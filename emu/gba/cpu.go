package gba

import (
	"unsafe"
)

type Cpu struct {
	gba               *GBA
	PcPtr             unsafe.Pointer
	ParallelDmaCycles uint32
	Op                [2]uint32
	Reg               Reg
	Halted            bool
	LastWasDma        bool
	IrqLine           bool
	Reloaded          bool
	Seq               uint32
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

	NONSEQ = 0
	SEQ    = 1
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
		gba: g,
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
				seq   = c.Seq
			)

			c.Seq = SEQ

			if thumb {
				c.CyclesInst(c.Reg.R[15], 2, seq)
				c.gba.Mem.Read16(c.Reg.R[15])
				c.LastWasDma = false
			} else {
				c.CyclesInst(c.Reg.R[15], 4, seq)
				c.gba.Mem.Read32(c.Reg.R[15])
				c.LastWasDma = false
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
	seq := c.Seq
	c.Seq = SEQ
	c.Op[0] = c.Op[1]

	if c.Reg.CPSR.T {

		c.CyclesInst(c.Reg.R[15], 2, seq)

		if c.PcPtr == nil {
			c.Op[1] = c.gba.Mem.Read16(c.Reg.R[15])
		} else {
			c.Op[1] = *(*uint32)(c.PcPtr) & 0xFFFF
		}
		c.LastWasDma = false

		c.DecodeTHUMB(uint16(inst))

		if !c.Reloaded {
			c.Reg.R[15] += 2
			if c.PcPtr != nil {
				c.PcPtr = unsafe.Add(c.PcPtr, 2)
			}
		}

	} else {

		c.CyclesInst(c.Reg.R[15], 4, seq)
		if c.PcPtr == nil {
			c.Op[1] = c.gba.Mem.Read32(c.Reg.R[15])
		} else {
			c.Op[1] = *(*uint32)(c.PcPtr)
		}
		c.LastWasDma = false

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

	c.CyclesInst(*pc+0, 2, NONSEQ)
	c.CyclesInst(*pc+2, 2, SEQ)

	if c.PcPtr == nil {
		c.Op[0] = c.gba.Mem.Read16(*pc + 0)
		c.Op[1] = c.gba.Mem.Read16(*pc + 2)
	} else {
		c.Op[0] = *(*uint32)(c.PcPtr) & 0xFFFF
		c.PcPtr = unsafe.Add(c.PcPtr, 2)
		c.Op[1] = *(*uint32)(c.PcPtr) & 0xFFFF
		c.PcPtr = unsafe.Add(c.PcPtr, 2)
	}
	c.LastWasDma = false

	*pc += 4
	c.Reloaded = true
	c.Seq = SEQ
}

func (c *Cpu) Reload32() {
	pc := &c.Reg.R[15]

	c.PcPtr = c.gba.Mem.ReadPtr(c.Reg.R[15])

	c.CyclesInst(*pc+0, 4, NONSEQ)
	c.CyclesInst(*pc+4, 4, SEQ)

	if c.PcPtr == nil {
		c.Op[0] = c.gba.Mem.Read32(*pc + 0)
		c.Op[1] = c.gba.Mem.Read32(*pc + 4)
	} else {
		c.Op[0] = *(*uint32)(c.PcPtr)
		c.PcPtr = unsafe.Add(c.PcPtr, 4)
		c.Op[1] = *(*uint32)(c.PcPtr)
		c.PcPtr = unsafe.Add(c.PcPtr, 4)
	}
	c.LastWasDma = false

	*pc += 8
	c.Reloaded = true
	c.Seq = SEQ
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
	c.Cycles(addr, 1, NONSEQ)
	c.gba.Mem.Write8(addr, v)
	c.LastWasDma = false
	c.Seq = NONSEQ
}

func (c *Cpu) Write16(addr uint32, v uint16) {
	c.Cycles(addr, 2, NONSEQ)
	c.gba.Mem.Write16(addr, v)
	c.LastWasDma = false
	c.Seq = NONSEQ
}

func (c *Cpu) Write32(addr uint32, v uint32) {
	c.Cycles(addr, 4, NONSEQ)
	c.gba.Mem.Write32(addr, v)
	c.LastWasDma = false
	c.Seq = NONSEQ
}

func (c *Cpu) Write32Block(addr, v, seq uint32) {
	c.Cycles(addr, 4, seq)
	c.gba.Mem.Write32(addr, v)
	c.LastWasDma = false
	c.Seq = NONSEQ
}

func (c *Cpu) Read8(addr uint32) uint32 {
	c.Cycles(addr, 1, NONSEQ)
	v := c.gba.Mem.Read8(addr)
	c.LastWasDma = false
	c.idle(1)
	return v
}

func (c *Cpu) Read16(addr uint32) uint32 {
	c.Cycles(addr, 2, NONSEQ)
	v := c.gba.Mem.Read16(addr)
	c.LastWasDma = false
	c.idle(1)
	return v
}

func (c *Cpu) Read32(addr uint32) uint32 {
	c.Cycles(addr, 4, NONSEQ)
	v := c.gba.Mem.Read32(addr)
	c.LastWasDma = false
	c.idle(1)
	return v
}

func (c *Cpu) Read32Block(addr, seq uint32) uint32 {
	c.Cycles(addr, 4, seq)
	v := c.gba.Mem.Read32(addr)
	c.LastWasDma = false
	return v
}

func (c *Cpu) CyclesDma(addr, width, seq uint32) {
	c.ParallelDmaCycles = 0

	switch region := addr >> 24; {
	case region < 8:
		c.gba.Tick(int64(c.gba.Mem.Timings.Timings[width>>2][0][region]))
	case region < 14:

		if addr&0x1FFFF == 0 {
			seq = NONSEQ
		}

		if c.gba.Mem.Timings.Active {
			c.gba.Mem.Timings.Cancel(c.Reg.R[15])
		}
		c.gba.Tick(int64(c.gba.Mem.Timings.Timings[width>>2][seq][region]))
	default:
		if c.gba.Mem.Timings.Active {
			c.gba.Mem.Timings.Cancel(c.Reg.R[15])
		}
		c.gba.Tick(int64(c.gba.Mem.Timings.Timings[width>>2][0][region]))
	}
}

func (c *Cpu) CyclesInst(addr, width, seq uint32) {
	if c.gba.Dma.IsRunning() {
		c.gba.CheckDmas()
	}

	c.ParallelDmaCycles = 0

	if region := addr >> 24; region < 8 {
		c.gba.Tick(int64(c.gba.Mem.Timings.Timings[width>>2][0][region]))
		return
	}

	if addr&0x1FFFF == 0 || c.LastWasDma {
		seq = NONSEQ
	}

	instWidth := uint32(4)
	if c.Reg.CPSR.T {
		instWidth = 2
	}

	t := c.gba.Mem.Timings

	t.Capacity = (1 << ((instWidth & 3) >> 1)) << 2

	if t.Active {
		if t.Opcodes != 0 && addr == t.Head {
			t.Opcodes--
			t.Head += instWidth
			t.Tick(1)
			return
		}

		if t.Countdown > 0 && addr == t.Addr {
			t.Tick(t.Countdown)
			t.Head = t.Addr
			t.Opcodes = 0
			return
		}

		t.Cancel(c.Reg.R[15])
	}

	cycles := int64(t.Timings[width>>2][seq][addr>>24])

	if t.Disabled {
		switch cycles {
		case int64(t.Timings[0][1][addr>>24]):
			cycles = int64(t.Timings[0][0][8])
		case int64(t.Timings[1][1][addr>>24]):
			cycles = int64(t.Timings[1][0][8])
		}
	}

	t.Disabled = false

	t.Tick(cycles)

	if t.Enabled {
		t.Active = true
		t.Opcodes = 0
		t.Width = instWidth
		t.AccessTime = int64(t.Timings[instWidth>>2][1][addr>>24])
		t.Countdown = int64(t.Timings[instWidth>>2][1][addr>>24])
		t.Addr = addr + instWidth
		t.Head = addr + instWidth
	}
}

func (c *Cpu) Cycles(addr, width, seq uint32) {
	if c.gba.Dma.IsRunning() {
		c.gba.CheckDmas()
	}

	c.ParallelDmaCycles = 0

	switch region := addr >> 24; {
	case region < 8:
		c.gba.Tick(int64(c.gba.Mem.Timings.Timings[width>>2][0][region]))
	case region < 14:

		if addr&0x1FFFF == 0 || c.LastWasDma {
			seq = NONSEQ
		}

		if c.gba.Mem.Timings.Active {
			c.gba.Mem.Timings.Cancel(c.Reg.R[15])
		}
		c.gba.Tick(int64(c.gba.Mem.Timings.Timings[width>>2][seq][region]))

	case region < 0x10:
		if c.gba.Mem.Timings.Active {
			c.gba.Mem.Timings.Cancel(c.Reg.R[15])
		}
		c.gba.Tick(int64(c.gba.Mem.Timings.Timings[width>>2][0][region]))

	default:
		c.gba.Tick(int64(c.gba.Mem.Timings.Timings[width>>2][0][0]))
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
	if c.gba.Dma.IsRunning() {
		c.ParallelDmaCycles = c.gba.CheckDmas()
	}

	if c.ParallelDmaCycles == 0 {
		c.gba.Tick(cycles)
		c.Seq = NONSEQ
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
