package gba

import (
	"fmt"
)

const (
	resetVec         uint32 = 0x00
	undVec           uint32 = 0x04
	swiVec           uint32 = 0x08
	prefetchAbortVec uint32 = 0xc
	dataAbortVec     uint32 = 0x10
	addr26BitVec     uint32 = 0x14
	irqVec           uint32 = 0x18
	fiqVec           uint32 = 0x1c
)

var (
    _ = fmt.Sprintf("")
)

func (gba *GBA) handleInterrupt() {

    mem := gba.Mem
    reg := &gba.Cpu.Reg
    r := &gba.Cpu.Reg.R

    irqBank := BANK_ID[MODE_IRQ]
    //fmt.Printf("ENTR SP %08X PC %08X THUMB %t MODE %02X CURR %08d LR %08X\n\n", r[SP], r[PC], reg.CPSR.GetFlag(FLAG_T), reg.getMode(), CURR_INST, r[LR])
    //fmt.Printf("SP ADDR %08X\n", r[SP])
    //fmt.Printf("LR %08X\n", r[LR])
    //fmt.Printf("12 %08X\n", r[12])
    //fmt.Printf("03 %08X\n", r[3])
    //fmt.Printf("02 %08X\n", r[2])
    //fmt.Printf("01 %08X\n", r[1])
    //fmt.Printf("00 %08X\n", r[0])

    curMode := uint32(reg.CPSR) & 0b11111
    curBank := BANK_ID[curMode]

    reg.SP[curBank] = r[SP]
    reg.LR[curBank] = r[LR]

    tmpCPSR := reg.SPSR[irqBank]
    reg.SPSR[irqBank] = reg.CPSR
    reg.CPSR = (tmpCPSR &^ 0b11111) | MODE_IRQ
    //reg.CPSR = (reg.CPSR &^ 0b11111) | MODE_IRQ

    // will need logic for FIQ at some point


    r[SP] = reg.SP[irqBank]
    r[LR] = reg.LR[irqBank]

    reg.CPSR |= 0b1000_0000 // disable IRQ

    reg.LR[irqBank] = r[PC]
    r[LR] = r[PC]

    reg.CPSR.SetFlag(FLAG_T, false)

    r[SP] -= 4
    mem.Write32(r[SP], r[LR])
    r[SP] -= 4
    mem.Write32(r[SP], r[12])
    r[SP] -= 4
    mem.Write32(r[SP], r[3])
    r[SP] -= 4
    mem.Write32(r[SP], r[2])
    r[SP] -= 4
    mem.Write32(r[SP], r[1])
    r[SP] -= 4
    mem.Write32(r[SP], r[0])

    userAddr := mem.Read32(0x03007FFC)
    r[PC] = userAddr

    //fmt.Printf("PC IN IRQ %08X\n", r[PC])

    return
}

func (gba *GBA) handleInterruptExit() {

    mem := gba.Mem
    reg := &gba.Cpu.Reg
    r := &gba.Cpu.Reg.R

    irqBank := BANK_ID[MODE_IRQ]

    r[0] = mem.Read32(r[SP])
    r[SP] += 4
    r[1] = mem.Read32(r[SP])
    r[SP] += 4
    r[2] = mem.Read32(r[SP])
    r[SP] += 4
    r[3] = mem.Read32(r[SP])
    r[SP] += 4
    r[12] = mem.Read32(r[SP])
    r[SP] += 4
    r[LR] = mem.Read32(r[SP])
    r[SP] += 4

    reg.SP[irqBank] = r[SP]
    reg.LR[irqBank] = r[LR]

    tmpCPSR := reg.CPSR
    reg.CPSR = reg.SPSR[irqBank]
    reg.SPSR[irqBank] = tmpCPSR

    curMode := uint32(reg.CPSR) & 0b11111
    curBank := BANK_ID[curMode]

    r[PC] = reg.LR[irqBank]
    r[LR] = reg.LR[curBank]
    r[SP] = reg.SP[curBank]

    //fmt.Printf("SP ADDR %08X\n", r[SP])
    //fmt.Printf("LR %08X\n", r[LR])
    //fmt.Printf("12 %08X\n", r[12])
    //fmt.Printf("03 %08X\n", r[3])
    //fmt.Printf("02 %08X\n", r[2])
    //fmt.Printf("01 %08X\n", r[1])
    //fmt.Printf("00 %08X\n", r[0])
    //fmt.Printf("EXIT SP %08X PC %08X THUMB %t MODE %02X CURR %08d LR %08X\n\n", r[SP], r[PC], reg.CPSR.GetFlag(FLAG_T), reg.getMode(), CURR_INST, r[LR])
    //fmt.Printf("-----------------------------------------------------\n")
}
