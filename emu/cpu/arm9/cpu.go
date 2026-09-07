package arm9

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/aabalke/guac/common/debug"
	"github.com/aabalke/guac/emu/cpu/arm7"
)

const (
	SP = 13
	LR = 14
	PC = 15
)

type Cpu struct {
	*arm7.Cpu
	Cp15    *Cp15
	Itcm    *Tcm
	Dtcm    *Tcm
	Timings *Timings

	InstCycles    int64
	DataCycles    int64
	IdleCycles    int64
	CyclesPerInst int64

	tick func(cycles int64)
}

func NewCpu(m arm7.Mem, tick func(cycles int64)) *Cpu {
	c := &Cpu{
		Cpu:  &arm7.Cpu{},
		tick: tick,
	}

	c.Mem = m
	c.Cycles = c.Cycles9
	c.Idle = c.Idle9

	c.Timings = NewTimings()
	c.Itcm = NewTcm(0x8000)
	c.Dtcm = NewTcm(0x4000)
	c.Cp15 = NewCp15(c)

	c.LowVector = false

	c.Bus = &Bus9{
		c: c,
	}

	return c
}

func (c *Cpu) Step() {
	if c.IrqLine {

		c.Halted = false

		if !c.Reg.CPSR.I {
			c.CheckIrq()
			c.ReloadPipe()
			//c.tick(max(c.InstCycles, c.DataCycles) + c.IdleCycles)
			//c.InstCycles, c.DataCycles, c.IdleCycles = 0, 0, 0
		}
	}

	inst := c.Op[0]

	c.print2(inst)

	c.Seq = arm7.SEQ
	c.Op[0] = c.Op[1]

	if c.Reg.CPSR.T {

		if c.Reg.R[PC]&2 == 0 {
			if c.Itcm.Readable(c.Reg.R[PC]) {
				c.InstCycles++
			} else {
				c.InstCycles += c.CyclesPerInst
			}
		}

		if c.PcPtr == nil {
			if v, ok := c.Itcm.Read16(c.Reg.R[PC]); ok {
				c.Op[1] = v
			} else {
				c.Op[1] = c.Mem.Read16(c.Reg.R[PC])
			}
		} else {
			c.Op[1] = *(*uint32)(c.PcPtr) & 0xFFFF
		}

		c.DecodeThumb(uint16(inst))

		if c.Reload {
			c.ReloadPipe()
		} else {
			c.Reg.R[PC] += 2
			if c.PcPtr != nil {
				c.PcPtr = unsafe.Add(c.PcPtr, 2)
			}
		}

	} else {

		if c.Itcm.Readable(c.Reg.R[PC]) {
			c.InstCycles++
		} else {
			c.InstCycles += c.CyclesPerInst
		}

		if c.PcPtr == nil {
			if v, ok := c.Itcm.Read32(c.Reg.R[PC]); ok {
				c.Op[1] = v
			} else {
				c.Op[1] = c.Mem.Read32(c.Reg.R[PC])
			}
		} else {
			c.Op[1] = *(*uint32)(c.PcPtr)
		}

		c.DecodeArm(inst)

		if c.Reload {
			c.ReloadPipe()
		} else {
			c.Reg.R[PC] += 4
			if c.PcPtr != nil {
				c.PcPtr = unsafe.Add(c.PcPtr, 4)
			}
		}
	}

	c.Reload = false

	c.print()
	c.tick(max(c.InstCycles, c.DataCycles) + c.IdleCycles)

	c.InstCycles, c.DataCycles, c.IdleCycles = 0, 0, 0
}

var prev int64

func (c *Cpu) print() {
	//if !debug.B[0] {
	//	return
	//}
	//fmt.Printf("PC %08X Stamp %08d Diff %08d: %02d %02d %02d\n", c.Reg.R[15], c.Timestamp, c.Timestamp-prev, c.InstCycles, c.DataCycles, c.IdleCycles)
	fmt.Printf("PC %08X Diff %08d: %02d %02d %02d\n", c.Reg.R[15], c.Timestamp-prev, c.InstCycles, c.DataCycles, c.IdleCycles)
	prev = c.Timestamp
}

func (c *Cpu) print2(inst uint32) {
	fmt.Printf("OP %08X\n", inst)
	if debug.V[0] > 10000 {
		os.Exit(0)
	} else {
		debug.V[0]++
	}
}

func (c *Cpu) ReloadPipe() {
	if c.Reg.CPSR.T {
		c.Reload16()
	} else {
		c.Reload32()
	}
}

func (c *Cpu) Reload16() {
	pc := c.Reg.R[PC] &^ 1

	if c.Itcm.Readable(pc) {
		c.PcPtr = c.Itcm.ReadPtr(pc)

		c.InstCycles++

		if c.PcPtr == nil {
			c.Op[0], _ = c.Itcm.Read16(pc + 0)
			c.Op[1], _ = c.Itcm.Read16(pc + 2)
		} else {
			c.Op[0] = *(*uint32)(c.PcPtr) & 0xFFFF
			c.PcPtr = unsafe.Add(c.PcPtr, 2)
			c.Op[1] = *(*uint32)(c.PcPtr) & 0xFFFF
			c.PcPtr = unsafe.Add(c.PcPtr, 2)
		}

	} else {

		c.PcPtr = c.Mem.ReadPtr(pc)

		c.SetCyclesPerInst(pc)
		c.InstCycles += c.CyclesPerInst // pc + 0
		if pc&2 != 0 {
			c.InstCycles += c.CyclesPerInst // pc + 2
		}

		if c.PcPtr == nil {
			c.Op[0] = c.Mem.Read16(pc + 0)
			c.Op[1] = c.Mem.Read16(pc + 2)
		} else {
			c.Op[0] = *(*uint32)(c.PcPtr) & 0xFFFF
			c.PcPtr = unsafe.Add(c.PcPtr, 2)
			c.Op[1] = *(*uint32)(c.PcPtr) & 0xFFFF
			c.PcPtr = unsafe.Add(c.PcPtr, 2)
		}
	}

	c.Reg.R[PC] += 4
	c.Seq = arm7.SEQ
}

func (c *Cpu) Reload32() {
	pc := c.Reg.R[PC] &^ 3

	if c.Itcm.Readable(pc) {
		c.PcPtr = c.Itcm.ReadPtr(pc)

		c.InstCycles++

		if c.PcPtr == nil {
			c.Op[0], _ = c.Itcm.Read32(pc + 0)
			c.Op[1], _ = c.Itcm.Read32(pc + 4)

		} else {
			c.Op[0] = *(*uint32)(c.PcPtr)
			c.PcPtr = unsafe.Add(c.PcPtr, 4)
			c.Op[1] = *(*uint32)(c.PcPtr)
			c.PcPtr = unsafe.Add(c.PcPtr, 4)
		}

	} else {

		c.PcPtr = c.Mem.ReadPtr(pc)

		c.SetCyclesPerInst(pc)
		c.InstCycles += c.CyclesPerInst * 2 // pc + 0, pc + 4

		if c.PcPtr == nil {
			c.Op[0] = c.Mem.Read32(pc + 0)
			c.Op[1] = c.Mem.Read32(pc + 4)
		} else {
			c.Op[0] = *(*uint32)(c.PcPtr)
			c.PcPtr = unsafe.Add(c.PcPtr, 4)
			c.Op[1] = *(*uint32)(c.PcPtr)
			c.PcPtr = unsafe.Add(c.PcPtr, 4)
		}
	}

	c.Reg.R[PC] += 8
	c.Seq = arm7.SEQ
}

//go:nosplit
func (c *Cpu) ToggleThumb() {
	c.Reg.CPSR.T = c.Reg.R[PC]&1 != 0
	c.Reload = true

	if c.Reg.CPSR.T {
		c.Reg.R[PC] &^= 1
		return
	}
	c.Reg.R[PC] &^= 3
}

func (c *Cpu) Exception(addr arm7.ExceptionVector, mode arm7.CpuMode) {
	cpsr := &c.Reg.CPSR
	thumb := cpsr.T

	c.ModeSwitch(cpsr.Mode, mode)

	i := arm7.ModeBank[mode]
	c.Reg.SPSR[i] = *cpsr

	if thumb {
		c.Reg.R[LR] = c.Reg.R[PC] - 2
	} else {
		c.Reg.R[LR] = c.Reg.R[PC] - 4
	}

	cpsr.Mode = mode
	cpsr.T = false
	cpsr.I = true
	if mode == arm7.MODE_FIQ {
		cpsr.F = true
	}

	if c.LowVector {
		addr &= 0xFFFF
	}

	c.Reg.R[PC] = uint32(addr)

	c.Reload = true
}

func (c *Cpu) SetCyclesPerInst(addr uint32) {
	// since always at reload, always instruction, nonseq, width 4

	// >> 12 4KB pages, uint64 bitmask, 1bit per 4kb page
	idx := (addr >> 12) / 64
	bit := (addr >> 12) & 63

	if c.Cp15.ProtectionUnit.InstCache.Pages[idx]&(1<<bit) != 0 {
		c.CyclesPerInst = 1
		return
	}

	region := addr >> 24
	cycles := int64(c.Timings[region][2]) << 1

	if penalty := region != 2; penalty {
		cycles += 6
	}

	c.CyclesPerInst = cycles
}

func (c *Cpu) Idle9(cycles int64) {
	// idle 9 is provided as 66mhz cycles
	c.IdleCycles += int64(cycles)
}

func (c *Cpu) Cycles9(addr, width, seq uint32, inst bool) {
	if inst {
		c.InstCycles += c.CyclesPerInst
		return
	}

	// >> 12 4KB pages, uint64 bitmask, 1bit per 4kb page
	idx := (addr >> 12) / 64
	bit := (addr >> 12) & 63

	if c.Cp15.ProtectionUnit.DataCache.Pages[idx]&(1<<bit) != 0 {

		if seq == arm7.SEQ {
			c.DataCycles++
		} else {
			c.DataCycles += 3 // this is rough estimate
		}
		return
	}

	region := addr >> 24
	cycles := int64(c.Timings[region][((width>>2)<<1)|seq]) << 1

	if penalty := region != 2 && seq == arm7.NONSEQ; penalty {
		cycles += 6
	}

	c.DataCycles += cycles
}
