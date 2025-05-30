package gba

import (
	"fmt"
)

var (
    SAVED_REGS = [6]uint32{}
    _ = fmt.Sprintf("")

    IRQ_RETURN_THUMB = false
    IRQ_SP = uint32(0)
    IRQ_LR = uint32(0)
    IRQ_SPSR = uint32(0)
)

func (gba *GBA) handleInterrupt() {

    //return

    mem := gba.Mem
    reg := &gba.Cpu.Reg
    r := &gba.Cpu.Reg.R

    fmt.Printf("ENTER SP %08X PC %08X\n\n", r[SP], r[PC])

    gba.Debugger.print(0)
    
    //currMode := uint32(reg.CPSR) & 0b11111
    //curBank := BANK_ID[currMode]

    // avoid multiple Interrupt handling per instruction
    mem.Write16(0x400_0208, 0)

    // CPU
    irqBank := BANK_ID[MODE_IRQ]
    IRQ_SPSR = uint32(reg.CPSR)

    reg.CPSR = (reg.CPSR &^ 0b11111) | MODE_IRQ

    // will need logic for FIQ at some point
    tmpSP := r[SP]
    tmpLR := r[LR]
    r[SP] = reg.SP[irqBank]
    r[LR] = reg.LR[irqBank]
    reg.SP[irqBank] = tmpSP
    reg.LR[irqBank] = tmpLR
    IRQ_LR = tmpLR
    IRQ_SP = tmpSP

    reg.SPSR[irqBank] = reg.CPSR

    reg.CPSR |= 0b1000_0000 // disable IRQ

    thumb := reg.CPSR.GetFlag(FLAG_T)
    if thumb {
        reg.LR[irqBank] = r[PC] + 2
        r[LR] = r[PC] + 2
        IRQ_RETURN_THUMB = true
    } else {
        reg.LR[irqBank] = r[PC] + 4
        r[LR] = r[PC] + 4
        IRQ_RETURN_THUMB = false
    }

    reg.CPSR.SetFlag(FLAG_T, false)

    // BIOS
    stackAddr := r[SP]
    mem.Write32(stackAddr - 4, r[LR])
    mem.Write32(stackAddr - 8, r[12])
    mem.Write32(stackAddr - 12, r[3])
    mem.Write32(stackAddr - 16, r[2])
    mem.Write32(stackAddr - 20, r[1])
    mem.Write32(stackAddr - 24, r[0])
    r[SP] -= 24

    // not sure one this, may also not matter
    //r[LR] = 0x12C
    //reg.LR[irqBank] = 0x12C

    //fmt.Printf("THUMB %t\n\n", thumb)

    userAddr := mem.Read32(0x03007FFC)

    fmt.Printf("USER ADDR %08X THUMB %t\n", userAddr, thumb)

    //r[PC] = userAddr &^ 11

    r[PC] = userAddr - 2 // -2 is temp, im not sure of a better way

    return
}

func (gba *GBA) handleInterruptExit() {

    //return

    mem := gba.Mem
    reg := &gba.Cpu.Reg
    r := &gba.Cpu.Reg.R

    //mem.Write16(0x400_0208, 1)
    //irqBank := BANK_ID[MODE_IRQ]
    //curBank := BANK_ID[reg.getMode()]

    //panic("INTERRUPT EXIT")
    stackAddr := r[SP]
    r[0] = mem.Read32(stackAddr + 0)
    r[1] = mem.Read32(stackAddr + 4)
    r[2] = mem.Read32(stackAddr + 8)
    r[3] = mem.Read32(stackAddr + 12)
    r[12] = mem.Read32(stackAddr + 16)
    r[LR] = mem.Read32(stackAddr + 20)

    r[SP] = IRQ_SP

    if IRQ_RETURN_THUMB {
        r[PC] = r[LR] - 2
    } else {
        r[PC] = r[LR] - 4
    }

    r[LR] = IRQ_LR

    //IRQ_RETURN_THUMB = false

    //reg.CPSR = reg.SPSR[curBank]
    reg.CPSR = Cond(IRQ_SPSR)
    fmt.Printf("EXIT SP %08X PC %08X THUMB %t\n\n", r[SP], r[PC], reg.CPSR.GetFlag(FLAG_T))
    //gba.Debugger.print(1)
}
