package arm7

import (
	"fmt"
	"os"
	"reflect"
	"unsafe"

	"github.com/aabalke/gojit"
	"golang.org/x/exp/constraints"
)

var (
	JIT = gojit.Rsi
	CPU = gojit.R9
)

type ReloadState uint32

const (
	NONE ReloadState = iota
	RELOAD
	POSSIBLE
)

type Jit struct {
	*gojit.Assembler
	cpu          *Cpu
	C            CpuPtrs
	Metrics      [][]uint32
	BlockCache   *BlockCache
	Pages        []*Page
	invalidPages []*Page
	Config       JitConfig
	ReloadState  ReloadState
	TestingCnt   int
}

type JitConfig struct {
	AddressSpace   int    // addr space with readable instructions
	PageShift      uint32 // density of pages, in address space
	PageMask       uint32 // mask used to calc blocks per page
	NativePagesize int    // byte cnt on native memory per block
	MinInstCnt     uint32 // blocks smaller than this are skipped
	MaxInstCnt     uint32 // block cannot be more inst than this
	LoopThreshold  uint32 // how many loops until create block
	BlockCnt       int    // max how many jit blocks
	Enabled        bool
}

// NOTE: Used so different shaped structs can be used as cpu (arm7, arm9...)
type CpuPtrs struct {
	Cpu           uintptr
	Cpsr          uintptr
	R             [16]gojit.Indirect
	Mode          gojit.Indirect
	N, Z, C, V, T gojit.Indirect
	Reload        gojit.Indirect
	Seq           gojit.Indirect
}

type Page struct {
	id     uint32
	Blocks []*JitBlock
	dead   bool
}

func NewJit(cpu *Cpu, config JitConfig, ptrs CpuPtrs) *Jit {
	if !config.Enabled { // testing jit
		return &Jit{cpu: cpu, Config: config, C: ptrs}
	}

	return &Jit{
		cpu:   cpu,
		Pages: make([]*Page, config.AddressSpace>>config.PageShift),
		BlockCache: InitBlockCache(
			uint32(config.BlockCnt),
			config.NativePagesize,
		),
		Metrics: make([][]uint32, config.AddressSpace>>config.PageShift),
		Config:  config,
		C:       ptrs,
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

	pageIdx := addr >> j.Config.PageShift

	page := j.Pages[pageIdx]
	if page == nil || page.dead {
		return
	}

	//fmt.Printf("Invalidated Page %08X Addr %08X\n", addr>>j.PageShift, addr)

	page.dead = true

	j.Pages[pageIdx] = nil
	j.Metrics[pageIdx] = make([]uint32, (1<<j.Config.PageShift)>>1)
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
	pageIdx := pc >> j.Config.PageShift
	blockIdx := (pc & j.Config.PageMask) >> 1 // aligned to word for thumb

	if metrics := j.Metrics[pageIdx]; metrics == nil {
		j.Metrics[pageIdx] = make([]uint32, (1<<j.Config.PageShift)>>1)
	}

	j.Metrics[pageIdx][blockIdx]++
	if j.Metrics[pageIdx][blockIdx] <= j.Config.LoopThreshold {
		return
	}

	j.CreateBlock(pc, w)
}

func (j *Jit) TryJit(pc uint32) bool {
	pageIdx := pc >> j.Config.PageShift
	blockIdx := (pc & j.Config.PageMask) >> 1

	page := j.Pages[pageIdx]

	if page == nil || page.dead {
		return false
	}

	block := page.Blocks[blockIdx]

	if block == nil || block.Skip || block.f == nil {
		return false
	}

	//fmt.Printf("Running Jit for PC %08X\n", pc)

	block.f()
	j.BlockCache.TouchBlock(block)
	return true
}

func (j *Jit) UseJit[T constraints.Unsigned](op T) {
	j.TestingCnt++

	fmt.Printf("starting test cnt %08d, op %08X\n", j.TestingCnt, op)

	asm, err := gojit.New(j.Config.NativePagesize)
	if err != nil {
		panic(err)
	}

	j.Assembler = asm

	j.MovAbs(uint64(uintptr(unsafe.Pointer(j))), JIT)
	j.MovAbs(uint64(j.C.Cpu), CPU)

	switch reflect.TypeOf(op).Kind() {
	case reflect.Uint16:
		j.emitThumb(uint16(op))
	case reflect.Uint32:
		j.emitArm(uint32(op))
	}

	asm.Exit()

	if err := asm.Error(); err != nil {
		panic(err)
	}

	gojit.CallJit(uintptr(unsafe.Pointer(&asm.Buf[0])))

	asm.Release()
}

func (j *Jit) RunTest[T constraints.Unsigned](op T) func() {
	cpu := j.cpu
	start := cpu.Reg
	staStamp := j.cpu.Timestamp

	//ewramPtr := j.cpu.Mem.ReadPtr(0x200_0000)
	//iwramPtr := j.cpu.Mem.ReadPtr(0x300_0000)
	//vramPtr := j.cpu.Mem.ReadPtr(0x600_0000)
	//ewram := *(*[0x40000]uint8)(ewramPtr)
	//iwram := *(*[0x8000]uint8)(iwramPtr)
	//vram := *(*[0x18001]uint8)(vramPtr)

	j.UseJit(op)

	//savedIwram := *(*[0x8000]uint8)(iwramPtr)

	sav := cpu.Reg
	savStamp := cpu.Timestamp

	//*(*[0x40000]uint8)(ewramPtr) = ewram
	//*(*[0x8000]uint8)(iwramPtr) = iwram
	//*(*[0x18001]uint8)(vramPtr) = vram

	cpu.Reg = start

	// returns exit test func, which should be deferred until end of interpreted func

	return func() {
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

//go:nosplit
func (j *Jit) Idle(cycles int64) { j.cpu.Idle(cycles) }

//go:nosplit
func (j *Jit) Read8(addr uint32) uint32 { return j.cpu.Read8(addr) }

//go:nosplit
func (j *Jit) Read16(addr uint32) uint32 { return j.cpu.Read16(addr) }

//go:nosplit
func (j *Jit) Read32(addr uint32) uint32 { return j.cpu.Read32(addr) }

//go:nosplit
func (j *Jit) Read32Block(addr, seq uint32) uint32 { return j.cpu.Read32Block(addr, seq) }

//go:nosplit
func (j *Jit) Write8(addr uint32, v uint8) { j.cpu.Write8(addr, v) }

//go:nosplit
func (j *Jit) Write16(addr uint32, v uint16) { j.cpu.Write16(addr, v) }

//go:nosplit
func (j *Jit) Write32(addr, v uint32) { j.cpu.Write32(addr, v) }

//go:nosplit
func (j *Jit) Write32Block(addr, v, seq uint32) { j.cpu.Write32Block(addr, v, seq) }

//go:nosplit
func (j *Jit) ModeSwitch(curr, next CpuMode) {
	j.cpu.ModeSwitch(curr, next)
}

//go:nosplit
func (j *Jit) GetSPSR(mode CpuMode) {
	j.cpu.GetSPSR(mode)
}

//go:nosplit
func (j *Jit) Step() bool {
	c := j.cpu

	if c.IrqLine {
		return true
	}

	seq := c.Seq
	c.Seq = SEQ

	w := uint32(4)
	if c.Reg.CPSR.T {
		w = 2
	}

	c.Cycles(c.Reg.R[PC], w, seq, true)

	return false
}

//go:nosplit
func (j *Jit) UpdatePc(p unsafe.Pointer, w uint32) {
	j.cpu.Reg.R[PC] += w
	if j.cpu.PcPtr != nil {
		j.cpu.PcPtr = unsafe.Add(j.cpu.PcPtr, w)
	}

	// NOTE: when exiting jit, need pipeline setup properly
	// ONLY when not reloading. This removes every inst pipeline adjustment

	mask := uint32(0xFFFF_FFFF >> ((w & 2) * 8))

	p = unsafe.Add(p, w)
	j.cpu.Op[0] = *(*uint32)(p) & mask
	p = unsafe.Add(p, w)
	j.cpu.Op[1] = *(*uint32)(p) & mask
	p = unsafe.Add(p, w)

	j.cpu.PcPtr = p
}

//go:nosplit
func (j *Jit) ReloadPipe() {
	j.cpu.ReloadPipe()
}

//go:nosplit
func (j *Jit) Exception(addr ExceptionVector, mode CpuMode) {
	j.cpu.Exception(addr, mode)
}

//go:nosplit
func (j *Jit) ExitException(mode CpuMode) {
	j.cpu.ExitException(mode)
}

//go:nosplit
func (j *Jit) ToggleThumb() {
	j.cpu.ToggleThumb()
}

//go:nosplit
func (j *Jit) DoMsrModeSwitch(spsrFlag bool, v, mask uint32) {
	j.cpu.DoMsrModeSwitch(spsrFlag, v, mask)
}

//go:nosplit
func (j *Jit) DoLdmLoadSwitch() {
	j.cpu.DoLdmLoadSwitch()
}
