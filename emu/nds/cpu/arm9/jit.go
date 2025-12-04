package arm9

import (
	"runtime/debug"
	"unsafe"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/nds/cpu/arm9/jit"
	"github.com/aabalke/guac/emu/nds/cpu/arm9/jit/amd64"
)

const (
    PAGE_MASK = 0xFFFF
    PAGE_SHIFT = 16
)

var (
    CpuPointer *Cpu
    CPU    = amd64.R11
    REG    = int32(unsafe.Offsetof(Cpu{}.Reg))
    R      = REG + int32(unsafe.Offsetof(Reg{}.R))
    CPSR   = REG + int32(unsafe.Offsetof(Reg{}.CPSR))
    MODE = amd64.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.Mode)), Bits: 32}
    N = amd64.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.N)), Bits: 8}
    Z = amd64.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.Z)), Bits: 8}
    C = amd64.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.C)), Bits: 8}
    V = amd64.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.V)), Bits: 8}
    Q = amd64.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.Q)), Bits: 8}
    I = amd64.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.I)), Bits: 8}
    F = amd64.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.F)), Bits: 8}
    T = amd64.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.T)), Bits: 8}
)

type Jit struct {
    *amd64.Assembler
	Cpu    *Cpu
	//Blocks map[uint32]*JitBlock
	//Blocks [0x825]*JitBlock

    Pages [0x1_0000_0000 >> PAGE_SHIFT]*Page // need 0xFFFF_FFFF for bios

    afterCall bool
    inCallBlock bool

    // used for testing individual instructions
    testFunc func()

    callFrameSize int32
    frameSize int32

    blockCh chan uint32
}

type Page struct {
    Metrics []uint32
    Blocks []*JitBlock
}

type JitBlock struct {
	f      func()
	initPc uint32
    finalPc uint32
    finalOp uint32
	Length uint32
}

func NewJit(cpu *Cpu) *Jit {

    j := &Jit{
        Cpu: cpu,
        blockCh: make(chan uint32),
        //Blocks: make(map[uint32]*JitBlock),
    }

    go j.bgBlockComp()

    return j
}

func (j *Jit) bgBlockComp() {
    for pc := range j.blockCh {
        j.CreateBlock(pc)
    }
}

func (j *Jit) UserBankReg(isLr bool) amd64.Operand {

    // user bank id is 0, so return first uint32 in SP and LR [6]uint32s

    if isLr {
        return amd64.Indirect{
            Base: CPU,
            Offset: REG + int32(unsafe.Offsetof(Reg{}.LR)),
            Bits: 32,
        }
    }

    return amd64.Indirect{
        Base: CPU,
        Offset: REG + int32(unsafe.Offsetof(Reg{}.SP)),
        Bits: 32,
    }
}

func (j *Jit) REG(i uint32) amd64.Operand {
    return amd64.Indirect{
        Base: CPU,
        Offset: R + int32(i * 4),
        Bits: 32,
    }
}

func (j *Jit) StackOffset(off int32, bits byte) amd64.Operand {
    return amd64.Indirect{Base: amd64.Rsp, Offset: off, Bits: bits}
}

func (j *Jit) InvalidatePage(addr uint32) {
    j.Pages[addr >> PAGE_SHIFT] = nil
}

func (j *Jit) CreateBlock(pc uint32) {

    reqInst := config.Conf.Nds.NdsJit.BatchInst

    old := debug.SetGCPercent(-1)

    page := j.Pages[pc >> PAGE_SHIFT]
    if page == nil {
        panic("CREATING BLOCK ON NIL PAGE")
    }

    page.Blocks[(pc & PAGE_MASK) >> 2] = &JitBlock{
        initPc: pc,
        Length: reqInst,
    }

    asm, err := amd64.New(gojit.PageSize)
    if err != nil {
        panic(err)
    }

    j.Assembler = asm

    // setup, not sure if needed per Block
    //asm.MovAbs(uint64(uintptr(unsafe.Pointer(j.Cpu))), amd64.R15)

    j.frameSize = 20 * 4
	j.Sub(amd64.Imm(int32(j.frameSize + 8)), amd64.Rsp)
	j.Mov(amd64.Rbp, amd64.Indirect{
        Base: amd64.Rsp, 
        Offset: int32(j.frameSize),
        Bits: 64},
    )
	j.Lea(amd64.Indirect{
        Base: amd64.Rsp, 
        Offset: int32(j.frameSize),
        Bits: 64},
        amd64.Rbp,
    )

    CpuPointer = j.Cpu
    j.MovAbs(uint64(uintptr(unsafe.Pointer(CpuPointer))), CPU)

    initPc, lastPc, length := pc, pc, uint32(0)
    for reqInst := uint32(32); reqInst > 0; {

        done, isBranch, resInst, finalPc := j.emitContBlock(initPc, lastPc, reqInst, length)
        length += resInst
        if done {
            break
        }

        if isBranch {
            reqInst -= resInst
            lastPc = finalPc
        }
    }

	j.Mov(amd64.Indirect{
        Base: amd64.Rsp,
        Offset: int32(j.frameSize),
        Bits: 64}, amd64.Rbp)
	j.Add(amd64.Imm(int32(j.frameSize + 8)), amd64.Rsp)
    asm.Ret()

    j.BuildTo(&j.Pages[pc >> PAGE_SHIFT].Blocks[(pc & PAGE_MASK) >> 2].f)

    debug.SetGCPercent(old)
}

func (j *Jit) emitContBlock(initPc, lastPc, reqInst, length uint32) (done, isBranch bool, resInst, finalPc uint32) {

    p, ok := j.Cpu.mem.ReadPtr(lastPc, true)
    if !ok {
        return true, false, 0, 0
    }

    for i := range uint32(reqInst) {

        op := *(*uint32)(unsafe.Add(p, i * 4))

        if ok, newPc := j.emitBranch(op, lastPc + i * 4); ok {
            return false, true, i, newPc
        }

        if ok := j.emitOp(op); !ok {
            block := &j.Pages[initPc >> 16].Blocks[(initPc & PAGE_MASK) / 4]
            (*block).Length = i + length
            (*block).finalOp = op
            (*block).finalPc = lastPc + i * 4
            return true, false, i, 0
        }
    }

    return true, false, 0, 0
}

func (j *Jit) emitBranch(op, lastPc uint32) (ok bool, newPc uint32) {
    // only emit branch if always exectues
	if cond := op >> 28; cond < 0xE {
        return false, 0
    }

    if !isB(op) {
        return false, 0
    }

    if isLink := (op >> 24) & 1 != 0; isLink {
        j.Movl(amd64.Imm(lastPc + 4), j.REG(14))
    }

    //j.Movl(j.REG(15), amd64.Eax)
    //j.Add(amd64.Imm(uint32((int32(op)<<8)>>6) + 8), amd64.Eax)
    //j.Movl(amd64.Eax, j.REG(15))

	newPc = lastPc + uint32((int32(op)<<8)>>6) + 8
    j.Movl(amd64.Imm(newPc), j.REG(15))

    return true, newPc

}

func (j *Jit) emitOp(op uint32) bool {

	cond := op >> 28
	var jcctargets []func()

	switch cond {
	case 0xE, 0xF:
		// nothing to do, always executed
	case 0x0: // Z
		j.Bt(amd64.Imm(0), Z)
		jcctargets = append(jcctargets, j.JccForward(amd64.CC_NC))
	case 0x1: // !Z
		j.Bt(amd64.Imm(0), Z)
		jcctargets = append(jcctargets, j.JccForward(amd64.CC_C))
	case 0x2: // C
		j.Bt(amd64.Imm(0), C)
		jcctargets = append(jcctargets, j.JccForward(amd64.CC_NC))
	case 0x3: // !C
		j.Bt(amd64.Imm(0), C)
		jcctargets = append(jcctargets, j.JccForward(amd64.CC_C))
	case 0x4: // N
		j.Bt(amd64.Imm(0), N)
		jcctargets = append(jcctargets, j.JccForward(amd64.CC_NC))
	case 0x5: // !N
		j.Bt(amd64.Imm(0), N)
		jcctargets = append(jcctargets, j.JccForward(amd64.CC_C))
	case 0x6: // V
		j.Bt(amd64.Imm(0), V)
		jcctargets = append(jcctargets, j.JccForward(amd64.CC_NC))
	case 0x7: // !V
		j.Bt(amd64.Imm(0), V)
		jcctargets = append(jcctargets, j.JccForward(amd64.CC_C))
	case 0x8: // C && !Z
		j.Bt(amd64.Imm(0), C)
		jcctargets = append(jcctargets, j.JccForward(amd64.CC_NC))
		j.Bt(amd64.Imm(0), Z)
		jcctargets = append(jcctargets, j.JccForward(amd64.CC_C))
	case 0x9: // !C || Z
		j.Movb(C, amd64.Al)
		j.Xorb(amd64.Imm(1), amd64.Al)
		j.Orb(Z, amd64.Al)
		jcctargets = append(jcctargets, j.JccForward(amd64.CC_Z))
	case 0xC: // !Z && N==V
		j.Bt(amd64.Imm(0), Z)
		jcctargets = append(jcctargets, j.JccForward(amd64.CC_C))
		fallthrough
	case 0xA, 0xB: // N==V / N!=V
		j.Movb(N, amd64.Al)
		j.Xorb(V, amd64.Al)
		if cond == 0xA || cond == 0xC {
			jcctargets = append(jcctargets, j.JccForward(amd64.CC_NZ))
		} else {
			jcctargets = append(jcctargets, j.JccForward(amd64.CC_Z))
		}
	case 0xD: // Z || N==V / N!=V
		j.Movb(N, amd64.Al)
		j.Xorb(V, amd64.Al)
		j.Orb(Z, amd64.Al)
		jcctargets = append(jcctargets, j.JccForward(amd64.CC_Z))
	default:
		panic("unreachable")
	}

    ok := j.DecodeARM(op)

	// Complete JCC instruction used for cond (if any)
	for _, tgt := range jcctargets {
		tgt()
	}

    return ok

    //return ok
}

func (jit *Jit) DecodeARM(opcode uint32) bool {

    switch {
    case isBLX(opcode):
        return false
    case isPLD(opcode):
        return false
    }

	if swi := (opcode>>24)&0xF == 0xF; swi {
        return false
	}

    switch {
    case isBkpt(opcode):
	case isB(opcode):
	case isBX(opcode):
	case isSDT(opcode):

        load := (opcode >> 20) & 1 != 0
        if rd := (opcode >> 12) & 0xF == 0xF; rd && load {
            return false
        }

        if rn := (opcode >> 16) & 0xF == 0xF; rn {
            return false
        }

        jit.emitSdt(opcode)
        return true
	case isBlock(opcode):

        if pcIncluded := opcode & 0x8000 != 0; pcIncluded {
            return false
        } 

        jit.emitBlock(opcode)
        return true

	case isHalf(opcode):

        if rd := (opcode >> 12) & 0xF == 0xF; rd {
            return false
        }

        jit.emitHalf(opcode)
        return true
	case isUD(opcode):
	case isPSR(opcode):
	case isSWP(opcode):
        jit.emitSwp(opcode)
        return true
    case isM(opcode):
        jit.emitMul(opcode)
        return true
    case isCLZ(opcode):
        jit.emitClz(opcode)
        return true
    case isQAlu(opcode):
        jit.emitQalu(opcode)
        return true
	case isALU(opcode):

        // required to skip for mario kart
        //return false

        if inst := (opcode >> 21) & 0xF; inst == 0x2 || inst == 0x4 || inst == 0xA || inst == 0xD {
            return false
        }
        
        //if mov := (opcode >> 21) & 0xF < 0xD; mov {
        if (opcode >> 25) & 1 == 0 {
            return false
        }

        if rd := (opcode >> 12) & 0xF == 0xF; rd {
            return false
        }

        jit.emitAlu(opcode)
        return true
    case isCoDataReg(opcode):
	}

    return false
}

func (j *Jit) TestInst(op uint32, f func(op uint32)) {

    old := debug.SetGCPercent(-1)

    asm, err := amd64.New(gojit.PageSize)
    if err != nil {
        panic(err)
    }

    j.Assembler = asm

    j.frameSize = 20 * 4
	j.Sub(amd64.Imm(int32(j.frameSize + 8)), amd64.Rsp)
	j.Mov(amd64.Rbp, amd64.Indirect{
        Base: amd64.Rsp, 
        Offset: int32(j.frameSize),
        Bits: 64},
    )

	j.Lea(amd64.Indirect{
        Base: amd64.Rsp, 
        Offset: int32(j.frameSize),
        Bits: 64},
        amd64.Rbp,
    )

    CpuPointer = j.Cpu
    j.MovAbs(uint64(uintptr(unsafe.Pointer(CpuPointer))), CPU)

    f(op)

	j.Mov(amd64.Indirect{
        Base: amd64.Rsp,
        Offset: int32(j.frameSize),
        Bits: 64}, amd64.Rbp)
	j.Add(amd64.Imm(int32(j.frameSize + 8)), amd64.Rsp)
    asm.Ret()

    asm.BuildTo(&j.testFunc)

    j.testFunc()
    debug.SetGCPercent(old)

    asm.Release()
}

func Read(addr uint32) uint32 {
    return CpuPointer.mem.Read8(addr, true)
}

func Read16(addr uint32) uint32 {
    return CpuPointer.mem.Read16(addr, true)
}

func Read32(addr uint32) uint32 {
    return CpuPointer.mem.Read32(addr, true)
}

func Write(addr uint32, v uint8) {
    CpuPointer.mem.Write8(addr, v, true)
}

func Write16(addr uint32, v uint16) {
    CpuPointer.mem.Write16(addr, v, true)
}

func Write32(addr, v uint32) {
    CpuPointer.mem.Write32(addr, v, true)
}

func (j *Jit) CallFunc(f any) {
    j.MovAbs(uint64(uintptr(unsafe.Pointer(CpuPointer))), CPU)
    j.InternalCallFunc(f)
    j.MovAbs(uint64(uintptr(unsafe.Pointer(CpuPointer))), CPU)
}

// SHL SAR vs movsx /

//const threshold = 255 / 2
const threshold = 255

func (j *Jit) UpdateMetrics(pc uint32) {

    if j.Pages[pc >> PAGE_SHIFT] == nil {
        j.Pages[pc >> PAGE_SHIFT] = &Page{
            Metrics: make([]uint32, (1 << PAGE_SHIFT) >> 2),
            Blocks:  make([]*JitBlock, (1 << PAGE_SHIFT) >> 2),
        }
    }

    j.Pages[pc >> PAGE_SHIFT].Metrics[(pc & PAGE_MASK) >> 2]++

    if j.Pages[pc >> PAGE_SHIFT].Metrics[(pc & PAGE_MASK) >> 2] > threshold {
        select {
        case j.blockCh <-pc:
        default:
            j.Pages[pc >> PAGE_SHIFT].Metrics[(pc & PAGE_MASK) >> 2]--
        }
    }
}
