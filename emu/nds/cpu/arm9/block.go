package arm9

import (
	"unsafe"

	"github.com/aabalke/guac/emu/nds/utils"
)

const (
    IA = 0
    IB = 1
    DB = 2
    DA = 3
)

//type Block struct {
//	Opcode, Rn, Rlist             uint32
//    Pre, Up bool
//	PSR, Writeback bool
//    Method uint8
//    ForceUser bool
//    Ref [16]*uint32
//}
//
//func (c *Cpu) Block(opcode uint32) {
//
//	block := &Block{
//		Opcode:      opcode,
//		Pre:         utils.BitEnabled(opcode, 24),
//		Up:          utils.BitEnabled(opcode, 23),
//		PSR:         utils.BitEnabled(opcode, 22),
//		Writeback:   utils.BitEnabled(opcode, 21),
//		Rn:          utils.GetVarData(opcode, 16, 19),
//		Rlist:       utils.GetVarData(opcode, 0, 15),
//	}
//
//    load := (opcode >> 20) & 1 != 0
//
//	mode := c.Reg.CPSR.Mode
//    block.ForceUser = (
//        block.PSR &&
//        mode != MODE_USR && 
//        (!load || (!utils.BitEnabled(opcode, 15))))
//
//    block.Ref = [16]*uint32{
//        &c.Reg.R[0],
//        &c.Reg.R[1],
//        &c.Reg.R[2],
//        &c.Reg.R[3],
//        &c.Reg.R[4],
//        &c.Reg.R[5],
//        &c.Reg.R[6],
//        &c.Reg.R[7],
//        &c.Reg.R[8],
//        &c.Reg.R[9],
//        &c.Reg.R[10],
//        &c.Reg.R[11],
//        &c.Reg.R[12],
//        &c.Reg.R[13],
//        &c.Reg.R[14],
//        &c.Reg.R[15],
//    }
//
//    if block.ForceUser {
//        block.Ref[13] = &c.Reg.SP[BANK_ID[MODE_USR]]
//        block.Ref[14] = &c.Reg.LR[BANK_ID[MODE_USR]]
//    }
//
//    switch {
//    case block.Pre && block.Up:
//        block.Method = IB
//    case !block.Pre && block.Up:
//        block.Method = IA
//    case block.Pre && !block.Up:
//        block.Method = DB
//    case !block.Pre && !block.Up:
//        block.Method = DA
//    }
//
//    //if load && block.Method == IA && opcode & 0x8000 != 0 {
//    if (
//        load && block.Method == IA && 
//        !block.ForceUser && opcode & 0x8000 != 0 ){
//
//            //fmt.Printf("PC %08X %08X\n", c.Reg.R[15], opcode)
//        c.Block2(opcode)
//        return
//    }
//
//
//	if load {
//		c.ldm(block)
//        return
//    }
//
//    c.stm(block)
//}
//
//func (cpu *Cpu) ldm(block *Block) {
//
//	r := &cpu.Reg.R
//
//	addr := r[block.Rn] &^ 0b11
//	wbValue := r[block.Rn]
//    keepPC := false
//
//	if block.Rlist == 0 {
//
//		cpu.Reg.R[PC] += 4
//
//		if block.Up {
//			r[block.Rn] += 0x40
//			return
//		}
//
//		r[block.Rn] -= 0x40
//		return
//	}
//
//	//mode := cpu.Reg.getMode()
//	//if forceUser := block.PSR; forceUser && mode != MODE_USR {
//	//	cpu.Reg.setMode(mode, MODE_USR)
//	//}
//
//	regCount := utils.CountBits(block.Rlist)
//
//	if rnIncluded := (block.Rlist>>block.Rn)&1 == 1; rnIncluded {
//        isLast := (block.Rlist < (1 << (block.Rn + 1)))
//        isOnly := regCount == 1 && rnIncluded
//        block.Writeback = (!isLast || isOnly) && block.Writeback
//
//        if !block.Writeback || block.PSR {
//            regCount--
//        }
//    }
//
//    switch block.Method {
//    case IB:
//        keepPC = ldmIB(cpu, addr, block.Rlist, block.Ref)
//    case IA:
//        keepPC = ldmIA(cpu, addr, block.Rlist, block.Ref)
//    case DB:
//        keepPC = ldmDB(cpu, addr, block.Rlist, block.Ref)
//    case DA:
//        keepPC = ldmDA(cpu, addr, block.Rlist, block.Ref)
//    }
//
//	//if forceUser := block.PSR; forceUser && mode != MODE_USR {
//	//	curr := cpu.Reg.getMode()
//	//	cpu.Reg.setMode(curr, mode)
//	//}
//
//	if block.Writeback {
//
//        if block.Up {
//            wbValue += regCount * 4
//
//        } else {
//            wbValue -= regCount * 4
//        }
//
//		r[block.Rn] = wbValue
//	}
//
//    if !keepPC {
//		cpu.Reg.R[PC] += 4
//        return
//    }
//
//	if utils.BitEnabled(block.Opcode, 15) && block.PSR {
//
//        curr := cpu.Reg.CPSR.Mode
//        spsr := cpu.Reg.SPSR[BANK_ID[curr]]
//
//        reg := &cpu.Reg
//
//        // I think this is necessary for irq exits - BoxTest FAILS, otherwise I dont think this matters
//        //if irqExit := curr == MODE_IRQ; irqExit {
//        //    r[PC] = r[LR]
//        //    //r[PC] += 4
//        //}
//
//        //cpu.Reg.setMode(cpu.Reg.getMode(), uint32(cpu.Reg.SPSR[BANK_ID[cpu.Reg.getMode()]]) & 0x1F)
//
//        next := spsr.Mode
//        reg.CPSR = spsr
//
//        if curr == MODE_USR {
//            panic("USER MODE LDM PC CHANGE")
//        }
//
//        if curr != MODE_FIQ {
//            for i := range 5 {
//                reg.USR[i] = r[8+i]
//            }
//        }
//
//        reg.SP[BANK_ID[curr]] = r[SP]
//        reg.LR[BANK_ID[curr]] = r[LR]
//
//        if curr == MODE_FIQ {
//            for i := range 5 {
//                reg.FIQ[i] = r[8+i]
//            }
//        }
//
//        if next != MODE_FIQ {
//            for i := range 5 {
//                r[8+i] = reg.USR[i]
//            }
//        }
//
//        r[SP] = reg.SP[BANK_ID[next]]
//        r[LR] = reg.LR[BANK_ID[next]]
//
//        if next == MODE_FIQ {
//            for i := range 5 {
//                r[8+i] = reg.FIQ[i]
//            }
//        }
//
//        return
//	}
//
//
//    cpu.toggleThumb()
//}
//
//func ldmIB(cpu *Cpu, addr, rlist uint32, ref [16]*uint32) bool {
//
//    keepPc := false
//
//    for reg := uint32(0); reg < 16; reg++ {
//        if disabled := !utils.BitEnabled(rlist, uint8(reg)); disabled {
//            continue
//        }
//
//        addr += 4
//
//        *ref[reg] = cpu.mem.Read32(addr, true)
//
//        if reg == PC {
//            keepPc = true
//        }
//    }
//
//    return keepPc
//}
//
//func ldmIA(cpu *Cpu, addr, rlist uint32, ref [16]*uint32) bool {
//
//    keepPc := false
//
//    for reg := uint32(0); reg < 16; reg++ {
//        if disabled := !utils.BitEnabled(rlist, uint8(reg)); disabled {
//            continue
//        }
//
//        *ref[reg] = cpu.mem.Read32(addr, true)
//
//        if reg == PC {
//            keepPc = true
//        }
//
//        addr += 4
//    }
//
//    return keepPc
//}
//
//func ldmDB(cpu *Cpu, addr, rlist uint32, ref [16]*uint32) bool {
//
//    keepPc := false
//
//    for reg := 15; reg >= 0; reg-- {
//        if disabled := !utils.BitEnabled(rlist, uint8(reg)); disabled {
//            continue
//        }
//
//        addr -= 4
//
//        *ref[reg] = cpu.mem.Read32(addr, true)
//
//        if reg == PC {
//            keepPc = true
//        }
//    }
//
//    return keepPc
//}
//
//func ldmDA(cpu *Cpu, addr, rlist uint32, ref [16]*uint32) bool {
//
//    keepPc := false
//
//    for reg := 15; reg >= 0; reg-- {
//        if disabled := !utils.BitEnabled(rlist, uint8(reg)); disabled {
//            continue
//        }
//
//        *ref[reg] = cpu.mem.Read32(addr, true)
//        addr -= 4
//
//        if reg == PC {
//            keepPc = true
//        }
//    }
//
//    return keepPc
//}
//
//// STM STM STM STM //
//
//func (cpu *Cpu) stm(block *Block) {
//
//	r := &cpu.Reg.R
//
//	addr := r[block.Rn] &^ 0b11
//	wbValue := r[block.Rn]
//
//	//mode := cpu.Reg.getMode()
//	//if forceUser := block.PSR; forceUser && mode != MODE_USR {
//	//	cpu.Reg.setMode(mode, MODE_USR)
//	//}
//
//	if block.Rlist == 0 {
//
//        r[PC] += 4
//
//        if block.Up {
//			r[block.Rn] += 0x40
//            return
//        }
//
//        r[block.Rn] -= 0x40
//
//		return
//	}
//
//    switch block.Method {
//    case IA:
//        stmIA(cpu, addr, block.Rlist, block.Rn, block.Ref)
//    case IB:
//        stmIB(cpu, addr, block.Rlist, block.Rn, block.Ref)
//    case DB:
//        stmDB(cpu, addr, block.Rlist, block.Rn, block.Ref)
//    case DA:
//        stmDA(cpu, addr, block.Rlist, block.Rn, block.Ref)
//    }
//
//	//if forceUser := block.PSR; forceUser && mode != MODE_USR {
//	//	curr := cpu.Reg.getMode()
//	//	cpu.Reg.setMode(curr, mode)
//	//}
//
//	if !block.Writeback {
//		r[block.Rn] = wbValue
//    }
//
//    r[PC] += 4
//}
//
//func stmIB(cpu *Cpu, addr, rlist, rn uint32, ref [16]*uint32) {
//
//	r := &cpu.Reg.R
//    rnValue := *ref[rn]
//
//    for reg := uint32(0); reg < 16; reg++ {
//        if disabled := !utils.BitEnabled(rlist, uint8(reg)); disabled {
//            continue
//        }
//
//        r[rn] += 4
//        addr += 4
//
//        switch reg {
//        case rn:
//            cpu.mem.Write32(addr, rnValue, true)
//        case PC:
//            cpu.mem.Write32(addr, *ref[reg]+12, true)
//        default:
//            cpu.mem.Write32(addr, *ref[reg], true)
//        }
//    }
//}
//
//func stmIA(cpu *Cpu, addr, rlist, rn uint32, ref [16]*uint32) {
//
//	r := &cpu.Reg.R
//    rnValue := *ref[rn]
//
//    for reg := uint32(0); reg < 16; reg++ {
//        if disabled := !utils.BitEnabled(rlist, uint8(reg)); disabled {
//            continue
//        }
//
//        switch reg {
//        case rn:
//            cpu.mem.Write32(addr, rnValue, true)
//        case PC:
//            cpu.mem.Write32(addr, *ref[reg]+12, true)
//        default:
//            cpu.mem.Write32(addr, *ref[reg], true)
//        }
//        r[rn] += 4
//        addr += 4
//    }
//}
//
//func stmDB(cpu *Cpu, addr, rlist, rn uint32, ref [16]*uint32) {
//
//	r := &cpu.Reg.R
//    rnValue := *ref[rn]
//
//    for reg := 15; reg >= 0; reg-- {
//        if disabled := !utils.BitEnabled(rlist, uint8(reg)); disabled {
//            continue
//        }
//
//        r[rn] -= 4
//        addr -= 4
//
//        switch uint32(reg) {
//        case rn:
//            cpu.mem.Write32(addr, rnValue, true)
//        case PC:
//            cpu.mem.Write32(addr, *ref[reg]+12, true)
//
//        default:
//
//            cpu.mem.Write32(addr, *ref[reg], true)
//        }
//    }
//}
//
//func stmDA(cpu *Cpu, addr, rlist, rn uint32, ref [16]*uint32) {
//
//	r := &cpu.Reg.R
//    rnValue := *ref[rn]
//
//    for reg := 15; reg >= 0; reg-- {
//        if disabled := !utils.BitEnabled(rlist, uint8(reg)); disabled {
//            continue
//        }
//
//        switch uint32(reg) {
//        case rn:
//            cpu.mem.Write32(addr, rnValue, true)
//        case PC:
//            cpu.mem.Write32(addr, *ref[reg]+12, true)
//        default:
//            cpu.mem.Write32(addr, *ref[reg], true)
//        }
//        r[rn] -= 4
//        addr -= 4
//    }
//}

func (c *Cpu) Block(opcode uint32) {

    r := &c.Reg.R

    pcIncluded := opcode & 0x8000 != 0
    pre   := (opcode >> 24) & 1 != 0
    up    := (opcode >> 23) & 1 != 0
    psr   := (opcode >> 22) & 1 != 0
    wb    := (opcode >> 21) & 1 != 0
    load  := (opcode >> 20) & 1 != 0
    rn    := (opcode >> 16) & 0xF
    rlist := opcode & 0xFFFF
    forceUser := psr && (c.Reg.CPSR.Mode != MODE_USR) && (!load || !pcIncluded)

    //compare := !pcIncluded && !load
    //if compare {
    //    c.Jit.TestInst(opcode, c.Jit.emitBlock)
    //    r[15] += 4
    //    return
    //}

    if rlist == 0 {

        r[PC] += 4

        if up {
            r[rn] += 0x40
            return
        }

        r[rn] -= 0x40
        return
    }

    addr := r[rn] &^ 0b11
    regCount := utils.CountBits(rlist)

    wbValue := r[rn]
    if up {
        wbValue += regCount * 4
    } else {
        wbValue -= regCount * 4
    }

    rnRef := &c.Reg.R[rn]
    if forceUser && rn == 13 {
        rnRef = &c.Reg.SP[BANK_ID[MODE_USR]]
    }
    if forceUser && rn == 14 {
        rnRef = &c.Reg.LR[BANK_ID[MODE_USR]]
    }

    rnv := *rnRef

    reg := uint32(0)
    if !up {
        reg = 15
    }

    p, ok := c.mem.ReadPtr(addr, true)
    if !ok {
        p = nil
    }

    for range 16 {

        if disabled := (rlist >> reg) & 1 == 0; disabled {
            if up { reg++ } else { reg-- }
            continue
        }

        ref := &c.Reg.R[reg]
        if forceUser && reg == 13 {
            ref = &c.Reg.SP[BANK_ID[MODE_USR]]
        }
        if forceUser && reg == 14 {
            ref = &c.Reg.LR[BANK_ID[MODE_USR]]
        }

        if pre {
            if up {
                addr += 4
                if p != nil { p = unsafe.Add(p, 4) }
            } else {
                addr -= 4
                if p != nil { p = unsafe.Add(p, -4) }
            }
        }

        if load {

            if p == nil {
                *ref = c.mem.Read32(addr, true)
            } else {
                *ref = *(*uint32)(p)
            }

        } else {

            if p == nil {
                switch reg {
                    case rn: c.mem.Write32(addr, rnv, true)
                    case PC: c.mem.Write32(addr, *ref+12, true)
                    default: c.mem.Write32(addr, *ref, true)
                }
            } else {
                switch reg {
                    case rn: *(*uint32)(p) = rnv
                    case PC: *(*uint32)(p) = *ref+12
                    default: *(*uint32)(p) = *ref
                }
            }
        }

        if !pre {
            if up {
                addr += 4
                if p != nil { p = unsafe.Add(p, 4) }
            } else {
                addr -= 4
                if p != nil { p = unsafe.Add(p, -4) }
            }
        }

        if up {
            reg++
        } else {
            reg--
        }
    }

    if !load {
        if wb {
            r[rn] = wbValue
        }

        r[PC] += 4
        return
    }

    if wb {
        if rnIncluded := (rlist>>rn) & 1 == 1; rnIncluded {
            isLast := (rlist < (1 << (rn + 1)))
            isOnly := regCount == 1
            if !isLast || isOnly {
                r[rn] = wbValue
            }
        } else {
            r[rn] = wbValue
        }
    }

    if !pcIncluded {
        r[PC] += 4
        return
    }

    if psr {
        c.ldmModeSwitch()
        return
    }

    c.toggleThumb()
}

func (cpu *Cpu) ldmModeSwitch() {

    r := &cpu.Reg.R
    curr := cpu.Reg.CPSR.Mode
    spsr := cpu.Reg.SPSR[BANK_ID[curr]]

    reg := &cpu.Reg

    next := spsr.Mode
    reg.CPSR = spsr

    if curr == MODE_USR {
        panic("USER MODE LDM PC CHANGE")
    }

    if curr != MODE_FIQ {
        for i := range 5 {
            reg.USR[i] = r[8+i]
        }
    }

    reg.SP[BANK_ID[curr]] = r[SP]
    reg.LR[BANK_ID[curr]] = r[LR]

    if curr == MODE_FIQ {
        for i := range 5 {
            reg.FIQ[i] = r[8+i]
        }
    }

    if next != MODE_FIQ {
        for i := range 5 {
            r[8+i] = reg.USR[i]
        }
    }

    r[SP] = reg.SP[BANK_ID[next]]
    r[LR] = reg.LR[BANK_ID[next]]

    if next == MODE_FIQ {
        for i := range 5 {
            r[8+i] = reg.FIQ[i]
        }
    }
}
