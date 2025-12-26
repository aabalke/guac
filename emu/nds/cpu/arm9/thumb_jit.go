package arm9

import (
	"fmt"

	amd64 "github.com/aabalke/gojit"
)

func (j *Jit) emitThumbStack(op uint32) {

    nn := (op & 0b111_1111) & 4

    j.Movl(j.REG(SP), amd64.Eax)

    if sub := (op >> 7) & 1 != 0; sub {
        j.Sub(amd64.Imm(nn), amd64.Eax)
    } else {
        j.Add(amd64.Imm(nn), amd64.Eax)
    }

    j.Movl(amd64.Eax, j.REG(SP))
}

func (j *Jit) emitThumbRelative(op uint32) {

    isSp := (op >> 11) & 1 != 0
    rd := (op >> 8) & 0b11
    nn := (op & 0xFF) * 4

	if isSp {
        j.Movl(j.REG(SP), amd64.Eax)
        j.Add(amd64.Imm(nn), amd64.Eax)
        j.Movl(amd64.Eax, j.REG(rd))
		return
	}

    j.Movl(j.REG(PC), amd64.Eax)
    j.Add(amd64.Imm(4), amd64.Eax)
    j.And(amd64.Imm(^0b11), amd64.Eax)
    j.Add(amd64.Imm(nn), amd64.Eax)
    j.Movl(amd64.Eax, j.REG(rd))
}

func (j *Jit) emitThumbAlu(op uint32) {

    inst := (op >> 6) & 0xF
    rs   := (op >> 3) & 0x7
    rd   := op & 0x7

    j.Movl(j.REG(rs), amd64.Eax)
    j.Movl(j.REG(rd), amd64.Ebx)

    switch inst {
    case THUMB_MUL:
        j.Imul(amd64.Ebx)
        j.Test(amd64.Eax, amd64.Eax)
        j.Movl(amd64.Eax, j.REG(rd))
	case THUMB_TST:
        j.Test(amd64.Ebx, amd64.Eax)
    case THUMB_CMP:
        j.Cmp(amd64.Ebx, amd64.Eax)
    case THUMB_CMN:
        j.Neg(amd64.Ebx)
        j.Cmp(amd64.Ebx, amd64.Eax)
    case THUMB_AND:
        j.And(amd64.Ebx, amd64.Eax)
        j.Movl(amd64.Eax, j.REG(rd))
    case THUMB_EOR:
        j.Xor(amd64.Ebx, amd64.Eax)
        j.Movl(amd64.Eax, j.REG(rd))
    case THUMB_ORR:
        j.Or(amd64.Ebx, amd64.Eax)
        j.Movl(amd64.Eax, j.REG(rd))
    case THUMB_BIC:
        j.Not(amd64.Ebx)
        j.And(amd64.Ebx, amd64.Eax)
        j.Movl(amd64.Eax, j.REG(rd))
    case THUMB_MVN:
        j.Not(amd64.Ebx)
        j.Test(amd64.Ebx, amd64.Ebx)
        j.Movl(amd64.Ebx, j.REG(rd))
    case THUMB_NEG:
        j.Neg(amd64.Eax)
        j.Movl(amd64.Eax, j.REG(rd))

    case THUMB_ADC:
    case THUMB_SBC:


    case THUMB_LSL:
    case THUMB_LSR:
    case THUMB_ASR:
    case THUMB_ROR:

    }

    j.SETcc(amd64.CC_S, N)
    j.SETcc(amd64.CC_Z, Z)

}

func (j *Jit) emitThumbAddSub(op uint32) {

	inst  := (op >> 9) & 0b11
    rnImm := (op >> 6) & 0b111
	rs    := (op >> 3) & 0b111
	rd    := (op >> 0) & 0b111


    j.Movl(j.REG(rs), amd64.Eax)
    j.Movl(j.REG(rnImm), amd64.Ebx)

    switch inst {
	case THUMB_ADD:
        j.Add(amd64.Ebx, amd64.Eax)
        j.Movl(amd64.Eax, j.REG(rd))
		//res = rsValue + rnValue
	case THUMB_SUB:
        fmt.Printf("Thumb Jit not sure if right order sub\n")
        j.Sub(amd64.Ebx, amd64.Eax)
        j.Movl(amd64.Eax, j.REG(rd))
		//res = rsValue - rnValue
	case THUMB_ADDMOV:
        j.Add(amd64.Imm(rnImm), amd64.Eax)
        j.Movl(amd64.Eax, j.REG(rd))
		//res = rsValue + rnImm
	case THUMB_SUBImm:
        println("A")
        j.Sub(amd64.Imm(rnImm), amd64.Eax)
        j.Movl(amd64.Eax, j.REG(rd))
		//res = rsValue - rnImm
    }

    j.SETcc(amd64.CC_O, V)
    j.SETcc(amd64.CC_B, C)
    j.SETcc(amd64.CC_S, N)
    j.SETcc(amd64.CC_Z, Z)


}
