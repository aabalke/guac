package arm7

import (
	"github.com/aabalke/gojit"
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
	Mem Mem
	//cpu *Cpu
	cpu          JittedCpu
	C            CpuPtrs
	Metrics      [][]uint32
	BlockCache   *BlockCache
	Pages        []*Page
	invalidPages []*Page
	Config       JitConfig
	ReloadState  ReloadState
	TestingCnt   int
}

type JittedCpu interface {
	Read8(addr uint32) uint32
	Read16(addr uint32) uint32
	Read32(addr uint32) uint32
	Read32Block(addr, seq uint32) uint32
	Write8(addr uint32, v uint8)
	Write16(addr uint32, v uint16)
	Write32(addr, v uint32)
	Write32Block(addr, v, seq uint32)

	Idle(cycles int64)
	Cycles(pc, w, seq uint32, inst bool)

	ModeSwitch(CpuMode, CpuMode)
	ReloadPipe()
	Exception(ExceptionVector, CpuMode)
	ExitException(CpuMode)
	DoMsrModeSwitch(bool, uint32, uint32)
	DoLdmLoadSwitch()
}

type JitConfig struct {
	AddressSpace   int    // addr space with readable instructions
	PageShift      uint32 // density of pages, in address space
	PageMask       uint32 // mask used to calc blocks per page
	NativePageSize int    // byte cnt on native memory per block
	MinInstCnt     uint32 // blocks smaller than this are skipped
	MaxInstCnt     uint32 // block cannot be more inst than this
	LoopThreshold  uint32 // how many loops until create block
	BlockCnt       int    // max how many jit blocks
	Enabled        bool
}

// NOTE: Used so different shaped structs can be used as cpu (arm7, arm9...)
type CpuPtrs struct {
	Cpu                  uintptr
	Cpsr, Spsr           uintptr
	R                    [16]gojit.Indirect
	Mode                 gojit.Indirect
	N, Z, C, V, T        gojit.Indirect
	Reload, Seq, IrqLine gojit.Indirect
	PcPtr                gojit.Indirect
	Op                   [2]gojit.Indirect
}

type Page struct {
	id     uint32
	Blocks []*JitBlock
	dead   bool
}

func NewJit(cpu *Cpu, config JitConfig, ptrs CpuPtrs) *Jit {
	if !config.Enabled { // testing jit
		return &Jit{cpu: cpu, Mem: cpu.Mem, Config: config, C: ptrs}
	}

	return &Jit{
		cpu:   cpu,
		Mem:   cpu.Mem,
		Pages: make([]*Page, config.AddressSpace>>config.PageShift),
		BlockCache: InitBlockCache(
			uint32(config.BlockCnt),
			config.NativePageSize,
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
func (j *Jit) ModeSwitch(curr, next CpuMode) { j.cpu.ModeSwitch(curr, next) }

//go:nosplit
func (j *Jit) ReloadPipe() { j.cpu.ReloadPipe() }

//go:nosplit
func (j *Jit) Exception(addr ExceptionVector, mode CpuMode) { j.cpu.Exception(addr, mode) }

//go:nosplit
func (j *Jit) ExitException(mode CpuMode) { j.cpu.ExitException(mode) }

//go:nosplit
func (j *Jit) InstCycles(pc, w, seq uint32) { j.cpu.Cycles(pc, w, seq, true) }

//go:nosplit
func (j *Jit) DoMsrModeSwitch(spsrFlag bool, v, mask uint32) {
	j.cpu.DoMsrModeSwitch(spsrFlag, v, mask)
}

//go:nosplit
func (j *Jit) DoLdmLoadSwitch() { j.cpu.DoLdmLoadSwitch() }
