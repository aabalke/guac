package arm7

import (
	"fmt"
	"math/bits"
)

func (c *Cpu) DecodeArm(op uint32) {
	if cond := op >> 28; !c.Reg.CPSR.CheckCond(cond) {
		c.Seq = SEQ
		return
	}

	switch {
	case (op>>24)&0xF == 0xF:
		c.Exception(VEC_SWI, MODE_SWI)
	case IsBranch(op):
		c.B(op)
	case IsBranchExchange(op):
		c.BranchExchange(op)
	case IsSdt(op):
		c.Sdt(op)
	case IsBlock(op):
		c.Block(op)
	case IsHalf(op):
		c.Half(op)
	case IsUndefined(op):
		c.Exception(VEC_UNDEFINED, MODE_UND)
	case IsMrs(op):
		c.Mrs(op)
	case IsMsr(op):
		c.Msr(op)
	case IsSwp(op):
		c.Swp(op)
	case IsMul(op):
		c.Mul(op)
	case IsAlu(op):
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
func IsSwp(op uint32) bool {
	return IsOpFormat(
		op,
		0b1111_1011_0000_0000_1111_1111_0000,
		0b0001_0000_0000_0000_0000_1001_0000,
	)
}

//go:inline
func IsBlock(op uint32) bool {
	is := false

	is = is || IsOpFormat(
		op,
		0b1110_0001_0000_0000_0000_0000_0000,
		0b1000_0001_0000_0000_0000_0000_0000,
	)

	is = is || IsOpFormat(
		op,
		0b1110_0001_0000_0000_0000_0000_0000,
		0b1000_0000_0000_0000_0000_0000_0000,
	)

	return is
}

//go:inline
func IsHalf(op uint32) bool {
	is := false

	is = is || IsOpFormat(
		op,
		0b1110_0000_0000_0000_0000_1101_0000,
		0b0000_0000_0000_0000_0000_1101_0000,
	)

	is = is || IsOpFormat(
		op,
		0b1110_0000_0000_0000_0000_1011_0000,
		0b0000_0000_0000_0000_0000_1011_0000,
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
func IsAlu(op uint32) bool {
	return IsAluImm(op) || IsAluShiftImm(op) || IsAluShiftReg(op)
}

//go:inline
func IsBranchExchange(op uint32) bool {
	return IsOpFormat(
		op,
		0b1111_1111_1111_1111_1111_1101_0000,
		0b0001_0010_1111_1111_1111_0001_0000,
	)
}

//go:inline
func IsBranch(op uint32) bool {
	return IsOpFormat(
		op,
		0b1110_0000_0000_0000_0000_0000_0000,
		0b1010_0000_0000_0000_0000_0000_0000,
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
func IsUndefined(op uint32) bool {
	return IsOpFormat(
		op,
		0b1110_0000_0000_0000_0000_0001_0000,
		0b0110_0000_0000_0000_0000_0001_0000,
	)
}

//go:inline
func IsSdt(op uint32) bool {
	is := false
	is = is || IsOpFormat(
		op,
		0b1100_0001_0000_0000_0000_0000_0000,
		0b0100_0001_0000_0000_0000_0000_0000,
	)
	is = is || IsOpFormat(
		op,
		0b1100_0001_0000_0000_0000_0000_0000,
		0b0100_0000_0000_0000_0000_0000_0000,
	)

	return is
}

//go:inline
func IsMrs(op uint32) bool {
	return IsOpFormat(
		op,
		0b1111_1011_1111_0000_1111_1111_1111,
		0b0001_0000_1111_0000_0000_0000_0000,
	)
}

//go:inline
func IsMsr(op uint32) bool {
	return IsOpFormat(
		op,
		0b1101_1011_0000_1111_0000_0000_0000,
		0b0001_0010_0000_1111_0000_0000_0000,
	)
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
		op2   uint32
	)

	if imm := (op>>25)&1 != 0; imm {

		op2 = op & 0xFF

		if shift := ((op >> 8) & 0xF) << 1; shift != 0 {
			carry = (op2>>(shift-1))&1 != 0
			op2 = bits.RotateLeft32(op&0xFF, -int(shift))
		}

	} else {
		op2 = c.getShiftedAluReg(op, &carry)

		if regShift := (op>>4)&1 != 0; regShift && rn == PC {
			rnv += 4
		}
	}

	switch inst := (op >> 21) & 0xF; {
	case inst == MOV:
		res := op2
		r[rd] = res

		if set := (op>>20)&1 != 0; set {
			cpsr.C = carry
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
			case TST, TEQ:
				cpsr.C = carry
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
			cpsr.C = carry
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
			if c.Reg.CPSR.C {
				res++
			}
		case SBC:
			res = uint64(rnv) - uint64(op2) - 1
			if c.Reg.CPSR.C {
				res++
			}
		case RSC:
			res = uint64(op2) - uint64(rnv) - 1
			if c.Reg.CPSR.C {
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
			c.Reload = true
		}
	}
}

func (c *Cpu) getShiftedAluReg(op uint32, carry *bool) uint32 {
	// https://iitd-plos.github.io/col718/ref/arm-instructionset.pdf
	var (
		rm  = op & 0xF
		op2 = c.Reg.R[rm]
	)

	if imm := (op>>4)&1 == 0; imm {
		shift := (op >> 7) & 0x1F

		return c.ShiftImm((op>>5)&3, op2, shift, carry)
	}

	rs := (op >> 8) & 0xF
	shift := c.Reg.R[rs] & 0xFF
	shType := (op >> 5) & 3

	if rm == PC {
		op2 += 4
	}

	op2 = c.ShiftReg(op2, shift, shType, carry)

	return op2
}

func (c *Cpu) ShiftImm(sType, v, shift uint32, carry *bool) uint32 {
	switch sType {
	case LSL:

		if shift != 0 {
			*carry = v&(1<<(32-shift)) != 0
			v <<= shift
		}

	case LSR, ASR:

		if shift == 0 {
			shift = 32
		}

		*carry = v&(1<<(shift-1)) != 0

		if sType == ASR {
			v = uint32(int32(v) >> shift)
		} else {
			v >>= shift
		}

	case ROR:

		if shift == 0 {

			c := v&1 != 0

			v >>= 1

			if *carry {
				v |= 0x8000_0000
			}

			*carry = c

		} else {
			*carry = (v>>((shift-1)&31))&1 != 0
			v = bits.RotateLeft32(v, -int(shift))
		}

	}
	return v
}

func (c *Cpu) ShiftReg(v, shift, shType uint32, carry *bool) uint32 {
	c.Idle(1)

	if regZero := shift != 0; regZero {
		switch shType {
		case LSL:
			*carry = v&(1<<(32-shift)) != 0
			v <<= shift

		case LSR:
			*carry = v&(1<<(shift-1)) != 0
			v >>= shift

		case ASR:
			*carry = v&(1<<min(31, shift-1)) != 0
			v = uint32(int32(v) >> shift)

		case ROR:
			*carry = (v>>((shift-1)&31))&1 != 0
			v = bits.RotateLeft32(v, -int(shift))
		}
	}

	return v
}

const (
	MUL   = 0b000
	MLA   = 0b001 //
	UMAAL = 0b010
	UMULL = 0b100
	UMLAL = 0b101 //
	SMULL = 0b110
	SMLAL = 0b111 //
)

func (c *Cpu) Mul(op uint32) {
	var (
		set  = (op>>20)&1 != 0
		rd   = (op >> 16) & 0xF
		rn   = (op >> 12) & 0xF
		rs   = (op >> 8) & 0xF
		rm   = (op >> 0) & 0xF
		r    = &c.Reg.R
		cpsr = &c.Reg.CPSR
		add  = (op>>21)&1 != 0
	)

	switch inst := (op >> 21) & 0xF; inst {
	case MUL, MLA:

		res := r[rm] * r[rs]

		c.Idle(idleMul(r[rs], true))

		if add {
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

	case UMAAL:
		panic("unsupported umaal instruction")

	case UMULL, UMLAL:

		c.Idle(idleMul(r[rs], false) + 1)
		res := uint64(r[rm]) * uint64(r[rs])

		if add {
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

	case SMULL, SMLAL:

		c.Idle(idleMul(r[rs], true) + 1)

		res := int64(int32(r[rm])) * int64(int32(r[rs]))
		if add {
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
	}
}

func (c *Cpu) Sdt(op uint32) {
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

		offset uint32
	)

	if reg {

		if (op>>4)&1 != 0 {
			panic("malformed single data transfer reg")
		}

		carry := c.Reg.CPSR.C
		offset = c.ShiftImm((op>>5)&3, r[op&0xF], (op>>7)&0x1F, &carry)

	} else {
		offset = op & 0xFFF
	}

	addr := r[rn]
	post := addr

	if up {
		post += offset
	} else {
		post -= offset
	}

	if pre {
		addr = post
	}

	v := r[rd]

	if wb {
		r[rn] = post
	}

	if load {
		if byte {
			r[rd] = c.Read8(addr)
		} else {

			v := c.Read32(addr)
			is := ((addr & 3) * 8) & 0x1F
			r[rd] = bits.RotateLeft32(v, -int(is))

			if rd == PC {
				c.ToggleThumb() // this is arm9 - not sure if arm7
			}
		}
	} else {

		if rd == PC {
			v += 4
		}

		if byte {
			c.Write8(addr, uint8(v))
		} else {
			c.Write32(addr, v)
		}
	}
}

func (c *Cpu) B(op uint32) {
	r := &c.Reg.R

	if link := (op>>24)&1 != 0; link {
		r[LR] = r[PC] - 4
	}

	r[PC] += uint32((int32(op) << 8) >> 6)
	c.Reload = true
}

const (
	INST_BX = iota + 1
	INST_BXJ
	INST_BLX
)

func (c *Cpu) BranchExchange(op uint32) {
	switch inst := (op >> 4) & 0xF; inst {
	case INST_BX:
		c.Reg.R[PC] = c.Reg.R[op&0xF]
		c.ToggleThumb()

	case INST_BXJ:
		panic("unsupported bxj instruction")
	case INST_BLX:
		panic("unsupported arm7 blx instruction")
	}
}

const (
	RESERVED = 0
	STRH     = 1
	LDRD     = 2
	STRD     = 3
	LDRH     = 1
	LDRSB    = 2
	LDRSH    = 3
)

func (c *Cpu) Half(op uint32) {
	var (
		r    = &c.Reg.R
		rn   = (op >> 16) & 0xF
		rd   = (op >> 12) & 0xF
		pre  = (op>>24)&1 != 0
		load = (op>>20)&1 != 0
		inst = (op >> 5) & 3
		wb   = (op>>21)&1 != 0 || !pre
		addr = r[rn]
		post = addr

		offset uint32
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

	if pre {
		addr = post
	}

	v := r[rd]

	if wb {
		r[rn] = post
	}

	if !load {

		if inst != STRH {
			panic("unsupported arm7 instruction (ldrd, strd, reserved)")
		}

		c.Write16(addr, uint16(v))
		return
	}

	switch inst {
	case LDRH:
		v := uint32(c.Read16(addr))
		r[rd] = bits.RotateLeft32(v, -int((addr&1)*8))
	case LDRSB:
		r[rd] = uint32(int32(int8(c.Read8(addr))))

	case LDRSH:
		if misaligned := addr&1 != 0; misaligned {
			r[rd] = uint32(int32(int8(c.Read8(addr))))
		} else {
			r[rd] = uint32(int32(int16(c.Read16(addr))))
		}
	default:
		panic("unsupported arm7 instruction (ldrd, strd, reserved)")
	}
}

func (c *Cpu) Mrs(op uint32) {
	r := &c.Reg.R
	rd := (op >> 12) & 0xF

	if spsr := (op>>22)&1 != 0; spsr {
		r[rd] = c.GetSPSR(c.Reg.CPSR.Mode)
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

func (c *Cpu) Msr(op uint32) {
	r := &c.Reg.R

	var v uint32
	if imm := (op>>25)&1 != 0; imm {
		shift := ((op >> 8) & 0xF) << 1
		v = bits.RotateLeft32(op&0xFF, -int(shift))
	} else {
		v = r[op&0xF]
	}

	var mask uint32
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

	spsrFlag := (op>>22)&1 != 0
	c.DoMsrModeSwitch(spsrFlag, v, mask)
}

func (c *Cpu) DoMsrModeSwitch(spsrFlag bool, v, mask uint32) {
	curr := c.Reg.CPSR.Mode

	secMask := PRIV_MASK
	if curr == MODE_USR {
		secMask = USR_MASK
	}

	if spsrFlag {

		secMask |= STATE_MASK
		mask &= secMask

		var spsr uint32
		if curr == MODE_USR || curr == MODE_SYS {
			spsr = c.Reg.CPSR.Get()
		} else {
			spsr = c.Reg.SPSR[ModeBank(curr)].Get()
		}

		spsr = (spsr &^ mask) | (v & mask)
		c.Reg.SPSR[ModeBank(curr)].Set(spsr)
		return
	}

	mask &= secMask

	cpsr := c.Reg.CPSR.Get()
	cpsr = (cpsr &^ mask) | (v & mask)
	c.Reg.CPSR.Set(cpsr)

	next := CpuMode(v&0x1F | 0x10)

	if ModeBank(curr) != ModeBank(next) {

		if curr == MODE_USR {
			panic("user mode msr")
		}

		c.ModeSwitch(curr, next)
	}
}

func (c *Cpu) Swp(op uint32) {
	var (
		r   = &c.Reg.R
		rd  = (op >> 12) & 0xF
		rmv = r[op&0xF]
		rnv = r[(op>>16)&0xF]
	)

	if isByte := (op>>22)&1 != 0; isByte {
		r[rd] = c.Read8(rnv)
		c.Write8(rnv, uint8(rmv))
	} else {
		v := c.Read32(rnv)
		r[rd] = bits.RotateLeft32(v, -int((rnv&3)*8))
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
		first = bits.TrailingZeros32(rlist)
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
			if i == PC {
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
	c.Reload = true

	if !psr {
		return
	}

	c.DoLdmLoadSwitch()
}

func (c *Cpu) DoLdmLoadSwitch() {
	var (
		curr = c.Reg.CPSR.Mode
		spsr = c.Reg.SPSR[ModeBank(curr)]
		next = spsr.Mode
	)

	if curr == MODE_USR {
		panic("user mode ldm^")
	}

	c.Reg.CPSR = spsr
	c.ModeSwitch(curr, next)
}
