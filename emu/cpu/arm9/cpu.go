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
	Cp15 *Cp15
	Itcm *Tcm
	Dtcm *Tcm

	tick             func(cycles int64)
	setCyclesPerInst func(addr uint32)

	InstCycles    int64
	DataCycles    int64
	IdleCycles    int64
	CyclesPerInst int64
}

func NewCpu(m arm7.Mem, idle, tick func(int64), cycles func(addr, width, seq uint32, inst bool), setCyclesPerInst func(addr uint32)) *Cpu {
	c := &Cpu{
		Cpu:              &arm7.Cpu{},
		Itcm:             NewTcm(0x8000),
		Dtcm:             NewTcm(0x4000),
		setCyclesPerInst: setCyclesPerInst,
		tick:             tick,
	}

	c.Mem = m
	c.Cp15 = NewCp15(c)
	c.LowVector = false
	c.Cycles = cycles
	c.Idle = idle

	c.Bus = &Bus9{
		c: c,
	}

	return c
}

func (c *Cpu) Step() {
	if c.IrqLine {

		c.Halted = false

		if !c.Reg.CPSR.I {
			c.DoIrq()
			c.ReloadPipe()
		}
	}

	inst := c.Op[0]

	//c.print2(inst)

	c.Seq = arm7.SEQ
	c.Op[0] = c.Op[1]

	w := uint32(4)
	if c.Reg.CPSR.T {
		w = 2
	}

	bus := c.Mem
	if c.Itcm.Readable(c.Reg.R[PC], w) {
		bus = c.Itcm
	}

	if w == 4 || c.Reg.R[PC]&2 == 0 {
		c.Cycles(c.Reg.R[PC], w, 0, true)
	}

	if c.PcPtr == nil {
		if w == 4 {
			c.Op[1] = bus.Read32(c.Reg.R[PC])
		} else {
			c.Op[1] = bus.Read16(c.Reg.R[PC])
		}
	} else {
		// 0xFFFF_FFFF uint32, 0xFFFF uint16
		mask := uint32(0xFFFF_FFFF >> ((w & 2) * 8))
		c.Op[1] = *(*uint32)(c.PcPtr) & mask
	}

	if w == 4 {
		c.DecodeArm(inst)
	} else {
		c.DecodeThumb(uint16(inst))
	}

	if c.Reload {
		c.ReloadPipe()
	} else {
		c.Reg.R[PC] += w
		if c.PcPtr != nil {
			c.PcPtr = unsafe.Add(c.PcPtr, w)
		}
	}

	//c.print()
	c.tick(max(c.InstCycles, c.DataCycles) + c.IdleCycles)

	c.InstCycles, c.DataCycles, c.IdleCycles = 0, 0, 0
}

func (c *Cpu) ReloadPipe() {
	w := uint32(4)
	if c.Reg.CPSR.T {
		w = 2
	}

	pc := c.Reg.R[PC] &^ (w - 1)

	bus := c.Mem
	if c.Itcm.Readable(pc, w) {
		bus = c.Itcm
		c.Cycles(pc, w, 0, true)
	} else {
		c.setCyclesPerInst(pc)

		c.Cycles(pc, w, 0, true) // pc + 0

		if w == 4 || pc&2 != 0 {
			c.Cycles(pc, w, 0, true) // pc + 2 or pc + 4
		}
	}

	c.PcPtr = bus.ReadPtr(pc)

	if c.PcPtr == nil {
		if w == 4 {
			c.Op[0] = bus.Read32(pc + 0)
			c.Op[1] = bus.Read32(pc + w)
		} else {
			c.Op[0] = bus.Read16(pc + 0)
			c.Op[1] = bus.Read16(pc + w)
		}
	} else {
		// 0xFFFF_FFFF uint32, 0xFFFF uint16
		mask := uint32(0xFFFF_FFFF >> ((w & 2) * 8))
		c.Op[0] = *(*uint32)(c.PcPtr) & mask
		c.PcPtr = unsafe.Add(c.PcPtr, w)
		c.Op[1] = *(*uint32)(c.PcPtr) & mask
		c.PcPtr = unsafe.Add(c.PcPtr, w)
	}

	c.Reg.R[PC] += w * 2
	c.Seq = arm7.SEQ
	c.Reload = false
}

func (c *Cpu) print() {
	fmt.Printf("PC %08X Diff %08d: %02d %02d %02d\n", c.Reg.R[15], c.Timestamp-debug.Vi64[0], c.InstCycles, c.DataCycles, c.IdleCycles)
	debug.Vi64[0] = c.Timestamp
}

func (c *Cpu) print2(inst uint32) {
	fmt.Printf("OP %08X\n", inst)
	if debug.V[0] > 10000 {
		os.Exit(0)
	} else {
		debug.V[0]++
	}
}
