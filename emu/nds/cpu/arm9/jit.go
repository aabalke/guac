package arm9

import (
	"fmt"
	"runtime/debug"
	"unsafe"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/jit"
	"github.com/aabalke/guac/emu/jit/amd64"
)

var _ = fmt.Sprintf

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

    Pages [0x1_0000_0000 >> PAGE_SHIFT]*Page // need 0xFFFF_FFFF for bios
    Head *Page
    Tail *Page
    Cnt uint64
    Capacity uint64

    Metrics [0x1_0000_0000 >> PAGE_SHIFT][]uint32

    invalidPages []uint32

    afterCall   bool
    inCallBlock bool

    // used for testing individual instructions
    testFunc func()

    callFrameSize int32
    frameSize     int32

    blockCh chan uint32
}

type Page struct {

    id uint32
	Prev *Page
	Next *Page
    Blocks []*JitBlock
    Written bool
}

type JitBlock struct {
	f      func()
	initPc uint32
    finalPc uint32
    finalOp uint32
	Length uint32
    assembler *amd64.Assembler
}

func NewJit(cpu *Cpu) *Jit {

    j := &Jit{
		Head:     &Page{},
		Tail:     &Page{},
        Capacity: 5, //0x1_0000_0000 >> PAGE_SHIFT,
        Cpu: cpu,
        blockCh: make(chan uint32),
    }

	j.Head.Next = j.Tail
	j.Tail.Prev = j.Head

    go j.bgBlockComp()

    return j
}

func (j *Jit) bgBlockComp() {
    for pc := range j.blockCh {
        j.CreateBlock(pc)
    }
}

func (j *Jit) UserBankReg(isLr bool) amd64.Indirect {

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

func (j *Jit) REG(i uint32) amd64.Indirect {
    return amd64.Indirect{
        Base: CPU,
        Offset: R + int32(i * 4),
        Bits: 32,
    }
}

func (j *Jit) InvalidatePage(addr uint32) {

    if j.Pages[addr >> PAGE_SHIFT] == nil {
        return
    }

    j.invalidPages = append(j.invalidPages, addr >> PAGE_SHIFT)
}

//func (j *Jit) DeletePage(id uint32) {
//
//    return
//
//    page := j.Pages[id]
//    if page == nil {
//        return
//    }
//
//    //fmt.Printf("PAGE DELETED %08X\n", id)
//    //j.GetPageData(id)
//
//    for i := range len(page.Blocks) {
//
//        //fmt.Printf("--- DEL BLOCK %02d\n", i)
//
//        //fmt.Printf("RELEASE CALLED\n")
//
//        if page.Blocks[i] == nil {
//            //fmt.Printf("------ BLOCK IS NIL, SKIP DEL %02d\n", i)
//            continue
//        }
//
//        if j.Pages[id].Blocks[i].assembler == nil {
//            continue
//        }
//
//        j.Pages[id].Blocks[i].assembler.Release()
//    }
//
//    j.Metrics[id] = nil
//    j.Pages[id] = nil
//}
//
//func (j *Jit) DeletePages() {
//    // this clears invalid pages after resources not being used by cpu implimentations (if cpu writes to its own blocks we get big errors
//
//    return
//
//    j.Cnt = uint64(max(0, int(j.Cnt) - len(j.invalidPages)))
//
//    for _, v := range j.invalidPages {
//
//        if j.Pages[v] == nil {
//            continue
//        }
//
//        for i := range j.Pages[v].Blocks {
//
//            if j.Pages[v].Blocks[i] == nil {
//                continue
//            }
//
//            if j.Pages[v].Blocks[i].assembler == nil {
//                continue
//            }
//
//            j.Pages[v].Blocks[i].assembler.Release()
//        }
//
//        j.remove(j.Pages[v])
//        j.Pages[v] = nil
//    }
//
//    j.invalidPages = []uint32{}
//}

func (j *Jit) CreateBlock(pc uint32) {

    reqInst := config.Conf.Nds.NdsJit.BatchInst

    old := debug.SetGCPercent(-1)

	pageIdx := pc >> PAGE_SHIFT

    page := j.Pages[pageIdx]
    if page == nil {

		page = &Page{
			id:     pageIdx,
			Blocks: make([]*JitBlock, (1<<PAGE_SHIFT)>>2),
		}

		//j.Metrics[pageIdx] = make([]uint32, (1<<PAGE_SHIFT)>>2)

		//fmt.Printf("PAGE %08X CNT %02d\n", pageIdx, j.Cnt)
        //fmt.Printf("ADDING %08X\n", page.id)
		j.add(page)
        j.Pages[pageIdx] = page

        //fmt.Printf("Creating Page\n")
        //j.GetPagesData()



		//j.Cnt++

		//if j.Cnt > j.Capacity {
		//	tail := j.popTail()
		//	//fmt.Printf("TAIL %08X\n", tail.id)
        //    j.invalidPages = append(j.invalidPages, tail.id)
        //    //j.DeletePage(tail.id)
		//	//j.remove(tail)
		//	//j.Cnt--
        //    //panic("COMPLETED FIRST DELETE")
		//}
    }

    blockIdx := (pc & PAGE_MASK) >> 2

    page.Blocks[blockIdx] = &JitBlock{
        initPc: pc,
        Length: reqInst,
    }

    //fmt.Printf("CREATING BLOCK PAGE %08X\n", pc >> PAGE_SHIFT)

    asm, err := amd64.New(gojit.PageSize)
    if err != nil {
        panic(err)
    }

    j.Assembler = asm

    block := page.Blocks[(pc & PAGE_MASK) >> 2]
    block.assembler = asm

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

    j.BuildTo(&block.f)

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

        if ok := j.emitOp(op, lastPc +i * 4); !ok {

            pageIdx := initPc >> PAGE_SHIFT
            blockIdx := (initPc & PAGE_MASK) >> 2

            block := j.Pages[pageIdx].Blocks[blockIdx]
            block.Length = i + length
            block.finalOp = op
            block.finalPc = lastPc + i * 4
            return true, false, i, lastPc + i * 4
        }
    }

    //op := *(*uint32)(unsafe.Add(p, reqInst * 4))
    //block := &j.Pages[initPc >> 16].Blocks[(initPc & PAGE_MASK) / 4]

    //(*block).Length = reqInst + length
    //(*block).finalOp = op
    //(*block).finalPc = lastPc + reqInst * 4
    //return true, false, reqInst, lastPc + reqInst * 4

    // if looped through all req Inst just keep going
    return false, false, reqInst, lastPc + reqInst * 4
}

func (j *Jit) emitBranch(op, lastPc uint32) (ok bool, newPc uint32) {
    // only emit branch if always exectues
	if cond := op >> 28; cond < 0xE {
        return false, 0
    }

    if !isB(op) {
        return false, 0
    }

    if immLoop := op == 0xEAFFFFFE; immLoop {
        j.Cpu.Halted = true
        return true, lastPc
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

func (j *Jit) emitOp(op uint32, pc uint32) bool {

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

    ok := j.DecodeARM(op, pc)

    j.Add(amd64.Imm(4), j.REG(PC))

	// Complete JCC instruction used for cond (if any)
	for _, tgt := range jcctargets {
		tgt()
	}

    return ok
}

func (jit *Jit) DecodeARM(opcode uint32, pc uint32) bool {

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

        rd   := (opcode >> 12) & 0xF
        rn   := (opcode >> 16) & 0xF
        pre  := (opcode >> 24) & 1 != 0
        load := (opcode >> 20) & 1 != 0

        wb   := (opcode >> 21) & 1 != 0
        wb    = (wb || !pre) && !(load && rn == rd)

        if rd == PC && (load || wb) {
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

        if rd := (opcode >> 12) & 0xF; rd == PC {
            return false
        }

        // I think this is fixed but be on the look out
        //if rn == PC {
        //    println("HERE")
        //    //return false
        //}

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

const threshold = 255

func (j *Jit) UpdateMetrics(pc uint32) {

    //j.set(pc)

    pageIdx := pc >> PAGE_SHIFT
    blockIdx := (pc & PAGE_MASK) >> 2 // aligned to arm

    if page := j.Pages[pageIdx]; page != nil {
		j.moveToHead(page)
    }

    if metrics := j.Metrics[pageIdx]; metrics == nil {
        j.Metrics[pageIdx] = make([]uint32, (1 << PAGE_SHIFT) >> 2)
    }

	j.Metrics[pageIdx][blockIdx]++

    if j.Metrics[pageIdx][blockIdx] > threshold {

        select {
        case j.blockCh <-pc:
        default:
            j.Metrics[pageIdx][blockIdx]--
        }
    }
}
