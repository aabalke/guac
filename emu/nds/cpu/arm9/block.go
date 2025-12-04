package arm9

import (
	"fmt"

	"github.com/aabalke/guac/emu/nds/utils"
)

var _ = fmt.Sprint

const (
    IA = 0
    IB = 1
    DB = 2
    DA = 3
)

//func (c *Cpu) Block(opcode uint32) {
//
//
//	r := &c.Reg.R
//
//    rlist := opcode & 0xFFFF
//    up := (opcode >> 23) & 1 != 0
//    rn := (opcode >> 16) & 0xF
//
//
//	if rlist == 0 {
//
//		r[PC] += 4
//
//		if up {
//			r[rn] += 0x40
//			return
//		}
//
//		r[rn] -= 0x40
//		return
//	}
//
//    pcIncluded := opcode & 0x8000 != 0
//    psr  := (opcode >> 22) & 1 != 0
//    wb   := (opcode >> 21) & 1 != 0
//    pre  := (opcode >> 24) & 1 != 0
//    load := (opcode >> 20) & 1 != 0
//    forceUser := psr && (c.Reg.CPSR.Mode != MODE_USR) && (!load || !pcIncluded)
//
//    ref := [16]*uint32{
//        &r[0],
//        &r[1],
//        &r[2],
//        &r[3],
//        &r[4],
//        &r[5],
//        &r[6],
//        &r[7],
//        &r[8],
//        &r[9],
//        &r[10],
//        &r[11],
//        &r[12],
//        &r[13],
//        &r[14],
//        &r[15],
//    }
//
//    if forceUser {
//        ref[13] = &c.Reg.SP[BANK_ID[MODE_USR]]
//        ref[14] = &c.Reg.LR[BANK_ID[MODE_USR]]
//    }
//
//    addr := r[rn] &^ 0b11
//    regCount := utils.CountBits(rlist)
//
//    wbValue := r[rn]
//    if up {
//        wbValue += regCount * 4
//    } else {
//        wbValue -= regCount * 4
//    }
//
//	if load {
//
//        if rnIncluded := (rlist>>rn) & 1 == 1; rnIncluded {
//            isLast := (rlist < (1 << (rn + 1)))
//            isOnly := regCount == 1
//            wb = (!isLast || isOnly) && wb
//
//            if !wb || psr {
//                regCount--
//            }
//        }
//
//        if up {
//            for reg := 0; reg < 16; reg++ {
//                if disabled := (rlist >> reg) & 1 == 0; disabled {
//                    continue
//                }
//
//                if pre { addr += 4 }
//                *ref[reg] = c.mem.Read32(addr, true)
//                if !pre { addr += 4 }
//            }
//        } else {
//
//            for reg := 15; reg >= 0; reg-- {
//                if disabled := (rlist >> reg) & 1 == 0; disabled {
//                    continue
//                }
//
//                if pre { addr -= 4 }
//                *ref[reg] = c.mem.Read32(addr, true)
//                if !pre { addr -= 4 }
//            }
//        }
//
//        if wb {
//            r[rn] = wbValue
//        }
//
//        if !pcIncluded {
//            r[PC] += 4
//            return
//        }
//
//        if psr {
//            c.ldmModeSwitch()
//        }
//
//        c.toggleThumb()
//
//        return
//    }
//
//    rnv := *ref[rn]
//
//    if up {
//        for reg := uint32(0); reg < 16; reg++ {
//            if disabled := (rlist >> reg) & 1 == 0; disabled {
//                continue
//            }
//
//            if pre { addr += 4 }
//
//            switch reg {
//            case rn: c.mem.Write32(addr, rnv, true)
//            case PC: c.mem.Write32(addr, *ref[reg]+12, true)
//            default: c.mem.Write32(addr, *ref[reg], true)
//            }
//
//            if !pre { addr += 4 }
//        }
//
//    } else {
//
//        for reg := 15; reg >= 0; reg-- {
//            if disabled := (rlist >> reg) & 1 == 0; disabled {
//                continue
//            }
//
//            if pre { addr -= 4 }
//
//            switch uint32(reg) {
//            case rn:
//                c.mem.Write32(addr, rnv, true)
//            case PC:
//                c.mem.Write32(addr, *ref[reg]+12, true)
//            default:
//                c.mem.Write32(addr, *ref[reg], true)
//            }
//
//            if !pre { addr -= 4 }
//        }
//    }
//
//    if wb {
//        r[rn] = wbValue
//    }
//
//    r[PC] += 4
//}
//
//func (cpu *Cpu) ldmModeSwitch() {
//
//    r := &cpu.Reg.R
//    curr := cpu.Reg.CPSR.Mode
//    spsr := cpu.Reg.SPSR[BANK_ID[curr]]
//
//    reg := &cpu.Reg
//
//    next := spsr.Mode
//    reg.CPSR = spsr
//
//    if curr == MODE_USR {
//        panic("USER MODE LDM PC CHANGE")
//    }
//
//    if curr != MODE_FIQ {
//        for i := range 5 {
//            reg.USR[i] = r[8+i]
//        }
//    }
//
//    reg.SP[BANK_ID[curr]] = r[SP]
//    reg.LR[BANK_ID[curr]] = r[LR]
//
//    if curr == MODE_FIQ {
//        for i := range 5 {
//            reg.FIQ[i] = r[8+i]
//        }
//    }
//
//    if next != MODE_FIQ {
//        for i := range 5 {
//            r[8+i] = reg.USR[i]
//        }
//    }
//
//    r[SP] = reg.SP[BANK_ID[next]]
//    r[LR] = reg.LR[BANK_ID[next]]
//
//    if next == MODE_FIQ {
//        for i := range 5 {
//            r[8+i] = reg.FIQ[i]
//        }
//    }
//}

func (c *Cpu) Block(opcode uint32) {

    r := &c.Reg.R

    rlist := opcode & 0xFFFF
    up := (opcode >> 23) & 1 != 0
    rn := (opcode >> 16) & 0xF
    pcIncluded := opcode & 0x8000 != 0
    psr  := (opcode >> 22) & 1 != 0
    wb   := (opcode >> 21) & 1 != 0
    pre  := (opcode >> 24) & 1 != 0
    load := (opcode >> 20) & 1 != 0
    forceUser := psr && (c.Reg.CPSR.Mode != MODE_USR) && (!load || !pcIncluded)
    //rnIncluded := (rlist>>rn) & 1 == 1

    //regz := c.Reg
    //c.Jit.TestInst(opcode, c.Jit.emitBlock)
    //fmt.Printf("R0 %08X\n", r[0])
    //c.Reg = regz

    //if !pcIncluded { //&& !rnIncluded {

    //    //reg := c.Reg
    //    c.Jit.TestInst(opcode, c.Jit.emitBlock)
    //    if !pcIncluded || !load {
    //        r[PC] += 4
    //    }
    //    //c.Reg = reg
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
            } else {
                addr -= 4
            }
        }

        if load {
            *ref = c.mem.Read32(addr, true)
        } else {
            switch reg {
                case rn: c.mem.Write32(addr, rnv, true)
                case PC: c.mem.Write32(addr, *ref+12, true)
                default: c.mem.Write32(addr, *ref, true)
            }
        }

        if !pre {
            if up {
                addr += 4
            } else {
                addr -= 4
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
