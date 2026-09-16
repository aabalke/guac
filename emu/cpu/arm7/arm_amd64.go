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

		j.CallFunc(Idle)

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

		j.CallFunc(Idle)

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

		j.CallFunc(Idle)

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

	j.Movl(j.REG(rn), gojit.Eax)
	j.Movl(j.REG(rm), gojit.Ebx)

	j.Movl(gojit.Eax, gojit.R8d)
	j.Movl(gojit.Ebx, gojit.Esi)

	if isByte := (op>>22)&1 != 0; isByte {

		j.CallFunc(Read8)
		j.Movl(gojit.Eax, j.REG(rd))

		j.Movl(gojit.R8d, gojit.Eax)
		j.Movl(gojit.Esi, gojit.Ebx)

		j.CallFunc(Write8)
		return
	}

	j.CallFunc(Read32)

	j.Movl(gojit.R8d, gojit.Ecx)
	j.And(gojit.Imm(3), gojit.Ecx)
	j.Shl(gojit.Imm(3), gojit.Ecx)

	j.RorCl(gojit.Eax)
	j.Movl(gojit.Eax, j.REG(rd))

	j.Movl(gojit.R8d, gojit.Eax)
	j.Movl(gojit.Esi, gojit.Ebx)
	j.CallFunc(Write32)
}

func (j *Jit) emitHalf(op uint32) {
	var (
		rn      = (op >> 16) & 0xF
		rd      = (op >> 12) & 0xF
		preFlag = (op>>24)&1 != 0
		load    = (op>>20)&1 != 0
		inst    = (op >> 5) & 3
		wb      = (op>>21)&1 != 0 || !preFlag
	)

	j.Movl(j.REG(rn), gojit.Eax)

	if imm := (op>>22)&1 != 0; imm {
		j.Mov(gojit.Imm((op&0xF)|((op>>4)&0xF0)), gojit.Edi)
	} else {
		j.Movl(j.REG(op&0xF), gojit.Edi)
	}

	j.Mov(gojit.Rax, gojit.Rcx)

	if up := (op>>23)&1 != 0; up {
		j.Add(gojit.Edi, gojit.Ecx)
	} else {
		j.Sub(gojit.Edi, gojit.Ecx)
	}

	if preFlag {
		j.Mov(gojit.Rcx, gojit.Rax)
	}

	if !load {

		if inst != STRH {
			panic("unsupported arm7 instruction (ldrd, strd, reserved)")
		}

		j.Movl(j.REG(rd), gojit.Ebx)

		if wb {
			j.Movl(gojit.Ecx, j.REG(rn))
		}

		j.CallFunc(Write16)
		return
	}

	if wb {
		j.Movl(gojit.Ecx, j.REG(rn))
	}

	switch inst {
	case LDRH:
		j.Movl(gojit.Eax, gojit.R8d)
		j.CallFunc(Read16)

		j.Movl(gojit.R8d, gojit.Ecx)
		j.And(gojit.Imm(1), gojit.Ecx)
		j.Shl(gojit.Imm(3), gojit.Ecx)
		j.RorCl(gojit.Eax)
		j.Movl(gojit.Eax, j.REG(rd))

	case LDRSB:
		// sign-expand byte value
		j.CallFunc(Read8)
		j.Movsx(gojit.Al, gojit.Rax)
		j.Movl(gojit.Eax, j.REG(rd))

	case LDRSH:

		j.Bt(gojit.Imm(0), gojit.Eax)
		half := j.JccForward(gojit.CC_NC)

		// sign-expand byte value
		j.CallFunc(Read8)
		j.Movsx(gojit.Al, gojit.Rax)
		j.Movl(gojit.Eax, j.REG(rd))
		byte := j.JmpForward()

		// sign-expand half value
		half()
		j.CallFunc(Read16)
		j.Movsx(gojit.Ax, gojit.Rax)
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

	if reg {
		j.emitSdtRegShift(op)

		j.Mov(gojit.Rbx, gojit.Rcx)
	} else {
		j.Mov(gojit.Imm(int32(op&0xFFF)), gojit.Rcx)
	}

	// rn ebx, shift ecx
	// ax pre
	// bc post

	j.Movl(j.REG(rn), gojit.Ebx)

	if up {
		j.Add(gojit.Ecx, gojit.Ebx)
	} else {
		j.Sub(gojit.Ecx, gojit.Ebx)
	}

	if pre {
		j.Movl(gojit.Ebx, gojit.Eax)
	} else {
		j.Movl(j.REG(rn), gojit.Eax)
	}

	if wb {
		j.Movl(gojit.Ebx, j.REG(rn))
	}

	if load {
		if byte {
			j.CallFunc(Read8)
			j.Movl(gojit.Eax, j.REG(rd))
		} else {
			j.Movl(gojit.Eax, gojit.R8d)

			j.CallFunc(Read32)

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

		j.Movl(j.REG(rd), gojit.Ebx)
		if rd == PC {
			j.Add(gojit.Imm(4), gojit.Ebx)
		}

		if byte {
			j.CallFunc(Write8)
		} else {
			j.CallFunc(Write32)
		}
	}
}

func (j *Jit) emitSdtRegShift(op uint32) {
	rm := op & 0xF
	shType := (op >> 5) & 0b11
	shift := (op >> 7) & 0x1F

	// ebx rm, ecx shift

	j.Movl(j.REG(rm), gojit.Ebx)

	if rm == PC {
		panic("rm cannot include pc sdt")
	}

	switch shType {
	case LSL:

		j.Mov(gojit.Imm(shift), gojit.Rcx)
		j.ShlCl(gojit.Ebx)

		j.Cmp(gojit.Imm(32), gojit.Ecx)
		j.Sbb(gojit.Eax, gojit.Eax)
		j.And(gojit.Eax, gojit.Ebx)
		return

	case LSR:

		j.Mov(gojit.Imm(shift), gojit.Rcx)
		j.ShrCl(gojit.Ebx)

		j.Cmp(gojit.Imm(32), gojit.Ecx)
		j.Sbb(gojit.Eax, gojit.Eax)
		j.And(gojit.Eax, gojit.Ebx)
		return

	case ASR:

		j.Movl(gojit.Imm(shift), gojit.Ecx)

		j.Cmp(gojit.Imm(32), gojit.Ecx)
		j.Sbb(gojit.Eax, gojit.Eax)
		j.Not(gojit.Eax)
		j.Or(gojit.Eax, gojit.Ecx)

		j.SarCl(gojit.Ebx)
		return

	case ROR:

		j.Movl(gojit.Imm(shift), gojit.Ecx)
		j.RorCl(gojit.Ebx)
		return
	}
}

func (j *Jit) emitMrs(op uint32) {
	rd := (op >> 12) & 0xF

	if spsr := (op>>22)&1 != 0; spsr {
		j.Movl(MODE, gojit.Eax)
		j.CallFunc(GetSpsr)
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
	var (
		shReg    = (op>>4)&1 != 0
		shType   = (op >> 5) & 3
		setCarry = (op>>20)&1 != 0
		inst     = (op >> 21) & 0xF
		logical  = inst&0b0110 == 0b0000 || inst&0b1100 == 0b1100
		rm       = op & 0xF
	)

	if shReg {
		j.Movl(gojit.Imm(1), gojit.Eax)
		j.CallFunc(Idle)
	}

	setCarry = setCarry && logical
	if setCarry {
		j.Mov(gojit.Imm(1), gojit.Rdi)
	} else {
		j.Mov(gojit.Imm(0), gojit.Rdi)
	}

	// rbx: op2
	// rcx: shift
	// rdx: original carry
	// rdi: setcarry

	j.Movb(jC, gojit.Dl)
	j.Movl(j.REG(rm), gojit.Ebx)

	if shReg {
		rs := (op >> 8) & 0xF

		j.Movl(j.REG(rs), gojit.Ecx)
		j.And(gojit.Imm(0xFF), gojit.Ecx)

		if rm == PC {
			j.Add(gojit.Imm(4), gojit.Ebx)
		}

	} else {

		shift := (op >> 7) & 0x1F

		if special := shift == 0; special {
			switch shType {
			case LSL:
			case LSR:

				j.Bt(gojit.Imm(31), gojit.Ebx)
				j.SETcc(gojit.CC_C, jC)

				// clear op2
				j.Xor(gojit.Ebx, gojit.Ebx)
			case ASR:

				// sar sets everything to top bit
				// if setcarry, set carry to bit as well

				j.Sar(gojit.Imm(31), gojit.Ebx)

				if setCarry {
					j.Bt(gojit.Imm(31), gojit.Ebx)
					j.SETcc(gojit.CC_C, jC)
				}

			case ROR:

				// CF = old carry (EDX & 1)
				j.Bt(gojit.Imm(0), gojit.Edx)

				// RRX
				j.Rcr(gojit.Imm(1), gojit.Ebx)

				// CPSR.C = new carry
				j.SETcc(gojit.CC_C, jC)
			}

			return
		}

		j.Mov(gojit.Imm(shift), gojit.Rcx)
	}

	j.Test(gojit.Rcx, gojit.Rcx)
	zeroJump := j.JccForward(gojit.CC_Z)

	// https://iitd-plos.github.io/col718/ref/arm-instructionset.pdf

	switch shType {
	case LSL, LSR:

		j.Cmp(gojit.Imm(32), gojit.Ecx)
		shift32 := j.JccForward(gojit.CC_A)
		equal := j.JccForward(gojit.CC_Z)

		if shType == LSL {
			// carry = op2 & (1 << (32-shift)) != 0
			j.Mov(gojit.Imm(32), gojit.Rax)
			j.Sub(gojit.Ecx, gojit.Eax)
			j.Bt(gojit.Eax, gojit.Ebx)
			j.SETcc(gojit.CC_C, gojit.Dl)
			// op2 <<= shift
			j.ShlCl(gojit.Ebx)
		} else {
			// carry = op2 & (1 << (shift-1)) != 0
			j.Mov(gojit.Rcx, gojit.Rax)
			j.Sub(gojit.Imm(1), gojit.Eax)
			j.Bt(gojit.Eax, gojit.Ebx)
			j.SETcc(gojit.CC_C, gojit.Dl)
			// op2 >>= shift
			j.ShrCl(gojit.Ebx)
		}

		done := j.JmpForward()

		shift32()

		// carry = op2 & 1 != 0
		j.Mov(gojit.Rbx, gojit.Rdx)
		j.And(gojit.Imm(1), gojit.Rdx)
		// op2 = 0
		j.Xor(gojit.Rbx, gojit.Rbx)

		done2 := j.JmpForward()

		// this causes errors???
		equal()

		// LSL: carry = op2 & 1 != 0
		// LSR: carry = op2 & 0x8000_0000 != 0
		j.Mov(gojit.Rbx, gojit.Rdx)
		if shType == LSL {
			j.And(gojit.Imm(1), gojit.Rdx)
		} else {
			j.Shr(gojit.Imm(31), gojit.Rdx)
		}

		// op2 = 0
		j.Xor(gojit.Rbx, gojit.Rbx)

		done()
		done2()

	case ASR:

		j.Cmp(gojit.Imm(32), gojit.Ecx)
		shift32ge := j.JccForward(gojit.CC_AE)

		// carry = op2 & (1 << (shift-1)) != 0
		j.Mov(gojit.Rcx, gojit.Rax)
		j.Sub(gojit.Imm(1), gojit.Eax)
		j.Bt(gojit.Eax, gojit.Ebx)
		j.SETcc(gojit.CC_C, gojit.Dl)
		// op2 <<= shift
		j.SarCl(gojit.Ebx)

		done := j.JmpForward()

		shift32ge()

		// op and carry == top bit sar
		j.Sar(gojit.Imm(31), gojit.Ebx)
		j.Bt(gojit.Imm(0), gojit.Ebx)
		j.SETcc(gojit.CC_C, gojit.Dl)

		done()

	case ROR:

		j.Cmp(gojit.Imm(32), gojit.Ecx)
		equal := j.JccForward(gojit.CC_Z)

		// carry = (op2 >> ((shift-1) & 31)) & 1 != 0
		j.Mov(gojit.Rcx, gojit.Rax)
		j.Sub(gojit.Imm(1), gojit.Eax)
		j.And(gojit.Imm(31), gojit.Eax)

		j.Bt(gojit.Eax, gojit.Ebx)
		j.SETcc(gojit.CC_C, gojit.Dl)

		// op2 ror shift
		j.RorCl(gojit.Ebx)

		done := j.JmpForward()

		equal()

		// op2 unchanged
		// carry = op2 & 0x8000_0000 != 0
		j.Mov(gojit.Rbx, gojit.Rdx)
		j.Shr(gojit.Imm(31), gojit.Rdx)

		done()
	}

	j.Test(gojit.Rdi, gojit.Rdi)

	skip := j.JccForward(gojit.CC_Z)

	j.Movb(gojit.Dl, jC)

	zeroJump()
	skip()
}

func (j *Jit) emitAlu(op uint32) {
	var (
		inst = (op >> 21) & 0xF
		rd   = (op >> 12) & 0xF
		rn   = (op >> 16) & 0xF
		set  = (op>>20)&1 != 0
	)

	if rd == PC {
		panic("rd == pc")
	}

	if inst == 5 || inst == 7 || inst == 6 {
		j.Xor(gojit.Rcx, gojit.Rcx)
		j.Movb(jC, gojit.Cl)
		j.Mov(gojit.Rcx, gojit.R8)
	}

	if imm := (op>>25)&1 != 0; imm {

		ro := ((op >> 8) & 0xF) << 1
		op2 := bits.RotateLeft32(op&0xFF, -int(ro))

		j.Mov(gojit.Imm(int32(op2)), gojit.Rbx)

		if set && ro != 0 {
			j.Movb(gojit.Imm((op2>>31)&1), jC)
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

	if inst == 5 || inst == 7 || inst == 6 {
		j.Mov(gojit.R8, gojit.Rcx)
	}

	aluInstJit[inst](j, op, rd)

	j.Movl(j.REG(PC), gojit.Eax)
}

var aluInstJit = [...]func(j *Jit, op, rd uint32){
	// AND
	func(j *Jit, op, rd uint32) {
		j.And(gojit.Ebx, gojit.Eax)

		if set := (op>>20)&1 != 0; set {
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
		}

		j.Movl(gojit.Eax, j.REG(rd))
	},

	// EOR
	func(j *Jit, op, rd uint32) {
		j.Xor(gojit.Ebx, gojit.Eax)

		if set := (op>>20)&1 != 0; set {
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
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
		}

		j.Movl(gojit.Eax, j.REG(rd))
	},

	// MOV
	func(j *Jit, op, rd uint32) {
		j.Movl(gojit.Ebx, j.REG(rd))

		if set := (op>>20)&1 != 0; set {
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
			j.SETcc(gojit.CC_S, jN)
			j.SETcc(gojit.CC_Z, jZ)
		}

		j.Movl(gojit.Eax, j.REG(rd))
	},

	// MVN
	func(j *Jit, op, rd uint32) {
		j.Not(gojit.Ebx)

		if set := (op>>20)&1 != 0; set {
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
	* 		R11: addr (also Rax)
	 */

	possibleForceUser := psr && (!load || !pcIncluded)

	j.Movl(gojit.Imm(0), gojit.R8d)

	if possibleForceUser {

		j.Movl(MODE, gojit.Eax)
		j.Cmp(gojit.Imm(MODE_USR), gojit.Eax)

		usr := j.JccForward(gojit.CC_Z)

		j.Cmp(gojit.Imm(MODE_SYS), gojit.Eax)

		sys := j.JccForward(gojit.CC_Z)

		j.Movl(gojit.Eax, gojit.R8d)
		j.Movl(gojit.Imm(MODE_USR), gojit.Ebx)
		j.CallFunc(ModeSwitch)

		usr()
		sys()
	}

	j.Movl(j.REG(rn), gojit.Eax)
	j.Movl(gojit.Eax, gojit.R10d)

	// even when decrementing, cpu increments from "final" reg
	// see mgba https://mgba.io/2014/12/28/classic-nes/

	if up {
		j.Add(gojit.Imm(bytes), gojit.R10d)
	} else {
		pre = !pre

		j.Sub(gojit.Imm(bytes), gojit.R10d)
		j.Sub(gojit.Imm(bytes), gojit.Eax)
	}

	seq := uint32(NONSEQ)

	for i := first; i < 0x10; i++ {
		if disabled := rlist&(1<<i) == 0; disabled {
			continue
		}

		if pre {
			j.Add(gojit.Imm(4), gojit.Eax)
		}

		j.Movl(gojit.Eax, gojit.R11d)

		if load {

			j.Movl(gojit.Imm(seq), gojit.Ebx)

			j.CallFunc(Read32Block)

			if wb && i == first {
				j.Movl(gojit.R10d, j.REG(rn))
			}

			j.Movl(gojit.Eax, j.REG(i))

		} else {

			j.Movl(j.REG(i), gojit.Ebx)

			if i == PC {
				j.Add(gojit.Imm(4), gojit.Ebx)
			}

			j.Movl(gojit.Imm(seq), gojit.Ecx)

			j.CallFunc(Write32Block)

			if wb && i == first {
				j.Movl(gojit.R10d, j.REG(rn))
			}
		}

		j.Movl(gojit.R11d, gojit.Eax)

		if !pre {
			j.Add(gojit.Imm(4), gojit.Eax)
		}

		seq = SEQ
	}

	if possibleForceUser {

		j.Cmp(gojit.Imm(0), gojit.R8d)

		notForceUser := j.JccForward(gojit.CC_Z)

		j.Movl(gojit.Imm(MODE_USR), gojit.Eax)
		j.Movl(gojit.R8d, gojit.Ebx)
		j.CallFunc(ModeSwitch)

		notForceUser()
	}

	if load {

		j.Movl(gojit.Imm(1), gojit.Eax)
		j.CallFunc(Idle)

		if pcIncluded {
			panic("load with pc")
		}
	}
}
