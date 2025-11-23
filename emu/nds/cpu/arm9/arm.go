package arm9

import (
	"fmt"
	"math"
	"math/bits"

	"github.com/aabalke/guac/emu/nds/cpu/cp15"
	"github.com/aabalke/guac/emu/nds/utils"
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

var aluInst = [16]func(cpu *Cpu, alu *Alu){

    // AND 
    func(cpu *Cpu, alu *Alu) {
        res := alu.RnValue & alu.Op2
        cpu.Reg.R[alu.Rd] = res
        cpu.logicalExit(alu, res)
    },

    // EOR
    func(cpu *Cpu, alu *Alu) {
        res := alu.RnValue ^ alu.Op2
        cpu.Reg.R[alu.Rd] = res
        cpu.logicalExit(alu, res)
    },

    // SUB
    func(cpu *Cpu, alu *Alu) {
        res := uint64(alu.RnValue) - uint64(alu.Op2)
        cpu.Reg.R[alu.Rd] = uint32(res)
        if set := utils.BitEnabled(alu.Opcode, 20); set {
            cpu.setSubFlags(alu, res)
        }
    },

    // RSB
    func(cpu *Cpu, alu *Alu) {
        res := uint64(alu.Op2) - uint64(alu.RnValue)
        cpu.Reg.R[alu.Rd] = uint32(res)
        if set := utils.BitEnabled(alu.Opcode, 20); set {
            cpu.setAluFlags(alu, res, 2)
        }
    },

    // ADD
    func(cpu *Cpu, alu *Alu) {
        res := uint64(alu.RnValue) + uint64(alu.Op2)
        cpu.Reg.R[alu.Rd] = uint32(res)
        if set := utils.BitEnabled(alu.Opcode, 20); set {
            cpu.setAluFlags(alu, res, 0)
        }
    },

    // ADC
    func(cpu *Cpu, alu *Alu) {
        carry := uint64(0)
        if alu.Carry {
            carry = 1
        }
        res := uint64(alu.RnValue) + uint64(alu.Op2) + carry
        cpu.Reg.R[alu.Rd] = uint32(res)
        if set := utils.BitEnabled(alu.Opcode, 20); set {
            cpu.setAluFlags(alu, res, 0)
        }
    },

    //SBC
    func(cpu *Cpu, alu *Alu) {
        carry := uint64(0)
        if alu.Carry {
            carry = 1
        }
        res := uint64(alu.RnValue) - uint64(alu.Op2) + carry - 1
        cpu.Reg.R[alu.Rd] = uint32(res)
        if set := utils.BitEnabled(alu.Opcode, 20); set {
            cpu.setAluFlags(alu, res, 1)
        }
    },

    // RSC
    func(cpu *Cpu, alu *Alu) {
        carry := uint64(0)
        if alu.Carry {
            carry = 1
        }
        res := uint64(alu.Op2) - uint64(alu.RnValue) + carry - 1
        cpu.Reg.R[alu.Rd] = uint32(res)
        if set := utils.BitEnabled(alu.Opcode, 20); set {
            cpu.setAluFlags(alu, res, 2)
        }
    },

    // TST
    func(cpu *Cpu, alu *Alu) {
        res := uint64(alu.RnValue) & uint64(alu.Op2)
        cpu.testExit(alu)
        if set := utils.BitEnabled(alu.Opcode, 20); set {
            cpu.Reg.CPSR.N = utils.BitEnabled(uint32(res), 31)
            cpu.Reg.CPSR.Z = uint32(res) == 0
        }
    },

    // TEQ
    func(cpu *Cpu, alu *Alu) {
        res := uint64(alu.RnValue) ^ uint64(alu.Op2)
        cpu.testExit(alu)
        if set := utils.BitEnabled(alu.Opcode, 20); set {
            cpu.Reg.CPSR.N = utils.BitEnabled(uint32(res), 31)
            cpu.Reg.CPSR.Z = uint32(res) == 0
        }
    },

    // CMP
    func(cpu *Cpu, alu *Alu) {
        res := uint64(alu.RnValue) - uint64(alu.Op2)
        if set := utils.BitEnabled(alu.Opcode, 20); set {
            cpu.setAluFlags(alu, res, 1)
        }
        cpu.testExit(alu)
    },

    // CMN
    func(cpu *Cpu, alu *Alu) {
        res := uint64(alu.RnValue) + uint64(alu.Op2)
        if set := utils.BitEnabled(alu.Opcode, 20); set {
            cpu.setAluFlags(alu, res, 0)
        }
        cpu.testExit(alu)
    },

    // ORR
    func(cpu *Cpu, alu *Alu) {
        res := alu.RnValue | alu.Op2
        cpu.Reg.R[alu.Rd] = res
        cpu.logicalExit(alu, res)
    },

    // MOV
    func(cpu *Cpu, alu *Alu) {
        res := alu.Op2
        cpu.Reg.R[alu.Rd] = res
        cpu.movExit(alu, res)
    },

    // BIC
    func(cpu *Cpu, alu *Alu) {
        res := alu.RnValue &^ alu.Op2
        cpu.Reg.R[alu.Rd] = res
        cpu.logicalExit(alu, res)
    },

    // MVN
    func(cpu *Cpu, alu *Alu) {
        res := ^alu.Op2
        cpu.Reg.R[alu.Rd] = res
        cpu.logicalExit(alu, res)
    },
}

type Alu struct {
	Opcode, Rd, RnValue, Op2       uint32
	Carry                          bool
}

var aluData Alu

func (cpu *Cpu) Alu(opcode uint32) {

	aluData.Opcode = opcode
	aluData.Rd = utils.GetByte(opcode, 12)
    aluData.Carry = cpu.Reg.CPSR.C
    rn := utils.GetByte(opcode, 16)
	aluData.RnValue = cpu.Reg.R[rn]

    if imm := (opcode>>25) & 1 == 1; imm {

        if opcode & 0xFFF == 0 {
            aluData.Op2 = 0
        } else {
            cpu.SetOp2Imm(&aluData, opcode)
        }

        if rn == PC {
            aluData.RnValue += 8
        }

    } else {

        if (opcode >> 4) & 0xFF == 0 && (opcode & 0xF != 0xF) {
            aluData.Op2 = cpu.Reg.R[opcode & 0xF]
        } else {
            cpu.SetOp2Reg(&aluData, opcode)
        }

        if rn == PC {
            if imm := !utils.BitEnabled(opcode, 4); imm {
                aluData.RnValue += 8
            } else {
                aluData.RnValue += 12
            }
        }
    }

    aluInst[utils.GetByte(opcode, 21)](cpu, &aluData)

    switch {
    case aluData.Rd != PC:
		cpu.Reg.R[15] += 4
    case cpu.Reg.CPSR.T:
        cpu.Reg.R[15] &^= 0b1
    case !cpu.Reg.CPSR.T:
        cpu.Reg.R[15] &^= 0b11
    }
}

func (cpu *Cpu) SetOp2Imm(alu *Alu, opcode uint32) {

    // Ror Special with assumptions
    nn := opcode & 0xFF
    ro := ((opcode >> 8) & 0xF) << 1
    carry := (nn>>((ro-1)&31))&0b1 > 0
    alu.Op2 = nn >> ro | (nn << (32 - ro))
    if setCarry := ro > 0 && (opcode >> 20) & 1 == 1; setCarry {
        cpu.Reg.CPSR.C = carry
    }
}

func (cpu *Cpu) SetOp2Reg(alu *Alu, opcode uint32) {

	reg := &cpu.Reg

    rm := opcode & 0xF
	var additional uint32
	if rm == PC {
		additional = 8
	}

    shiftRegister := (opcode >> 4) & 1 == 1
    is := ((opcode >> 7) & 0x1F)

	if shiftRegister {
		is = reg.R[(opcode>>8) & 0xF] & 0xFF

		if rm == PC {
			additional = 12
		}
	}

	op2, setCarry, carry := utils.ShiftFuncs[opcode >> 5 & 0b11](
		reg.R[rm] + additional,
		is,
        (opcode >> 20) & 1 == 1, // set
		!shiftRegister,
		alu.Carry,
    )

	if setCarry {
        reg.CPSR.C = carry
	}

    alu.Op2 = op2
}

func (cpu *Cpu) movExit(alu *Alu, res uint32) {

    if alu.Rd == PC {
        //cpu.toggleThumb()
        // this may be a problem still
        cpu.Reg.R[alu.Rd] &^= 0b1
    }

	if set := utils.BitEnabled(alu.Opcode, 20); !set {
		return
	}

    rm := utils.GetByte(alu.Opcode, 0) == LR && (alu.Opcode >> 25) & 1 == 0
	if swiExit := alu.Rd == PC && rm; swiExit {
		cpu.ExitException(MODE_SWI)
        if cpu.Reg.R[15] & 1 == 1 {
            cpu.toggleThumb()
        }

		return
	}

    cpu.Reg.CPSR.N = utils.BitEnabled(uint32(res), 31)
    cpu.Reg.CPSR.Z = uint32(res) == 0
}

func (cpu *Cpu) logicalExit(alu *Alu, res uint32) {

    if alu.Rd == PC {
        //cpu.toggleThumb()
        // this may be a problem still
        cpu.Reg.R[alu.Rd] &^= 0b1
    }

	if set := utils.BitEnabled(alu.Opcode, 20); !set {
		return
	}

    cpu.Reg.CPSR.N = utils.BitEnabled(uint32(res), 31)
    cpu.Reg.CPSR.Z = uint32(res) == 0
}

func (cpu *Cpu) testExit(alu *Alu) {
	if alu.Rd == PC {
		// ARM 3: Bad CMP / CMN / TST / TEQ change the mode
		// i know it isnt stored spsr, becuase in tests this was zero
		//cpu.Reg.setMode(cpu.Reg.getMode(), MODE_SYS) // this may be not SYS MODE
		// maybe spsr but with some more work???

		cpu.Reg.R[PC] += 4
		return
	}
}

func (cpu *Cpu) psrSwitch() {

	reg := &cpu.Reg
	r := &cpu.Reg.R

	// PC is updated in final bios inst

	curr := cpu.Reg.CPSR.Mode

	i := BANK_ID[curr]
	reg.CPSR = reg.SPSR[i]

	next := cpu.Reg.CPSR.Mode
	c := BANK_ID[next]

	// if you set this up for fiq, get the special registers
	reg.LR[i] = r[LR]
	reg.SP[i] = r[SP]
	r[SP] = reg.SP[c]
	r[LR] = reg.LR[c]

	if curr != MODE_FIQ {
		for i := range 5 {
			reg.USR[i] = r[8+i]
		}
	}

	reg.SP[BANK_ID[curr]] = r[SP]
	reg.LR[BANK_ID[curr]] = r[LR]

	if curr == MODE_FIQ {
		for i := range 5 {
			reg.FIQ[i] = r[8+i]
		}
	}

	if next != MODE_FIQ {
		for i := range 5 {
			r[8+i] = reg.USR[i]
		}
	}

	r[SP] = reg.SP[BANK_ID[next]]
	r[LR] = reg.LR[BANK_ID[next]]

	if next == MODE_FIQ {
		for i := range 5 {
			r[8+i] = reg.FIQ[i]
		}
	}
}

func (cpu *Cpu) setSubFlags(alu *Alu, res uint64) {

    if alu.Rd == PC {
        if rn := utils.GetByte(alu.Opcode, 16); rn == LR {
            switch cpu.Reg.CPSR.Mode {
            case MODE_ABT:
                cpu.Reg.R[15] += 4
                cpu.ExitException(MODE_ABT)
            case MODE_SWI:
                cpu.ExitException(MODE_SWI)
            default:
                cpu.ExitException(MODE_IRQ)
            }

            if cpu.Reg.R[15] & 1 == 1 {
                cpu.toggleThumb()
            }

            return
        }

        // force exit
        cpu.psrSwitch() // not sure if needed
        if cpu.Reg.R[15] & 1 == 1 {
            cpu.toggleThumb()
        }

        return
    }

	rnSign := uint8(alu.RnValue>>31) & 1
	opSign := uint8(alu.Op2>>31) & 1
	rSign := uint8(res>>31) & 1

    v := (rnSign != opSign) && (rSign != rnSign)
	c := res < 0x1_0000_0000

    cpu.Reg.CPSR.V = v
    cpu.Reg.CPSR.C = c
    cpu.Reg.CPSR.N = utils.BitEnabled(uint32(res), 31)
    cpu.Reg.CPSR.Z = uint32(res) == 0
}

func (cpu *Cpu) setAluFlags(alu *Alu, res uint64, instSet uint32) {

	var v, c bool
	rnSign := uint8(alu.RnValue>>31) & 1
	opSign := uint8(alu.Op2>>31) & 1
	rSign := uint8(res>>31) & 1

	switch instSet {
	//case ADD, ADC, CMN:
    case 0:
		v = (rnSign == opSign) && (rSign != rnSign)
		c = res >= 0x1_0000_0000
    case 1:
	//case SUB, SBC, CMP:
		v = (rnSign != opSign) && (rSign != rnSign)
		c = res < 0x1_0000_0000
    case 2:
	//case RSB, RSC:
		v = (rnSign != opSign) && (rSign != opSign)
		c = res < 0x1_0000_0000
	}

    cpu.Reg.CPSR.V = v
    cpu.Reg.CPSR.C = c
    cpu.Reg.CPSR.N = utils.BitEnabled(uint32(res), 31)
    cpu.Reg.CPSR.Z = uint32(res) == 0
}

const (
	MUL   = 0b0
	MLA   = 0b1
	UMAAL = 0b010
	UMULL = 0b100
	UMLAL = 0b101
	SMULL = 0b110
	SMLAL = 0b111

    SMLAxy = 0b1000
    SMLAWySMLALWy = 0b1001
    SMLALxy =0b1010
    SMULxy = 0b1011
)

func (cpu *Cpu) Mul(opcode uint32) {

	inst := utils.GetByte(opcode, 21)
	set := utils.BitEnabled(opcode, 20)
	rd := utils.GetByte(opcode, 16)
	rn := utils.GetByte(opcode, 12)
	rs := utils.GetByte(opcode, 8)
	rm := utils.GetByte(opcode, 0)
	r := &cpu.Reg.R

    switch inst {
    case MUL, MLA:

		res := r[rm] * r[rs]

		if inst == MLA {
			res += r[rn]
		}

		r[rd] = res

		if set {
            cpu.Reg.CPSR.N = utils.BitEnabled(uint32(res), 31)
            cpu.Reg.CPSR.Z = uint32(res) == 0
			// FLAG_C "destroyed" ARM <5, ignored ARM >=5
            //cpu.Reg.CPSR.C = false
		}

		r[PC] += 4
		return


    case UMAAL:
	    panic("UMAAL is UNSUPPORTED")
    
    case UMULL, UMLAL:

		res := uint64(r[rm]) * uint64(r[rs])

		if inst == UMLAL {
			res += uint64(r[rd])<<32 | uint64(r[rn])
		}

		r[rd] = uint32(res >> 32)
		r[rn] = uint32(res)

		if set {
            cpu.Reg.CPSR.N = (res >> 63 & 1) != 0
            cpu.Reg.CPSR.Z = res == 0
			// FLAG_C "destroyed" ARM <5, ignored ARM >=5
			// need carry to pass mgba suite
            cpu.Reg.CPSR.C = false
			// FLAG_V maybe destroyed on ARM <5. ignored ARM <=5
		}

		r[PC] += 4
		return

    case SMULL, SMLAL:

        res := int64(int32(r[rm])) * int64(int32(r[rs]))
        if inst == SMLAL {
            res += int64(r[rd])<<32 | int64(r[rn])
        }

        r[rd] = uint32(res >> 32)
        r[rn] = uint32(res)

        if set {
            cpu.Reg.CPSR.N = (res >> 63) & 1 == 1
            cpu.Reg.CPSR.Z = res == 0
            cpu.Reg.CPSR.C = false
            // FLAG_C "destroyed" ARM <5, ignored ARM >=5
            // FLAG_V maybe destroyed on ARM <5. ignored ARM <=5
        }

        r[PC] += 4
		return
	}

    // arm9 muliplies

    x := (opcode >> 5) & 1
    y := (opcode >> 6) & 1

    switch inst {
    case SMLAxy :

        rmV := int64(int16((r[rm] >> (16 * x)) & 0xFFFF))
        rsV := int64(int16((r[rs] >> (16 * y)) & 0xFFFF))
        rnV := int64(int32(r[rn]))

        res := rmV * rsV

        res += rnV

        r[rd] = uint32(res)

        if (res > math.MaxInt32 || res < math.MinInt32) {
            cpu.Reg.CPSR.Q = true
        }

    case SMLAWySMLALWy:

        rsV := int64(int16((r[rs] >> (16 * y) & 0xFFFF)))
        rmV := int64(int32(r[rm]))
        res := (rmV * rsV) >> 16

        if smulwa := x == 0; smulwa {
            add := int64(int32(r[rn]))

            if res + add > math.MaxInt32 || res + add < math.MinInt32 {
                cpu.Reg.CPSR.Q = true
            }

            res += add
        }

        r[rd] = uint32(res)

    case SMLALxy:

        rsV := int64(int16((r[rs] >> (16 * y) & 0xFFFF)))
        rmV := int64(int16((r[rm] >> (16 * x) & 0xFFFF)))

        res := rsV * rmV
        add := int64(int32(r[rd]))<<32 | int64(int32(r[rn]))
        res += add

        r[rd] = uint32(res >> 32)
        r[rn] = uint32(res)

    case SMULxy:

        rmV := int64(int16((r[rm] >> (16 * x)) & 0xFFFF))
        rsV := int64(int16((r[rs] >> (16 * y)) & 0xFFFF))

        res := rmV * rsV

        r[rd] = uint32(res)
    }

    r[PC] += 4
}

const (
	STR = iota
	LDR_PLD
)

func (c *Cpu) Sdt(opcode uint32) {

	r := &c.Reg.R

	if valid := utils.GetVarData(opcode, 26, 27) == 0b01; !valid {
		panic("Malformed Sdt Instruction")
	}


    rd   := (opcode >> 12) & 0xF
    rn   := (opcode >> 16) & 0xF
    preFlag  := (opcode >> 24) & 1 != 0
    byte := (opcode >> 22) & 1 != 0
    wb   := (opcode >> 21) & 1 != 0
    load := (opcode >> 20) & 1 != 0

	post := generateSdtAddress(c, opcode)
    pre := r[rn]
    if preFlag {
        pre = post
	}

	switch {
	case load && byte:
		// DO NOT WORD ALIGN
		r[rd] = c.mem.Read8(pre, true)

	case load && !byte:

        addr := pre &^ 0b11
		v := c.mem.Read32(addr, true)
		is := (pre & 0b11) << 3
		v = utils.RorSimple(v, is)

        r[rd] = v

        if rd == PC {
            c.toggleThumb()
            r[rd] -= 4

            if c.Reg.CPSR.T {
                r[rd] &^= 0b1
            } else {
                r[rd] &^= 0b11
            }
        }

	case !load && byte:

		c.mem.Write8(pre, uint8(r[rd]), true)

	case !load && !byte:

		v := r[rd]
		if rd == PC {
			v += 12
		}

        addr := pre &^ 0b11

		c.mem.Write32(addr, v, true)
	}

    if writeback := (!preFlag || wb) && !(load && rn == rd); writeback {
		r[rn] = post
    }

	c.Reg.R[PC] += 4
}

func generateSdtAddress(cpu *Cpu, opcode uint32) uint32 {

	r := &cpu.Reg.R

	var offset uint32
    if shReg := (opcode >> 25) & 1 != 0; shReg {

        if utils.BitEnabled(opcode, 4) {
            panic("Malformed Single Data Transfer")
        }

        shift := utils.GetVarData(opcode, 7, 11)
        shiftType := utils.GetVarData(opcode, 5, 6)
        rm := utils.GetByte(opcode, 0)

        offset, _, _ = utils.ShiftFuncs[shiftType](
			r[rm],
			shift,
			false,
			true,
			cpu.Reg.CPSR.C,
        )

    } else {

        offset = utils.GetVarData(opcode, 0, 11)
    }

	rn := utils.GetByte(opcode, 16)
	addr := r[rn]
	if rn == PC {
		addr += 8
	}

    if up := utils.BitEnabled(opcode, 23); up {
        return addr + offset
	}

    return addr - offset
}

func (cpu *Cpu) BLX(opcode uint32) {

    r := &cpu.Reg.R

    r[14] = r[15] + 4

    r[PC] += uint32((int32(opcode)<<8)>>6) + 8

    if halfOffset := utils.BitEnabled(opcode, 24); halfOffset {
        r[PC] += 2
    }

    cpu.Reg.CPSR.T = true
}

func (cpu *Cpu) B(opcode uint32) {

    r := &cpu.Reg.R

	if isLink := utils.BitEnabled(opcode, 24);isLink {
		r[14] = r[15] + 4
    }

	r[PC] += uint32((int32(opcode)<<8)>>6) + 8
}

func (cpu *Cpu) BX(opcode uint32) {

	const (
		INST_BX  = 1
		INST_BXJ = 2
		INST_BLX = 3
	)

	inst := utils.GetByte(opcode, 4)
	rn := utils.GetByte(opcode, 0)
	r := &cpu.Reg.R

	switch inst {
	case INST_BX:
		r[PC] = r[rn]
		cpu.toggleThumb()
	case INST_BXJ:
		panic("Unsupported BXJ Instruction")
	case INST_BLX:

		if rn == 14 {
			// Using BLX R14 is possible (sets PC=Old_LR, and New_LR=retadr).
            tmp := r[14]
            r[14] = r[PC] + 4
            r[PC] = tmp
            cpu.toggleThumb()
            return
		}

		r[14] = r[PC] + 4
		r[PC] = r[rn]
		cpu.toggleThumb()
	}
}

const (
	RESERVED = 0
	STRH     = 1
	LDRD     = 2
	STRD     = 3

	LDRH  = 1
	LDRSB = 2
	LDRSH = 3
)

var halfData Half

type Half struct {
	Rn, Rd, Imm, Inst, Rm, RdValue, RnValue, RdValue2, RmValue      uint32
	Pre, Up, Immediate, WriteBack, Load, MemoryManagement bool
}

func NewHalf(opcode uint32, c *Cpu) *Half {

	r := &c.Reg.R

	halfData.Rn = utils.GetByte(opcode, 16)
	halfData.Rd = utils.GetByte(opcode, 12)
	halfData.Pre = utils.BitEnabled(opcode, 24)
	halfData.Up = utils.BitEnabled(opcode, 23)
	halfData.Immediate = utils.BitEnabled(opcode, 22)
	halfData.Load = utils.BitEnabled(opcode, 20)
	halfData.Inst = utils.GetVarData(opcode, 5, 6)

	if halfData.Pre {
		halfData.WriteBack = utils.BitEnabled(opcode, 21)
	} else {
		halfData.WriteBack = true
	}

	fails := []bool{
		!halfData.Pre && utils.BitEnabled(opcode, 21),
		!utils.BitEnabled(opcode, 7),
		!utils.BitEnabled(opcode, 4),
		//halfData.Immediate && !(utils.GetByte(opcode, 8) == 0b0000),
	}

	for i, fail := range fails {
		if fail {
			panic(fmt.Sprintf("Malformed Half Instruction %d %08X", i, opcode))
		}
	}

	halfData.Rm = utils.GetByte(opcode, 0)
	halfData.RmValue = r[halfData.Rm]
	halfData.Imm = utils.GetByte(opcode, 8)<<4 | utils.GetByte(opcode, 0)

	halfData.RnValue = r[halfData.Rn]
	if halfData.Rn == PC {
		halfData.RnValue += 8
	}

	halfData.RdValue = r[halfData.Rd]
	if halfData.Rd == PC {
		halfData.RdValue += 12
	}

    if double := !halfData.Load && halfData.Inst != 1; double {

        if odd := halfData.Rd & 1 == 1; odd {
            panic(fmt.Sprintf("DOUBLE INST WITH ODD RD REG %02X OPCODE %08X\n", r[halfData.Rd], opcode))
        }

        halfData.RdValue2 = r[halfData.Rd + 1]
        if halfData.Rd + 1 == PC {
            halfData.RdValue2 += 12
        }
    }

	return &halfData
}

func (c *Cpu) Half(opcode uint32) {

	half := NewHalf(opcode, c)

	if !half.Load {
		switch half.Inst {
		case RESERVED:
			panic("RESERVED HALF (Load) NOT SUPPORTED")
		case STRH:
			unsignedHalfStd(half, c)
		case LDRD:
            doubleStd(half, c, true)
		case STRD:
            doubleStd(half, c, false)
		}

		c.Reg.R[15] += 4
		return
	}

	switch half.Inst {
	case RESERVED:
		panic("RESERVED HALF (Store) NOT SUPPORTED")
	case LDRH:
		unsignedHalfStd(half, c)
	case LDRSB:
		signedByteStd(half, c)
	case LDRSH:
		signedHalfStd(half, c)
	}

	c.Reg.R[15] += 4
}

func doubleStd(half *Half, cpu *Cpu, load bool) {

    //STRD/LDRD: base writeback: Rn should not be same as R(d) or R(d+1).
    //STRD: index register: Rm should not be same as R(d) or R(d+1).
    //STRD/LDRD: Rd must be an even numbered register (R0,R2,R4,R6,R8,R10,R12).
    //STRD/LDRD: Address must be double-word aligned (multiple of eight).

	r := &cpu.Reg.R
	pre, post := halfUnsignedAddress(half, cpu)
	addr := pre &^ 0b111

	if load {
		r[half.Rd] = cpu.mem.Read32(addr, true)
		r[half.Rd + 1] = cpu.mem.Read32(addr + 4, true)
	} else {
		cpu.mem.Write32(addr, half.RdValue, true)
		cpu.mem.Write32(addr + 4, half.RdValue2, true)
	}

	skipLoadWriteBack := half.Load && (half.Rn == half.Rd)
	if (half.WriteBack || !half.Pre) && !skipLoadWriteBack {
		r[half.Rn] = post
	}
}

func signedByteStd(half *Half, cpu *Cpu) {

	r := &cpu.Reg.R
	pre, post := halfUnsignedAddress(half, cpu)
	addr := pre &^ 0b1

	if half.Load {
		// sign-expand byte value
		unexpanded := int8(cpu.mem.Read8(pre, true))
		expanded := uint32(unexpanded)

		if unexpanded < 0 {
			expanded |= (0xFFFFFF << 8)
		}

		r[half.Rd] = expanded
	} else {
		cpu.mem.Write16(addr, uint16(int16(half.RdValue)), true)
	}

	skipLoadWriteBack := half.Load && (half.Rn == half.Rd)
	if (half.WriteBack || !half.Pre) && !skipLoadWriteBack {
		r[half.Rn] = post
	}
}

func signedHalfStd(half *Half, cpu *Cpu) {

	r := &cpu.Reg.R
	pre, post := halfUnsignedAddress(half, cpu)

	if half.Load {
			// sign-expand half value
            unexpanded := int16(cpu.mem.Read16(pre &^ 0b1, true))
            expanded := uint32(unexpanded)

            if unexpanded < 0 {
                expanded |= (0xFFFF << 16)
            }

            r[half.Rd] = expanded
        //v := uint32(uint16(int16(cpu.mem.Read16(pre&^0b1, true))))
        //r[half.Rd] = v
	} else {
		addr := pre &^ 0b1
		cpu.mem.Write16(addr, uint16(int16(half.RdValue)), true)
	}

	skipLoadWriteBack := half.Load && (half.Rn == half.Rd)
	if (half.WriteBack || !half.Pre) && !skipLoadWriteBack {
		r[half.Rn] = post
	}
}

func unsignedHalfStd(half *Half, cpu *Cpu) {
	r := &cpu.Reg.R
	pre, post := halfUnsignedAddress(half, cpu)
	addr := pre &^ 0b1


	if half.Load {
		v := uint32(cpu.mem.Read16(addr, true))
		r[half.Rd] = v

	} else {
		cpu.mem.Write16(addr, uint16(half.RdValue), true)
	}

	skipLoadWriteBack := half.Load && (half.Rn == half.Rd)
	if (half.WriteBack || !half.Pre) && !skipLoadWriteBack {
		r[half.Rn] = post
	}
}

func halfUnsignedAddress(half *Half, cpu *Cpu) (uint32, uint32) {

	r := &cpu.Reg.R

	var offset uint32
	if half.Immediate {
		offset = half.Imm
	} else {
		offset = half.RmValue
	}

	addr := r[half.Rn]

	if half.Up {
		addr += offset
	} else {
		addr -= offset
	}

	if half.Pre {
		return addr, addr
	} else {
		return r[half.Rn], addr
	}
}

type PSR struct {
	Opcode, Rd, Rm, Shift, Imm       uint32
	SPSR, MSR, Immediate, F, S, X, C bool
}

func NewPSR(opcode uint32, cpu *Cpu) *PSR {

	psr := &PSR{
		Opcode:    opcode,
		Immediate: utils.BitEnabled(opcode, 25),
		SPSR:      utils.BitEnabled(opcode, 22),
		MSR:       utils.BitEnabled(opcode, 21),
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
		mode := cpu.Reg.CPSR.Mode
        r[psr.Rd] = cpu.Reg.SPSR[BANK_ID[mode]].Get()
		return
	}

	mask := PRIV_MASK
	if cpu.Reg.CPSR.Mode == MODE_USR {
		mask = USR_MASK
	}


	r[psr.Rd] = uint32(cpu.Reg.CPSR.Get()) & mask
}

const (
	PRIV_MASK  uint32 = 0xF8FF_03DF
	USR_MASK   uint32 = 0xF8FF_0000
	STATE_MASK uint32 = 0x0100_0020
)

func (cpu *Cpu) msr(psr *PSR) {

	reg := &cpu.Reg
	r := &cpu.Reg.R

	var v uint32
	if psr.Immediate {
		v = utils.RorSimple(psr.Imm, psr.Shift)
		//v, _, _ = utils.Ror(psr.Imm, psr.Shift, false, false, false)
	} else {
		v = r[psr.Rm]
	}

	mask := uint32(0)
	if psr.C {
		mask |= 0x0000_00FF
	}
	if psr.X {
		mask |= 0x0000_FF00
	}
	if psr.S {
		mask |= 0x00FF_0000
	}
	if psr.F {
		mask |= 0xFF00_0000
	}

	secMask := PRIV_MASK
	curr := cpu.Reg.CPSR.Mode
	if curr == MODE_USR {
		secMask = USR_MASK
	}

	if psr.SPSR {
		secMask |= STATE_MASK
	}

	mask &= secMask

	if psr.SPSR {

		var spsr uint32

		if curr == MODE_USR || curr == MODE_SYS {
			spsr = uint32(reg.CPSR.Get()) &^ mask
		} else {
			spsr = uint32(reg.SPSR[BANK_ID[curr]].Get()) &^ mask
		}

		spsr |= v & mask
		reg.SPSR[BANK_ID[curr]].Set(spsr)

		return
	}

	next := v & 0b11111
	cpsr := uint32(reg.CPSR.Get()) &^ mask

	cpsr |= v & mask

	reg.CPSR.Set(cpsr)

	if skip := BANK_ID[curr] == BANK_ID[next]; skip {
		return
	}

	if curr == MODE_USR {
		panic("USER MODE MSR")
	}

	if curr != MODE_FIQ {
		for i := range 5 {
			reg.USR[i] = r[8+i]
		}
	}

	reg.SP[BANK_ID[curr]] = r[SP]
	reg.LR[BANK_ID[curr]] = r[LR]

	if curr == MODE_FIQ {
		for i := range 5 {
			reg.FIQ[i] = r[8+i]
		}
	}

	if next != MODE_FIQ {
		for i := range 5 {
			r[8+i] = reg.USR[i]
		}
	}

	r[SP] = reg.SP[BANK_ID[next]]
	r[LR] = reg.LR[BANK_ID[next]]

	if next == MODE_FIQ {
		for i := range 5 {
			r[8+i] = reg.FIQ[i]
		}
	}

}

func (cpu *Cpu) Swp(opcode uint32) {

	isByte := utils.BitEnabled(opcode, 22)
	rn := utils.GetByte(opcode, 16)
	rd := utils.GetByte(opcode, 12)
	rm := utils.GetByte(opcode, 0)

	r := &cpu.Reg.R

	rmValue := r[rm]
	rnValue := r[rn]

	if isByte {
		r[rd] = cpu.mem.Read8(rnValue, true)
		cpu.mem.Write8(rnValue, uint8(rmValue), true)
		r[PC] += 4
		return
    }

    v := cpu.mem.Read32(rnValue &^ 0b11, true)
    is := (rnValue & 0b11) << 3

    //rnMemValue, _, _ = utils.Ror(rnMemValue, is, false, false, false)
    v = utils.RorSimple(v, is)

	r[rd] = v
	cpu.mem.Write32(rnValue, rmValue, true)
	r[PC] += 4
}

const (
    QADD = 0
    QSUB = 2
    QDADD = 4
    QDSUB = 6
)

func (cpu *Cpu) Qalu(opcode uint32) {

    r := &cpu.Reg.R

    inst := (opcode >> 20) & 0xF
    rnV := int64(int32(r[(opcode >> 16) & 0xF]))
    rmV := int64(int32(r[opcode & 0xF]))

    if double := inst >= 4; double {

        rnV *= 2

        if rnV > math.MaxInt32 {
            cpu.Reg.CPSR.Q = true
            rnV = math.MaxInt32
        }

        if rnV < math.MinInt32 {
            cpu.Reg.CPSR.Q = true
            rnV = math.MinInt32
        }
    }

    if inst == QADD || inst == QDADD {
        rnV += rmV
    } else {
        rnV = rmV - rnV
    }

    if rnV > math.MaxInt32 {
        cpu.Reg.CPSR.Q = true
        rnV = math.MaxInt32
    }

    if rnV < math.MinInt32 {
        cpu.Reg.CPSR.Q = true
        rnV = math.MinInt32
    }

    rd := (opcode >> 12) & 0xF
    r[rd] = uint32(rnV)

    r[PC] += 4
}

func (cpu *Cpu) Clz(opcode uint32) {

    r := &cpu.Reg.R
    rm := opcode & 0b1111


    rd := utils.GetVarData(opcode, 12, 15)

    r[rd] = uint32(bits.LeadingZeros32(r[rm]))

    r[PC] += 4
}

func (cpu *Cpu) CoDataReg(opcode uint32) {

    reg := cp15.CpRegister{
        Op: uint8(utils.GetVarData(opcode, 21, 23)),
        Cn: uint8(utils.GetVarData(opcode, 16, 19)),
        Pn: uint8(utils.GetVarData(opcode, 8, 11)),
        Cp: uint8(utils.GetVarData(opcode, 5, 7)),
        Cm: uint8(utils.GetVarData(opcode, 0, 3)),
    }

    r := &cpu.Reg.R

    rd := utils.GetVarData(opcode, 12, 15)

    if (opcode >> 28) == 0xF {
        panic("MRC2/MCR2")
    }

    if rd == 15 { panic("SETUP PIPELINE OFFSET CO DATA REG")}

    if mrc := (opcode >> 20) & 1 == 1; mrc {
        r[rd] = cpu.Cp15.Read(reg)
        cpu.Reg.R[15] += 4
        return
    }

    if rd == 0 && (reg == cp15.HALT || reg == cp15.HALT2) {

        if !cpu.Irq.IME {
            panic("ARM9 CPU HALTED WITHOUT IME ENABLED")
            //cpu.Irq.IME = true
        }

        cpu.Halted = true
    } else {
        cpu.Cp15.Write(r[rd], reg, &cpu.LowVector)
    }

    cpu.Reg.R[15] += 4
}
