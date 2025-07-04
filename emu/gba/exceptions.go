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

var IN_INTERRUPT = false

func (gba *GBA) exception(addr uint32, mode uint32) {

    IN_INTERRUPT = true

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

    //fmt.Printf("LR %08X: LR0 %08X, LR1 %08X, LR2 %08X, LR3 %08X, LR4 %08X, LR5 %08X\n",
    //reg.R[LR],
    //reg.LR[0],
    //reg.LR[1],
    //reg.LR[2], //
    //reg.LR[3],
    //reg.LR[4],
    //reg.LR[5])

    //fmt.Printf("SPSR0 %08X, SPSR1 %08X, SPSR2 %08X, SPSR3 %08X SPSR4 %08X, SPSR5 %08X\n",
    //reg.SPSR[0],
    //reg.SPSR[1],
    //reg.SPSR[2], //
    //reg.SPSR[3],
    //reg.SPSR[4],
    //reg.SPSR[5])

    i := BANK_ID[MODE_IRQ]
    reg.CPSR = reg.SPSR[i]

    reg.LR[i] = r[LR]
    reg.SP[i] = r[SP]

    c := BANK_ID[cpu.Reg.getMode()]
    r[SP] = reg.SP[c]
    r[LR] = reg.LR[c]

    r[PC] = reg.LR[i] - 4

    //fmt.Printf("EXIT PC %08X OP %08X CPSR %08X LR %08X SP %08X T %t\n", r[PC], cpu.Gba.Mem.Read32(r[PC]), cpu.Reg.CPSR, r[LR], r[SP], cpu.Reg.CPSR.GetFlag(FLAG_T))

}
