package gba

//  |..3 ..................2 ..................1 ..................0|
//  |1_0_9_8_7_6_5_4_3_2_1_0_9_8_7_6_5_4_3_2_1_0_9_8_7_6_5_4_3_2_1_0|
//  |_Cond__|0_0_0|___Op__|S|__Rn___|__Rd___|__Shift__|Typ|0|__Rm___| DataProc
//  |_Cond__|0_0_0|___Op__|S|__Rn___|__Rd___|__Rs___|0|Typ|1|__Rm___| DataProc
//  |_Cond__|0_0_1|___Op__|S|__Rn___|__Rd___|_Shift_|___Immediate___| DataProc
//  |_Cond__|0_0_1_1_0_0_1_0_0_0_0_0_1_1_1_1_0_0_0_0|_____Hint______| ARM11:Hint
//  |_Cond__|0_0_1_1_0|P|1|0|_Field_|__Rd___|_Shift_|___Immediate___| PSR Imm
//  |_Cond__|0_0_0_1_0|P|L|0|_Field_|__Rd___|0_0_0_0|0_0_0_0|__Rm___| PSR Reg
//  |_Cond__|0_0_0_1_0_0_1_0_1_1_1_1_1_1_1_1_1_1_1_1|0_0|L|1|__Rn___| BX,BLX
//  |1_1_1_0|0_0_0_1_0_0_1_0|_____immediate_________|0_1_1_1|_immed_| ARM9:BKPT
//  |_Cond__|0_0_0_1_0_1_1_0_1_1_1_1|__Rd___|1_1_1_1|0_0_0_1|__Rm___| ARM9:CLZ
//  |_Cond__|0_0_0_1_0|Op_|0|__Rn___|__Rd___|0_0_0_0|0_1_0_1|__Rm___| ARM9:QALU
//  |_Cond__|0_0_0_0_0_0|A|S|__Rd___|__Rn___|__Rs___|1_0_0_1|__Rm___| Multiply
//  |_Cond__|0_0_0_0_0_1_0_0|_RdHi__|_RdLo__|__Rs___|1_0_0_1|__Rm___| ARM11:UMAAL
//  |_Cond__|0_0_0_0_1|U|A|S|_RdHi__|_RdLo__|__Rs___|1_0_0_1|__Rm___| MulLong
//  |_Cond__|0_0_0_1_0|Op_|0|Rd/RdHi|Rn/RdLo|__Rs___|1|y|x|0|__Rm___| MulHalfARM9
//  |_Cond__|0_0_0_1_0|B|0_0|__Rn___|__Rd___|0_0_0_0|1_0_0_1|__Rm___| TransSwp12
//  |_Cond__|0_0_0_1_1|_Op__|__Rn___|__Rd___|1_1_1_1|1_0_0_1|__Rm___| ARM11:LDREX
//  |_Cond__|0_0_0|P|U|0|W|L|__Rn___|__Rd___|0_0_0_0|1|S|H|1|__Rm___| TransReg10
//  |_Cond__|0_0_0|P|U|1|W|L|__Rn___|__Rd___|OffsetH|1|S|H|1|OffsetL| TransImm10
//  |_Cond__|0_1_0|P|U|B|W|L|__Rn___|__Rd___|_________Offset________| TransImm9
//  |_Cond__|0_1_1|P|U|B|W|L|__Rn___|__Rd___|__Shift__|Typ|0|__Rm___| TransReg9
//  |_Cond__|0_1_1|________________xxx____________________|1|__xxx__| Undefined
//  |_Cond__|0_1_1|Op_|x_x_x_x_x_x_x_x_x_x_x_x_x_x_x_x_x_x|1|x_x_x_x| ARM11:Media
//  |1_1_1_1_0_1_0_1_0_1_1_1_1_1_1_1_1_1_1_1_0_0_0_0_0_0_0_1_1_1_1_1| ARM11:CLREX
//  |_Cond__|1_0_0|P|U|S|W|L|__Rn___|__________Register_List________| BlockTrans
//  |_Cond__|1_0_1|L|___________________Offset______________________| B,BL,BLX
//  |_Cond__|1_1_0|P|U|N|W|L|__Rn___|__CRd__|__CP#__|____Offset_____| CoDataTrans
//  |_Cond__|1_1_0_0_0_1_0|L|__Rn___|__Rd___|__CP#__|_CPopc_|__CRm__| CoRR ARM9
//  |_Cond__|1_1_1_0|_CPopc_|__CRn__|__CRd__|__CP#__|_CP__|0|__CRm__| CoDataOp
//  |_Cond__|1_1_1_0|CPopc|L|__CRn__|__Rd___|__CP#__|_CP__|1|__CRm__| CoRegTrans
//  |_Cond__|1_1_1_1|_____________Ignored_by_Processor______________| SWI

func (cpu *Cpu) DecodeARM(opcode uint32) {

    if !cpu.CheckCond(opcode) {
        
        cpu.Reg.R[PC] += 4
        return
    }

	switch {
    case isSWI(opcode): panic("Need SWI Functionality")
    case isB(opcode): cpu.B(opcode)
	case isBX(opcode): panic("Need BX functionality")
    case isSDT(opcode): cpu.Sdt(opcode)
    case isHalf(opcode): cpu.Half(opcode)
	case isBlock(opcode): cpu.Block(opcode)
    case isUD(opcode): panic("Need Undefined functionality")
    case isALU(opcode): cpu.Alu(opcode)
    default: panic("Unable to Decode")
	}
}


func isOpcodeFormat(opcode, mask, format uint32) bool {
    return opcode&mask == format
}

func isBlock(opcode uint32) bool {
    return isOpcodeFormat( opcode,
		0b0000_1110_0000_0000_0000_0000_0000_0000,
		0b0000_1000_0000_0000_0000_0000_0000_0000,
    )
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
