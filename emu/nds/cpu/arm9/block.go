package arm9

import (
	"fmt"
	"unsafe"

	"github.com/aabalke/guac/emu/nds/utils"
)

var _ = fmt.Sprint

const (
    IA = 0
    IB = 1
    DB = 2
    DA = 3
)

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
