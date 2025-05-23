package gba

import (
	"github.com/aabalke33/guac/emu/gba/utils"
    "fmt"
)

func (cpu *Cpu) DecodeARM(opcode uint32) int {

    r := &cpu.Reg.R

	if !cpu.CheckCond(utils.GetByte(opcode, 28)) {
		r[PC] += 4
		return 4
	}

	switch {
	case isSWI(opcode):
        cpu.Gba.Mem.BIOS_MODE = BIOS_SWI
        cpu.Gba.SysCall(utils.GetVarData(opcode, 16, 23))
        r[PC] += 4

        return 20

	case isB(opcode):
		cpu.B(opcode)
	case isBX(opcode):
		cpu.BX(opcode)
	case isSDT(opcode):
		cpu.Sdt(opcode)
	case isBlock(opcode):
		cpu.Block(opcode)
	case isHalf(opcode):
		cpu.Half(opcode)
	case isUD(opcode):
		panic("Need Undefined functionality")
	case isPSR(opcode):
		cpu.Psr(opcode)
	case isSWP(opcode):
		cpu.Swp(opcode)
    case isM(opcode):
        cpu.Mul(opcode)
	case isALU(opcode):
		cpu.Alu(opcode)

        if cpu.Gba.Ct.instCycles != 0 {
            return cpu.Gba.Ct.instCycles
        }

	default:
		panic(fmt.Sprintf("Unable to Decode %X, at PC %X, INSTR %d", opcode, r[PC], CURR_INST))
	}

    return 4

    // Notes: Coprocessor instructions do not matter since gba 
    // uses a single processor (NDS is a different story)
}

func isOpcodeFormat(opcode, mask, format uint32) bool {
	return opcode&mask == format
}

func isSWP(opcode uint32) bool {

	return isOpcodeFormat(opcode,
		0b0000_1111_1011_0000_0000_1111_1111_0000,
		0b0000_0001_0000_0000_0000_0000_1001_0000,
	)
}

func isBlock(opcode uint32) bool {

	is := false

	is = is || isOpcodeFormat(opcode,
		0b0000_1110_0001_0000_0000_0000_0000_0000,
		0b0000_1000_0001_0000_0000_0000_0000_0000,
	)

	is = is || isOpcodeFormat(opcode,
		0b0000_1110_0001_0000_0000_0000_0000_0000,
		0b0000_1000_0000_0000_0000_0000_0000_0000,
	)

	return is
}

func isHalf(opcode uint32) bool {

	is := false

	// LDRH
	is = is || isOpcodeFormat(opcode,
		0b0000_1110_0001_0000_0000_0000_1111_0000,
		0b0000_0000_0001_0000_0000_0000_1011_0000,
	)

	// LDRSB
	is = is || isOpcodeFormat(opcode,
		0b0000_1110_0001_0000_0000_0000_1111_0000,
		0b0000_0000_0001_0000_0000_0000_1101_0000,
	)

	// LDRSH
	is = is || isOpcodeFormat(opcode,
		0b0000_1110_0001_0000_0000_0000_1111_0000,
		0b0000_0000_0001_0000_0000_0000_1111_0000,
	)

	// STRH
	is = is || isOpcodeFormat(opcode,
		0b0000_1110_0001_0000_0000_0000_1111_0000,
		0b0000_0000_0000_0000_0000_0000_1011_0000,
	)

	return is
}

func isALU(opcode uint32) bool {
	return isOpcodeFormat(opcode,
		0b0000_1100_0000_0000_0000_0000_0000_0000,
		0b0000_0000_0000_0000_0000_0000_0000_0000,
	)
}

func isBX(opcode uint32) bool {
	return isOpcodeFormat(opcode,
		0b0000_1111_1111_1111_1111_1111_1111_0000,
		0b0000_0001_0010_1111_1111_1111_0001_0000,
	)
}

func isB(opcode uint32) bool {
	return isOpcodeFormat(opcode,
		0b0000_1110_0000_0000_0000_0000_0000_0000,
		0b0000_1010_0000_0000_0000_0000_0000_0000,
	)
}

func isM(opcode uint32) bool {

    is := false

	is = is || isOpcodeFormat(opcode,
		0b0000_1110_1000_0000_0000_0000_1111_0000,
		0b0000_0000_0000_0000_0000_0000_1001_0000,
	)
	is = is || isOpcodeFormat(opcode,
		0b0000_1110_1000_0000_0000_0000_1111_0000,
		0b0000_0000_1000_0000_0000_0000_1001_0000,
	)

    return is
}

func isSWI(opcode uint32) bool {
	return isOpcodeFormat(
		opcode,
		0b0000_1111_0000_0000_0000_0000_0000_0000,
		0b0000_1111_0000_0000_0000_0000_0000_0000,
	)
}

func isUD(opcode uint32) bool {
	return isOpcodeFormat(
		opcode,
		0b0000_1110_0000_0000_0000_0000_0000_0000,
		0b0000_0110_0000_0000_0000_0000_0000_0000,
	)
}

func isSDT(opcode uint32) bool {
	is := false
	is = is || isOpcodeFormat(
		opcode,
        0b0000_1100_0001_0000_0000_0000_0000_0000,
        0b0000_0100_0001_0000_0000_0000_0000_0000,
	)
	is = is || isOpcodeFormat(
		opcode,
        0b0000_1100_0001_0000_0000_0000_0000_0000,
        0b0000_0100_0000_0000_0000_0000_0000_0000,
	)

	return is
}

func isPSR(opcode uint32) bool {

	is := false

	is = is || isOpcodeFormat(
		opcode,
		0b0000_1111_1011_1111_0000_1111_1111_1111,
		0b0000_0001_0000_1111_0000_0000_0000_0000,
	)

	is = is || isOpcodeFormat(
		opcode,
		0b0000_1101_1011_0000_1111_0000_0000_0000,
		0b0000_0001_0010_0000_1111_0000_0000_0000,
	)

	return is
}
