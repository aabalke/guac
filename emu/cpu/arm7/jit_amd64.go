package arm7

import (
	"fmt"
	"unsafe"

	"github.com/aabalke/gojit"
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
	tempPc := pc - (w * 2)

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
		op = *(*uint32)(unsafe.Add(p, i*w))

		if length >= BATCH_INST_MAX {
			break
		}

		if ok := j.TryEmitOp(op, w); !ok {
			break
		}

		fmt.Printf("emitOp PC %08X OP %08X\n", tempPc, op)

		i++
		length++
		tempPc += w
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

	fmt.Printf("Block Created for PC %08X EXIT PC %08X OP %08X\n", pc, tempPc, op)
}

func (j *Jit) TryEmitOp(op, w uint32) bool {
	if w == 4 {

		ok := j.IsJittableArm(op)
		if ok {

			condTargets := j.emitCond(op)

			j.Mov(JIT, gojit.Rax)
			j.CallFunc((*Jit).Step)

			j.emitArm(op)

			for _, target := range condTargets {
				target()
			}

			j.Mov(JIT, gojit.Rax)
			j.CallFunc((*Jit).UpdatePcArm)
		}

		return ok
	} else {

		ok := j.IsJittableThumb(uint16(op))
		if ok {

			j.Mov(JIT, gojit.Rax)
			j.CallFunc((*Jit).Step)

			j.emitThumb(uint16(op))

			j.Mov(JIT, gojit.Rax)
			j.CallFunc((*Jit).UpdatePcThumb)
		}

		return ok
	}
}

func (j *Jit) UpdatePcArm() {
	j.cpu.Reg.R[PC] += 4
	if j.cpu.PcPtr != nil {
		j.cpu.PcPtr = unsafe.Add(j.cpu.PcPtr, 4)
	}
}

func (j *Jit) UpdatePcThumb() {
	j.cpu.Reg.R[PC] += 2
	if j.cpu.PcPtr != nil {
		j.cpu.PcPtr = unsafe.Add(j.cpu.PcPtr, 2)
	}
}

func (j *Jit) emitCond(op uint32) []func() {
	var jcctargets []func()

	switch cond := op >> 28; cond {
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

func (j *Jit) IsJittableArm(op uint32) bool {
	switch {
	case (op>>24)&0xF == 0xF, IsBranch(op), IsBranchExchange(op):
		return false
	case IsSdt(op), IsHalf(op):
		load := (op>>20)&1 != 0
		rdpc := op&0xF000 == 0xF000
		return !load || !rdpc
	case IsBlock(op):
		load := (op>>20)&1 != 0
		pcIncluded := op&0x8000 != 0
		rlist := op & 0xFFFF
		return rlist != 0 && (!load || !pcIncluded)
	case IsUndefined(op), IsMsr(op):
		return false
	case IsMrs(op), IsSwp(op), IsMul(op):
		return true
	case IsAlu(op):
		rdpc := op&0xF000 == 0xF000
		swiExit := op&0x3F0_000F == 0x3F0_000F
		return !rdpc && !swiExit
	default:
		return false
	}
}

func (j *Jit) emitArm(op uint32) {
	switch {
	case IsSdt(op):
		j.emitSdt(op)
	case IsBlock(op):
		j.emitBlock(op)
	case IsHalf(op):
		j.emitHalf(op)
	case IsMrs(op):
		j.emitMrs(op)
	case IsSwp(op):
		j.emitSwp(op)
	case IsMul(op):
		j.emitMul(op)
	case IsAlu(op):
		j.emitAlu(op)
	}
}

func (j *Jit) IsJittableThumb(op uint16) bool {
	switch {
	case IsthumbSWI(op):
		return false
	case IsThumbAddSub(op), IsThumbShift(op), IsThumbImm(op), IsThumbAlu(op):
		return true
	case IsThumbHiReg(op):

		var (
			inst = (op >> 8) & 0b11
			mSBd = (op>>7)&1 != 0
			rd   = op & 0x7
		)

		if inst != 3 && mSBd {
			rd |= 0b1000
		}

		if inst == 3 || rd == PC {
			return false
		}

		return true

	case IsLSHalf(op), IsThumbSdt(op), IsLPC(op), IsLSImm(op):
		return true
	case IsPushPop(op):
		pclr := (op>>8)&1 != 0
		pop := (op>>11)&1 != 0
		if pop && pclr {
			return false
		}
		return true
	case IsRelative(op):
		return true
	case IsThumbB(op), IsJumpCall(op):
		return false
	case IsStack(op):
		return true
	case IsLongBranch(op), IsShortLongBranch(op):
		return false
	case IsLSSP(op):
		return true
	case IsMulti(op):
		ldmia := (op>>11)&1 != 0
		rlist := op & 0xFF

		if ldmia && rlist == 0 {
			return false
		}
		return true
	}

	return false
}

func (j *Jit) emitThumb(op uint16) {
	switch {
	case IsThumbAddSub(op):
		j.emitThumbAddSub(op)
	case IsThumbShift(op):
		j.emitThumbShifted(op)
	case IsThumbImm(op):
		j.emitThumbImm(op)
	case IsThumbAlu(op):
		j.emitThumbAlu(op)
	case IsThumbHiReg(op):
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
	case IsStack(op):
		j.emitThumbStack(op)
	case IsLSSP(op):
		j.emitThumbLSSP(op)
	case IsMulti(op):
		j.emitThumbBlock(op)
	}
}
