package gba

import (
	"fmt"
	"github.com/aabalke33/guac/emu/gba/utils"
)

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

type Alu struct {
    Opcode, Rd, Rn, RnValue, Cond, Op2, Inst uint32
    Immediate, Set bool
}

func NewAluData(opcode uint32, cpu *Cpu) *Alu {

    alu := &Alu{
        Opcode: opcode,
        Cond: utils.GetByte(opcode, 28),
        Inst: utils.GetByte(opcode, 21),
        Immediate: utils.BitEnabled(opcode, 25),
        Set: utils.BitEnabled(opcode, 20),
        Rd: utils.GetByte(opcode, 12),
        Rn: utils.GetByte(opcode, 16),
    }

    alu.Op2 = cpu.GetOp2(opcode)

    if alu.Rn != PC {
        alu.RnValue = cpu.Reg.R[alu.Rn]
        return alu
    }
    
    shiftImmediate := alu.Immediate || !utils.BitEnabled(opcode, 4)
    if shiftImmediate {
        alu.RnValue = cpu.Reg.R[PC] + 8
        return alu
    }

    alu.RnValue = cpu.Reg.R[PC] + 12

    return alu
}

func (cpu *Cpu) Alu(opcode uint32) {

    alu := NewAluData(opcode, cpu)

    switch alu.Inst {
    case AND, EOR, ORR, MOV, MVN, BIC: cpu.logical(alu)
    case ADD, ADC, SUB, SBC, RSB, RSC: cpu.arithmetic(alu)
    case TST, TEQ, CMP, CMN:           cpu.test(alu)
    }

    if alu.Rd != PC {
        cpu.Reg.R[15] += 4
    }
}

func (cpu *Cpu) GetOp2(opcode uint32) uint32 {

    reg := &cpu.Reg

    isCarry := utils.BitEnabled(opcode, 20)
    currCarry := reg.CPSR.GetFlag(FLAG_C)

    if immediate := utils.BitEnabled(opcode, 25); immediate {

        nn := utils.GetVarData(opcode, 0, 7)
        ro := utils.GetVarData(opcode, 8, 11) * 2
        op2, setCarry, carry := utils.Ror(nn, ro, isCarry, false, currCarry)

        if setCarry {
            reg.CPSR.SetFlag(FLAG_C, carry)
        }

        return op2
    }

    is := utils.GetVarData(opcode, 7, 11)
    var additional uint32
    rm := utils.GetByte(opcode, 0)
    if rm == PC {
        additional += 8
    }

    shiftRegister := utils.BitEnabled(opcode, 4)
    if shiftRegister {
        // timer increase 1
        is = reg.R[(opcode>>8)&0b1111] & 0b1111_1111

    }

    shiftArgs := utils.ShiftArgs{
        SType: opcode >> 5 & 0b11,
        Val: reg.R[rm] + additional,
        Is: is,
        IsCarry: isCarry,
        Immediate: !shiftRegister,
        CurrCarry: currCarry,
    }

    op2, setCarry, carry := utils.Shift(shiftArgs)

    if setCarry {
        reg.CPSR.SetFlag(FLAG_C, carry)
    }

    return op2
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

    if CURR_INST == MAX_COUNT {
        printer(map[string]any{"OP2": alu.Op2})
    }

    cpu.Reg.R[alu.Rd] = res

    cpu.setLogicalFlags(alu, res)
}

func (cpu *Cpu) setLogicalFlags(alu *Alu, res uint32) {
    switch {
    case alu.Rd == PC && alu.Set: // private mode
        panic("Unhandled logic")
        return
    //case alu.Rd == PC && !test: pipelining
    case alu.Set:
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
    case RSC: oper = func(u1, u2, u3 uint64) uint64 { return u2 - u1 + u3 - 1 }
    }

    carry := uint64(0)
    if cpu.Reg.CPSR.GetFlag(FLAG_C) {
        carry = 1
    }

    var res uint64

    res = oper(uint64(alu.RnValue), uint64(alu.Op2), carry)
    cpu.Reg.R[alu.Rd] = uint32(res)

    cpu.setArithmeticFlags(alu, res)
}

func (cpu *Cpu) setArithmeticFlags(alu *Alu, res uint64) {
    switch {
    case alu.Rd == PC && alu.Set: // private mode
        panic("Unhandled arith")
        return
    //case alu.Rd == PC && !test:
    case alu.Rd == PC:
        panic("pipelining")
    case alu.Set:
        var v, c bool
        rnSign := uint8(alu.RnValue >> 31) & 1
        opSign := uint8(alu.Op2 >> 31) & 1
        rSign  := uint8(res >> 31) & 1

        switch alu.Inst {
        case ADD, ADC, CMN:
            v = (rnSign == opSign) && (rSign != rnSign)
            c = res >= 0x1_0000_0000
        case SUB, SBC, CMP:
            v = (rnSign != opSign) && (rSign != rnSign)
            c = res < 0x1_0000_0000

        case RSB, RSC:
            v = (rnSign != opSign) && (rSign != opSign)
            c = res < 0x1_0000_0000
        }

        cpu.Reg.CPSR.SetFlag(FLAG_V, v)
        cpu.Reg.CPSR.SetFlag(FLAG_C, c)
        cpu.Reg.CPSR.SetFlag(FLAG_N, utils.BitEnabled(uint32(res), 31))
        cpu.Reg.CPSR.SetFlag(FLAG_Z, uint32(res) == 0)
        return
    }
}

func (cpu *Cpu) test(alu *Alu) {

    var oper func (uint64, uint64) uint64

    switch alu.Inst {
    case TST: oper = func(u1, u2 uint64) uint64 { return u1 & u2 }
    case TEQ: oper = func(u1, u2 uint64) uint64 { return u1 ^ u2 }
    case CMP: oper = func(u1, u2 uint64) uint64 { return u1 - u2 }
    case CMN: oper = func(u1, u2 uint64) uint64 { return u1 + u2 }
    }

    res := oper(uint64(alu.RnValue), uint64(alu.Op2))

    switch alu.Inst {
    case TST, TEQ: cpu.setLogicalFlags(alu, uint32(res))
    case CMP, CMN:
        cpu.setArithmeticFlags(alu, res)
    }
}

const (
    MUL = 0b0
    MLA = 0b1
    UMAAL = 0b010
    UMULL = 0b100
    UMLAL = 0b101
    SMULL = 0b110
    SMLAL = 0b111
)

func (cpu *Cpu) Mul(opcode uint32) {

    inst := utils.GetByte(opcode, 21)
    set := utils.BitEnabled(opcode, 20)
    rd := utils.GetByte(opcode, 16)
    rn := utils.GetByte(opcode, 12)
    rs := utils.GetByte(opcode, 8)
    rm := utils.GetByte(opcode, 0)
    r := &cpu.Reg.R

    if mulHalf := inst == MUL || inst == MLA; mulHalf {

        res := r[rd] * r[rs]

        if inst == MLA {
            res += r[rn]
        }

        r[rd] = res

        if set {
            cpu.Reg.CPSR.SetFlag(FLAG_Z, res == 0)
            cpu.Reg.CPSR.SetFlag(FLAG_N, (res >> 31 & 0b1) != 0)
            // FLAG_C "destroyed" ARM <5, ignored ARM >=5
            cpu.Reg.CPSR.SetFlag(FLAG_C, false)
        }

        r[PC] += 4
        return
    }

    if inst == UMAAL {
        panic("UMAAL is UNSUPPORTED")
    }

    if mulUnsignedWord := inst == UMULL || inst == UMLAL; mulUnsignedWord {
        res := uint64(r[rm]) * uint64(r[rs])

        if inst == UMLAL {
            res += uint64(r[rd])<<32 | uint64(r[rn])
        }

        r[rd] = uint32(res >> 32)
        r[rn] = uint32(res)

        if set {
            cpu.Reg.CPSR.SetFlag(FLAG_Z, res == 0)
            cpu.Reg.CPSR.SetFlag(FLAG_N, (res >> 63 & 0b1) != 0)
            // FLAG_C "destroyed" ARM <5, ignored ARM >=5
            cpu.Reg.CPSR.SetFlag(FLAG_C, false)
            // FLAG_V maybe destroyed on ARM <5. ignored ARM <=5
        }

        r[PC] += 4
        return
    }

    res := int64(int32(r[rm])) * int64(int32(r[rs]))
    if inst == SMLAL {
        res += int64(r[rd])<<32 | int64(r[rn])
    }

    r[rd] = uint32(res >> 32)
    r[rn] = uint32(res)

    if set {
        cpu.Reg.CPSR.SetFlag(FLAG_Z, res == 0)
        cpu.Reg.CPSR.SetFlag(FLAG_N, (res >> 63 & 0b1) != 0)
        // FLAG_C "destroyed" ARM <5, ignored ARM >=5
        cpu.Reg.CPSR.SetFlag(FLAG_C, false)
        // FLAG_V maybe destroyed on ARM <5. ignored ARM <=5
    }

    r[PC] += 4
}

const (
    STR = iota
    LDR_PLD
)

type Sdt struct {
    Opcode, Rd, Rn, RnValue, RdValue, Cond, Offset, Shift, ShiftType, Rm uint32
    Set, Immediate, Load, WriteBack, MemoryMgmt, Pre, Up, Byte, Pld bool
}

func NewSdtData(opcode uint32, cpu *Cpu) *Sdt {
    valid := utils.GetVarData(opcode, 26, 27) == 0b01
    if !valid {
        panic("Malformed Sdt Instruction")
    }

    sdt := Sdt{
        Opcode: opcode,
        Cond: utils.GetByte(opcode, 28),
        Immediate: !utils.BitEnabled(opcode, 25),
        Pre: utils.BitEnabled(opcode, 24),
        Up: utils.BitEnabled(opcode, 23),
        Byte: utils.BitEnabled(opcode, 22),
        Load: utils.BitEnabled(opcode, 20),
        Rn: utils.GetByte(opcode, 16),
        Rd: utils.GetByte(opcode, 12),
    }

    if sdt.Pre {
        sdt.WriteBack = utils.BitEnabled(opcode, 21)
    } else {
        sdt.MemoryMgmt = utils.BitEnabled(opcode, 21)
    }

    if sdt.Immediate {
        sdt.Offset = utils.GetVarData(opcode, 0, 11)
    } else {
        sdt.Shift = utils.GetVarData(opcode, 7, 11)
        sdt.ShiftType = utils.GetVarData(opcode, 5, 6)

        if utils.BitEnabled(opcode, 4) {
            panic("Malformed Single Data Transfer")
        }

        sdt.Rm = utils.GetByte(opcode, 0)
    }

    sdt.RdValue = cpu.Reg.R[sdt.Rd]
    sdt.RnValue = cpu.Reg.R[sdt.Rn]

    sdt.Pld = sdt.Cond == 0b1111 &&
            sdt.Pre == true &&
            sdt.Byte == true &&
            sdt.WriteBack == false &&
            sdt.Load == true &&
            sdt.Rd == 0b1111

    return &sdt
}

func (c *Cpu) Sdt(opcode uint32) {

    r := &c.Reg.R

    sdt := NewSdtData(opcode, c)

    pre, post, _ := generateSdtAddress(sdt, c)

    if sdt.Pld {
        panic("Need to handle PLD Inst")
    }



    switch {
    case sdt.Load && sdt.Byte:
        r[sdt.Rd] = uint32(c.Gba.Mem.Read8(pre))
    case sdt.Load && !sdt.Byte:
        r[sdt.Rd] = c.Gba.Mem.Read32(pre)
    case !sdt.Load && sdt.Byte:
        c.Gba.Mem.Write8(pre, uint8(r[sdt.Rd]))
    case !sdt.Load && !sdt.Byte:
        c.Gba.Mem.Write32(pre, r[sdt.Rd])
    }

    if (sdt.WriteBack || !sdt.Pre) && sdt.Rn != sdt.Rd {
        r[sdt.Rn] = post
    }

    c.Reg.R[PC] += 4
}

func generateSdtAddress(sdt *Sdt, cpu *Cpu) (pre uint32, post uint32, writeBack bool) {

    r := &cpu.Reg.R

    var offset uint32
    if sdt.Immediate {
        offset = sdt.Offset
    } else {
        shift := sdt.Opcode >> 7 & 0b11111 // I = 1 shift reg
        offset = cpu.RegShiftOffset(sdt.Opcode, shift)
    }

    addr := r[sdt.Rn]
    if sdt.Up {
        addr += offset
    } else {
        addr -= offset
    }

    if sdt.Pre {
        switch {
        case offset == 0: return r[sdt.Rn], 0, false
        case sdt.Immediate: return r[sdt.Rn], 0, true
        default: return r[sdt.Rn], 0, true
        }
    }

    switch {
    case sdt.Immediate: return r[sdt.Rn], addr, false
    default: return r[sdt.Rn], addr, false
    }
}

func (cpu *Cpu) RegShiftOffset(opcode uint32, shift uint32) uint32 {

    if !utils.BitEnabled(opcode, 25) {
        return opcode & 0b1111_1111_1111
    }

    reg := &cpu.Reg
	rm := opcode & 0b1111

    shiftArgs := utils.ShiftArgs{
        SType: opcode >> 5 & 0b11,
        Val: reg.R[rm],
        Is: shift,
        IsCarry: false,
        Immediate: true,
        CurrCarry: reg.CPSR.GetFlag(FLAG_C),
    }

    ofs, setCarry, carry := utils.Shift(shiftArgs)

    if setCarry {
        reg.CPSR.SetFlag(FLAG_C, carry)
    }

    return ofs
}

const (
    SWP = iota
)

type Sds struct {
    Cond, Opcode, Inst, Rn, Rd, Rm uint32
    Byte bool
}

func NewSds(opcode uint32, cpu *Cpu) *Sds {

    sds := &Sds{
        Cond: utils.GetByte(opcode, 28),
        Opcode: opcode,
        Inst: utils.GetByte(opcode, 23),
        Byte: utils.BitEnabled(opcode, 22),
        Rn: utils.GetByte(opcode, 16),
        Rd: utils.GetByte(opcode, 12),
        Rm: utils.GetByte(opcode, 0),
    }

    valid := utils.GetVarData(opcode, 23, 27) == 0b00010 &&
             utils.GetVarData(opcode, 4, 11) == 0b00001001

    if !valid {
        panic("Malformed SDS (SWP) Instruction")
    }

    return sds
}

func (cpu *Cpu) Sds(opcode uint32) {

    r := &cpu.Reg.R

    sds := NewSds(opcode, cpu)

    if sds.Inst != SWP {
        panic("Malformed SDS Instruction. Is not SWP.")
    }


    r[sds.Rd] = r[sds.Rn]
    r[sds.Rn] = r[sds.Rm]

}

func (cpu *Cpu) B(opcode uint32) {

    isLink := utils.BitEnabled(opcode, 24)

    r := &cpu.Reg.R

    if isLink {
        r[14] = r[15] + 4
    }

    r[PC] += uint32((int32(opcode) << 8) >> 6) + 8
}

const (
    INST_BX = 1 + iota
    INST_BXJ
    INST_BLX
)

func (c *Cpu) BX(opcode uint32) {

    inst := utils.GetByte(opcode, 4)
    rn := utils.GetByte(opcode, 0)

    switch inst {
    case INST_BX:
        c.Reg.R[PC] = c.Reg.R[rn]
        c.Gba.toggleThumb()
    case INST_BXJ: panic("Unsupported BXJ Instruction")
    case INST_BLX: panic("Unsupported BLX Instruction")
    }
}

const (
    RESERVED = 0
    STRH = 1
    LDRD = 2
    STRD = 3

    LDRH =  1
    LDRSB = 2
    LDRSH = 3
)

type Half struct {
    Cond, Rn, Rd, Offset, Inst, Rm, RdValue, RnValue, RmValue uint32
    Pre, Up, Immediate, WriteBack, Load, MemoryManagement bool
}

func NewHalf(opcode uint32, c *Cpu) *Half {

    r := &c.Reg.R

    h := &Half{
        Cond: utils.GetByte(opcode, 28),
        Rn: utils.GetByte(opcode, 16),
        Rd: utils.GetByte(opcode, 12),
        Pre: utils.BitEnabled(opcode, 24),
        Up: utils.BitEnabled(opcode, 23),
        Immediate: utils.BitEnabled(opcode, 22),
        Load: utils.BitEnabled(opcode, 20),
        Inst: utils.GetVarData(opcode, 5, 6),
    }

    if h.Pre {
        h.WriteBack = utils.BitEnabled(opcode, 21)
    } else {
        h.WriteBack = true
    }

    fails := []bool{
        !h.Pre && utils.BitEnabled(opcode, 21),
        !utils.BitEnabled(opcode, 7),
        !utils.BitEnabled(opcode, 4),
        h.Immediate && !(utils.GetByte(opcode, 8) == 0b0000),
    }

    for i, fail := range fails {
        if fail { panic(fmt.Sprintf("Malformed Half Instruction %d", i)) }
    }

    if h.Immediate {
        h.Rm = utils.GetByte(opcode, 0)
        h.RmValue = r[h.Rm]
    } else {
        h.Offset = utils.GetByte(opcode, 8) << 4 | utils.GetByte(opcode, 0)
    }


    h.RnValue = r[h.Rn]
    if h.Rn == PC {
        h.RnValue += 8
    }

    h.RdValue = r[h.Rd]
    if h.Rd == PC {
        h.RdValue += 12
    }

    return h
}

func (c *Cpu) Half(opcode uint32) {

    r := &c.Reg.R
    half := NewHalf(opcode, c)
    //pre, post, writeBack := generateAddress(half, c)
    pre, _, _ := generateAddress(half, c)

    if !half.Load {
        switch half.Inst {
        case RESERVED: panic("RESERVED HALF (Load) NOT SUPPORTED")
        case STRH: c.Gba.Mem.Write16(pre, uint16(half.RdValue))
        case LDRD: panic("LDRD NOT SUPPORTED")
        case STRD: panic("STRD NOT SUPPORTED")
        }

        c.Reg.R[15] += 4
        return
    }

    switch half.Inst {
    case RESERVED: panic("RESERVED HALF (Store) NOT SUPPORTED")
    case LDRH: // unsigned half
        r[half.Rd] = c.Gba.Mem.Read16(r[half.Rn])
    case LDRSB: // signed byte
        r[half.Rd] = uint32(int8(uint8(c.Gba.Mem.Read8(r[half.Rn]))))
    case LDRSH: // signed half
        r[half.Rd] = uint32(int16(c.Gba.Mem.Read16(r[half.Rn])))
    }

    c.Reg.R[15] += 4
}

func generateAddress(half *Half, cpu *Cpu) (pre uint32, post int32, writeBack bool) {

    r := &cpu.Reg.R

    var offset int32 = int32(half.Offset)
    var rm int32 = int32(half.Rm)
    if half.Up {
        offset = int32(half.Offset) * -1
        rm = int32(half.Rm) * -1
    }

    if half.Pre {
        switch {
        case offset == 0: return r[half.Rn], 0, false
        case half.Immediate: return uint32(int32(r[half.Rn]) + offset), 0, true
        default: return uint32(int32(r[half.Rn]) + rm), 0, true
        }
    }

    switch {
    case half.Immediate: return r[half.Rn], offset, false
    default: return r[half.Rn], rm, false
    }

    // writeback???
}

type Block struct {
    Opcode, Cond, Rn, RnValue, Rlist uint32
    Pre, Up, PSR, Writeback, Load bool
}

func NewBlock(opcode uint32, c *Cpu) *Block {

    b := &Block{
        Opcode: opcode,
        Cond: utils.GetByte(opcode, 28),
        Pre: utils.BitEnabled(opcode, 24),
        Up: utils.BitEnabled(opcode, 23),
        PSR: utils.BitEnabled(opcode, 22),
        Writeback: utils.BitEnabled(opcode, 21),
        Load: utils.BitEnabled(opcode, 20),
        Rn: utils.GetByte(opcode, 16),
        Rlist: utils.GetVarData(opcode, 0, 15),
    }

    if utils.GetVarData(opcode, 25, 27) != 0b100 {
        panic("Malformed Block Instruction")
    }


    b.RnValue = c.Reg.R[b.Rn]

    return b
}

func (c *Cpu) Block(opcode uint32) {

    r := &c.Reg.R

    block := NewBlock(opcode, c)

    incPc := true

    mode := c.Reg.getMode()

    if forceUser := block.PSR; forceUser {
        c.Reg.setMode(MODE_USR)
    }

    if block.Load {
        incPc = c.ldm(block)
    } else {
        c.stm(block)
    }

    if block.Pre && !block.Writeback {
        r[block.Rn] = block.RnValue
    }

    if forceUser := block.PSR; forceUser {
        c.Reg.setMode(mode)
    }

    if incPc {
        c.Reg.R[PC] += 4
    }
}

func (c *Cpu) ldm(block *Block) bool {

    incPC := true

    r := &c.Reg.R

    for reg := range 16 {

        regBitEnabled := utils.BitEnabled(block.Rlist, uint8(reg))
        decRegBitEnabled := utils.BitEnabled(block.Rlist, uint8(15 - reg))

        switch {
        case block.Pre  &&  block.Up && regBitEnabled:

            r[block.Rn] += 4
            r[reg] = c.Gba.Mem.Read32(r[block.Rn])

            if reg == PC {
                incPC = false
            }

        case !block.Pre &&  block.Up && regBitEnabled:

            r[reg] = c.Gba.Mem.Read32(r[block.Rn])

            if CURR_INST == MAX_COUNT {
                printer(map[string]any{"REG": reg, "R": r[reg]})
            }

            r[block.Rn] += 4

            if reg == PC {
                incPC = false
            }

        case block.Pre  &&  !block.Up && decRegBitEnabled: // pop

            r[block.Rn] -= 4
            r[15 - reg] = c.Gba.Mem.Read32(r[block.Rn])

            if 15 - reg == PC {
                incPC = false
            }

        case !block.Pre &&  !block.Up && decRegBitEnabled:

            r[15 - reg] = c.Gba.Mem.Read32(r[block.Rn])
            r[block.Rn] -= 4

            if 15 - reg == PC {
                incPC = false
            }

        }
    }

    return incPC
}

func (c *Cpu) stm(block *Block) {
    r := &c.Reg.R

    for reg := range 16 {

        regBitEnabled := utils.BitEnabled(block.Rlist, uint8(reg))
        decRegBitEnabled := utils.BitEnabled(block.Rlist, uint8(15 - reg))

        switch {
        case block.Pre  &&  block.Up && regBitEnabled:

            r[block.Rn] += 4
            c.Gba.Mem.Write32(r[block.Rn], r[reg])

        case !block.Pre &&  block.Up && regBitEnabled:

            c.Gba.Mem.Write32(r[block.Rn], r[reg])
            r[block.Rn] += 4

        case block.Pre  &&  !block.Up && decRegBitEnabled: // push

            r[block.Rn] -= 4
            c.Gba.Mem.Write32(r[block.Rn], r[15 - reg])

        case !block.Pre &&  !block.Up && decRegBitEnabled:

            c.Gba.Mem.Write32(r[block.Rn], r[15 - reg])
            r[block.Rn] -= 4
        }
    }
}

type PSR struct {
    Opcode, Cond, Rd, Rm, Shift, Imm uint32
    SPSR, MSR, Immediate, F, S, X, C bool 
}

func NewPSR(opcode uint32, cpu *Cpu) *PSR {

    psr := &PSR{
        Opcode: opcode,
        Cond: utils.GetByte(opcode, 28),
        Immediate: utils.BitEnabled(opcode, 25),
        SPSR: utils.BitEnabled(opcode, 22),
        MSR: utils.BitEnabled(opcode, 21),
    }

    if !psr.MSR {
        psr.Rd = utils.GetByte(opcode, 12)
        return psr
    }

    if psr.MSR {
        psr.F = utils.BitEnabled(opcode, 19)
        psr.S = utils.BitEnabled(opcode, 18)
        psr.X = utils.BitEnabled(opcode, 17)
        psr.C = utils.BitEnabled(opcode, 16)
    }

    if psr.Immediate {
        psr.Shift = utils.GetByte(opcode, 8) * 2
        psr.Imm = utils.GetVarData(opcode, 0, 7)
        return psr
    }

    psr.Rm = utils.GetByte(opcode, 0)


    return psr
}

func (cpu *Cpu) Psr(opcode uint32) {

    psr := NewPSR(opcode, cpu)

    if psr.MSR {
        cpu.msr(psr)
        cpu.Reg.R[15] += 4
        return
    }

    cpu.mrs(psr)
    cpu.Reg.R[15] += 4
}

func (cpu *Cpu) mrs(psr *PSR) {

    r := &cpu.Reg.R

    if psr.SPSR {
        mode := cpu.Reg.getMode()
        r[psr.Rd] = uint32(cpu.Reg.SPSR[BANK_ID[mode]])
        return
    }

    // masks

    r[psr.Rd] = uint32(cpu.Reg.CPSR)
}

func (cpu *Cpu) msr(psr *PSR) {

    reg := &cpu.Reg
    r := &cpu.Reg.R

    p := &cpu.Reg.CPSR

    currMode := cpu.Reg.getMode()

    if psr.SPSR {
        mode := cpu.Reg.getMode()
        p = &cpu.Reg.SPSR[BANK_ID[mode]]
    }


    if psr.Immediate {

        // assumes user mode
        imm, _, _:= utils.Ror(psr.Imm, psr.Shift, false, false, false)

        if psr.SPSR {
            //spsr := cpu.Reg.SPSR[BANK_ID[currMode]]

            // set spsr
            panic("NEED SET SPRS LOGIC IN MSR")
            return
        }

        if psr.C { *p = Cond((uint32(*p) &^ 0x000000FF) | (imm & 0x000000FF)) }
        if psr.F { *p = Cond((uint32(*p) &^ 0xF0000000) | (imm & 0xF0000000)) }
        if psr.X { *p = Cond((uint32(*p) &^ 0x0FF00000) | (imm & 0x0FF00000)) }
        if psr.S { *p = Cond((uint32(*p) &^ 0x000FFF00) | (imm & 0x000FFF00)) }

        mode := cpu.Reg.getMode()

        if BANK_ID[currMode] == BANK_ID[mode] {
            return
        }

        reg.SP[BANK_ID[currMode]] = reg.R[SP]
        reg.LR[BANK_ID[currMode]] = reg.R[LR]

        reg.R[SP] = reg.SP[BANK_ID[mode]]
        reg.R[LR] = reg.LR[BANK_ID[mode]]

        // check irq
        return
    }

    // unsure on this part
    if psr.F { p.SetField(24, (r[psr.Rm] >> 24) & 0xFF) }
    if psr.S { p.SetField(16, (r[psr.Rm] >> 16) & 0xFF) }
    if psr.X { p.SetField(8, (r[psr.Rm] >> 8) & 0xFF) }
    if psr.C { p.SetField(0, r[psr.Rm] & 0xFF) }
}

func (cpu *Cpu) Swp(opcode uint32) {

    isByte := utils.BitEnabled(opcode, 22)
    rn := utils.GetByte(opcode, 16)
    rd := utils.GetByte(opcode, 12)
    rm := utils.GetByte(opcode, 0)

    r := &cpu.Reg.R

    rmValue := r[rm]
    rnValue := r[rn]

    var rnMemValue uint32
    if isByte {
        rnMemValue = cpu.Gba.Mem.Read8(rnValue)
    } else {
        rnMemValue = cpu.Gba.Mem.Read32(rnValue)
    }

    r[rd] = rnMemValue
    cpu.Gba.Mem.Write32(rnValue, rmValue)

    cpu.Reg.R[15] += 4
}
