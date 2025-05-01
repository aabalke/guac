package gba

//import "fmt"

func (cpu *Cpu) DecodeTHUMB(opcode uint16) {

	// initial switch 0x803155C, 40D3

    //fmt.Printf("OP: %X %b\n", opcode, opcode >> 12)

	switch {
	case isThumbAddSub(opcode):
        cpu.ThumbAddSub(opcode)
	case isThumbShift(opcode):
		panic("SHIFTED")
	case isThumbImm(opcode):
		panic("IMMEDIATE")
	case isThumbAlu(opcode):
        cpu.ThumbAlu(opcode)
	case isThumbHiReg(opcode):
        cpu.HiRegBX(opcode)
	default:
		panic("UNKNOWN THUMB OPCODE")
	}

	//	if !cpu.CheckCond(opcode) {
	//		cpu.Reg.R[PC] += 4
	//		return
	//	}
	//
	//	switch {
	//	case isSWI(opcode):
	//		panic("Need SWI Functionality")
	//	case isB(opcode):
	//		cpu.B(opcode)
	//	case isBX(opcode):
	//		cpu.BX(opcode)
	//	case isSDT(opcode):
	//		cpu.Sdt(opcode)
	//	case isHalf(opcode):
	//		cpu.Half(opcode)
	//	case isBlock(opcode):
	//		cpu.Block(opcode)
	//	case isUD(opcode):
	//		panic("Need Undefined functionality")
	//	case isPSR(opcode):
	//		cpu.Psr(opcode)
	//	case isSWP(opcode):
	//		cpu.Swp(opcode)
	//    case isM(opcode):
	//        cpu.Mul(opcode)
	//	case isALU(opcode):
	//		cpu.Alu(opcode)
	//	default:
	//		panic("Unable to Decode")
	//	}
	//
	//    // Notes: Coprocessor instructions do not matter since gba
	//    // uses a single processor (NDS is a different story)

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
