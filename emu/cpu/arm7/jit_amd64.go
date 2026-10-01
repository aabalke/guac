package arm7

import (
	"fmt"
	"unsafe"

	"github.com/aabalke/gojit"
)

type BlockData struct {
	Size       int
	BranchSize int
	Ptr        unsafe.Pointer
	RealPc, Pc uint32
	W          uint32
}

func (bd *BlockData) Set(pc, w uint32) {
	bd.Size = 0
	bd.BranchSize = 0
	// offset for pipelining
	bd.RealPc = (pc - (w * 2)) &^ (w - 1)
	bd.Pc = pc
	bd.W = w
}

func (j *Jit) CreateBlock(pc, w uint32) {
	pageIdx := pc >> j.Config.PageShift
	blockIdx := (pc & j.Config.PageMask) >> 1

	page := j.Pages[pageIdx]
	if page == nil {

		page = &Page{
			Id:     pageIdx,
			Blocks: make([]*JitBlock, (1<<j.Config.PageShift)>>1),
		}

		j.Pages[pageIdx] = page
	} else if page.Dead {
		println("page dead, block not created")
		return
	} else if block := page.Blocks[blockIdx]; block != nil && block.Skip {
		return
	}

	block := j.BlockCache.AssignBlock(&j.Pages, j.Config.PageShift, j.Config.PageMask)
	if block == nil {
		return
	}

	j.Assembler = block.Assembler

	j.MovAbs(uint64(uintptr(unsafe.Pointer(j))), JIT)
	j.MovAbs(uint64(j.C.Cpu), CPU)

	data := &j.BlockData
	data.Set(pc, w)

	data.Ptr = j.Mem.ReadPtr(data.RealPc)
	if data.Ptr == nil {
		j.BlockCache.PushTail(block)
		page.Blocks[blockIdx] = j.BlockCache.SkipBlock
		return
	}

	for build := true; build; {
		build = j.TryEmitOp(data)
		data.Size++
		data.BranchSize++
	}

	if data.Size < j.Config.MinInstCnt {
		//fmt.Printf("Rejecting Block Size: %d < %d\n", size, j.Config.MinInstCnt)
		j.BlockCache.PushTail(block)
		page.Blocks[blockIdx] = j.BlockCache.SkipBlock
		return
	}

	j.Exit()

	if err := j.Error(); err != nil {
		panic(err)
	}

	block.InitPc = pc
	block.Size = data.Size
	block.F = func() {
		gojit.CallJit(uintptr(unsafe.Pointer(&block.Assembler.Buf[0])))
	}

	page.Blocks[blockIdx] = block

	//if w == 2 {
	//	fmt.Printf("Block Created for Page %08X PC %08X END PC %08X SIZE %08d\n", pageIdx, pc, data.Pc, data.Size)
	//} else {
	//	fmt.Printf("Block Created for Page %08X PC %08X END PC %08X SIZE %08d\n", pageIdx, pc, data.Pc, data.Size)
	//}
}

func (j *Jit) TryEmitOp(data *BlockData) (build bool) {
	op := *(*uint32)(data.Ptr)

	irq := j.EmitStep()

	if data.W == 4 {

		conds := j.EmitCond(op >> 28)

		j.EmitArm(op)
		done := j.JmpForward()

		for _, cond := range conds {
			cond()
		}

		j.Movl(gojit.Imm(SEQ), j.C.Seq)

		done()

	} else {
		j.EmitThumb(uint16(op))
	}

	// NOTE: condition branching instructions require ending the block
	// but need to use c.Reload to find out if branch was taken
	reloadState := j.ReloadState
	j.ReloadState = NONE
	switch reloadState {
	case NONE:
		j.EmitStepPc(data.W)
		j.EmitPipelineUpdate(data.Ptr, data.W)

		build = data.Size < j.Config.MaxInstCnt

		data.Ptr = unsafe.Add(data.Ptr, data.W)

	case RELOAD:
		j.Mov(JIT, gojit.Rax)
		j.CallFunc((*Jit).ReloadPipe)

	case POSSIBLE:

		j.Movb(j.C.Reload, gojit.Al)
		j.Testb(gojit.Al, gojit.Al)

		reload := j.JccForward(gojit.CC_NZ)

		j.EmitStepPc(data.W)
		j.EmitPipelineUpdate(data.Ptr, data.W)

		notReload := j.JmpForward()
		reload()

		j.Mov(JIT, gojit.Rax)
		j.CallFunc((*Jit).ReloadPipe)

		notReload()
	}

	irq()

	//if w == 2 {
	//	fmt.Printf("emitOp PC %08X OP %04X END BLOCK %t\n", pc+(size*w), uint16(op), endBlock)
	//} else {
	//	fmt.Printf("emitOp PC %08X OP %08X END BLOCK %t\n", pc+(size*w), op, endBlock)
	//}

	return build
}

func (j *Jit) EmitCond(cond uint32) []func() {
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

func (j *Jit) EmitArm(op uint32) {
	switch {
	case (op>>24)&0xF == 0xF:
		j.EmitSWI(op)
	case IsBranch(op):
		j.EmitBranch(op)
	case IsBranchExchange(op):
		j.EmitBranchExchange(op)
	case IsSdt(op):
		j.EmitSdt(op)
	case IsBlock(op):
		j.EmitBlock(op)
	case IsHalf(op):
		j.EmitHalf(op)
	case IsUndefined(op):
		j.EmitUndefined(op)
	case IsMsr(op):
		j.EmitMsr(op)
	case IsMrs(op):
		j.EmitMrs(op)
	case IsSwp(op):
		j.EmitSwp(op)
	case IsMul(op):
		j.EmitMul(op)
	case IsAlu(op):
		j.EmitAlu(op)
	default:
		panic(fmt.Sprintf("unemittable amd64 jit instruction ARM OP %08X", op))
	}
}

func (j *Jit) EmitThumb(op uint16) {
	switch {
	case IsThumbSWI(op):
		j.EmitThumbSWI(op)
	case IsThumbAddSub(op):
		j.EmitThumbAddSub(op)
	case IsThumbShift(op):
		j.EmitThumbShifted(op)
	case IsThumbImm(op):
		j.EmitThumbImm(op)
	case IsThumbAlu(op):
		j.EmitThumbAlu(op)
	case IsThumbHiBx(op):
		j.EmitThumbHiBx(op)
	case IsThumbHi(op):
		j.EmitThumbHi(op)
	case IsLSHalf(op):
		j.EmitThumbLSHalf(op)
	case IsThumbLDSH(op):
		j.EmitThumbLDSH(op)
	case IsThumbSdt(op):
		j.EmitThumbSdt(op)
	case IsLPC(op):
		j.EmitThumbLPC(op)
	case IsLSImm(op):
		j.EmitThumbLSImm(op)
	case IsPopPc(op):
		j.EmitThumbPopPc(op)
	case IsPushPop(op):
		j.EmitThumbPushPop(op)
	case IsRelative(op):
		j.EmitThumbRelative(op)
	case IsThumbBranch(op):
		j.EmitThumbBranch(op)
	case IsJumpCall(op):
		j.EmitJumpCall(op)
	case IsStack(op):
		j.EmitThumbStack(op)
	case IsLongBranch(op):
		j.EmitLongBranch(op)
	case IsShortLongBranch(op):
		j.EmitShortLongBranch(op)
	case IsLSSP(op):
		j.EmitThumbLSSP(op)
	case IsThumbBlock(op):
		j.EmitThumbBlock(op)
	default:
		panic(fmt.Sprintf("unemittable amd64 jit instruction THUMB OP %04X", op))
	}
}

func (j *Jit) EmitToggleThumb() {
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

func (j *Jit) EmitStep() func() {
	j.Movb(j.C.IrqLine, gojit.Al)

	j.Testb(gojit.Al, gojit.Al)

	irq := j.JccForward(gojit.CC_NZ)

	j.Mov(JIT, gojit.Rax)

	j.Movl(j.C.R[PC], gojit.Ebx)

	j.Movb(j.C.T, gojit.Cl)
	j.Test(gojit.Ecx, gojit.Ecx)
	j.Movl(gojit.Imm(4), gojit.Ecx)
	j.Movl(gojit.Imm(2), gojit.Edi)
	j.Cmovcc(gojit.CC_NZ, gojit.Edi, gojit.Ecx)

	j.Movl(j.C.Seq, gojit.Edi)

	j.Movl(gojit.Imm(SEQ), j.C.Seq)

	j.CallFunc((*Jit).InstCycles)

	return irq
}

func (j *Jit) EmitStepPc(w uint32) {
	j.Add(gojit.Imm(w), j.C.R[PC])

	j.Mov(j.C.PcPtr, gojit.Rax)

	j.Test(gojit.Rax, gojit.Rax)

	noPtr := j.JccForward(gojit.CC_Z)

	j.Add(gojit.Imm(w), j.C.PcPtr)

	noPtr()
}

func (j *Jit) EmitPipelineUpdate(p unsafe.Pointer, w uint32) {
	// NOTE: when exiting jit, need pipeline setup properly
	// ONLY when not reloading. This removes every inst pipeline adjustment
	mask := uint32(0xFFFF_FFFF >> ((w & 2) * 8))
	j.MovAbs(uint64(uintptr(p)), gojit.Rax)
	j.Add(gojit.Imm(w), gojit.Rax)
	ptr := gojit.Indirect{Base: gojit.Rax, Offset: 0, Bits: 32}

	for i := range 2 {
		j.Movl(ptr, gojit.Ebx)
		j.And(gojit.Imm(mask), gojit.Ebx)
		j.Movl(gojit.Ebx, j.C.Op[i])
		j.Add(gojit.Imm(w), gojit.Rax)
	}

	j.Mov(gojit.Rax, j.C.PcPtr)
}
