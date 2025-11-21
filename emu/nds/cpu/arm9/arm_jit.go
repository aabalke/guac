package arm9

import (
	"log"
	"math"

	"github.com/aabalke/guac/emu/nds/cpu/arm9/gojit/amd64"
	sys "golang.org/x/sys/cpu"
)

func (j *Jit) emitClz(op uint32) {

    // https://www.felixcloutier.com/x86/lzcnt
    // https://www.felixcloutier.com/x86/bsr
    // BMI1 adds LZCNT in ~ 2013. Older cpus req older method

	rd := (op >> 12) & 0xF
	rm := op & 0xF

    if !sys.X86.HasBMI1 {
        log.Printf("CPU is missing BMI1 inst set. CLZ Fallback called.\n")
        j.Xor(amd64.Rax, amd64.Rax)
        j.Movl(j.REG(rm), amd64.Eax)
        j.Shl(amd64.Imm{Val: 1}, amd64.Rax)
        j.Bts(amd64.Imm{Val: 0}, amd64.Rax)
        j.Bsr(amd64.Rax, amd64.Rbx)
        j.Sub(amd64.Imm{Val: 32}, amd64.Ebx)
        j.Neg(amd64.Ebx)
        j.Movl(amd64.Ebx, j.REG(rd))
        return
    }

    j.Movl(j.REG(rm), amd64.Eax)
    j.Lzcnt(amd64.Eax, amd64.Eax)
    j.Movl(amd64.Eax, j.REG(rd))
}

func (j *Jit) emitMul(op uint32) {

    inst := (op >> 21) & 0xF
    set := (op >> 20) & 1 != 0
    rd := (op >> 16) & 0xF
    rn := (op >> 12) & 0xF
    rs := (op >> 8) & 0xF
    rm := op & 0xF

    switch inst {
    case MUL, MLA:

        j.Xor(amd64.Rax, amd64.Rax)
        j.Movl(j.REG(rs), amd64.Eax)
        j.Mul(j.REG(rm))

        if inst == MLA {
            j.Add(j.REG(rn), amd64.Eax)
        }

        if set {

            if inst == MUL { // Mul does not update flags
                j.Test(amd64.Eax, amd64.Eax)
            }

            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }

        j.Movl(amd64.Eax, j.REG(rd))
        return

    case UMULL, UMLAL:

        j.Xor(amd64.Rax, amd64.Rax)

        j.Movl(j.REG(rs), amd64.Eax)
        j.Movl(j.REG(rm), amd64.Ebx)
        j.Mul(amd64.Rbx)

        if inst == UMLAL {
            //j.Xor(amd64.Rcx, amd64.Rcx)
            //j.Xor(amd64.Rdx, amd64.Rdx)
            j.Movl(j.REG(rd), amd64.Ecx)
            j.Shl(amd64.Imm{Val: 32}, amd64.Rcx)
            j.Movl(j.REG(rn), amd64.Edx)
            j.Add(amd64.Rcx, amd64.Rax)
            j.Add(amd64.Rdx, amd64.Rax)
        }

        if set {

            if inst == UMULL {
                j.Test(amd64.Rax, amd64.Rax)
            }

            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }

        j.Movl(amd64.Eax, j.REG(rn))
        j.Shr(amd64.Imm{Val: 32}, amd64.Rax)
        j.Movl(amd64.Eax, j.REG(rd))
        return

    case SMULL, SMLAL:

        j.Xor(amd64.Rax, amd64.Rax)

        j.Movl(j.REG(rs), amd64.Eax)
        j.Movl(j.REG(rm), amd64.Ebx)

        // sign extend int64(uint32()), replace with MODSX?
        j.Shl(amd64.Imm{Val: 32}, amd64.Rbx)
        j.Shl(amd64.Imm{Val: 32}, amd64.Rax)
        j.Sar(amd64.Imm{Val: 32}, amd64.Rbx)
        j.Sar(amd64.Imm{Val: 32}, amd64.Rax)

        j.Imul(amd64.Rbx)

        if inst == SMLAL {
            //j.Xor(amd64.Rcx, amd64.Rcx)
            //j.Xor(amd64.Rdx, amd64.Rdx)
            j.Movl(j.REG(rd), amd64.Ecx)
            j.Shl(amd64.Imm{Val: 32}, amd64.Rcx)
            j.Movl(j.REG(rn), amd64.Edx)
            j.Add(amd64.Rcx, amd64.Rax)
            j.Add(amd64.Rdx, amd64.Rax)
        }

        if set {

            if inst == SMULL {
                j.Test(amd64.Rax, amd64.Rax)
            }

            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }

        j.Movl(amd64.Eax, j.REG(rn))
        j.Shr(amd64.Imm{Val: 32}, amd64.Rax)
        j.Movl(amd64.Eax, j.REG(rd))

        return
    }

    x := (op >> 5) & 1 != 0
    y := (op >> 6) & 1 != 0

    switch inst {

    case SMLAxy:

        j.Movl(j.REG(rm), amd64.Eax)

        if !x {
            j.Shl(amd64.Imm{Val: 48}, amd64.Rax)
        } else {
            j.Shl(amd64.Imm{Val: 32}, amd64.Rax)
        }

        j.Sar(amd64.Imm{Val: 48}, amd64.Rax)

        j.Movl(j.REG(rs), amd64.Ebx)

        if !y {
            j.Shl(amd64.Imm{Val: 48}, amd64.Rbx)
        } else {
            j.Shl(amd64.Imm{Val: 32}, amd64.Rbx)
        }

        j.Sar(amd64.Imm{Val: 48}, amd64.Rbx)


        j.Movl(j.REG(rn), amd64.Ecx)
        j.Shl(amd64.Imm{Val: 32}, amd64.Rcx)
        j.Sar(amd64.Imm{Val: 32}, amd64.Rcx)

        j.Imul(amd64.Rbx)

        j.Add(amd64.Rcx, amd64.Rax)

        j.Movl(amd64.Eax, j.REG(rd))

        j.MulQFlag(amd64.Rax)

    case SMULxy:

        j.Movl(j.REG(rm), amd64.Eax)
        j.Movl(j.REG(rs), amd64.Ebx)

        if !x {
            j.Shl(amd64.Imm{Val: 16}, amd64.Eax)
        }

        if !y {
            j.Shl(amd64.Imm{Val: 16}, amd64.Ebx)
        }

        j.Sar(amd64.Imm{Val: 16}, amd64.Eax)
        j.Sar(amd64.Imm{Val: 16}, amd64.Ebx)

        j.Imul(amd64.Ebx)

        j.Movl(amd64.Eax, j.REG(rd))

    case SMLALxy:

        j.Movl(j.REG(rs), amd64.Eax)
        j.Movl(j.REG(rm), amd64.Ebx)

        if x {
            j.Shr(amd64.Imm{Val: 16}, amd64.Ebx)
        }

        if y {
            j.Shr(amd64.Imm{Val: 16}, amd64.Eax)
        }

        j.Shl(amd64.Imm{Val: 48}, amd64.Rax)
        j.Shl(amd64.Imm{Val: 48}, amd64.Rbx)

        j.Sar(amd64.Imm{Val: 48}, amd64.Rax)
        j.Sar(amd64.Imm{Val: 48}, amd64.Rbx)

        j.Imul(amd64.Rbx)

        j.Xor(amd64.Rbx, amd64.Rbx)
        j.Movl(j.REG(rd), amd64.Ebx)
        j.Movl(j.REG(rn), amd64.Ecx)

        j.Shl(amd64.Imm{Val: 32}, amd64.Rbx)

        j.Shl(amd64.Imm{Val: 32}, amd64.Rcx)
        j.Sar(amd64.Imm{Val: 32}, amd64.Rcx)


        j.Add(amd64.Rbx, amd64.Rax)
        j.Add(amd64.Rcx, amd64.Rax)

        j.Movl(amd64.Eax, j.REG(rn))
        j.Shr(amd64.Imm{Val: 32}, amd64.Rax)
        j.Movl(amd64.Eax, j.REG(rd))

    case SMLAWySMLALWy:

        j.Movl(j.REG(rs), amd64.Eax)
        j.Movl(j.REG(rm), amd64.Ebx)

        if y {
            j.Shr(amd64.Imm{Val: 16}, amd64.Rax)
        }

        j.Shl(amd64.Imm{Val: 32}, amd64.Rbx)
        j.Shl(amd64.Imm{Val: 48}, amd64.Rax)
        j.Sar(amd64.Imm{Val: 32}, amd64.Rbx)
        j.Sar(amd64.Imm{Val: 48}, amd64.Rax)

        j.Imul(amd64.Rbx)
        j.Sar(amd64.Imm{Val: 16}, amd64.Rax)

        if !x {

            j.Movl(j.REG(rn), amd64.Ebx)
            j.Shl(amd64.Imm{Val: 32}, amd64.Rbx)
            j.Sar(amd64.Imm{Val: 32}, amd64.Rbx)

            j.Add(amd64.Rbx, amd64.Rax)

            j.MulQFlag(amd64.Rax)
        }

        j.Movl(amd64.Eax, j.REG(rd))
    }
}

func (j *Jit) MulQFlag(r amd64.Register) {

    // set q flag if > MaxInt32 || < MinInt32
    // NEVER SET FALSE

    j.Mov(Q, amd64.Dl)
    j.Mov(amd64.Imm{Val: 1}, amd64.R10)

    j.Cmp(amd64.Imm{Val: math.MaxInt32}, r)
    j.Cmovcc(amd64.CC_GE, amd64.R10, amd64.Rdx)

    j.Cmp(amd64.Imm{Val: math.MinInt32}, r)
    j.Cmovcc(amd64.CC_LE, amd64.R10, amd64.Rdx)

    j.Mov(amd64.Dl, Q)
}

func (j *Jit) emitSwp(op uint32) {

    isByte := (op >> 22) & 1 != 0
    rn := (op >> 16) & 0xF
    rd := (op >> 12) & 0xF
    rm := op & 0xF

    if isByte {

        // rax ptr
        // rbx rn
        // rcx rm

        j.MovAbs(uint64(cpuPtr), amd64.Rax)
        j.Movl(j.REG(rn), amd64.Ebx)

        j.Movl(j.REG(rm), amd64.Ecx)
        j.And(amd64.Imm{Val: 0xFF}, amd64.Rcx)

        j.Push(amd64.Rbx)
        j.Push(amd64.Rcx)

        j.CallFuncGo((*Cpu).Read8)
        j.Movl(amd64.Eax, j.REG(rd))

        j.Pop(amd64.Rcx)
        j.Pop(amd64.Rbx)

        j.MovAbs(uint64(cpuPtr), amd64.Rax)
        j.CallFuncGo((*Cpu).Write8)

        return
    }

    j.MovAbs(uint64(cpuPtr), amd64.Rax)

    j.Movl(j.REG(rn), amd64.Ebx)
    j.Movl(j.REG(rm), amd64.Ecx)

    j.Push(amd64.Rbx)
    j.Push(amd64.Rcx)

    j.Xor(amd64.Imm{Val: 0b11}, amd64.Ebx)
    j.CallFuncGo((*Cpu).Read32)

    // rcx is rn after call since x86 needs CL for Ror

    j.Mov(amd64.Indirect{Base: amd64.Rsp, Offset: 16}, amd64.Rcx)
    j.And(amd64.Imm{Val: 0b11}, amd64.Ecx)
    j.Shl(amd64.Imm{Val: 0b11}, amd64.Ecx)
    j.And(amd64.Imm{Val: 31}, amd64.Ecx)

    j.RorCl(amd64.Eax)

    j.Movl(amd64.Eax, j.REG(rd))

    j.Pop(amd64.Rcx)
    j.Pop(amd64.Rbx)

    j.MovAbs(uint64(cpuPtr), amd64.Rax)
    j.CallFuncGo((*Cpu).Write32)
}

func (j *Jit) emitQalu(op uint32) {

    boolean := amd64.Imm{Val: 1}
    maxInt32 := amd64.Imm{Val: math.MaxInt32}
    minInt32 := amd64.Imm{Val: math.MinInt32}

    inst := (op >> 20) & 0xF
    rn := (op >> 16) & 0xF
    rd := (op >> 12) & 0xF
    rm := op & 0xF

    j.Mov(j.REG(rm), amd64.Eax)
    j.Shl(amd64.Imm{Val: 32}, amd64.Rax)
    j.Sar(amd64.Imm{Val: 32}, amd64.Rax)

    j.Mov(j.REG(rn), amd64.Ebx)
    j.Shl(amd64.Imm{Val: 32}, amd64.Rbx)
    j.Sar(amd64.Imm{Val: 32}, amd64.Rbx)

    switch inst{
    case QADD, QSUB:

        j.Mov(Q, amd64.Cl)

        if inst == QADD {
            j.Add(amd64.Rbx, amd64.Rax)
        } else {
            j.Sub(amd64.Rbx, amd64.Rax)
        }

        j.Cmp(maxInt32, amd64.Rax)
        skipA := j.JccForward(amd64.CC_LE)
        j.Mov(boolean, amd64.Rcx)
        j.Mov(maxInt32, amd64.Rax)
        skipA()

        j.Cmp(minInt32, amd64.Rax)
        skipB := j.JccForward(amd64.CC_GE)
        j.Mov(boolean, amd64.Rcx)
        j.MovAbs(uint64(0x8000_0000), amd64.Rax)
        skipB()

        j.Mov(amd64.Cl, Q)
        j.Movl(amd64.Eax, j.REG(rd))

    case QDADD, QDSUB:

        j.Mov(Q, amd64.Cl)

        // imul needs rax so tmp mov rax to r10 and back
        // also imul dst is rax, set to rbx to get proper order
        j.Mov(amd64.Rax, amd64.R10)
        j.Mov(amd64.Imm{Val: 2}, amd64.Rax)

        j.Imul(amd64.Rbx)
        j.Mov(amd64.Rax, amd64.Rbx)
        j.Mov(amd64.R10, amd64.Rax)

        // check if rn * 2 flips q flag

        j.Cmp(maxInt32, amd64.Rbx)
        skipA1 := j.JccForward(amd64.CC_LE)
        j.Mov(boolean, amd64.Rcx)
        j.Mov(maxInt32, amd64.Rbx)
        skipA1()

        j.Cmp(minInt32, amd64.Rbx)
        skipB1 := j.JccForward(amd64.CC_GE)
        j.Mov(boolean, amd64.Rcx)
        j.Mov(minInt32, amd64.Rbx)
        skipB1()

        var skip func()
        if inst == QDADD {
            j.Add(amd64.Rbx, amd64.Rax)
            // im not sure why special case for -1 is needed here but not golang
            j.MovAbs(uint64(0xFFFF_FFFF), amd64.Rdx)
            j.Cmp(amd64.Rdx, amd64.Rax)
            skip = j.JccForward(amd64.CC_Z)
        } else {
            j.Sub(amd64.Rbx, amd64.Rax)
            // im not sure why special case for 0 is needed here but not golang
            j.MovAbs(uint64(0x0), amd64.Rdx)
            j.Cmp(amd64.Edx, amd64.Eax)
            skip = j.JccForward(amd64.CC_Z)
        }

        j.Cmp(maxInt32, amd64.Rax)
        skipA := j.JccForward(amd64.CC_LE)
        j.Mov(boolean, amd64.Rcx)
        j.Mov(maxInt32, amd64.Rax)
        skipA()

        j.Cmp(minInt32, amd64.Rax)
        skipB := j.JccForward(amd64.CC_GE)
        j.Mov(boolean, amd64.Rcx)
        j.Mov(minInt32, amd64.Rax)
        skipB()

        skip()

        j.Mov(amd64.Cl, Q)
        j.Movl(amd64.Eax, j.REG(rd))
    }
}
