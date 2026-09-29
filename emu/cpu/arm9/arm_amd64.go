package arm9

import (
	"math"
	"math/bits"

	"github.com/aabalke/gojit"
	"github.com/aabalke/guac/emu/cpu/arm7"
)

func (j *Jit) EmitBKPT(op uint32) {
	j.EmitException(arm7.VEC_PREFETCH, arm7.MODE_ABT)
	j.EmitCondReloadState(op >> 28)
}

func (j *Jit) EmitQalu(op uint32) {
	var (
		rn = (op >> 16) & 0xF
		rd = (op >> 12) & 0xF
		rm = op & 0xF
	)

	j.Mov(j.C.R[rm], gojit.Eax)
	j.Movsxd(gojit.Eax, gojit.Rax)

	j.Mov(j.C.R[rn], gojit.Ebx)
	j.Movsxd(gojit.Ebx, gojit.Rbx)

	if double := op&(1<<22) != 0; double {
		// double with add to get int32 overflow
		j.Add(gojit.Ebx, gojit.Ebx)
		j.emitQaluClamp(gojit.Ebx)
	}

	if sub := op&(1<<21) != 0; sub {
		j.Sub(gojit.Ebx, gojit.Eax)
	} else {
		j.Add(gojit.Ebx, gojit.Eax)
	}

	j.emitQaluClamp(gojit.Eax)

	j.Movl(gojit.Eax, j.C.R[rd])
}

func (j *Jit) emitQaluClamp(dst gojit.Register) {
	j.Mov(gojit.Imm(math.MaxInt32), gojit.Rcx)
	j.Mov(gojit.Imm(math.MinInt32), gojit.Rdx)

	j.Cmovcc(gojit.CC_NS, gojit.Edx, gojit.Ecx)
	j.Cmovcc(gojit.CC_O, gojit.Ecx, dst)

	j.SETcc(gojit.CC_O, gojit.Dl)
	j.Orb(gojit.Dl, j.C.Q)
}

func (j *Jit) EmitClz(op uint32) {
	rd := (op >> 12) & 0xF
	rm := op & 0xF

	j.Movl(j.C.R[rm], gojit.Eax)
	j.Lzcnt(gojit.Eax, gojit.Eax)

	// lzcnt returns op size when input is zero
	// j.Mov(gojit.Imm(0), gojit.Ebx)
	// j.Cmovcc(gojit.CC_Z, gojit.Ebx, gojit.Eax)

	j.Movl(gojit.Eax, j.C.R[rd])
}

func (j *Jit) EmitExtendedMul(op uint32) {
	var (
		rd = (op >> 16) & 0xF
		rn = (op >> 12) & 0xF
		rs = (op >> 8) & 0xF
		rm = op & 0xF
		x  = (op>>5)&1 != 0
		y  = (op>>6)&1 != 0
	)

	j.emitExtMulInt16(gojit.Rax, j.C.R[rs], y)

	switch inst := (op >> 21) & 3; inst {
	case SMULxy, SMLAxy:

		j.emitExtMulInt16(gojit.Rbx, j.C.R[rm], x)

		j.Imul(gojit.Rbx)

		if inst == SMLAxy {
			j.emitExtMulAdd(j.C.R[rn])
		}

	case SMLAWySMLALWy:

		// rmv
		j.Movl(j.C.R[rm], gojit.Ebx)
		j.Movsxd(gojit.Ebx, gojit.Rbx)

		j.Imul(gojit.Rbx)
		j.Sar(gojit.Imm(16), gojit.Rax)

		if !x {
			j.emitExtMulAdd(j.C.R[rn])
		}

	case SMLALxy:

		j.emitExtMulInt16(gojit.Rbx, j.C.R[rm], x)

		j.Imul(gojit.Rbx)

		j.Movl(j.C.R[rd], gojit.Ebx)
		j.Shl(gojit.Imm(32), gojit.Rbx)
		j.Add(gojit.Rbx, gojit.Rax)

		j.Movl(j.C.R[rn], gojit.Ebx)
		j.Movsxd(gojit.Ebx, gojit.Rbx)
		j.Add(gojit.Rbx, gojit.Rax)

		j.Movl(gojit.Eax, j.C.R[rn])
		j.Shr(gojit.Imm(32), gojit.Rax)
	}

	j.Movl(gojit.Eax, j.C.R[rd])
}

func (j *Jit) emitExtMulInt16(dst gojit.Register, src gojit.Indirect, shiftBit bool) {
	// dst = int64(int16(uint16(src >> (16 * shiftBit))))
	reg16 := gojit.Register{Val: dst.Val, Bits: 16}
	reg32 := gojit.Register{Val: dst.Val, Bits: 32}
	reg64 := gojit.Register{Val: dst.Val, Bits: 64}

	j.Movl(src, reg32)
	if shiftBit {
		j.Sar(gojit.Imm(16), reg32)
	}
	j.Movsx(reg16, reg64)
}

func (j *Jit) emitExtMulAdd(src gojit.Indirect) {
	j.Movl(src, gojit.Ebx)
	j.Movsxd(gojit.Ebx, gojit.Rbx)
	j.Add(gojit.Ebx, gojit.Eax)
	j.SETcc(gojit.CC_O, gojit.Dl)
	j.Orb(gojit.Dl, j.C.Q)
}

func (j *Jit) EmitBlx(op uint32) {
	j.Movl(j.C.R[PC], gojit.Eax)
	j.Sub(gojit.Imm(4), gojit.Eax)
	j.Movl(gojit.Eax, j.C.R[LR])

	jmp := uint32((int32(op)<<8)>>6) + ((op>>24)&1)<<1
	j.Add(gojit.Imm(jmp), j.C.R[PC])

	j.Movb(gojit.Imm(1), j.C.T)
	j.Movb(gojit.Imm(1), j.C.Reload)
	j.ReloadState = arm7.RELOAD
}

func (j *Jit) EmitCoDataReg(op uint32) {
	var (
		cn = (op >> 16) & 0xF
		cp = (op >> 5) & 7
		cm = (op >> 0) & 0xF
		rd = (op >> 12) & 0xF
	)

	if (op >> 28) == 0xF {
		panic("MRC2/MCR2")
	}

	if processor := uint8((op >> 8) & 0xF); processor != 15 {
		panic("co data register with pn != cp15")
	}

	if cpopc := (op >> 21) & 7; cpopc != 0 {
		panic("co data register with cpopc != 0")
	}

	idx := (cn << 8) | (cm << 4) | cp

	if mrc := (op>>20)&1 != 0; mrc {

		j.MovAbs(uint64(j.C.Cp15), gojit.Rax)
		j.Movl(gojit.Imm(idx), gojit.Ebx)

		j.CallFunc((*Cp15).Read)

		j.Movl(gojit.Eax, j.C.R[rd])

		j.Mov(arm7.JIT, gojit.Rax)
		j.Movl(gojit.Imm(3), gojit.Ebx)
		j.CallFunc((*Jit).Idle)
		return
	}

	j.MovAbs(uint64(j.C.Cp15), gojit.Rax)
	j.Movl(gojit.Imm(idx), gojit.Ebx)
	j.Movl(j.C.R[rd], gojit.Ecx)

	j.CallFunc((*Cp15).Write)

	j.Mov(arm7.JIT, gojit.Rax)
	j.Movl(gojit.Imm(2), gojit.Ebx)
	j.CallFunc((*Jit).Idle)
}

func (j *Jit) EmitBranchExchange(op uint32) {
	rn := op & 0xF

	j.Movl(j.C.R[rn], gojit.Eax)

	switch inst := (op >> 4) & 0xF; inst {
	case arm7.INST_BXJ:
		panic("unsupported bxj instruction")
	case arm7.INST_BLX:
		j.Movl(j.C.R[PC], gojit.Ebx)
		j.Sub(gojit.Imm(4), gojit.Ebx)
		j.Movl(gojit.Ebx, j.C.R[LR])
	}

	j.Movl(gojit.Eax, j.C.R[PC])

	j.EmitToggleThumb()
	j.EmitCondReloadState(op >> 28)
}

func (j *Jit) EmitBlock(op uint32) {
	var (
		rlist = op & 0xFFFF
		up    = (op>>23)&1 != 0
		rn    = (op >> 16) & 0xF
	)

	if rlist == 0 {
		if up {
			j.Add(gojit.Imm(0x40), j.C.R[rn])
		} else {
			j.Sub(gojit.Imm(0x40), j.C.R[rn])
		}
		return
	}

	var (
		pcIncluded = rlist&0x8000 != 0
		rnIncluded = (rlist>>rn)&1 != 0
		pre        = (op>>24)&1 != 0
		psr        = (op>>22)&1 != 0
		wb         = (op>>21)&1 != 0
		load       = (op>>20)&1 != 0
		bytes      = uint32(bits.OnesCount32(rlist)) * 4
		first      = uint32(bits.TrailingZeros32(rlist))
		regCount   = uint32(bits.OnesCount32(rlist))
		last       = bits.Len32(rlist) - 1
		isOnly     = regCount == 1
		isLast     = uint32(last) == rn
	)

	/*
	* 		R8: possible mode prev on user mode force
	* 		R9: Jit compiler CpuPtr
	* 		R10: Rn New
	* 		R11: addr (also Ebx)
	 */

	possibleForceUser := psr && (!load || !pcIncluded)

	j.Movl(j.C.R[rn], gojit.R11d)
	j.Movl(gojit.R11d, gojit.R10d)

	j.Movl(gojit.Imm(0), gojit.R8d)

	if possibleForceUser {

		j.Movl(j.C.Mode, gojit.Ebx)
		j.Cmp(gojit.Imm(arm7.MODE_USR), gojit.Ebx)

		usr := j.JccForward(gojit.CC_Z)

		j.Cmp(gojit.Imm(arm7.MODE_SYS), gojit.Ebx)

		sys := j.JccForward(gojit.CC_Z)

		j.Mov(arm7.JIT, gojit.Rax)
		j.Movl(gojit.Ebx, gojit.R8d)
		j.Movl(gojit.Imm(arm7.MODE_USR), gojit.Ecx)
		j.CallFunc((*Jit).ModeSwitch)

		usr()
		sys()
	}

	// even when decrementing, cpu increments from "final" reg
	// see mgba https://mgba.io/2014/12/28/classic-nes/

	if up {
		j.Add(gojit.Imm(bytes), gojit.R10d)
	} else {
		pre = !pre

		j.Sub(gojit.Imm(bytes), gojit.R10d)
		j.Sub(gojit.Imm(bytes), gojit.R11d)
	}

	seq := uint32(arm7.NONSEQ)

	for i := first; i < 0x10; i++ {
		if disabled := rlist&(1<<i) == 0; disabled {
			continue
		}

		if pre {
			j.Add(gojit.Imm(4), gojit.R11d)
		}

		j.Mov(arm7.JIT, gojit.Rax)
		j.Movl(gojit.R11d, gojit.Ebx)

		if load {
			j.Movl(gojit.Imm(seq), gojit.Ecx)
			j.CallFunc((*Jit).Read32Block)
			j.Movl(gojit.Eax, j.C.R[i])
		} else {
			j.Movl(j.C.R[i], gojit.Ecx)
			j.Movl(gojit.Imm(seq), gojit.Edi)
			j.CallFunc((*Jit).Write32Block)
		}

		if !pre {
			j.Add(gojit.Imm(4), gojit.R11d)
		}

		seq = arm7.SEQ
	}

	if possibleForceUser {

		j.Test(gojit.R8d, gojit.R8d)

		notForceUser := j.JccForward(gojit.CC_Z)

		j.Mov(arm7.JIT, gojit.Rax)
		j.Movl(gojit.Imm(arm7.MODE_USR), gojit.Ebx)
		j.Movl(gojit.R8d, gojit.Ecx)
		j.CallFunc((*Jit).ModeSwitch)

		notForceUser()
	}

	if wb {
		if load && rnIncluded {
			if isOnly || !isLast {
				j.Movl(gojit.R10d, j.C.R[rn])
			}
		} else {
			j.Movl(gojit.R10d, j.C.R[rn])
		}
	}

	if !load {
		return
	}

	//j.Mov(arm7.JIT, gojit.Rax)
	//j.Movl(gojit.Imm(1), gojit.Ebx)
	//j.CallFunc((*Jit).Idle)

	if !pcIncluded {
		return
	}

	j.EmitCondReloadState(op >> 28)

	if !psr {
		j.EmitToggleThumb()
		return
	}

	j.Mov(arm7.JIT, gojit.Rax)
	j.CallFunc((*Jit).DoLdmLoadSwitch)
}

func (j *Jit) EmitHalf(op uint32) {
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

	j.Mov(arm7.JIT, gojit.Rax)
	j.Movl(j.C.R[rn], gojit.Ebx)
	j.Movl(gojit.Ebx, gojit.Edx)

	switch {
	case imm && up:
		j.Add(gojit.Imm((op&0xF)|((op>>4)&0xF0)), gojit.Edx)
	case imm && !up:
		j.Sub(gojit.Imm((op&0xF)|((op>>4)&0xF0)), gojit.Edx)
	case !imm && up:
		j.Add(j.C.R[op&0xF], gojit.Edx)
	case !imm && !up:
		j.Sub(j.C.R[op&0xF], gojit.Edx)
	}

	if pre {
		j.Movl(gojit.Edx, gojit.Ebx)
	}

	j.Movl(j.C.R[rd], gojit.Ecx)

	if wb {
		j.Movl(gojit.Edx, j.C.R[rn])
	}

	if !load {

		switch inst {
		case arm7.STRH:
			j.CallFunc((*Jit).Write16)

		case arm7.LDRD:
			j.And(gojit.Imm(^7), gojit.Ebx)
			j.Movl(gojit.Ebx, gojit.R8d)

			j.CallFunc((*Jit).Read32)
			j.Movl(gojit.Eax, j.C.R[rd])

			j.Movl(gojit.R8d, gojit.Ebx)
			j.Add(gojit.Imm(4), gojit.Ebx)

			j.Mov(arm7.JIT, gojit.Rax)
			j.CallFunc((*Jit).Read32)
			j.Movl(gojit.Eax, j.C.R[rd+1])

		case arm7.STRD:
			j.And(gojit.Imm(^7), gojit.Ebx)
			j.Movl(gojit.Ebx, gojit.R8d)

			j.CallFunc((*Jit).Write32)

			j.Mov(arm7.JIT, gojit.Rax)

			j.Movl(gojit.R8d, gojit.Ebx)
			j.Add(gojit.Imm(4), gojit.Ebx)

			j.Movl(gojit.R9d, gojit.Ecx)

			j.CallFunc((*Jit).Write32)

		default:
			panic("unsupported arm9 instruction (reserved)")
		}

		return
	}

	switch inst {
	case arm7.LDRH:
		j.Movl(gojit.Ebx, gojit.R8d)
		j.CallFunc((*Jit).Read16)

		j.Movl(gojit.R8d, gojit.Ecx)
		j.And(gojit.Imm(1), gojit.Ecx)
		j.Shl(gojit.Imm(3), gojit.Ecx)
		j.RorCl(gojit.Eax)
		j.Movl(gojit.Eax, j.C.R[rd])

	case arm7.LDRSB:
		// sign-expand byte value
		j.CallFunc((*Jit).Read8)
		j.Movsx(gojit.Al, gojit.Eax)
		j.Movl(gojit.Eax, j.C.R[rd])

	case arm7.LDRSH:

		j.CallFunc((*Jit).Read16)
		j.Movsx(gojit.Ax, gojit.Eax)
		j.Movl(gojit.Eax, j.C.R[rd])

	default:
		panic("unsupported arm9 instruction (reserved)")
	}
}
