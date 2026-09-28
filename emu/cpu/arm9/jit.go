package arm9

import "github.com/aabalke/guac/emu/cpu/arm7"

type Jit struct {
	arm7.Jit
}

func NewJit(cpu arm7.JittedCpu, mem arm7.Mem, config arm7.JitConfig, ptrs arm7.CpuPtrs) *Jit {
	if !config.Enabled { // testing jit
		return &Jit{Cpu: cpu, Mem: mem, Config: config, C: ptrs}
	}

	return &Jit{
		Cpu:   cpu,
		Mem:   mem,
		Pages: make([]*arm7.Page, config.AddressSpace>>config.PageShift),
		BlockCache: arm7.InitBlockCache(
			uint32(config.BlockCnt),
			config.NativePageSize,
		),
		Metrics: make([][]uint32, config.AddressSpace>>config.PageShift),
		Config:  config,
		C:       ptrs,
	}
}

// NOTE: Nosplit directive is not applied on inherited struct methods
// all methods called by arm9 jit inherited from arm7 must be explicitly added

// TODO: need to check Cond, cpu etc has proper explicit methods

//go:nosplit
func (j *Jit) Idle(cycles int64) { j.Cpu.Idle(cycles) }

//go:nosplit
func (j *Jit) Read8(addr uint32) uint32 { return j.Cpu.Read8(addr) }

//go:nosplit
func (j *Jit) Read16(addr uint32) uint32 { return j.Cpu.Read16(addr) }

//go:nosplit
func (j *Jit) Read32(addr uint32) uint32 { return j.Cpu.Read32(addr) }

//go:nosplit
func (j *Jit) Read32Block(addr, seq uint32) uint32 { return j.Cpu.Read32Block(addr, seq) }

//go:nosplit
func (j *Jit) Write8(addr uint32, v uint8) { j.Cpu.Write8(addr, v) }

//go:nosplit
func (j *Jit) Write16(addr uint32, v uint16) { j.Cpu.Write16(addr, v) }

//go:nosplit
func (j *Jit) Write32(addr, v uint32) { j.Cpu.Write32(addr, v) }

//go:nosplit
func (j *Jit) Write32Block(addr, v, seq uint32) { j.Cpu.Write32Block(addr, v, seq) }

//go:nosplit
func (j *Jit) ModeSwitch(curr, next arm7.CpuMode) { j.Cpu.ModeSwitch(curr, next) }

//go:nosplit
func (j *Jit) ReloadPipe() { j.Cpu.ReloadPipe() }

//go:nosplit
func (j *Jit) Exception(addr arm7.ExceptionVector, mode arm7.CpuMode) { j.Cpu.Exception(addr, mode) }

//go:nosplit
func (j *Jit) ExitException(mode arm7.CpuMode) { j.Cpu.ExitException(mode) }

//go:nosplit
func (j *Jit) InstCycles(pc, w, seq uint32) { j.Cpu.Cycles(pc, w, seq, true) }

//go:nosplit
func (j *Jit) DoMsrModeSwitch(spsrFlag bool, v, mask uint32) {
	j.Cpu.DoMsrModeSwitch(spsrFlag, v, mask)
}

//go:nosplit
func (j *Jit) DoLdmLoadSwitch() { j.Cpu.DoLdmLoadSwitch() }
