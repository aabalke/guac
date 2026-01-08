package arm9

import (
	"math/bits"
	"unsafe"

	"github.com/aabalke/guac/emu/cpu/arm9/cp15"
	"math"
)

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

func (cpu *Cpu) Alu(op uint32) {

	var (
		r     = &cpu.Reg.R
		cpsr  = &cpu.Reg.CPSR
		rd    = (op >> 12) & 0xF
		rn    = (op >> 16) & 0xF
		carry = cpsr.C
		rnv   = r[rn]
		imm   = (op>>25)&1 != 0

		op2 uint32
	)

	//compare := rd != PC
	//cpu.Jit.StartTest(op, compare, cpu.Jit.emitAlu)

	if imm {

		ro := ((op >> 8) & 0xF) << 1
		op2 = bits.RotateLeft32(op&0xFF, -int(ro))

		if setCarry := ro != 0 && (op>>20)&1 != 0; setCarry {
			// I believe this matches
			//carry := (nn >> (ro-1)) & 1 != 0 // this line must be before op
			cpsr.C = op2>>31 != 0
		}

		if rn == PC {
			rnv += 8
		}

	} else {

		op2 = cpu.getShiftedAluReg(op)

		if rn == PC {
			if regShift := (op>>4)&1 != 0; regShift {
				rnv += 12
			} else {
				rnv += 8
			}
		}
	}

	inst := (op >> 21) & 0xF

	switch {
	case inst == MOV:
		res := op2
		r[rd] = res

		if rd == PC {
			// not sure on this
			r[rd] &^= 0b1
		}

		if set := (op>>20)&1 != 0; set {

			rm := op & 0xF

			if swiExit := !imm && rd == PC && rm == LR; swiExit {
				cpu.ExitException(MODE_SWI)
				if r[PC]&1 != 0 {
					cpu.toggleThumb()
				}
			} else {
				cpsr.N = (uint32(res)>>31)&1 != 0
				cpsr.Z = uint32(res) == 0
			}
		}

	case inst == SUB:
		res := uint64(rnv) - uint64(op2)
		r[rd] = uint32(res)
		if set := (op>>20)&1 != 0; set {
			if rd == PC {
				if rn == LR {
					switch cpsr.Mode {
					case MODE_ABT:
						r[PC] += 4
						cpu.ExitException(MODE_ABT)
					case MODE_SWI:
						cpu.ExitException(MODE_SWI)
					default:
						cpu.ExitException(MODE_IRQ)
					}
				} else {
					// force exit
					cpu.psrSwitch() // not sure if needed
				}

				if r[PC]&1 != 0 {
					cpu.toggleThumb()
				}

			} else {
				cpsr.V = ((rnv^op2)&(rnv^uint32(res)))>>31 != 0
				cpsr.C = res < 0x1_0000_0000
				cpsr.N = (uint32(res)>>31)&1 != 0
				cpsr.Z = uint32(res) == 0
			}
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

		if rd == PC {
			r[PC] += 4
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

	//cpu.Jit.EndTest(op, compare)

	switch {
	case rd != PC:
		r[PC] += 4
	case cpsr.T:
		r[PC] &^= 0b1
	case !cpsr.T:
		r[PC] &^= 0b11
	}
}

func (cpu *Cpu) getShiftedAluReg(op uint32) uint32 {

	var (
		r = &cpu.Reg.R

		carry = cpu.Reg.CPSR.C

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

		if rm == PC {
			op2 += 12
		}

	} else {

		shift = (op >> 7) & 0x1F

		if rm == PC {
			op2 += 8
		}

		if special := shift == 0; special {
			switch shType {
			case LSL:
				return op2
			case LSR:
				cpu.Reg.CPSR.C = op2&0x8000_0000 != 0
				return 0
			case ASR:

				signed := op2&0x8000_0000 != 0

				if setCarry {
					cpu.Reg.CPSR.C = signed
				}

				if signed {
					return 0xFFFF_FFFF
				}

				return 0

			case ROR:

				cpu.Reg.CPSR.C = op2&1 != 0

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
		cpu.Reg.CPSR.C = carry
	}

	return op2
}

func (cpu *Cpu) psrSwitch() {

	var (
		reg  = &cpu.Reg
		r    = &cpu.Reg.R
		cpsr = &cpu.Reg.CPSR
		curr = cpsr.Mode
		i    = BANK_ID[curr]
	)

	*cpsr = reg.SPSR[i]
	next := cpsr.Mode
	c := BANK_ID[next]

	reg.LR[i] = r[LR]
	reg.SP[i] = r[SP]
	r[SP] = reg.SP[c]
	r[LR] = reg.LR[c]

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

const (
	MUL   = 0b0
	MLA   = 0b1
	UMAAL = 0b010
	UMULL = 0b100
	UMLAL = 0b101
	SMULL = 0b110
	SMLAL = 0b111

	SMLAxy        = 0b1000
	SMLAWySMLALWy = 0b1001
	SMLALxy       = 0b1010
	SMULxy        = 0b1011
)

func (cpu *Cpu) Mul(op uint32) {

	var (
		inst = (op >> 21) & 0xF
		set  = (op>>20)&1 != 0
		rd   = (op >> 16) & 0xF
		rn   = (op >> 12) & 0xF
		rs   = (op >> 8) & 0xF
		rm   = (op >> 0) & 0xF
		r    = &cpu.Reg.R
		cpsr = &cpu.Reg.CPSR
	)

	switch inst {
	case MUL, MLA:

		res := r[rm] * r[rs]

		if inst == MLA {
			res += r[rn]
		}

		r[rd] = res

		if set {
			cpsr.N = (uint32(res)>>31)&1 != 0
			cpsr.Z = uint32(res) == 0
			// FLAG_C "destroyed" ARM <5, ignored ARM >=5
			//cpsr.C = false
		}

		r[PC] += 4
		return

	case UMAAL:
		panic("unsupported umaal instruction")

	case UMULL, UMLAL:

		res := uint64(r[rm]) * uint64(r[rs])

		if inst == UMLAL {
			res += uint64(r[rd])<<32 | uint64(r[rn])
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

		r[PC] += 4
		return

	case SMULL, SMLAL:

		res := int64(int32(r[rm])) * int64(int32(r[rs]))
		if inst == SMLAL {
			res += int64(r[rd])<<32 | int64(r[rn])
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

		r[PC] += 4
		return
	}

	x := (op >> 5) & 1
	y := (op >> 6) & 1

	switch inst {
	case SMLAxy:

		rmV := int64(int16((r[rm] >> (16 * x)) & 0xFFFF))
		rsV := int64(int16((r[rs] >> (16 * y)) & 0xFFFF))
		rnV := int64(int32(r[rn]))

		res := rmV * rsV

		res += rnV

		r[rd] = uint32(res)

		if res > math.MaxInt32 || res < math.MinInt32 {
			cpsr.Q = true
		}

	case SMLAWySMLALWy:

		rsV := int64(int16((r[rs] >> (16 * y) & 0xFFFF)))
		rmV := int64(int32(r[rm]))
		res := (rmV * rsV) >> 16

		if smulwa := x == 0; smulwa {
			add := int64(int32(r[rn]))

			if res+add > math.MaxInt32 || res+add < math.MinInt32 {
				cpsr.Q = true
			}

			res += add
		}

		r[rd] = uint32(res)

	case SMLALxy:

		rsV := int64(int16((r[rs] >> (16 * y) & 0xFFFF)))
		rmV := int64(int16((r[rm] >> (16 * x) & 0xFFFF)))

		res := rsV * rmV
		add := int64(int32(r[rd]))<<32 | int64(int32(r[rn]))
		res += add

		r[rd] = uint32(res >> 32)
		r[rn] = uint32(res)

	case SMULxy:

		rmV := int64(int16((r[rm] >> (16 * x)) & 0xFFFF))
		rsV := int64(int16((r[rs] >> (16 * y)) & 0xFFFF))

		res := rmV * rsV

		r[rd] = uint32(res)
	}

	r[PC] += 4

}

const (
	STR = iota
	LDR_PLD
)

func (c *Cpu) Sdt(op uint32) {

	if valid := (op>>26)&0b11 == 0b01; !valid {
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
		sType := (op >> 5) & 0b11
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
	if rn == PC {
		post += 8
	}

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
			r[rd] = c.mem.Read8(prev, true)
		} else {

			v := c.mem.Read32(prev&^0b11, true)
			is := ((prev & 0b11) << 3) & 0x1F
			r[rd] = bits.RotateLeft32(v, -int(is))

			if rd == PC {
				c.toggleThumb() // this is arm9 - not sure if arm7
				r[rd] -= 4
			}
		}
	} else {
		v := r[rd]
		if rd == PC {
			v += 12
		}

		if byte {
			c.mem.Write8(prev, uint8(v), true)
		} else {
			c.mem.Write32(prev&^0b11, v, true)
		}
	}

	if wb && !(load && rn == rd) {
		r[rn] = post
	}

	r[PC] += 4
}

func (cpu *Cpu) BLX(op uint32) {

	r := &cpu.Reg.R

	r[14] = r[15] + 4

	r[PC] += uint32((int32(op)<<8)>>6) + 8

	if halfOffset := (op>>24)&1 != 0; halfOffset {
		r[PC] += 2
	}

	cpu.Reg.CPSR.T = true
}

func (cpu *Cpu) B(op uint32) {

	if immLoop := op == 0xEAFFFFFE; immLoop {
		cpu.Halted = true
		return
	}

	r := &cpu.Reg.R

	if link := (op>>24)&1 != 0; link {
		r[14] = r[15] + 4
	}

	r[PC] += uint32((int32(op)<<8)>>6) + 8
}

const (
	INST_BX  = 1
	INST_BXJ = 2
	INST_BLX = 3
)

func (cpu *Cpu) BX(op uint32) {

	var (
		r    = &cpu.Reg.R
		inst = (op >> 4) & 0xF
		rn   = op & 0xF
	)

	switch inst {
	case INST_BX:
		r[PC] = r[rn]

		if rn == PC {
			r[PC] += 8
		}

		cpu.toggleThumb()
	case INST_BXJ:
		panic("Unsupported BXJ Instruction")
	case INST_BLX:

		if rn == 14 {
			// Using BLX R14 is possible (sets PC=Old_LR, and New_LR=retadr).
			tmp := r[14]
			r[14] = r[PC] + 4
			r[PC] = tmp
			cpu.toggleThumb()
			return
		}

		r[14] = r[PC] + 4
		r[PC] = r[rn]
		cpu.toggleThumb()

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

	if rn == PC {
		rnv += 8
	}

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
		if rd == PC {
			rdv += 12
		}

		rd2v := r[rd+1]
		if rd+1 == PC {
			rd2v += 12
		}

		if wb {
			r[rn] = post
		}

		switch inst {
		case STRH:
			c.mem.Write16(pre&^1, uint16(rdv), true)

		case LDRD:
			addr := pre &^ 0b111
			r[rd] = c.mem.Read32(addr, true)
			r[rd+1] = c.mem.Read32(addr+4, true)

		case STRD:

			addr := pre &^ 0b111
			c.mem.Write32(addr, rdv, true)
			c.mem.Write32(addr+4, rd2v, true)

		}

		r[PC] += 4
		return
	}

	if wb {
		r[rn] = post
	}

	switch inst {
	case LDRH:
		v := uint32(c.mem.Read16(pre&^1, true))
		is := (pre & 1) << 3
		r[rd] = bits.RotateLeft32(v, -int(is))

	case LDRSB:
		// sign-expand byte value
		r[rd] = uint32(int32(int8(c.mem.Read8(pre, true))))

	case LDRSH:
		// sign-expand half value
		r[rd] = uint32(int32(int16(c.mem.Read16(pre&^1, true))))

	}

	r[PC] += 4
}

func (cpu *Cpu) Psr(op uint32) {

	r := &cpu.Reg.R

	if msr := (op>>21)&1 != 0; msr {
		cpu.msr(op)
		r[PC] += 4
		return
	}

	rd := (op >> 12) & 0xF

	if spsr := (op>>22)&1 != 0; spsr {
		mode := cpu.Reg.CPSR.Mode
		r[rd] = cpu.Reg.SPSR[BANK_ID[mode]].Get()
		r[PC] += 4
		return
	}

	mask := PRIV_MASK
	if cpu.Reg.CPSR.Mode == MODE_USR {
		mask = USR_MASK
	}

	r[rd] = uint32(cpu.Reg.CPSR.Get()) & mask
	r[PC] += 4
}

const (
	PRIV_MASK  uint32 = 0xF8FF_03DF
	USR_MASK   uint32 = 0xF8FF_0000
	STATE_MASK uint32 = 0x0100_0020
)

func (cpu *Cpu) msr(op uint32) {

	r := &cpu.Reg.R

	spsrFlag := (op>>22)&1 != 0

	var v uint32
	if imm := (op>>25)&1 != 0; imm {
		shift := ((op >> 8) & 0xF) * 2
		v = bits.RotateLeft32(op&0xFF, -int(shift&31))

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
	curr := cpu.Reg.CPSR.Mode
	if curr == MODE_USR {
		secMask = USR_MASK
	}

	if spsrFlag {
		secMask |= STATE_MASK
	}

	mask &= secMask

	reg := &cpu.Reg

	if spsrFlag {

		var spsr uint32

		if curr == MODE_USR || curr == MODE_SYS {
			spsr = uint32(reg.CPSR.Get()) &^ mask
		} else {
			spsr = uint32(reg.SPSR[BANK_ID[curr]].Get()) &^ mask
		}

		spsr |= v & mask
		reg.SPSR[BANK_ID[curr]].Set(spsr)

		return
	}

	next := v & 0x1F
	cpsr := uint32(reg.CPSR.Get()) &^ mask

	cpsr |= v & mask

	reg.CPSR.Set(cpsr)

	if skip := BANK_ID[curr] == BANK_ID[next]; skip {
		return
	}

	if curr == MODE_USR {
		panic("USER MODE MSR")
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

func (cpu *Cpu) Swp(op uint32) {

	var (
		r      = &cpu.Reg.R
		isByte = (op>>22)&1 != 0
		rn     = (op >> 16) & 0xF
		rd     = (op >> 12) & 0xF
		rm     = op & 0xF
		rmv    = r[rm]
		rnv    = r[rn]
	)

	if isByte {
		r[rd] = cpu.mem.Read8(rnv, true)
		cpu.mem.Write8(rnv, uint8(rmv), true)
	} else {
		v := cpu.mem.Read32(rnv&^0b11, true)
		is := (rnv & 0b11) << 3
		r[rd] = bits.RotateLeft32(v, -int(is))
		cpu.mem.Write32(rnv&^0b11, rmv, true)
	}

	r[PC] += 4
}

const (
	QADD  = 0
	QSUB  = 2
	QDADD = 4
	QDSUB = 6
)

func (cpu *Cpu) Qalu(op uint32) {

	var (
		r    = &cpu.Reg.R
		inst = (op >> 20) & 0xF
		rnV  = int64(int32(r[(op>>16)&0xF]))
		rmV  = int64(int32(r[op&0xF]))
	)

	if double := inst >= 4; double {

		rnV *= 2

		if rnV > math.MaxInt32 {
			cpu.Reg.CPSR.Q = true
			rnV = math.MaxInt32
		}

		if rnV < math.MinInt32 {
			cpu.Reg.CPSR.Q = true
			rnV = math.MinInt32
		}
	}

	if inst == QADD || inst == QDADD {
		rnV += rmV
	} else {
		rnV = rmV - rnV
	}

	if rnV > math.MaxInt32 {
		cpu.Reg.CPSR.Q = true
		rnV = math.MaxInt32
	}

	if rnV < math.MinInt32 {
		cpu.Reg.CPSR.Q = true
		rnV = math.MinInt32
	}

	rd := (op >> 12) & 0xF
	r[rd] = uint32(rnV)

	r[PC] += 4
}

func (cpu *Cpu) Clz(op uint32) {

	r := &cpu.Reg.R
	rm := op & 0xF
	rd := (op >> 12) & 0xF
	r[rd] = uint32(bits.LeadingZeros32(r[rm]))
	r[PC] += 4
}

func (cpu *Cpu) CoDataReg(op uint32) {

	var (
		reg = cp15.CpRegister{
			Op: uint8((op >> 21) & 0x7),
			Cn: uint8((op >> 16) & 0xF),
			Pn: uint8((op >> 8) & 0xF),
			Cp: uint8((op >> 5) & 0x7),
			Cm: uint8((op >> 0) & 0xF),
		}

		r  = &cpu.Reg.R
		rd = (op >> 12) & 0xF
	)

	if (op >> 28) == 0xF {
		panic("MRC2/MCR2")
	}

	if rd == 15 {
		panic("SETUP PIPELINE OFFSET CO DATA REG")
	}

	if mrc := (op>>20)&1 == 1; mrc {
		r[rd] = cpu.Cp15.Read(reg)
		r[PC] += 4
		return
	}

	if rd == 0 && (reg == cp15.HALT || reg == cp15.HALT2) {

		if !cpu.Irq.IME {
			panic("ARM9 CPU HALTED WITHOUT IME ENABLED")
		}

		cpu.Halted = true
		r[15] += 4
		return
	}

	cpu.Cp15.Write(r[rd], reg, &cpu.LowVector)
	r[15] += 4
}

func (c *Cpu) Block(op uint32) {

	var (
		r     = &c.Reg.R
		rlist = op & 0xFFFF
		up    = (op>>23)&1 != 0
		rn    = (op >> 16) & 0xF
	)

	if rlist == 0 {

		r[PC] += 4

		if up {
			r[rn] += 0x40
			return
		}

		r[rn] -= 0x40
		return
	}

	var (
		pcIncluded = rlist&0x8000 != 0
		rnIncluded = (rlist>>rn)&1 != 0
		pre        = (op>>24)&1 != 0
		psr        = (op>>22)&1 != 0
		wb         = (op>>21)&1 != 0
		load       = (op>>20)&1 != 0
		forceUser  = psr && (c.Reg.CPSR.Mode != MODE_USR) && (!load || !pcIncluded)
		addr       = r[rn] &^ 0b11
		regCount   = uint32(bits.OnesCount32(rlist))
		wbValue    = r[rn]
		// fiq switch has additional r8 - r12 use mode switch registers
		forceFIQSwitch = forceUser && c.Reg.CPSR.Mode == MODE_FIQ
	)

	if up {
		wbValue += regCount * 4
	} else {
		wbValue -= regCount * 4
	}

	rnRef := &c.Reg.R[rn]
	switch {
	case forceFIQSwitch:

		switch {
		case rn == 13:
			rnRef = &c.Reg.SP[BANK_ID[MODE_USR]]
		case rn == 14:
			rnRef = &c.Reg.LR[BANK_ID[MODE_USR]]
		case rn >= 8:
			rnRef = &c.Reg.USR[rn-8]

		}

	case forceUser:
		switch {
		case rn == 13:
			rnRef = &c.Reg.SP[BANK_ID[MODE_USR]]
		case rn == 14:
			rnRef = &c.Reg.LR[BANK_ID[MODE_USR]]
		}
	}

	var (
		rnv = *rnRef
		reg = uint32(0)

		p unsafe.Pointer
	)

	if !up {
		reg = 15
	}

	if load {
		p, _ = c.mem.ReadPtr(addr, true)
	} else {
		p, _ = c.mem.WritePtr(addr, true)
	}

	for range 16 {

		if disabled := (rlist>>reg)&1 == 0; disabled {
			if up {
				reg++
			} else {
				reg--
			}
			continue
		}

		ref := &c.Reg.R[reg]
		switch {
		case forceFIQSwitch:

			switch {
			case reg == 13:
				ref = &c.Reg.SP[BANK_ID[MODE_USR]]
			case reg == 14:
				ref = &c.Reg.LR[BANK_ID[MODE_USR]]
			case reg >= 8:
				ref = &c.Reg.USR[reg-8]
			}

		case forceUser:
			switch {
			case reg == 13:
				ref = &c.Reg.SP[BANK_ID[MODE_USR]]
			case reg == 14:
				ref = &c.Reg.LR[BANK_ID[MODE_USR]]
			}
		}

		if pre {
			if up {
				if p != nil {
					p = unsafe.Add(p, 4)
				} else {
					addr += 4
				}
			} else {
				if p != nil {
					p = unsafe.Add(p, -4)
				} else {
					addr -= 4
				}
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
				case rn:
					c.mem.Write32(addr, rnv, true)
				case PC:
					c.mem.Write32(addr, *ref+12, true)
				default:
					c.mem.Write32(addr, *ref, true)
				}
			} else {
				switch reg {
				case rn:
					*(*uint32)(p) = rnv
				case PC:
					*(*uint32)(p) = *ref + 12
				default:
					*(*uint32)(p) = *ref
				}
			}
		}

		if !pre {
			if up {
				if p != nil {
					p = unsafe.Add(p, 4)
				} else {
					addr += 4
				}
			} else {
				if p != nil {
					p = unsafe.Add(p, -4)
				} else {
					addr -= 4
				}
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
		if rnIncluded {
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

	if !psr {
		c.toggleThumb()
		return
	}

	var (
		curr = c.Reg.CPSR.Mode
		spsr = c.Reg.SPSR[BANK_ID[curr]]
		next = spsr.Mode
	)

	c.Reg.CPSR = spsr

	if curr == MODE_USR {
		panic("USER MODE LDM PC CHANGE")
	}

	if curr != MODE_FIQ {
		for i := range 5 {
			c.Reg.USR[i] = r[8+i]
		}
	}

	c.Reg.SP[BANK_ID[curr]] = r[SP]
	c.Reg.LR[BANK_ID[curr]] = r[LR]

	if curr == MODE_FIQ {
		for i := range 5 {
			c.Reg.FIQ[i] = r[8+i]
		}
	}

	if next != MODE_FIQ {
		for i := range 5 {
			r[8+i] = c.Reg.USR[i]
		}
	}

	r[SP] = c.Reg.SP[BANK_ID[next]]
	r[LR] = c.Reg.LR[BANK_ID[next]]

	if next == MODE_FIQ {
		for i := range 5 {
			r[8+i] = c.Reg.FIQ[i]
		}
	}
}
