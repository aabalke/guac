package arm7

import (
	"fmt"
	"math/bits"
)

func (c *Cpu) DecodeARM(op uint32) {
	if cond := op >> 28; !c.Reg.CPSR.CheckCond(cond) {
		c.Seq = SEQ
		return
	}

	switch {
	case (op>>24)&0xF == 0xF:
		c.Exception(VEC_SWI, MODE_SWI)
	case IsB(op):
		c.B(op)
	case IsBX(op):
		c.BX(op)
	case IsSDT(op):
		c.Sdt(op)
	case IsBlock(op):
		c.Block(op)
	case IsHalf(op):
		c.Half(op)
	case IsUD(op):
		c.Exception(VEC_UNDEFINED, MODE_UND)
	case IsPSR(op):
		c.Psr(op)
	case IsSWP(op):
		c.Swp(op)
	case IsMul(op):
		c.Mul(op)
	case IsALU(op):
		c.Alu(op)

	default:
		panic(fmt.Sprintf("Unable to Decode ARM false %08X, at PC %08X\n", op, c.Reg.R[PC]))
	}
}

//go:inline
func IsOpFormat(op, mask, format uint32) bool {
	return op&mask == format
}

//go:inline
func IsSWP(op uint32) bool {
	return IsOpFormat(
		op,
		0b0000_1111_1011_0000_0000_1111_1111_0000,
		0b0000_0001_0000_0000_0000_0000_1001_0000,
	)
}

//go:inline
func IsBlock(op uint32) bool {
	is := false

	is = is || IsOpFormat(
		op,
		0b0000_1110_0001_0000_0000_0000_0000_0000,
		0b0000_1000_0001_0000_0000_0000_0000_0000,
	)

	is = is || IsOpFormat(
		op,
		0b0000_1110_0001_0000_0000_0000_0000_0000,
		0b0000_1000_0000_0000_0000_0000_0000_0000,
	)

	return is
}

//go:inline
func IsHalf(op uint32) bool {
	is := false

	is = is || IsOpFormat(
		op,
		0b0000_1110_0000_0000_0000_0000_1101_0000,
		0b0000_0000_0000_0000_0000_0000_1101_0000,
	)

	is = is || IsOpFormat(
		op,
		0b0000_1110_0000_0000_0000_0000_1011_0000,
		0b0000_0000_0000_0000_0000_0000_1011_0000,
	)

	return is
}

func IsAluImm(op uint32) bool {
	return IsOpFormat(
		op,
		0b1110_0000_0000_0000_0000_0000_0000,
		0b0010_0000_0000_0000_0000_0000_0000,
	)
}

func IsAluShiftReg(op uint32) bool {
	return IsOpFormat(
		op,
		0b1110_0000_0000_0000_0000_1001_0000,
		0b0000_0000_0000_0000_0000_0001_0000,
	)
}

func IsAluShiftImm(op uint32) bool {
	return IsOpFormat(
		op,
		0b1110_0000_0000_0000_0000_0001_0000,
		0b0000_0000_0000_0000_0000_0000_0000,
	)
}

//go:inline
func IsALU(op uint32) bool {
	return IsAluImm(op) || IsAluShiftImm(op) || IsAluShiftReg(op)
}

//go:inline
func IsBX(op uint32) bool {
	return IsOpFormat(
		op,
		0b0000_1111_1111_1111_1111_1111_1101_0000,
		0b0000_0001_0010_1111_1111_1111_0001_0000,
	)
}

//go:inline
func IsB(op uint32) bool {
	return IsOpFormat(
		op,
		0b0000_1110_0000_0000_0000_0000_0000_0000,
		0b0000_1010_0000_0000_0000_0000_0000_0000,
	)
}

//go:inline
func IsMul(op uint32) bool {
	return IsOpFormat(
		op,
		0b1110_0000_0000_0000_0000_1111_0000,
		0b0000_0000_0000_0000_0000_1001_0000,
	)
}

//go:inline
func IsUD(op uint32) bool {
	return IsOpFormat(
		op,
		0b0000_1110_0000_0000_0000_0000_0000_0000,
		0b0000_0110_0000_0000_0000_0000_0000_0000,
	)
}

//go:inline
func IsSDT(op uint32) bool {
	is := false
	is = is || IsOpFormat(
		op,
		0b0000_1100_0001_0000_0000_0000_0000_0000,
		0b0000_0100_0001_0000_0000_0000_0000_0000,
	)
	is = is || IsOpFormat(
		op,
		0b0000_1100_0001_0000_0000_0000_0000_0000,
		0b0000_0100_0000_0000_0000_0000_0000_0000,
	)

	return is
}

//go:inline
func IsPSR(op uint32) bool {
	is := false

	is = is || IsOpFormat(
		op,
		0b0000_1111_1011_1111_0000_1111_1111_1111,
		0b0000_0001_0000_1111_0000_0000_0000_0000,
	)

	is = is || IsOpFormat(
		op,
		0b0000_1101_1011_0000_1111_0000_0000_0000,
		0b0000_0001_0010_0000_1111_0000_0000_0000,
	)

	return is
}

const (
	LSL = iota
	LSR
	ASR
	ROR
)

const (
	AND = iota
	EOR
	SUB
	RSB
	ADD
	ADC
	SBC
	RSC
	TST
	TEQ
	CMP
	CMN
	ORR
	MOV
	BIC
	MVN
)

func (c *Cpu) Alu(op uint32) {
	var (
		r     = &c.Reg.R
		cpsr  = &c.Reg.CPSR
		rd    = (op >> 12) & 0xF
		rn    = (op >> 16) & 0xF
		carry = cpsr.C
		rnv   = r[rn]
		imm   = (op>>25)&1 != 0

		op2 uint32
	)

	if imm {

		ro := ((op >> 8) & 0xF) << 1
		op2 = bits.RotateLeft32(op&0xFF, -int(ro))

		if setCarry := ro != 0 && (op>>20)&1 != 0; setCarry {
			// I believe this matches
			// carry := (nn >> (ro-1)) & 1 != 0 // this line must be before op
			cpsr.C = (op2>>31)&1 != 0
		}

	} else {
		op2 = c.getShiftedAluReg(op)

		if regShift := (op>>4)&1 != 0; regShift && rn == PC {
			rnv += 4
		}
	}

	inst := (op >> 21) & 0xF

	switch {
	case inst == MOV:
		res := op2
		r[rd] = res

		if set := (op>>20)&1 != 0; set {
			cpsr.N = (uint32(res)>>31)&1 != 0
			cpsr.Z = uint32(res) == 0
		}

	case inst == SUB:
		res := uint64(rnv) - uint64(op2)
		r[rd] = uint32(res)
		if set := (op>>20)&1 != 0; set {
			cpsr.V = ((rnv^op2)&(rnv^uint32(res)))>>31 != 0
			cpsr.C = res < 0x1_0000_0000
			cpsr.N = (uint32(res)>>31)&1 != 0
			cpsr.Z = uint32(res) == 0
		}

	// test alu
	case inst >= 0b1000 && inst < 0b1100:

		var res uint64

		switch inst {
		case TST:
			res = uint64(rnv) & uint64(op2)
		case TEQ:
			res = uint64(rnv) ^ uint64(op2)
		case CMP:
			res = uint64(rnv) - uint64(op2)
		case CMN:
			res = uint64(rnv) + uint64(op2)
		}

		if set := (op>>20)&1 != 0; set {

			switch inst {
			case CMN:
				cpsr.V = ((^(rnv ^ op2))&(rnv^uint32(res)))>>31 != 0
				cpsr.C = res >= 0x1_0000_0000
			case CMP:
				cpsr.V = ((rnv^op2)&(rnv^uint32(res)))>>31 != 0
				cpsr.C = res < 0x1_0000_0000
			}

			cpsr.N = (uint32(res)>>31)&1 != 0
			cpsr.Z = uint32(res) == 0
		}

	// logical
	case inst&0b0110 == 0b0000 || inst&0b1100 == 0b1100:

		var res uint32

		switch inst {
		case AND:
			res = rnv & op2
		case EOR:
			res = rnv ^ op2
		case ORR:
			res = rnv | op2
		case BIC:
			res = rnv &^ op2
		case MVN:
			res = ^op2
		}

		r[rd] = res

		if set := (op>>20)&1 != 0; set {
			cpsr.N = (uint32(res)>>31)&1 != 0
			cpsr.Z = uint32(res) == 0
		}

	// arthmetic
	default:

		var res uint64

		switch inst {
		case RSB:
			res = uint64(op2) - uint64(rnv)
		case ADD:
			res = uint64(rnv) + uint64(op2)
		case ADC:
			res = uint64(rnv) + uint64(op2)
			if carry {
				res++
			}
		case SBC:
			res = uint64(rnv) - uint64(op2) - 1
			if carry {
				res++
			}
		case RSC:
			res = uint64(op2) - uint64(rnv) - 1
			if carry {
				res++
			}
		}

		r[rd] = uint32(res)

		if set := (op>>20)&1 != 0; set {

			switch inst {
			case ADD, ADC:
				cpsr.V = ((^(rnv ^ op2))&(rnv^uint32(res)))>>31 != 0
				cpsr.C = res >= 0x1_0000_0000
			case SBC:
				cpsr.V = ((rnv^op2)&(rnv^uint32(res)))>>31 != 0
				cpsr.C = res < 0x1_0000_0000
			case RSB, RSC:
				cpsr.V = ((rnv^op2)&(op2^uint32(res)))>>31 != 0
				cpsr.C = res < 0x1_0000_0000
			}

			cpsr.N = (uint32(res)>>31)&1 != 0
			cpsr.Z = uint32(res) == 0
		}
	}
	if rd := (op >> 12) & 0xF; rd == PC {
		inst := (op >> 21) & 0xF

		if op&(1<<20) != 0 {
			c.ExitException(c.Reg.CPSR.Mode)
		}

		if c.Reg.CPSR.T {
			c.Reg.R[PC] &^= 1
		} else {
			c.Reg.R[PC] &^= 3
		}
		if inst < 0b1000 || inst > 0b1011 {
			if c.Reg.CPSR.T {
				c.Reload16()
			} else {
				c.Reload32()
			}
		}
	}
}

func (c *Cpu) getShiftedAluReg(op uint32) uint32 {
	var (
		r = &c.Reg.R

		carry = c.Reg.CPSR.C

		shReg  = (op>>4)&1 != 0
		shType = (op >> 5) & 0b11

		inst     = (op >> 21) & 0xF
		logical  = inst&0b0110 == 0b0000 || inst&0b1100 == 0b1100
		setCarry = (op>>20)&1 != 0 && logical

		rm  = op & 0xF
		op2 = r[rm]

		shift uint32
	)

	if shReg {
		rs := (op >> 8) & 0xF
		shift = r[rs] & 0xFF

		c.Idle(1)

		if rm == PC {
			op2 += 4
		}

	} else {

		shift = (op >> 7) & 0x1F

		if special := shift == 0; special {
			switch shType {
			case LSL:
				return op2
			case LSR:
				c.Reg.CPSR.C = op2&0x8000_0000 != 0
				return 0
			case ASR:

				signed := op2&0x8000_0000 != 0

				if setCarry {
					c.Reg.CPSR.C = signed
				}

				if signed {
					return 0xFFFF_FFFF
				}

				return 0

			case ROR:

				c.Reg.CPSR.C = op2&1 != 0

				op2 >>= 1
				if carry {
					op2 |= 0x8000_0000
				}

				return op2
			}
		}
	}

	// https://iitd-plos.github.io/col718/ref/arm-instructionset.pdf

	if regZero := shift == 0; regZero {
		// op2 unchanges, carry is set to original carry (no change)
		return op2
	}

	switch shType {
	case LSL:

		switch {
		case shift > 32:
			op2 = 0
			carry = false
		default:
			carry = op2&(1<<(32-shift)) != 0
			op2 <<= shift
		}

	case LSR:

		switch {
		case shift > 32:
			op2 = 0
			carry = false
		case shift == 32:
			carry = op2&0x8000_0000 != 0
			op2 = 0
		default:
			carry = op2&(1<<(shift-1)) != 0
			op2 >>= shift
		}

	case ASR:

		switch {
		case shift >= 32:
			signed := op2&0x8000_0000 != 0
			carry = signed

			if signed {
				op2 = 0xFFFF_FFFF
			} else {
				op2 = 0x0
			}

		default:
			carry = op2&(1<<(shift-1)) != 0
			op2 = uint32(int32(op2) >> shift)
		}

	case ROR:

		if shift == 32 {
			// op2 unchanges
			carry = op2&0x8000_0000 != 0
		} else {
			carry = (op2>>((shift-1)&31))&1 != 0
			op2 = bits.RotateLeft32(op2, -int(shift))
		}
	}

	if setCarry {
		c.Reg.CPSR.C = carry
	}

	return op2
}

const (
	MUL   = 0b000
	MLA   = 0b001
	UMAAL = 0b010
	UMULL = 0b100
	UMLAL = 0b101
	SMULL = 0b110
	SMLAL = 0b111
)

func (c *Cpu) Mul(op uint32) {
	var (
		inst = (op >> 21) & 0xF
		set  = (op>>20)&1 != 0
		rd   = (op >> 16) & 0xF
		rn   = (op >> 12) & 0xF
		rs   = (op >> 8) & 0xF
		rm   = (op >> 0) & 0xF
		r    = &c.Reg.R
		cpsr = &c.Reg.CPSR
	)

	switch inst {
	case MUL, MLA:

		res := r[rm] * r[rs]

		c.Idle(idleMul(r[rs], true))

		if inst == MLA {
			res += r[rn]
			c.Idle(1)
		}

		r[rd] = res

		if set {
			cpsr.N = (uint32(res)>>31)&1 != 0
			cpsr.Z = uint32(res) == 0
			// FLAG_C "destroyed" ARM <5, ignored ARM >=5
			// cpsr.C = false
		}

		return

	case UMAAL:
		panic("unsupported umaal instruction")

	case UMULL, UMLAL:

		c.Idle(idleMul(r[rs], false) + 1)
		res := uint64(r[rm]) * uint64(r[rs])

		if inst == UMLAL {
			res += uint64(r[rd])<<32 | uint64(r[rn])
			c.Idle(1)
		}

		r[rd] = uint32(res >> 32)
		r[rn] = uint32(res)

		if set {
			cpsr.N = (res >> 63 & 1) != 0
			cpsr.Z = res == 0
			// FLAG_C "destroyed" ARM <5, ignored ARM >=5
			// need carry to pass mgba suite
			cpsr.C = false
			// FLAG_V maybe destroyed on ARM <5. ignored ARM <=5
		}
		return

	case SMULL, SMLAL:

		c.Idle(idleMul(r[rs], true) + 1)

		res := int64(int32(r[rm])) * int64(int32(r[rs]))
		if inst == SMLAL {
			res += int64(r[rd])<<32 | int64(r[rn])
			c.Idle(1)
		}

		r[rd] = uint32(res >> 32)
		r[rn] = uint32(res)

		if set {
			cpsr.N = (res>>63)&1 != 0
			cpsr.Z = res == 0
			cpsr.C = false
			// FLAG_C "destroyed" ARM <5, ignored ARM >=5
			// FLAG_V maybe destroyed on ARM <5. ignored ARM <=5
		}

		return

	}
}

func (c *Cpu) Sdt(op uint32) {
	if valid := (op>>26)&3 == 1; !valid {
		panic("Malformed Sdt Instruction")
	}

	var (
		r    = &c.Reg.R
		reg  = (op>>25)&1 != 0
		pre  = (op>>24)&1 != 0
		up   = (op>>23)&1 != 0
		byte = (op>>22)&1 != 0
		wb   = (op>>21)&1 != 0 || !pre
		load = (op>>20)&1 != 0
		rn   = (op >> 16) & 0xF
		rd   = (op >> 12) & 0xF

		offset, prev uint32
	)

	if reg {

		if (op>>4)&1 != 0 {
			panic("Malformed Single Data Transfer O_o")
		}

		shift := (op >> 7) & 0x1F
		sType := (op >> 5) & 3
		rm := op & 0xF

		switch sType {
		case LSL:

			offset = r[rm] << shift

		case LSR:

			if shift == 0 {
				shift = 32
			}

			offset = r[rm] >> shift

		case ASR:

			if shift == 0 {
				shift = 32
			}

			offset = uint32(int32(r[rm]) >> shift)

		case ROR:

			if shift == 0 {

				offset = r[rm] >> 1

				if c.Reg.CPSR.C {
					offset |= 0x8000_0000
				}
			} else {
				offset = bits.RotateLeft32(r[rm], -int(shift))
			}
		}

	} else {
		offset = op & 0xFFF
	}

	post := r[rn]

	if up {
		post += offset
	} else {
		post -= offset
	}

	if pre {
		prev = post
	} else {
		prev = r[rn]
	}

	if load {
		if byte {
			// DO NOT WORD ALIGN
			r[rd] = c.Read8(prev)
		} else {

			v := c.Read32(prev)
			is := ((prev & 3) << 3) & 0x1F
			r[rd] = bits.RotateLeft32(v, -int(is))

			if rd == PC {
				c.ToggleThumb() // this is arm9 - not sure if arm7
			}
		}
	} else {
		v := r[rd]
		if rd == PC {
			v += 4
		}

		if byte {
			c.Write8(prev, uint8(v))
		} else {
			c.Write32(prev, v)
		}
	}

	if wb && (!load || rn != rd) {
		r[rn] = post
	}
}

func (c *Cpu) B(op uint32) {
	r := &c.Reg.R

	if link := (op>>24)&1 != 0; link {
		r[14] = r[15] - 4
	}

	r[PC] += uint32((int32(op) << 8) >> 6)
	c.Reload32()
}

const (
	INST_BX  = 1
	INST_BXJ = 2
	INST_BLX = 3
)

func (c *Cpu) BX(op uint32) {
	var (
		r    = &c.Reg.R
		inst = (op >> 4) & 0xF
		rn   = op & 0xF
	)

	switch inst {
	case INST_BX:
		r[PC] = r[rn]
		c.ToggleThumb()

	case INST_BXJ:
		panic("Unsupported BXJ Instruction")
	case INST_BLX:

		panic("unsupported arm7 blx instruction")

	}
}

const (
	RESERVED = 0
	STRH     = 1
	LDRD     = 2
	STRD     = 3

	LDRH  = 1
	LDRSB = 2
	LDRSH = 3
)

func (c *Cpu) Half(op uint32) {
	var (
		r       = &c.Reg.R
		rn      = (op >> 16) & 0xF
		rd      = (op >> 12) & 0xF
		preFlag = (op>>24)&1 != 0
		load    = (op>>20)&1 != 0
		inst    = (op >> 5) & 0b11
		wb      = (op>>21)&1 != 0 || !preFlag
		rnv     = r[rn]
		post    = rnv

		pre, offset uint32
	)

	if imm := (op>>22)&1 != 0; imm {
		offset = (op & 0xF) | ((op >> 4) & 0xF0)
	} else {
		offset = r[op&0xF]
	}

	if up := (op>>23)&1 != 0; up {
		post += offset
	} else {
		post -= offset
	}

	if preFlag {
		pre = post
	} else {
		pre = rnv
	}

	if inst == RESERVED {
		panic("unsupported half (reserved)")
	}

	if !load {
		rdv := r[rd]

		if wb {
			r[rn] = post
		}

		switch inst {
		case STRH:
			c.Write16(pre, uint16(rdv))

		case LDRD:
			panic("unsupported arm7 ldrd instruction")
		case STRD:
			panic("unsupported arm7 strd instruction")
		}

		return
	}

	if wb {
		r[rn] = post
	}

	switch inst {
	case LDRH:
		//  LDRH Rd,[odd]   -->  LDRH Rd,[odd-1] ROR 8  ;read to bit0-7 and bit24-31
		v := uint32(c.Read16(pre))
		is := (pre & 1) << 3
		r[rd] = bits.RotateLeft32(v, -int(is))
	case LDRSB:
		// sign-expand byte value
		r[rd] = uint32(int32(int8(c.Read8(pre))))

	case LDRSH:
		// On ARM7 aka ARMv4 aka NDS7/GBA:
		// LDRSH Rd,[odd]  -->  LDRSB Rd,[odd];sign-expand BYTE value
		if misaligned := pre&1 != 0; misaligned {
			// sign-expand byte value
			r[rd] = uint32(int32(int8(c.Read8(pre))))
		} else {
			// sign-expand half value
			r[rd] = uint32(int32(int16(c.Read16(pre))))
		}
	}
}

func (c *Cpu) Psr(op uint32) {
	r := &c.Reg.R

	if msr := (op>>21)&1 != 0; msr {
		c.msr(op)
		return
	}

	rd := (op >> 12) & 0xF

	if spsr := (op>>22)&1 != 0; spsr {
		mode := c.Reg.CPSR.Mode
		r[rd] = c.Reg.SPSR[ModeBank[mode]].Get()
		return
	}

	mask := PRIV_MASK
	if c.Reg.CPSR.Mode == MODE_USR {
		mask = USR_MASK
	}

	r[rd] = uint32(c.Reg.CPSR.Get()) & mask
}

const (
	PRIV_MASK  uint32 = 0xF8FF_03DF
	USR_MASK   uint32 = 0xF8FF_0000
	STATE_MASK uint32 = 0x0100_0020
)

func (c *Cpu) msr(op uint32) {
	r := &c.Reg.R

	var v uint32
	if imm := (op>>25)&1 != 0; imm {
		shift := ((op >> 8) & 0xF) << 1
		v = bits.RotateLeft32(op&0xFF, -int(shift))

	} else {
		v = r[op&0xF]
	}

	mask := uint32(0)
	if C := (op>>16)&1 != 0; C {
		mask |= 0x0000_00FF
	}
	if X := (op>>17)&1 != 0; X {
		mask |= 0x0000_FF00
	}
	if S := (op>>18)&1 != 0; S {
		mask |= 0x00FF_0000
	}
	if F := (op>>19)&1 != 0; F {
		mask |= 0xFF00_0000
	}

	secMask := PRIV_MASK
	curr := c.Reg.CPSR.Mode

	if curr == MODE_USR {
		secMask = USR_MASK
	}

	spsrFlag := (op>>22)&1 != 0

	if spsrFlag {
		secMask |= STATE_MASK
	}

	mask &= secMask

	reg := &c.Reg

	if spsrFlag {

		var spsr uint32

		if curr == MODE_USR || curr == MODE_SYS {
			spsr = uint32(reg.CPSR.Get()) &^ mask
		} else {
			spsr = uint32(reg.SPSR[ModeBank[curr]].Get()) &^ mask
		}

		spsr |= v & mask
		reg.SPSR[ModeBank[curr]].Set(spsr)

		return
	}

	next := CpuMode(v&0x1F | 0x10)
	cpsr := uint32(reg.CPSR.Get()) &^ mask

	cpsr |= v & mask

	reg.CPSR.Set(cpsr)

	if skip := ModeBank[curr] == ModeBank[next]; skip {
		return
	}

	if curr == MODE_USR {
		panic("USER MODE MSR")
	}

	c.ModeSwitch(curr, next)
}

func (c *Cpu) Swp(op uint32) {
	var (
		r      = &c.Reg.R
		isByte = (op>>22)&1 != 0
		rn     = (op >> 16) & 0xF
		rd     = (op >> 12) & 0xF
		rm     = op & 0xF
		rmv    = r[rm]
		rnv    = r[rn]
	)

	if isByte {
		r[rd] = c.Read8(rnv)
		c.Write8(rnv, uint8(rmv))
	} else {
		v := c.Read32(rnv)
		is := (rnv & 3) << 3
		r[rd] = bits.RotateLeft32(v, -int(is))
		c.Write32(rnv, rmv)
	}
}

func (c *Cpu) Block(op uint32) {
	var (
		r          = &c.Reg.R
		rlist      = op & 0xFFFF
		rn         = (op >> 16) & 0xF
		pcIncluded = rlist&0x8000 != 0
		pre        = (op>>24)&1 != 0
		up         = (op>>23)&1 != 0
		psr        = (op>>22)&1 != 0
		wb         = (op>>21)&1 != 0
		load       = (op>>20)&1 != 0

		addr = r[rn]

		first int
		bytes uint32
	)

	if rlist == 0 {
		rlist = 0x8000
		first = 0xF
		bytes = 16 * 4
		pcIncluded = true
	} else {
		bytes = uint32(bits.OnesCount32(rlist)) * 4

		// can this be sped up? Its just log2(rlist & -rlist)
		// first = int(math.Log2(float64(rlist & -rlist)))
		for i := 0xF; i >= 0; i-- {
			if rlist&(1<<i) != 0 {
				first = i
			}
		}
	}

	mode := c.Reg.CPSR.Mode
	forceUser := psr && (mode != MODE_USR && mode != MODE_SYS) && (!load || !pcIncluded)

	if forceUser {
		c.ModeSwitch(mode, MODE_USR)
	}

	rnNew := addr

	// even when decrementing, cpu increments from "final" reg
	// see mgba https://mgba.io/2014/12/28/classic-nes/

	if up {
		rnNew += bytes
	} else {
		pre = !pre
		rnNew -= bytes
		addr -= bytes
	}

	seq := uint32(NONSEQ)

	for i := first; i < 0x10; i++ {
		if disabled := rlist&(1<<i) == 0; disabled {
			continue
		}

		if pre {
			addr += 4
		}

		if load {

			v := c.Read32Block(addr, seq)
			if wb && i == first {
				r[rn] = rnNew
			}

			r[i] = v

		} else {

			v := r[i]
			if i == 15 {
				v += 4
			}

			c.Write32Block(addr, v, seq)
			if wb && i == first {
				r[rn] = rnNew
			}

		}

		if !pre {
			addr += 4
		}

		seq = SEQ
	}

	if forceUser {
		c.ModeSwitch(MODE_USR, mode)
	}

	if !load {
		return
	}

	c.Idle(1)

	if !pcIncluded {
		return
	}

	if c.Reg.CPSR.T {
		c.Reload16()
	} else {
		c.Reload32()
	}

	if !psr {
		return
	}

	var (
		curr = c.Reg.CPSR.Mode
		spsr = c.Reg.SPSR[ModeBank[curr]]
		next = spsr.Mode
	)

	c.Reg.CPSR = spsr

	if curr == MODE_USR {
		panic("USER MODE LDM PC CHANGE")
	}

	c.ModeSwitch(curr, next)
}
