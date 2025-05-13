package gba

import "fmt"

func (cpu *Cpu) DecodeTHUMB(opcode uint16) {

    r := &cpu.Reg.R

	switch {
	case isThumbAddSub(opcode):
		cpu.ThumbAddSub(opcode)
	case isThumbShift(opcode):
        cpu.thumbShifted(opcode)
	case isThumbImm(opcode):
        cpu.thumbImm(opcode)
	case isThumbAlu(opcode):
		cpu.ThumbAlu(opcode)
	case isThumbHiReg(opcode):
		cpu.HiRegBX(opcode)
	case isLSHalf(opcode):
		cpu.thumbLSHalf(opcode)
	case isLSSigned(opcode):
		cpu.thumbLSSigned(opcode)
	case isLPC(opcode):
		cpu.thumbLPC(opcode)
	case isLSR(opcode):
		cpu.thumbLSR(opcode)
	case isLSImm(opcode):
		cpu.thumbLSImm(opcode)
	case isPushPop(opcode):
		cpu.thumbPushPop(opcode)
    case isRelative(opcode):
        cpu.thumbRelative(opcode)
    case isThumbB(opcode):
        cpu.thumbB(opcode)
    case isJumpCall(opcode):
        cpu.thumbJumpCalls(opcode)
    case isStack(opcode):
        cpu.thumbStack(opcode)
    case isLongBranch(opcode):
        cpu.thumbLongBranch(opcode)
    case isLSSP(opcode):
        cpu.thumbLSSP(opcode)
    case isMulti(opcode):
        cpu.thumbMulti(opcode)
	default:
		panic(fmt.Sprintf("Unable to Decode %X, at PC %X, INSTR %d", opcode, r[PC], CURR_INST))
	}
}

func isThumbOpcodeFormat(opcode, mask, format uint16) bool {
	return opcode&mask == format
}

func isThumbShift(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
		0b1110_0000_0000_0000,
		0b0000_0000_0000_0000,
	)
}

func isThumbAddSub(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
		0b1111_1000_0000_0000,
		0b0001_1000_0000_0000,
	)
}

func isThumbImm(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
		0b1110_0000_0000_0000,
		0b0010_0000_0000_0000,
	)
}

func isThumbAlu(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
		0b1111_1100_0000_0000,
		0b0100_0000_0000_0000,
	)
}

func isThumbHiReg(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
		0b1111_1100_0000_0000,
		0b0100_0100_0000_0000,
	)
}

func isLSHalf(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
		0b1111_0000_0000_0000,
		0b1000_0000_0000_0000,
	)
}

func isLSSigned(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
		0b1111_0010_0000_0000,
		0b0101_0010_0000_0000,
	)
}

func isLPC(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
		0b1111_1000_0000_0000,
		0b0100_1000_0000_0000,
	)
}

func isLSR(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
		0b1111_0010_0000_0000,
		0b0101_0000_0000_0000,
	)
}

func isLSImm(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
		0b1110_0000_0000_0000,
		0b0110_0000_0000_0000,
	)
}

func isPushPop(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
        0b1111_0110_0000_0000,
        0b1011_0100_0000_0000,
	)
}

func isRelative(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
        0b1111_0000_0000_0000,
        0b1010_0000_0000_0000,
	)
}

func isJumpCall(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
        0b1111_0000_0000_0000,
        0b1101_0000_0000_0000,
	)
}

func isThumbB(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
        0b1111_1000_0000_0000,
        0b1110_0000_0000_0000,
	)
}

func isStack(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
        0b1111_1111_0000_0000,
        0b1011_0000_0000_0000,
	)
}

func isLongBranch(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
        0b1111_1000_0000_0000,
        0b1111_0000_0000_0000,
	)
}

func isLSSP(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
        0b1111_0000_0000_0000,
        0b1001_0000_0000_0000,
	)
}

func isMulti(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
        0b1111_0000_0000_0000,
        0b1100_0000_0000_0000,
	)
}
