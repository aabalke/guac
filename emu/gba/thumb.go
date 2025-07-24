package gba

import (
	"fmt"

	"github.com/aabalke33/guac/emu/gba/utils"
)

var (
    _ = fmt.Sprintf("")
)

const (
	THUMB_AND = iota
	THUMB_EOR
	THUMB_LSL
	THUMB_LSR
	THUMB_ASR
	THUMB_ADC
	THUMB_SBC
	THUMB_ROR
	THUMB_TST
	THUMB_NEG // arm eq is RSBS
	THUMB_CMP
	THUMB_CMN
	THUMB_ORR
	THUMB_MUL
	THUMB_BIC
	THUMB_MVN
)

type ThumbAlu struct {
	Opcode, Inst, Rs, Rd uint16
}

var thumbAluData ThumbAlu

func (cpu *Cpu) ThumbAlu(opcode uint16) {

	thumbAluData.Opcode = opcode
	thumbAluData.Inst = uint16(utils.GetByte(uint32(opcode), 6))
    thumbAluData.Rs = uint16(utils.GetVarData(uint32(opcode), 3, 5))
    thumbAluData.Rd = uint16(utils.GetVarData(uint32(opcode), 0, 2))

    alu := &thumbAluData

    switch alu.Inst {
    case THUMB_MUL: cpu.thumbMuliply(alu)
    case THUMB_TST, THUMB_CMN, THUMB_CMP: cpu.thumbTest(alu)
    case THUMB_AND, THUMB_EOR, THUMB_ORR, THUMB_BIC, THUMB_MVN: cpu.thumbLogical(alu)
    default: cpu.thumbArithmetic(alu)
    }

    cpu.Reg.R[15] += 2
}

func (cpu *Cpu) thumbMuliply(alu *ThumbAlu) {

    r := &cpu.Reg.R

    res := uint64(r[alu.Rd]) * uint64(r[alu.Rs])

    r[alu.Rd] = uint32(res)

    // ARM < 4, carry flag destroyed, ARM >= 5, carry flag unchanged
    //cpu.Reg.CPSR.SetFlag(FLAG_C, false)

    cpu.Reg.CPSR.SetFlag(FLAG_N, utils.BitEnabled(uint32(res), 31))
    cpu.Reg.CPSR.SetFlag(FLAG_Z, uint32(res) == 0)
}

func (cpu *Cpu) thumbLogical(alu *ThumbAlu) {

    r := &cpu.Reg.R

    var res uint32
    a, b := r[alu.Rd], r[alu.Rs]

    switch alu.Inst {
    case THUMB_AND: res = a & b
    case THUMB_EOR: res = a ^ b
    case THUMB_ORR: res = a | b
    case THUMB_BIC: res = a &^ b
    case THUMB_MVN: res = ^b
    }

    r[alu.Rd] = res

    cpu.Reg.CPSR.SetFlag(FLAG_N, utils.BitEnabled(uint32(res), 31))
    cpu.Reg.CPSR.SetFlag(FLAG_Z, uint32(res) == 0)
}

func (cpu *Cpu) thumbArithmetic(alu *ThumbAlu) {

    r := &cpu.Reg.R

    var oper func (uint64, uint64, uint64) uint64

    var v, c bool

    switch alu.Inst {
    case THUMB_LSL: oper = func(u1, u2, u3 uint64) uint64 {

            if u2 > 32 {
                return 0
            }
           
            c = u1&(1<<(32-u2)) > 0
            
            return u1 << (u2 & 0xFF)
        }
    case THUMB_LSR: oper = func(u1, u2, u3 uint64) uint64 {

            c = u1&(1<<(u2-1)) > 0

            return u1 >> (u2 & 0xFF)
        }
    case THUMB_ASR: oper = func(u1, u2, u3 uint64) uint64 {

            if u2 > 32 {
                u2 = 32
            }

            if u2 > 0 {
                c = u1&(1<<(u2-1)) > 0
            }

            tmp := u1
            msb := tmp & 0x8000_0000

            for range u2 {
                tmp = (tmp >> 1) | msb
            }

            return uint64(tmp)
        }
    case THUMB_ADC: oper = func(u1, u2, u3 uint64) uint64 { return u1 + u2 + u3 }
    case THUMB_SBC: oper = func(u1, u2, u3 uint64) uint64 { 

            if u3 == 1 { u3 = 0 } else { u3 = 1 }

            return u1 - u2 - u3
        }
    case THUMB_ROR: oper = func(u1, u2, u3 uint64) uint64 {

            c = (u1>>((u2-1)%32))&1 > 0
            
            shift := u2 % 32
            tmp0 := u1 >> shift
            tmp1 := u1 << (32 - (shift))
            return tmp0 | tmp1
        }
    case THUMB_NEG: oper = func(_, u2, _ uint64) uint64 { return 0 - u2 }
    }

    carry := uint64(0)
    if cpu.Reg.CPSR.GetFlag(FLAG_C) {
        carry = 1
    }

    rdValue := uint64(r[alu.Rd])

    res := oper(rdValue, uint64(r[alu.Rs]), carry)

    r[alu.Rd] = uint32(res)

    rdSign := uint8(rdValue >> 31) & 1
    rsSign := uint8(r[alu.Rs] >> 31) & 1
    rSign  := uint8(res >> 31) & 1

    switch alu.Inst {
    case THUMB_ADC:
        v = (rdSign == rsSign) && (rSign != rdSign)
        c = res >= 0x1_0000_0000
    case THUMB_SBC, THUMB_NEG:
        v = (rdSign != rsSign) && (rSign != rdSign)
        c =  res < 0x1_0000_0000
    }

    cpu.Reg.CPSR.SetFlag(FLAG_N, utils.BitEnabled(uint32(res), 31))
    cpu.Reg.CPSR.SetFlag(FLAG_Z, uint32(res) == 0)

    switch alu.Inst {
    case THUMB_LSL, THUMB_LSR, THUMB_ASR, THUMB_ROR:
        if (r[alu.Rs] & 0xFF) == 0 {
            return
        }
    }

    cpu.Reg.CPSR.SetFlag(FLAG_C, c)

    switch alu.Inst {
    case THUMB_LSL, THUMB_LSR, THUMB_ASR, THUMB_ROR:
        return
    }

    cpu.Reg.CPSR.SetFlag(FLAG_V, v)
}

func (cpu *Cpu) thumbTest(alu *ThumbAlu) {

    r := &cpu.Reg.R

    var res uint64
    a, b := uint64(r[alu.Rd]), uint64(r[alu.Rs])
    rdValue := uint64(r[alu.Rd])

    switch alu.Inst {
    case THUMB_TST: res = a & b
    case THUMB_CMP: res = a - b
    case THUMB_CMN: res = a + b
    }

    rdSign := uint8(rdValue >> 31) & 1
    rsSign := uint8(r[alu.Rs] >> 31) & 1
    resSign  := uint8(res >> 31) & 1
    switch alu.Inst {
    case THUMB_CMP:
        v := (rdSign != rsSign) && (resSign != rdSign)
        c := res < 0x1_0000_0000

        cpu.Reg.CPSR.SetFlag(FLAG_V, v)
        cpu.Reg.CPSR.SetFlag(FLAG_C, c)
    case THUMB_CMN:
        v := (rdSign == rsSign) && (resSign != rdSign)
        c := res >= 0x1_0000_0000
        cpu.Reg.CPSR.SetFlag(FLAG_V, v)
        cpu.Reg.CPSR.SetFlag(FLAG_C, c)
    }

    cpu.Reg.CPSR.SetFlag(FLAG_N, utils.BitEnabled(uint32(res), 31))
    cpu.Reg.CPSR.SetFlag(FLAG_Z, uint32(res) == 0)
}

func (cpu *Cpu) HiRegBX(opcode uint16) int {

    // only cmp effects flags

    inst := uint16(utils.GetVarData(uint32(opcode), 8, 9))
    mSBd := utils.BitEnabled(uint32(opcode), 7)
    mSBs := utils.BitEnabled(uint32(opcode), 6)
    rs := uint16(utils.GetVarData(uint32(opcode), 3, 6))
    rd := uint16(utils.GetVarData(uint32(opcode), 0, 2))

    if inst != 3 && mSBd {
        rd |= 0b1000
    }

    if mSBs {
        rs |= 0b1000
    }

    r := &cpu.Reg.R
    cpsr := &cpu.Reg.CPSR

    switch {
    case inst == 0: 

        rsValue := uint64(r[rs])

        if rs == PC {
            rsValue += 4
        }

        rdValue := uint64(r[rd])

        res := rsValue + rdValue

        if rd == PC {
            r[rd] = utils.WordAlign(uint32(res)) + 4
        } else {
            r[rd] = uint32(res)
            r[PC] += 2
        }
        return 4

    case inst == 1:

        // HI REG CMP

        rsValue := uint64(r[rs])

        if rs == PC {
            rsValue += 4
        }

        rdValue := uint64(r[rd])

        res := rdValue - rsValue

        if rd != PC {
            r[PC] += 2
        }

        rdSign := uint8(rdValue >> 31) & 1
        rsSign := uint8(rsValue >> 31) & 1
        rSign  := uint8(res >> 31) & 1

        v := (rdSign != rsSign) && (rSign != rdSign)
        c := res < 0x1_0000_0000

        cpsr.SetFlag(FLAG_N, utils.BitEnabled(uint32(res), 31))
        cpsr.SetFlag(FLAG_Z, uint32(res) == 0)
        cpsr.SetFlag(FLAG_C, c)
        cpsr.SetFlag(FLAG_V, v)
        return 4

    case inst == 2:

        if nop := rs == 8 && rd == 8; nop {

            cycles := 3

            r[PC] += 2
            return cycles
        }

        // MOV HI REG //

        rsValue := uint64(r[rs])
        if rs == PC {
            rsValue += 4
        }

        if rd == PC {
            r[rd] = utils.HalfAlign(uint32(rsValue))

            return 4
        }

        r[rd] = uint32(rsValue)
        r[PC] += 2
        return 4

    case inst == 3 && mSBd: panic("UNSUPPORTED HI BLX")
    case inst == 3:

        if rs == PC {
            cpsr.SetFlag(FLAG_T, false)
            //R15: CPU switches to ARM state, and PC is auto-aligned as (($+4) AND NOT 2).
            r[PC] = (r[PC] + 4) &^ 2

            return 4
        }

        if setThumb := !utils.BitEnabled(r[rs], 0); setThumb {
            cpsr.SetFlag(FLAG_T, false)
            r[PC] = r[rs] &^ 0b11
        } else {
            r[PC] = r[rs] &^ 0b1
        }

        return 4
    }

    return 4
}

const (
    THUMB_ADD = iota
    THUMB_SUB
    THUMB_ADDMOV
    THUMB_SUBImm
)

func (cpu *Cpu) ThumbAddSub(opcode uint16) {

    inst := uint64(utils.GetVarData(uint32(opcode), 9, 10))
    rnImm :=  uint64(utils.GetVarData(uint32(opcode), 6, 8))
    rs := uint64(utils.GetVarData(uint32(opcode), 3, 5))
    rd := uint64(utils.GetVarData(uint32(opcode), 0, 2))

    r := &cpu.Reg.R
    cpsr := &cpu.Reg.CPSR

    rsValue := uint64(r[rs])
    rnValue := uint64(r[rnImm])

    var res uint64

    switch inst {
    case THUMB_ADD: res = rsValue + rnValue
    case THUMB_SUB: res = rsValue - rnValue
    case THUMB_ADDMOV: res = rsValue + rnImm
    case THUMB_SUBImm: res = rsValue - rnImm
    }

    r[rd] = uint32(res)

    var v, c bool
    rsSign := (rsValue >> 31) & 1
    rnSign := (rnValue >> 31) & 1
    imSign := uint64(uint32(rnImm) >> 31) & 1
    rSign  := uint64(int32(uint32(res)) >> 31) & 1

    switch inst {
    case THUMB_ADD:
        v = (rsSign == rnSign) && (rSign != rsSign)
        c = res >= 0x1_0000_0000
    case THUMB_ADDMOV:
        v = (rsSign == imSign) && (rSign != imSign)
        c = res >= 0x1_0000_0000
    case THUMB_SUB:
        v = (rsSign != rnSign) && (rSign != rsSign)
        c = res < 0x1_0000_0000
    case THUMB_SUBImm:
        v = (rsSign != imSign) && (rSign != rsSign)
        c = res < 0x1_0000_0000
    }

    cpsr.SetFlag(FLAG_N, utils.BitEnabled(uint32(res), 31))
    cpsr.SetFlag(FLAG_Z, uint32(res) == 0)
    cpsr.SetFlag(FLAG_C, c)
    cpsr.SetFlag(FLAG_V, v)

    if PC == rd {
        return
    }

    r[PC] += 2
}

const (
    THUMB_IMM_MOV = iota
    THUMB_IMM_CMP
    THUMB_IMM_ADD
    THUMB_IMM_SUB
)

func (cpu *Cpu) thumbImm(opcode uint16) {

    r := &cpu.Reg.R
    cpsr := &cpu.Reg.CPSR
    inst := utils.GetVarData(uint32(opcode), 11, 12)
    rd := uint64(utils.GetVarData(uint32(opcode), 8, 10))
    nn := uint64(utils.GetVarData(uint32(opcode), 0, 7))
    rdValue := uint64(r[rd])
    rdSign := uint8(rdValue >> 31) & 1
    nnSign := uint8(nn >> 31) & 1

    var res uint64
    var v, c bool

    switch inst {
    case THUMB_IMM_MOV:
        res = nn
    case THUMB_IMM_CMP:

        res = rdValue - nn
        rSign  := uint8(res >> 31) & 1
        v = (rdSign != nnSign) && (rSign != rdSign)
        c = res < 0x1_0000_0000
    case THUMB_IMM_ADD:
        res = rdValue + nn
        rSign  := uint8(res >> 31) & 1
        v = (rdSign == nnSign) && (rSign != rdSign)
        c = res >= 0x1_0000_0000
    case THUMB_IMM_SUB:
        res = rdValue - nn
        rSign  := uint8(res >> 31) & 1
        v = (rdSign != nnSign) && (rSign != rdSign)
        c = res < 0x1_0000_0000
    }

    if inst != THUMB_IMM_CMP {
        r[rd] = uint32(res)
    }

    if inst != THUMB_IMM_MOV {
        cpsr.SetFlag(FLAG_C, c)
        cpsr.SetFlag(FLAG_V, v)
    }

    cpsr.SetFlag(FLAG_N, utils.BitEnabled(uint32(res), 31))
    cpsr.SetFlag(FLAG_Z, uint32(res) == 0)

    if PC == rd {
        return
    }

    r[PC] += 2
}

func (cpu *Cpu) thumbLSHalf(opcode uint16) {

    // STRH // LDRH

    offset := utils.GetVarData(uint32(opcode), 6, 10) << 1
    rb := utils.GetVarData(uint32(opcode), 3, 5)
    rd := utils.GetVarData(uint32(opcode), 0, 2)
    ldr := utils.BitEnabled(uint32(opcode), 11)

    r := &cpu.Reg.R

    addr := r[rb] + offset

    if ldr {
        v := uint32(cpu.Gba.Mem.Read16(addr &^ 1))
        is := (addr & 1) * 8
        v, _, _ = utils.Ror(v, is, false, false ,false)
        r[rd] = v
    } else {
        cpu.Gba.Mem.Write16(addr, uint16(r[rd]))
    }

    r[PC] += 2
}

const (
    THUMB_STRH = iota
    THUMB_LDSB
    THUMB_LDRH
    THUMB_LDSH
)

func (cpu *Cpu) thumbLSSigned(opcode uint16) {

    inst := utils.GetVarData(uint32(opcode), 10, 11)
    ro := utils.GetVarData(uint32(opcode), 6, 8)
    rb := utils.GetVarData(uint32(opcode), 3, 5)
    rd := utils.GetVarData(uint32(opcode), 0, 2)

    r := &cpu.Reg.R

    addr := r[rb] + r[ro]
    switch inst {
    case THUMB_STRH: 
        cpu.Gba.Mem.Write16(addr, uint16(r[rd]))

    case THUMB_LDSB: r[rd] = uint32(int8(uint8(cpu.Gba.Mem.Read8(addr))))
    case THUMB_LDRH:
        v := cpu.Gba.Mem.Read16(utils.HalfAlign(addr))
        is := (addr & 0b1) * 8
        v, _, _ = utils.Ror(v, is, false, false ,false)
        r[rd] = v

    case THUMB_LDSH:

        // On ARM7 aka ARMv4 aka NDS7/GBA:
        // LDRSH Rd,[odd]  -->  LDRSB Rd,[odd]         ;sign-expand BYTE value
        // On ARM9 aka ARMv5 aka NDS9:
        // LDRSH Rd,[odd]  -->  LDRSH Rd,[odd-1]       ;forced align

        if misaligned := addr & 1 == 1; misaligned {
            // sign-expand BYTE value
            unexpanded := int16(cpu.Gba.Mem.Read16(addr))
            expanded := uint32(unexpanded)

            if int8(unexpanded) < 0 {
                expanded |= (0xFFFFFF << 8)
            }

            r[rd] = expanded
        } else {

            aligned := utils.HalfAlign(addr)

            // sign-expand half value
            unexpanded := int16(cpu.Gba.Mem.Read16(aligned))
            expanded := uint32(unexpanded)

            if unexpanded < 0 {
                expanded |= (0xFFFF << 16)
            }

            r[rd] = expanded
        }
    }

    r[PC] += 2
}

func (cpu *Cpu) thumbLPC(opcode uint16) {

    r := &cpu.Reg.R

    rd := utils.GetVarData(uint32(opcode), 8, 10)
    nn := utils.GetVarData(uint32(opcode), 0, 7) << 2
    addr := utils.WordAlign(r[PC] + 4 + nn)

    r[rd] = cpu.Gba.Mem.Read32(addr)
    r[PC] += 2
}

const (
    THUMB_STR_REG = iota
    THUMB_STRB_REG
    THUMB_LDR_REG
    THUMB_LDRB_REG
)

func (cpu *Cpu) thumbLSR(opcode uint16) {

    r := &cpu.Reg.R

    inst := utils.GetVarData(uint32(opcode), 10, 11)
    ro := utils.GetVarData(uint32(opcode), 6, 8)
    rb := utils.GetVarData(uint32(opcode), 3, 5)
    rd := utils.GetVarData(uint32(opcode), 0, 2)

    addr := r[rb] + r[ro]

    switch inst {
    case THUMB_STR_REG:
        cpu.Gba.Mem.Write32(addr, r[rd])
    case THUMB_STRB_REG:
        cpu.Gba.Mem.Write8(addr, uint8(r[rd]))
    case THUMB_LDR_REG:
        v := cpu.Gba.Mem.Read32(utils.WordAlign(addr))
        is := (addr & 0b11) * 8
        r[rd], _, _ = utils.Ror(v, is, false, false ,false)
    case THUMB_LDRB_REG:
        r[rd] = cpu.Gba.Mem.Read8(addr)
    }

    r[PC] += 2
}

const (
    THUMB_STR_IMM = iota
    THUMB_LDR_IMM
    THUMB_STRB_IMM
    THUMB_LDRB_IMM
)

func (cpu *Cpu) thumbLSImm(opcode uint16) {

    r := &cpu.Reg.R

    inst := utils.GetVarData(uint32(opcode), 11, 12)
    nn := utils.GetVarData(uint32(opcode), 6, 10)
    rb := utils.GetVarData(uint32(opcode), 3, 5)
    rd := utils.GetVarData(uint32(opcode), 0, 2)

    switch inst {
    case THUMB_STR_IMM:
        addr := r[rb] + (nn << 2)
        cpu.Gba.Mem.Write32(addr, r[rd])
    case THUMB_LDR_IMM:
        addr := r[rb] + (nn << 2)
        v := cpu.Gba.Mem.Read32(utils.WordAlign(addr))
        is := (addr & 0b11) * 8
        r[rd], _, _ = utils.Ror(v, is, false, false ,false)

    case THUMB_STRB_IMM:
        addr := r[rb] + nn
        cpu.Gba.Mem.Write8(addr, uint8(r[rd]))
    case THUMB_LDRB_IMM:
        addr := r[rb] + nn
        r[rd] = cpu.Gba.Mem.Read8(addr)
    }

    r[PC] += 2
}

func (cpu *Cpu) thumbPushPop(opcode uint16) {

    r := &cpu.Reg.R

    isPop := utils.BitEnabled(uint32(opcode), 11)
    pclr := utils.BitEnabled(uint32(opcode), 8)
    rlist := utils.GetVarData(uint32(opcode), 0, 7)

    if isPop {
        for reg := range 8 {
            if utils.BitEnabled(rlist, uint8(reg)) {
                r[reg] = cpu.Gba.Mem.Read32(r[SP])
                r[SP] += 4
            }
        }

        if pclr {

            r[PC] = utils.HalfAlign(cpu.Gba.Mem.Read32(r[SP]))
            r[SP] += 4
            //piplining
            return
        }

        r[PC] += 2
        return
    }

    //if CURR_INST > 0xc000 && opcode == 0xB5F0 { fmt.Printf("CURR %08X PC %08X\n\n", CURR_INST, r[PC])}

    if pclr {
        r[SP] -= 4
        cpu.Gba.Mem.Write32(r[SP], r[14])
    }

    for reg := 7; reg >= 0; reg-- {
        if utils.BitEnabled(rlist, uint8(reg)) {
            r[SP] -= 4
            cpu.Gba.Mem.Write32(r[SP], r[reg])
        }
    }

    r[PC] += 2
}

func (cpu *Cpu) thumbRelative(opcode uint16) {


    r := &cpu.Reg.R
    isSP := utils.BitEnabled(uint32(opcode), 11)
    rd := utils.GetVarData(uint32(opcode), 8, 10)
    nn := utils.GetVarData(uint32(opcode), 0, 7) * 4

    if isSP {
        r[rd] = r[13] + nn
        r[PC] += 2
        return
    }

    //r[rd] = r[PC] + 4 + nn
    r[rd] = utils.WordAlign(r[PC] + 4) + nn
    r[PC] += 2
}

func (cpu *Cpu) thumbJumpCalls(opcode uint16) {

    r := &cpu.Reg.R

    if !cpu.CheckCond(utils.GetByte(uint32(opcode), 8)) {
        r[PC] += 2
        return
    }

    mask := uint32(0b1_1111_1111)
    nn := utils.GetVarData(uint32(opcode), 0, 7) << 1
    negative := (nn >> 8) == 1
    nn &= mask

    if negative {
        nn |= ^mask
    }

    offset := int(nn)

    r[PC] = (uint32(int(r[PC]) + 4 + offset))
}

func (cpu *Cpu) thumbB(opcode uint16) {
    r := &cpu.Reg.R

    mask := uint32(0b111_1111_1111)

    nn := (utils.GetVarData(uint32(opcode), 0, 10))
    negative := (nn >> 10) == 1
    nn = nn & mask

    if negative {
        nn |= ^mask
    }

    offset := int32(nn) * 2

    r[PC] = uint32(int32(r[PC]) + 4 + offset)
}

func (cpu *Cpu) thumbShifted(opcode uint16) {
    reg := &cpu.Reg
    r := &cpu.Reg.R

    inst := utils.GetVarData(uint32(opcode), 11, 12)
    offset := utils.GetVarData(uint32(opcode), 6, 10)
    rs := utils.GetVarData(uint32(opcode), 3, 5)
    rd := utils.GetVarData(uint32(opcode), 0, 2)


    shiftArgs := utils.ShiftArgs{
        SType: inst, // ROR NOT POSSIBLE
        Val: r[rs],
        Is: offset,
        IsCarry: true,
        Immediate: true,
    }


    res, setCarry, carry := utils.Shift(shiftArgs)

    if setCarry {
        reg.CPSR.SetFlag(FLAG_C, carry)
    }

    //cpu.Reg.CPSR.SetFlag(FLAG_V, v)
    cpu.Reg.CPSR.SetFlag(FLAG_N, utils.BitEnabled(uint32(res), 31))
    cpu.Reg.CPSR.SetFlag(FLAG_Z, uint32(res) == 0)

    r[rd] = res

    r[PC] += 2
}

func (cpu *Cpu) thumbStack(opcode uint16) {

    r := &cpu.Reg.R
    nn := utils.GetVarData(uint32(opcode), 0, 6) * 4
    sub := utils.BitEnabled(uint32(opcode), 7)

    if sub {
        r[SP] -= nn
    } else {
        r[SP] += nn
    }

    r[PC] += 2
}

func (cpu *Cpu) thumbLongBranch(opcode uint16) {


    r := &cpu.Reg.R

    op2 := cpu.Gba.Mem.Read16(r[PC] + 2)


    upper := utils.GetVarData(uint32(opcode), 0, 10)
    lower := utils.GetVarData(uint32(op2), 0, 10)
    exchange := utils.GetVarData(uint32(op2), 11, 15) == 0b11101

    if exchange {
        panic("BLX LONG BRANCH WITH LINK NOT SUPPORTED")
    }

    mask := uint32(0b111_1111_1111_1111_1111_1111)
    nn := (upper << 12) | (lower << 1)
    negative := (nn >> 22) == 1
    nn &= mask

    if negative {
        nn |= ^mask
    }

    offset := int32(nn)

    r[LR] = utils.HalfAlign(r[PC] + 4) + 1
    r[PC] = uint32(int32(r[PC]) + 4 + offset)
}

func (cpu *Cpu) thumbShortLongBranch(opcode uint16) {
    // Using only the 2nd half of BL as "BL LR+imm" is possible
    // (for example, Mario Golf Advance Tour for GBA uses opcode F800h as "BL LR+0").
    // BL LR + nn

    r := &cpu.Reg.R

    lower := utils.GetVarData(uint32(opcode), 0, 10)

    mask := uint32(0b111_1111_1111_1111_1111_1111)
    nn := (lower << 1)
    negative := (nn >> 22) == 1
    nn &= mask

    if negative {
        nn |= ^mask
    }

    offset := int32(nn)

    tmpLR := r[LR]
    r[LR] = utils.HalfAlign(r[PC] + 2) + 1
    r[PC] = utils.HalfAlign(uint32(int32(tmpLR) + offset))
}

func (cpu *Cpu) thumbLSSP(opcode uint16) {

    r := &cpu.Reg.R

    rd := utils.GetVarData(uint32(opcode), 8, 10)
    nn := utils.GetVarData(uint32(opcode), 0, 7) << 2
    ldr := utils.BitEnabled(uint32(opcode), 11)

    addr := r[SP] + nn

    if ldr {
        v := cpu.Gba.Mem.Read32(utils.WordAlign(addr))
        is := (addr & 0b11) * 8
        r[rd], _, _ = utils.Ror(v, is, false, false ,false)
    } else {
        cpu.Gba.Mem.Write32(addr, r[rd])
    }

    r[PC] += 2
}

func (cpu *Cpu) thumbMulti(opcode uint16) {

    r := &cpu.Reg.R

    ldmia := utils.BitEnabled(uint32(opcode), 11)
    rb := utils.GetVarData(uint32(opcode), 8, 10)
    rlist := utils.GetVarData(uint32(opcode), 0, 7)
    addr := utils.WordAlign(r[rb])

    if !ldmia {

        regCount := utils.CountBits(rlist)
        matchingValue := uint32(0)
        matchingAddr := uint32(0) // rn during regs
        smallest := (rlist & -rlist) == 1 << rb
        matchingRb := (rlist >> rb) & 1 == 1

        rbIdx := uint32(0)
        count := uint32(0)

        if rlist == 0 {
            cpu.Gba.Mem.Write32(r[rb], r[PC]+6)
            r[rb] += 0x40
            r[PC] += 2
            return
        }

        for reg := range 8 {
            if notEnabled := !utils.BitEnabled(rlist, uint8(reg)); notEnabled {
                continue
            }

            if reg == int(rb) {
                cpu.Gba.Mem.Write32(addr, r[reg])
                matchingValue = r[reg] + 4
                matchingAddr = addr
                rbIdx = regCount - count
                r[rb] += 4
                addr += 4
                continue
            }

            cpu.Gba.Mem.Write32(addr, r[reg])

            r[rb] += 4
            addr += 4
        }

        if smallest {
            v := cpu.Gba.Mem.Read32(addr)
            cpu.Gba.Mem.Write32(r[rb], v - (regCount * 2))
            r[PC] += 2
            return
        }

        if matchingRb {
            cpu.Gba.Mem.Write32(matchingAddr, matchingValue + (rbIdx * 2))
            r[PC] += 2
            return
        }

        r[PC] += 2
        return
    }

    rbValue := r[rb]
    matchingRb := false

    if rlist == 0 {
        r[rb] += 0x40
        r[PC] += 8
        return
    }

    for reg := range 8 {
        if notEnabled := !utils.BitEnabled(rlist, uint8(reg)); notEnabled {
            continue
        }

        r[reg] = cpu.Gba.Mem.Read32(addr)

        if reg == int(rb) {
            matchingRb = true
            // do not remove this, needed for golden sun and others
            rbValue = r[rb]
        }

        r[rb] += 4
        addr += 4
    }

    if matchingRb {
        r[rb] = rbValue
    }

    r[PC] += 2
}
