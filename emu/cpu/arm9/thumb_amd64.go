package arm9

import (
	"github.com/aabalke/gojit"
	"github.com/aabalke/guac/emu/cpu/arm7"
)

func (j *Jit) EmitThumbBKPT(_ uint16) {
	j.EmitException(arm7.VEC_PREFETCH, arm7.MODE_ABT)
	j.ReloadState = arm7.RELOAD
}

func (j *Jit) EmitThumbShortBlx(op uint16) {
	nn := uint32(op&0x7FF) << 1

	j.Movl(j.C.R[PC], gojit.Eax)
	j.Sub(gojit.Imm(2), gojit.Eax)
	j.Or(gojit.Imm(1), gojit.Eax)

	j.Movl(j.C.R[LR], gojit.Ebx)
	j.Add(gojit.Imm(nn), gojit.Ebx)

	j.Movl(gojit.Ebx, j.C.R[PC])
	j.Movl(gojit.Eax, j.C.R[LR])

	j.EmitToggleThumb()
	j.ReloadState = arm7.RELOAD
}

func (j *Jit) EmitThumbHiBx(op uint16) {
	rs := (op >> 3) & 0xF

	j.Movl(j.C.R[rs], gojit.Eax)
	if rs == PC {
		j.And(gojit.Imm(^1), gojit.Eax)
	}

	if blx := op&(1<<7) != 0; blx {
		j.Movl(j.C.R[PC], gojit.Ebx)
		j.Sub(gojit.Imm(2), gojit.Ebx)
		j.Or(gojit.Imm(1), gojit.Ebx)
		j.Movl(gojit.Ebx, j.C.R[LR])
	}

	j.Movl(gojit.Eax, j.C.R[PC])

	j.EmitToggleThumb()

	j.Movb(gojit.Imm(0), j.C.Reload)
	j.ReloadState = arm7.RELOAD
}

func (j *Jit) EmitThumbLDSH(op uint16) {
	j.Mov(arm7.JIT, gojit.Rax)

	j.Movl(j.C.R[(op>>3)&7], gojit.Ebx)
	j.Add(j.C.R[(op>>6)&7], gojit.Ebx)

	j.CallFunc((*Jit).Read16)
	j.Movsx(gojit.Ax, gojit.Eax)

	j.Movl(gojit.Eax, j.C.R[op&7])
}

func (j *Jit) EmitThumbPopPc(op uint16) {
	seq := j.EmitThumbPushPop(op)

	j.Mov(arm7.JIT, gojit.Rax)
	j.Movl(j.C.R[SP], gojit.Ebx)
	j.Movl(gojit.Imm(seq), gojit.Ecx)

	j.CallFunc((*Jit).Read32Block)
	j.Movl(gojit.Eax, j.C.R[PC])
	j.Add(gojit.Imm(4), j.C.R[SP])

	//j.Mov(arm7.JIT, gojit.Rax)
	//j.Movl(gojit.Imm(1), gojit.Ebx)
	//j.CallFunc((*Jit).Idle)

	j.EmitToggleThumb()
	j.ReloadState = arm7.RELOAD
}
