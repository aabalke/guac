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

type InterruptStack struct {
    Gba *GBA
    Interrupts []Interrupt
    Skip bool
    Print bool
}

type Interrupt struct {
    Reg Reg
    ReturnAddr uint32
    IF uint32
}

func (s *InterruptStack) IsEmpty() bool {
    return len(s.Interrupts) == 0
}

func (s *InterruptStack) ReturnAddr() uint32 {
    if s.IsEmpty() { return 0 }
    return s.Interrupts[len(s.Interrupts) - 1].Reg.R[LR]
}

func (s *InterruptStack) print(exit bool) {

    if !(CURR_INST >= 18_000_000 && CURR_INST < 19_000_000) {
        return
    }

    if !s.Print {
        return
    }

    reg := &s.Gba.Cpu.Reg
    r := &s.Gba.Cpu.Reg.R

    a := "ENTER"
    if exit {
        a = "EXIT "
    }

    fmt.Printf("%s PC %08X LR %08X SP %08X CPSR %08X T %t MODE %02X OP %08X CURR %08d\n", a, r[PC], r[LR], r[SP], reg.CPSR, reg.CPSR.GetFlag(FLAG_T), reg.getMode(), s.Gba.Mem.Read32(r[PC]), CURR_INST)
    //fmt.Printf("SP ADDR %08X\n", r[SP])
    //fmt.Printf("LR %08X\n", r[LR])
    fmt.Printf("0x300_22DC %08X\n", s.Gba.Mem.Read32(r[2] + 0x1C))
    fmt.Printf("0x300_7FFC %08X\n", s.Gba.Mem.Read32(0x3007FFC))
    fmt.Printf("12 %08X\n", r[12])
    fmt.Printf("03 %08X\n", r[3])
    fmt.Printf("02 %08X\n", r[2])
    fmt.Printf("01 %08X\n", r[1])
    fmt.Printf("00 %08X\n", r[0])

    if exit {
        fmt.Printf("--------------------------------------------------------\n")
    }
}

func (s *InterruptStack) Execute() {
    // PUSH

    if s.Skip {
        return
    }

    mem := s.Gba.Mem
    reg := &s.Gba.Cpu.Reg
    r := &s.Gba.Cpu.Reg.R

    //if CURR_INST >= 17_985_169 && CURR_INST < 19_000_000 {
    //    fmt.Printf("USER ENTE PC %08X CURR %d\n", r[PC], CURR_INST)
    //}

    s.print(false)

    s.Gba.Mem.BIOS_MODE = BIOS_IRQ

    irqBank := BANK_ID[MODE_IRQ]

    curMode := uint32(reg.CPSR) & 0b11111
    curBank := BANK_ID[curMode]

    reg.SP[curBank] = r[SP]
    reg.LR[curBank] = r[LR]

    tmpCPSR := reg.SPSR[irqBank]
    reg.SPSR[irqBank] = reg.CPSR
    reg.CPSR = (tmpCPSR &^ 0b11111) | MODE_IRQ
    //reg.CPSR = (reg.CPSR &^ 0b11111) | MODE_IRQ

    r[SP] = reg.SP[irqBank]
    r[LR] = reg.LR[irqBank]

    reg.CPSR.SetFlag(FLAG_I, true) // true is disabled
    reg.CPSR.SetFlag(FLAG_T, false)

    //{
    //    for i := range 13 {

    //        ifBit := (s.Gba.Mem.Read16(0x400_0202) >> i) & 1 == 1
    //        ieBit := (s.Gba.Mem.Read16(0x400_0200) >> i) & 1 == 1

    //        if ifBit && ieBit {
    //            if i == 0 { // vblank breaks things for some reason
    //                break // do i need to check if irq enabled to check priority? IE??
    //            }
    //            mem.Write16(0x400_0202, 1 << i)

    //            //irq = uint32(i)
    //            break
    //        }
    //    }
    //}

    s.Interrupts = append(s.Interrupts, Interrupt{
        Reg: s.Gba.Cpu.Reg,
        ReturnAddr: r[PC],
        IF: s.Gba.Mem.Read16(0x400_0202),
    })

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

}

func (s *InterruptStack) Exit() {

    if s.Skip {
        return
    }


    if len(s.Interrupts) == 0 {
        panic(fmt.Sprintf("ERROR: A Interrupt Exit was called without an Interrupt PC %08X, CURR %d", s.Gba.Cpu.Reg.R[PC], CURR_INST))
    }

    mem := s.Gba.Mem
    reg := &s.Gba.Cpu.Reg
    r := &s.Gba.Cpu.Reg.R

    //if CURR_INST >= 18_000_000 && CURR_INST < 19_000_000 {
    //    fmt.Printf("USER EXIT PC %08X CURR %d\n", r[PC], CURR_INST)
    //}

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

    maxIdx := len(s.Interrupts)-1

    s.Gba.Cpu.Reg = s.Interrupts[maxIdx].Reg
    r[PC] = s.Interrupts[maxIdx].ReturnAddr

    s.Interrupts = s.Interrupts[:maxIdx]

    reg.SP[irqBank] = r[SP]
    reg.LR[irqBank] = r[LR]

    // DO NOT REMOVE TEMP
    tmpCPSR := reg.CPSR
    reg.CPSR = reg.SPSR[irqBank]
    reg.SPSR[irqBank] = tmpCPSR

    curMode := uint32(reg.CPSR) & 0b11111
    curBank := BANK_ID[curMode]

    r[LR] = reg.LR[curBank]
    r[SP] = reg.SP[curBank]

    s.Gba.Mem.BIOS_MODE = BIOS_IRQ_POST

    reg.CPSR.SetFlag(FLAG_I, false) // disable IRQ

    s.print(true)
}
