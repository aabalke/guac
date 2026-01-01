package arm7

import (
	"fmt"
)

func (cpu *Cpu) DecodeARM() (int, bool) {

	r := &cpu.Reg.R

	//opcode := cpu.mem.Read32(r[PC], false)
	opcode, cycles := cpu.GetOpArm()

	if !cpu.CheckCond(opcode >> 28) {
		r[PC] += 4
		return cycles + 1, true
	}

	if swi := (opcode>>24)&0xF == 0xF; swi {
		cpu.exception(VEC_SWI, MODE_SWI)
		return cycles + 1, true
	}

	switch (opcode >> 25) & 0b111 {
	case 0b000:
		switch {
		case isBX(opcode):
			cpu.BX(opcode)
		case isSDT(opcode):
			cpu.Sdt(opcode)
		case isHalf(opcode):
			cpu.Half(opcode)
		case isPSR(opcode):
			cpu.Psr(opcode)
		case isSWP(opcode):
			cpu.Swp(opcode)
		case isM(opcode):
			cpu.Mul(opcode)
		case isALU(opcode):
			cpu.Alu(opcode)
		default:
			panic(fmt.Sprintf("Unable to Decode ARM %08X, at PC %08X", opcode, r[PC]))
		}

		return cycles + 1, true
	case 0b001:

		switch {
		case isSDT(opcode):
			cpu.Sdt(opcode)
			return cycles + 1, true
		case isHalf(opcode):
			cpu.Half(opcode)
		case isPSR(opcode):
			cpu.Psr(opcode)
		case isSWP(opcode):
			cpu.Swp(opcode)
		case isALU(opcode):
			cpu.Alu(opcode)
		default:
			panic(fmt.Sprintf("Unable to Decode ARM %08X, at PC %08X", opcode, r[PC]))
		}

		return cycles + 1, true
	case 0b010:
	case 0b011:
	case 0b100:
		cpu.Block(opcode)
		return cycles + 1, true
	case 0b101:
		cpu.B(opcode)
		return cycles + 1, true
	}

	switch {
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
	default:
		fmt.Printf("Unable to Decode ARM 7 %08X, at PC %08X\n", opcode, r[PC])
		return 0, false
	}

	return cycles + 1, true
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
		0b0000_1111_1111_1111_1111_1111_1101_0000,
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
