package gba

import (
    //"os"
)

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

func (gba *GBA) exception(addr uint32, mode uint32) {

	reg := &gba.Cpu.Reg
	r := &gba.Cpu.Reg.R

    c := BANK_ID[reg.getMode()]
    i := BANK_ID[mode]

    reg.SP[c] = r[SP]
    reg.LR[c] = r[LR]
    r[SP] = reg.SP[i]
    r[LR] = reg.LR[i]
    reg.SPSR[i] = reg.CPSR

    r[LR] = r[PC] + 4
    reg.LR[i] = r[PC] + 4

    reg.CPSR.SetMode(mode)
    reg.CPSR.SetFlag(FLAG_T, false)
    reg.CPSR.SetFlag(FLAG_I, true)

    r[PC] = addr
    return
}

func (gba *GBA) IrqExit() {

    cpu := gba.Cpu
    reg := &cpu.Reg
    r := &cpu.Reg.R

    r[PC] = r[LR] - 4

    i := BANK_ID[MODE_IRQ]
    reg.CPSR = reg.SPSR[i]
    c := BANK_ID[cpu.Reg.getMode()]

    reg.LR[i] = r[LR]
    reg.SP[i] = r[SP]
    r[SP] = reg.SP[c]
    r[LR] = reg.LR[c]

}
