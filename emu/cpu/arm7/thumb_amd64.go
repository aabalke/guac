package arm7

import (
	"math/bits"

	"github.com/aabalke/gojit"
)

func (j *Jit) emitThumbLSSP(op uint16) {
	rd := uint32(op>>8) & 7

	j.Mov(JIT, gojit.Rax)

	j.Movl(j.REG(SP), gojit.Ebx)
	j.Add(gojit.Imm((op&0xFF)<<2), gojit.Ebx)

	if ldr := (op>>11)&1 != 0; ldr {

		j.Movl(gojit.Ebx, gojit.R8d)

		j.CallFunc((*Jit).Read32)

		j.Movl(gojit.R8d, gojit.Ecx)
		j.And(gojit.Imm(3), gojit.Ecx)
		j.Shl(gojit.Imm(3), gojit.Ecx)
		j.RorCl(gojit.Eax)

		j.Movl(gojit.Eax, j.REG(rd))

	} else {
		j.Movl(j.REG(rd), gojit.Ecx)
		j.CallFunc((*Jit).Write32)
	}
}

func (j *Jit) emitThumbStack(op uint16) {
	nn := int(op&0x7F) << 2
	if sub := (op>>7)&1 != 0; sub {
		j.Sub(gojit.Imm(nn), j.REG(SP))
	} else {
		j.Add(gojit.Imm(nn), j.REG(SP))
	}
}

func (j *Jit) emitThumbRelative(op uint16) {
	if isSP := (op>>11)&1 != 0; isSP {
		j.Movl(j.REG(SP), gojit.Eax)
	} else {
		j.Movl(j.REG(PC), gojit.Eax)
		j.And(gojit.Imm(^3), gojit.Eax)
	}

	j.Add(gojit.Imm(uint32(op&0xFF)<<2), gojit.Eax)
	j.Movl(gojit.Eax, j.REG(uint32(op>>8)&7))
}

func (j *Jit) emitThumbLPC(op uint16) {
	var (
		rd = uint32(op>>8) & 7
		nn = uint32(op&0xFF) << 2
	)

	j.Mov(JIT, gojit.Rax)

	j.Movl(j.REG(PC), gojit.Ebx)
	j.And(gojit.Imm(^3), gojit.Ebx)
	j.Add(gojit.Imm(nn), gojit.Ebx)

	j.CallFunc((*Jit).Read32)

	j.Movl(gojit.Eax, j.REG(rd))
}

func (j *Jit) emitThumbLSImm(op uint16) {
	var (
		rd = uint32(op & 7)
		rb = uint32((op >> 3) & 7)
		nn = uint32(op>>6) & 0x1F
	)

	j.Mov(JIT, gojit.Rax)

	j.Movl(j.REG(rb), gojit.Ebx)

	if byte := (op>>12)&1 != 0; byte {
		j.Add(gojit.Imm(nn), gojit.Ebx)
	} else {
		j.Add(gojit.Imm(nn*4), gojit.Ebx)
	}

	switch inst := (op >> 11) & 3; inst {
	case THUMB_STR_IMM, THUMB_STRB_IMM:

		j.Movl(j.REG(rd), gojit.Ecx)

		if inst == THUMB_STRB_IMM {
			j.CallFunc((*Jit).Write8)
		} else {
			j.CallFunc((*Jit).Write32)
		}

	case THUMB_LDR_IMM:

		j.Movl(gojit.Ebx, gojit.R8d)
		j.CallFunc((*Jit).Read32)

		j.Movl(gojit.R8d, gojit.Ecx)
		j.And(gojit.Imm(3), gojit.Ecx)
		j.Shl(gojit.Imm(3), gojit.Ecx)
		j.RorCl(gojit.Eax)

		j.Movl(gojit.Eax, j.REG(rd))

	case THUMB_LDRB_IMM:

		j.CallFunc((*Jit).Read8)
		j.Movl(gojit.Eax, j.REG(rd))
	}
}

func (j *Jit) emitThumbSdt(op uint16) {
	var (
		inst = (op >> 10) & 3
		rd   = uint32(op & 7)
	)

	j.Mov(JIT, gojit.Rax)

	j.Movl(j.REG(uint32(op>>3)&7), gojit.Ebx)
	j.Add(j.REG(uint32(op>>6)&7), gojit.Ebx)

	if signed := (op>>9)&1 != 0; signed {

		switch inst {
		case THUMB_STRH:

			j.Movl(j.REG(rd), gojit.Ecx)
			j.CallFunc((*Jit).Write16)

		case THUMB_LDSB:

			j.CallFunc((*Jit).Read8)
			j.Movsx(gojit.Al, gojit.Eax)
			j.Movl(gojit.Eax, j.REG(rd))

		case THUMB_LDRH:

			j.Movl(gojit.Ebx, gojit.R8d)
			j.CallFunc((*Jit).Read16)

			j.Movl(gojit.R8d, gojit.Ecx)
			j.And(gojit.Imm(1), gojit.Ecx)
			j.Shl(gojit.Imm(3), gojit.Ecx)
			j.RorCl(gojit.Eax)

			j.Movl(gojit.Eax, j.REG(rd))

		case THUMB_LDSH:

			j.Bt(gojit.Imm(0), gojit.Ebx)
			half := j.JccForward(gojit.CC_NC)

			j.CallFunc((*Jit).Read8)
			j.Movsx(gojit.Al, gojit.Eax)
			byte := j.JmpForward()

			half()
			j.CallFunc((*Jit).Read16)
			j.Movsx(gojit.Ax, gojit.Eax)

			byte()

			j.Movl(gojit.Eax, j.REG(rd))

		}
		return
	}

	switch inst {
	case THUMB_STR_REG:
		j.Movl(j.REG(rd), gojit.Ecx)
		j.CallFunc((*Jit).Write32)
	case THUMB_LDR_REG:

		j.Movl(gojit.Ebx, gojit.R8d)
		j.CallFunc((*Jit).Read32)

		j.Movl(gojit.R8d, gojit.Ecx)
		j.And(gojit.Imm(3), gojit.Ecx)
		j.Shl(gojit.Imm(3), gojit.Ecx)
		j.RorCl(gojit.Eax)

		j.Movl(gojit.Eax, j.REG(rd))
	case THUMB_STRB_REG:

		j.Movl(j.REG(rd), gojit.Ecx)
		j.CallFunc((*Jit).Write8)
	case THUMB_LDRB_REG:

		j.CallFunc((*Jit).Read8)
		j.Movl(gojit.Eax, j.REG(rd))
	}
}

func (j *Jit) emitThumbLSHalf(op uint16) {
	var (
		offset = uint32((op >> 6) & 0x1F << 1)
		rd     = uint32(op) & 7
	)

	j.Mov(JIT, gojit.Rax)

	j.Movl(j.REG(uint32(op>>3)&7), gojit.Ebx)
	j.Add(gojit.Imm(offset), gojit.Ebx)

	if ldr := (op>>11)&1 != 0; ldr {

		j.Movl(gojit.Ebx, gojit.R8d)
		j.CallFunc((*Jit).Read16)

		j.Movl(gojit.R8d, gojit.Ecx)
		j.And(gojit.Imm(1), gojit.Ecx)
		j.Shl(gojit.Imm(3), gojit.Ecx)
		j.RorCl(gojit.Eax)

		j.Movl(gojit.Eax, j.REG(rd))

	} else {

		j.Movl(j.REG(rd), gojit.Ecx)
		j.CallFunc((*Jit).Write16)
	}
}

func (j *Jit) emitThumbImm(op uint16) {
	var (
		rd = uint32(op>>8) & 7
		nn = uint32(op & 0xFF)
	)

	switch inst := (op >> 11) & 3; inst {
	case THUMB_IMM_MOV:

		j.Movl(gojit.Imm(nn), gojit.Eax)
		j.Movl(gojit.Eax, j.REG(rd))
		j.Test(gojit.Eax, gojit.Eax)

	case THUMB_IMM_CMP, THUMB_IMM_SUB:

		j.Movl(j.REG(rd), gojit.Eax)
		j.Sub(gojit.Imm(nn), gojit.Eax)
		if inst == THUMB_IMM_SUB {
			j.Movl(gojit.Eax, j.REG(rd))
		}

		j.SETcc(gojit.CC_O, jV)
		j.SETcc(gojit.CC_NC, jC)

	case THUMB_IMM_ADD:

		j.Add(gojit.Imm(nn), j.REG(rd))
		j.SETcc(gojit.CC_O, jV)
		j.SETcc(gojit.CC_C, jC)
	}

	j.SETcc(gojit.CC_S, jN)
	j.SETcc(gojit.CC_Z, jZ)
}

func (j *Jit) emitThumbAddSub(op uint16) {
	var (
		inst = (op >> 9) & 3
		rd   = uint32(op & 7)
	)

	j.Movl(j.REG(uint32(op>>3)&7), gojit.Eax)

	if reg := inst < 2; reg {
		j.Movl(j.REG(uint32(op>>6)&7), gojit.Ebx)
	} else {
		j.Movl(gojit.Imm((op>>6)&7), gojit.Ebx)
	}

	switch inst {
	case THUMB_ADD, THUMB_ADDImm:
		j.Add(gojit.Ebx, gojit.Eax)
		j.SETcc(gojit.CC_C, jC)
	case THUMB_SUB, THUMB_SUBImm:
		j.Sub(gojit.Ebx, gojit.Eax)
		j.SETcc(gojit.CC_NC, jC)
	}

	j.Movl(gojit.Eax, j.REG(rd))

	j.SETcc(gojit.CC_O, jV)
	j.SETcc(gojit.CC_S, jN)
	j.SETcc(gojit.CC_Z, jZ)
}

func (j *Jit) emitThumbAlu(op uint16) {
	var (
		inst = (op >> 6) & 0xF
		rd   = uint32(op & 7)
	)

	if inst == THUMB_LSL || inst == THUMB_LSR ||
		inst == THUMB_ASR || inst == THUMB_ROR {
		j.Mov(JIT, gojit.Rax)
		j.Movl(gojit.Imm(1), gojit.Ebx)
		j.CallFunc((*Jit).Idle)
	}

	// rdv = eax, rsv = ebx

	j.Movl(j.REG(rd), gojit.Eax)
	j.Movl(j.REG(uint32(op>>3)&7), gojit.Ebx)

	switch inst {
	case THUMB_MUL:

		j.Movl(gojit.Eax, gojit.R8d)
		j.Movl(gojit.Ebx, gojit.R10d)

		j.Movl(TRUE, gojit.Ebx)
		j.CallFunc(idleMul)

		j.Movl(gojit.Eax, gojit.Ebx)
		j.Mov(JIT, gojit.Rax)
		j.CallFunc((*Jit).Idle)

		j.Movl(gojit.R8d, gojit.Eax)
		j.Movl(gojit.R10d, gojit.Ebx)

		j.Mul(gojit.Ebx)
		j.Movl(gojit.Eax, j.REG(rd))
		j.Test(gojit.Eax, gojit.Eax)
		// ARM < 4, carry flag destroyed, ARM >= 5, carry flag unchanged
		// cpsr.C = false

	case THUMB_TST:

		j.Test(gojit.Ebx, gojit.Eax)

	case THUMB_CMN:
		j.Add(gojit.Ebx, gojit.Eax)

		j.SETcc(gojit.CC_O, jV)
		j.SETcc(gojit.CC_C, jC)

	case THUMB_CMP:
		j.Cmp(gojit.Ebx, gojit.Eax)
		j.Movl(gojit.Eax, j.REG(rd))
		j.SETcc(gojit.CC_O, jV)
		j.SETcc(gojit.CC_NC, jC)

	case THUMB_AND:
		j.And(gojit.Ebx, gojit.Eax)
		j.Movl(gojit.Eax, j.REG(rd))

	case THUMB_EOR:
		j.Xor(gojit.Ebx, gojit.Eax)
		j.Movl(gojit.Eax, j.REG(rd))

	case THUMB_ORR:
		j.Or(gojit.Ebx, gojit.Eax)
		j.Movl(gojit.Eax, j.REG(rd))

	case THUMB_BIC:
		j.Not(gojit.Ebx)
		j.And(gojit.Ebx, gojit.Eax)
		j.Movl(gojit.Eax, j.REG(rd))

	case THUMB_MVN:
		j.Not(gojit.Ebx)
		j.Test(gojit.Ebx, gojit.Ebx)
		j.Movl(gojit.Ebx, j.REG(rd))

	case THUMB_NEG:

		j.Neg(gojit.Ebx)
		j.Movl(gojit.Ebx, j.REG(rd))

		j.SETcc(gojit.CC_O, jV)
		j.SETcc(gojit.CC_NC, jC)

	case THUMB_SBC:
		j.Movb(jC, gojit.Cl)

		j.Bt(gojit.Imm(0), gojit.Cl)
		j.Cmc() // compliment carry (reverse for sub)
		j.Sbb(gojit.Ebx, gojit.Eax)

		j.SETcc(gojit.CC_O, jV)
		j.SETcc(gojit.CC_NC, jC)

		j.Movl(gojit.Eax, j.REG(rd))

	case THUMB_ADC:
		j.Movb(jC, gojit.Cl)

		j.Bt(gojit.Imm(0), gojit.Cl)
		j.Adc(gojit.Ebx, gojit.Eax)

		j.SETcc(gojit.CC_O, jV)
		j.SETcc(gojit.CC_C, jC)

		j.Movl(gojit.Eax, j.REG(rd))

	case THUMB_LSL, THUMB_LSR, THUMB_ASR, THUMB_ROR:

		j.And(gojit.Imm(0xFF), gojit.Ebx)

		shType := ROR
		switch inst {
		case THUMB_LSL:
			shType = LSL
		case THUMB_LSR:
			shType = LSR
		case THUMB_ASR:
			shType = ASR
		}

		// carry to dl
		j.Movb(jC, gojit.Dl)

		// shift reg ebx, v eax
		j.ShiftReg(uint32(shType))

		j.Movb(gojit.Dl, jC)
		j.Movl(gojit.Eax, j.REG(rd))

		j.Test(gojit.Eax, gojit.Eax)
	}

	j.SETcc(gojit.CC_S, jN)
	j.SETcc(gojit.CC_Z, jZ)
}

func (j *Jit) emitThumbPushPop(op uint16) {
	var (
		pclr  = (op>>8)&1 != 0
		rlist = op & 0xFF
		pop   = (op>>11)&1 != 0
		seq   = uint32(NONSEQ)
	)

	// thank you nano
	if rlist == 0 && !pclr {
		if pop {
			panic("thumb push pop rlist == 0 and !pclr pop")
		}

		j.Sub(gojit.Imm(0x40), j.REG(SP))

		j.Mov(JIT, gojit.Rax)
		j.Movl(j.REG(SP), gojit.Ebx)
		j.Movl(j.REG(PC), gojit.Ecx)
		j.Movl(gojit.Imm(seq), gojit.Edx)

		j.CallFunc((*Jit).Write32Block)

		return
	}

	if pop {
		for reg := range uint32(8) {
			if disabled := (rlist>>reg)&1 == 0; disabled {
				continue
			}

			j.Mov(JIT, gojit.Rax)
			j.Movl(j.REG(SP), gojit.Ebx)
			j.Movl(gojit.Imm(seq), gojit.Ecx)

			j.CallFunc((*Jit).Read32Block)
			j.Movl(gojit.Eax, j.REG(reg))

			j.Add(gojit.Imm(4), j.REG(SP))

			seq = SEQ
		}

		if pclr {
			panic("pclr")
		}

		j.Mov(JIT, gojit.Rax)
		j.Movl(gojit.Imm(1), gojit.Ebx)
		j.CallFunc((*Jit).Idle)

	} else {

		if pclr {
			j.Sub(gojit.Imm(4), j.REG(SP))

			j.Mov(JIT, gojit.Rax)
			j.Movl(j.REG(SP), gojit.Ebx)
			j.Movl(j.REG(LR), gojit.Ecx)
			j.Movl(gojit.Imm(seq), gojit.Edi)

			j.CallFunc((*Jit).Write32Block)

			seq = SEQ
		}

		for reg := 7; reg >= 0; reg-- {
			if disabled := (rlist>>reg)&1 == 0; disabled {
				continue
			}

			j.Sub(gojit.Imm(4), j.REG(SP))

			j.Mov(JIT, gojit.Rax)
			j.Movl(j.REG(SP), gojit.Ebx)
			j.Movl(j.REG(uint32(reg)), gojit.Ecx)
			j.Movl(gojit.Imm(seq), gojit.Edi)

			j.CallFunc((*Jit).Write32Block)

			seq = SEQ
		}
	}
}

func (j *Jit) emitThumbHi(op uint16) {
	var (
		rd = uint32((op & 7) | (((op >> 7) & 1) << 3))
		rs = uint32(op>>3) & 0xF
	)

	if rd == PC {
		panic("hi jit thumb jit rd == PC")
	}

	switch inst := (op >> 8) & 3; inst {
	case HI_ADD:

		j.Movl(j.REG(rs), gojit.Eax)
		j.Add(j.REG(rd), gojit.Eax)
		j.Movl(gojit.Eax, j.REG(rd))

	case HI_CMP:

		j.Movl(j.REG(rd), gojit.Eax)
		j.Sub(j.REG(rs), gojit.Eax)

		j.SETcc(gojit.CC_O, jV)
		j.SETcc(gojit.CC_NC, jC)
		j.SETcc(gojit.CC_S, jN)
		j.SETcc(gojit.CC_Z, jZ)

	case HI_MOV:
		if nop := op == 0x46C0; nop {
			return
		}

		j.Movl(j.REG(rs), gojit.Eax)
		j.Movl(gojit.Eax, j.REG(rd))

	case HI_BX:
		panic("hi bx on jit")
	}
}

func (j *Jit) emitThumbShifted(op uint16) {
	shift := uint32(op>>6) & 0x1F

	j.Movl(j.REG(uint32(op>>3)&7), gojit.Eax)
	j.Movb(jC, gojit.Bl)

	j.ShiftImm(uint32(op>>11)&3, shift)
	j.Movb(gojit.Bl, jC)

	j.Movl(gojit.Eax, j.REG(uint32(op&7)))

	j.Test(gojit.Eax, gojit.Eax)
	j.SETcc(gojit.CC_S, jN)
	j.SETcc(gojit.CC_Z, jZ)
}

func (j *Jit) emitThumbBlock(op uint16) {
	var (
		ldmia = (op>>11)&1 != 0
		rb    = uint32(op>>8) & 7
		rlist = uint32(op & 0xFF)
	)

	if rlist == 0 {
		if ldmia {
			panic("thumb block rlist == 0 ldmia")
		}

		j.Mov(JIT, gojit.Rax)

		j.Movl(j.REG(rb), gojit.Ebx)
		j.Movl(j.REG(PC), gojit.Ecx)
		j.Add(gojit.Ecx, gojit.Imm(2))

		j.CallFunc((*Jit).Write32)

		j.Add(gojit.Imm(0x40), j.REG(rb))
		return
	}

	if !ldmia {
		var (
			count = uint32(bits.OnesCount16(uint16(rlist))) * 4
			first = uint32(0)
		)

		// can this be sped up? Its just log2(rlist & -rlist)
		// first = int(math.Log2(float64(rlist & -rlist)))
		for reg := 7; reg >= 0; reg-- {
			if rlist&(1<<reg) != 0 {
				first = uint32(reg)
			}
		}

		// addr stored in r8

		j.Movl(j.REG(rb), gojit.R8d)

		j.Mov(JIT, gojit.Rax)
		j.Movl(gojit.R8d, gojit.Ebx)
		j.Movl(j.REG(first), gojit.Ecx)
		j.Movl(gojit.Imm(NONSEQ), gojit.Edx)
		j.CallFunc((*Jit).Write32Block)

		j.Movl(gojit.R8d, gojit.Eax)
		j.Add(gojit.Imm(count), gojit.Eax)
		j.Movl(gojit.Eax, j.REG(rb))

		j.Add(gojit.Imm(4), gojit.R8d)

		for reg := first + 1; reg < 8; reg++ {
			if enabled := (rlist>>reg)&1 != 0; enabled {

				j.Mov(JIT, gojit.Rax)
				j.Movl(gojit.R8d, gojit.Ebx)
				j.Movl(j.REG(reg), gojit.Ecx)
				j.Movl(gojit.Imm(SEQ), gojit.Edx)
				j.CallFunc((*Jit).Write32Block)

				j.Add(gojit.Imm(4), gojit.R8d)
			}
		}
	} else {

		// addr stored in r8
		j.Movl(j.REG(rb), gojit.R8d)

		seq := uint32(NONSEQ)

		for reg := range uint32(8) {
			if enabled := (rlist>>reg)&1 != 0; enabled {

				j.Mov(JIT, gojit.Rax)
				j.Movl(gojit.R8d, gojit.Ebx)
				j.Movl(gojit.Imm(seq), gojit.Ecx)

				j.CallFunc((*Jit).Read32Block)

				j.Movl(gojit.Eax, j.REG(reg))

				j.Add(gojit.Imm(4), gojit.R8d)
				seq = SEQ
			}
		}

		if ^rlist&(1<<rb) != 0 {
			j.Movl(gojit.R8d, j.REG(rb))
		}

		j.Mov(JIT, gojit.Rax)
		j.Movl(gojit.Imm(1), gojit.Ebx)
		j.CallFunc((*Jit).Idle)
	}
}
