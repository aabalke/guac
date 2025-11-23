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

        j.Mov(amd64.R15, amd64.Rax)
        j.Movl(j.REG(rn), amd64.Ebx)

        j.Movl(j.REG(rm), amd64.Ecx)
        j.And(amd64.Imm{Val: 0xFF}, amd64.Rcx)

        j.Push(amd64.Rbx)
        j.Push(amd64.Rcx)

        j.CallFuncGo((*Cpu).Read8)
        j.Movl(amd64.Eax, j.REG(rd))

        j.Pop(amd64.Rcx)
        j.Pop(amd64.Rbx)

        j.Mov(amd64.R15, amd64.Rax)
        j.CallFuncGo((*Cpu).Write8)

        return
    }

    j.Mov(amd64.R15, amd64.Rax)

    j.Movl(j.REG(rn), amd64.Ebx)
    j.Movl(j.REG(rm), amd64.Ecx)

    j.Push(amd64.Rbx)
    j.Push(amd64.Rcx)

    j.And(amd64.Imm{Val: ^0b11}, amd64.Ebx)
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

    j.Mov(amd64.R15, amd64.Rax)
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

func (j *Jit) emitSdt(op uint32) {

    rd   := (op >> 12) & 0xF
    rn   := (op >> 16) & 0xF
    reg  := (op >> 25) & 1 != 0
    pre  := (op >> 24) & 1 != 0
    up   := (op >> 23) & 1 != 0
    byte := (op >> 22) & 1 != 0
    wb   := (op >> 21) & 1 != 0
    load := (op >> 20) & 1 != 0

    // gen address

    // eax off / ptr, ebx pre, ecx post

    OFF := amd64.Eax
    PRE := amd64.Ebx
    POST := amd64.Ecx

    // offset
    if reg {
        // alu op2 reg outputs in Ebx
        j.emitAluOp2Reg(op, false)
        j.Mov(amd64.Ebx, amd64.Eax)
    } else {
        j.Movl(amd64.Imm{Val: int32(op & 0xFFF)}, OFF)
    }

    j.Movl(j.REG(rn), POST)

    if rn == PC {
        j.Add(amd64.Imm{Val: 8}, POST)
    }

    if up {
        j.Add(OFF, POST)
    } else {
        j.Sub(OFF, POST)
    }

    if pre {
        j.Movl(POST, PRE)
    } else {
        j.Movl(j.REG(rn), PRE)
    }

    j.Push(PRE)
    j.Push(POST)

    j.Mov(amd64.R15, amd64.Rax)

    switch {
    case load && byte:
        j.CallFuncGo((*Cpu).Read8)
        j.Movl(amd64.Eax, j.REG(rd))

    case load && !byte:

        j.And(amd64.Imm{Val: ^0b11}, amd64.Rbx)
        j.CallFuncGo((*Cpu).Read32)

        // shift
        j.StackOffset(1, amd64.Rcx)
        j.And(amd64.Imm{Val: 0b11}, amd64.Rcx)
        j.Shl(amd64.Imm{Val: 3}, amd64.Rcx)
        j.And(amd64.Imm{Val: 31}, amd64.Ecx)
        j.RorCl(amd64.Eax)

        j.Movl(amd64.Eax, j.REG(rd))

        //if rd == PC {
        //    j.ToggleThumb()
        //    j.Sub(amd64.Imm{Val: 4}, j.REG(PC))
        //}

    case !load && byte:

        j.Movl(j.REG(rd), amd64.Ecx)

        j.CallFuncGo((*Cpu).Write8)

    case !load && !byte:

        j.And(amd64.Imm{Val: ^0b11}, amd64.Rbx)
        j.Movl(j.REG(rd), amd64.Ecx)
        j.CallFuncGo((*Cpu).Write32)

    }

    j.Pop(POST)
    j.Pop(PRE)

    if writeback := (!pre || wb) && !(load && rn == rd); writeback {
        j.Movl(POST, j.REG(rn))
    }
}

func (j *Jit) ToggleThumb() {

    j.Movl(j.REG(PC), amd64.Edx)
    j.And(amd64.Imm{Val: 1}, amd64.Edx)
    //j.Cmp(amd64.Imm{Val: 0}, amd64.Edx)
    j.SETcc(amd64.CC_NZ, T)
    j.Cmovcc(amd64.CC_NZ, amd64.Imm{Val: 1}, amd64.Ecx)
    j.Cmovcc(amd64.CC_Z, amd64.Imm{Val: 3}, amd64.Ecx)

    j.Not(amd64.Ecx)
    j.And(amd64.Ecx, amd64.Edx)
    j.Movl(amd64.Edx, j.REG(PC))

	//reg.CPSR.T = reg.R[PC]&1 > 0

	//if reg.CPSR.T {
	//	reg.R[PC] &^= 1
	//	return
	//}

	//reg.R[PC] &^= 3
}

func (j *Jit) emitAluOp2Reg(op uint32, setcarry bool) {
	shtype := (op >> 5) & 3
	byreg := op&0x10 != 0
    rm := op & 0xF

	// load value to be shifted/rotated into ebx
	j.Movl(j.REG(rm), amd64.Ebx)

	// now load into ECX that shift amount, that can be
	// either foudn in a register or as immediate
	if byreg {

        if rm == PC {
            j.Add(amd64.Imm{Val: 12}, amd64.Ebx)
        }

		if (op>>7)&1 != 0 {
			panic("bit7 in op2 with reg, should not be here")
		}
		// cpu.Regs[15] += 4
		// move shift amount from armreg into ECX
		j.Movl(j.REG((op>>8)&0xF), amd64.Ecx)
		// and ecx & 0xFF
		j.And(amd64.Imm{Val: 0xFF}, amd64.Ecx)
		// if ecx == 0 -> jump forward (ebx is ok as-is)
		op2end := j.JccShortForward(amd64.CC_Z)
		//j.AddCycles(1)

		switch shtype {
		case 3: // rot
			j.RorCl(amd64.Ebx)
			if setcarry {
				// set carry from x86 sign. We can't rely on the x86 carry
				// flag because it is different when CL=32 (for x86 it means
				// 0, so x86 carry is not affected).
				j.Test(amd64.Ebx, amd64.Ebx)
				j.SETcc(amd64.CC_S, amd64.R10)
			}
		case 2: // asr
			// Calculate shift = max(shift, 31). We actually put 0xFFFFFFFF
			// in ECX, but that is parsed as 31 by x86
			j.Cmp(amd64.Imm{Val: 32}, amd64.Ecx)
			j.Sbb(amd64.Eax, amd64.Eax)
			j.Not(amd64.Eax)
			j.Or(amd64.Eax, amd64.Ecx)

			// Shift right. This is now always performed correctly as we
			// maxed out the value before.
			j.SarCl(amd64.Ebx)

			if setcarry {
				j.SETcc(amd64.CC_B, amd64.R10) // x86 carry in R10

				// If the shift value was >= 32, EBX is either 0 or FFFFFFFF,
				// and the carry must be 0 or 1 (respectively).
				j.Test(amd64.Eax, amd64.Eax)
				j.Cmovcc(amd64.CC_NZ, amd64.Ebx, amd64.R10d)
				j.And(amd64.Imm{Val: 1}, amd64.R10d)
			}

		case 0, 1: // lsl / lsr
			if shtype == 0 {
				j.ShlCl(amd64.Ebx)
			} else {
				j.ShrCl(amd64.Ebx)
			}
			if !setcarry {
				// Adjust shifts for amounts >= 32; in ARM, shift amounts
				// are well-defined for amounts >= 32, like in Go.
				j.Cmp(amd64.Imm{Val: 32}, amd64.Ecx)
				j.Sbb(amd64.Eax, amd64.Eax)
				j.And(amd64.Eax, amd64.Ebx)
			} else {
				// We need to both adjust the result for shift >= 32 and
				// compute carry flag. The ARM carry flag can be computed like this:
				//   shift < 32: use x86 carry
				//   shift == 32: nothing was shifted (it's shift=0 in x86 semantic);
				//                use bit 0 or 31 of EBX (depending on shift direction)
				//   shift > 32: carry must be zero
				j.SETcc(amd64.CC_B, amd64.R10) // x86 carry in R10
				if shtype == 0 {
					j.Bt(amd64.Imm{Val: 0}, amd64.Ebx)
				} else {
					j.Bt(amd64.Imm{Val: 31}, amd64.Ebx)
				}
				j.SETcc(amd64.CC_B, amd64.R11) // EBX bit 0 or 31 in R11 (this will only be used if shift==32)

				j.Cmp(amd64.Imm{Val: 32}, amd64.Ecx)
				j.Cmovcc(amd64.CC_Z, amd64.R11, amd64.R10) // shift == 32 -> EBX 0/31 bit in R10
				j.Sbb(amd64.Eax, amd64.Eax)
				j.And(amd64.Eax, amd64.Ebx)

				j.Cmp(amd64.Imm{Val: 33}, amd64.Ecx) // shift >= 33 -> clear R10
				j.Sbb(amd64.Eax, amd64.Eax)
				j.And(amd64.Eax, amd64.Ebx)
				j.And(amd64.Eax, amd64.R10d)
			}
		}

		if setcarry {
			j.Movl(amd64.R10d, amd64.Eax)
			j.Movb(amd64.Al, C)
		}

		op2end()
	} else {
        if rm == PC {
            j.Add(amd64.Imm{Val: 8}, amd64.Ebx)
        }

		shift := (op >> 7) & 0x1F

		switch shtype {
		case 0: // lsl
			if shift == 0 {
				return
			}
			j.Shl(amd64.Imm{Val: int32(shift)}, amd64.Ebx)
			if setcarry {
				j.SETcc(amd64.CC_B, C)
			}
		case 1, 2: // lsr/asr
			if shift == 0 {
				// Equal to >>32 in Go, so bit31 is carry
				// and then clear the output or set it to -1
				if setcarry {
					j.Bt(amd64.Imm{Val: 31}, amd64.Ebx)
					j.SETcc(amd64.CC_B, C)
				}
				if shtype == 1 {
					j.Xor(amd64.Ebx, amd64.Ebx)
				} else {
					j.Sar(amd64.Imm{Val: 31}, amd64.Ebx)
				}
			} else {
				if shtype == 1 {
					j.Shr(amd64.Imm{Val: int32(shift)}, amd64.Ebx)
				} else {
					j.Sar(amd64.Imm{Val: int32(shift)}, amd64.Ebx)
				}
				if setcarry {
					j.SETcc(amd64.CC_B, C)
				}
			}
		case 3: // ror
			if shift == 0 {
				// shift == 0 -> rcr #1
				j.Bt(amd64.Imm{Val: 0}, C)
				j.Rcr(amd64.Imm{Val: 1}, amd64.Ebx)
			} else {
				j.Ror(amd64.Imm{Val: int32(shift)}, amd64.Ebx)
			}
			if setcarry {
				j.SETcc(amd64.CC_B, C)
			}
		}
	}
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
    }

    j.Movl(j.REG(rn), amd64.Eax)

    if imm {

        rot := uint((op >> 7) & 0x1E)
        op2 := ((op & 0xFF) >> rot) | ((op & 0xFF) << (32 - rot))

        j.Mov(amd64.Imm{Val: int32(op2)}, amd64.Rbx)

        if set {
            if rot != 0 {
                if op2>>31 != 0 {
                    j.Movb(amd64.Imm{Val: 1}, C)
                } else {
                    j.Movb(amd64.Imm{Val: 0}, C)
                }
            }
        }

        if rn == PC {
            j.Add(amd64.Imm{Val: 8}, amd64.Eax)
        }
    } else {

        // get op2
        // shift
        j.emitAluOp2Reg(op, set)

        if rn == PC {
            if imm := (op >> 4) & 1 != 0; imm {
                j.Add(amd64.Imm{Val: 8}, amd64.Eax)
            } else {
                j.Add(amd64.Imm{Val: 12}, amd64.Eax)
            }
        }
    }

    j.Movl(j.REG(rn), amd64.Eax)

    aluInstJit[inst](j, op, rd)
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
            j.SETcc(amd64.CC_B, C)
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
            j.SETcc(amd64.CC_B, C)
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
            j.SETcc(amd64.CC_B, C)
            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }

        j.Movl(amd64.Eax, j.REG(rd))
    },

    // ADC
    func(j *Jit, op, rd uint32) {
        j.Bt(amd64.Imm{Val: 0}, amd64.Cl)
        j.Adc(amd64.Ebx, amd64.Eax)

        if set := (op >> 20) & 1 != 0; set {
            j.SETcc(amd64.CC_O, V)
            j.SETcc(amd64.CC_B, C)
            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }

        j.Movl(amd64.Eax, j.REG(rd))
    },

    // SBC
    func(j *Jit, op, rd uint32) {
        j.Bt(amd64.Imm{Val: 0}, amd64.Cl)
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
        j.Bt(amd64.Imm{Val: 0}, amd64.Cl)
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
            j.SETcc(amd64.CC_B, C)
            j.SETcc(amd64.CC_S, N)
            j.SETcc(amd64.CC_Z, Z)
        }
    },

    // CMN
    func(j *Jit, op, rd uint32) {
        j.Add(amd64.Ebx, amd64.Eax)

        if set := (op >> 20) & 1 != 0; set {
            j.SETcc(amd64.CC_O, V)
            j.SETcc(amd64.CC_B, C)
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
