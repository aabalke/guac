package arm9

import (
	"unsafe"

	"github.com/aabalke/guac/emu/nds/cpu/arm9/gojit"
	"github.com/aabalke/guac/emu/nds/cpu/arm9/gojit/amd64"
)

type Jit struct {
	Cpu    *Cpu
	Blocks [0x10]JitBlock
    *amd64.Assembler
}

type JitBlock struct {
	f      func() int64
	initPc uint32
    finalPc uint32
    finalOp uint32
	Length uint32
}

func NewJit(cpu *Cpu) *Jit {

    cpuPtr = uintptr(unsafe.Pointer(cpu))
    return &Jit{
        Cpu: cpu,
        Blocks: [0x10]JitBlock{},
    }
}

func CallBlk(asm *amd64.Assembler, framesize int32, f func()) {

    if framesize & 7 != 0 {
        panic("UNALIGNED FRAMESIZE")
    }

    // Requires manual sp adjustments

    asm.Sub(amd64.Imm{Val: framesize}, amd64.Rsp)

    f()

    asm.Add(amd64.Imm{Val: framesize}, amd64.Rsp)
}

//func CallStack(off int32, bits uint8) amd64.Operand {
//    return amd64.Indirect{Base: amd64.Rsp, Offset: off, Bits: bits}
//}

var (
    cpuPtr uintptr

    CPU = amd64.R15 // always set R15 to CPU ptr
    REG = int32(unsafe.Offsetof(Cpu{}.Reg))
    CPSR = REG + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR))
    PC_REG = amd64.Indirect{Base: CPU, Offset: REG + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.N)), Bits: 32}
    N = amd64.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.N)), Bits: 8}
    Z = amd64.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.Z)), Bits: 8}
    C = amd64.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.C)), Bits: 8}
    V = amd64.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.V)), Bits: 8}
    Q = amd64.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.Q)), Bits: 8}
    I = amd64.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.I)), Bits: 8}
    F = amd64.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.F)), Bits: 8}
    T = amd64.Indirect{Base: CPU, Offset: CPSR + int32(unsafe.Offsetof(Cpu{}.Reg.CPSR.T)), Bits: 8}
)

func (j *Jit) REG(i uint32) amd64.Operand {
    return amd64.Indirect{
        Base: CPU,
        Offset: REG + int32(i << 2),
        Bits: 32,
    }
}

func (j *Jit) CreateBlocks() {

    pc := uint32(0x2039A88)

    j.Blocks[0] = JitBlock{
        initPc: pc,
        Length: 8,
        finalPc: 0x0203_9AA0,
        finalOp: 0x1AFF_FFF8,
    }

    asm, err := amd64.New(gojit.PageSize)
    if err != nil {
        panic(err)
    }

    // setup, not sure if needed per Block
    asm.ABI = amd64.GoABI
    asm.MovAbs(uint64(cpuPtr), amd64.R15)


    //0x02039A88
    asm.MovAbs(uint64(cpuPtr), amd64.Rax)
    asm.Mov(amd64.Imm{Val: 0x2}, amd64.Rbx) // addr
    asm.CallFuncGo((*Cpu).Read32)

    asm.MovAbs(2, amd64.Rbx)
    asm.Cmp(amd64.Rax, amd64.Rbx)
    asm.SETcc(amd64.CC_Z, Z)

    asm.Mov(amd64.Imm{Val: 0x2039AA0}, PC_REG)

    asm.Ret() // make sure RAX is return value on amd64

    asm.BuildTo(&j.Blocks[0].f)
}

func (j *Jit) TestInst(op uint32) {

    asm, err := amd64.New(gojit.PageSize)
    if err != nil {
        panic(err)
    }

    j.Assembler = asm

    asm.ABI = amd64.GoABI
    asm.MovAbs(uint64(cpuPtr), amd64.R15)

    //j.emitClz(op)
    j.emitMul(op)

    asm.Ret()

    asm.BuildTo(&j.Blocks[10].f)
}
