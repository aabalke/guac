package gba

func (cpu *Cpu) DecodeARM(opcode uint32) {

    if !cpu.CheckCond(opcode) {
        cpu.Reg.R[PC] += 4
        return
    }

	switch {
    case isSWI(opcode):     panic("Need SWI Functionality")
    case isB(opcode):       cpu.B(opcode)
	case isBX(opcode):      cpu.BX(opcode)
    case isSDT(opcode):     cpu.Sdt(opcode)
    case isHalf(opcode):    cpu.Half(opcode)
	case isBlock(opcode):   cpu.Block(opcode)
    case isUD(opcode):      panic("Need Undefined functionality")
    case isPSR(opcode):     cpu.Psr(opcode)
    case isALU(opcode):     cpu.Alu(opcode)
    default:                panic("Unable to Decode")
	}
}

func isOpcodeFormat(opcode, mask, format uint32) bool {
    return opcode&mask == format
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
func isCo(opcode uint32) bool {
    return isOpcodeFormat( opcode,
		0b0000_1111_0000_0000_0000_0000_0000_0000,
		0b0000_1110_0000_0000_0000_0000_0000_0000,
    )
}

func isHalf(opcode uint32) bool {

    is := false

    // LDRH
    is = is || isOpcodeFormat( opcode,
        0b0000_1110_0001_0000_0000_0000_1111_0000,
        0b0000_0000_0001_0000_0000_0000_1011_0000,
    )

    // LDRSB
    is = is || isOpcodeFormat( opcode,
        0b0000_1110_0001_0000_0000_0000_1111_0000,
        0b0000_0000_0001_0000_0000_0000_1101_0000,
    )

    // LDRSH
    is = is || isOpcodeFormat( opcode,
        0b0000_1110_0001_0000_0000_0000_1111_0000,
        0b0000_0000_0001_0000_0000_0000_1111_0000,
    )

    // STRH
    is = is || isOpcodeFormat( opcode,
        0b0000_1110_0001_0000_0000_0000_1111_0000,
        0b0000_0000_0000_0000_0000_0000_1011_0000,
    )

    return is
}

func isALU(opcode uint32) bool {
    return isOpcodeFormat( opcode,
		0b0000_1100_0000_0000_0000_0000_0000_0000,
		0b0000_0000_0000_0000_0000_0000_0000_0000,
	)
}

func isBX(opcode uint32) bool {
    return isOpcodeFormat( opcode,
		0b0000_1111_1111_1111_1111_1111_1111_0000,
		0b0000_0001_0010_1111_1111_1111_0001_0000,
	)
}

func isB(opcode uint32) bool {
    return isOpcodeFormat( opcode,
		0b0000_1110_0000_0000_0000_0000_0000_0000,
		0b0000_1010_0000_0000_0000_0000_0000_0000,
	)
}

func isM(opcode uint32) bool {
    return isOpcodeFormat( opcode,
        0b0000_1110_1000_0000_0000_0000_1111_0000,
		0b0000_0000_0000_0000_0000_0000_1001_0000,
	)
}

func isML(opcode uint32) bool {
    return isOpcodeFormat( opcode,
		0b0000_1110_1000_0000_0000_0000_1111_0000,
		0b0000_0000_1000_0000_0000_0000_1001_0000,
	)
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
    return isOpcodeFormat(
        opcode,
		0b0000_1100_0000_0000_0000_0000_0000_0000,
		0b0000_0100_0000_0000_0000_0000_0000_0000,
    )
}

func isDP(opcode uint32) bool {
    return isOpcodeFormat(
        opcode,
		0b0000_0010_0000_0000_0000_0000_0000_0000,
		0b0000_0010_0000_0000_0000_0000_0000_0000,
    )
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
