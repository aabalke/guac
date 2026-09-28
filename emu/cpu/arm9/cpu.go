package arm9

import (
	"fmt"
	"os"
	"reflect"
	"unsafe"

	"github.com/aabalke/gojit"
	"github.com/aabalke/guac/emu/cpu/arm7"
	"golang.org/x/exp/constraints"
)

const (
	SP = 13
	LR = 14
	PC = 15
)

type Cpu struct {
	arm7.Cpu
	Cp15         *Cp15
	Itcm         *Tcm
	Dtcm         *Tcm
	TestJit, Jit *Jit

	tick             func(cycles int64)
	setCyclesPerInst func(addr uint32)

	InstCycles    int64
	DataCycles    int64
	IdleCycles    int64
	CyclesPerInst int64
}

func NewCpu(m arm7.Mem, jitConfig arm7.JitConfig, idle, tick func(int64), cycles func(addr, width, seq uint32, inst bool), setCyclesPerInst func(addr uint32)) *Cpu {
	c := &Cpu{
		Cpu:              arm7.Cpu{},
		Itcm:             NewTcm(0x8000),
		Dtcm:             NewTcm(0x4000),
		setCyclesPerInst: setCyclesPerInst,
		tick:             tick,
	}

	c.Mem = m
	c.Cp15 = NewCp15(c)
	c.LowVector = false
	c.CyclesFunc = cycles
	c.IdleFunc = idle

	c.Bus = &Bus9{
		c: c,
	}

	cpuPtrs := GetCpuPtrs(c)

	if jitConfig.Enabled {
		c.Jit = NewJit(c, m, jitConfig, cpuPtrs)
	} else {
		c.TestJit = NewJit(c, m, jitConfig, cpuPtrs)
	}

	return c
}

func GetCpuPtrs(cpu *Cpu) arm7.CpuPtrs {
	var (
		reg  = int32(unsafe.Offsetof(Cpu{}.Reg))
		r    = reg + int32(unsafe.Offsetof(arm7.Reg{}.R))
		cpsr = reg + int32(unsafe.Offsetof(arm7.Reg{}.CPSR))
		op   = int32(unsafe.Offsetof(Cpu{}.Op))
	)

	c := arm7.CpuPtrs{
		Mode:    gojit.Indirect{Base: arm7.CPU, Offset: cpsr + int32(unsafe.Offsetof(arm7.Cond{}.Mode)), Bits: 32},
		N:       gojit.Indirect{Base: arm7.CPU, Offset: cpsr + int32(unsafe.Offsetof(arm7.Cond{}.N)), Bits: 8},
		Z:       gojit.Indirect{Base: arm7.CPU, Offset: cpsr + int32(unsafe.Offsetof(arm7.Cond{}.Z)), Bits: 8},
		C:       gojit.Indirect{Base: arm7.CPU, Offset: cpsr + int32(unsafe.Offsetof(arm7.Cond{}.C)), Bits: 8},
		V:       gojit.Indirect{Base: arm7.CPU, Offset: cpsr + int32(unsafe.Offsetof(arm7.Cond{}.V)), Bits: 8},
		Q:       gojit.Indirect{Base: arm7.CPU, Offset: cpsr + int32(unsafe.Offsetof(arm7.Cond{}.Q)), Bits: 8},
		T:       gojit.Indirect{Base: arm7.CPU, Offset: cpsr + int32(unsafe.Offsetof(arm7.Cond{}.T)), Bits: 8},
		IrqLine: gojit.Indirect{Base: arm7.CPU, Offset: int32(unsafe.Offsetof(Cpu{}.IrqLine)), Bits: 8},
		Reload:  gojit.Indirect{Base: arm7.CPU, Offset: int32(unsafe.Offsetof(Cpu{}.Reload)), Bits: 8},
		Seq:     gojit.Indirect{Base: arm7.CPU, Offset: int32(unsafe.Offsetof(Cpu{}.Seq)), Bits: 32},
		PcPtr:   gojit.Indirect{Base: arm7.CPU, Offset: int32(unsafe.Offsetof(Cpu{}.PcPtr)), Bits: 64},
		Spsr:    uintptr(unsafe.Pointer(&cpu.Reg.SPSR)),
		Cpsr:    uintptr(unsafe.Pointer(&cpu.Reg.CPSR)),
		Cp15:    uintptr(unsafe.Pointer(cpu.Cp15)),
		Cpu:     uintptr(unsafe.Pointer(cpu)),
	}

	for i := range 16 {
		c.R[i] = gojit.Indirect{Base: arm7.CPU, Offset: r + int32(i*4), Bits: 32}
	}

	for i := range 2 {
		c.Op[i] = gojit.Indirect{Base: arm7.CPU, Offset: op + int32(i*4), Bits: 32}
	}

	return c
}

func (c *Cpu) Step() {
	if c.IrqLine {

		c.Halted = false

		if !c.Reg.CPSR.I {
			c.DoIrq()
			c.ReloadPipe()
		}
	}

	inst := c.Op[0]

	//c.print2(inst)

	c.Seq = arm7.SEQ
	c.Op[0] = c.Op[1]

	w := uint32(4)
	if c.Reg.CPSR.T {
		w = 2
	}

	bus := c.Mem
	if c.Itcm.Readable(c.Reg.R[PC], w) {
		bus = c.Itcm
	}

	if w == 4 || c.Reg.R[PC]&2 == 0 {
		c.Cycles(c.Reg.R[PC], w, 0, true)
	}

	if c.PcPtr == nil {
		if w == 4 {
			c.Op[1] = bus.Read32(c.Reg.R[PC])
		} else {
			c.Op[1] = bus.Read16(c.Reg.R[PC])
		}
	} else {
		// 0xFFFF_FFFF uint32, 0xFFFF uint16
		mask := uint32(0xFFFF_FFFF >> ((w & 2) * 8))
		c.Op[1] = *(*uint32)(c.PcPtr) & mask
	}

	if w == 4 {
		c.DecodeArm(inst)
	} else {
		c.DecodeThumb(uint16(inst))
	}

	if c.Reload {
		c.ReloadPipe()
	} else {
		c.Reg.R[PC] += w
		if c.PcPtr != nil {
			c.PcPtr = unsafe.Add(c.PcPtr, w)
		}
	}

	//c.print()
	c.tick(max(c.InstCycles, c.DataCycles) + c.IdleCycles)

	c.InstCycles, c.DataCycles, c.IdleCycles = 0, 0, 0
}

func (c *Cpu) ReloadPipe() {
	w := uint32(4)
	if c.Reg.CPSR.T {
		w = 2
	}

	pc := c.Reg.R[PC] &^ (w - 1)

	bus := c.Mem
	if c.Itcm.Readable(pc, w) {
		bus = c.Itcm
		c.Cycles(pc, w, 0, true)
	} else {
		c.setCyclesPerInst(pc)

		c.Cycles(pc, w, 0, true) // pc + 0

		if w == 4 || pc&2 != 0 {
			c.Cycles(pc, w, 0, true) // pc + 2 or pc + 4
		}
	}

	c.PcPtr = bus.ReadPtr(pc)

	if c.PcPtr == nil {
		if w == 4 {
			c.Op[0] = bus.Read32(pc + 0)
			c.Op[1] = bus.Read32(pc + w)
		} else {
			c.Op[0] = bus.Read16(pc + 0)
			c.Op[1] = bus.Read16(pc + w)
		}
	} else {
		// 0xFFFF_FFFF uint32, 0xFFFF uint16
		mask := uint32(0xFFFF_FFFF >> ((w & 2) * 8))
		c.Op[0] = *(*uint32)(c.PcPtr) & mask
		c.PcPtr = unsafe.Add(c.PcPtr, w)
		c.Op[1] = *(*uint32)(c.PcPtr) & mask
		c.PcPtr = unsafe.Add(c.PcPtr, w)
	}

	c.Reg.R[PC] += w * 2
	c.Seq = arm7.SEQ
	c.Reload = false
}

//func (c *Cpu) print() {
//	fmt.Printf("PC %08X Diff %08d: %02d %02d %02d\n", c.Reg.R[15], c.Timestamp-debug.Vi64[0], c.InstCycles, c.DataCycles, c.IdleCycles)
//	debug.Vi64[0] = c.Timestamp
//}
//
//func (c *Cpu) print2(inst uint32) {
//	fmt.Printf("OP %08X\n", inst)
//	if debug.V[0] > 10000 {
//		os.Exit(0)
//	} else {
//		debug.V[0]++
//	}
//}

func (c *Cpu) UseJit[T constraints.Unsigned](op T) {
	j := c.TestJit

	if j == nil {
		panic("attemped to test jit with jit active (no test jit present)")
	}

	j.TestingCnt++

	fmt.Printf("starting test cnt %08d, op %08X\n", j.TestingCnt, op)

	asm, err := gojit.New(j.Config.NativePageSize)
	if err != nil {
		panic(err)
	}

	j.Assembler = asm

	j.MovAbs(uint64(uintptr(unsafe.Pointer(j))), arm7.JIT)
	j.MovAbs(uint64(uintptr(unsafe.Pointer(c))), arm7.CPU)

	switch reflect.TypeOf(op).Kind() {
	case reflect.Uint16:
		j.EmitThumb(uint16(op))
	case reflect.Uint32:
		j.EmitArm(uint32(op))
	}

	asm.Exit()

	if err := asm.Error(); err != nil {
		panic(err)
	}

	gojit.CallJit(uintptr(unsafe.Pointer(&asm.Buf[0])))

	asm.Release()
}

func (c *Cpu) RunJitTest[T constraints.Unsigned](op T) func() {
	start := c.Reg
	staStamp := c.Timestamp
	staReload := c.Reload

	//var cp15 Cp15
	//if IsCoDataReg(uint32(op)) {
	//	cp15 = *c.Cp15
	//}

	//ewramPtr := j.cpu.Mem.ReadPtr(0x200_0000)
	//iwramPtr := j.cpu.Mem.ReadPtr(0x300_0000)
	//vramPtr := j.cpu.Mem.ReadPtr(0x600_0000)
	//ewram := *(*[0x40000]uint8)(ewramPtr)
	//iwram := *(*[0x8000]uint8)(iwramPtr)
	//vram := *(*[0x18001]uint8)(vramPtr)

	c.UseJit(op)

	//savedIwram := *(*[0x8000]uint8)(iwramPtr)

	sav := c.Reg
	savStamp := c.Timestamp
	savReload := c.Reload
	//var savCp15 Cp15
	//if IsCoDataReg(uint32(op)) {
	//	savCp15 = *c.Cp15
	//}

	//*(*[0x40000]uint8)(ewramPtr) = ewram
	//*(*[0x8000]uint8)(iwramPtr) = iwram
	//*(*[0x18001]uint8)(vramPtr) = vram

	//*c.Cp15 = cp15
	c.Reg = start
	c.Reload = staReload

	// returns exit test func, which should be deferred until end of interpreted func

	return func() {
		//if savCp15.Ctrl != 0 && (savCp15.Ctrl != c.Cp15.Ctrl ||
		//	savCp15.ProtectionUnit != c.Cp15.ProtectionUnit) {
		//	fmt.Printf("CoDataInvalid\n")
		//}

		// do not (Reg) == (Reg), sta = cpu.Reg does not promise padding

		// dirty := false
		// for i := 0; i < len(savedIwram); i += 4 {
		//	jit := binary.LittleEndian.Uint32(savedIwram[i:])
		//	interpreter := binary.LittleEndian.Uint32((*[0x8000]uint8)(iwramPtr)[i:])

		//	if jit != interpreter {
		//		dirty = true
		//		fmt.Printf("ADDR 0x300...%04X: Jit %08X Interpreter %08X\n", i, jit, interpreter)
		//	}
		//}

		//if dirty {
		//	panic("invalid memory values")
		//}

		if match := (c.Reg.R == sav.R &&
			c.Reg.CPSR == sav.CPSR &&
			c.Reg.SPSR == sav.SPSR &&
			c.Reg.FIQ == sav.FIQ &&
			c.Reg.LR == sav.LR &&
			c.Reg.SP == sav.SP &&
			c.Reg.USR == sav.USR &&
			c.Timestamp-savStamp == savStamp-staStamp &&
			c.Reload == savReload); match {
			return // match
		}

		s := ""
		s += fmt.Sprintf("STA REG %08X CPSR %08X\n", start.R, start.CPSR.Get())
		s += fmt.Sprintf("JIT REG %08X CPSR %08X\n", sav.R, sav.CPSR.Get())
		s += fmt.Sprintf("COR REG %08X CPSR %08X\n", c.Reg.R, c.Reg.CPSR.Get())

		s += fmt.Sprintf("Reload Prior %t Cor %t Jit %t\n", staReload, c.Reload, savReload)
		s += fmt.Sprintf("Time Diff Cor %08X Jit %08X\n", c.Timestamp-savStamp, savStamp-staStamp)

		s += fmt.Sprintf("STA USRREG %08X\n", start.USR)
		s += fmt.Sprintf("JIT USRREG %08X\n", sav.USR)
		s += fmt.Sprintf("COR USRREG %08X\n", c.Reg.USR)

		s += fmt.Sprintf("STA LR %08X\n", start.LR)
		s += fmt.Sprintf("JIT LR %08X\n", sav.LR)
		s += fmt.Sprintf("COR LR %08X\n", c.Reg.LR)

		s += fmt.Sprintf("STA SP %08X\n", start.SP)
		s += fmt.Sprintf("JIT SP %08X\n", sav.SP)
		s += fmt.Sprintf("COR SP %08X\n", c.Reg.SP)

		s += fmt.Sprintf("STA FIQ %08X\n", start.FIQ)
		s += fmt.Sprintf("JIT FIQ %08X\n", sav.FIQ)
		s += fmt.Sprintf("COR FIQ %08X\n", c.Reg.FIQ)

		fmt.Printf("%s", s)

		os.Exit(0)
	}
}
