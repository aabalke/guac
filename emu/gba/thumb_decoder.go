package gba

import (
	"encoding/binary"
	"fmt"
	"log"
)

func (cpu *Cpu) DecodeTHUMB() int {

	r := &cpu.Reg.R
	mem := &cpu.Gba.Mem

	var opcode uint16
	switch r[PC] >> 24 {
	case 0x0:
		opcode = binary.LittleEndian.Uint16(mem.BIOS[r[PC]:])
	case 0x2:
		opcode = binary.LittleEndian.Uint16(mem.WRAM1[r[PC]&0x3FFFF:])
	case 0x3:
		opcode = binary.LittleEndian.Uint16(mem.WRAM2[r[PC]&0x7FFF:])
	case 0x8, 0x9, 0xA, 0xB, 0xC, 0xD:
		opcode = binary.LittleEndian.Uint16(cpu.Gba.Cartridge.Rom[r[PC]&0x1FFFFFF:])
	default:
		log.Printf("Unexpected Arm PC at %08X CURR %d\n", r[PC], CURR_INST)
		opcode = uint16(cpu.Gba.Mem.Read16(r[PC]))
	}

	switch {
	case isthumbSWI(opcode):
		//cpu.Gba.Mem.BIOS_MODE = BIOS_SWI
		//cpu.Gba.exception(VEC_SWI, MODE_SWI)
		//return 2
		cpu.Gba.Mem.BIOS_MODE = BIOS_SWI
		cycles, incPc := cpu.Gba.SysCall(uint32(opcode) & 0xFF)

		if incPc {
			cpu.Reg.R[PC] += 2
		}

		return cycles

	case isThumbAddSub(opcode):
		cpu.ThumbAddSub(opcode)
	case isThumbShift(opcode):
		cpu.thumbShifted(opcode)
	case isThumbImm(opcode):
		cpu.thumbImm(opcode)
	case isThumbAlu(opcode):
		cpu.ThumbAlu(opcode)
	case isThumbHiReg(opcode):
		return cpu.HiRegBX(opcode)
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
	case isShortLongBranch(opcode):
		cpu.thumbShortLongBranch(opcode)
	case isLSSP(opcode):
		cpu.thumbLSSP(opcode)
	case isMulti(opcode):
		cpu.thumbMulti(opcode)
	default:
		r := &cpu.Reg.R
		panic(fmt.Sprintf("Unable to Decode Thumb %X, at PC %X, INSTR %d", opcode, r[PC], CURR_INST))
	}

	return 2
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

func isShortLongBranch(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
		0b1111_1000_0000_0000,
		0b1111_1000_0000_0000,
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

func isthumbSWI(opcode uint16) bool {
	return isThumbOpcodeFormat(opcode,
		0b1111_1111_0000_0000,
		0b1101_1111_0000_0000,
	)
}
