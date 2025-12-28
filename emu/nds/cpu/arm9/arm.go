package arm9

import (
	"fmt"
	"math"
	"math/bits"

	"github.com/aabalke/guac/emu/nds/cpu/arm9/cp15"
	"github.com/aabalke/guac/emu/nds/utils"
)

const (
	LSL = iota
	LSR
	ASR
	ROR
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

var _ = fmt.Sprintf

var aluInst = [16]func(cpu *Cpu, alu *Alu){

    // AND 
    func(cpu *Cpu, alu *Alu) {
        res := alu.RnValue & alu.Op2
        cpu.Reg.R[alu.Rd] = res
        cpu.logicalExit(alu)
        if set := utils.BitEnabled(alu.Opcode, 20); set {
            cpu.Reg.CPSR.N = utils.BitEnabled(uint32(res), 31)
            cpu.Reg.CPSR.Z = uint32(res) == 0
        }
    },

    // EOR
    func(cpu *Cpu, alu *Alu) {
        res := alu.RnValue ^ alu.Op2
        cpu.Reg.R[alu.Rd] = res
        cpu.logicalExit(alu)
        if set := utils.BitEnabled(alu.Opcode, 20); set {
            cpu.Reg.CPSR.N = utils.BitEnabled(uint32(res), 31)
            cpu.Reg.CPSR.Z = uint32(res) == 0
        }
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
        cpu.logicalExit(alu)
        if set := utils.BitEnabled(alu.Opcode, 20); set {
            cpu.Reg.CPSR.N = utils.BitEnabled(uint32(res), 31)
            cpu.Reg.CPSR.Z = uint32(res) == 0
        }
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
        cpu.logicalExit(alu)
        if set := utils.BitEnabled(alu.Opcode, 20); set {
            cpu.Reg.CPSR.N = utils.BitEnabled(uint32(res), 31)
            cpu.Reg.CPSR.Z = uint32(res) == 0
        }
    },

    // MVN
    func(cpu *Cpu, alu *Alu) {
        res := ^alu.Op2
        cpu.Reg.R[alu.Rd] = res
        cpu.logicalExit(alu)
        if set := utils.BitEnabled(alu.Opcode, 20); set {
            cpu.Reg.CPSR.N = utils.BitEnabled(uint32(res), 31)
            cpu.Reg.CPSR.Z = uint32(res) == 0
        }
    },
}

type Alu struct {
	Opcode, Rd, RnValue, Op2       uint32
	Carry                          bool
}

var aluData Alu

func (cpu *Cpu) Alu(opcode uint32) {

    //reggy   := (opcode >> 25) & 1 == 0
    //rd := (opcode >>12) & 0xF
    //compare := !reggy && rd != PC
    //cpu.Jit.StartTest(opcode, compare, cpu.Jit.emitAlu)

	aluData.Opcode = opcode
	aluData.Rd = utils.GetByte(opcode, 12)
    aluData.Carry = cpu.Reg.CPSR.C
    rn := (opcode >> 16) & 0xF
    aluData.RnValue = cpu.Reg.R[rn]

    if imm := (opcode>>25) & 1 == 1; imm {

        ro := ((opcode >> 8) & 0xF) << 1
        aluData.Op2 = bits.RotateLeft32(opcode & 0xFF, -int(ro))

        if setCarry := ro != 0 && (opcode >> 20) & 1 == 1; setCarry {
            // I believe this matches
            //carry := (nn >> (ro-1)) & 1 != 0 // this line must be before opcode
            cpu.Reg.CPSR.C = aluData.Op2 >> 31 != 0
            //cpu.Reg.CPSR.C = carry
        }

        if rn == PC {
            aluData.RnValue += 8
        }

    } else {

        aluData.Op2 = cpu.getShiftedAluReg(opcode)

        if rn == PC {
            if imm := !utils.BitEnabled(opcode, 4); imm {
                aluData.RnValue += 8
            } else {
                aluData.RnValue += 12
            }
        }
    }

    inst := (opcode >> 21) & 0xF
    aluInst[inst](cpu, &aluData)


    switch {
    case aluData.Rd != PC:
		cpu.Reg.R[15] += 4
    case cpu.Reg.CPSR.T:
        cpu.Reg.R[15] &^= 0b1
    case !cpu.Reg.CPSR.T:
        cpu.Reg.R[15] &^= 0b11
    }

    //cpu.Jit.EndTest(opcode, compare)
}

func (cpu *Cpu) getShiftedAluReg(op uint32) uint32 {

    r := &cpu.Reg.R

    carry := cpu.Reg.CPSR.C

    shReg    := (op >> 4) & 1 != 0
    shType   := (op >> 5) & 0b11

    setCarry := (op >> 20) & 1 != 0
    inst     := (op >> 21) & 0xF
    logical  := inst & 0b0110 == 0b0000 || inst & 0b1100 == 0b1100
    setCarry = setCarry && logical

    rm       := op & 0xF
    op2      := r[rm]
    var shift uint32

    if shReg {
        rs := (op >> 8) & 0xF
        shift = r[rs] & 0xFF

        if rm == PC {
            op2 += 12
        }

    } else {

        shift = (op >> 7) & 0x1F

        if rm == PC {
            op2 += 8
        }

        if special := shift == 0; special {
            switch shType{
            case LSL:
                return op2
            case LSR:
                cpu.Reg.CPSR.C = op2 & 0x8000_0000 != 0
                return 0
            case ASR:

                signed := op2 & 0x8000_0000 != 0

                if setCarry {
                    cpu.Reg.CPSR.C = signed
                }

                if signed {
                    return 0xFFFF_FFFF
                }

                return 0

            case ROR:

                cpu.Reg.CPSR.C = op2 & 1 != 0

                op2 >>= 1
                if carry {
                    op2 |= 0x8000_0000
                }

                return op2
            }
        }
    }

    // https://iitd-plos.github.io/col718/ref/arm-instructionset.pdf

    if regZero := shift == 0; regZero {
        // op2 unchanges, carry is set to original carry (no change)
        return op2
    }

    switch shType {
    case LSL:

        switch {
        case shift > 32:
            op2 = 0
            carry = false
        case shift == 32:
            carry = op2 & 1 != 0
            op2 = 0
        default:
            carry = op2 & (1 << (32-shift)) != 0
            op2 <<= shift
        }

    case LSR:

        switch {
        case shift > 32:
            op2 = 0
            carry = false
        case shift == 32:
            carry = op2 & 0x8000_0000 != 0
            op2 = 0
        default:
            carry = op2 & (1 << (shift-1)) != 0
            op2 >>= shift
        }

    case ASR:

        switch {
        case shift >= 32:
            signed := op2 & 0x8000_0000 != 0
            carry = signed

            if signed {
                op2 = 0xFFFF_FFFF
            } else {
                op2 = 0x0
            }

        default:
            carry = op2 & (1 << (shift-1)) != 0
            op2 = uint32(int32(op2) >> shift)
        }

    case ROR:

        switch {
        case shift == 32:
            // op2 unchanges
            carry = op2 & 0x8000_0000 != 0
        default:
            carry = (op2 >> ((shift-1) & 31)) & 1 != 0
            op2 = bits.RotateLeft32(op2, -int(shift))
        }
    }

    if setCarry {
        cpu.Reg.CPSR.C = carry
    }

    return op2
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

func (cpu *Cpu) logicalExit(alu *Alu) {

    // I think this is already handled by &^ 0b11, &^ 1 in end of alu
    //if alu.Rd == PC {
    //    //cpu.toggleThumb()
    //    // this may be a problem still
    //    cpu.Reg.R[alu.Rd] &^= 0b1
    //}
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

	rnSign := (alu.RnValue>>31) & 1 != 0
	opSign := (alu.Op2>>31) & 1 != 0
	rSign := (res>>31) & 1 != 0

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

func (c *Cpu) Sdt(op uint32) {

	r := &c.Reg.R

    if valid   := (op >> 26) & 0b11 == 0b01; !valid {
		panic("Malformed Sdt Instruction")
	}

    reg  := (op >> 25) & 1 != 0
    pre  := (op >> 24) & 1 != 0
    up   := (op >> 23) & 1 != 0
    byte := (op >> 22) & 1 != 0
    wb   := (op >> 21) & 1 != 0 || !pre
    load := (op >> 20) & 1 != 0
    rn   := (op >> 16) & 0xF
    rd   := (op >> 12) & 0xF

    //compare := !byte && load && pre && rd != PC
    //if compare {
    //    c.Jit.TestInst(op, c.Jit.emitSdt)
    //    r[15] += 4
    //    return
    //}

	var offset, prev uint32
    if reg {

        if (op >> 4) & 1 != 0 {
            panic("Malformed Single Data Transfer O_o")
        }

        shift := (op >> 7) & 0x1F
        sType := (op >> 5) & 0b11
        rm    := op & 0xF

        switch sType {
        case LSL:

            offset = r[rm] << shift

        case LSR:

            if shift == 0 {
                shift = 32
            }

            offset = r[rm] >> shift

        case ASR:

            if shift == 0 {
                shift = 32
            }
            
            offset = uint32(int32(r[rm]) >> shift)

        case ROR:

            if shift == 0 {

                offset = r[rm] >> 1

                if c.Reg.CPSR.C {
                    offset |= 0x8000_0000
                }
            } else {
                offset = bits.RotateLeft32(r[rm], -int(shift))
            }
        }

    } else {
        offset = op & 0xFFF
    }

	post := r[rn]
	if rn == PC {
		post += 8
	}

    if up {
        post += offset
	} else {
        post -= offset
    }

    if pre {
        prev = post
	} else {
        prev = r[rn]
    }

    //compare := (
    //    rd != PC &&
    //    prev != 0x410_0000 &&
    //    prev != 0x410_0010 &&
    //    !(prev >= 0x400_0180 && prev < 0x400_0200) &&
    //    !(prev >= 0x400_0400 && prev < 0x400_0600))

    //c.Jit.StartTest(op, compare, c.Jit.emitSdt)


    //compare := (
    //    rd != PC &&
    //    load &&
    //    !(wb && rn == rd) &&
    //    prev & 0xF00_0000 != 0x400_0000)
    //c.Jit.StartTest(op, compare, c.Jit.emitSdt)

    if wb {
		r[rn] = post
    }

    if load {
        if byte {
            // DO NOT WORD ALIGN
            r[rd] = c.mem.Read8(prev, true)
        } else {

            v := c.mem.Read32(prev &^ 0b11, true)
            is := ((prev & 0b11) << 3) & 0x1F
            r[rd] = bits.RotateLeft32(v, -int(is))

            if rd == PC {
                c.toggleThumb()
                r[rd] -= 4
            }
        }
    } else {
        v := r[rd]
        if rd == PC {
            v += 12
        }

        if byte {
		    c.mem.Write8(prev, uint8(v), true)
        } else {
            c.mem.Write32(prev &^ 0b11, v, true)
        }
    }

    //c.Jit.EndTest(op, compare)

	r[PC] += 4
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

    if immLoop := opcode == 0xEAFFFFFE; immLoop {
        cpu.Halted = true
        return
    }

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

func (c *Cpu) Half(op uint32) {

    r := &c.Reg.R

    rn := (op >> 16) & 0xF
    rd := (op >> 12) & 0xF
    preFlag := (op >> 24) & 1 != 0
    load := (op >> 20) & 1 != 0
    inst := (op >> 5)  & 0b11
    wb := (op >> 21) & 1 != 0 || !preFlag

    //if (rd != PC) {
    //    //reg := c.Reg
    //    c.Jit.TestInst(op, c.Jit.emitHalf)
    //    c.Reg.R[PC] += 4
    //    //c.Reg = reg
    //    return
    //}

    rnv := r[rn]
	if rn == PC {
        rnv += 8
	}

	var offset uint32
    if imm := (op >> 22) & 1 != 0; imm {
        offset = (op & 0xF) | (((op >> 8) & 0xF) << 4)
	} else {
        offset = r[op & 0xF]
	}

	post := rnv

    if up := (op >> 23) & 1 != 0; up {
		post += offset
	} else {
		post -= offset
	}

    pre := post
    if !preFlag {
        pre = rnv
    }

    //compare := (
    //    rd != PC &&
    //    load &&
    //    !(wb && rn == rd) &&
    //    pre & 0xF00_0000 != 0x400_0000)
    //c.Jit.StartTest(op, compare, c.Jit.emitHalf)

    if inst == RESERVED {
        panic("unsupported half (reserved)")
    }


	if !load {
        rdv := r[rd]

        if rd == PC {
            rdv += 12
        }
        rd2v := r[rd + 1]
        if rd + 1 == PC {
            rd2v += 12
        }

        if wb {
            r[rn] = post
        }

		switch inst {
		case STRH:
            c.mem.Write16(pre &^ 1, uint16(rdv), true)

		case LDRD:
            addr := pre &^ 0b111
            r[rd] = c.mem.Read32(addr, true)
            r[rd + 1] = c.mem.Read32(addr + 4, true)

		case STRD:

            addr := pre &^ 0b111
            c.mem.Write32(addr, rdv, true)
            c.mem.Write32(addr + 4, rd2v, true)
		}

        //c.Jit.EndTest(op, compare)

		c.Reg.R[15] += 4
		return
	}

    if wb {
        r[rn] = post
    }

	switch inst {
	case LDRH:
        r[rd] = uint32(c.mem.Read16(pre &^ 1, true))

	case LDRSB:
        // sign-expand byte value
        r[rd] = uint32(int32(int8(c.mem.Read8(pre, true))))
        // may have extended in correctly

	case LDRSH:
        // sign-expand half value
        r[rd] = uint32(int32(int16(c.mem.Read16(pre &^ 1, true))))
	}

    //c.Jit.EndTest(op, compare)

	c.Reg.R[15] += 4
}

func (cpu *Cpu) Psr(opcode uint32) {

	r := &cpu.Reg.R

    if msr := (opcode >> 21) & 1 != 0; msr {
		cpu.msr(opcode)
		r[PC] += 4
		return
	}

    rd := (opcode >> 12) & 0xF

    if spsr := (opcode >> 22) & 1 != 0; spsr {
		mode := cpu.Reg.CPSR.Mode
        r[rd] = cpu.Reg.SPSR[BANK_ID[mode]].Get()
        r[PC] += 4
		return
	}

	mask := PRIV_MASK
	if cpu.Reg.CPSR.Mode == MODE_USR {
		mask = USR_MASK
	}

	r[rd] = uint32(cpu.Reg.CPSR.Get()) & mask
    r[PC] += 4
}

const (
	PRIV_MASK  uint32 = 0xF8FF_03DF
	USR_MASK   uint32 = 0xF8FF_0000
	STATE_MASK uint32 = 0x0100_0020
)

func (cpu *Cpu) msr(op uint32) {

	r := &cpu.Reg.R

    spsrFlag := (op >> 22) & 1 != 0

	var v uint32
    if imm := (op >> 25) & 1 != 0; imm {
        immv := op & 0xFF
        shift := ((op >> 8) & 0xF) * 2
		v = utils.RorSimple(immv, shift)
		//v, _, _ = utils.Ror(psr.Imm, psr.Shift, false, false, false)
	} else {
		v = r[op & 0xF]
	}

	mask := uint32(0)
    if C := (op >> 16) & 1 != 0; C {
		mask |= 0x0000_00FF
	}
    if X := (op >> 17) & 1 != 0; X {
		mask |= 0x0000_FF00
	}
    if S := (op >> 18) & 1 != 0; S {
		mask |= 0x00FF_0000
	}
    if F := (op >> 19) & 1 != 0; F {
		mask |= 0xFF00_0000
	}

	secMask := PRIV_MASK
	curr := cpu.Reg.CPSR.Mode
	if curr == MODE_USR {
		secMask = USR_MASK
	}

	if spsrFlag {
		secMask |= STATE_MASK
	}

	mask &= secMask

	reg := &cpu.Reg

	if spsrFlag {

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

    isByte := (opcode >> 22) & 1 != 0
    rn := (opcode >> 16) & 0xF
    rd := (opcode >> 12) & 0xF
    rm := opcode & 0xF

	r := &cpu.Reg.R

	rmValue := r[rm]
	rnValue := r[rn]

    //compare := isByte && rm != rn

    //rn0 := cpu.mem.Read32(rnValue, true)
    //rm0 := cpu.mem.Read32(rmValue, true)
    //cpu.Jit.TestInst(opcode, cpu.Jit.emitSwp)
    //cpu.Jit.StartTest(opcode, compare, cpu.Jit.emitSwp)
    //cpu.mem.Write32(rnValue, rn0, true)
    //cpu.mem.Write32(rmValue, rm0, true)

	if isByte {
		r[rd] = cpu.mem.Read8(rnValue, true)
		cpu.mem.Write8(rnValue, uint8(rmValue), true)
    } else {
        v := cpu.mem.Read32(rnValue &^ 0b11, true)
        is := (rnValue & 0b11) << 3
        v = bits.RotateLeft32(v, -int(is & 31))
        r[rd] = v
        cpu.mem.Write32(rnValue, rmValue, true)
    }

    //cpu.Jit.EndTest(opcode, compare)

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
    rm := opcode & 0xF
    rd := (opcode >> 12) & 0xF
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
