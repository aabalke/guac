package arm7

import (
	"fmt"
	"math/bits"
)

// NOTE: order of decoders effects order checked and therefore implimentation
const (
	INST_THUMB_BRA Decoder = iota
	INST_THUMB_SVC
	INST_THUMB_HIBX
	INST_THUMB_HI
	INST_THUMB_CBR
	INST_THUMB_REL
	INST_THUMB_SP
	INST_THUMB_PCR
	INST_THUMB_LDSH
	INST_THUMB_SDTR
	INST_THUMB_SDTI
	INST_THUMB_SDTH
	INST_THUMB_SPR
	INST_THUMB_BLK
	INST_THUMB_POPPC
	INST_THUMB_PUS
	INST_THUMB_ADD
	INST_THUMB_IMM
	INST_THUMB_SHI
	INST_THUMB_ALU
	INST_THUMB_BL_LONG
	INST_THUMB_BL_SHORT
)

var thumbFormatters = []Formatter{
	INST_THUMB_BRA:      {0xF800, 0xE000},
	INST_THUMB_SVC:      {0xFF00, 0xDF00},
	INST_THUMB_HIBX:     {0xFF00, 0x4700},
	INST_THUMB_HI:       {0xFC00, 0x4400},
	INST_THUMB_CBR:      {0xF000, 0xD000},
	INST_THUMB_ALU:      {0xFC00, 0x4000},
	INST_THUMB_IMM:      {0xE000, 0x2000},
	INST_THUMB_ADD:      {0xF800, 0x1800},
	INST_THUMB_SHI:      {0xE000, 0x0000},
	INST_THUMB_POPPC:    {0xFF00, 0xBD00},
	INST_THUMB_PUS:      {0xF600, 0xB400},
	INST_THUMB_BLK:      {0xF000, 0xC000},
	INST_THUMB_REL:      {0xF000, 0xA000},
	INST_THUMB_SP:       {0xFF00, 0xB000},
	INST_THUMB_PCR:      {0xF800, 0x4800},
	INST_THUMB_LDSH:     {0xFE00, 0x5E00},
	INST_THUMB_SDTR:     {0xF000, 0x5000},
	INST_THUMB_SDTI:     {0xE000, 0x6000},
	INST_THUMB_SDTH:     {0xF000, 0x8000},
	INST_THUMB_SPR:      {0xF000, 0x9000},
	INST_THUMB_BL_LONG:  {0xF800, 0xF000},
	INST_THUMB_BL_SHORT: {0xE800, 0xE800},
}

func (c *Cpu) DecodeThumb(op32 uint32) {
	op := uint16(op32)

	for decoder, desc := range thumbFormatters {

		if (op32 & desc.Mask) != desc.Match {
			continue
		}

		switch Decoder(decoder) {
		case INST_THUMB_SVC:
			c.Exception(VEC_SWI, MODE_SWI)
		case INST_THUMB_ADD:
			c.ThumbAddSub(op)
		case INST_THUMB_SHI:
			c.ThumbShifted(op)
		case INST_THUMB_IMM:
			c.ThumbImm(op)
		case INST_THUMB_ALU:
			c.ThumbAlu(op)
		case INST_THUMB_HIBX:
			c.ThumbHiBx(op)
		case INST_THUMB_HI:
			c.ThumbHi(op)
		case INST_THUMB_SDTH:
			c.ThumbLSHalf(op)
		case INST_THUMB_SDTI:
			c.ThumbLSImm(op)
		case INST_THUMB_SDTR:
			c.ThumbSdt(op)
		case INST_THUMB_POPPC:
			c.ThumbPopPc(op)
		case INST_THUMB_PUS:
			c.ThumbPushPop(op)
		case INST_THUMB_REL:
			c.ThumbRelative(op)
		case INST_THUMB_BRA:
			c.ThumbBranch(op)
		case INST_THUMB_BL_SHORT:
			c.ThumbShortLongBranch(op)
		case INST_THUMB_BL_LONG:
			c.ThumbLongBranch(op)
		case INST_THUMB_BLK:
			c.ThumbBlock(op)
		case INST_THUMB_PCR:
			c.ThumbLPC(op)
		case INST_THUMB_CBR:
			c.ThumbJumpCalls(op)
		case INST_THUMB_SP:
			c.ThumbStack(op)
		case INST_THUMB_SPR:
			c.ThumbLSSP(op)
		case INST_THUMB_LDSH:
			c.ThumbLDSH(op)
		}
		return
	}

	panic(fmt.Sprintf("arm7: decode failed pc %08X op %04X\n", c.Reg.R[PC], op))
}

//go:inline
func IsThumbOpFormat(op, mask, fmt uint16) bool {
	return op&mask == fmt
}

const (
	THUMB_AND = iota
	THUMB_EOR
	THUMB_LSL
	THUMB_LSR
	THUMB_ASR
	THUMB_ADC
	THUMB_SBC
	THUMB_ROR
	THUMB_TST
	THUMB_NEG // arm eq is RSBS
	THUMB_CMP
	THUMB_CMN
	THUMB_ORR
	THUMB_MUL
	THUMB_BIC
	THUMB_MVN
)

func (c *Cpu) ThumbAlu(op uint16) {
	var (
		r    = &c.Reg.R
		cpsr = &c.Reg.CPSR

		rsv = r[(op>>3)&7]
		rd  = op & 7
		rdv = r[rd]
		res uint64
	)

	switch inst := (op >> 6) & 0xF; inst {
	case THUMB_MUL:
		c.Idle(idleMul(rdv, true))

		res = uint64(rdv) * uint64(rsv)
		r[rd] = uint32(res)
		// ARM < 4, carry flag destroyed, ARM >= 5, carry flag unchanged
		// cpsr.C = false

	case THUMB_TST:
		res = uint64(rdv) & uint64(rsv)

	case THUMB_CMN:
		res = uint64(rdv) + uint64(rsv)
		cpsr.V = ((^(rdv ^ rsv))&(rdv^uint32(res)))>>31 != 0
		cpsr.C = res >= 0x1_0000_0000

	case THUMB_CMP:
		res = uint64(rdv) - uint64(rsv)
		cpsr.C = res < 0x1_0000_0000
		cpsr.V = ((rdv^rsv)&(rdv^uint32(res)))>>31 != 0

	case THUMB_AND:
		res = uint64(rdv) & uint64(rsv)
		r[rd] = uint32(res)

	case THUMB_EOR:
		res = uint64(rdv) ^ uint64(rsv)
		r[rd] = uint32(res)

	case THUMB_ORR:
		res = uint64(rdv) | uint64(rsv)
		r[rd] = uint32(res)

	case THUMB_BIC:
		res = uint64(rdv) &^ uint64(rsv)
		r[rd] = uint32(res)

	case THUMB_MVN:
		res = ^uint64(rsv)
		r[rd] = uint32(res)

	case THUMB_NEG:
		res = 0 - uint64(rsv)
		r[rd] = uint32(res)

		cpsr.C = res < 0x1_0000_0000
		cpsr.V = ((rdv^rsv)&(rdv^uint32(res)))>>31 != 0

	case THUMB_SBC:
		res = uint64(rdv) - uint64(rsv)
		if !cpsr.C {
			res--
		}

		cpsr.V = ((rdv^rsv)&(rdv^uint32(res)))>>31 != 0
		cpsr.C = res < 0x1_0000_0000

		r[rd] = uint32(res)

	case THUMB_ADC:
		res = uint64(rdv) + uint64(rsv)
		if cpsr.C {
			res++
		}

		cpsr.V = ((^(rdv ^ rsv))&(rdv^uint32(res)))>>31 != 0
		cpsr.C = res >= 0x1_0000_0000

		r[rd] = uint32(res)

	case THUMB_LSL, THUMB_LSR, THUMB_ASR, THUMB_ROR:

		rsv &= 0xFF

		if rsv != 0 {

			var shType uint32
			switch inst {
			case THUMB_LSL:
				shType = LSL
			case THUMB_LSR:
				shType = LSR
			case THUMB_ASR:
				shType = ASR
			case THUMB_ROR:
				shType = ROR
			}

			res = uint64(c.ShiftReg(rdv, rsv, shType, &cpsr.C))

		} else {
			c.Idle(1)
			res = uint64(rdv)
		}

		r[rd] = uint32(res)
	}

	cpsr.N = (res>>31)&1 != 0
	cpsr.Z = uint32(res) == 0
}

const (
	HI_ADD = iota
	HI_CMP
	HI_MOV
	HI_BX
)

func (c *Cpu) ThumbHi(op uint16) {
	var (
		r  = &c.Reg.R
		rd = (op & 7) | (((op >> 7) & 1) << 3)
		rs = (op >> 3) & 0xF
	)

	v := r[rs]
	if rs == PC {
		v &^= 1
	}

	switch inst := (op >> 8) & 3; inst {
	case HI_ADD:

		r[rd] += v

		if rd == PC {
			c.Reload = true
		}

	case HI_CMP:

		rdv := r[rd]
		res := uint64(rdv) - uint64(v)

		c.Reg.CPSR.N = (uint32(res)>>31)&1 != 0
		c.Reg.CPSR.Z = uint32(res) == 0
		c.Reg.CPSR.C = res < 0x1_0000_0000
		c.Reg.CPSR.V = ((rdv^v)&(rdv^uint32(res)))>>31 != 0

	case HI_MOV:
		if nop := op == 0x46C0; nop {
			return
		}

		r[rd] = v

		if rd == PC {
			c.Reload = true
		}

	case HI_BX:
		panic("use thumb hi bx not thumb hi switch case")
	}
}

func (c *Cpu) ThumbHiBx(op uint16) {
	var (
		r  = &c.Reg.R
		rs = (op >> 3) & 0xF
	)

	v := r[rs]
	if rs == PC {
		v &^= 1
	}
	r[PC] = v
	c.ToggleThumb()
}

const (
	THUMB_ADD = iota
	THUMB_SUB
	THUMB_ADDImm
	THUMB_SUBImm
)

func (c *Cpu) ThumbAddSub(op uint16) {
	var (
		r    = &c.Reg.R
		cpsr = &c.Reg.CPSR
		inst = (op >> 9) & 3
		rsv  = uint32(r[(op>>3)&7])
		rd   = op & 7
		res  uint64
	)

	var op2 uint32
	if reg := inst < 2; reg {
		op2 = uint32(r[(op>>6)&7])
	} else {
		op2 = uint32((op >> 6) & 7)
	}

	switch inst {
	case THUMB_ADD, THUMB_ADDImm:
		res = uint64(rsv) + uint64(op2)
		cpsr.V = ((^(rsv ^ op2))&(rsv^uint32(res)))>>31 != 0
		cpsr.C = res >= 0x1_0000_0000
	case THUMB_SUB, THUMB_SUBImm:
		res = uint64(rsv) - uint64(op2)
		cpsr.V = ((rsv^op2)&(rsv^uint32(res)))>>31 != 0
		cpsr.C = res < 0x1_0000_0000
	}

	r[rd] = uint32(res)
	cpsr.N = (uint32(res)>>31)&1 != 0
	cpsr.Z = uint32(res) == 0
}

const (
	THUMB_IMM_MOV = iota
	THUMB_IMM_CMP
	THUMB_IMM_ADD
	THUMB_IMM_SUB
)

func (c *Cpu) ThumbImm(op uint16) {
	var (
		r    = &c.Reg.R
		cpsr = &c.Reg.CPSR
		inst = (op >> 11) & 3
		rd   = (op >> 8) & 7
		nn   = uint32(op & 0xFF)
		rdv  = r[rd]
		res  uint64
	)

	switch inst {
	case THUMB_IMM_MOV:
		res = uint64(nn)
		r[rd] = uint32(res)
	case THUMB_IMM_CMP:
		res = uint64(rdv) - uint64(nn)
		cpsr.V = ((rdv^nn)&(rdv^uint32(res)))>>31 != 0
		cpsr.C = res < 0x1_0000_0000
	case THUMB_IMM_ADD:
		res = uint64(rdv) + uint64(nn)
		r[rd] = uint32(res)
		cpsr.V = ((^(rdv ^ nn))&(rdv^uint32(res)))>>31 != 0
		cpsr.C = res >= 0x1_0000_0000
	case THUMB_IMM_SUB:
		res = uint64(rdv) - uint64(nn)
		r[rd] = uint32(res)
		cpsr.V = ((rdv^nn)&(rdv^uint32(res)))>>31 != 0
		cpsr.C = res < 0x1_0000_0000
	}

	cpsr.N = (res>>31)&1 != 0
	cpsr.Z = uint32(res) == 0
}

func (c *Cpu) ThumbLSHalf(op uint16) {
	var (
		r    = &c.Reg.R
		addr = r[(op>>3)&7] + uint32((op>>6)&0x1F<<1)
		rd   = op & 7
	)

	if ldr := (op>>11)&1 != 0; ldr {
		v := uint32(c.Read16(addr))
		r[rd] = bits.RotateLeft32(v, -int((addr&1)*8))
	} else {
		c.Write16(addr, uint16(r[rd]))
	}
}

const (
	THUMB_STR = iota
	THUMB_STRH
	THUMB_STRB
	THUMB_LDSB
	THUMB_LDR
	THUMB_LDRH
	THUMB_LDRB
	THUMB_LDSH
)

func (c *Cpu) ThumbSdt(op uint16) {
	var (
		r    = &c.Reg.R
		rd   = op & 7
		addr = r[(op>>3)&7] + r[(op>>6)&7]
	)

	switch inst := (op >> 9) & 7; inst {
	case THUMB_STR:
		c.Write32(addr, r[rd])
	case THUMB_STRH:
		c.Write16(addr, uint16(r[rd]))
	case THUMB_STRB:
		c.Write8(addr, uint8(r[rd]))
	case THUMB_LDSB:
		r[rd] = uint32(int32(int8(c.Read8(addr))))
	case THUMB_LDR:
		v := c.Read32(addr)
		r[rd] = bits.RotateLeft32(v, -int((addr&3)*8))
	case THUMB_LDRH:
		v := c.Read16(addr)
		r[rd] = bits.RotateLeft32(v, -int((addr&1)<<3))
	case THUMB_LDRB:
		r[rd] = c.Read8(addr)
	case THUMB_LDSH:
		panic("use ThumbLDRSH function separately")
	}
}

func (c *Cpu) ThumbLDSH(op uint16) {
	var (
		r    = &c.Reg.R
		rd   = op & 7
		addr = r[(op>>3)&7] + r[(op>>6)&7]
	)
	if misaligned := addr&1 != 0; misaligned {
		r[rd] = uint32(int32(int8(c.Read8(addr))))
		return
	}
	r[rd] = uint32(int32(int16(c.Read16(addr))))
}

func (c *Cpu) ThumbLPC(op uint16) {
	c.Reg.R[(op>>8)&7] = c.Read32((c.Reg.R[PC] &^ 3) + uint32(op&0xFF)<<2)
}

const (
	THUMB_STR_IMM = iota
	THUMB_LDR_IMM
	THUMB_STRB_IMM
	THUMB_LDRB_IMM
)

func (c *Cpu) ThumbLSImm(op uint16) {
	var (
		r    = &c.Reg.R
		inst = (op >> 11) & 3
		rd   = op & 7
		rb   = (op >> 3) & 7
		nn   = uint32(op>>6) & 0x1F
	)

	switch inst {
	case THUMB_STR_IMM:
		addr := r[rb] + (nn << 2)
		c.Write32(addr, r[rd])
	case THUMB_LDR_IMM:
		addr := r[rb] + (nn << 2)
		v := c.Read32(addr)
		r[rd] = bits.RotateLeft32(v, -int((addr&3)*8))
	case THUMB_STRB_IMM:
		addr := r[rb] + nn
		c.Write8(addr, uint8(r[rd]))
	case THUMB_LDRB_IMM:
		addr := r[rb] + nn
		r[rd] = uint32(c.Read8(addr))
	}
}

func (c *Cpu) ThumbPopPc(op uint16) {
	r := &c.Reg.R
	seq := c.ThumbPushPop(op)
	r[PC] = c.Read32Block(r[SP], seq)
	r[SP] += 4
	c.Idle(1) // gba has, not sure if arm9
	c.ToggleThumb()
}

func (c *Cpu) ThumbPushPop(op uint16) uint32 {
	var (
		r     = &c.Reg.R
		pclr  = (op>>8)&1 != 0
		rlist = op & 0xFF
		pop   = (op>>11)&1 != 0
		seq   = uint32(NONSEQ)
	)

	// thank you nano
	if rlist == 0 && !pclr {
		if pop {
			r[PC] = c.Read32Block(r[SP], seq)
			c.Reload = true
			r[SP] += 0x40
		} else {
			// alyosha test fails this.
			// i think it is timing related
			r[SP] -= 0x40
			c.Write32Block(r[SP], r[PC], seq)
		}

		return seq
	}

	if pop {
		for reg := range 8 {
			if disabled := (rlist>>reg)&1 == 0; disabled {
				continue
			}

			r[reg] = c.Read32Block(r[SP], seq)
			r[SP] += 4

			seq = SEQ
		}

		if !pclr {
			c.Idle(1)
		}

	} else {

		if pclr {
			r[SP] -= 4
			c.Write32Block(r[SP], r[LR], seq)
			seq = SEQ
		}

		for reg := 7; reg >= 0; reg-- {
			if disabled := (rlist>>reg)&1 == 0; disabled {
				continue
			}

			r[SP] -= 4
			c.Write32Block(r[SP], r[reg], seq)

			seq = SEQ
		}
	}

	return seq
}

func (c *Cpu) ThumbRelative(op uint16) {
	var (
		r  = &c.Reg.R
		rd = (op >> 8) & 7
		nn = uint32(op&0xFF) << 2
	)

	if isSP := (op>>11)&1 != 0; isSP {
		r[rd] = r[SP] + nn

		return
	}

	r[rd] = (r[PC] &^ 3) + nn
}

func (c *Cpu) ThumbJumpCalls(op uint16) {
	r := &c.Reg.R

	if !c.Reg.CPSR.CheckCond(uint32(op>>8) & 0xF) {
		c.Seq = SEQ
		return
	}

	nn := int(int8(op&0xFF)) << 1
	r[PC] = uint32(int(r[PC]) + nn)
	c.Reload = true
}

func (c *Cpu) ThumbBranch(op uint16) {
	c.Reg.R[PC] += uint32(int16((op&0x7FF)<<5) >> 4)
	c.Reload = true
}

func (c *Cpu) ThumbShifted(op uint16) {
	v := c.Reg.R[(op>>3)&7]
	v = c.ShiftImm(uint32(op>>11)&3, v, uint32(op>>6)&0x1F, &c.Reg.CPSR.C)
	c.Reg.CPSR.N = (v>>31)&1 != 0
	c.Reg.CPSR.Z = uint32(v) == 0
	c.Reg.R[op&7] = v
}

func (c *Cpu) ThumbStack(op uint16) {
	// NOTE: not 2s compliment, do not int8 to remove branch
	nn := int(op&0x7F) << 2
	if sub := (op>>7)&1 != 0; sub {
		nn = -nn
	}

	c.Reg.R[SP] = uint32(int(c.Reg.R[SP]) + nn)
}

func (c *Cpu) ThumbLongBranch(op uint16) {
	c.Reg.R[LR] = c.Reg.R[PC] + uint32(int32(uint32(op&0x7FF)<<21)>>9)
}

func (c *Cpu) ThumbShortLongBranch(op uint16) {
	r := &c.Reg.R
	ret := (r[PC] - 2) | 1
	r[PC] = r[LR] + (uint32(op&0x7FF) << 1)
	r[LR] = ret
	c.Reload = true
}

func (c *Cpu) ThumbLSSP(op uint16) {
	r := &c.Reg.R
	rd := (op >> 8) & 7
	addr := r[SP] + (uint32(op&0xFF) << 2)

	if ldr := (op>>11)&1 != 0; ldr {
		v := c.Read32(addr)
		r[rd] = bits.RotateLeft32(v, -int((addr&3)*8))
	} else {
		c.Write32(addr, r[rd])
	}
}

func (c *Cpu) ThumbBlock(op uint16) {
	var (
		r     = &c.Reg.R
		ldmia = (op>>11)&1 != 0
		rb    = (op >> 8) & 7
		rlist = op & 0xFF
	)

	if rlist == 0 {

		if ldmia {
			r[PC] = c.Read32Block(r[rb], NONSEQ)
			c.Reload = true
		} else {
			c.Write32(r[rb], r[PC]+2)
		}
		r[rb] += 0x40
		return
	}

	if !ldmia {

		var (
			count = uint32(bits.OnesCount16(rlist))
			first = 0
			addr  = r[rb]
		)

		// can this be sped up? Its just log2(rlist & -rlist)
		// first = int(math.Log2(float64(rlist & -rlist)))
		for reg := 7; reg >= 0; reg-- {
			if rlist&(1<<reg) != 0 {
				first = reg
			}
		}

		c.Write32Block(addr, r[first], NONSEQ)
		r[rb] = addr + count*4
		addr += 4

		for reg := first + 1; reg < 8; reg++ {
			if enabled := (rlist>>reg)&1 != 0; enabled {
				c.Write32Block(addr, r[reg], SEQ)
				addr += 4
			}
		}

	} else {

		addr := r[rb]

		seq := uint32(NONSEQ)

		for reg := range 8 {
			if enabled := (rlist>>reg)&1 != 0; enabled {
				r[reg] = c.Read32Block(addr, seq)
				addr += 4
				seq = SEQ
			}
		}

		c.Idle(1)

		if ^rlist&(1<<rb) != 0 {
			r[rb] = addr
		}
	}
}
