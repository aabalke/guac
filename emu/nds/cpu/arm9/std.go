package arm9

import (
	"fmt"

	"github.com/aabalke/guac/emu/nds/utils"
)

const (
    IA = 0
    IB = 1
    DB = 2
    DA = 3
)

type Block struct {
	Opcode, Rn, Rlist             uint32
	Pre, Up, PSR, Writeback, Load bool
    Method uint8
}

func (c *Cpu) Block(opcode uint32) {

	block := &Block{
		Opcode:      opcode,
		Pre:         utils.BitEnabled(opcode, 24),
		Up:          utils.BitEnabled(opcode, 23),
		PSR:         utils.BitEnabled(opcode, 22),
		Writeback:   utils.BitEnabled(opcode, 21),
		Load:        utils.BitEnabled(opcode, 20),
		Rn:          utils.GetVarData(opcode, 16, 19),
		Rlist:       utils.GetVarData(opcode, 0, 15),
	}

    switch {
    case block.Pre && block.Up:
        block.Method = IB
    case !block.Pre && block.Up:
        block.Method = IA
    case block.Pre && !block.Up:
        block.Method = DB
    case !block.Pre && !block.Up:
        block.Method = DA
    }

	if block.Load {
		c.ldm(block)
        return
    }

    c.stm(block)
}

func (cpu *Cpu) ldm(block *Block) {

	r := &cpu.Reg.R

	if utils.BitEnabled(block.Opcode, 15) && block.PSR {
        fmt.Printf("ON ENTRY PC %08X\n", r[15])
    }


	addr := r[block.Rn] &^ 0b11
	wbValue := r[block.Rn]
    keepPC := false

	if block.Rlist == 0 {

		cpu.Reg.R[PC] += 4

		if block.Up {
			r[block.Rn] += 0x40
			return
		}

		r[block.Rn] -= 0x40
		return
	}

	mode := cpu.Reg.getMode()
	if forceUser := block.PSR; forceUser && mode != MODE_USR {
		cpu.Reg.setMode(mode, MODE_USR)
	}

	regCount := utils.CountBits(block.Rlist)

	if rnIncluded := (block.Rlist>>block.Rn)&1 == 1; rnIncluded {
        isLast := (block.Rlist < (1 << (block.Rn + 1)))
        isOnly := regCount == 1 && rnIncluded
        block.Writeback = !isLast || isOnly

        if !block.Writeback || block.PSR {
            regCount--
        }
    }

    switch block.Method {
    case IB:
        keepPC = ldmIB(cpu, addr, block.Rlist)
    case IA:
        keepPC = ldmIA(cpu, addr, block.Rlist)
    case DB:
        keepPC = ldmDB(cpu, addr, block.Rlist)
    case DA:
        keepPC = ldmDA(cpu, addr, block.Rlist)
    }

	if forceUser := block.PSR; forceUser && mode != MODE_USR {
		curr := cpu.Reg.getMode()
		cpu.Reg.setMode(curr, mode)
	}

	if block.Writeback {

        if block.Up {
            wbValue += regCount * 4

        } else {
            wbValue -= regCount * 4
        }

		r[block.Rn] = wbValue
	}

    if !keepPC {
		cpu.Reg.R[PC] += 4
        return
    }

	if utils.BitEnabled(block.Opcode, 15) && block.PSR {

        curr := cpu.Reg.getMode()
        spsr := cpu.Reg.SPSR[BANK_ID[curr]]

        reg := &cpu.Reg

        // I think this is necessary for irq exits

        if irqExit := curr == MODE_IRQ; irqExit {

            fmt.Printf("CALLED LDM^ R15 %08X -> %08X\n", r[PC], r[LR])

            r[PC] = r[LR]

            //r[PC] += 4
        }

        //cpu.Reg.setMode(cpu.Reg.getMode(), uint32(cpu.Reg.SPSR[BANK_ID[cpu.Reg.getMode()]]) & 0x1F)

        fmt.Printf("PC %08X\n", r[15])

        next := uint32(spsr) & 0b11111
        //cpsr := uint32(reg.CPSR)

        reg.CPSR = Cond(spsr)
        reg.IsThumb = reg.CPSR.GetFlag(FLAG_T)

        //if curr == MODE_USR {
        //    panic("USER MODE LDM PC CHANGE")
        //}

        //if curr != MODE_FIQ {
        //    for i := range 5 {
        //        reg.USR[i] = r[8+i]
        //    }
        //}

        reg.SP[BANK_ID[curr]] = r[SP]
        reg.LR[BANK_ID[curr]] = r[LR]

        //if curr == MODE_FIQ {
        //    for i := range 5 {
        //        reg.FIQ[i] = r[8+i]
        //    }
        //}

        //if next != MODE_FIQ {
        //    for i := range 5 {
        //        r[8+i] = reg.USR[i]
        //    }
        //}

        r[SP] = reg.SP[BANK_ID[next]]
        r[LR] = reg.LR[BANK_ID[next]]

        //if next == MODE_FIQ {
        //    for i := range 5 {
        //        r[8+i] = reg.FIQ[i]
        //    }
        //}

        return
	}


    cpu.toggleThumb()
}

func ldmIB(cpu *Cpu, addr, rlist uint32) bool {

	r := &cpu.Reg.R
    keepPc := false

    for reg := uint32(0); reg < 16; reg++ {
        if disabled := !utils.BitEnabled(rlist, uint8(reg)); disabled {
            continue
        }

        addr += 4

        r[reg] = cpu.mem.Read32(addr, true)

        if reg == PC {
            keepPc = true
        }
    }

    return keepPc
}

func ldmIA(cpu *Cpu, addr, rlist uint32) bool {

	r := &cpu.Reg.R
    keepPc := false

    for reg := uint32(0); reg < 16; reg++ {
        if disabled := !utils.BitEnabled(rlist, uint8(reg)); disabled {
            continue
        }

        r[reg] = cpu.mem.Read32(addr, true)

        if reg == PC {
            keepPc = true
        }

        addr += 4
    }

    return keepPc
}

func ldmDB(cpu *Cpu, addr, rlist uint32) bool {

	r := &cpu.Reg.R
    keepPc := false

    for reg := 15; reg >= 0; reg-- {
        if disabled := !utils.BitEnabled(rlist, uint8(reg)); disabled {
            continue
        }

        addr -= 4

        r[reg] = cpu.mem.Read32(addr, true)

        if reg == PC {
            keepPc = true
        }

    }

    return keepPc
}

func ldmDA(cpu *Cpu, addr, rlist uint32) bool {

	r := &cpu.Reg.R
    keepPc := false

    for reg := 15; reg >= 0; reg-- {
        if disabled := !utils.BitEnabled(rlist, uint8(reg)); disabled {
            continue
        }

        r[reg] = cpu.mem.Read32(addr, true)
        addr -= 4

        if reg == PC {
            keepPc = true
        }

    }

    return keepPc
}

// STM STM STM STM //

func (cpu *Cpu) stm(block *Block) {

	r := &cpu.Reg.R

	addr := r[block.Rn] &^ 0b11
	wbValue := r[block.Rn]

	mode := cpu.Reg.getMode()
	if forceUser := block.PSR; forceUser && mode != MODE_USR {
		cpu.Reg.setMode(mode, MODE_USR)
	}

	if block.Rlist == 0 {

        r[PC] += 4

        if block.Up {
			r[block.Rn] += 0x40
            return
        }

        r[block.Rn] -= 0x40

		return
	}

    switch block.Method {
    case IA:
        stmIA(cpu, addr, block.Rlist, block.Rn)
    case IB:
        stmIB(cpu, addr, block.Rlist, block.Rn)
    case DB:
        stmDB(cpu, addr, block.Rlist, block.Rn)
    case DA:
        stmDA(cpu, addr, block.Rlist, block.Rn)
    }

	if forceUser := block.PSR; forceUser && mode != MODE_USR {
		curr := cpu.Reg.getMode()
		cpu.Reg.setMode(curr, mode)
	}

	if !block.Writeback {
		r[block.Rn] = wbValue
    }

    r[PC] += 4
}

func stmIB(cpu *Cpu, addr, rlist, rn uint32) {

	r := &cpu.Reg.R
    rnValue := r[rn]

    for reg := uint32(0); reg < 16; reg++ {
        if disabled := !utils.BitEnabled(rlist, uint8(reg)); disabled {
            continue
        }

        r[rn] += 4
        addr += 4

        switch reg {
        case rn:
            cpu.mem.Write32(addr, rnValue, true)

        case PC:
            cpu.mem.Write32(addr, r[reg]+12, true)

        default:
            cpu.mem.Write32(addr, r[reg], true)
        }
    }
}

func stmIA(cpu *Cpu, addr, rlist, rn uint32) {

	r := &cpu.Reg.R

    rnValue := r[rn]

    for reg := uint32(0); reg < 16; reg++ {
        if disabled := !utils.BitEnabled(rlist, uint8(reg)); disabled {
            continue
        }

        switch reg {
        case rn:
            cpu.mem.Write32(addr, rnValue, true)
            r[rn] += 4
            addr += 4

        case PC:
            cpu.mem.Write32(addr, r[reg]+12, true)
            r[rn] += 4
            addr += 4

        default:
            cpu.mem.Write32(addr, r[reg], true)
            r[rn] += 4
            addr += 4
        }
    }
}

func stmDB(cpu *Cpu, addr, rlist, rn uint32) {

	r := &cpu.Reg.R
    rnValue := r[rn]

    for reg := 15; reg >= 0; reg-- {
        if disabled := !utils.BitEnabled(rlist, uint8(reg)); disabled {
            continue
        }

        r[rn] -= 4
        addr -= 4

        switch uint32(reg) {
        case rn:
            cpu.mem.Write32(addr, rnValue, true)
        case PC:
            cpu.mem.Write32(addr, r[reg]+12, true)

        default:

            cpu.mem.Write32(addr, r[reg], true)
        }
    }
}

func stmDA(cpu *Cpu, addr, rlist, rn uint32) {

	r := &cpu.Reg.R
    rnValue := r[rn]

    for reg := 15; reg >= 0; reg-- {
        if disabled := !utils.BitEnabled(rlist, uint8(reg)); disabled {
            continue
        }

        switch uint32(reg) {
        case rn:
            cpu.mem.Write32(addr, rnValue, true)
            r[rn] -= 4
            addr -= 4

        case PC:
            cpu.mem.Write32(addr, r[reg]+12, true)
            r[rn] -= 4
            addr -= 4

        default:
            cpu.mem.Write32(addr, r[reg], true)
            r[rn] -= 4
            addr -= 4
        }
    }
}
