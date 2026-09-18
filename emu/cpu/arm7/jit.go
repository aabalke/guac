package arm7

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/aabalke/gojit"
	"golang.org/x/exp/constraints"
)

var (
	CPU  = gojit.R9
	REG  = int32(unsafe.Offsetof(Cpu{}.Reg))
	R    = REG + int32(unsafe.Offsetof(Reg{}.R))
	CPSR = REG + int32(unsafe.Offsetof(Reg{}.CPSR))

	MODE = gojit.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cond{}.Mode)), Bits: 32}
	jN   = gojit.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cond{}.N)), Bits: 8}
	jZ   = gojit.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cond{}.Z)), Bits: 8}
	jC   = gojit.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cond{}.C)), Bits: 8}
	jV   = gojit.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cond{}.V)), Bits: 8}

	FALSE = gojit.Imm(0)
	TRUE  = gojit.Imm(1)
)

type Jit struct {
	*gojit.Assembler
	cpu *Cpu

	EndBlock bool

	TestingCnt int
}

func NewJit(cpu *Cpu) *Jit {
	return &Jit{
		cpu: cpu,
	}
}

func (j *Jit) UseJit[T constraints.Unsigned](op T, f func(op T)) {
	j.TestingCnt++

	fmt.Printf("starting test cnt %08d, op %08X\n", j.TestingCnt, op)

	asm, err := gojit.New(gojit.PageSize)
	if err != nil {
		panic(err)
	}

	j.Assembler = asm

	j.MovAbs(uint64(uintptr(unsafe.Pointer(j.cpu))), CPU)

	f(op)

	asm.Exit()

	if err := asm.Error(); err != nil {
		panic(err)
	}

	gojit.CallJit(uintptr(unsafe.Pointer(&asm.Buf[0])))

	asm.Release()
}

func (j *Jit) RunTest[T constraints.Unsigned](op T, f func(op T)) func() {
	cpu := j.cpu
	start := cpu.Reg
	staStamp := j.cpu.Timestamp

	ewramPtr := j.cpu.Mem.ReadPtr(0x200_0000)
	iwramPtr := j.cpu.Mem.ReadPtr(0x300_0000)
	ewram := *(*[0x40000]uint8)(ewramPtr)
	iwram := *(*[0x8000]uint8)(iwramPtr)

	j.UseJit(op, f)

	sav := cpu.Reg
	savStamp := cpu.Timestamp

	*(*[0x40000]uint8)(ewramPtr) = ewram
	*(*[0x8000]uint8)(iwramPtr) = iwram

	cpu.Reg = start

	// returns exit test func, which should be deferred until end of interpreted func

	return func() {
		// do not (Reg) == (Reg), sta = cpu.Reg does not promise padding
		if match := (cpu.Reg.R == sav.R &&
			cpu.Reg.CPSR == sav.CPSR &&
			cpu.Reg.SPSR == sav.SPSR &&
			cpu.Reg.FIQ == sav.FIQ &&
			cpu.Reg.LR == sav.LR &&
			cpu.Reg.SP == sav.SP &&
			cpu.Reg.USR == sav.USR &&
			cpu.Timestamp-savStamp == savStamp-staStamp); match {
			return // match
		}

		s := ""
		s += fmt.Sprintf("STA REG %08X CPSR %08X\n", start.R, start.CPSR.Get())
		s += fmt.Sprintf("JIT REG %08X CPSR %08X\n", sav.R, sav.CPSR.Get())
		s += fmt.Sprintf("COR REG %08X CPSR %08X\n", cpu.Reg.R, cpu.Reg.CPSR.Get())

		s += fmt.Sprintf("Time Diff Cor %08X Jit %08X\n", cpu.Timestamp-savStamp, savStamp-staStamp)

		s += fmt.Sprintf("STA USRREG %08X\n", start.USR)
		s += fmt.Sprintf("JIT USRREG %08X\n", sav.USR)
		s += fmt.Sprintf("COR USRREG %08X\n", cpu.Reg.USR)

		s += fmt.Sprintf("STA LR %08X\n", start.LR)
		s += fmt.Sprintf("JIT LR %08X\n", sav.LR)
		s += fmt.Sprintf("COR LR %08X\n", cpu.Reg.LR)

		s += fmt.Sprintf("STA SP %08X\n", start.SP)
		s += fmt.Sprintf("JIT SP %08X\n", sav.SP)
		s += fmt.Sprintf("COR SP %08X\n", cpu.Reg.SP)

		s += fmt.Sprintf("STA FIQ %08X\n", start.FIQ)
		s += fmt.Sprintf("JIT FIQ %08X\n", sav.FIQ)
		s += fmt.Sprintf("COR FIQ %08X\n", cpu.Reg.FIQ)

		fmt.Printf("%s", s)

		os.Exit(0)
	}
}

func (j *Jit) CallFunc(f any) {
	j.InternalCallFunc(f)
}

func (j *Jit) REG(i uint32) gojit.Indirect {
	return gojit.Indirect{
		Base:   CPU,
		Offset: R + int32(i*4),
		Bits:   32,
	}
}

//go:nosplit
func Idle(c *Cpu, cycles int64) {
	c.Idle(cycles)
}
