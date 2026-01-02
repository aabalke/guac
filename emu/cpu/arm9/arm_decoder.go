package arm9

import (
	"fmt"
)

func (cpu *Cpu) DecodeARM() (int, bool) {

	r := &cpu.Reg.R

	op, cycles := cpu.GetOpArm()

	switch cond := op >> 28; cond {
	case 0xF:

		switch {
		case isBLX(op):
			cpu.BLX(op)
			return cycles + 1, true

		case isPLD(op):
			fmt.Printf("PC %08X OPCODE %08X CPSR %08X\n", r[PC], op, cpu.Reg.CPSR)
			panic("PLD")
			return cycles + 1, true
		}

	case 0xE:
		if isBkpt(op) {
			cpu.Exception(VEC_PREFETCHABORT, MODE_ABT)
			return cycles + 1, true
		}

	default:
		if !cpu.CheckCond(cond) {
			r[PC] += 4
			return cycles + 1, true
		}
	}

	if swi := (op>>24)&0xF == 0xF; swi {
		cpu.Exception(VEC_SWI, MODE_SWI)
		return cycles + 1, true
	}

	switch {
	case isBkpt(op):
		cpu.Exception(VEC_PREFETCHABORT, MODE_ABT)
	case isB(op):
		cpu.B(op)
	case isBX(op):
		cpu.BX(op)
	case isSDT(op):
		cpu.Sdt(op)
	case isBlock(op):
		cpu.Block(op)
	case isHalf(op):
		cpu.Half(op)
	case isUD(op):
		panic("Need Undefined functionality")
	case isPSR(op):
		cpu.Psr(op)
	case isSWP(op):
		cpu.Swp(op)
	case isM(op):
		cpu.Mul(op)
	case isCLZ(op):
		cpu.Clz(op)
	case isQAlu(op):
		cpu.Qalu(op)
	case isALU(op):
		cpu.Alu(op)
	case isCoDataReg(op):
		cpu.CoDataReg(op)
	default:
		fmt.Printf("Unable to Decode ARM 9 %08X, at PC %08X\n", op, r[PC])
		return 0, false
	}

	return cycles + 1, true
}

//go:inline
func isOpFormat(op, mask, format uint32) bool {
	return op&mask == format
}

//go:inline
func isBLX(op uint32) bool {
	return isOpFormat(op,
		0b1111_1110_0000_0000_0000_0000_0000_0000,
		0b1111_1010_0000_0000_0000_0000_0000_0000,
	)
}

//go:inline
func isBkpt(op uint32) bool {
	return isOpFormat(op,
		0b1111_1111_1111_0000_0000_0000_1111_0000,
		0b1110_0001_0010_0000_0000_0000_0111_0000,
	)
}

//go:inline
func isCLZ(op uint32) bool {
	return isOpFormat(op,
		0b0000_1111_1111_1111_0000_1111_1111_0000,
		0b0000_0001_0110_1111_0000_1111_0001_0000,
	)
}

//go:inline
func isQAlu(op uint32) bool {
	return isOpFormat(op,
		0b0000_1111_1001_0000_0000_1111_1111_0000,
		0b0000_0001_0000_0000_0000_0000_0101_0000,
	)
}

//go:inline
func isCoDataReg(op uint32) bool {
	return isOpFormat(op,
		0b0000_1111_0000_0000_0000_0000_0000_0000,
		0b0000_1110_0000_0000_0000_0000_0000_0000,
	)
}

//go:inline
func isSWP(op uint32) bool {

	return isOpFormat(op,
		0b0000_1111_1011_0000_0000_1111_1111_0000,
		0b0000_0001_0000_0000_0000_0000_1001_0000,
	)
}

//go:inline
func isBlock(op uint32) bool {

	is := false

	is = is || isOpFormat(op,
		0b0000_1110_0001_0000_0000_0000_0000_0000,
		0b0000_1000_0001_0000_0000_0000_0000_0000,
	)

	is = is || isOpFormat(op,
		0b0000_1110_0001_0000_0000_0000_0000_0000,
		0b0000_1000_0000_0000_0000_0000_0000_0000,
	)

	return is
}

//go:inline
func isHalf(op uint32) bool {

	is := false

	is = is || isOpFormat(op,
		0b0000_1110_0000_0000_0000_0000_1101_0000,
		0b0000_0000_0000_0000_0000_0000_1101_0000,
	)

	is = is || isOpFormat(op,
		0b0000_1110_0000_0000_0000_0000_1011_0000,
		0b0000_0000_0000_0000_0000_0000_1011_0000,
	)

	return is
}

//go:inline
func isALU(op uint32) bool {
	return isOpFormat(op,
		0b0000_1100_0000_0000_0000_0000_0000_0000,
		0b0000_0000_0000_0000_0000_0000_0000_0000,
	)
}

//go:inline
func isBX(op uint32) bool {
	return isOpFormat(op,
		0b0000_1111_1111_1111_1111_1111_1101_0000,
		0b0000_0001_0010_1111_1111_1111_0001_0000,
	)
}

//go:inline
func isB(op uint32) bool {
	return isOpFormat(op,
		0b0000_1110_0000_0000_0000_0000_0000_0000,
		0b0000_1010_0000_0000_0000_0000_0000_0000,
	)
}

//go:inline
func isM(op uint32) bool {

	is := false

	is = is || isOpFormat(op,
		0b0000_1110_1000_0000_0000_0000_1111_0000,
		0b0000_0000_0000_0000_0000_0000_1001_0000,
	)
	is = is || isOpFormat(op,
		0b0000_1110_1000_0000_0000_0000_1111_0000,
		0b0000_0000_1000_0000_0000_0000_1001_0000,
	)

	is = is || isOpFormat(op,
		0b0000_1111_1001_0000_0000_0000_1001_0000,
		0b0000_0001_0000_0000_0000_0000_1000_0000,
	)

	return is
}

//go:inline
func isUD(op uint32) bool {
	return isOpFormat(
		op,
		0b0000_1110_0000_0000_0000_0000_0000_0000,
		0b0000_0110_0000_0000_0000_0000_0000_0000,
	)
}

//go:inline
func isSDT(op uint32) bool {
	is := false
	is = is || isOpFormat(
		op,
		0b0000_1100_0001_0000_0000_0000_0000_0000,
		0b0000_0100_0001_0000_0000_0000_0000_0000,
	)
	is = is || isOpFormat(
		op,
		0b0000_1100_0001_0000_0000_0000_0000_0000,
		0b0000_0100_0000_0000_0000_0000_0000_0000,
	)

	return is
}

//go:inline
func isPLD(op uint32) bool {
	return isOpFormat(
		op,
		0b1111_1100_0001_0000_0000_0000_0000_0000,
		0b1111_0100_0001_0000_0000_0000_0000_0000,
	)
}

//go:inline
func isPSR(op uint32) bool {

	is := false

	is = is || isOpFormat(
		op,
		0b0000_1111_1011_1111_0000_1111_1111_1111,
		0b0000_0001_0000_1111_0000_0000_0000_0000,
	)

	is = is || isOpFormat(
		op,
		0b0000_1101_1011_0000_1111_0000_0000_0000,
		0b0000_0001_0010_0000_1111_0000_0000_0000,
	)

	return is
}
