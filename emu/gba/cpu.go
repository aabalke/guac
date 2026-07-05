package gba

import "unsafe"

type Cpu struct {
	gba               *GBA
	PcPtr             unsafe.Pointer
	ParallelDmaCycles uint32
	Op                [2]uint32
	Reg               Reg
	Halted            bool
	NonSeq            bool
	LastWasDma        bool
	IrqLine           bool
	Reloaded          bool
}

const (
	SP = 13
	LR = 14
	PC = 15

	FLAG_N = 31
	FLAG_Z = 30
	FLAG_C = 29
	FLAG_V = 28
	FLAG_Q = 27
	FLAG_I = 7
	FLAG_F = 6
	FLAG_T = 5

	MODE_USR = 0x10
	MODE_FIQ = 0x11
	MODE_IRQ = 0x12
	MODE_SWI = 0x13
	MODE_ABT = 0x17
	MODE_UND = 0x1B
	MODE_SYS = 0x1F
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

var BANK_ID = map[uint32]uint32{
	MODE_USR: 0,
	MODE_SYS: 0,
	MODE_FIQ: 1,
	MODE_IRQ: 2,
	MODE_SWI: 3,
	MODE_ABT: 4,
	MODE_UND: 5,
}

func NewCpu(g *GBA) *Cpu {
	return &Cpu{
		NonSeq: true,
		gba:    g,
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
				mode  = uint32(MODE_IRQ)
				seq   = !c.NonSeq
			)

			c.NonSeq = false

			if thumb {
				c.Cycles(c.Reg.R[15], 2, false, seq, true)
				c.gba.Mem.Read16(c.Reg.R[15])
			} else {
				c.Cycles(c.Reg.R[15], 4, false, seq, true)
				c.gba.Mem.Read32(c.Reg.R[15])
			}

			c.ModeSwitch(cpsr.Mode, mode)

			i := BANK_ID[mode]
			c.Reg.SPSR[i] = *cpsr

			if thumb {
				c.Reg.R[LR] = c.Reg.R[15]
				c.Reg.LR[i] = c.Reg.R[15]

			} else {
				c.Reg.R[LR] = c.Reg.R[15] - 4
				c.Reg.LR[i] = c.Reg.R[15] - 4
			}

			cpsr.Mode = mode
			cpsr.T = false
			cpsr.I = true

			c.Reg.R[PC] = addr

			c.Reload32()
		}
	}

	inst := c.Op[0]

	seq := !c.NonSeq
	c.NonSeq = false
	c.Op[0] = c.Op[1]

	if c.Reg.CPSR.T {

		c.Cycles(c.Reg.R[15], 2, false, seq, true)

		if c.PcPtr == nil {
			c.Op[1] = c.gba.Mem.Read16(c.Reg.R[15])
		} else {
			c.Op[1] = *(*uint32)(c.PcPtr) & 0xFFFF
		}

		c.DecodeTHUMB(uint16(inst))

		if !c.Reloaded {
			c.Reg.R[15] += 2
			if c.PcPtr != nil {
				c.PcPtr = unsafe.Add(c.PcPtr, 2)
			}
		}

	} else {

		c.Cycles(c.Reg.R[15], 4, false, seq, true)
		if c.PcPtr == nil {
			c.Op[1] = c.gba.Mem.Read32(c.Reg.R[15])
		} else {
			c.Op[1] = *(*uint32)(c.PcPtr)
		}

		c.DecodeARM(inst)

		if !c.Reloaded {
			c.Reg.R[15] += 4
			if c.PcPtr != nil {
				c.PcPtr = unsafe.Add(c.PcPtr, 4)
			}
		}
	}

	c.Reloaded = false
}

func (c *Cpu) Reload16() {
	pc := &c.Reg.R[15]

	c.PcPtr = c.gba.Mem.ReadPtr(c.Reg.R[15])

	c.Cycles(*pc, 2, false, false, true)
	c.Cycles(*pc+2, 2, false, true, true)

	if c.PcPtr == nil {
		c.Op[0] = c.gba.Mem.Read16(*pc + 0)
		c.Op[1] = c.gba.Mem.Read16(*pc + 2)
	} else {
		c.Op[0] = *(*uint32)(c.PcPtr) & 0xFFFF
		c.PcPtr = unsafe.Add(c.PcPtr, 2)
		c.Op[1] = *(*uint32)(c.PcPtr) & 0xFFFF
		c.PcPtr = unsafe.Add(c.PcPtr, 2)
	}

	*pc += 4
	c.Reloaded = true
	c.NonSeq = false
}

func (c *Cpu) Reload32() {
	pc := &c.Reg.R[15]

	c.PcPtr = c.gba.Mem.ReadPtr(c.Reg.R[15])

	c.Cycles(*pc, 4, false, false, true)
	c.Cycles(*pc+4, 4, false, true, true)

	if c.PcPtr == nil {
		c.Op[0] = c.gba.Mem.Read32(*pc + 0)
		c.Op[1] = c.gba.Mem.Read32(*pc + 4)
	} else {
		c.Op[0] = *(*uint32)(c.PcPtr)
		c.PcPtr = unsafe.Add(c.PcPtr, 4)
		c.Op[1] = *(*uint32)(c.PcPtr)
		c.PcPtr = unsafe.Add(c.PcPtr, 4)
	}

	*pc += 8
	c.Reloaded = true
	c.NonSeq = false
}

type Reg struct {
	R    [16]uint32
	CPSR Cond

	SP   [6]uint32
	LR   [6]uint32
	FIQ  [5]uint32 // r8 - r12
	USR  [5]uint32 // r8 - r12 // tmp to restore after FIQ
	SPSR [6]Cond
}

type Cond struct {
	Mode                   uint32
	N, Z, C, V, Q, I, F, T bool
}

//go:nosplit
func (c *Cond) Get() uint32 {
	v := c.Mode

	if c.N {
		v |= 1 << FLAG_N
	}
	if c.Z {
		v |= 1 << FLAG_Z
	}
	if c.C {
		v |= 1 << FLAG_C
	}
	if c.V {
		v |= 1 << FLAG_V
	}
	if c.Q {
		v |= 1 << FLAG_Q
	}
	if c.I {
		v |= 1 << FLAG_I
	}
	if c.F {
		v |= 1 << FLAG_F
	}
	if c.T {
		v |= 1 << FLAG_T
	}

	return v
}

func (c *Cond) Set(v uint32) {
	c.N = (v>>FLAG_N)&1 != 0
	c.Z = (v>>FLAG_Z)&1 != 0
	c.C = (v>>FLAG_C)&1 != 0
	c.V = (v>>FLAG_V)&1 != 0
	c.Q = (v>>FLAG_Q)&1 != 0
	c.I = (v>>FLAG_I)&1 != 0
	c.F = (v>>FLAG_F)&1 != 0
	c.T = (v>>FLAG_T)&1 != 0
	c.Mode = v & 0x1F
}

func (cpu *Cpu) ToggleThumb() {
	cpu.Reg.CPSR.T = cpu.Reg.R[15]&1 != 0

	if cpu.Reg.CPSR.T {
		cpu.Reg.R[15] &^= 1
		return
	}
	cpu.Reg.R[15] &^= 3
}

func (c *Cpu) Write8(addr uint32, v uint8) {
	c.Cycles(addr, 1, false, false, false)
	c.gba.Mem.Write8(addr, v)
	c.NonSeq = true
}

func (c *Cpu) Write16(addr uint32, v uint16) {
	c.Cycles(addr, 2, false, false, false)
	c.gba.Mem.Write16(addr, v)
	c.NonSeq = true
}

func (c *Cpu) Write32(addr uint32, v uint32) {
	c.Cycles(addr, 4, false, false, false)
	c.gba.Mem.Write32(addr, v)
	c.NonSeq = true
}

func (c *Cpu) Write32Block(addr, v uint32, seq bool) {
	c.Cycles(addr, 4, false, seq, false)
	c.gba.Mem.Write32(addr, v)
	c.NonSeq = true
}

func (c *Cpu) Read8(addr uint32) uint32 {
	c.Cycles(addr, 1, false, false, false)
	v := c.gba.Mem.Read8(addr)
	c.idle(1)
	return v
}

func (c *Cpu) Read16(addr uint32) uint32 {
	c.Cycles(addr, 2, false, false, false)
	v := c.gba.Mem.Read16(addr)
	c.idle(1)
	return v
}

func (c *Cpu) Read32(addr uint32) uint32 {
	c.Cycles(addr, 4, false, false, false)
	v := c.gba.Mem.Read32(addr)
	c.idle(1)
	return v
}

func (c *Cpu) Read32Block(addr uint32, seq bool) uint32 {
	c.Cycles(addr, 4, false, seq, false)
	return c.gba.Mem.Read32(addr)
}

func (c *Cpu) Cycles(addr, width uint32, dma, seq, inst bool) {
	if dma {
		c.LastWasDma = true
	} else if c.gba.IsRunning() {
		c.gba.CheckDmas()
	}

	c.ParallelDmaCycles = 0

	switch addr >> 24 {
	case 2:
		c.gba.Tick(3 << (width >> 2))
	case 5, 6:
		c.gba.Tick(1 << (width >> 2))
	case 8, 9, 10, 11, 12, 13:

		if addr&0x1FFFF == 0 {
			seq = false
		}

		if c.LastWasDma && !dma {
			c.LastWasDma = false
			seq = false
		}

		if !inst {
			c.gba.Mem.Prefetch.Cancel(c.Reg.R[15])
			c.gba.Tick(c.gba.Mem.Waitstate.Get(width, addr, seq))
			return
		}

		w := uint32(4)
		if c.Reg.CPSR.T {
			w = 2
		}

		cycles := c.gba.Mem.Waitstate.Get(width, addr, seq)
		c.gba.Mem.Prefetch.Wait(c.Reg.R[15], addr, w, cycles)

	case 14, 15:
		c.gba.Mem.Prefetch.Cancel(c.Reg.R[15])
		c.gba.Tick(c.gba.Mem.Waitstate.Get(width, addr, seq))
	default:
		c.gba.Tick(1)
	}
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

func (c *Cpu) idle(cycles int64) {
	if c.gba.IsRunning() {
		c.ParallelDmaCycles = c.gba.CheckDmas()
	}

	if c.ParallelDmaCycles == 0 {
		c.gba.Tick(cycles)
		c.NonSeq = true
	} else {
		c.ParallelDmaCycles--
	}
}

func (c *Cpu) ModeSwitch(curr, next uint32) {
	// DO NOT RELOAD PIPE AFTER CALLING ModeSwitch

	reg := &c.Reg
	r := &c.Reg.R

	if curr != MODE_FIQ {
		for i := range 5 {
			reg.USR[i] = r[8+i]
		}
	}

	reg.SP[BANK_ID[curr]] = r[SP]
	reg.LR[BANK_ID[curr]] = r[LR]

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

	r[SP] = reg.SP[BANK_ID[next]]
	r[LR] = reg.LR[BANK_ID[next]]

	if next == MODE_FIQ {
		for i := range 5 {
			r[8+i] = reg.FIQ[i]
		}
	}
}

const (
	VEC_RESET         = 0x00
	VEC_UND           = 0x04
	VEC_SWI           = 0x08
	VEC_PREFETCHABORT = 0x0C
	VEC_DATAABORT     = 0x10
	VEC_ADDR26BIT     = 0x14
	VEC_IRQ           = 0x18
	VEC_FIQ           = 0x1C
)

func (c *Cpu) Exception(addr uint32, mode uint32) {
	cpsr := &c.Reg.CPSR
	thumb := cpsr.T

	c.ModeSwitch(cpsr.Mode, mode)

	i := BANK_ID[mode]
	c.Reg.SPSR[i] = *cpsr

	if thumb {
		c.Reg.R[LR] = c.Reg.R[15] - 2
		c.Reg.LR[i] = c.Reg.R[15] - 2

	} else {
		c.Reg.R[LR] = c.Reg.R[15] - 4
		c.Reg.LR[i] = c.Reg.R[15] - 4

	}

	cpsr.Mode = mode
	cpsr.T = false
	cpsr.I = true
	if mode == MODE_FIQ {
		cpsr.F = true
	}

	c.Reg.R[PC] = addr
	c.Reload32()
}

func (c *Cpu) ExitException(mode uint32) {
	c.Reg.CPSR = c.Reg.SPSR[BANK_ID[mode]]
	c.ModeSwitch(mode, c.Reg.CPSR.Mode)
}
