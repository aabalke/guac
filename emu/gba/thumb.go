package gba

import (
	"fmt"
	"math/bits"
)

func (c *Cpu) DecodeTHUMB(op uint16) {
	switch {
	case IsthumbSWI(op):
		c.Exception(VEC_SWI, MODE_SWI)
	case IsThumbAddSub(op):
		c.ThumbAddSub(op)
	case IsThumbShift(op):
		c.ThumbShifted(op)
	case IsThumbImm(op):
		c.ThumbImm(op)
	case IsThumbAlu(op):
		c.ThumbAlu(op)
	case IsThumbHiReg(op):
		c.HiRegBX(op)
	case IsLSHalf(op):
		c.ThumbLSHalf(op)
	case IsThumbSdt(op):
		c.ThumbSdt(op)
	case IsLPC(op):
		c.ThumbLPC(op)
	case IsLSImm(op):
		c.ThumbLSImm(op)
	case IsPushPop(op):
		c.ThumbPushPop(op)
	case IsRelative(op):
		c.ThumbRelative(op)
	case IsThumbB(op):
		c.ThumbB(op)
	case IsJumpCall(op):
		c.ThumbJumpCalls(op)
	case IsStack(op):
		c.ThumbStack(op)
	case IsLongBranch(op):
		c.ThumbLongBranch(op)
	case IsShortLongBranch(op):
		c.ThumbShortLongBranch(op)
	case IsLSSP(op):
		c.ThumbLSSP(op)
	case IsMulti(op):
		c.ThumbBlock(op)
	default:
		r := &c.Reg.R
		panic(fmt.Sprintf("Unable to Decode ARM false %04X, at PC %08X\n", op, r[PC]))
	}
}

//go:inline
func IsThumbOpFormat(op, mask, fmt uint16) bool {
	return op&mask == fmt
}

//go:inline
func IsThumbShift(op uint16) bool {
	return IsThumbOpFormat(
		op,
		0b1110_0000_0000_0000,
		0b0000_0000_0000_0000,
	)
}

//go:inline
func IsThumbAddSub(op uint16) bool {
	return IsThumbOpFormat(
		op,
		0b1111_1000_0000_0000,
		0b0001_1000_0000_0000,
	)
}

//go:inline
func IsThumbImm(op uint16) bool {
	return IsThumbOpFormat(
		op,
		0b1110_0000_0000_0000,
		0b0010_0000_0000_0000,
	)
}

//go:inline
func IsThumbAlu(op uint16) bool {
	return IsThumbOpFormat(
		op,
		0b1111_1100_0000_0000,
		0b0100_0000_0000_0000,
	)
}

//go:inline
func IsThumbHiReg(op uint16) bool {
	return IsThumbOpFormat(
		op,
		0b1111_1100_0000_0000,
		0b0100_0100_0000_0000,
	)
}

//go:inline
func IsLSHalf(op uint16) bool {
	return IsThumbOpFormat(
		op,
		0b1111_0000_0000_0000,
		0b1000_0000_0000_0000,
	)
}

//go:inline
func IsThumbSdt(op uint16) bool {
	return IsThumbOpFormat(
		op,
		0b1111_0000_0000_0000,
		0b0101_0000_0000_0000,
	)
}

//go:inline
func IsLPC(op uint16) bool {
	return IsThumbOpFormat(
		op,
		0b1111_1000_0000_0000,
		0b0100_1000_0000_0000,
	)
}

//go:inline
func IsLSImm(op uint16) bool {
	return IsThumbOpFormat(
		op,
		0b1110_0000_0000_0000,
		0b0110_0000_0000_0000,
	)
}

//go:inline
func IsPushPop(op uint16) bool {
	return IsThumbOpFormat(
		op,
		0b1111_0110_0000_0000,
		0b1011_0100_0000_0000,
	)
}

//go:inline
func IsRelative(op uint16) bool {
	return IsThumbOpFormat(
		op,
		0b1111_0000_0000_0000,
		0b1010_0000_0000_0000,
	)
}

//go:inline
func IsJumpCall(op uint16) bool {
	return IsThumbOpFormat(
		op,
		0b1111_0000_0000_0000,
		0b1101_0000_0000_0000,
	)
}

//go:inline
func IsThumbB(op uint16) bool {
	return IsThumbOpFormat(
		op,
		0b1111_1000_0000_0000,
		0b1110_0000_0000_0000,
	)
}

//go:inline
func IsStack(op uint16) bool {
	return IsThumbOpFormat(
		op,
		0b1111_1111_0000_0000,
		0b1011_0000_0000_0000,
	)
}

//go:inline
func IsLongBranch(op uint16) bool {
	return IsThumbOpFormat(
		op,
		0b1111_1000_0000_0000,
		0b1111_0000_0000_0000,
	)
}

//go:inline
func IsShortLongBranch(op uint16) bool {
	return IsThumbOpFormat(
		op,
		0b1111_1000_0000_0000,
		0b1111_1000_0000_0000,
	)
}

//go:inline
func IsLSSP(op uint16) bool {
	return IsThumbOpFormat(
		op,
		0b1111_0000_0000_0000,
		0b1001_0000_0000_0000,
	)
}

//go:inline
func IsMulti(op uint16) bool {
	return IsThumbOpFormat(
		op,
		0b1111_0000_0000_0000,
		0b1100_0000_0000_0000,
	)
}

//go:inline
func IsthumbSWI(op uint16) bool {
	return IsThumbOpFormat(
		op,
		0b1111_1111_0000_0000,
		0b1101_1111_0000_0000,
	)
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

func (cpu *Cpu) ThumbAlu(op uint16) {
	var (
		r    = &cpu.Reg.R
		cpsr = &cpu.Reg.CPSR

		inst = (op >> 6) & 0xF
		rsv  = r[(op>>3)&0x7]
		rd   = op & 0x7
		rdv  = r[rd]

		res uint64
	)

	switch inst {
	case THUMB_MUL:
		cpu.idle(idleMul(rdv, true))

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

	case THUMB_LSL:

		cpu.idle(1)

		rsv &= 0xFF

		if rsv > 32 {
			res = 0
			if rsv != 0 {
				cpsr.C = false
			}
		} else {
			res = uint64(rdv) << rsv
			if rsv != 0 {
				cpsr.C = rdv&(1<<(32-rsv)) != 0
			}
		}
		r[rd] = uint32(res)

	case THUMB_LSR:
		cpu.idle(1)
		rsv &= 0xFF

		res = uint64(rdv) >> rsv
		r[rd] = uint32(res)

		if rsv != 0 {
			cpsr.C = rdv&(1<<(rsv-1)) != 0
		}

	case THUMB_ASR:
		cpu.idle(1)
		rsv &= 0xFF

		if rsv > 32 {
			rsv = 32
		}

		if rsv != 0 {
			cpsr.C = rdv&(1<<(rsv-1)) != 0
		}

		res = uint64(int32(rdv) >> rsv)
		r[rd] = uint32(res)

	case THUMB_ROR:
		cpu.idle(1)
		rsv &= 0xFF

		if rsv != 0 {
			cpsr.C = (rdv>>((rsv-1)%32))&1 != 0
		}

		res = uint64(bits.RotateLeft32(rdv, -int(rsv)))
		r[rd] = uint32(res)
	}

	cpsr.N = (res>>31)&1 != 0
	cpsr.Z = uint32(res) == 0
}

const (
	HI_ADD = 0
	HI_CMP = 1
	HI_MOV = 2
	HI_BX  = 3
)

func (cpu *Cpu) HiRegBX(op uint16) {
	var (
		r    = &cpu.Reg.R
		cpsr = &cpu.Reg.CPSR

		inst = (op >> 8) & 0b11
		mSBd = (op>>7)&1 != 0
		mSBs = (op>>6)&1 != 0
		rs   = (op >> 3) & 0xF
		rd   = op & 0x7
	)

	if inst != 3 && mSBd {
		rd |= 0b1000
	}

	if mSBs {
		rs |= 0b1000
	}

	switch inst {
	case HI_ADD:

		rsv := r[rs]
		rdv := r[rd]

		//if rs == PC {
		//	rsv += 4
		//}

		res := uint64(rsv) + uint64(rdv)

		if rd == PC {
			r[rd] = uint32(res &^ 1)
			cpu.P.Reload = true
		} else {
			r[rd] = uint32(res)
		}
		return

	case HI_CMP:

		rsv := r[rs]
		rdv := r[rd]

		//if rs == PC {
		//	rsv += 4
		//}

		res := uint64(rdv) - uint64(rsv)

		//if rd != PC {
		//}

		cpsr.N = (uint32(res)>>31)&1 != 0
		cpsr.Z = uint32(res) == 0
		cpsr.C = res < 0x1_0000_0000
		cpsr.V = ((rdv^rsv)&(rdv^uint32(res)))>>31 != 0
		return

	case HI_MOV:

		if nop := rs == 8 && rd == 8; nop {
			return
		}

		rsv := r[rs]
		//if rs == PC {
		//	rsv += 4
		//}

		if rd == PC {
			r[rd] = rsv &^ 0b1
			cpu.P.Reload = true
			return
		}

		r[rd] = rsv

		return

	case HI_BX:

		if blx := mSBd; blx {
			if rs == LR {
				panic("Setup Thumb Blx LR, LR")
			}

			r[LR] = r[PC] // + 3
		}

		if thumb := r[rs]&1 != 0; thumb {
			r[PC] = r[rs] &^ 0b1
			cpu.P.Reload = true
			return
		}

		cpsr.T = false
		r[PC] = r[rs] &^ 0b11

		cpu.P.Reload = true

		return
	}
}

const (
	THUMB_ADD = iota
	THUMB_SUB
	THUMB_ADDImm
	THUMB_SUBImm
)

func (cpu *Cpu) ThumbAddSub(op uint16) {
	var (
		r    = &cpu.Reg.R
		cpsr = &cpu.Reg.CPSR

		inst = (op >> 9) & 0b11
		rsv  = uint32(r[(op>>3)&0x7])
		rd   = op & 0x7

		res uint64
	)

	var op2 uint32
	if reg := inst < 2; reg {
		op2 = uint32(r[(op>>6)&0x7])
	} else {
		op2 = uint32((op >> 6) & 0x7)
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

func (cpu *Cpu) ThumbImm(op uint16) {
	var (
		r    = &cpu.Reg.R
		cpsr = &cpu.Reg.CPSR
		inst = (op >> 11) & 0b11
		rd   = (op >> 8) & 0x7
		nn   = uint32(op & 0xFF)
		rdv  = r[rd]

		res uint64
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

func (cpu *Cpu) ThumbLSHalf(op uint16) {
	var (
		r = &cpu.Reg.R

		offset = uint32((op >> 6) & 0x1F << 1)
		addr   = r[(op>>3)&0x7] + offset
		rd     = op & 0x7
	)

	if ldr := (op>>11)&1 != 0; ldr {
		v := uint32(cpu.Read16(addr))
		is := (addr & 1) << 3
		r[rd] = bits.RotateLeft32(v, -int(is))
	} else {
		cpu.Write16(addr, uint16(r[rd]))
	}
}

const (
	THUMB_STRH = iota
	THUMB_LDSB
	THUMB_LDRH
	THUMB_LDSH
)

const (
	THUMB_STR_REG = iota
	THUMB_STRB_REG
	THUMB_LDR_REG
	THUMB_LDRB_REG
)

func (cpu *Cpu) ThumbSdt(op uint16) {
	var (
		r    = &cpu.Reg.R
		inst = (op >> 10) & 0b11
		rd   = op & 0x7
		addr = r[(op>>3)&0x7] + r[(op>>6)&0x7]
	)

	if signed := (op>>9)&1 != 0; signed {

		switch inst {
		case THUMB_STRH:

			cpu.Write16(addr, uint16(r[rd]))

		case THUMB_LDSB:

			// sign-expand byte value
			r[rd] = uint32(int32(int8(cpu.Read8(addr))))

		case THUMB_LDRH:

			v := cpu.Read16(addr)
			is := (addr & 1) << 3
			r[rd] = bits.RotateLeft32(v, -int(is))

		case THUMB_LDSH:

			// On ARM7 aka ARMv4 aka NDS7/GBA:
			// LDRSH Rd,[odd]  -->  LDRSB Rd,[odd];sign-expand BYTE value
			if misaligned := addr&1 != 0; misaligned {
				// sign-expand byte value
				r[rd] = uint32(int32(int8(cpu.Read8(addr))))
			} else {
				// sign-expand half value
				r[rd] = uint32(int32(int16(cpu.Read16(addr))))
			}

		}

		return
	}

	switch inst {
	case THUMB_STR_REG:
		cpu.Write32(addr, r[rd])
	case THUMB_STRB_REG:
		cpu.Write8(addr, uint8(r[rd]))
	case THUMB_LDR_REG:
		v := cpu.Read32(addr)
		is := (addr & 3) << 3
		r[rd] = bits.RotateLeft32(v, -int(is))
	case THUMB_LDRB_REG:
		r[rd] = cpu.Read8(addr)
	}
}

func (cpu *Cpu) ThumbLPC(op uint16) {
	var (
		r    = &cpu.Reg.R
		rd   = (op >> 8) & 0x7
		nn   = uint32(op&0xFF) << 2
		addr = (r[PC] &^ 0b11) + nn
	)

	r[rd] = cpu.Read32(addr)
}

const (
	THUMB_STR_IMM = iota
	THUMB_LDR_IMM
	THUMB_STRB_IMM
	THUMB_LDRB_IMM
)

func (cpu *Cpu) ThumbLSImm(op uint16) {
	var (
		r = &cpu.Reg.R

		inst = (op >> 11) & 0b11
		rd   = op & 0x7
		rb   = (op >> 3) & 0x7
		nn   = uint32(op>>6) & 0x1F
	)

	switch inst {
	case THUMB_STR_IMM:
		addr := r[rb] + (nn << 2)
		cpu.Write32(addr, r[rd])
	case THUMB_LDR_IMM:
		addr := r[rb] + (nn << 2)
		v := cpu.Read32(addr)
		is := (addr & 3) << 3
		r[rd] = bits.RotateLeft32(v, -int(is))

	case THUMB_STRB_IMM:
		addr := r[rb] + nn
		cpu.Write8(addr, uint8(r[rd]))
	case THUMB_LDRB_IMM:
		addr := r[rb] + nn
		r[rd] = uint32(cpu.Read8(addr))
	}
}

func (cpu *Cpu) ThumbPushPop(op uint16) {
	var (
		r     = &cpu.Reg.R
		pclr  = (op>>8)&1 != 0
		rlist = op & 0xFF
		pop   = (op>>11)&1 != 0
	)

	// thank you nano
	if rlist == 0 && !pclr {
		if pop {
			r[PC] = cpu.Read32(r[SP])
			cpu.P.Reload = true
			r[SP] += 0x40
		} else {
			// alyosha test fails this.
			// i think it is timing related
			r[SP] -= 0x40
			cpu.Write32(r[SP], r[PC])
		}

		return
	}

	seq := false

	reg := 0
	if !pop {
		reg = 7
	}

	if !pop && pclr {
		r[SP] -= 4
		cpu.Write32(r[SP], r[14])
	}

	for range 8 {
		if disabled := (rlist>>reg)&1 == 0; disabled {
			if pop {
				reg++
			} else {
				reg--
			}
			continue
		}

		if pop {
			r[reg] = cpu.Read32Block(r[SP], seq)
			r[SP] += 4
		} else {
			r[SP] -= 4
			cpu.Write32Block(r[SP], r[reg], seq)
		}

		seq = true

		if pop {
			reg++
		} else {
			reg--
		}
	}

	if pop && pclr {

		r[PC] = cpu.Read32(r[SP])

		r[PC] &^= 1
		r[SP] += 4

		cpu.idle(1)

		cpu.P.Reload = true
	}
}

func (cpu *Cpu) ThumbRelative(op uint16) {
	var (
		r  = &cpu.Reg.R
		rd = (op >> 8) & 0x7
		nn = uint32(op&0xFF) << 2
	)

	if isSP := (op>>11)&1 != 0; isSP {
		r[rd] = r[SP] + nn

		return
	}

	r[rd] = (r[PC] &^ 0b11) + nn
}

func (cpu *Cpu) ThumbJumpCalls(op uint16) {
	r := &cpu.Reg.R

	if !cpu.Reg.CPSR.CheckCond(uint32(op>>8) & 0xF) {
		return
	}

	nn := int(int8(op&0xFF)) << 1
	r[PC] = uint32(int(r[PC]) + nn)
	cpu.P.Reload = true
}

func (cpu *Cpu) ThumbB(op uint16) {
	r := &cpu.Reg.R

	offset := int16((op&0x7FF)<<5) >> 4
	r[15] += uint32(offset)
	cpu.P.Reload = true
}

func (cpu *Cpu) ThumbShifted(op uint16) {
	var (
		cpsr = &cpu.Reg.CPSR
		r    = &cpu.Reg.R

		shType = (op >> 11) & 0b11
		is     = (op >> 6) & 0x1F
		rsv    = r[(op>>3)&0x7]
		rd     = op & 0x7

		res uint32
	)

	switch shType {
	case LSL:

		switch {
		case is == 0:
			res = rsv
		case is > 32:
			res = 0
			cpsr.C = false
		default:
			res = rsv << is
			cpsr.C = rsv&(1<<(32-is)) != 0
		}

	case LSR:

		if is == 0 {
			is = 32
		}

		cpsr.C = rsv&(1<<(is-1)) != 0
		res = rsv >> is

	case ASR:

		if (is == 0) || is > 32 {
			is = 32
		}

		cpsr.C = rsv&(1<<(is-1)) != 0
		res = uint32(int32(rsv) >> is)
	}

	cpsr.N = (res>>31)&1 != 0
	cpsr.Z = uint32(res) == 0

	r[rd] = res
}

func (cpu *Cpu) ThumbStack(op uint16) {
	r := &cpu.Reg.R
	nn := int(op&0x7F) << 2

	if sub := (op>>7)&1 != 0; sub {
		nn = -nn
	}

	r[SP] = uint32(int(r[SP]) + nn)
}

func (cpu *Cpu) ThumbLongBranch(op uint16) {
	r := &cpu.Reg.R
	offset := int32(uint32(op&0x7FF)<<21) >> 9
	r[14] = r[15] + uint32(offset)

	//const shift = 32 - 23 // 22 is bits, + 1 for * 2
	//var (
	//	r = &cpu.Reg.R
	//	//op2 = cpu.Read16(cpu.P.Execute.Addr + 2)
	//	op2 = cpu.P.Decode.Op
	//	hi  = uint32(op & 0x7FF)
	//	lo  = uint32(op2 & 0x7FF)
	//	nn  = int32(((hi<<12)|(lo<<1))<<shift) >> shift
	//)

	//r[LR] = (r[PC] &^ 1) + 1
	//r[PC] = uint32(int32(r[PC]) + nn)

	//if exc := (op2>>12)&1 == 0; exc {
	//	cpu.ToggleThumb()
	//}
	//cpu.P.Reload = true
}

func (cpu *Cpu) ThumbShortLongBranch(op uint16) {
	r := &cpu.Reg.R

	nn := uint32(op&0x7FF) << 1
	r[15], cpu.P.Reload = r[14]+nn, true
	r[14] = cpu.P.Decode.Addr | 1

	// Using only the 2nd half of BL as "BL LR+imm" is possible
	// (for example, Mario Golf Advance Tour for GBA uses op F800h as "BL LR+0").
	// BL LR + nn
	// bottom half never signed?

	//nn := uint32(op&0x7FF) << 1
	//tmpLR := r[LR]
	//r[LR] = ((r[PC] + 2) &^ 0b1) + 1
	//r[PC] = (tmpLR + nn) &^ 0b1
	//cpu.P.Reload = true
}

func (cpu *Cpu) ThumbLSSP(op uint16) {
	r := &cpu.Reg.R
	rd := (op >> 8) & 0x7
	addr := r[SP] + (uint32(op&0xFF) << 2)

	if ldr := (op>>11)&1 != 0; ldr {
		v := cpu.Read32(addr)
		is := (addr & 3) << 3
		r[rd] = bits.RotateLeft32(v, -int(is))
	} else {
		cpu.Write32(addr, r[rd])
	}
}

func (cpu *Cpu) ThumbBlock(op uint16) {
	var (
		r     = &cpu.Reg.R
		ldmia = (op>>11)&1 != 0
		rb    = (op >> 8) & 7
		rlist = uint32(op & 0xFF)
		addr  = r[rb]
		seq   = false
	)

	if !ldmia {

		regCount := uint32(bits.OnesCount32(rlist))
		matchingValue := uint32(0)
		matchingAddr := uint32(0) // rn during regs
		smallest := (rlist & -rlist) == 1<<rb
		matchingRb := (rlist>>rb)&1 == 1

		rbIdx := uint32(0)
		count := uint32(0)

		if rlist == 0 {
			cpu.Write32(r[rb], r[PC]+2)
			r[rb] += 0x40

			return
		}

		for reg := range 8 {
			if disabled := (rlist>>reg)&1 == 0; disabled {
				continue
			}

			if reg == int(rb) {
				cpu.Write32Block(addr, r[reg], seq)
				matchingValue = r[reg] + 4
				matchingAddr = addr
				rbIdx = regCount - count
				r[rb] += 4
				addr += 4
				seq = true
				continue
			}

			cpu.Write32Block(addr, r[reg], seq)

			r[rb] += 4
			addr += 4
			seq = true
		}

		if smallest {
			v := cpu.Read32(addr)
			cpu.Write32(r[rb], v-(regCount*2))

			return
		}

		if matchingRb {
			cpu.Write32(matchingAddr, matchingValue+(rbIdx*2))

			return
		}

		return
	}

	rbValue := r[rb]
	matchingRb := false

	if rlist == 0 {
		r[rb] += 0x40
		r[PC] += 4
		cpu.P.Reload = true

		return
	}

	for reg := range 8 {
		if disabled := (rlist>>reg)&1 == 0; disabled {
			continue
		}

		r[reg] = cpu.Read32Block(addr, seq)

		if reg == int(rb) {
			matchingRb = true
			// do not remove this, needed for golden sun and others
			rbValue = r[rb]
		}

		r[rb] += 4
		addr += 4
		seq = true
	}

	cpu.idle(1)

	if matchingRb {
		r[rb] = rbValue
	}
}
