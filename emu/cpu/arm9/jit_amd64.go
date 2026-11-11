package arm9

import (
	"fmt"
	"unsafe"

	"github.com/aabalke/gojit"
	"github.com/aabalke/guac/emu/cpu/arm7"
)

func (j *Jit) CreateBlock(pc, w uint32) {
	pageIdx := pc >> j.Config.PageShift
	blockIdx := (pc & j.Config.PageMask) >> 1

	page := j.Pages[pageIdx]
	if page == nil {

		page = &arm7.Page{
			Id:     pageIdx,
			Blocks: make([]*arm7.JitBlock, (1<<j.Config.PageShift)>>1),
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

	j.MovAbs(uint64(uintptr(unsafe.Pointer(j))), arm7.JIT)
	j.MovAbs(uint64(j.C.Cpu), arm7.CPU)

	data := &j.BlockData
	data.Set(pc, w)

	c := j.Cpu.(*Cpu)
	bus := c.Mem
	if c.Itcm.Readable(c.Reg.R[PC], w) {
		bus = c.Itcm
	}

	data.Ptr = bus.ReadPtr(data.RealPc)
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

	page.Blocks[blockIdx] = block
}

func (j *Jit) TryEmitOp(data *arm7.BlockData) (build bool) {
	op := *(*uint32)(data.Ptr)

	irq := j.EmitStep()

	if data.W == 4 {

		conds := j.EmitCond(op >> 28)

		j.EmitArm(op)
		done := j.JmpForward()

		for _, cond := range conds {
			cond()
		}

		j.Movl(gojit.Imm(arm7.SEQ), j.C.Seq)

		done()

	} else {
		j.EmitThumb(op)
	}

	// NOTE: condition branching instructions require ending the block
	// but need to use c.Reload to find out if branch was taken
	reloadState := j.ReloadState
	j.ReloadState = arm7.NONE
	switch reloadState {
	case arm7.NONE:
		j.EmitStepPc(data.W)
		j.EmitPipelineUpdate(data.Ptr, data.W)

		build = data.Size < j.Config.MaxInstCnt

		data.Ptr = unsafe.Add(data.Ptr, data.W)

	case arm7.RELOAD:
		j.Mov(arm7.JIT, gojit.Rax)
		j.CallFunc((*Jit).ReloadPipe)

	case arm7.POSSIBLE:

		j.Movb(j.C.Reload, gojit.Al)
		j.Testb(gojit.Al, gojit.Al)

		reload := j.JccForward(gojit.CC_NZ)

		j.EmitStepPc(data.W)
		j.EmitPipelineUpdate(data.Ptr, data.W)

		notReload := j.JmpForward()
		reload()

		j.Mov(arm7.JIT, gojit.Rax)
		j.CallFunc((*Jit).ReloadPipe)

		notReload()
	}

	j.Mov(arm7.JIT, gojit.Rax)
	j.CallFunc((*Jit).StepCleanup)

	irq()

	return build
}

func (j *Jit) EmitArm(op uint32) {
	for decoder, desc := range formatters {

		if (op & desc.Mask) != desc.Match {
			continue
		}

		switch arm7.Decoder(decoder) {
		case INST_BLX:
			j.EmitBlx(op)
		case INST_PLD:
			panic("unsetup pld instruction")
		case INST_SVC:
			j.EmitSWI(op)
		case INST_BKP:
			j.EmitBKPT(op)
		case INST_CRT:
			j.EmitCoDataReg(op)
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
		case INST_CLZ:
			j.EmitClz(op)
		case INST_QAL:
			j.EmitQalu(op)
		case INST_HWM:
			j.EmitExtendedMul(op)
		case INST_ALU, INST_ALU_SHIFT_IMM, INST_ALU_SHIFT_REG:
			j.EmitAlu(op)
		}
		return
	}

	panic(fmt.Sprintf("arm9: decode failed op %08X\n", op))
}

func (j *Jit) EmitThumb(op32 uint32) {
	op := uint16(op32)

	for decoder, desc := range thumbFormatters {

		if (op32 & desc.Mask) != desc.Match {
			continue
		}

		switch arm7.Decoder(decoder) {
		case INST_THUMB_SVC:
			j.EmitThumbSWI(op)
		case INST_THUMB_BPT:
			j.EmitThumbBKPT(op)
		case INST_THUMB_SHORT_BLX:
			j.EmitThumbShortBlx(op)
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

	panic(fmt.Sprintf("arm9: decode failed op %04X\n", op))
}

func (j *Jit) EmitStep() func() {
	j.Movb(j.C.IrqLine, gojit.Al)

	j.Testb(gojit.Al, gojit.Al)

	irq := j.JccForward(gojit.CC_NZ)

	j.Mov(arm7.JIT, gojit.Rax)
	j.CallFunc((*Jit).StepKindaButNotReally)

	return irq
}

//go:nosplit
func (j *Jit) StepKindaButNotReally() {
	c := j.Cpu.(*Cpu)

	c.Seq = arm7.SEQ

	w := uint32(4)
	if c.Reg.CPSR.T {
		w = 2
	}

	if w == 4 || c.Reg.R[PC]&2 == 0 {
		c.Cycles(c.Reg.R[PC], w, 0, true)
	}
}

//go:nosplit
func (j *Jit) StepCleanup() {
	c := j.Cpu.(*Cpu)
	c.tick(max(c.InstCycles, c.DataCycles) + c.IdleCycles)
	c.InstCycles, c.DataCycles, c.IdleCycles = 0, 0, 0
}
