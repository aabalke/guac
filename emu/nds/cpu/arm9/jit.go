package arm9

import (
	"fmt"
	"unsafe"

	"github.com/aabalke/gojit"
	"github.com/aabalke/guac/config"
)

var _ = fmt.Sprintf
var _ = config.Conf

const (
    PAGE_MASK = 0xFFFF
    PAGE_SHIFT = 16

    CONCURRENT_BLOCKS = true // this is for testing
)

var (
    CpuPointer *Cpu
    CPU    = gojit.R9
    REG    = int32(unsafe.Offsetof(Cpu{}.Reg))
    R      = REG + int32(unsafe.Offsetof(Reg{}.R))
    CPSR   = REG + int32(unsafe.Offsetof(Reg{}.CPSR))
    JIT   = int32(unsafe.Offsetof(Cpu{}.Jit))
    SCRATCH   = JIT + int32(unsafe.Offsetof(Jit{}.Scratch))
    MODE = gojit.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.Mode)), Bits: 32}
    N = gojit.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.N)), Bits: 8}
    Z = gojit.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.Z)), Bits: 8}
    C = gojit.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.C)), Bits: 8}
    V = gojit.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.V)), Bits: 8}
    Q = gojit.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.Q)), Bits: 8}
    I = gojit.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.I)), Bits: 8}
    F = gojit.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.F)), Bits: 8}
    T = gojit.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.T)), Bits: 8}
    HALTED_FLAG = gojit.Indirect{Base: CPU, Offset: int32(unsafe.Offsetof(Cpu{}.Halted)), Bits: 8}
)

type Jit struct {
    *gojit.Assembler
	Cpu    *Cpu

    Pages   [0x1_0000_0000 >> PAGE_SHIFT]*Page // need 0xFFFF_FFFF for bios
    Metrics [0x1_0000_0000 >> PAGE_SHIFT][]uint32
    Cnt int

    invalidPages []uint32

    // used for testing individual instructions
    testFunc func()

    callFrameSize int32
    frameSize     int32

    blockCh chan uint32

    Scratch [0x10]uint32
}

type Page struct {
    id      uint32
    Blocks  []*JitBlock
    Written bool
}

type JitBlock struct {
    Skip      bool
    f         func()
	initPc    uint32
    finalPc   uint32
    finalOp   uint32
	Length    uint32
    assembler *gojit.Assembler
}

func NewJit(cpu *Cpu) *Jit {

    j := &Jit{
        Cpu: cpu,
        blockCh: make(chan uint32, 1024),
    }

    if CONCURRENT_BLOCKS {
        go j.concBlockComp()
    }

    CpuPointer = cpu

    return j
}

func (j *Jit) concBlockComp() {
    for pc := range j.blockCh {
        j.CreateBlock(pc)
    }
}

func (j *Jit) UserBankReg(isLr bool) gojit.Indirect {

    // user bank id is 0, so return first uint32 in SP and LR [6]uint32s

    if isLr {
        return gojit.Indirect{
            Base: CPU,
            Offset: REG + int32(unsafe.Offsetof(Reg{}.LR)),
            Bits: 32,
        }
    }

    return gojit.Indirect{
        Base: CPU,
        Offset: REG + int32(unsafe.Offsetof(Reg{}.SP)),
        Bits: 32,
    }
}

func (j *Jit) REG(i uint32) gojit.Indirect {
    return gojit.Indirect{
        Base: CPU,
        Offset: R + int32(i * 4),
        Bits: 32,
    }
}

func (j *Jit) SCRATCH(i uint32) gojit.Indirect {

    // i do not understand how clobbering of stack and registers works with
    // function calls at this time. Simple solution is writing them to memory
    // this is slower and should be replaced with a stac kor register based method

    if i > uint32(len(j.Scratch)) {
        panic("Called scratch register in jit > len of scratch registers")
    }

    return gojit.Indirect{
        Base: CPU,
        Offset: SCRATCH + int32(i * 4),
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
    // this clears invalid pages after resources not being used by
    // cpu implimentations (if cpu writes to its own blocks we get big errors

    if len(j.invalidPages) == 0 {
        return
    }

    for _, v := range j.invalidPages {

        if j.Pages[v] == nil {
            continue
        }

        j.Cnt--

        for i := range j.Pages[v].Blocks {

            if j.Pages[v].Blocks[i] == nil {
                continue
            }

            if j.Pages[v].Blocks[i].assembler == nil {
                continue
            }

            j.Pages[v].Blocks[i].assembler.Release()
        }

        j.Pages[v] = nil
        j.Metrics[v] = nil
    }

    j.invalidPages = []uint32{}
}

func (j *Jit) CreateBlock(pc uint32) {

	pageIdx := pc >> PAGE_SHIFT
    blockIdx := (pc & PAGE_MASK) >> 2

    page := j.Pages[pageIdx]
    if page == nil {
		page = &Page{
			id:     pageIdx,
			Blocks: make([]*JitBlock, (1<<PAGE_SHIFT)>>2),
		}

        j.Pages[pageIdx] = page

        j.Cnt++
    }

    block := page.Blocks[blockIdx]
    if block != nil && block.Skip {
        return
    }

    const pagesize = 0x100000 // 1024 * 1024
    asm, err := gojit.New(pagesize)
    if err != nil {
        panic(err)
    }

    j.Assembler = asm

    j.MovAbs(uint64(uintptr(unsafe.Pointer(CpuPointer))), CPU)

    tempPc := pc
    var length, op, i, iA uint32

    p, ok := j.Cpu.mem.ReadPtr(tempPc, true)
    if !ok {
        panic("READ BAD")
    }

    for {

        op = *(*uint32)(unsafe.Add(p, i*4))

        //// this slows things
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

        if iA >= config.Conf.Nds.NdsJit.BatchInst {
            length += i
            break
        }

        if ok := j.emitOp(op); !ok {
            length += i
            break
        }

        i++
        iA++
        tempPc += 4
    }

    if length == 0 {
        page.Blocks[blockIdx] = &JitBlock{
            Skip: true,
        }

        return
    }

    gojit.ExitAssembler(asm)

    if err := asm.Error(); err != nil {
        println("err in block creation, skipping")
        return
    }

    page.Blocks[blockIdx] = &JitBlock{
        initPc: pc,
        assembler: asm,
        Length: length,
        finalOp: op,
        //finalPc: tempPc,
        f: func () {
            gojit.CallJit(&asm.Buf[0])
        },
    }
}

func (j *Jit) emitBranch(op, lastPc uint32) (ok bool, newPc uint32) {
    // only emit branch if always exectues
	if cond := op & 0xF000_0000; cond < 0xE000_0000 {
        return false, 0
    }

    if !isB(op) {
        return false, 0
    }


    if immLoop := op == 0xEAFFFFFE; immLoop {
        panic("IMM LOOP")
        // j.Cpu.Halted = true
        j.Movb(gojit.Imm(1), HALTED_FLAG)
        return true, lastPc
    }

    if isLink := (op >> 24) & 1 != 0; isLink {

        j.Movl(gojit.Imm(lastPc + 4), j.REG(14))

        //return false, 0
        //j.Movl(j.REG(15), gojit.Eax)
        //j.Add(gojit.Imm(4), gojit.Eax)
        //j.Movl(gojit.Eax, j.REG(14))
    }

    //j.Movl(j.REG(15), gojit.Eax)
    //j.Add(gojit.Imm((int32(op)<<8)>>6) + 8, gojit.Eax)
    //j.Movl(gojit.Eax, j.REG(15))

	newPc = lastPc + uint32((int32(op)<<8)>>6) + 8
    j.Movl(gojit.Imm(newPc), j.REG(15))

    return true, newPc
}

func (j *Jit) emitOp(op uint32) bool {


    jcctargets := j.emitCond(op)

    ok := j.DecodeARM(op)

	for _, tgt := range jcctargets {
		tgt()
	}

    if ok {
        j.Add(gojit.Imm(4), j.REG(PC))
    }

    return ok
}

//go:inline
func (j *Jit) emitCond(op uint32) []func() {

    // thank you rasky

	cond := op >> 28
	var jcctargets []func()

	switch cond {
	case 0xE, 0xF:
		// nothing to do, always executed
	case 0x0: // Z
		j.Bt(gojit.Imm(0), Z)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_NC))
	case 0x1: // !Z
		j.Bt(gojit.Imm(0), Z)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_C))
	case 0x2: // C
		j.Bt(gojit.Imm(0), C)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_NC))
	case 0x3: // !C
		j.Bt(gojit.Imm(0), C)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_C))
	case 0x4: // N
		j.Bt(gojit.Imm(0), N)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_NC))
	case 0x5: // !N
		j.Bt(gojit.Imm(0), N)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_C))
	case 0x6: // V
		j.Bt(gojit.Imm(0), V)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_NC))
	case 0x7: // !V
		j.Bt(gojit.Imm(0), V)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_C))
	case 0x8: // C && !Z
		j.Bt(gojit.Imm(0), C)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_NC))
		j.Bt(gojit.Imm(0), Z)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_C))
	case 0x9: // !C || Z
		j.Movb(C, gojit.Al)
		j.Xorb(gojit.Imm(1), gojit.Al)
		j.Orb(Z, gojit.Al)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_Z))
	case 0xC: // !Z && N==V
		j.Bt(gojit.Imm(0), Z)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_C))
		fallthrough
	case 0xA, 0xB: // N==V / N!=V
		j.Movb(N, gojit.Al)
		j.Xorb(V, gojit.Al)
		if cond == 0xA || cond == 0xC {
			jcctargets = append(jcctargets, j.JccForward(gojit.CC_NZ))
		} else {
			jcctargets = append(jcctargets, j.JccForward(gojit.CC_Z))
		}
	case 0xD: // Z || N==V / N!=V
		j.Movb(N, gojit.Al)
		j.Xorb(V, gojit.Al)
		j.Orb(Z, gojit.Al)
		jcctargets = append(jcctargets, j.JccForward(gojit.CC_Z))
	default:
		panic("unreachable")
	}

    return jcctargets
}

func (jit *Jit) DecodeARM(opcode uint32) bool {

    switch {
    case isBLX(opcode):
        return false
    case isPLD(opcode):
        return false
    }

    if swi := opcode & 0xF00_0000 == 0xF00_0000; swi {
        return false
	}

    switch {
    case isBkpt(opcode):
	case isB(opcode):
	case isBX(opcode):
	case isSDT(opcode):

        load := (opcode >> 20) & 1 != 0
        if rdpc := opcode & 0xF000 == 0xF000; rdpc && load {
            return false
        }

        jit.emitSdt(opcode)
        return true
	case isBlock(opcode):

        load := (opcode >> 20) & 1 != 0
        pcIncluded := opcode & 0x8000 != 0
        if pcIncluded && load {
            return false
        } 

        jit.emitBlock(opcode)
        return true

	case isHalf(opcode):

        load := (opcode >> 20) & 1 != 0
        if rdpc := opcode & 0xF000 == 0xF000; rdpc && load {
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

        if rdpc := opcode & 0xF000 == 0xF000; rdpc  {
            return false
        }

        if swiExit := opcode & 0x3F0_000F == 0x3F0_000F; swiExit {
            return false
        }

        jit.emitAlu(opcode)
        return true

    case isCoDataReg(opcode):
	}

    return false
}

func (j *Jit) TestInst(op uint32, f func(op uint32)) {

    asm, err := gojit.New(gojit.PageSize)
    if err != nil {
        panic(err)
    }

    j.Assembler = asm

    j.MovAbs(uint64(uintptr(unsafe.Pointer(CpuPointer))), CPU)

    f(op)

    gojit.ExitAssembler(asm)

    if err := asm.Error(); err != nil {
        panic(err)
    }

    gojit.CallJit(&asm.Buf[0])

    asm.Release()
}

//go:nosplit
func Read(addr uint32) uint32 {
    return CpuPointer.mem.Read8(addr, true)
}

//go:nosplit
func Read16(addr uint32) uint32 {
    return CpuPointer.mem.Read16(addr, true)
}

//go:nosplit
func Read32(addr uint32) uint32 {
    return CpuPointer.mem.Read32(addr, true)
}

//go:nosplit
func Write(addr uint32, v uint8) {
    CpuPointer.mem.Write8(addr, v, true)
}

//go:nosplit
func Write16(addr uint32, v uint16) {
    CpuPointer.mem.Write16(addr, v, true)
}

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

    pageIdx  := pc >> PAGE_SHIFT
    blockIdx := (pc & PAGE_MASK) >> 2 // aligned to word for arm

    if metrics := j.Metrics[pageIdx]; metrics == nil {
        j.Metrics[pageIdx] = make([]uint32, (1 << PAGE_SHIFT) >> 2)
    }

	j.Metrics[pageIdx][blockIdx]++
    if j.Metrics[pageIdx][blockIdx] <= threshold {
        return
    }

    if CONCURRENT_BLOCKS {
        select {
        case j.blockCh <-pc:
        default:
            j.Metrics[pageIdx][blockIdx]--
        }

        return
    }

    j.CreateBlock(pc)
}

var (
    rpc uint32
    sav Reg
    sta Reg
)

func (j *Jit) StartTest(op uint32, compare bool, f func(op uint32)) {

    if config.Conf.Nds.NdsJit.Enabled {
        panic("Jit Instruction Test is running with Jit Running")
    }

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
