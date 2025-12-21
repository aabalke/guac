package arm9

import (
	"fmt"
	d "runtime/debug"
	"unsafe"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/jit"
	"github.com/aabalke/guac/emu/jit/amd64"
)

var _ = fmt.Sprintf

const (
    PAGE_MASK = 0xFFFF
    PAGE_SHIFT = 16

    conc = !true
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

    if conc {
        go j.bgBlockComp()
    }

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

func (j *Jit) DeletePages() {
    // this clears invalid pages after resources not being used by cpu implimentations (if cpu writes to its own blocks we get big errors

    j.Cnt = uint64(max(0, int(j.Cnt) - len(j.invalidPages)))

    for _, v := range j.invalidPages {

        if j.Pages[v] == nil {
            continue
        }

        for i := range j.Pages[v].Blocks {

            if j.Pages[v].Blocks[i] == nil {
                continue
            }

            if j.Pages[v].Blocks[i].assembler == nil {
                continue
            }

            j.Pages[v].Blocks[i].assembler.Release()
        }

        j.remove(j.Pages[v])
        j.Pages[v] = nil
    }

    j.invalidPages = []uint32{}
}

func (j *Jit) CreateBlock(pc uint32) {

    //reqInst := config.Conf.Nds.NdsJit.BatchInst

    old := d.SetGCPercent(-1)

	pageIdx := pc >> PAGE_SHIFT
    blockIdx := (pc & PAGE_MASK) >> 2

    page := j.Pages[pageIdx]
    if page == nil {

		page = &Page{
			id:     pageIdx,
			Blocks: make([]*JitBlock, (1<<PAGE_SHIFT)>>2),
		}

		j.add(page)
        j.Pages[pageIdx] = page
    }

    asm, err := amd64.New(gojit.PageSize)
    if err != nil {
        panic(err)
    }

    j.Assembler = asm

    block := &JitBlock{
        initPc: pc,
        assembler: asm,
    }

    fs := 0

    j.Push(amd64.Rbp)
    j.Mov(amd64.Rsp, amd64.Rbp)
    j.Sub(amd64.Imm(fs + 8), amd64.Rsp)

    //j.frameSize = 20 * 8
	//j.Sub(amd64.Imm(int32(j.frameSize + 8)), amd64.Rsp)
	//j.Mov(amd64.Rbp, amd64.Indirect{
    //    Base: amd64.Rsp, 
    //    Offset: int32(j.frameSize),
    //    Bits: 64},
    //)
	//j.Lea(amd64.Indirect{
    //    Base: amd64.Rsp, 
    //    Offset: int32(j.frameSize),
    //    Bits: 64},
    //    amd64.Rbp,
    //)

    CpuPointer = j.Cpu
    j.MovAbs(uint64(uintptr(unsafe.Pointer(CpuPointer))), CPU)

    tempPc := pc
    var length, op, i, iA uint32

    p, ok := j.Cpu.mem.ReadPtr(tempPc, true)
    if !ok {
        panic("READ BAD")
    }

    for {

        op = *(*uint32)(unsafe.Add(p, i*4))

        //if ok, newPc := j.emitBranch(op, tempPc); ok {
        //    tempPc = newPc
        //    length += i
        //    i = 0

        //    p, ok = j.Cpu.mem.ReadPtr(tempPc, true)
        //    if !ok {
        //        panic("READ BAD")
        //    }
        //    continue
        //}

        if ok := j.emitOp(op, tempPc); !ok {
            length += i
            break
        }

        if iA >= config.Conf.Nds.NdsJit.BatchInst {
            length += i
            break
        }

        i++
        iA++
        tempPc += 4
    }

    block.Length = length
    block.finalOp = op
    block.finalPc = tempPc

    //if debug.B[6] {
    //    fmt.Printf(
    //        "LEN %08d INIT PC %08X FINAL OP %08X PC %08X\n",
    //        block.Length,
    //        block.initPc,
    //        block.finalOp, block.finalPc)

    //    debug.B[6] = false
    //}

	//j.Mov(amd64.Indirect{
    //    Base: amd64.Rsp,
    //    Offset: int32(j.frameSize),
    //    Bits: 64}, amd64.Rbp)
	//j.Add(amd64.Imm(int32(j.frameSize + 8)), amd64.Rsp)
    //asm.Ret()

    j.Add(amd64.Imm(fs + 8), amd64.Rsp)
    j.Pop(amd64.Rbp)
    j.Ret()

    for j.Off&15 != 0 {
        j.Int3()
    }

    j.BuildTo(&block.f)

    page.Blocks[blockIdx] = block
    d.SetGCPercent(old)
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

	// Complete JCC instruction used for cond (if any)
	for _, tgt := range jcctargets {
		tgt()
	}

    j.Add(amd64.Imm(4), j.REG(PC))

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

        if rd == PC {
            return false
        }

        jit.emitSdt(opcode)
        return true
	case isBlock(opcode):

        return false
        if pcIncluded := opcode & 0x8000 != 0; pcIncluded {
            return false
        } 

        jit.emitBlock(opcode)
        return true

	case isHalf(opcode):

        return false
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

        inst := (opcode >> 21) & 0xF
        set := (opcode >> 20) & 1 != 0
        imm := (opcode >> 25) & 1 != 0
        rm := (opcode >> 0) & 0xF

        // swi exit
        if inst == MOV && set && imm && rm == LR {
            return false
        }

        //shReg := (opcode >> 4) & 1 != 0
        //shType := (opcode >> 5) & 0b11
        ////rd := (opcode >> 12) & 0xF
        //shift := (opcode >> 7) & 0x1F
        //if inst == 0x4 && !set && !imm && !shReg && shType == 0 && shift != 0 && rm >= 0x8 {
        //if opcode & 0xFFFF_00FF == 0xE082_008E {
        //    fmt.Printf("PC %08X OP %08X\n", pc, opcode)
        //    return false
        //}

        jit.emitAlu(opcode)

        return true
    case isCoDataReg(opcode):
	}

    return false
}

func (j *Jit) TestInst(op uint32, f func(op uint32)) {

    old := d.SetGCPercent(-1)

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
    d.SetGCPercent(old)

    asm.Release()
}

//go:noinline
//go:nosplit
func Read(addr uint32) uint32 {
    return CpuPointer.mem.Read8(addr, true)
}

//go:noinline
//go:nosplit
func Read16(addr uint32) uint32 {
    return CpuPointer.mem.Read16(addr, true)
}

//go:noinline
//go:nosplit
func Read32(addr uint32) uint32 {
    return CpuPointer.mem.Read32(addr, true)
}

//go:noinline
//go:nosplit
func Write(addr uint32, v uint8) {
    CpuPointer.mem.Write8(addr, v, true)
}

//go:noinline
//go:nosplit
func Write16(addr uint32, v uint16) {
    CpuPointer.mem.Write16(addr, v, true)
}

//go:noinline
//go:nosplit
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
        if conc {
            select {
            case j.blockCh <-pc:
            default:
                j.Metrics[pageIdx][blockIdx]--
            }
        } else {
            j.CreateBlock(pc)
        }
    }
}

var (
    rpc uint32
    sav Reg
    sta Reg
)

func (j *Jit) StartTest(op uint32, compare bool, f func(op uint32)) {

    if !compare {
        return
    }

    cpu := j.Cpu
    rpc = cpu.Reg.R[15]
    sta = cpu.Reg

    cpu.Jit.TestInst(op, f)

    sav = cpu.Reg

    cpu.Reg = sta
}

func (j *Jit) EndTest(op uint32, compare bool) {

    cpu := j.Cpu
    if !(compare && cpu.Reg != sav) {
        return
    }

    fmt.Printf("STA REG %08X CPSR %08X\n", sta.R, sta.CPSR.Get())
    fmt.Printf("JIT REG %08X CPSR %08X\n", sav.R, sav.CPSR.Get())
    fmt.Printf("COR REG %08X CPSR %08X\n", cpu.Reg.R, cpu.Reg.CPSR.Get())
    panic(fmt.Sprintf("Bad Compare %08X %08X", rpc, op))
}
