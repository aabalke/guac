package arm9

import (
	"fmt"

	"github.com/aabalke/guac/emu/cpu/arm7"
)

func (j *Jit) EmitArm(op uint32) {
	// this will have to be placed differently
	if op>>28 == 0xF {
		switch {
		case IsBlx(op):
			j.EmitBlx(op)
			return
		case IsPld(op):
			panic("unsetup pld instruction")
		}
	}

	switch {
	case (op>>24)&0xF == 0xF:
		j.EmitSWI(op)
	case IsBkpt(op):
		j.EmitBKPT(op)
	case IsCoDataReg(op):
		j.EmitCoDataReg(op)
	case arm7.IsBranch(op):
		j.EmitBranch(op)
	case arm7.IsBranchExchange(op):
		j.EmitBranchExchange(op)
	case arm7.IsSdt(op):
		j.EmitSdt(op)
	case arm7.IsBlock(op):
		j.EmitBlock(op)
	case arm7.IsHalf(op):
		j.EmitHalf(op)
	case arm7.IsUndefined(op):
		j.EmitUndefined(op)
	case arm7.IsMrs(op):
		j.EmitMrs(op)
	case arm7.IsMsr(op):
		j.EmitMsr(op)
	case arm7.IsSwp(op):
		j.EmitSwp(op)
	case arm7.IsMul(op):
		j.EmitMul(op)
	case IsCLZ(op):
		j.EmitClz(op)
	case IsQAlu(op):
		j.EmitQalu(op)
	case IsExtMul(op):
		j.EmitExtendedMul(op)
	case arm7.IsAlu(op):
		j.EmitAlu(op)
	default:
		panic(fmt.Sprintf("unemittable amd64 jit instruction ARM OP %08X", op))
	}
}

func (j *Jit) EmitThumb(op uint16) {
	switch {
	case IsThumbBkpt(op):
		j.EmitThumbBKPT(op)
	case arm7.IsThumbSWI(op):
		j.EmitThumbSWI(op)
	case arm7.IsThumbAddSub(op):
		j.EmitThumbAddSub(op)
	case arm7.IsThumbShift(op):
		j.EmitThumbShifted(op)
	case arm7.IsThumbImm(op):
		j.EmitThumbImm(op)
	case arm7.IsThumbAlu(op):
		j.EmitThumbAlu(op)
	case arm7.IsThumbHiBx(op):
		j.EmitThumbHiBx(op)
	case arm7.IsThumbHi(op):
		j.EmitThumbHi(op)
	case arm7.IsLSHalf(op):
		j.EmitThumbLSHalf(op)
	case arm7.IsThumbLDSH(op):
		j.EmitThumbLDSH(op)
	case arm7.IsThumbSdt(op):
		j.EmitThumbSdt(op)
	case arm7.IsLPC(op):
		j.EmitThumbLPC(op)
	case arm7.IsLSImm(op):
		j.EmitThumbLSImm(op)
	case arm7.IsPopPc(op):
		j.EmitThumbPopPc(op)
	case arm7.IsPushPop(op):
		j.EmitThumbPushPop(op)
	case arm7.IsRelative(op):
		j.EmitThumbRelative(op)
	case arm7.IsThumbBranch(op):
		j.EmitThumbBranch(op)
	case arm7.IsJumpCall(op):
		j.EmitJumpCall(op)
	case arm7.IsStack(op):
		j.EmitThumbStack(op)
	case arm7.IsLongBranch(op):
		j.EmitLongBranch(op)
	case arm7.IsShortLongBranch(op):
		j.EmitShortLongBranch(op)
	case IsThumbShortBlx(op):
		j.EmitThumbShortBlx(op)
	case arm7.IsLSSP(op):
		j.EmitThumbLSSP(op)
	case arm7.IsThumbBlock(op):
		j.EmitThumbBlock(op)
	default:
		panic(fmt.Sprintf("unemittable amd64 jit instruction THUMB OP %04X", op))
	}
}
