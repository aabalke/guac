package arm7

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/aabalke/gojit"
	"golang.org/x/exp/constraints"
)

var (
	JIT  = gojit.Rsi
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

const (
	ADDRESS_SPACE = 0x1_0000_0000
	PAGE_SHIFT    = 16
	PAGE_SIZE     = 0x10000
	PAGE_MASK     = (1 << PAGE_SHIFT) - 1

	BLOCK_CNT      = 4096
	BATCH_INST_MAX = 32
	LOOP_CNT       = 255
)

type Jit struct {
	*gojit.Assembler
	cpu *Cpu

	EndBlock bool

	TestingCnt int

	BlockCache    *BlockCache
	Pages         []*Page
	Metrics       [ADDRESS_SPACE >> PAGE_SHIFT][]uint32
	invalidPages  []*Page
	LoopThreshold uint32
	PageShift     uint32
	PageMask      uint32
}

type Page struct {
	id     uint32
	Blocks []*JitBlock
	dead   bool
}

func NewJit(cpu *Cpu) *Jit {
	return &Jit{
		cpu:   cpu,
		Pages: make([]*Page, ADDRESS_SPACE>>PAGE_SHIFT),
		BlockCache: InitBlockCache(
			BLOCK_CNT,
			PAGE_SIZE,
		),
		LoopThreshold: LOOP_CNT,
		PageShift:     PAGE_SHIFT,
		PageMask:      PAGE_MASK,
	}
}

func (j *Jit) Close() {
	if j.BlockCache != nil {
		j.BlockCache.Close()
	}
}

func (j *Jit) InvalidatePage(addr uint32) {
	if j.Pages == nil {
		return
	}

	page := j.Pages[addr>>j.PageShift]
	if page == nil || page.dead {
		return
	}

	page.dead = true

	j.Pages[addr>>j.PageShift] = nil
	j.Metrics[addr>>j.PageShift] = make([]uint32, (1<<j.PageShift)>>1)
	j.invalidPages = append(j.invalidPages, page)
}

func (j *Jit) DeletePages() {
	if len(j.invalidPages) == 0 {
		return
	}

	for _, page := range j.invalidPages {
		for i := range page.Blocks {
			if block := page.Blocks[i]; block != nil {
				page.Blocks[i] = nil

				if !block.Skip {
					j.BlockCache.InvalidateBlock(block)
				}
			}
		}
	}

	j.invalidPages = j.invalidPages[:0]
}

func (j *Jit) UpdateMetrics(pc, w uint32) {
	pageIdx := pc >> j.PageShift
	blockIdx := (pc & j.PageMask) >> 1 // aligned to word for thumb

	if metrics := j.Metrics[pageIdx]; metrics == nil {
		j.Metrics[pageIdx] = make([]uint32, (1<<j.PageShift)>>1)
	}

	j.Metrics[pageIdx][blockIdx]++
	if j.Metrics[pageIdx][blockIdx] <= j.LoopThreshold {
		return
	}

	j.CreateBlock(pc, w)
}

func (j *Jit) UseJit[T constraints.Unsigned](op T, f func(op T)) {
	j.TestingCnt++

	fmt.Printf("starting test cnt %08d, op %08X\n", j.TestingCnt, op)

	asm, err := gojit.New(gojit.PageSize)
	if err != nil {
		panic(err)
	}

	j.Assembler = asm

	j.MovAbs(uint64(uintptr(unsafe.Pointer(j))), JIT)
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

func (j *Jit) REG(i uint32) gojit.Indirect {
	return gojit.Indirect{
		Base:   CPU,
		Offset: R + int32(i*4),
		Bits:   32,
	}
}

//go:nosplit
func (j *Jit) Idle(cycles int64) {
	j.cpu.Idle(cycles)
}

//go:nosplit
func (j *Jit) Read8(addr uint32) uint32 {
	return j.cpu.Read8(addr)
}

//go:nosplit
func (j *Jit) Read16(addr uint32) uint32 {
	return j.cpu.Read16(addr)
}

//go:nosplit
func (j *Jit) Read32(addr uint32) uint32 {
	return j.cpu.Read32(addr)
}

//go:nosplit
func (j *Jit) Read32Block(addr, seq uint32) uint32 {
	return j.cpu.Read32Block(addr, seq)
}

//go:nosplit
func (j *Jit) Write8(addr uint32, v uint8) {
	j.cpu.Write8(addr, v)
}

//go:nosplit
func (j *Jit) Write16(addr uint32, v uint16) {
	j.cpu.Write16(addr, v)
}

//go:nosplit
func (j *Jit) Write32(addr, v uint32) {
	j.cpu.Write32(addr, v)
}

//go:nosplit
func (j *Jit) Write32Block(addr, v, seq uint32) {
	j.cpu.Write32Block(addr, seq, v)
}

//go:nosplit
func (j *Jit) ModeSwitch(curr, next CpuMode) {
	j.cpu.ModeSwitch(curr, next)
}

//go:nosplit
func (j *Jit) GetSPSR(mode CpuMode) {
	j.cpu.GetSPSR(mode)
}

//go:nosplit
func (j *Jit) Step() {
	c := j.cpu

	if c.IrqLine {
		panic("irq called during jit step")
	}

	seq := c.Seq
	c.Seq = SEQ
	c.Op[0] = c.Op[1]

	w := uint32(4)
	if c.Reg.CPSR.T {
		w = 2
	}

	c.Cycles(c.Reg.R[PC], w, seq, true)

	if c.PcPtr == nil {
		if w == 4 {
			c.Op[1] = c.Mem.Read32(c.Reg.R[PC])
		} else {
			c.Op[1] = c.Mem.Read16(c.Reg.R[PC])
		}
	} else {
		// 0xFFFF_FFFF uint32, 0xFFFF uint16
		mask := uint32(0xFFFF_FFFF >> ((w & 2) * 8))
		c.Op[1] = *(*uint32)(c.PcPtr) & mask
	}
}
