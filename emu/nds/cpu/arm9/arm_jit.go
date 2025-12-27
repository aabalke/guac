package arm9

import (
	"log"
	"math"
	"math/bits"
	"unsafe"

	amd64 "github.com/aabalke/gojit"
	"github.com/aabalke/guac/emu/nds/utils"
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
        j.Shl(amd64.Imm(1), amd64.Rax)
        j.Bts(amd64.Imm(0), amd64.Rax)
        j.Bsr(amd64.Rax, amd64.Rbx)
        j.Sub(amd64.Imm(32), amd64.Ebx)
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
            //j.Xor(amd64.Rdi, amd64.Rdi)
            j.Movl(j.REG(rd), amd64.Ecx)
            j.Shl(amd64.Imm(32), amd64.Rcx)
            j.Movl(j.REG(rn), amd64.Edi)
            j.Add(amd64.Rcx, amd64.Rax)
            j.Add(amd64.Rdi, amd64.Rax)
        }

        if set {

            if inst == UMULL {
                j.Test(amd64.Rax, amd64.Rax)
            }

            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }

        j.Movl(amd64.Eax, j.REG(rn))
        j.Shr(amd64.Imm(32), amd64.Rax)
        j.Movl(amd64.Eax, j.REG(rd))
        return

    case SMULL, SMLAL:

        j.Xor(amd64.Rax, amd64.Rax)

        j.Movl(j.REG(rs), amd64.Eax)
        j.Movl(j.REG(rm), amd64.Ebx)

        // sign extend int64(uint32()), replace with MODSX?
        j.Shl(amd64.Imm(32), amd64.Rbx)
        j.Shl(amd64.Imm(32), amd64.Rax)
        j.Sar(amd64.Imm(32), amd64.Rbx)
        j.Sar(amd64.Imm(32), amd64.Rax)

        j.Imul(amd64.Rbx)

        if inst == SMLAL {
            //j.Xor(amd64.Rcx, amd64.Rcx)
            //j.Xor(amd64.Rdi, amd64.Rdi)
            j.Movl(j.REG(rd), amd64.Ecx)
            j.Shl(amd64.Imm(32), amd64.Rcx)
            j.Movl(j.REG(rn), amd64.Edi)
            j.Add(amd64.Rcx, amd64.Rax)
            j.Add(amd64.Rdi, amd64.Rax)
        }

        if set {

            if inst == SMULL {
                j.Test(amd64.Rax, amd64.Rax)
            }

            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }

        j.Movl(amd64.Eax, j.REG(rn))
        j.Shr(amd64.Imm(32), amd64.Rax)
        j.Movl(amd64.Eax, j.REG(rd))

        return
    }

    x := (op >> 5) & 1 != 0
    y := (op >> 6) & 1 != 0

    switch inst {

    case SMLAxy:

        j.Movl(j.REG(rm), amd64.Eax)

        if !x {
            j.Shl(amd64.Imm(48), amd64.Rax)
        } else {
            j.Shl(amd64.Imm(32), amd64.Rax)
        }

        j.Sar(amd64.Imm(48), amd64.Rax)

        j.Movl(j.REG(rs), amd64.Ebx)

        if !y {
            j.Shl(amd64.Imm(48), amd64.Rbx)
        } else {
            j.Shl(amd64.Imm(32), amd64.Rbx)
        }

        j.Sar(amd64.Imm(48), amd64.Rbx)


        j.Movl(j.REG(rn), amd64.Ecx)
        j.Shl(amd64.Imm(32), amd64.Rcx)
        j.Sar(amd64.Imm(32), amd64.Rcx)

        j.Imul(amd64.Rbx)

        j.Add(amd64.Rcx, amd64.Rax)

        j.Movl(amd64.Eax, j.REG(rd))

        j.MulQFlag(amd64.Rax)

    case SMULxy:

        j.Movl(j.REG(rm), amd64.Eax)
        j.Movl(j.REG(rs), amd64.Ebx)

        if !x {
            j.Shl(amd64.Imm(16), amd64.Eax)
        }

        if !y {
            j.Shl(amd64.Imm(16), amd64.Ebx)
        }

        j.Sar(amd64.Imm(16), amd64.Eax)
        j.Sar(amd64.Imm(16), amd64.Ebx)

        j.Imul(amd64.Ebx)

        j.Movl(amd64.Eax, j.REG(rd))

    case SMLALxy:

        j.Movl(j.REG(rs), amd64.Eax)
        j.Movl(j.REG(rm), amd64.Ebx)

        if x {
            j.Shr(amd64.Imm(16), amd64.Ebx)
        }

        if y {
            j.Shr(amd64.Imm(16), amd64.Eax)
        }

        j.Shl(amd64.Imm(48), amd64.Rax)
        j.Shl(amd64.Imm(48), amd64.Rbx)

        j.Sar(amd64.Imm(48), amd64.Rax)
        j.Sar(amd64.Imm(48), amd64.Rbx)

        j.Imul(amd64.Rbx)

        j.Xor(amd64.Rbx, amd64.Rbx)
        j.Movl(j.REG(rd), amd64.Ebx)
        j.Movl(j.REG(rn), amd64.Ecx)

        j.Shl(amd64.Imm(32), amd64.Rbx)

        j.Shl(amd64.Imm(32), amd64.Rcx)
        j.Sar(amd64.Imm(32), amd64.Rcx)


        j.Add(amd64.Rbx, amd64.Rax)
        j.Add(amd64.Rcx, amd64.Rax)

        j.Movl(amd64.Eax, j.REG(rn))
        j.Shr(amd64.Imm(32), amd64.Rax)
        j.Movl(amd64.Eax, j.REG(rd))

    case SMLAWySMLALWy:

        j.Movl(j.REG(rs), amd64.Eax)
        j.Movl(j.REG(rm), amd64.Ebx)

        if y {
            j.Shr(amd64.Imm(16), amd64.Rax)
        }

        j.Shl(amd64.Imm(32), amd64.Rbx)
        j.Shl(amd64.Imm(48), amd64.Rax)
        j.Sar(amd64.Imm(32), amd64.Rbx)
        j.Sar(amd64.Imm(48), amd64.Rax)

        j.Imul(amd64.Rbx)
        j.Sar(amd64.Imm(16), amd64.Rax)

        if !x {

            j.Movl(j.REG(rn), amd64.Ebx)
            j.Shl(amd64.Imm(32), amd64.Rbx)
            j.Sar(amd64.Imm(32), amd64.Rbx)

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
    j.Mov(amd64.Imm(1), amd64.R10)

    j.Cmp(amd64.Imm(math.MaxInt32), r)
    j.Cmovcc(amd64.CC_GE, amd64.R10, amd64.Rdi)

    j.Cmp(amd64.Imm(math.MinInt32), r)
    j.Cmovcc(amd64.CC_LE, amd64.R10, amd64.Rdi)

    j.Mov(amd64.Dl, Q)
}

func (j *Jit) emitSwp(op uint32) {

    isByte := (op >> 22) & 1 != 0
    rn := (op >> 16) & 0xF
    rd := (op >> 12) & 0xF
    rm := op & 0xF

    j.Movl(j.REG(rn), amd64.Eax)
    j.Movl(j.REG(rm), amd64.Ebx)

    j.Movl(amd64.Eax, j.SCRATCH(0))
    j.Movl(amd64.Ebx, j.SCRATCH(1))

    //j.Mov(amd64.Rax, amd64.R12)
    //j.Mov(amd64.Rbx, amd64.R13)

    if isByte {

        j.CallFunc(Read)
        j.Movl(amd64.Eax, j.REG(rd))

        j.Movl(j.SCRATCH(0), amd64.Eax)
        j.Movl(j.SCRATCH(1), amd64.Ebx)

        //j.Mov(amd64.R12, amd64.Rax)
        //j.Mov(amd64.R13, amd64.Rbx)
        j.And(amd64.Imm(0xFF), amd64.Rbx)
        j.CallFunc(Write)
        return
    }

    j.And(amd64.Imm(^0b11), amd64.Rax)
    j.CallFunc(Read32)
    //j.Mov(amd64.R12, amd64.Rcx)
    j.Movl(j.SCRATCH(0), amd64.Ecx)
    j.And(amd64.Imm(0b11), amd64.Ecx)
    j.Shl(amd64.Imm(0b11), amd64.Ecx)
    j.And(amd64.Imm(31), amd64.Ecx)
    j.RorCl(amd64.Eax)
    j.Movl(amd64.Eax, j.REG(rd))

    j.Movl(j.SCRATCH(0), amd64.Eax)
    j.Movl(j.SCRATCH(1), amd64.Ebx)
    //j.Mov(amd64.R12, amd64.Rax)
    //j.Mov(amd64.R13, amd64.Rbx)
    j.CallFunc(Write32)
}

func (j *Jit) emitQalu(op uint32) {

    boolean := amd64.Imm(1)
    maxInt32 := amd64.Imm(math.MaxInt32)
    minInt32 := amd64.Imm(math.MinInt32)

    inst := (op >> 20) & 0xF
    rn := (op >> 16) & 0xF
    rd := (op >> 12) & 0xF
    rm := op & 0xF

    j.Mov(j.REG(rm), amd64.Eax)
    j.Shl(amd64.Imm(32), amd64.Rax)
    j.Sar(amd64.Imm(32), amd64.Rax)

    j.Mov(j.REG(rn), amd64.Ebx)
    j.Shl(amd64.Imm(32), amd64.Rbx)
    j.Sar(amd64.Imm(32), amd64.Rbx)

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
        j.Mov(amd64.Imm(2), amd64.Rax)

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
            j.MovAbs(uint64(0xFFFF_FFFF), amd64.Rdi)
            j.Cmp(amd64.Rdi, amd64.Rax)
            skip = j.JccForward(amd64.CC_Z)
        } else {
            j.Sub(amd64.Rbx, amd64.Rax)
            // im not sure why special case for 0 is needed here but not golang
            j.MovAbs(uint64(0x0), amd64.Rdi)
            j.Cmp(amd64.Edi, amd64.Eax)
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

func (j *Jit) emitHalf(op uint32) {

    rn      := (op >> 16) & 0xF
    rd      := (op >> 12) & 0xF
    preFlag := (op >> 24) & 1 != 0
    load    := (op >> 20) & 1 != 0
    inst    := (op >> 5)  & 0b11
    wb      := (op >> 21) & 1 != 0 || !preFlag

    // ax rnv / pre
    // di offset
    // cx post
    // 

    j.Movl(j.REG(rn), amd64.Eax)

    if rn == PC {
        j.Add(amd64.Imm(8), amd64.Eax)
    }

    if imm := (op >> 22) & 1 != 0; imm {
        j.MovAbs(uint64((op & 0xF) | (((op >> 8) & 0xF) << 4)), amd64.Rdi)
	} else {
        j.Movl(j.REG(op & 0xF), amd64.Edi)
	}

    j.Mov(amd64.Rax, amd64.Rcx)

    if up := (op >> 23) & 1 != 0; up {
        j.Add(amd64.Edi, amd64.Ecx)
	} else {
        j.Sub(amd64.Edi, amd64.Ecx)
	}

    if preFlag {
        j.Mov(amd64.Rcx, amd64.Rax)
    }

    if inst == RESERVED {
        panic("unsupported half (reserved)")
    }

    if !load {

        j.Movl(j.REG(rd), amd64.Ebx)
        if rd == PC {
            j.Add(amd64.Imm(12), amd64.Ebx)
        }

        if wb {
            j.Movl(amd64.Ecx, j.REG(rn))
        }

		switch inst {
		case STRH:

            j.And(amd64.Imm(^1), amd64.Rax)
            j.And(amd64.Imm(0xFFFF), amd64.Rbx)
            j.CallFunc(Write16)

		case LDRD:

            j.And(amd64.Imm(^0b111), amd64.Rax)
            j.Movl(amd64.Eax, j.SCRATCH(0))

            j.CallFunc(Read32)
            j.Movl(amd64.Eax, j.REG(rd))

            j.Movl(j.SCRATCH(0), amd64.Eax)
            j.Add(amd64.Imm(4), amd64.Rax)

            j.CallFunc(Read32)
            j.Movl(amd64.Eax, j.REG(rd+1))

		case STRD:

            j.And(amd64.Imm(^0b111), amd64.Rax)
            j.Movl(amd64.Eax, j.SCRATCH(0))

            j.CallFunc(Write32)
            j.Movl(amd64.Eax, j.REG(rd))

            j.Movl(j.SCRATCH(0), amd64.Eax)
            j.Add(amd64.Imm(4), amd64.Rax)

            j.Movl(j.REG(rd+1), amd64.Ebx)
            if rd + 1 == PC {
                j.Add(amd64.Imm(12), amd64.Ebx)
            }

            j.CallFunc(Write32)
            j.Movl(amd64.Eax, j.REG(rd+1))
		}

    } else {

        if wb {
            j.Movl(amd64.Ecx, j.REG(rn))
        }

        switch inst {
        case LDRH:

            j.And(amd64.Imm(^1), amd64.Rax)
            j.CallFunc(Read16)
            j.Movl(amd64.Eax, j.REG(rd))

        case LDRSB:
            // sign-expand byte value
            j.CallFunc(Read)

            j.Shl(amd64.Imm(56), amd64.Rax)
            j.Sar(amd64.Imm(56), amd64.Rax)

            j.Movl(amd64.Eax, j.REG(rd))

        case LDRSH:
            // sign-expand half value
            j.And(amd64.Imm(^1), amd64.Rax)
            j.CallFunc(Read16)

            j.Shl(amd64.Imm(48), amd64.Rax)
            j.Sar(amd64.Imm(48), amd64.Rax)

            j.Movl(amd64.Eax, j.REG(rd))
        }
    }
}

func (j *Jit) emitSdt(op uint32) {

    rd   := (op >> 12) & 0xF
    rn   := (op >> 16) & 0xF
    reg  := (op >> 25) & 1 != 0
    pre  := (op >> 24) & 1 != 0
    up   := (op >> 23) & 1 != 0
    byte := (op >> 22) & 1 != 0
    load := (op >> 20) & 1 != 0
    wb   := (op >> 21) & 1 != 0 || !pre

    // offset
    if reg {
        j.emitSdtRegShift(op)
        CpuPointer = j.Cpu
        j.MovAbs(uint64(uintptr(unsafe.Pointer(CpuPointer))), CPU)

        j.Mov(amd64.Rbx, amd64.Rcx)
    } else {
        j.Mov(amd64.Imm(int32(op & 0xFFF)), amd64.Rcx)
    }

    // rn ebx, shift ecx

    // ax pre
    // bc post

    j.Movl(j.REG(rn), amd64.Ebx)

    if rn == PC {
        j.Add(amd64.Imm(8), amd64.Ebx)
    }

    if up {
        j.Add(amd64.Ecx, amd64.Ebx)
    } else {
        j.Sub(amd64.Ecx, amd64.Ebx)
    }

    if pre {
        j.Movl(amd64.Ebx, amd64.Eax)
    } else {
        j.Movl(j.REG(rn), amd64.Eax)
    }

    if wb {
        j.Movl(amd64.Ebx, j.REG(rn))
    }

    if load {
        if byte {
            j.CallFunc(Read)
        } else {
            j.Movl(amd64.Eax, j.SCRATCH(0))

            j.And(amd64.Imm(^0b11), amd64.Eax)
            j.CallFunc(Read32)

            j.Movl(j.SCRATCH(0), amd64.Ecx)
            j.And(amd64.Imm(0b11), amd64.Ecx)
            j.Shl(amd64.Imm(0b11), amd64.Ecx)
            j.And(amd64.Imm(31), amd64.Ecx)
            j.RorCl(amd64.Eax)
        }

        j.Movl(amd64.Eax, j.REG(rd))

    } else {

        j.Movl(j.REG(rd), amd64.Ebx)
        if rd == PC {
            j.Add(amd64.Imm(12), amd64.Ebx)
        }

        if byte {
            j.And(amd64.Imm(0xFF), amd64.Ebx)
            j.CallFunc(Write)
        } else {
            j.And(amd64.Imm(^0b11), amd64.Eax)
            j.CallFunc(Write32)
        }
    }
}

func (j *Jit) emitSdtRegShift(op uint32) {

    rm     := op & 0xF
    shType := (op >> 5) & 0b11
    shift  := (op >> 7) & 0x1F

    // ebx rm, ecx shift

	j.Movl(j.REG(rm), amd64.Ebx)

    if rm == PC {
        panic("rm cannot include pc sdt")
    }

    //if special := shift == 0; special {

    //    if rm == PC {
    //        j.Add(amd64.Imm(8), amd64.Ebx)
    //    }

    //    switch shType{
    //    case LSL:
    //    case LSR:
    //        j.Xor(amd64.Ebx, amd64.Ebx)
    //    case ASR:
    //        j.Sar(amd64.Imm(31), amd64.Ebx)
    //    case ROR:
    //        j.Rcr(amd64.Imm(1), amd64.Ebx)
    //    }
    //    return
    //}

    switch shType {
    case LSL:

        j.Mov(amd64.Imm(shift), amd64.Rcx)
        j.ShlCl(amd64.Ebx)

        j.Cmp(amd64.Imm(32), amd64.Ecx)
        j.Sbb(amd64.Eax, amd64.Eax)
        j.And(amd64.Eax, amd64.Ebx)
        return

    case LSR:

        j.Mov(amd64.Imm(shift), amd64.Rcx)
        j.ShrCl(amd64.Ebx)

        j.Cmp(amd64.Imm(32), amd64.Ecx)
        j.Sbb(amd64.Eax, amd64.Eax)
        j.And(amd64.Eax, amd64.Ebx)
        return

    case ASR:

        j.Movl(amd64.Imm(shift), amd64.Ecx)

        j.Cmp(amd64.Imm(32), amd64.Ecx)
        j.Sbb(amd64.Eax, amd64.Eax)
        j.Not(amd64.Eax)
        j.Or(amd64.Eax, amd64.Ecx)

        j.SarCl(amd64.Ebx)
        return

    case ROR:

        j.Movl(amd64.Imm(shift), amd64.Ecx)
        j.RorCl(amd64.Ebx)
        return
    }
}

func (j *Jit) ToggleThumb() {

    j.Movl(j.REG(PC), amd64.Edi)
    j.And(amd64.Imm(1), amd64.Edi)
    //j.Cmp(amd64.Imm(0}, amd64.Edi)
    j.SETcc(amd64.CC_NZ, T)
    j.Cmovcc(amd64.CC_NZ, amd64.Imm(1), amd64.Ecx)
    j.Cmovcc(amd64.CC_Z, amd64.Imm(3), amd64.Ecx)

    j.Not(amd64.Ecx)
    j.And(amd64.Ecx, amd64.Edi)
    j.Movl(amd64.Edi, j.REG(PC))

	//reg.CPSR.T = reg.R[PC]&1 > 0

	//if reg.CPSR.T {
	//	reg.R[PC] &^= 1
	//	return
	//}

	//reg.R[PC] &^= 3
}

func (j *Jit) emitAluOp2Reg(op uint32) {

    shReg    := (op >> 4) & 1 != 0
    shType   := (op >> 5) & 0b11
    setCarry := (op >> 20) & 1 != 0
    inst     := (op >> 21) & 0xF
    logical  := inst & 0b0110 == 0b0000 || inst & 0b1100 == 0b1100
    rm       := op & 0xF

    setCarry  = setCarry && logical
    if setCarry {
        j.Mov(amd64.Imm(1), amd64.Rdi)
    } else {
        j.Mov(amd64.Imm(0), amd64.Rdi)
    }

    // rbx: op2
    // rcx: shift
    // rdx: original carry
    // rdi: setcarry

    j.Movb(C, amd64.Dl)
	j.Movl(j.REG(rm), amd64.Ebx)

    if shReg {
        rs := (op >> 8) & 0xF

        j.Movl(j.REG(rs), amd64.Ecx)
        j.And(amd64.Imm(0xFF), amd64.Ecx)

        if rm == PC {
            j.Add(amd64.Imm(12), amd64.Ebx)
        }


    } else {

        shift := (op >> 7) & 0x1F

        if rm == PC {
            j.Add(amd64.Imm(8), amd64.Ebx)
        }

        if special := shift == 0; special {
            switch shType{
            case LSL:
            case LSR:

                j.Bt(amd64.Imm(31), amd64.Ebx)
                j.SETcc(amd64.CC_C, amd64.Al)
                j.Movb(amd64.Al, C)

                // clear op2
                j.Xor(amd64.Ebx, amd64.Ebx)
            case ASR:

                // sar sets everything to top bit
                // if setcarry, set carry to bit as well

                j.Sar(amd64.Imm(31), amd64.Ebx)

                if setCarry {
                    j.Bt(amd64.Imm(31), amd64.Ebx)
                    j.SETcc(amd64.CC_C, amd64.Al)
                    j.Movb(amd64.Al, C)
                }

            case ROR:

                // CF = old carry (EDX & 1)
                j.Bt(amd64.Imm(0), amd64.Edx)

                // RRX
                j.Rcr(amd64.Imm(1), amd64.Ebx)

                // CPSR.C = new carry
                j.SETcc(amd64.CC_C, amd64.Al)
                j.Movb(amd64.Al, C)

            }

            return
        }

        j.Mov(amd64.Imm(shift), amd64.Rcx)
    }

    j.Test(amd64.Rcx, amd64.Rcx)
    zeroJump := j.JccForward(amd64.CC_Z)

    //// https://iitd-plos.github.io/col718/ref/arm-instructionset.pdf

    switch shType {
    case LSL, LSR:

        j.Cmp(amd64.Imm(32), amd64.Ecx)
        shift32 := j.JccForward(amd64.CC_A)
        equal := j.JccForward(amd64.CC_Z)

        if shType == LSL {
            // carry = op2 & (1 << (32-shift)) != 0
            j.Mov(amd64.Imm(32), amd64.Rax)
            j.Sub(amd64.Ecx, amd64.Eax)
            j.Bt(amd64.Eax, amd64.Ebx)
            j.SETcc(amd64.CC_C, amd64.Dl)
            // op2 <<= shift
            j.ShlCl(amd64.Ebx)
        } else {
            // carry = op2 & (1 << (shift-1)) != 0
            j.Mov(amd64.Rcx, amd64.Rax)
            j.Sub(amd64.Imm(1), amd64.Eax)
            j.Bt(amd64.Eax, amd64.Ebx)
            j.SETcc(amd64.CC_C, amd64.Dl)
            // op2 <<= shift
            j.ShrCl(amd64.Ebx)
        }

        done := j.JmpForward()

        shift32()

        // carry = op2 & 1 != 0
        j.Mov(amd64.Rbx, amd64.Rdx)
        j.And(amd64.Imm(1), amd64.Rdx)
        // op2 = 0
        j.Xor(amd64.Rbx, amd64.Rbx)

        done2 := j.JmpForward()

        equal()

        // LSL: carry = op2 & 1 != 0 
        // LSR: carry = op2 & 0x8000_0000 != 0 
        j.Mov(amd64.Rbx, amd64.Rdx)
        if shType == LSL {
            j.And(amd64.Imm(1), amd64.Rdx)
        } else {
            j.Shr(amd64.Imm(31), amd64.Rdx)
        }

        // op2 = 0
        j.Xor(amd64.Rbx, amd64.Rbx)

        done()
        done2()

    case ASR:

        j.Cmp(amd64.Imm(32), amd64.Ecx)
        shift32ge := j.JccForward(amd64.CC_AE)

        // carry = op2 & (1 << (shift-1)) != 0
        j.Mov(amd64.Rcx, amd64.Rax)
        j.Sub(amd64.Imm(1), amd64.Eax)
        j.Bt(amd64.Eax, amd64.Ebx)
        j.SETcc(amd64.CC_C, amd64.Dl)
        // op2 <<= shift
        j.SarCl(amd64.Ebx)

        done := j.JmpForward()

        shift32ge()

        // op and carry == top bit sar
        j.Sar(amd64.Imm(31), amd64.Ebx)
        j.Bt(amd64.Imm(0), amd64.Ebx)
        j.SETcc(amd64.CC_C, amd64.Dl)

        done()

    case ROR:

        j.Cmp(amd64.Imm(32), amd64.Ecx)
        equal := j.JccForward(amd64.CC_Z)

        // carry = (op2 >> ((shift-1) & 31)) & 1 != 0
        j.Mov(amd64.Rcx, amd64.Rax)
        j.Sub(amd64.Imm(1), amd64.Eax)
        j.And(amd64.Imm(31), amd64.Eax)

        j.Bt(amd64.Eax, amd64.Ebx)
        j.SETcc(amd64.CC_C, amd64.Dl)

        // op2 ror shift
        j.RorCl(amd64.Ebx)

        done := j.JmpForward()

        equal()

        // op2 unchanged
        // carry = op2 & 0x8000_0000 != 0
        j.Mov(amd64.Rbx, amd64.Rdx)
        j.Shr(amd64.Imm(31), amd64.Rdx)

        done()
    }

    j.Test(amd64.Rdi, amd64.Rdi)

    skip := j.JccForward(amd64.CC_Z)

    j.Movb(amd64.Dl, C)

    zeroJump()
    skip()
}

func (j *Jit) emitAlu(op uint32) {

    inst := (op >> 21) & 0xF
    rd   := (op >> 12) & 0xF
    rn   := (op >> 16) & 0xF
    imm  := (op >> 25) & 1 != 0
    set  := (op >> 20) & 1 != 0

    // reg and imm

    if inst == 5 || inst == 7 || inst == 6 {
        j.Xor(amd64.Rcx, amd64.Rcx)
        j.Mov(C, amd64.Cl)
        j.Mov(amd64.Rcx, amd64.R8)
    }

    if imm {

        ro := ((op >> 8) & 0xF) << 1
        op2 := bits.RotateLeft32(op & 0xFF, -int(ro))

        j.Mov(amd64.Imm(int32(op2)), amd64.Rbx)

        if set && ro != 0 {
            j.Movb(amd64.Imm((op2 >> 31) & 1), C)
        }

        j.Movl(j.REG(rn), amd64.Eax)

        if rn == PC {
            j.Add(amd64.Imm(8), amd64.Eax)
        }
    } else {

        // get op2, op2 will be in bx
        // shift
        j.emitAluOp2Reg(op)

        j.Movl(j.REG(rn), amd64.Eax)

        if rn == PC {
            if imm := (op >> 4) & 1 != 0; imm {
                j.Add(amd64.Imm(8), amd64.Eax)
            } else {
                j.Add(amd64.Imm(12), amd64.Eax)
            }
        }
    }

    if inst == 5 || inst == 7 || inst == 6 {
        j.Mov(amd64.R8, amd64.Rcx)
    }

    CpuPointer = j.Cpu
    j.MovAbs(uint64(uintptr(unsafe.Pointer(CpuPointer))), CPU)

    aluInstJit[inst](j, op, rd)

    //j.Movl(j.REG(PC), amd64.Eax)

    //if rd != PC {
    //    j.Add(amd64.Imm(4), amd64.Eax)
    //    j.Movl(amd64.Eax, j.REG(PC))
    //    return
    //}

    //panic("jit alu pc == rd not supported, need to setup ind inst and exit")
}

var aluInstJit = [...]func(j *Jit, op, rd uint32) {

    // AND 
    func(j *Jit, op, rd uint32) {
        j.And(amd64.Ebx, amd64.Eax)

        if set := (op >> 20) & 1 != 0; set {
            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }

        j.Movl(amd64.Eax, j.REG(rd))
    },

    // EOR
    func(j *Jit, op, rd uint32) {
        j.Xor(amd64.Ebx, amd64.Eax)

        if set := (op >> 20) & 1 != 0; set {
            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }

        j.Movl(amd64.Eax, j.REG(rd))
    },

    // SUB
    func(j *Jit, op, rd uint32) {
        j.Sub(amd64.Ebx, amd64.Eax)

        if set := (op >> 20) & 1 != 0; set {
            j.SETcc(amd64.CC_O, V)
            j.SETcc(amd64.CC_NC, C)
            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }

        j.Movl(amd64.Eax, j.REG(rd))
    },

    // RSB
    func(j *Jit, op, rd uint32) {
        j.Sub(amd64.Eax, amd64.Ebx)

        if set := (op >> 20) & 1 != 0; set {
            j.SETcc(amd64.CC_O, V)
            j.SETcc(amd64.CC_NC, C)
            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }

        j.Movl(amd64.Ebx, j.REG(rd))
    },

    // ADD
    func(j *Jit, op, rd uint32) {
        j.Add(amd64.Ebx, amd64.Eax)

        if set := (op >> 20) & 1 != 0; set {
            j.SETcc(amd64.CC_O, V)
            j.SETcc(amd64.CC_C, C)
            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }

        j.Movl(amd64.Eax, j.REG(rd))
    },

    // ADC
    func(j *Jit, op, rd uint32) {
        j.Bt(amd64.Imm(0), amd64.Cl)
        j.Adc(amd64.Ebx, amd64.Eax)

        if set := (op >> 20) & 1 != 0; set {
            j.SETcc(amd64.CC_O, V)
            j.SETcc(amd64.CC_C, C)
            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }

        j.Movl(amd64.Eax, j.REG(rd))
    },

    // SBC
    func(j *Jit, op, rd uint32) {
        j.Bt(amd64.Imm(0), amd64.Cl)
        j.Cmc() // compliment carry (reverse for sub)
        j.Sbb(amd64.Ebx, amd64.Eax)

        if set := (op >> 20) & 1 != 0; set {
            j.SETcc(amd64.CC_O, V)
            j.SETcc(amd64.CC_NC, C)
            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }

        j.Movl(amd64.Eax, j.REG(rd))
    },

    // RSC
    func(j *Jit, op, rd uint32) {
        j.Bt(amd64.Imm(0), amd64.Cl)
        j.Cmc() // compliment carry (reverse for sub)
        j.Sbb(amd64.Eax, amd64.Ebx)

        if set := (op >> 20) & 1 != 0; set {
            j.SETcc(amd64.CC_O, V)
            j.SETcc(amd64.CC_NC, C)
            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }

        j.Movl(amd64.Ebx, j.REG(rd))
    },

    // TST 
    func(j *Jit, op, rd uint32) {
        j.And(amd64.Ebx, amd64.Eax)
        if set := (op >> 20) & 1 != 0; set {
            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }
    },

    // TEQ 
    func(j *Jit, op, rd uint32) {
        j.Xor(amd64.Ebx, amd64.Eax)
        if set := (op >> 20) & 1 != 0; set {
            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }
    },


    // CMP
    func(j *Jit, op, rd uint32) {
        j.Sub(amd64.Ebx, amd64.Eax)

        if set := (op >> 20) & 1 != 0; set {
            j.SETcc(amd64.CC_O, V)
            j.SETcc(amd64.CC_NC, C)
            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }
    },

    // CMN
    func(j *Jit, op, rd uint32) {
        j.Add(amd64.Ebx, amd64.Eax)

        if set := (op >> 20) & 1 != 0; set {
            j.SETcc(amd64.CC_O, V)
            j.SETcc(amd64.CC_C, C)
            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }
    },

    // ORR
    func(j *Jit, op, rd uint32) {
        j.Or(amd64.Ebx, amd64.Eax)

        if set := (op >> 20) & 1 != 0; set {
            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }

        j.Movl(amd64.Eax, j.REG(rd))
    },

    // MOV
    func(j *Jit, op, rd uint32) {
        j.Movl(amd64.Ebx, j.REG(rd))

        if set := (op >> 20) & 1 != 0; set {
            j.Test(amd64.Ebx, amd64.Ebx)
            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }
    },

    // BIC
    func(j *Jit, op, rd uint32) {

        j.Not(amd64.Ebx)
        j.And(amd64.Ebx, amd64.Eax)

        if set := (op >> 20) & 1 != 0; set {
            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }

        j.Movl(amd64.Eax, j.REG(rd))
    },

    // MVN
    func(j *Jit, op, rd uint32) {

        j.Not(amd64.Ebx)

        if set := (op >> 20) & 1 != 0; set {
            j.Test(amd64.Ebx, amd64.Ebx)
            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }

        j.Movl(amd64.Ebx, j.REG(rd))
    },
}

func (j *Jit) emitBlock(op uint32) {

    rlist := op & 0xFFFF
    up := (op >> 23) & 1 != 0
    rn := (op >> 16) & 0xF

	if rlist == 0 {

        j.Movl(j.REG(rn), amd64.Eax)

		if up {
            j.Add(amd64.Imm(0x40), amd64.Eax)
		} else {
            j.Sub(amd64.Imm(0x40), amd64.Eax)
        }

        j.Movl(amd64.Eax, j.REG(rn))
		return
	}

    pcIncluded := op & 0x8000 != 0
    pre  := (op >> 24) & 1 != 0
    psr  := (op >> 22) & 1 != 0
    wb   := (op >> 21) & 1 != 0
    load := (op >> 20) & 1 != 0
    //forceUserTemp := psr && (j.Cpu.Reg.CPSR.Mode != MODE_USR) && (!load || !pcIncluded)
    j.Xor(amd64.Rdi, amd64.Rdi)
    j.Xor(amd64.Rdi, amd64.Rdi)
    if psr && (!load || !pcIncluded) {
        j.Mov(MODE, amd64.Edi)
        j.Cmp(amd64.Imm(MODE_USR), amd64.Edi)
        j.SETcc(amd64.CC_NZ, amd64.Rdi)
    }

    j.Push(amd64.Edi)

    regCount := utils.CountBits(rlist)

    j.Movl(j.REG(rn), amd64.Eax)
    j.Mov(amd64.Rax, amd64.Rbx)

    j.And(amd64.Imm(^0b11), amd64.Rax)

    if up {
        j.Add(amd64.Imm(regCount * 4), amd64.Rbx)
    } else {
        j.Sub(amd64.Imm(regCount * 4), amd64.Rbx)
    }

    //rnRef := &c.Reg.R[rn]

    //if forceUser && rn == 13 {
    //    rnRef = &c.Reg.SP[BANK_ID[MODE_USR]]
    //}

    //if forceUser && rn == 14 {
    //    rnRef = &c.Reg.LR[BANK_ID[MODE_USR]]
    //}

    //rnv := *rnRef

    if rn == 13 || rn == 14 {

        // rdx already has forceuser bit

        j.And(amd64.Imm(1), amd64.Edi)

        j.Cmp(amd64.Imm(1), amd64.Edi)
        normal := j.JccForward(amd64.CC_NZ)

        switch rn {
        case 13:
            j.Movl(j.UserBankReg(false), amd64.Edi)
        case 14:
            j.Movl(j.UserBankReg(true), amd64.Edi)
        }

        userModeJump := j.JmpForward()

        normal()

        j.Movl(j.REG(rn), amd64.Edi)

        userModeJump()

    } else {
        j.Movl(j.REG(rn), amd64.Edi)
    }

    j.Push(amd64.Edi)

    reg := uint32(0)
    if !up {
        reg = 15
    }

    for range 16 {

        if disabled := (rlist >> reg) & 1 == 0; disabled {
            if up { reg++ } else { reg-- }
            continue
        }

        if pre {
            if up {
                j.Add(amd64.Imm(4), amd64.Rax)
            } else {
                j.Sub(amd64.Imm(4), amd64.Rax)
            }
        }

        if load {

            j.Push(amd64.Eax)
            j.Push(amd64.Ebx)
            j.Push(amd64.Ecx)

            j.CallFunc(Read32)

            if reg == 13 || reg == 14 {

                j.Movl(amd64.Indirect{ // push / pop is always 64bit
                    Base: amd64.Rsp,
                    Offset: int32(8 * 4),
                    Bits: 64}, 
                    amd64.Rdi,
                )

                j.And(amd64.Imm(1), amd64.Edi)

                j.Cmp(amd64.Imm(1), amd64.Edi)
                normal := j.JccForward(amd64.CC_NZ)

                switch reg {
                case 13:
                    j.Movl(amd64.Eax, j.UserBankReg(false))
                case 14:
                    j.Movl(amd64.Eax, j.UserBankReg(true))
                }

                userModeJump := j.JmpForward()

                normal()

                j.Movl(amd64.Eax, j.REG(reg))

                userModeJump()

            } else {
                j.Movl(amd64.Eax, j.REG(reg))
            }

            j.Pop(amd64.Ecx)
            j.Pop(amd64.Ebx)
            j.Pop(amd64.Eax)
        } else {

            j.Push(amd64.Eax)
            j.Push(amd64.Ebx)
            j.Push(amd64.Ecx)

            switch reg {
            case rn:
                j.Movl(amd64.Indirect{ // push / pop is always 64bit
                    Base: amd64.Rsp,
                    Offset: int32(8 * 3),
                    Bits: 64}, 
                    amd64.Rbx,
                )
            case PC:
                j.Movl(j.REG(15), amd64.Ebx)
                j.Add(amd64.Imm(12), amd64.Ebx)
            default:

                if reg == 13 || reg == 14 {

                    j.Movl(amd64.Indirect{ // push / pop is always 64bit
                        Base: amd64.Rsp,
                        Offset: int32(8 * 4),
                        Bits: 64}, 
                        amd64.Rdi,
                    )

                    j.And(amd64.Imm(1), amd64.Edi)

                    j.Cmp(amd64.Imm(1), amd64.Edi)
                    normal := j.JccForward(amd64.CC_NZ)
                    switch reg {
                    case 13:
                        j.Movl(j.UserBankReg(false), amd64.Ebx)
                    case 14:
                        j.Movl(j.UserBankReg(true), amd64.Ebx)
                    }

                    userModeJump := j.JmpForward()

                    normal()

                    j.Movl(j.REG(reg), amd64.Ebx)

                    userModeJump()

                } else {
                    j.Movl(j.REG(reg), amd64.Ebx)
                }
            }

            j.CallFunc(Write32)

            j.Pop(amd64.Ecx)
            j.Pop(amd64.Ebx)
            j.Pop(amd64.Eax)
        }

        if !pre {
            if up {
                j.Add(amd64.Imm(4), amd64.Rax)
            } else {
                j.Sub(amd64.Imm(4), amd64.Rax)
            }
        }

        if up {
            reg++
        } else {
            reg--
        }
    }

    j.Pop(amd64.Eax)
    j.Pop(amd64.Eax)

    if !load {
        if wb {
            j.Movl(amd64.Ebx, j.REG(rn))
        }

        return
    }

    if wb {
        if rnIncluded := (rlist>>rn) & 1 == 1; rnIncluded {
            isLast := (rlist < (1 << (rn + 1)))
            isOnly := regCount == 1
            if !isLast || isOnly {
                j.Movl(amd64.Ebx, j.REG(rn))
            }
        } else {
            j.Movl(amd64.Ebx, j.REG(rn))
        }
    }

    if !pcIncluded {
        return
    }

    panic("UNSETUP LDR MODE SWITCH")

    //if psr {
    //    c.ldmModeSwitch()
    //}

    //c.toggleThumb()
}
