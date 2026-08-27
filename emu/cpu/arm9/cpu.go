package arm9

import (
	"unsafe"

	"github.com/aabalke/guac/emu/cpu/arm7"
	"github.com/aabalke/guac/emu/cpu/arm9/cp15"
)

const (
	SP = 13
	LR = 14
	PC = 15
)

type Cpu struct {
	*arm7.Cpu
	Cp15 *cp15.Cp15
}

func NewCpu(m arm7.Mem, cycles func(addr, width, seq uint32, inst bool), idle func(cycles int64), cp15 *cp15.Cp15) *Cpu {
	c := &Cpu{
		Cpu:  arm7.NewCpu(m, cycles, idle),
		Cp15: cp15,
	}

	c.LowVector = false

	return c
}

func (c *Cpu) Step() {
	if c.IrqLine {

		c.Halted = false

		if !c.Reg.CPSR.I {

			var (
				cpsr  = &c.Reg.CPSR
				thumb = cpsr.T
				mode  = arm7.MODE_IRQ
				seq   = c.Seq
			)

			c.Seq = arm7.SEQ

			if thumb {
				c.Cycles(c.Reg.R[PC], 2, seq, true)
				c.Mem.Read16(c.Reg.R[PC])
			} else {
				c.Cycles(c.Reg.R[PC], 4, seq, true)
				c.Mem.Read32(c.Reg.R[PC])
			}

			c.ModeSwitch(cpsr.Mode, mode)

			i := arm7.ModeBank[mode]
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

			addr := uint32(arm7.VEC_IRQ)
			if c.LowVector {
				addr &= 0xFFFF
			}

			c.Reg.R[PC] = addr

			c.Reload32()
		}
	}

	inst := c.Op[0]
	seq := c.Seq
	c.Seq = arm7.SEQ
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
