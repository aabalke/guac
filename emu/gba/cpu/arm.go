package cpu

import (
	"fmt"

	utils "github.com/aabalke33/guac/emu/gba/util"
)

// ALU

const (
    AND = iota
    EOR
    SUB
    RSB
    ADD
    ADC
    SBC
    RSC
    TST
    TEQ
    CMP
    CMN
    ORR
    MOV
    BIC
    MVN
)

func (cpu *Cpu) GetByte(i uint32, offsetBit uint8) uint32 {
    return cpu.GetVarData(i, offsetBit + 4, offsetBit)
}

func (cpu *Cpu) GetVarData(i uint32, s, e uint8) uint32 {

    switch {
    case s >= 0x20:
        msg := fmt.Sprintf("GetVarData max s is 31. got=%d", s)
        panic(msg)
    case e >= 0x20:
        msg := fmt.Sprintf("GetVarData max e is 31. got=%d", e)
        panic(msg)
    case e <= s:
        msg := fmt.Sprintf("GetVarData e >= s. got=%d >= %d", e, s)
        panic(msg)
    }

    width := e - s
    mask := uint32((1 << width) - 1)
    return (i >> s) & mask
}

func (cpu *Cpu) GetOp2(opcode uint32) uint32 {

    if immediate := utils.BitEnabled(opcode, 25); immediate {
        // shifted immediate
        ror := cpu.GetVarData(opcode, 11, 8)
        nn := cpu.GetVarData(opcode, 7, 0)

        _ = ror * nn

        // will need to ror shift nn to get op2
        var op2 uint32 = 0
        return op2
    }

    // shifted reg

    return 0
}

func NewAluData(opcode uint32, cpu *Cpu) *Alu {

    alu := &Alu{
        Opcode: opcode,
        Cond: cpu.GetByte(opcode, 28),
        Inst: cpu.GetByte(opcode, 21),
        Immediate: utils.BitEnabled(opcode, 25),
        Set: utils.BitEnabled(opcode, 20),
        Rd: cpu.GetByte(opcode, 12),
        Rn: cpu.GetByte(opcode, 16),
        Op2: cpu.GetOp2(opcode),
    }

    if alu.Rn == PC {
        if shiftImmediate := alu.Immediate || !utils.BitEnabled(opcode, 4); shiftImmediate {
            alu.RnValue = cpu.Loc + 8
        } else {
            alu.RnValue = cpu.Loc + 12
        }
    } else {
        alu.RnValue = cpu.Reg.R[alu.Rn]
    }

    return alu
}

type Alu struct {
    Opcode, Rd, Rn, RnValue, Cond, Op2, Inst uint32
    Immediate, Set bool
}

func (cpu *Cpu) Alu(opcode uint32) {

    alu := NewAluData(opcode, cpu)

    switch alu.Inst {
    case AND, EOR, ORR, MOV, MVN, BIC: cpu.logical(alu)
    case ADD, ADC, SUB, SBC, RSB, RSC: cpu.arithmetic(alu)
    case TST, TEQ, CMP, CMN:           cpu.test(alu)
    }
}

func (cpu *Cpu) logical(alu *Alu) {

    var oper func (uint32, uint32) uint32

    switch alu.Inst {
    case AND: oper = func(u1, u2 uint32) uint32 { return u1 & u2 }
    case EOR: oper = func(u1, u2 uint32) uint32 { return u1 ^ u2 }
    case ORR: oper = func(u1, u2 uint32) uint32 { return u1 | u2 }
    case MOV: oper = func(u1, u2 uint32) uint32 { return u2 }
    case MVN: oper = func(u1, u2 uint32) uint32 { return ^u2 }
    case BIC: oper = func(u1, u2 uint32) uint32 { return u1 &^ u2 }
    }

    res := oper(alu.RnValue, alu.Op2)
    cpu.Reg.R[alu.Rd] = res

    if alu.Rd == 15 {

        if alu.Set {
//If S=1, Rd=R15; should not be used in user mode:
//  CPSR = SPSR_<current mode>
//  PC = result
//  For example: MOVS PC,R14  ;return from SWI (PC=R14_svc, CPSR=SPSR_svc).

            //
            return
        }

        return
    }

    if alu.Set {
        // not sure if c matters
        // C=carryflag of shift operation (not affected if LSL#0 or Rs=00h)
        //cpu.Reg.CPSR.SetFlag(FLAG_C, )
        cpu.Reg.CPSR.SetFlag(FLAG_N, utils.BitEnabled(res, 31))
        cpu.Reg.CPSR.SetFlag(FLAG_Z, res == 0)
        return
    }
}

func (cpu *Cpu) arithmetic(alu *Alu) {

    var oper func (uint64, uint64, uint64) uint64

    switch alu.Inst {
    case ADD: oper = func(u1, u2, u3 uint64) uint64 { return u1 + u2 }
    case ADC: oper = func(u1, u2, u3 uint64) uint64 { return u1 + u2 + u3 }
    case SUB: oper = func(u1, u2, u3 uint64) uint64 { return u1 - u2 }
    case SBC: oper = func(u1, u2, u3 uint64) uint64 { return u1 - u2 + u3 - 1 }
    case RSB: oper = func(u1, u2, u3 uint64) uint64 { return u2 - u1 }
    case RSC: oper = func(u1, u2, u3 uint64) uint64 { return u2 + u1 + u3 - 1 }
    }

    res := oper(uint64(alu.RnValue), uint64(alu.Op2), 0) // cy
    cpu.Reg.R[alu.Rd] = uint32(res)

    if alu.Rd == 15 {

        if alu.Set {
//If S=1, Rd=R15; should not be used in user mode:
//  CPSR = SPSR_<current mode>
//  PC = result
//  For example: MOVS PC,R14  ;return from SWI (PC=R14_svc, CPSR=SPSR_svc).

            //
            return
        }

        return

    }

    if alu.Set {

        var v bool
        rnSign := uint8(alu.RnValue >> 31) & 1
        opSign := uint8(alu.Op2 >> 31) & 1
        rSign  := uint8(res >> 31) & 1

        switch alu.Inst {
        case ADD, ADC: v = (rnSign == opSign) && (rSign != rnSign)
        case SUB, SBC: v = (rnSign != opSign) && (rSign != rnSign)
        case RSB, RSC: v = (rnSign != opSign) && (rSign != opSign)
        }

        cpu.Reg.CPSR.SetFlag(FLAG_V, v)
        cpu.Reg.CPSR.SetFlag(FLAG_C, res >= 0x1_0000_0000)
        cpu.Reg.CPSR.SetFlag(FLAG_N, utils.BitEnabled(uint32(res), 31))
        cpu.Reg.CPSR.SetFlag(FLAG_Z, uint32(res) == 0)
        return
    }
}

func (cpu *Cpu) test(alu *Alu) {

    var oper func (uint32, uint32) uint32

    switch alu.Inst {
    case TST: oper = func(u1, u2 uint32) uint32 { return u1 & u2 }
    case TEQ: oper = func(u1, u2 uint32) uint32 { return u1 ^ u2 }
    case CMP: oper = func(u1, u2 uint32) uint32 { return u1 - u2 }
    case CMN: oper = func(u1, u2 uint32) uint32 { return u1 + u2 }
    }

    _ = oper

    //res := oper(alu.RnValue, alu.Op2) // cy
    //cpu.SetAluArithmeticData(alu, res)
}

const (
    MUL = iota
    MLA
    UMAAL
    UMULL
    UMLAL
    SMULL
    SMLAL
)

type Mul struct {
    Opcode, Rd, Rn, Rs, Rm, Cond, Inst uint32
    Set bool
}

func NewMulData(opcode uint32, cpu *Cpu) *Mul {

    valid := cpu.GetByte(opcode, 4) == 0b1001
    valid = valid && cpu.GetVarData(opcode, 25, 27) == 0b000

    if !valid {
        panic("Malformed Muliply Instruction")
    }

    mul := Mul{
        // halfword multiplies are for ARM9 +
        Opcode: opcode,
        Cond: cpu.GetByte(opcode, 28),
        Inst: cpu.GetByte(opcode, 21),
        Set: utils.BitEnabled(opcode, 20),
        Rd: cpu.GetByte(opcode, 16),
        Rn: cpu.GetByte(opcode, 12),
        Rs: cpu.GetByte(opcode, 8),
        Rm: cpu.GetByte(opcode, 0),
    }
    return &mul
}

func (cpu *Cpu) Mul(opcode uint32) {

    mul := NewMulData(opcode, cpu)

    switch mul.Inst {
    case MUL: cpu.mul(mul)
    case MLA: cpu.mla(mul)
    case UMAAL: panic("UMAAL is unsupported")
    case UMULL: cpu.umull(mul)
    case UMLAL: cpu.umlal(mul)
    case SMULL: cpu.smull(mul)
    case SMLAL: cpu.smlal(mul)
    }
}

func (cpu *Cpu) mul(mul *Mul) {

    r := &cpu.Reg.R

    var oper func (uint32, uint32) uint32

    oper = func (u1, u2 uint32) uint32 { return u1 * u2 }

    res := oper(r[mul.Rm], r[mul.Rs])
    r[mul.Rd] = res

    // set flags
}

func (cpu *Cpu) mla(mul *Mul) {

    r := &cpu.Reg.R

    var oper func (uint32, uint32, uint32) uint32

    oper = func (u1, u2, u3 uint32) uint32 { return u1 * u2 + u3}

    res := oper(r[mul.Rm], r[mul.Rs], r[mul.Rn])
    r[mul.Rd] = res

    // set flags
}

func (cpu *Cpu) umull(mul *Mul) {

    r := &cpu.Reg.R

    var oper func (uint64, uint64) uint64

    oper = func (u1, u2 uint64) uint64 { return u1 * u2}

    res := oper(uint64(r[mul.Rm]), uint64(r[mul.Rs]))
    r[mul.Rd] = uint32(res >> 32)
    r[mul.Rn] = uint32(res)

    // set flags
}

func (cpu *Cpu) umlal(mul *Mul) {

    r := &cpu.Reg.R

    var oper func (uint64, uint64, uint64, uint64) uint64

    oper = func (u1, u2, u3, u4 uint64) uint64 { return u1 * u2 + (u3<<32 | u4) }

    res := oper(uint64(r[mul.Rm]), uint64(r[mul.Rs]), uint64(r[mul.Rd]), uint64(r[mul.Rn]))
    r[mul.Rd] = uint32(res >> 32)
    r[mul.Rn] = uint32(res)

    // set flags
}

func (cpu *Cpu) smull(mul *Mul) {

    r := &cpu.Reg.R

    var oper func (int64, int64) int64

    oper = func (u1, u2 int64) int64 { return u1 * u2 }

    res := oper(int64(int32(r[mul.Rm])), int64(int32(r[mul.Rs])))
    r[mul.Rd] = uint32(res >> 32)
    r[mul.Rn] = uint32(res)

    // set flags

}

func (cpu *Cpu) smlal(mul *Mul) {

    r := &cpu.Reg.R

    var oper func (int64, int64, int64, int64) int64

    oper = func (u1, u2, u3, u4 int64) int64 { return u1 * u2 + (u3<<32 | u4) }

    res := oper(int64(int32(r[mul.Rm])), int64(int32(r[mul.Rs])), int64(int32(r[mul.Rd])), int64(int32(r[mul.Rn])))
    r[mul.Rd] = uint32(res >> 32)
    r[mul.Rn] = uint32(res)

    // set flags
}

const (
    STR = iota
    LDR_PLD
)

type Sdt struct {
    Opcode, Rd, Rn, Cond, Inst, Offset, Shift, ShiftType, Rm uint32
    Set, Immediate, Load, WriteBack, MemoryMgmt, Pre, Up, Byte bool
}

func NewSdtData(opcode uint32, cpu *Cpu) *Sdt {
    valid := cpu.GetVarData(opcode, 26, 27) == 0b01
    if !valid {
        panic("Malformed Sdt Instruction")
    }

    sdt := Sdt{
        Opcode: opcode,
        Cond: cpu.GetByte(opcode, 28),
        Immediate: utils.BitEnabled(opcode, 25),
        Pre: utils.BitEnabled(opcode, 24),
        Up: utils.BitEnabled(opcode, 23),
        Byte: utils.BitEnabled(opcode, 22),
        Inst: cpu.GetVarData(opcode, 20, 20),
        Rn: cpu.GetByte(opcode, 16),
        Rd: cpu.GetByte(opcode, 12),
    }

    if sdt.Pre {
        sdt.WriteBack = utils.BitEnabled(opcode, 21)
    } else {
        sdt.MemoryMgmt = utils.BitEnabled(opcode, 21)
    }

    if sdt.Immediate {
        sdt.Offset = cpu.GetVarData(opcode, 0, 11)
    } else {
        sdt.Shift = cpu.GetVarData(opcode, 7, 11)
        sdt.ShiftType = cpu.GetVarData(opcode, 5, 6)

        if utils.BitEnabled(opcode, 4) {
            panic("Malformed Single Data Transfer")
        }

        sdt.Rm = cpu.GetByte(opcode, 0)
    }

    return &sdt
}

func (cpu *Cpu) Sdt(opcode uint32) {

    sdt := NewSdtData(opcode, cpu)

    switch sdt.Inst {
    case STR:
    case LDR_PLD:
    }
}
