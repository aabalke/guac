package arm9

import (

	"github.com/aabalke/guac/emu/nds/cpu"
	"github.com/aabalke/guac/emu/nds/cpu/cp15"
	"github.com/aabalke/guac/emu/nds/mem/dma"
)

type Cpu struct {
	mem    cpu.MemoryInterface
	Irq    *cpu.Irq
	Reg    Reg
	Halted bool

    LowVector bool

    Cp15 *cp15.Cp15

    Dma [4]dma.DMA
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

	BIOS_STARTUP  = 0
	BIOS_SWI      = 1
	BIOS_IRQ      = 2
	BIOS_IRQ_POST = 3
)

var masks = [16]uint16{
	0:  0xF0F0,
	1:  0x0F0F,
	2:  0xCCCC,
	3:  0x3333,
	4:  0xFF00,
	5:  0x00FF,
	6:  0xAAAA,
	7:  0x5555,
	8:  0x0C0C,
	9:  0xF3F3,
	10: 0xAA55,
	11: 0x55AA,
	12: 0x0A05,
	13: 0xF5FA,
	14: 0xFFFF,
	15: 0x0000,
}

func (cpu *Cpu) CheckCond(cond uint32) bool {
	return (masks[cond] & (1 << (cpu.Reg.CPSR >> 28))) != 0
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

var BIOS_ADDR = map[uint32]uint32{
	BIOS_STARTUP:  0xE129F000,
	BIOS_SWI:      0xE3A02004,
	BIOS_IRQ:      0xE25EF004,
	BIOS_IRQ_POST: 0xE55EC002,
}

func NewCpu(m cpu.MemoryInterface, irq *cpu.Irq, cp15 *cp15.Cp15) *Cpu {

	c := &Cpu{
		mem: m,
		Irq: irq,
        Cp15: cp15,
	}

    // skip bios
    c.Irq.IME = true
    // IrqIpcRecvFifo, IrqTimers, IrqVBlank
    //c.Irq.IE |= 1 << 0
    //c.Irq.IE |= 1 << 3
    //c.Irq.IE |= 1 << 17

	return c
}

func (c *Cpu) Execute() (int, bool) {
	if c.Reg.IsThumb {
		return c.DecodeTHUMB(), true
	}

	return c.DecodeARM()
}

type Reg struct {
	R    [16]uint32
	SP   [6]uint32
	LR   [6]uint32
	FIQ  [5]uint32 // r8 - r12
	USR  [5]uint32 // r8 - r12 // tmp to restore after FIQ
	CPSR Cond
	SPSR [6]Cond

	IsThumb bool
}

type Cond uint32

func (c *Cond) GetFlag(flag uint32) bool {
	return (uint32(*c)>>flag)&0b1 == 0b1
}

func (c *Cond) SetThumb(value bool, cpu *Cpu) {
	cpu.Reg.IsThumb = value
	c.SetFlag(FLAG_T, value)
}

func (c *Cond) SetFlag(flag uint32, value bool) {

	if value {
		*c |= (0b1 << flag)
		return
	}

	*c &^= (0b1 << flag)
}

func (c *Cond) SetField(loBit uint32, value uint32) {
	mask := 0b1111_1111 << loBit
	*c &^= Cond(mask)
	value <<= loBit
	*c |= Cond(value)
}

func (c *Cond) SetMode(mode uint32) {
	*c = Cond((uint32(*c) &^ 0b11111) | mode)
}

func (r *Reg) getMode() uint32 {
	return uint32(r.CPSR) & 0b11111
}

func (r *Reg) setMode(prev, curr uint32) {

	if prev == curr {
		return
	}

	r.CPSR.SetMode(curr)

	//r._setMode(prev, curr)
	//}
	//
	//func (r *Reg) _setMode(prev, curr uint32) {

	if BANK_ID[prev] == BANK_ID[curr] {
		return
	}

	r.switchRegisterBanks(prev, curr)
}

func (r *Reg) switchRegisterBanks(prev, curr uint32) {

	//if BANK_ID[prev] == BANK_ID[curr] {
	//    return
	//}

	if prev != MODE_FIQ {
		for i := range 5 {
			r.USR[i] = r.R[8+i]
		}
	}

	r.SP[BANK_ID[prev]] = r.R[SP]
	r.LR[BANK_ID[prev]] = r.R[LR]

	if prev == MODE_FIQ {
		for i := range 5 {
			r.FIQ[i] = r.R[8+i]
		}
	}

	if curr != MODE_FIQ {
		for i := range 5 {
			r.R[8+i] = r.USR[i]
		}
	}

	r.R[SP] = r.SP[BANK_ID[curr]]
	r.R[LR] = r.LR[BANK_ID[curr]]

	if curr == MODE_FIQ {
		for i := range 5 {
			r.R[8+i] = r.FIQ[i]
		}
	}
}

func (cpu *Cpu) toggleThumb() {

	reg := &cpu.Reg

	newFlag := reg.R[PC]&1 > 0

	reg.CPSR.SetThumb(newFlag, cpu)

	if newFlag {
		reg.R[PC] &^= 1
		return
	}

	reg.R[PC] &^= 3
}

func (cpu *Cpu) CheckIrq() {

	if interrupts := cpu.Irq.IE&cpu.Irq.IF != 0; !interrupts {
        return
    }

    cpu.Halted = false

	if !cpu.Reg.CPSR.GetFlag(FLAG_I) && cpu.Irq.IME {
		cpu.exception(VEC_IRQ, MODE_IRQ)
	}
}
