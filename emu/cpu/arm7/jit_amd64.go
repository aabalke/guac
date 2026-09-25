package arm7

import (
	"fmt"
	"unsafe"

	"github.com/aabalke/gojit"
)

func (j *Jit) CreateBlock(pc, w uint32) {
	pageIdx := pc >> j.Config.PageShift
	blockIdx := (pc & j.Config.PageMask) >> 1

	page := j.Pages[pageIdx]
	if page == nil {

		page = &Page{
			id:     pageIdx,
			Blocks: make([]*JitBlock, (1<<j.Config.PageShift)>>1),
		}

		j.Pages[pageIdx] = page
	} else if page.dead {
		println("page dead, block not created")
		return
	} else if block := page.Blocks[blockIdx]; block != nil && block.Skip {
		return
	}

	block := j.BlockCache.AssignBlock(j)
	if block == nil {
		return
	}

	j.Assembler = block.assembler

	j.MovAbs(uint64(uintptr(unsafe.Pointer(j))), JIT)
	j.MovAbs(uint64(j.C.Cpu), CPU)

	// offset for pipelining
	realPc := (pc - (w * 2)) &^ (w - 1)
	p := j.cpu.Mem.ReadPtr(realPc)
	if p == nil {
		j.BlockCache.PushTail(block)
		page.Blocks[blockIdx] = j.BlockCache.SkipBlock
		return
	}

	var size uint32
	for size < j.Config.MaxInstCnt {

		if reloaded := j.TryEmitOp(p, w); reloaded {
			break
		}

		size++
		p = unsafe.Add(p, w)
	}

	if size < j.Config.MinInstCnt {
		j.BlockCache.PushTail(block)
		page.Blocks[blockIdx] = j.BlockCache.SkipBlock
		return
	}

	j.Exit()

	if err := j.Error(); err != nil {
		panic(err)
	}

	block.initPc = pc
	block.Size = size
	block.f = func() {
		gojit.CallJit(uintptr(unsafe.Pointer(&block.assembler.Buf[0])))
	}

	page.Blocks[blockIdx] = block

	//if w == 2 {
	//	fmt.Printf("Block Created for Page %08X PC %08X EXIT PC %08X OP %04X\n", pageIdx, pc, tempPc, uint16(op))
	//} else {
	//	fmt.Printf("Block Created for Page %08X PC %08X EXIT PC %08X OP %08X\n", pageIdx, pc, tempPc, op)
	//}
}

func (j *Jit) TryEmitOp(p unsafe.Pointer, w uint32) bool {
	op := *(*uint32)(p)

	endBlock := false

	j.Mov(JIT, gojit.Rax)
	j.CallFunc((*Jit).Step)
	j.Test(gojit.Ax, gojit.Ax)
	irq := j.JccForward(gojit.CC_NZ)

	if w == 4 {
		condTargets := j.emitCond(op >> 28)

		j.emitArm(op)
		done := j.JmpForward()

		j.Movl(gojit.Imm(SEQ), j.C.Seq)

		for _, target := range condTargets {
			target()
		}

		done()

	} else {
		j.emitThumb(uint16(op))
	}

	// NOTE: condition branching instructions require ending the block
	// but need to use c.Reload to find out if branch was taken
	reloadState := j.ReloadState
	j.ReloadState = NONE
	switch reloadState {
	case NONE:
		endBlock = false
		j.Mov(JIT, gojit.Rax)
		j.MovAbs(uint64(uintptr(p)), gojit.Rbx)
		j.Movl(gojit.Imm(w), gojit.Ecx)
		j.CallFunc((*Jit).UpdatePc)
	case RELOAD:
		endBlock = true
		j.Mov(JIT, gojit.Rax)
		j.CallFunc((*Jit).ReloadPipe)

	case POSSIBLE:
		endBlock = true

		j.Movb(j.C.Reload, gojit.Al)
		j.Testb(gojit.Al, gojit.Al)

		reload := j.JccForward(gojit.CC_NZ)

		j.Mov(JIT, gojit.Rax)
		j.MovAbs(uint64(uintptr(p)), gojit.Rbx)
		j.Movl(gojit.Imm(w), gojit.Ecx)
		j.CallFunc((*Jit).UpdatePc)

		notReload := j.JmpForward()
		reload()

		j.Mov(JIT, gojit.Rax)
		j.CallFunc((*Jit).ReloadPipe)

		notReload()
	}

	irq()

	//if w == 2 {
	//	fmt.Printf("emitOp PC %08X OP %04X\n", tempPc, uint16(op))
	//} else {
	//	fmt.Printf("emitOp PC %08X OP %08X\n", tempPc, op)
	//}

	return endBlock
}

func (j *Jit) emitCond(cond uint32) []func() {
	var jcctargets []func()

	switch cond {
	case 0xE, 0xF:
		// nothing to do, always executed
	case 0x0: // Z
		j.Bt(gojit.Imm(0), j.C.Z)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_NC))
	case 0x1: // !Z
		j.Bt(gojit.Imm(0), j.C.Z)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_C))
	case 0x2: // C
		j.Bt(gojit.Imm(0), j.C.C)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_NC))
	case 0x3: // !C
		j.Bt(gojit.Imm(0), j.C.C)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_C))
	case 0x4: // N
		j.Bt(gojit.Imm(0), j.C.N)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_NC))
	case 0x5: // !N
		j.Bt(gojit.Imm(0), j.C.N)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_C))
	case 0x6: // V
		j.Bt(gojit.Imm(0), j.C.V)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_NC))
	case 0x7: // !V
		j.Bt(gojit.Imm(0), j.C.V)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_C))
	case 0x8: // C && !Z
		j.Bt(gojit.Imm(0), j.C.C)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_NC))
		j.Bt(gojit.Imm(0), j.C.Z)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_C))
	case 0x9: // !C || Z
		j.Movb(j.C.C, gojit.Al)
		j.Xorb(gojit.Imm(1), gojit.Al)
		j.Orb(j.C.Z, gojit.Al)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_Z))
	case 0xC: // !Z && N==V
		j.Bt(gojit.Imm(0), j.C.Z)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_C))
		fallthrough
	case 0xA, 0xB: // N==V / N!=V
		j.Movb(j.C.N, gojit.Al)
		j.Xorb(j.C.V, gojit.Al)
		if cond == 0xA || cond == 0xC {
			jcctargets = append(jcctargets, j.JccForward(gojit.CC_NZ))
		} else {
			jcctargets = append(jcctargets, j.JccForward(gojit.CC_Z))
		}
	case 0xD: // Z || N==V / N!=V
		j.Movb(j.C.N, gojit.Al)
		j.Xorb(j.C.V, gojit.Al)
		j.Orb(j.C.Z, gojit.Al)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_Z))
	}

	return jcctargets
}

func (j *Jit) emitArm(op uint32) {
	switch {
	case (op>>24)&0xF == 0xF:
		j.emitSWI(op)
	case IsBranch(op):
		j.emitBranch(op)
	case IsBranchExchange(op):
		j.emitBranchExchange(op)
	case IsSdt(op):
		j.emitSdt(op)
	case IsBlock(op):
		j.emitBlock(op)
	case IsHalf(op):
		j.emitHalf(op)
	case IsUndefined(op):
		j.emitUndefined(op)
	case IsMsr(op):
		j.emitMsr(op)
	case IsMrs(op):
		j.emitMrs(op)
	case IsSwp(op):
		j.emitSwp(op)
	case IsMul(op):
		j.emitMul(op)
	case IsAlu(op):
		j.emitAlu(op)
	default:
		panic(fmt.Sprintf("unemittable amd64 jit instruction ARM OP %08X", op))
	}
}

func (j *Jit) emitThumb(op uint16) {
	switch {
	case IsThumbSWI(op):
		j.emitThumbSWI(op)
	case IsThumbAddSub(op):
		j.emitThumbAddSub(op)
	case IsThumbShift(op):
		j.emitThumbShifted(op)
	case IsThumbImm(op):
		j.emitThumbImm(op)
	case IsThumbAlu(op):
		j.emitThumbAlu(op)
	case IsThumbHi(op):
		j.emitThumbHi(op)
	case IsLSHalf(op):
		j.emitThumbLSHalf(op)
	case IsThumbSdt(op):
		j.emitThumbSdt(op)
	case IsLPC(op):
		j.emitThumbLPC(op)
	case IsLSImm(op):
		j.emitThumbLSImm(op)
	case IsPushPop(op):
		j.emitThumbPushPop(op)
	case IsRelative(op):
		j.emitThumbRelative(op)
	case IsThumbBranch(op):
		j.emitThumbBranch(op)
	case IsJumpCall(op):
		j.emitJumpCall(op)
	case IsStack(op):
		j.emitThumbStack(op)
	case IsLongBranch(op):
		j.emitLongBranch(op)
	case IsShortLongBranch(op):
		j.emitShortLongBranch(op)
	case IsLSSP(op):
		j.emitThumbLSSP(op)
	case IsThumbBlock(op):
		j.emitThumbBlock(op)
	default:
		panic(fmt.Sprintf("unemittable amd64 jit instruction THUMB OP %04X", op))
	}
}

func (j *Jit) emitToggleThumb() {
	j.Movl(j.C.R[PC], gojit.Eax)
	j.And(gojit.Imm(1), gojit.Eax)
	j.Movb(gojit.Al, j.C.T)

	j.Movb(gojit.Imm(1), j.C.Reload)

	j.Testb(gojit.Al, gojit.Al)

	j.Movl(gojit.Imm(^1), gojit.Eax)
	j.Movl(gojit.Imm(^3), gojit.Ebx)
	j.Cmovcc(gojit.CC_Z, gojit.Ebx, gojit.Eax)

	j.And(gojit.Eax, j.C.R[PC])
}
