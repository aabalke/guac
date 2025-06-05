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
    SKIP_INTERRUPT = false
    PRINT_INTERRUPT = false
    INTERRUPT_CAUSE = ""
    IRQ = Interrupt{}
)

type Interrupt struct {
    LR, R0, R1, R2, R3, R12 uint32
}

func (gba *GBA) printInterrupt(exit bool) {

    if CURR_INST <= 379180 {
        return
    }

    if !PRINT_INTERRUPT {
        return
    }

    reg := &gba.Cpu.Reg
    r := &gba.Cpu.Reg.R

    s := "ENTER"
    if exit {
        s = "EXIT "
    }

    fmt.Printf("%s CAUSE %s PC %08X LR %08X SP %08X CPSR %08X T %t MODE %02X OP %08X CURR %08d\n", s, INTERRUPT_CAUSE, r[PC], r[LR], r[SP], reg.CPSR, reg.CPSR.GetFlag(FLAG_T), reg.getMode(), gba.Mem.Read32(r[PC]), CURR_INST)
    fmt.Printf("SP ADDR %08X\n", r[SP])
    fmt.Printf("LR %08X\n", r[LR])
    fmt.Printf("12 %08X\n", r[12])
    fmt.Printf("03 %08X\n", r[3])
    fmt.Printf("02 %08X\n", r[2])
    fmt.Printf("01 %08X\n", r[1])
    fmt.Printf("00 %08X\n", r[0])

    if exit {
        fmt.Printf("--------------------------------------------------------\n")
    }
}

func (gba *GBA) handleInterrupt() {

    if SKIP_INTERRUPT == true {
        return
    }

    gba.Mem.BIOS_MODE = BIOS_IRQ

    mem := gba.Mem
    reg := &gba.Cpu.Reg
    r := &gba.Cpu.Reg.R

    irqBank := BANK_ID[MODE_IRQ]

    gba.printInterrupt(false)

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

    if reg.CPSR.GetFlag(FLAG_T) {
        reg.LR[irqBank] = r[PC] + 2
        r[LR] = r[PC] + 2
    } else {
        reg.LR[irqBank] = r[PC] + 4
        r[LR] = r[PC] + 4
    }

    reg.CPSR.SetFlag(FLAG_T, false)

    IRQ = Interrupt{
        LR: r[LR],
        R0: r[0],
        R1: r[1],
        R2: r[2],
        R3: r[3],
        R12: r[12],
    }

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

    //fmt.Printf("PC IN IRQ %08X IRQ LR %08X \n", r[PC], reg.LR[irqBank])

    return
}

func (gba *GBA) handleInterruptExit() {

    if SKIP_INTERRUPT == true {
        return
    }

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

    r[0] = IRQ.R0
    r[1] = IRQ.R1
    r[2] = IRQ.R2
    r[3] = IRQ.R3
    r[12] = IRQ.R12
    r[LR] = IRQ.LR

    //r[PC] = reg.LR[irqBank]
    if reg.CPSR.GetFlag(FLAG_T) {
        r[PC] = reg.LR[irqBank] - 2
    } else {
        r[PC] = reg.LR[irqBank] - 4
    }

    reg.SP[irqBank] = r[SP]
    reg.LR[irqBank] = r[LR]

    tmpCPSR := reg.CPSR
    reg.CPSR = reg.SPSR[irqBank]
    reg.SPSR[irqBank] = tmpCPSR

    curMode := uint32(reg.CPSR) & 0b11111
    curBank := BANK_ID[curMode]

    r[LR] = reg.LR[curBank]
    r[SP] = reg.SP[curBank]

    reg.CPSR &^= 0b1000_0000 // disable IRQ

    gba.Mem.BIOS_MODE = BIOS_IRQ_POST

    gba.printInterrupt(true)
}
