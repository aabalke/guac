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
	blockIdx := (((pc - (w * 2)) &^ (w - 1)) & j.Config.PageMask) >> 1

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

	//fmt.Printf("Create: pageIdx %08X blockIdx %08X Size:%02d Pc:%08X W:%d\n", pageIdx, blockIdx, data.Size-1, data.RealPc, data.W)

	page.Blocks[blockIdx] = block
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
		j.EmitThumb(op)
	}

	// NOTE: condition branching instructions require ending the block
	// but need to use c.Reload to find out if branch was taken
	reloadState := j.ReloadState
	j.ReloadState = NONE
	switch reloadState {
	case NONE:
		j.EmitStepPc(data.W)
		j.EmitPipelineUpdate(data.Ptr, data.W)
		data.Ptr = unsafe.Add(data.Ptr, data.W)

		build = data.Size < j.Config.MaxInstCnt

	case RELOAD:
		j.Mov(JIT, gojit.Rax)
		j.CallFunc((*Jit).ReloadPipe)

	case POSSIBLE:

		j.Movb(j.C.Reload, gojit.Al)
		j.Testb(gojit.Al, gojit.Al)

		reload := j.JccForward(gojit.CC_NZ)

		j.EmitStepPc(data.W)
		j.EmitPipelineUpdate(data.Ptr, data.W)
		data.Ptr = unsafe.Add(data.Ptr, data.W)

		notReload := j.JmpForward()
		reload()

		j.Mov(JIT, gojit.Rax)
		j.CallFunc((*Jit).ReloadPipe)

		notReload()
	}

	irq()

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
	for decoder, desc := range formatters {

		if (op & desc.Mask) != desc.Match {
			continue
		}

		switch Decoder(decoder) {
		case INST_SVC:
			j.EmitSWI(op)
		case INST_BRA:
			j.EmitBranch(op)
		case INST_BRX:
			j.EmitBranchExchange(op)
		case INST_SDT_IMM, INST_SDT_REG:
			j.EmitSdt(op)
		case INST_BLK:
			j.EmitBlock(op)
		case INST_HAL:
			j.EmitHalf(op)
		case INST_UND:
			j.EmitUndefined(op)
		case INST_MRS:
			j.EmitMrs(op)
		case INST_MSR:
			j.EmitMsr(op)
		case INST_SWP:
			j.EmitSwp(op)
		case INST_MUL:
			j.EmitMul(op)
		case INST_ALU, INST_ALU_SHIFT_IMM, INST_ALU_SHIFT_REG:
			j.EmitAlu(op)
		}
		return
	}

	panic(fmt.Sprintf("arm7: decode failed op %08X\n", op))
}

func (j *Jit) EmitThumb(op32 uint32) {
	op := uint16(op32)

	for decoder, desc := range thumbFormatters {

		if (op32 & desc.Mask) != desc.Match {
			continue
		}

		switch Decoder(decoder) {
		case INST_THUMB_SVC:
			j.EmitThumbSWI(op)
		case INST_THUMB_ADD:
			j.EmitThumbAddSub(op)
		case INST_THUMB_SHI:
			j.EmitThumbShifted(op)
		case INST_THUMB_IMM:
			j.EmitThumbImm(op)
		case INST_THUMB_ALU:
			j.EmitThumbAlu(op)
		case INST_THUMB_HIBX:
			j.EmitThumbHiBx(op)
		case INST_THUMB_HI:
			j.EmitThumbHi(op)
		case INST_THUMB_SDTH:
			j.EmitThumbSdth(op)
		case INST_THUMB_SDTI:
			j.EmitThumbSdti(op)
		case INST_THUMB_SDTR:
			j.EmitThumbSdtr(op)
		case INST_THUMB_POPPC:
			j.EmitThumbPopPc(op)
		case INST_THUMB_PUS:
			j.EmitThumbPushPop(op)
		case INST_THUMB_REL:
			j.EmitThumbRelative(op)
		case INST_THUMB_BRA:
			j.EmitThumbBranch(op)
		case INST_THUMB_BL_SHORT:
			j.EmitShortLongBranch(op)
		case INST_THUMB_BL_LONG:
			j.EmitLongBranch(op)
		case INST_THUMB_BLK:
			j.EmitThumbBlock(op)
		case INST_THUMB_PCR:
			j.EmitThumbPCR(op)
		case INST_THUMB_CBR:
			j.EmitJumpCall(op)
		case INST_THUMB_SP:
			j.EmitThumbSP(op)
		case INST_THUMB_SPR:
			j.EmitThumbSPR(op)
		case INST_THUMB_LDSH:
			j.EmitThumbLDSH(op)
		}
		return
	}

	panic(fmt.Sprintf("arm7: decode failed op %04X\n", op))
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
