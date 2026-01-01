package arm7

import (
	"github.com/aabalke/guac/emu/gba/utils"
)

type Block struct {
	Opcode, Rn, RnValue, Rlist    uint32
	Pre, Up, PSR, Writeback, Load bool
	ForceUser                     bool
}

func (c *Cpu) Block(opcode uint32) {

	block := &Block{
		Opcode:    opcode,
		Pre:       utils.BitEnabled(opcode, 24),
		Up:        utils.BitEnabled(opcode, 23),
		PSR:       utils.BitEnabled(opcode, 22),
		Writeback: utils.BitEnabled(opcode, 21),
		Load:      utils.BitEnabled(opcode, 20),
		Rn:        (opcode >> 16) & 0xF,
		Rlist:     opcode & 0xFFFF,
	}

	mode := c.Reg.CPSR.Mode
	block.ForceUser = (block.PSR &&
		mode != MODE_USR &&
		(!block.Load || (block.Load && !utils.BitEnabled(opcode, 15))))

	block.RnValue = c.Reg.R[block.Rn]

	if utils.BitEnabled(block.Opcode, 15) && block.PSR && block.Load {
		panic("ARM7 LDM WITH R15 AND SET USED")
	}

	incPc := true
	if block.Load {
		incPc = c.ldm(block)
	} else {
		c.stm(block)
	}

	if incPc {
		c.Reg.R[PC] += 4
	}
}

func (c *Cpu) ldm(block *Block) bool {

	ref := [16]*uint32{
		&c.Reg.R[0],
		&c.Reg.R[1],
		&c.Reg.R[2],
		&c.Reg.R[3],
		&c.Reg.R[4],
		&c.Reg.R[5],
		&c.Reg.R[6],
		&c.Reg.R[7],
		&c.Reg.R[8],
		&c.Reg.R[9],
		&c.Reg.R[10],
		&c.Reg.R[11],
		&c.Reg.R[12],
		&c.Reg.R[13],
		&c.Reg.R[14],
		&c.Reg.R[15],
	}

	if block.ForceUser {
		ref[13] = &c.Reg.SP[BANK_ID[MODE_USR]]
		ref[14] = &c.Reg.LR[BANK_ID[MODE_USR]]
	}

	incPC := true
	r := &c.Reg.R

	ib := block.Pre && block.Up
	ia := !block.Pre && block.Up
	db := block.Pre && !block.Up
	da := !block.Pre && !block.Up

	if block.Rlist == 0 {

		c.Reg.R[PC] += 16 // i believe this is short cut for {} => {r15} behavior

		if block.Up {
			r[block.Rn] += 0x40
			return false
		}

		r[block.Rn] -= 0x40
		return false
	}

	regCount := utils.CountBits(block.Rlist)

	if (block.Rlist>>block.Rn)&1 == 1 {
		regCount--
		block.Writeback = false
	}

	addr := r[block.Rn] &^ 0b11

	reg := uint32(0)
	for reg = range 16 {

		regBitEnabled := utils.BitEnabled(block.Rlist, uint8(reg))
		decRegBitEnabled := utils.BitEnabled(block.Rlist, uint8(15-reg))

		switch {
		case ib && regBitEnabled:

			addr += 4
			*ref[reg] = c.mem.Read32(addr, false)

			if reg == PC {
				incPC = incPC && false
			}
			if reg == block.Rn {
				block.RnValue = r[block.Rn]
			}

		case ia && regBitEnabled:

			*ref[reg] = c.mem.Read32(addr, false)

			if reg == PC {
				incPC = incPC && false
			}
			if reg == block.Rn {
				block.RnValue = r[block.Rn]
			}

			addr += 4

		case db && decRegBitEnabled: // pop

			addr -= 4
			*ref[15-reg] = c.mem.Read32(addr, false)

			if 15-reg == PC {
				incPC = incPC && false
			}
			if 15-reg == block.Rn {
				block.RnValue = r[block.Rn]
			}

		case da && decRegBitEnabled:

			*ref[15-reg] = c.mem.Read32(addr, false)
			addr -= 4

			if 15-reg == PC {
				incPC = incPC && false
			}
			if 15-reg == block.Rn {
				block.RnValue = r[block.Rn]
			}
		}
	}

	if block.Up {
		r[block.Rn] += (regCount * 4)
	} else {
		r[block.Rn] -= (regCount * 4)
	}

	if !block.Writeback {
		r[block.Rn] = block.RnValue
	}

	return incPC
}

func (c *Cpu) stm(block *Block) {

	ref := [16]*uint32{
		&c.Reg.R[0],
		&c.Reg.R[1],
		&c.Reg.R[2],
		&c.Reg.R[3],
		&c.Reg.R[4],
		&c.Reg.R[5],
		&c.Reg.R[6],
		&c.Reg.R[7],
		&c.Reg.R[8],
		&c.Reg.R[9],
		&c.Reg.R[10],
		&c.Reg.R[11],
		&c.Reg.R[12],
		&c.Reg.R[13],
		&c.Reg.R[14],
		&c.Reg.R[15],
	}

	if block.ForceUser {
		ref[13] = &c.Reg.SP[BANK_ID[MODE_USR]]
		ref[14] = &c.Reg.LR[BANK_ID[MODE_USR]]
	}

	r := &c.Reg.R

	ib := block.Pre && block.Up
	ia := !block.Pre && block.Up
	db := block.Pre && !block.Up
	da := !block.Pre && !block.Up

	if block.Rlist == 0 {
		// stm {} => {PC}
		switch {
		case ib:
			addr := r[block.Rn] + 4
			r[block.Rn] += 0x40
			c.mem.Write32(addr, r[PC]+12, false)
		case ia:
			c.mem.Write32(r[block.Rn], r[PC]+12, false)
			r[block.Rn] += 0x40
		case db:
			r[block.Rn] -= 0x40
			c.mem.Write32(r[block.Rn], r[PC]+12, false)
		case da:
			r[block.Rn] -= 0x40
			c.mem.Write32(r[block.Rn]+4, r[PC]+12, false)
		}

		return
	}

	regCount := utils.CountBits(block.Rlist)

	smallest := (block.Rlist & -block.Rlist) == 1<<block.Rn
	matchingRn := (block.Rlist>>block.Rn)&1 == 1
	matchingValue := uint32(0)
	matchingAddr := uint32(0) // rn during regs

	addr := r[block.Rn] &^ 0b11

	count := uint32(0)
	rnIdx := uint32(0)
	for reg := range 16 {

		regBitEnabled := utils.BitEnabled(block.Rlist, uint8(reg))
		decRegBitEnabled := utils.BitEnabled(block.Rlist, uint8(15-reg))

		switch {
		case ib && regBitEnabled:

			count++

			r[block.Rn] += 4
			addr += 4

			if reg == int(block.Rn) {
				c.mem.Write32(addr, *ref[reg]-4, false)
				matchingValue = *ref[reg]
				matchingAddr = addr
				rnIdx = regCount - count
				continue
			}

			if reg == PC {
				c.mem.Write32(addr, *ref[reg]+12, false)
				continue
			}

			c.mem.Write32(addr, *ref[reg], false)

		case ia && regBitEnabled:

			count++

			if reg == int(block.Rn) {

				c.mem.Write32(addr, *ref[reg], false)
				matchingValue = *ref[reg] + 4
				matchingAddr = addr
				rnIdx = regCount - count
				r[block.Rn] += 4
				addr += 4
				continue
			}

			if reg == PC {
				c.mem.Write32(addr, *ref[reg]+12, false)
				continue
			}

			c.mem.Write32(addr, *ref[reg], false)

			r[block.Rn] += 4
			addr += 4

		case db && decRegBitEnabled: // push

			count++

			r[block.Rn] -= 4
			addr -= 4

			if 15-reg == int(block.Rn) {
				matchingValue = *ref[15-reg]
				matchingAddr = addr
				rnIdx = regCount - count // regCount only for 15 - reg
			}
			if 15-reg == PC {
				c.mem.Write32(addr, *ref[15-reg]+12, false)
				continue
			}

			c.mem.Write32(addr, *ref[15-reg], false)

		case da && decRegBitEnabled:

			count++

			decReg := 15 - reg

			if decReg == int(block.Rn) {
				c.mem.Write32(addr, *ref[decReg]+(count-1)*4, false)
				matchingValue = *ref[decReg] - 4 // -4 offsets above +4 when matching Value (not first smallest)
				matchingAddr = addr
				rnIdx = regCount - count
				r[block.Rn] -= 4
				addr -= 4
				continue
			}

			if decReg == PC {
				c.mem.Write32(addr, *ref[decReg]+12, false)
				continue
			}

			c.mem.Write32(addr, *ref[decReg], false)

			r[block.Rn] -= 4
			addr -= 4
		}
	}

	if !block.Writeback {
		r[block.Rn] = block.RnValue
		return
	}

	if block.Writeback && smallest {

		v := c.mem.Read32(addr, false)

		if block.Up {
			c.mem.Write32(r[block.Rn], v-(regCount*4), false)
			return
		}
		c.mem.Write32(r[block.Rn], v+(regCount*4), false)
		return
	}

	if block.Writeback && matchingRn {
		if block.Up {
			c.mem.Write32(matchingAddr, matchingValue+(rnIdx*4), false)
			return
		}

		c.mem.Write32(matchingAddr, matchingValue-(rnIdx*4), false)
		return
	}
}
