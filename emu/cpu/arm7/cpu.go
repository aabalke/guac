package arm7

import (
	"unsafe"
)

type Cpu struct {
	Mem        Mem
	Cycles     func(addr, width, seq uint32, inst bool)
	Idle       func(cycles int64)
	PcPtr      unsafe.Pointer
	Reg        Reg
	Op         [2]uint32
	Seq        uint32
	Halted     bool
	LastWasDma bool
	IrqLine    bool
	Reloaded   bool
	LowVector  bool
}

type Mem interface {
	Read8(addr uint32) uint32
	Read16(addr uint32) uint32
	Read32(addr uint32) uint32
	Write8(addr uint32, v uint8)
	Write16(addr uint32, v uint16)
	Write32(addr uint32, v uint32)
	ReadPtr(addr uint32) unsafe.Pointer
}

type Reg struct {
	R    [16]uint32
	CPSR Cond

	SP   [6]uint32
	LR   [6]uint32
	FIQ  [5]uint32 // r8 - r12
	USR  [5]uint32 // r8 - r12
	SPSR [6]Cond
}

type Cond struct {
	Mode                   CpuMode
	N, Z, C, V, Q, I, F, T bool
}

//go:nosplit
func (c *Cond) Get() uint32 {
	v := uint32(c.Mode)

	if c.N {
		v |= 1 << N
	}
	if c.Z {
		v |= 1 << Z
	}
	if c.C {
		v |= 1 << C
	}
	if c.V {
		v |= 1 << V
	}
	if c.Q {
		v |= 1 << Q
	}
	if c.I {
		v |= 1 << I
	}
	if c.F {
		v |= 1 << F
	}
	if c.T {
		v |= 1 << T
	}

	return v
}

//go:nosplit
func (c *Cond) Set(v uint32) {
	c.N = (v>>N)&1 != 0
	c.Z = (v>>Z)&1 != 0
	c.C = (v>>C)&1 != 0
	c.V = (v>>V)&1 != 0
	c.Q = (v>>Q)&1 != 0
	c.I = (v>>I)&1 != 0
	c.F = (v>>F)&1 != 0
	c.T = (v>>T)&1 != 0
	c.Mode = CpuMode(v & 0x1F)
}

const (
	SP = 13
	LR = 14
	PC = 15

	N = 31
	Z = 30
	C = 29
	V = 28
	Q = 27
	I = 7
	F = 6
	T = 5

	NONSEQ = 0
	SEQ    = 1
)

type CpuMode uint32

const (
	MODE_USR CpuMode = 0x10
	MODE_FIQ CpuMode = 0x11
	MODE_IRQ CpuMode = 0x12
	MODE_SWI CpuMode = 0x13
	MODE_ABT CpuMode = 0x17
	MODE_UND CpuMode = 0x1B
	MODE_SYS CpuMode = 0x1F
)

var ModeBank = map[CpuMode]uint32{
	MODE_USR: 0,
	MODE_SYS: 0,
	MODE_FIQ: 1,
	MODE_IRQ: 2,
	MODE_SWI: 3,
	MODE_ABT: 4,
	MODE_UND: 5,
}

type ExceptionVector uint32

const (
	VEC_RESET     ExceptionVector = 0xFFFF0000
	VEC_UNDEFINED ExceptionVector = 0xFFFF0004
	VEC_SWI       ExceptionVector = 0xFFFF0008
	VEC_PREFETCH  ExceptionVector = 0xFFFF000C
	VEC_IRQ       ExceptionVector = 0xFFFF0018
	VEC_FIQ       ExceptionVector = 0xFFFF001C
)

func (c *Cond) CheckCond(cond uint32) bool {
	switch cond {
	case 0xE: // AL (always)
		return true
	case 0x0: // EQ
		return c.Z
	case 0x1: // NE
		return !c.Z
	case 0x2: // CS/HS
		return c.C
	case 0x3: // CC/LO
		return !c.C
	case 0x4: // MI
		return c.N
	case 0x5: // PL
		return !c.N
	case 0x6: // VS
		return c.V
	case 0x7: // VC
		return !c.V
	case 0x8: // HI
		return c.C && !c.Z
	case 0x9: // LS
		return !c.C || c.Z
	case 0xA: // GE
		return c.N == c.V
	case 0xB: // LT
		return c.N != c.V
	case 0xC: // GT
		return !c.Z && (c.N == c.V)
	case 0xD: // LE
		return c.Z || (c.N != c.V)
	default: // NV
		return false
	}
}

func NewCpu(mem Mem, cycles func(addr, width, seq uint32, inst bool), idle func(cycles int64)) *Cpu {
	return &Cpu{
		Mem:       mem,
		Cycles:    cycles,
		Idle:      idle,
		LowVector: true,
	}
}

func (c *Cpu) Step() {
	if c.IrqLine {
		c.Halted = false

		if !c.Reg.CPSR.I {

			var (
				cpsr  = &c.Reg.CPSR
				thumb = cpsr.T
				addr  = uint32(VEC_IRQ)
				mode  = MODE_IRQ
				seq   = c.Seq
			)

			c.Seq = SEQ

			if thumb {
				c.Cycles(c.Reg.R[PC], 2, seq, true)
				c.Mem.Read16(c.Reg.R[PC])
			} else {
				c.Cycles(c.Reg.R[PC], 4, seq, true)
				c.Mem.Read32(c.Reg.R[PC])
			}

			c.ModeSwitch(cpsr.Mode, mode)

			i := ModeBank[mode]
			c.Reg.SPSR[i] = *cpsr

			if thumb {
				c.Reg.R[LR] = c.Reg.R[PC]
				c.Reg.LR[i] = c.Reg.R[PC]

			} else {
				c.Reg.R[LR] = c.Reg.R[PC] - 4
				c.Reg.LR[i] = c.Reg.R[PC] - 4
			}

			cpsr.Mode = mode
			cpsr.T = false
			cpsr.I = true

			if c.LowVector {
				addr &= 0xFFFF
			}

			c.Reg.R[PC] = addr

			c.Reload32()
		}
	}

	inst := c.Op[0]
	seq := c.Seq
	c.Seq = SEQ
	c.Op[0] = c.Op[1]

	if c.Reg.CPSR.T {

		c.Cycles(c.Reg.R[PC], 2, seq, true)

		if c.PcPtr == nil {
			c.Op[1] = c.Mem.Read16(c.Reg.R[PC])
		} else {
			c.Op[1] = *(*uint32)(c.PcPtr) & 0xFFFF
		}

		c.DecodeTHUMB(uint16(inst))

		if !c.Reloaded {
			c.Reg.R[PC] += 2
			if c.PcPtr != nil {
				c.PcPtr = unsafe.Add(c.PcPtr, 2)
			}
		}

	} else {

		c.Cycles(c.Reg.R[PC], 4, seq, true)
		if c.PcPtr == nil {
			c.Op[1] = c.Mem.Read32(c.Reg.R[PC])
		} else {
			c.Op[1] = *(*uint32)(c.PcPtr)
		}

		c.DecodeARM(inst)

		if !c.Reloaded {
			c.Reg.R[PC] += 4
			if c.PcPtr != nil {
				c.PcPtr = unsafe.Add(c.PcPtr, 4)
			}
		}
	}

	c.Reloaded = false
}

func (c *Cpu) Reload16() {
	pc := c.Reg.R[PC] &^ 1

	c.PcPtr = c.Mem.ReadPtr(pc)

	c.Cycles(pc+0, 2, NONSEQ, true)
	c.Cycles(pc+2, 2, SEQ, true)

	if c.PcPtr == nil {
		c.Op[0] = c.Mem.Read16(pc + 0)
		c.Op[1] = c.Mem.Read16(pc + 2)
	} else {
		c.Op[0] = *(*uint32)(c.PcPtr) & 0xFFFF
		c.PcPtr = unsafe.Add(c.PcPtr, 2)
		c.Op[1] = *(*uint32)(c.PcPtr) & 0xFFFF
		c.PcPtr = unsafe.Add(c.PcPtr, 2)
	}

	c.Reg.R[PC] += 4
	c.Reloaded = true
	c.Seq = SEQ
}

func (c *Cpu) Reload32() {
	pc := c.Reg.R[PC] &^ 3

	c.PcPtr = c.Mem.ReadPtr(pc)

	c.Cycles(pc+0, 4, NONSEQ, true)
	c.Cycles(pc+4, 4, SEQ, true)

	if c.PcPtr == nil {
		c.Op[0] = c.Mem.Read32(pc + 0)
		c.Op[1] = c.Mem.Read32(pc + 4)
	} else {
		c.Op[0] = *(*uint32)(c.PcPtr)
		c.PcPtr = unsafe.Add(c.PcPtr, 4)
		c.Op[1] = *(*uint32)(c.PcPtr)
		c.PcPtr = unsafe.Add(c.PcPtr, 4)
	}

	c.Reg.R[PC] += 8
	c.Reloaded = true
	c.Seq = SEQ
}

//go:nosplit
func (c *Cpu) ToggleThumb() {
	c.Reg.CPSR.T = c.Reg.R[PC]&1 != 0

	if c.Reg.CPSR.T {
		c.Reg.R[PC] &^= 1
		c.Reload16()
		return
	}
	c.Reg.R[PC] &^= 3
	c.Reload32()
}

//go:nosplit
func (c *Cpu) Write8(addr uint32, v uint8) {
	c.Cycles(addr, 1, NONSEQ, false)
	c.Mem.Write8(addr, v)
	c.Seq = NONSEQ
	c.LastWasDma = false
}

//go:nosplit
func (c *Cpu) Write16(addr uint32, v uint16) {
	c.Cycles(addr, 2, NONSEQ, false)
	c.Mem.Write16(addr, v)
	c.Seq = NONSEQ
	c.LastWasDma = false
}

//go:nosplit
func (c *Cpu) Write32(addr, v uint32) {
	c.Cycles(addr, 4, NONSEQ, false)
	c.Mem.Write32(addr, v)
	c.Seq = NONSEQ
	c.LastWasDma = false
}

//go:nosplit
func (c *Cpu) Write32Block(addr, v, seq uint32) {
	c.Cycles(addr, 4, seq, false)
	c.Mem.Write32(addr, v)
	c.Seq = NONSEQ
	c.LastWasDma = false
}

//go:nosplit
func (c *Cpu) Read8(addr uint32) uint32 {
	c.Cycles(addr, 1, NONSEQ, false)
	v := c.Mem.Read8(addr)
	c.Idle(1)
	c.LastWasDma = false
	return v
}

//go:nosplit
func (c *Cpu) Read16(addr uint32) uint32 {
	c.Cycles(addr, 2, NONSEQ, false)
	v := c.Mem.Read16(addr)
	c.Idle(1)
	c.LastWasDma = false
	return v
}

//go:nosplit
func (c *Cpu) Read32(addr uint32) uint32 {
	c.Cycles(addr, 4, NONSEQ, false)
	v := c.Mem.Read32(addr)
	c.Idle(1)
	c.LastWasDma = false
	return v
}

//go:nosplit
func (c *Cpu) Read32Block(addr, seq uint32) uint32 {
	c.Cycles(addr, 4, seq, false)
	v := c.Mem.Read32(addr)
	c.LastWasDma = false
	return v
}

func idleMul(rs uint32, sign bool) int64 {
	cycles := int64(1)
	mask := uint32(0xFFFFFF00)
	for {
		rs &= mask
		if rs == 0 {
			break
		}
		if sign && (rs == mask) {
			break
		}
		mask <<= 8
		cycles++
	}
	return cycles
}

func (c *Cpu) ModeSwitch(curr, next CpuMode) {
	// DO NOT RELOAD PIPE AFTER CALLING ModeSwitch

	reg := &c.Reg
	r := &c.Reg.R

	if curr != MODE_FIQ {
		for i := range 5 {
			reg.USR[i] = r[8+i]
		}
	}

	reg.SP[ModeBank[curr]] = r[SP]
	reg.LR[ModeBank[curr]] = r[LR]

	if curr == MODE_FIQ {
		for i := range 5 {
			reg.FIQ[i] = r[8+i]
		}
	}

	if next != MODE_FIQ {
		for i := range 5 {
			r[8+i] = reg.USR[i]
		}
	}

	r[SP] = reg.SP[ModeBank[next]]
	r[LR] = reg.LR[ModeBank[next]]

	if next == MODE_FIQ {
		for i := range 5 {
			r[8+i] = reg.FIQ[i]
		}
	}
}

func (c *Cpu) Exception(addr ExceptionVector, mode CpuMode) {
	cpsr := &c.Reg.CPSR
	thumb := cpsr.T

	c.ModeSwitch(cpsr.Mode, mode)

	i := ModeBank[mode]
	c.Reg.SPSR[i] = *cpsr

	if thumb {
		c.Reg.R[LR] = c.Reg.R[PC] - 2
		c.Reg.LR[i] = c.Reg.R[PC] - 2

	} else {
		c.Reg.R[LR] = c.Reg.R[PC] - 4
		c.Reg.LR[i] = c.Reg.R[PC] - 4

	}

	cpsr.Mode = mode
	cpsr.T = false
	cpsr.I = true
	if mode == MODE_FIQ {
		cpsr.F = true
	}

	if c.LowVector {
		addr &= 0xFFFF
	}

	c.Reg.R[PC] = uint32(addr)
	c.Reload32()
}

func (c *Cpu) ExitException(mode CpuMode) {
	c.Reg.CPSR = c.Reg.SPSR[ModeBank[mode]]
	c.ModeSwitch(mode, c.Reg.CPSR.Mode)
}
