package arm7

import (
	"math/bits"

	"github.com/aabalke/gojit"
)

func (j *Jit) emitMul(op uint32) {
	var (
		set = (op>>20)&1 != 0
		rd  = (op >> 16) & 0xF
		rn  = (op >> 12) & 0xF
		rs  = (op >> 8) & 0xF
		rm  = op & 0xF
		add = (op>>21)&1 != 0
	)

	switch inst := (op >> 21) & 0xF; inst {
	case MUL, MLA:

		// timings
		j.Movl(j.REG(rs), gojit.Eax)
		j.Movl(TRUE, gojit.Ebx)
		j.CallFunc(idleMul)

		if add {
			j.Add(gojit.Imm(1), gojit.Eax)
		}

		j.Movl(gojit.Eax, gojit.Ebx)
		j.Mov(JIT, gojit.Rax)
		j.CallFunc((*Jit).Idle)

		// multiply
		j.Movl(j.REG(rs), gojit.Eax)
		j.Mul(j.REG(rm))

		if add {
			j.Add(j.REG(rn), gojit.Eax)
		}

		if set {
			j.Test(gojit.Eax, gojit.Eax)
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
			//j.Movb(gojit.Imm(0), C)
		}

		j.Movl(gojit.Eax, j.REG(rd))

	case UMAAL:
		panic("unsupported umaal instruction")

	case UMULL, UMLAL:

		// timings
		j.Movl(j.REG(rs), gojit.Eax)
		j.Movl(FALSE, gojit.Ebx)
		j.CallFunc(idleMul)
		j.Add(gojit.Imm(1), gojit.Eax)

		if add {
			j.Add(gojit.Imm(1), gojit.Eax)
		}

		j.Movl(gojit.Eax, gojit.Ebx)
		j.Mov(JIT, gojit.Rax)
		j.CallFunc((*Jit).Idle)

		// multiply
		j.Movl(j.REG(rs), gojit.Eax)
		j.Movl(j.REG(rm), gojit.Ebx)
		j.Mul(gojit.Rbx)

		if add {
			j.Movl(j.REG(rd), gojit.Ecx)
			j.Shl(gojit.Imm(32), gojit.Rcx)
			j.Movl(j.REG(rn), gojit.Edi)
			j.Add(gojit.Rcx, gojit.Rax)
			j.Add(gojit.Rdi, gojit.Rax)
		}

		if set {
			j.Test(gojit.Rax, gojit.Rax)
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
			j.Movb(gojit.Imm(0), jC)
		}

		j.Movl(gojit.Eax, j.REG(rn))
		j.Shr(gojit.Imm(32), gojit.Rax)
		j.Movl(gojit.Eax, j.REG(rd))
		return

	case SMULL, SMLAL:

		// timings
		j.Movl(j.REG(rs), gojit.Eax)
		j.Movl(TRUE, gojit.Ebx)
		j.CallFunc(idleMul)
		j.Add(gojit.Imm(1), gojit.Eax)

		if add {
			j.Add(gojit.Imm(1), gojit.Eax)
		}

		j.Movl(gojit.Eax, gojit.Ebx)
		j.Mov(JIT, gojit.Rax)
		j.CallFunc((*Jit).Idle)

		// multiply
		j.Movl(j.REG(rs), gojit.Eax)
		j.Movl(j.REG(rm), gojit.Ebx)

		// sign extend 32 -> 64
		j.Movsxd(gojit.Eax, gojit.Rax)
		j.Movsxd(gojit.Ebx, gojit.Rbx)

		j.Imul(gojit.Rbx)

		if add {
			j.Movl(j.REG(rd), gojit.Ecx)
			j.Shl(gojit.Imm(32), gojit.Rcx)
			j.Movl(j.REG(rn), gojit.Edi)
			j.Add(gojit.Rcx, gojit.Rax)
			j.Add(gojit.Rdi, gojit.Rax)
		}

		if set {
			j.Test(gojit.Rax, gojit.Rax)
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
			j.Movb(gojit.Imm(0), jC)
		}

		j.Movl(gojit.Eax, j.REG(rn))
		j.Shr(gojit.Imm(32), gojit.Rax)
		j.Movl(gojit.Eax, j.REG(rd))
	}
}

func (j *Jit) emitSwp(op uint32) {
	rn := (op >> 16) & 0xF
	rd := (op >> 12) & 0xF
	rm := op & 0xF

	j.Movl(j.REG(rn), gojit.R8d)
	j.Movl(j.REG(rm), gojit.R10d)

	j.Mov(JIT, gojit.Rax)
	j.Movl(gojit.R8d, gojit.Ebx)

	if isByte := (op>>22)&1 != 0; isByte {

		j.CallFunc((*Jit).Read8)

		j.Movl(gojit.Eax, j.REG(rd))

		j.Mov(JIT, gojit.Rax)
		j.Movl(gojit.R8d, gojit.Ebx)
		j.Movl(gojit.R10d, gojit.Ecx)

		j.CallFunc((*Jit).Write8)
		return
	}

	j.CallFunc((*Jit).Read32)

	j.Movl(gojit.R8d, gojit.Ecx)
	j.And(gojit.Imm(3), gojit.Ecx)
	j.Shl(gojit.Imm(3), gojit.Ecx)
	j.RorCl(gojit.Eax)

	j.Movl(gojit.Eax, j.REG(rd))

	j.Mov(JIT, gojit.Rax)
	j.Movl(gojit.R8d, gojit.Ebx)
	j.Movl(gojit.R10d, gojit.Ecx)
	j.CallFunc((*Jit).Write32)
}

func (j *Jit) emitHalf(op uint32) {
	var (
		rn   = (op >> 16) & 0xF
		rd   = (op >> 12) & 0xF
		pre  = (op>>24)&1 != 0
		load = (op>>20)&1 != 0
		inst = (op >> 5) & 3
		wb   = (op>>21)&1 != 0 || !pre
		imm  = (op>>22)&1 != 0
		up   = (op>>23)&1 != 0
	)

	// cpu rax, addr rbx, rdv rcx, post rdx

	j.Mov(JIT, gojit.Rax)
	j.Movl(j.REG(rn), gojit.Ebx)
	j.Movl(gojit.Ebx, gojit.Edx)

	switch {
	case imm && up:
		j.Add(gojit.Imm((op&0xF)|((op>>4)&0xF0)), gojit.Edx)
	case imm && !up:
		j.Sub(gojit.Imm((op&0xF)|((op>>4)&0xF0)), gojit.Edx)
	case !imm && up:
		j.Add(j.REG(op&0xF), gojit.Edx)
	case !imm && !up:
		j.Sub(j.REG(op&0xF), gojit.Edx)
	}

	if pre {
		j.Movl(gojit.Edx, gojit.Ebx)
	}

	j.Movl(j.REG(rd), gojit.Ecx)

	if wb {
		j.Movl(gojit.Edx, j.REG(rn))
	}

	if !load {

		if inst != STRH {
			panic("unsupported arm7 instruction (ldrd, strd, reserved)")
		}

		j.CallFunc((*Jit).Write16)
		return
	}

	switch inst {
	case LDRH:
		j.Movl(gojit.Ebx, gojit.R8d)
		j.CallFunc((*Jit).Read16)

		j.Movl(gojit.R8d, gojit.Ecx)
		j.And(gojit.Imm(1), gojit.Ecx)
		j.Shl(gojit.Imm(3), gojit.Ecx)
		j.RorCl(gojit.Eax)
		j.Movl(gojit.Eax, j.REG(rd))

	case LDRSB:
		// sign-expand byte value
		j.CallFunc((*Jit).Read8)
		j.Movsx(gojit.Al, gojit.Eax)
		j.Movl(gojit.Eax, j.REG(rd))

	case LDRSH:

		j.Bt(gojit.Imm(0), gojit.Ebx)
		half := j.JccForward(gojit.CC_NC)

		// sign-expand byte value
		j.CallFunc((*Jit).Read8)
		j.Movsx(gojit.Al, gojit.Eax)
		j.Movl(gojit.Eax, j.REG(rd))
		byte := j.JmpForward()

		// sign-expand half value
		half()
		j.CallFunc((*Jit).Read16)
		j.Movsx(gojit.Ax, gojit.Eax)
		j.Movl(gojit.Eax, j.REG(rd))

		byte()

	default:
		panic("unsupported arm7 instruction (ldrd, strd, reserved)")
	}
}

func (j *Jit) emitSdt(op uint32) {
	var (
		reg  = (op>>25)&1 != 0
		pre  = (op>>24)&1 != 0
		up   = (op>>23)&1 != 0
		byte = (op>>22)&1 != 0
		wb   = (op>>21)&1 != 0 || !pre
		load = (op>>20)&1 != 0
		rn   = (op >> 16) & 0xF
		rd   = (op >> 12) & 0xF
	)

	if load && rd == PC {
		j.EndBlock = true
		return
	}

	if reg {
		j.Movl(j.REG(op&0xF), gojit.Eax)
		j.Movb(jC, gojit.Bl)

		j.ShiftImm((op>>5)&3, (op>>7)&0x1F)
		j.Movl(gojit.Eax, gojit.Edx)
	} else {
		j.Mov(gojit.Imm(int32(op&0xFFF)), gojit.Edx)
	}

	// rn ebx, shift ecx
	// ax pre
	// bc post

	j.Movl(j.REG(rn), gojit.Ecx)

	if up {
		j.Add(gojit.Edx, gojit.Ecx)
	} else {
		j.Sub(gojit.Edx, gojit.Ecx)
	}

	if pre {
		j.Movl(gojit.Ecx, gojit.Ebx)
	} else {
		j.Movl(j.REG(rn), gojit.Ebx)
	}

	if wb {
		j.Movl(gojit.Ecx, j.REG(rn))
	}

	j.Mov(JIT, gojit.Rax)

	if load {
		if byte {
			j.CallFunc((*Jit).Read8)
			j.Movl(gojit.Eax, j.REG(rd))
		} else {

			j.Movl(gojit.Ebx, gojit.R8d)

			j.CallFunc((*Jit).Read32)

			j.Movl(gojit.R8d, gojit.Ecx)
			j.And(gojit.Imm(3), gojit.Ecx)
			j.Shl(gojit.Imm(3), gojit.Ecx)
			j.RorCl(gojit.Eax)
			j.Movl(gojit.Eax, j.REG(rd))

			if rd == PC {
				panic("toggle thumb")
			}
		}
	} else {

		j.Movl(j.REG(rd), gojit.Ecx)
		if rd == PC {
			j.Add(gojit.Imm(4), gojit.Ecx)
		}

		if byte {
			j.CallFunc((*Jit).Write8)
		} else {
			j.CallFunc((*Jit).Write32)
		}
	}
}

func (j *Jit) emitMrs(op uint32) {
	rd := (op >> 12) & 0xF

	if spsr := (op>>22)&1 != 0; spsr {
		j.Mov(JIT, gojit.Rax)
		j.Movl(MODE, gojit.Ebx)
		j.CallFunc((*Jit).GetSPSR)
		j.Movl(gojit.Eax, j.REG(rd))
		return
	}

	//mask := PRIV_MASK
	//if cpu.Reg.CPSR.Mode == MODE_USR {
	//	mask = USR_MASK
	//}

	//r[rd] = uint32(cpu.Reg.CPSR.Get()) & mask

	j.Mov(gojit.Imm(CPSR), gojit.Rax)
	j.Add(CPU, gojit.Rax)
	j.CallFunc((*Cond).Get)

	j.Cmp(gojit.Imm(MODE_USR), MODE)
	user := j.JccForward(gojit.CC_Z)

	j.MovAbs(uint64(PRIV_MASK), gojit.Rbx)
	priv := j.JmpForward()

	user()
	j.MovAbs(uint64(USR_MASK), gojit.Rbx)
	priv()

	j.And(gojit.Rbx, gojit.Rax)

	j.Movl(gojit.Eax, j.REG(rd))
}

func (j *Jit) emitAluOp2Reg(op uint32) {
	// op2 ax, carry dl

	if imm := (op>>4)&1 == 0; imm {
		shift := (op >> 7) & 0x1F

		// imm shift function uses bx for carry
		j.Movl(gojit.Edx, gojit.Ebx)
		j.Movl(j.REG(op&0xF), gojit.Eax)

		j.ShiftImm((op>>5)&3, shift)

		j.Movl(gojit.Ebx, gojit.Edx)
		j.Movl(gojit.Eax, gojit.Ebx)

		return
	}

	// TODO: have to move carry to not be clobbered, might be better method
	j.Movl(gojit.Rdx, gojit.R8)

	j.Mov(JIT, gojit.Rax)
	j.Movl(gojit.Imm(1), gojit.Ebx)
	j.CallFunc((*Jit).Idle)

	j.Movl(gojit.R8, gojit.Rdx)

	j.Movl(j.REG(op&0xF), gojit.Eax)
	if op&0xF == PC {
		j.Add(gojit.Imm(4), gojit.Eax)
	}

	// shift bx
	rs := (op >> 8) & 0xF
	j.Movl(j.REG(rs), gojit.Ebx)

	shType := (op >> 5) & 3
	j.ShiftReg(shType)

	j.Movl(gojit.Eax, gojit.Ebx)
}

func (j *Jit) ShiftImm(sType, shift uint32) {
	// v rax, carry rbx

	switch sType {
	case LSL:
		if shift != 0 {
			j.Bt(gojit.Imm(32-shift), gojit.Eax)
			j.SETcc(gojit.CC_C, gojit.Ebx)

			j.Shl(gojit.Imm(shift), gojit.Eax)
		}

	case LSR, ASR:

		if shift == 0 {
			shift = 32
		}

		j.Bt(gojit.Imm(shift-1), gojit.Eax)
		j.SETcc(gojit.CC_C, gojit.Ebx)

		if sType == ASR {
			// 31 for extending
			j.Sar(gojit.Imm(min(31, shift)), gojit.Eax)
		} else {
			// rax for extending
			j.Shr(gojit.Imm(shift), gojit.Rax)
		}

	case ROR:

		if rrx := shift == 0; rrx {
			// emulated carry flag -> CL -> CF
			j.Movl(gojit.Ebx, gojit.Ecx)
			j.Movl(gojit.Imm(1), gojit.Ebx)
			j.ShrCl(gojit.Ebx)
			j.Rcr(gojit.Imm(1), gojit.Eax)
			j.SETcc(gojit.CC_C, gojit.Ebx)

		} else {
			j.Bt(gojit.Imm((shift-1)&31), gojit.Eax)
			j.SETcc(gojit.CC_C, gojit.Ebx)

			j.Ror(gojit.Imm(shift), gojit.Eax)
		}
	}
}

func (j *Jit) ShiftReg(shType uint32) {
	// v == rax, shift == rbx

	j.And(gojit.Imm(0xFF), gojit.Ebx)

	j.Test(gojit.Ebx, gojit.Ebx)

	zero := j.JccForward(gojit.CC_Z)

	j.Movl(gojit.Ebx, gojit.Ecx)

	switch shType {
	case LSL, LSR:

		j.Cmp(gojit.Imm(32), gojit.Cl)

		isg32 := j.JccForward(gojit.CC_A)
		is32 := j.JccForward(gojit.CC_Z)

		if shType == LSR {
			j.ShrCl(gojit.Eax)
		} else {
			j.ShlCl(gojit.Eax)
		}

		j.SETcc(gojit.CC_C, gojit.Dl)

		done := j.JmpForward()

		is32()

		if shType == LSR {
			j.Test(gojit.Eax, gojit.Eax)
			j.SETcc(gojit.CC_S, gojit.Dl)
		} else {
			j.Test(gojit.Imm(1), gojit.Eax)
			j.SETcc(gojit.CC_NZ, gojit.Dl)
		}

		j.Xor(gojit.Eax, gojit.Eax)

		done2 := j.JmpForward()

		isg32()

		j.Xor(gojit.Eax, gojit.Eax)
		j.Xor(gojit.Edx, gojit.Edx)

		done()
		done2()

	case ASR:

		j.Cmp(gojit.Imm(32), gojit.Cl)

		isge32 := j.JccForward(gojit.CC_AE)

		j.SarCl(gojit.Eax)
		j.SETcc(gojit.CC_C, gojit.Dl)

		done := j.JmpForward()

		isge32()

		j.Test(gojit.Eax, gojit.Eax)
		j.SETcc(gojit.CC_S, gojit.Dl)
		j.Sar(gojit.Imm(31), gojit.Eax)

		done()

	case ROR:

		j.RorCl(gojit.Eax)
		j.SETcc(gojit.CC_C, gojit.Dl)

		j.Test(gojit.Imm(0x1F), gojit.Cl)

		done := j.JccForward(gojit.CC_NZ)

		j.Test(gojit.Eax, gojit.Eax)
		j.SETcc(gojit.CC_S, gojit.Dl)

		done()
	}

	zero()
}

func (j *Jit) emitAlu(op uint32) {
	var (
		inst = (op >> 21) & 0xF
		rd   = (op >> 12) & 0xF
		rn   = (op >> 16) & 0xF
	)

	if rd == PC {
		panic("rd == pc")
	}

	// eax rnv
	// ebx op2
	// carry to r8d
	j.Movb(jC, gojit.Dl)

	if imm := (op>>25)&1 != 0; imm {

		if shift := ((op >> 8) & 0xF) << 1; shift != 0 {
			j.Movl(gojit.Imm(((op&0xFF)>>(shift-1))&1), gojit.Edx)
			j.Movl(gojit.Imm(bits.RotateLeft32(op&0xFF, -int(shift))), gojit.Ebx)
		} else {
			j.Movl(gojit.Imm(op&0xFF), gojit.Ebx)
		}

		j.Movl(j.REG(rn), gojit.Eax)

	} else {

		// get op2, op2 will be in bx
		// shift
		j.emitAluOp2Reg(op)

		j.Movl(j.REG(rn), gojit.Eax)
		if regShift := (op>>4)&1 != 0; regShift && rn == PC {
			j.Add(gojit.Imm(4), gojit.Eax)
		}
	}

	aluInstJit[inst](j, op, rd)

	//j.Movl(j.REG(PC), gojit.Eax)
}

var aluInstJit = [...]func(j *Jit, op, rd uint32){
	// AND
	func(j *Jit, op, rd uint32) {
		j.And(gojit.Ebx, gojit.Eax)

		if set := (op>>20)&1 != 0; set {
			j.Movb(gojit.Dl, jC)
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
		}

		j.Movl(gojit.Eax, j.REG(rd))
	},

	// EOR
	func(j *Jit, op, rd uint32) {
		j.Xor(gojit.Ebx, gojit.Eax)

		if set := (op>>20)&1 != 0; set {
			j.Movb(gojit.Dl, jC)
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
		}

		j.Movl(gojit.Eax, j.REG(rd))
	},

	// SUB
	func(j *Jit, op, rd uint32) {
		j.Sub(gojit.Ebx, gojit.Eax)

		if set := (op>>20)&1 != 0; set {
			j.SETcc(gojit.CC_O, jV)
			j.SETcc(gojit.CC_NC, jC)
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
		}

		j.Movl(gojit.Eax, j.REG(rd))
	},

	// RSB
	func(j *Jit, op, rd uint32) {
		j.Sub(gojit.Eax, gojit.Ebx)

		if set := (op>>20)&1 != 0; set {
			j.SETcc(gojit.CC_O, jV)
			j.SETcc(gojit.CC_NC, jC)
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
		}

		j.Movl(gojit.Ebx, j.REG(rd))
	},

	// ADD
	func(j *Jit, op, rd uint32) {
		j.Add(gojit.Ebx, gojit.Eax)

		if set := (op>>20)&1 != 0; set {
			j.SETcc(gojit.CC_O, jV)
			j.SETcc(gojit.CC_C, jC)
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
		}

		j.Movl(gojit.Eax, j.REG(rd))
	},

	// ADC
	func(j *Jit, op, rd uint32) {
		j.Movb(jC, gojit.Cl)
		j.Bt(gojit.Imm(0), gojit.Cl)
		j.Adc(gojit.Ebx, gojit.Eax)

		if set := (op>>20)&1 != 0; set {
			j.SETcc(gojit.CC_O, jV)
			j.SETcc(gojit.CC_C, jC)
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
		}

		j.Movl(gojit.Eax, j.REG(rd))
	},

	// SBC
	func(j *Jit, op, rd uint32) {
		j.Movb(jC, gojit.Cl)
		j.Bt(gojit.Imm(0), gojit.Cl)
		j.Cmc() // compliment carry (reverse for sub)
		j.Sbb(gojit.Ebx, gojit.Eax)

		if set := (op>>20)&1 != 0; set {
			j.SETcc(gojit.CC_O, jV)
			j.SETcc(gojit.CC_NC, jC)
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
		}

		j.Movl(gojit.Eax, j.REG(rd))
	},

	// RSC
	func(j *Jit, op, rd uint32) {
		j.Movb(jC, gojit.Cl)
		j.Bt(gojit.Imm(0), gojit.Cl)
		j.Cmc() // compliment carry (reverse for sub)
		j.Sbb(gojit.Eax, gojit.Ebx)

		if set := (op>>20)&1 != 0; set {
			j.SETcc(gojit.CC_O, jV)
			j.SETcc(gojit.CC_NC, jC)
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
		}

		j.Movl(gojit.Ebx, j.REG(rd))
	},

	// TST
	func(j *Jit, op, rd uint32) {
		j.And(gojit.Ebx, gojit.Eax)
		if set := (op>>20)&1 != 0; set {
			j.Movb(gojit.Dl, jC)
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
		}

		if rd == PC {
			j.Add(gojit.Imm(4), j.REG(15))
		}
	},

	// TEQ
	func(j *Jit, op, rd uint32) {
		j.Xor(gojit.Ebx, gojit.Eax)
		if set := (op>>20)&1 != 0; set {
			j.Movb(gojit.Dl, jC)
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
		}

		if rd == PC {
			j.Add(gojit.Imm(4), j.REG(15))
		}
	},

	// CMP
	func(j *Jit, op, rd uint32) {
		j.Sub(gojit.Ebx, gojit.Eax)

		if set := (op>>20)&1 != 0; set {
			j.SETcc(gojit.CC_O, jV)
			j.SETcc(gojit.CC_NC, jC)
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
		}

		if rd == PC {
			j.Add(gojit.Imm(4), j.REG(15))
		}
	},

	// CMN
	func(j *Jit, op, rd uint32) {
		j.Add(gojit.Ebx, gojit.Eax)

		if set := (op>>20)&1 != 0; set {
			j.SETcc(gojit.CC_O, jV)
			j.SETcc(gojit.CC_C, jC)
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
		}

		if rd == PC {
			j.Add(gojit.Imm(4), j.REG(15))
		}
	},

	// ORR
	func(j *Jit, op, rd uint32) {
		j.Or(gojit.Ebx, gojit.Eax)

		if set := (op>>20)&1 != 0; set {
			j.Movb(gojit.Dl, jC)
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
		}

		j.Movl(gojit.Eax, j.REG(rd))
	},

	// MOV
	func(j *Jit, op, rd uint32) {
		j.Movl(gojit.Ebx, j.REG(rd))

		if set := (op>>20)&1 != 0; set {
			j.Movb(gojit.Dl, jC)
			j.Test(gojit.Ebx, gojit.Ebx)
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
		}
	},

	// BIC
	func(j *Jit, op, rd uint32) {
		j.Not(gojit.Ebx)
		j.And(gojit.Ebx, gojit.Eax)

		if set := (op>>20)&1 != 0; set {
			j.Movb(gojit.Dl, jC)
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
		}

		j.Movl(gojit.Eax, j.REG(rd))
	},

	// MVN
	func(j *Jit, op, rd uint32) {
		j.Not(gojit.Ebx)

		if set := (op>>20)&1 != 0; set {
			j.Movb(gojit.Dl, jC)
			j.Test(gojit.Ebx, gojit.Ebx)
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
		}

		j.Movl(gojit.Ebx, j.REG(rd))
	},
}

func (j *Jit) emitBlock(op uint32) {
	var (
		rlist      = op & 0xFFFF
		rn         = (op >> 16) & 0xF
		pcIncluded = rlist&0x8000 != 0
		pre        = (op>>24)&1 != 0
		up         = (op>>23)&1 != 0
		psr        = (op>>22)&1 != 0
		wb         = (op>>21)&1 != 0
		load       = (op>>20)&1 != 0

		first uint32
		bytes uint32
	)

	if rlist == 0 {
		rlist = 0x8000
		first = 0xF
		bytes = 16 * 4
		pcIncluded = true
	} else {
		bytes = uint32(bits.OnesCount32(rlist)) * 4
		first = uint32(bits.TrailingZeros32(rlist))
	}

	/*
	* 		R8: possible mode prev on user mode force
	* 		R9: Jit compiler CpuPtr
	* 		R10: Rn New
	* 		R11: addr (also Ebx)
	 */

	possibleForceUser := psr && (!load || !pcIncluded)

	j.Movl(gojit.Imm(0), gojit.R8d)

	if possibleForceUser {

		j.Movl(MODE, gojit.Ebx)
		j.Cmp(gojit.Imm(MODE_USR), gojit.Ebx)

		usr := j.JccForward(gojit.CC_Z)

		j.Cmp(gojit.Imm(MODE_SYS), gojit.Ebx)

		sys := j.JccForward(gojit.CC_Z)

		j.Mov(JIT, gojit.Rax)
		j.Movl(gojit.Ebx, gojit.R8d)
		j.Movl(gojit.Imm(MODE_USR), gojit.Ecx)
		j.CallFunc((*Jit).ModeSwitch)

		usr()
		sys()
	}

	j.Movl(j.REG(rn), gojit.R11d)
	j.Movl(gojit.R11d, gojit.R10d)

	// even when decrementing, cpu increments from "final" reg
	// see mgba https://mgba.io/2014/12/28/classic-nes/

	if up {
		j.Add(gojit.Imm(bytes), gojit.R10d)
	} else {
		pre = !pre

		j.Sub(gojit.Imm(bytes), gojit.R10d)
		j.Sub(gojit.Imm(bytes), gojit.R11d)
	}

	seq := uint32(NONSEQ)

	for i := first; i < 0x10; i++ {
		if disabled := rlist&(1<<i) == 0; disabled {
			continue
		}

		if pre {
			j.Add(gojit.Imm(4), gojit.R11d)
		}

		j.Mov(JIT, gojit.Rax)
		j.Movl(gojit.R11d, gojit.Ebx)

		if load {

			j.Movl(gojit.Imm(seq), gojit.Ecx)
			j.CallFunc((*Jit).Read32Block)

			if wb && i == first {
				j.Movl(gojit.R10d, j.REG(rn))
			}

			j.Movl(gojit.Eax, j.REG(i))

		} else {

			j.Movl(j.REG(i), gojit.Ecx)
			if i == PC {
				j.Add(gojit.Imm(4), gojit.Ecx)
			}

			j.Movl(gojit.Imm(seq), gojit.Edi)
			j.CallFunc((*Jit).Write32Block)

			if wb && i == first {
				j.Movl(gojit.R10d, j.REG(rn))
			}
		}

		if !pre {
			j.Add(gojit.Imm(4), gojit.R11d)
		}

		seq = SEQ
	}

	if possibleForceUser {

		j.Cmp(gojit.Imm(0), gojit.R8d)

		notForceUser := j.JccForward(gojit.CC_Z)

		j.Mov(JIT, gojit.Rax)
		j.Movl(gojit.Imm(MODE_USR), gojit.Ebx)
		j.Movl(gojit.R8d, gojit.Ecx)
		j.CallFunc((*Jit).ModeSwitch)

		notForceUser()
	}

	if load {

		j.Mov(JIT, gojit.Rax)
		j.Movl(gojit.Imm(1), gojit.Ebx)
		j.CallFunc((*Jit).Idle)

		if pcIncluded {
			panic("load with pc")
		}
	}
}
