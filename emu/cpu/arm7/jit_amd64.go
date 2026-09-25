package arm7

import (
	"fmt"
	"unsafe"

	"github.com/aabalke/gojit"
	"github.com/aabalke/guac/config"
)

func (j *Jit) CreateBlock(pc, w uint32) {
	pageIdx := pc >> j.PageShift
	blockIdx := (pc & j.PageMask) >> 1

	page := j.Pages[pageIdx]

	if page == nil {
		page = &Page{
			id:     pageIdx,
			Blocks: make([]*JitBlock, (1<<j.PageShift)>>1),
		}

		j.Pages[pageIdx] = page

	} else if page.dead {
		println("page dead, block not created")
		return
	}

	if block := page.Blocks[blockIdx]; block != nil && block.Skip {
		return
	}

	newBlock := j.BlockCache.AssignBlock(j)
	if newBlock == nil {
		return
	}

	j.Assembler = newBlock.assembler

	j.MovAbs(uint64(uintptr(unsafe.Pointer(j))), JIT)
	j.MovAbs(uint64(uintptr(unsafe.Pointer(j.cpu))), CPU)

	// offset for pipelining
	tempPc := (pc - (w * 2)) &^ (w - 1)

	var length, op, i uint32

	p := j.cpu.Mem.ReadPtr(tempPc)
	if p == nil {

		if tempPc>>24 == 6 {
			panic("need to setup VRAM as work ram")
		}

		//panic(fmt.Sprintf("read ptr bad jit arm7 ADDR %08X", tempPc))
		j.BlockCache.PushTail(newBlock)
		page.Blocks[blockIdx] = j.BlockCache.SkipBlock
		return
	}

	for {
		op = *(*uint32)(p)

		if length >= config.Conf.Nds.Jit.BatchInstA7 {
			break
		}

		if reloaded := j.TryEmitOp(op, w); reloaded {
			break
		}

		//if w == 2 {
		//	fmt.Printf("emitOp PC %08X OP %04X\n", tempPc, uint16(op))
		//} else {
		//	fmt.Printf("emitOp PC %08X OP %08X\n", tempPc, op)
		//}

		i++
		length++
		tempPc += w
		p = unsafe.Add(p, w)
	}

	if length == 0 {
		j.BlockCache.PushTail(newBlock)
		page.Blocks[blockIdx] = j.BlockCache.SkipBlock
		return
	}

	j.Exit()

	if err := j.Error(); err != nil {
		panic(err)
	}

	newBlock.initPc = pc
	newBlock.Length = length
	newBlock.finalOp = op
	newBlock.f = func() {
		gojit.CallJit(uintptr(unsafe.Pointer(&newBlock.assembler.Buf[0])))
	}

	page.Blocks[blockIdx] = newBlock

	//if w == 2 {
	//	fmt.Printf("Block Created for Page %08X PC %08X EXIT PC %08X OP %04X\n", pageIdx, pc, tempPc, uint16(op))
	//} else {
	//	fmt.Printf("Block Created for Page %08X PC %08X EXIT PC %08X OP %08X\n", pageIdx, pc, tempPc, op)
	//}
}

func (j *Jit) TryEmitOp(op, w uint32) bool {
	endBlock := false

	//if w == 4 {
	//	return true
	//}

	j.Mov(JIT, gojit.Rax)
	j.CallFunc((*Jit).Step)
	j.Test(gojit.Ax, gojit.Ax)
	irq := j.JccForward(gojit.CC_NZ)

	if w == 4 {
		condTargets := j.emitCond(op >> 28)
		j.emitArm(op)
		for _, target := range condTargets {
			target()
		}
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
		j.Movl(gojit.Imm(w), gojit.Ebx)
		j.CallFunc((*Jit).UpdatePc)
	case RELOAD:
		endBlock = true
		j.Mov(JIT, gojit.Rax)
		j.CallFunc((*Jit).ReloadPipe)
		j.Mov(JIT, gojit.Rax)
		j.CallFunc((*Jit).DoJit)

	case POSSIBLE:
		endBlock = true

		j.Movb(RELOAD_FLAG, gojit.Al)
		j.Testb(gojit.Al, gojit.Al)

		reload := j.JccForward(gojit.CC_NZ)

		j.Mov(JIT, gojit.Rax)
		j.Movl(gojit.Imm(w), gojit.Ebx)
		j.CallFunc((*Jit).UpdatePc)

		notReload := j.JmpForward()
		reload()

		j.Mov(JIT, gojit.Rax)
		j.CallFunc((*Jit).ReloadPipe)
		j.Mov(JIT, gojit.Rax)
		j.CallFunc((*Jit).DoJit)

		notReload()
	}

	irq()

	return endBlock
}

func (j *Jit) emitCond(cond uint32) []func() {
	var jcctargets []func()

	switch cond {
	case 0xE, 0xF:
		// nothing to do, always executed
	case 0x0: // Z
		j.Bt(gojit.Imm(0), jZ)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_NC))
	case 0x1: // !Z
		j.Bt(gojit.Imm(0), jZ)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_C))
	case 0x2: // C
		j.Bt(gojit.Imm(0), jC)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_NC))
	case 0x3: // !C
		j.Bt(gojit.Imm(0), jC)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_C))
	case 0x4: // N
		j.Bt(gojit.Imm(0), jN)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_NC))
	case 0x5: // !N
		j.Bt(gojit.Imm(0), jN)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_C))
	case 0x6: // V
		j.Bt(gojit.Imm(0), jV)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_NC))
	case 0x7: // !V
		j.Bt(gojit.Imm(0), jV)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_C))
	case 0x8: // C && !Z
		j.Bt(gojit.Imm(0), jC)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_NC))
		j.Bt(gojit.Imm(0), jZ)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_C))
	case 0x9: // !C || Z
		j.Movb(jC, gojit.Al)
		j.Xorb(gojit.Imm(1), gojit.Al)
		j.Orb(jZ, gojit.Al)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_Z))
	case 0xC: // !Z && N==V
		j.Bt(gojit.Imm(0), jZ)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_C))
		fallthrough
	case 0xA, 0xB: // N==V / N!=V
		j.Movb(jN, gojit.Al)
		j.Xorb(jV, gojit.Al)
		if cond == 0xA || cond == 0xC {
			jcctargets = append(jcctargets, j.JccForward(gojit.CC_NZ))
		} else {
			jcctargets = append(jcctargets, j.JccForward(gojit.CC_Z))
		}
	case 0xD: // Z || N==V / N!=V
		j.Movb(jN, gojit.Al)
		j.Xorb(jV, gojit.Al)
		j.Orb(jZ, gojit.Al)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_Z))
	default:
		panic("not possible")
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
	case IsthumbSWI(op):
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
