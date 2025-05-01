package gba

import (

	"github.com/aabalke33/guac/emu/gba/utils"
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

func NewThumbAlu(opcode uint16, cpu *Cpu) *ThumbAlu {

	alu := &ThumbAlu{
		Opcode: opcode,
		Inst: uint16(utils.GetByte(uint32(opcode), 6)),
        Rs: uint16(utils.GetVarData(uint32(opcode), 3, 5)),
        Rd: uint16(utils.GetVarData(uint32(opcode), 0, 2)),
	}

    return alu
}

func (cpu *Cpu) ThumbAlu(opcode uint16) {

    alu := NewThumbAlu(opcode, cpu)

    switch alu.Inst {
    case THUMB_LSL, THUMB_LSR, THUMB_ASR, THUMB_ADC, THUMB_SBC, THUMB_ROR, THUMB_NEG: cpu.thumbArithmetic(alu)
    case THUMB_AND, THUMB_EOR, THUMB_ORR, THUMB_BIC, THUMB_MVN: cpu.thumbLogical(alu)
    case THUMB_TST, THUMB_CMN, THUMB_CMP: cpu.thumbTest(alu)
    case THUMB_MUL: cpu.thumbMuliply(alu)
    }

    cpu.Reg.R[15] += 2
}

func (cpu *Cpu) thumbMuliply(alu *ThumbAlu) {

    r := &cpu.Reg.R

    res := uint64(r[alu.Rd]) * uint64(r[alu.Rs])

    r[alu.Rd] = uint32(res)

    // ARM < 4, carry flag destroyed, ARM >= 5, carry flag unchanged
    cpu.Reg.CPSR.SetFlag(FLAG_C, false)

    cpu.Reg.CPSR.SetFlag(FLAG_N, utils.BitEnabled(uint32(res), 31))
    cpu.Reg.CPSR.SetFlag(FLAG_Z, uint32(res) == 0)
}

func (cpu *Cpu) thumbLogical(alu *ThumbAlu) {

    r := &cpu.Reg.R

    var oper func (uint32, uint32) uint32

    switch alu.Inst {
    case THUMB_AND: oper = func(u1, u2 uint32) uint32 { return u1 & u2 }
    case THUMB_EOR: oper = func(u1, u2 uint32) uint32 { return u1 ^ u2 }
    case THUMB_ORR: oper = func(u1, u2 uint32) uint32 { return u1 | u2 }
    case THUMB_BIC: oper = func(u1, u2 uint32) uint32 { return u1 &^ u2 }
    case THUMB_MVN: oper = func(_, u2 uint32) uint32 { return ^u2 }
    }

    res := oper(r[alu.Rd], r[alu.Rs])

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
        c = uint64(res) >= 0x1_0000_0000
    case THUMB_SBC, THUMB_NEG:
        v = (rdSign != rsSign) && (rSign != rdSign)
        c = uint64(res) < 0x1_0000_0000
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

    var oper func (uint64, uint64) uint64

    switch alu.Inst {
    case THUMB_TST: oper = func(u1, u2 uint64) uint64 { return u1 & u2 }
    case THUMB_CMP: oper = func(u1, u2 uint64) uint64 { return u1 - u2 }
    case THUMB_CMN: oper = func(u1, u2 uint64) uint64 { return u1 + u2 }
    }

    rdValue := uint64(r[alu.Rd])

    res := oper(rdValue, uint64(r[alu.Rs]))

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
        c := res >= 0x1_0000_0000 // HERE MIGHT BE eRROR
        cpu.Reg.CPSR.SetFlag(FLAG_V, v)
        cpu.Reg.CPSR.SetFlag(FLAG_C, c)
    }

    cpu.Reg.CPSR.SetFlag(FLAG_N, utils.BitEnabled(uint32(res), 31))
    cpu.Reg.CPSR.SetFlag(FLAG_Z, uint32(res) == 0)
}


// HI REG / BX

type HiRegBX struct {
    Opcode, Inst, Rs, Rd uint16
    MSBd bool
}

func NewHiRegBX(opcode uint16, cpu *Cpu) *HiRegBX {

    hr := &HiRegBX{
        Opcode: opcode,
        Inst: uint16(utils.GetVarData(uint32(opcode), 8, 9)),
        MSBd: utils.BitEnabled(uint32(opcode), 7),
        Rs: uint16(utils.GetVarData(uint32(opcode), 3, 6)),
        Rd: uint16(utils.GetVarData(uint32(opcode), 0, 2)),
    }

    if hr.Inst != 3 && hr.MSBd {
        hr.Rd |= 0b1000
    }

    return hr
}

func (cpu *Cpu) HiRegBX(opcode uint16) {

    r := &cpu.Reg.R
    cpsr := &cpu.Reg.CPSR
    hr := NewHiRegBX(opcode, cpu)

    switch {
    case hr.Inst == 0: panic("HI ADD")
    case hr.Inst == 1: panic("HI CMP")
    case hr.Inst == 2:

        if nop := hr.Rs == 8 && hr.Rd == 8; nop {
            panic("HI NOP")
            return
        }

        r[hr.Rd] = r[hr.Rs]
        r[PC] += 2
        return

    case hr.Inst == 3 && hr.MSBd: panic("UNSUPPORTED HI BLX")
    case hr.Inst == 3:

        if !utils.BitEnabled(r[hr.Rs], 0) {
            cpsr.SetFlag(FLAG_T, false)
        }

        r[PC] = r[hr.Rs] //&^ 0b1

        if hr.Rs == PC {
            r[PC] += 4
        }
        return
    }
}

const (
    THUMB_ADD = iota
    THUMB_SUB
    THUMB_ADDMOV
    THUMB_SUBImm
)

type ThumbAddSub struct {
    Opcode, Inst, Rn, Rs, Rd, Imm uint16
}

func NewThumbAddSub(opcode uint16, cpu *Cpu) *ThumbAddSub {

    as := &ThumbAddSub{
        Opcode: uint16(opcode),
        Inst: uint16(utils.GetVarData(uint32(opcode), 9, 10)),
        Rn: uint16(utils.GetVarData(uint32(opcode), 6, 8)),
        Imm: uint16(utils.GetVarData(uint32(opcode), 6, 8)),
        Rs: uint16(utils.GetVarData(uint32(opcode), 3, 5)),
        Rd: uint16(utils.GetVarData(uint32(opcode), 0, 2)),
    }

    return as
}

func (cpu *Cpu) ThumbAddSub(opcode uint16) {

    r := &cpu.Reg.R
    cpsr := &cpu.Reg.CPSR

    as := NewThumbAddSub(opcode, cpu)

    var oper func (u1, u2, u3 uint64) uint64

    if CURR_INST == MAX_COUNT {
        printer(map[string]any{"RS": r[as.Rs], "RN": r[as.Rn]})
    }

    switch as.Inst {
    case THUMB_ADD: oper = func(u1, u2, _ uint64) uint64 {return u1 + u2}
    case THUMB_SUB: oper = func(u1, u2, _ uint64) uint64 {return u1 - u2}
    case THUMB_ADDMOV: oper = func(u1, _, u3 uint64) uint64 {return u1 + u3}
    case THUMB_SUBImm: oper = func(u1, _, u3 uint64) uint64 {return u1 - u3}
    }

    rsValue := r[as.Rs]

    res := oper(uint64(rsValue), uint64(r[as.Rn]), uint64(as.Imm))

    r[as.Rd] = uint32(res)

    var v, c bool
    rsSign := uint8(rsValue >> 31) & 1
    rnSign := uint8(uint32(r[as.Rn]) >> 31) & 1
    imSign := uint8(uint32(as.Imm) >> 31) & 1
    rSign  := uint8(int32(uint32(res)) >> 31) & 1


    switch as.Inst {
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
        v = (rsSign != imSign) && (rSign != imSign)
        c = res < 0x1_0000_0000
    }

    cpsr.SetFlag(FLAG_N, utils.BitEnabled(uint32(res), 31))
    cpsr.SetFlag(FLAG_Z, uint32(res) == 0)
    cpsr.SetFlag(FLAG_C, c)
    cpsr.SetFlag(FLAG_V, v)

    r[PC] += 2
}

type ThumbLSHalf struct {
    Opcode, Offset, Rb, Rd uint32
    Ldr bool
}

func NewThumbLSHalf(opcode uint16, cpu *Cpu) *ThumbLSHalf {
    ls := &ThumbLSHalf{
        Opcode: uint32(opcode),
        Offset: utils.GetVarData(uint32(opcode), 6, 10) * 2,
        Rb: utils.GetVarData(uint32(opcode), 3, 5),
        Rd: utils.GetVarData(uint32(opcode), 0, 2),
        Ldr: utils.BitEnabled(uint32(opcode), 11),
    }

    return ls
}

func (cpu *Cpu) thumbLSHalf(opcode uint16) {

    r := &cpu.Reg.R

    ls := NewThumbLSHalf(opcode, cpu)

    addr := r[ls.Rb] + ls.Offset

    if ls.Ldr {
        r[ls.Rd] = cpu.Gba.Mem.Read16(addr)
    } else {
        cpu.Gba.Mem.Write16(addr, uint16(r[ls.Rd]))
    }

    r[PC] += 2
}

const (
    THUMB_STRH = iota
    THUMB_LDSB
    THUMB_LDRH
    THUMB_LDSH
)

type ThumbLSSigned struct {
    Opcode, Inst, Ro, Rb, Rd uint32
}

func NewThumbLSSigned(opcode uint16, cpu *Cpu) *ThumbLSSigned {
    return &ThumbLSSigned{
        Opcode: uint32(opcode),
        Inst: utils.GetVarData(uint32(opcode), 10, 11),
        Ro: utils.GetVarData(uint32(opcode), 6, 8),
        Rb: utils.GetVarData(uint32(opcode), 3, 5),
        Rd: utils.GetVarData(uint32(opcode), 0, 2),
    }
}

func (cpu *Cpu) thumbLSSigned(opcode uint16) {

    ls := NewThumbLSSigned(opcode, cpu)
    r := &cpu.Reg.R

    addr := r[ls.Rb] + r[ls.Ro]

    switch ls.Inst {
    case THUMB_STRH:
        cpu.Gba.Mem.Write16(addr, uint16(r[ls.Rd]))

    case THUMB_LDSB:
        r[ls.Rd] = uint32(int8(uint8(cpu.Gba.Mem.Read8(addr))))
    case THUMB_LDRH:
        r[ls.Rd] = cpu.Gba.Mem.Read16(addr)
    case THUMB_LDSH:
        r[ls.Rd] = uint32(int16(cpu.Gba.Mem.Read16(addr)))
    }

    r[PC] += 2
}

func (cpu *Cpu) thumbLPC(opcode uint16) {

    r := &cpu.Reg.R

    rd := utils.GetVarData(uint32(opcode), 8, 10)
    nn := utils.GetVarData(uint32(opcode), 0, 7)
    addr := r[PC] + 2 + nn

    r[rd] = cpu.Gba.Mem.Read32(addr)

    r[PC] += 4
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
        r[rd] = cpu.Gba.Mem.Read32(addr)
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
    addr := r[rb] + nn

    switch inst {
    case THUMB_STR_IMM:
        cpu.Gba.Mem.Write32(addr, r[rd])
    case THUMB_LDR_IMM:
        r[rd] = cpu.Gba.Mem.Read32(addr)
    case THUMB_STRB_IMM:
        cpu.Gba.Mem.Write8(addr, uint8(r[rd]))
    case THUMB_LDRB_IMM:
        r[rd] = cpu.Gba.Mem.Read8(addr)
    }

    r[PC] += 2
}

// need thumb 9 imm and thumb 11 sp relative

func (cpu *Cpu) thumbPushPop(opcode uint16) {

    r := &cpu.Reg.R

    isPop := utils.BitEnabled(uint32(opcode), 11)
    pclr := utils.BitEnabled(uint32(opcode), 8)
    rlist := utils.GetVarData(uint32(opcode), 0, 7)

    if isPop {
        for reg := range 7 {
            if utils.BitEnabled(rlist, uint8(reg)) {
                r[reg] = cpu.Gba.Mem.Read32(r[13])
                r[13] += 4
            }
        }

        if pclr {
            r[PC] = cpu.Gba.Mem.Read32(r[13])
            r[13] += 4
            //piplining
        }

        r[PC] += 2

        return
    }

    if pclr {
        r[13] -= 4
        cpu.Gba.Mem.Write32(r[13], r[14])
    }

    for reg := range 7 {
        if utils.BitEnabled(rlist, uint8(7 - reg)) {
            r[13] -= 4
            cpu.Gba.Mem.Write32(r[13], r[7-reg])
        }
    }

    r[PC] += 2

}
