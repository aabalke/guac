package gba

import (
	"fmt"

	"github.com/aabalke/guac/emu/gba/utils"
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

type Alu struct {
	Opcode, Rd, Rn, Rm, RnValue, Op2, Inst uint32
	Immediate, Set                         bool
	LogicalFlags, Test                     bool
	Carry                                  bool
}

var aluData Alu

func NewAluData(opcode uint32, cpu *Cpu) *Alu {

	aluData.Opcode = opcode
	aluData.Inst = utils.GetByte(opcode, 21)
	aluData.Immediate = utils.BitEnabled(opcode, 25)
	aluData.Set = utils.BitEnabled(opcode, 20)
	aluData.Rd = utils.GetByte(opcode, 12)
	aluData.Rn = utils.GetByte(opcode, 16)
	aluData.Rm = utils.GetByte(opcode, 0)
	aluData.LogicalFlags = false
	aluData.Test = false

	aluData.Op2, aluData.Carry = cpu.GetOp2(opcode)

	if aluData.Rn != PC {
		aluData.RnValue = cpu.Reg.R[aluData.Rn]
		return &aluData
	}

	shiftImmediate := aluData.Immediate || !utils.BitEnabled(opcode, 4)
	if shiftImmediate {
		aluData.RnValue = cpu.Reg.R[PC] + 8
		return &aluData
	}

	aluData.RnValue = cpu.Reg.R[PC] + 12

	return &aluData
}

func (cpu *Cpu) Alu(opcode uint32) {

	alu := NewAluData(opcode, cpu)

	switch alu.Inst {
	case AND, EOR, ORR, MOV, MVN, BIC:
		alu.LogicalFlags = true
		alu.Test = false
		cpu.logical(alu)
	case ADD, ADC, SUB, SBC, RSB, RSC:
		alu.LogicalFlags = false
		alu.Test = false
		cpu.arithmetic(alu)
	case TST, TEQ:
		alu.Test = true
		alu.LogicalFlags = true
		cpu.test(alu)
	case CMP, CMN:
		alu.Test = true
		alu.LogicalFlags = false
		cpu.test(alu)
	}

	if alu.Rd != PC {
		cpu.Reg.R[15] += 4
	}
}

func (cpu *Cpu) GetOp2(opcode uint32) (uint32, bool) {

	reg := &cpu.Reg

	isCarry := utils.BitEnabled(opcode, 20)
	currCarry := reg.CPSR.GetFlag(FLAG_C)

	if immediate := utils.BitEnabled(opcode, 25); immediate {

		nn := utils.GetVarData(opcode, 0, 7)
		ro := utils.GetVarData(opcode, 8, 11) * 2
		op2, setCarry, carry := utils.Ror(nn, ro, isCarry, false, currCarry)

		if setCarry {
			reg.CPSR.SetFlag(FLAG_C, carry)
		}

		return op2, currCarry
	}

	is := utils.GetVarData(opcode, 7, 11)
	var additional uint32
	rm := utils.GetByte(opcode, 0)
	if rm == PC {
		additional += 8
	}

	shiftRegister := utils.BitEnabled(opcode, 4)
	if shiftRegister {
		is = reg.R[(opcode>>8)&0b1111] & 0b1111_1111

		if rm == PC {
			additional += 4
		}

	}

	shiftArgs := utils.ShiftArgs{
		SType:     opcode >> 5 & 0b11,
		Val:       reg.R[rm] + additional,
		Is:        is,
		IsCarry:   isCarry,
		Immediate: !shiftRegister,
		CurrCarry: currCarry,
	}

	op2, setCarry, carry := utils.Shift(&shiftArgs)

	if setCarry {
		reg.CPSR.SetFlag(FLAG_C, carry)
	}

	return op2, currCarry
}

func (cpu *Cpu) logical(alu *Alu) {

    var res uint32
    switch alu.Inst {
    case AND:
        res = alu.RnValue & alu.Op2
    case EOR:
        res = alu.RnValue ^ alu.Op2
    case ORR:
        res = alu.RnValue | alu.Op2
    case MOV:
        res = alu.Op2
    case MVN:
        res = ^alu.Op2
    case BIC:
        res = alu.RnValue &^ alu.Op2
    }

	cpu.Reg.R[alu.Rd] = res

	cpu.setAluFlags(alu, uint64(res))
}

func (cpu *Cpu) arithmetic(alu *Alu) {

	carry := uint64(0)
	if alu.Carry {
		carry = 1
	}

    var res uint64

    switch alu.Inst {
    case ADD: res = uint64(alu.RnValue) + uint64(alu.Op2)
    case ADC: res = uint64(alu.RnValue) + uint64(alu.Op2) + carry
    case SUB: res = uint64(alu.RnValue) - uint64(alu.Op2)
    case SBC: res = uint64(alu.RnValue) - uint64(alu.Op2) + carry - 1
    case RSB: res = uint64(alu.Op2) - uint64(alu.RnValue)
    case RSC: res = uint64(alu.Op2) - uint64(alu.RnValue) + carry - 1
    }

	cpu.Reg.R[alu.Rd] = uint32(res)

	cpu.setAluFlags(alu, res)
}

func (cpu *Cpu) test(alu *Alu) {

    var res uint64
    switch alu.Inst {
    case TST: res = uint64(alu.RnValue) & uint64(alu.Op2)
    case TEQ: res = uint64(alu.RnValue) ^ uint64(alu.Op2)
    case CMP: res = uint64(alu.RnValue) - uint64(alu.Op2)
    case CMN: res = uint64(alu.RnValue) + uint64(alu.Op2)
    }

	//incPc := true

	switch alu.Inst {
	case TST, TEQ:
		cpu.setAluFlags(alu, uint64(uint32(res)))
	case CMP, CMN:
		cpu.setAluFlags(alu, res)
	}

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

    curr := cpu.Reg.getMode()

	i := BANK_ID[curr]
	reg.CPSR = reg.SPSR[i]
    reg.isThumb = reg.CPSR.GetFlag(FLAG_T)

    next := cpu.Reg.getMode()
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

func (cpu *Cpu) setAluFlags(alu *Alu, res uint64) {

    //    Returned CPSR Flags
    //If S=1, Rd<>R15, logical operations (AND,EOR,TST,TEQ,ORR,MOV,BIC,MVN):
    //  V=not affected
    //  C=carryflag of shift operation (not affected if LSL#0 or Rs=00h)
    //  Z=zeroflag of result
    //  N=signflag of result (result bit 31)
    //If S=1, Rd<>R15, arithmetic operations (SUB,RSB,ADD,ADC,SBC,RSC,CMP,CMN):
    //  V=overflowflag of result
    //  C=carryflag of result
    //  Z=zeroflag of result
    //  N=signflag of result (result bit 31)
    //IF S=1, with unused Rd bits=1111b, {P} opcodes (CMPP/CMNP/TSTP/TEQP):
    //  R15=result  ;modify PSR bits in R15, ARMv2 and below only.
    //  In user mode only N,Z,C,V bits of R15 can be changed.
    //  In other modes additionally I,F,M1,M0 can be changed.
    //  The PC bits in R15 are left unchanged in all modes.
    //If S=1, Rd=R15; should not be used in user mode:
    //  CPSR = SPSR_<current mode>
    //  PC = result
    //  For example: MOVS PC,R14  ;return from SWI (PC=R14_svc, CPSR=SPSR_svc).
    //If S=0: Flags are not affected (not allowed for CMP,CMN,TEQ,TST).

	if !alu.Set {
		return
	}

	if irqExit := alu.Rd == PC && alu.Rn == LR && alu.Inst == SUB; irqExit {
		cpu.Gba.ExitException(MODE_IRQ)
		return
	}

	if swiExit := alu.Rd == PC && alu.Rm == LR && alu.Inst == MOV; swiExit {
		cpu.Gba.ExitException(MODE_SWI)
		return
	}

    if forceExit := alu.Rd == PC; forceExit {
        cpu.psrSwitch()
        return
    }

	if alu.LogicalFlags {
		cpu.Reg.CPSR.SetFlag(FLAG_N, utils.BitEnabled(uint32(res), 31))
		cpu.Reg.CPSR.SetFlag(FLAG_Z, uint32(res) == 0)
		return
	}

	var v, c bool
	rnSign := uint8(alu.RnValue>>31) & 1
	opSign := uint8(alu.Op2>>31) & 1
	rSign := uint8(res>>31) & 1

	switch alu.Inst {
	case ADD, ADC, CMN:
		v = (rnSign == opSign) && (rSign != rnSign)
		c = res >= 0x1_0000_0000
	case SUB, SBC, CMP:
		v = (rnSign != opSign) && (rSign != rnSign)
		c = res < 0x1_0000_0000

	case RSB, RSC:
		v = (rnSign != opSign) && (rSign != opSign)
		c = res < 0x1_0000_0000
	}

	cpu.Reg.CPSR.SetFlag(FLAG_V, v)
	cpu.Reg.CPSR.SetFlag(FLAG_C, c)
	cpu.Reg.CPSR.SetFlag(FLAG_N, utils.BitEnabled(uint32(res), 31))
	cpu.Reg.CPSR.SetFlag(FLAG_Z, uint32(res) == 0)
	return
}

const (
	MUL   = 0b0
	MLA   = 0b1
	UMAAL = 0b010
	UMULL = 0b100
	UMLAL = 0b101
	SMULL = 0b110
	SMLAL = 0b111
)

func (cpu *Cpu) Mul(opcode uint32) {

	inst := utils.GetByte(opcode, 21)
	set := utils.BitEnabled(opcode, 20)
	rd := utils.GetByte(opcode, 16)
	rn := utils.GetByte(opcode, 12)
	rs := utils.GetByte(opcode, 8)
	rm := utils.GetByte(opcode, 0)
	r := &cpu.Reg.R

	if mulHalf := inst == MUL || inst == MLA; mulHalf {

		res := r[rm] * r[rs]

		if inst == MLA {
			res += r[rn]
		}

		r[rd] = res

		if set {
			cpu.Reg.CPSR.SetFlag(FLAG_Z, res == 0)
			cpu.Reg.CPSR.SetFlag(FLAG_N, (res>>31&0b1) != 0)
			// FLAG_C "destroyed" ARM <5, ignored ARM >=5
			cpu.Reg.CPSR.SetFlag(FLAG_C, false)
		}

		r[PC] += 4
		return
	}

	if inst == UMAAL {
		panic("UMAAL is UNSUPPORTED")
	}

	if mulUnsignedWord := inst == UMULL || inst == UMLAL; mulUnsignedWord {
		res := uint64(r[rm]) * uint64(r[rs])

		if inst == UMLAL {
			res += uint64(r[rd])<<32 | uint64(r[rn])
		}

		r[rd] = uint32(res >> 32)
		r[rn] = uint32(res)

		if set {
			//cpu.Reg.CPSR.SetFlag(FLAG_N, (res >> 63 & 0b1) != 0)
			cpu.Reg.CPSR.SetFlag(FLAG_N, (res>>63&1) == 1)
			cpu.Reg.CPSR.SetFlag(FLAG_Z, res == 0)
			// FLAG_C "destroyed" ARM <5, ignored ARM >=5
			// need carry to pass mgba suite
			//c := res >= 0x1_0000_0000
			//cpu.Reg.CPSR.SetFlag(FLAG_C, c)
			cpu.Reg.CPSR.SetFlag(FLAG_C, false)
			// FLAG_V maybe destroyed on ARM <5. ignored ARM <=5
		}

		r[PC] += 4
		return
	}

	res := int64(int32(r[rm])) * int64(int32(r[rs]))
	if inst == SMLAL {
		res += int64(r[rd])<<32 | int64(r[rn])
	}

	r[rd] = uint32(res >> 32)
	r[rn] = uint32(res)

	if set {
		cpu.Reg.CPSR.SetFlag(FLAG_N, (res>>63&1) == 1)
		cpu.Reg.CPSR.SetFlag(FLAG_Z, res == 0)
		// FLAG_C "destroyed" ARM <5, ignored ARM >=5
		cpu.Reg.CPSR.SetFlag(FLAG_C, false)
		// FLAG_V maybe destroyed on ARM <5. ignored ARM <=5
	}

	r[PC] += 4
}

const (
	STR = iota
	LDR_PLD
)

var sdtInstance Sdt

type Sdt struct {
	Opcode, Rd, Rn, RnValue, RdValue, Offset, Shift, ShiftType, Rm uint32
	Set, I, Load, WriteBack, MemoryMgmt, Pre, Up, Byte, Pld              bool
}

func NewSdtData(opcode uint32, cpu *Cpu) *Sdt {
	valid := utils.GetVarData(opcode, 26, 27) == 0b01
	if !valid {
		panic("Malformed Sdt Instruction")
	}

	sdtInstance = Sdt{
		Opcode: opcode,
		I:      utils.BitEnabled(opcode, 25),
		Pre:    utils.BitEnabled(opcode, 24),
		Up:     utils.BitEnabled(opcode, 23),
		Byte:   utils.BitEnabled(opcode, 22),
		Load:   utils.BitEnabled(opcode, 20),
		Rn:     utils.GetByte(opcode, 16),
		Rd:     utils.GetByte(opcode, 12),
	}

	sdt := &sdtInstance

	if sdt.Pre {
		sdt.WriteBack = utils.BitEnabled(opcode, 21)
	} else {
		sdt.MemoryMgmt = utils.BitEnabled(opcode, 21)
	}

	if sdt.I {
		sdt.Shift = utils.GetVarData(opcode, 7, 11)
		sdt.ShiftType = utils.GetVarData(opcode, 5, 6)

		if utils.BitEnabled(opcode, 4) {
			panic("Malformed Single Data Transfer")
		}

		sdt.Rm = utils.GetByte(opcode, 0)
	} else {
		sdt.Offset = utils.GetVarData(opcode, 0, 11)
	}

	sdt.RdValue = cpu.Reg.R[sdt.Rd]
	sdt.RnValue = cpu.Reg.R[sdt.Rn]

	sdt.Pld = (opcode >> 28) == 0b1111 &&
		sdt.Pre == true &&
		sdt.Byte == true &&
		sdt.WriteBack == false &&
		sdt.Load == true &&
		sdt.Rd == 0b1111

	return sdt
}

func (c *Cpu) Sdt(opcode uint32) uint32 {

	r := &c.Reg.R

	sdt := NewSdtData(opcode, c)

	pre, post, _ := generateSdtAddress(sdt, c)

	addr := pre &^ 0b11

	if sram := addr >= 0xE00_0000 && addr < 0x1000_0000; sram {
		addr = pre
	}

	if sdt.Pld {
		panic("Need to handle PLD Inst")
	}

	switch {
	case sdt.Load && sdt.Byte:

		// DO NOT WORD ALIGN
		r[sdt.Rd] = uint32(c.Gba.Mem.Read8(pre))

	case sdt.Load && !sdt.Byte:

		v := c.Gba.Mem.Read32(addr)
		is := (pre & 0b11) << 3
        v = utils.RorSimple(v, is)
		//v, _, _ = utils.Ror(v, is, false, false, false)

		if sdt.Rd == PC { // not sure if this is right
			v -= 4
		}

		r[sdt.Rd] = v

	case !sdt.Load && sdt.Byte:

		c.Gba.Mem.Write8(pre, uint8(r[sdt.Rd]))

	case !sdt.Load && !sdt.Byte:

		v := r[sdt.Rd]
		if sdt.Rd == PC {
			v += 12
		}

		c.Gba.Mem.Write32(addr, v)
	}

	skipLoadWriteBack := sdt.Load && (sdt.Rn == sdt.Rd)

	if (sdt.WriteBack || !sdt.Pre) && !skipLoadWriteBack {
		r[sdt.Rn] = post
	}

	c.Reg.R[PC] += 4

	return 4
}

func generateSdtAddress(sdt *Sdt, cpu *Cpu) (pre uint32, post uint32, writeBack bool) {

	r := &cpu.Reg.R

	var offset uint32
	if !sdt.I {
		offset = sdt.Offset
	} else {
		shift := sdt.Opcode >> 7 & 0b11111

		rm := sdt.Opcode & 0b1111

		shiftArgs := utils.ShiftArgs{
			SType:     sdt.Opcode >> 5 & 0b11,
			Val:       r[rm],
			Is:        shift,
			IsCarry:   false,
			Immediate: true,
			CurrCarry: cpu.Reg.CPSR.GetFlag(FLAG_C),
		}

		offset, _, _ = utils.Shift(&shiftArgs)
	}

	addr := r[sdt.Rn]
	if sdt.Rn == PC {
		addr += 8
	}
	if sdt.Up {
		addr += offset

	} else {
		addr -= offset
	}

	if sdt.Pre {
		if offset == 0 {
			// this may be a problem, post addr was a problem on ldm
			return addr, 0, false
			//return addr, addr, false
		}
		return addr, addr, true
	}

	return r[sdt.Rn], addr, false
}

func (cpu *Cpu) B(opcode uint32) {

	isLink := utils.BitEnabled(opcode, 24)

	r := &cpu.Reg.R

	if isLink {
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

	switch inst {
	case INST_BX:
		cpu.Reg.R[PC] = cpu.Reg.R[rn]
		cpu.Gba.toggleThumb()
	case INST_BXJ:
		panic("Unsupported BXJ Instruction")
	case INST_BLX:
		panic("Unsupported BLX Instruction")
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
	Rn, Rd, Imm, Inst, Rm, RdValue, RnValue, RmValue uint32
	Pre, Up, Immediate, WriteBack, Load, MemoryManagement  bool
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
			panic(fmt.Sprintf("Malformed Half Instruction %d %08X %d", i, opcode, CURR_INST))
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
			panic("LDRD NOT SUPPORTED")
		case STRD:
			panic("STRD NOT SUPPORTED")
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

func signedByteStd(half *Half, cpu *Cpu) {

	r := &cpu.Reg.R
	pre, post := halfUnsignedAddress(half, cpu)
	addr := pre &^ 0b1

	if half.Load {
		// sign-expand byte value
		unexpanded := int8(cpu.Gba.Mem.Read8(pre))
		expanded := uint32(unexpanded)

		if unexpanded < 0 {
			expanded |= (0xFFFFFF << 8)
		}

		r[half.Rd] = expanded
	} else {
		cpu.Gba.Mem.Write16(addr, uint16(int16(half.RdValue)))
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
		// On ARM7 aka ARMv4 aka NDS7/GBA:
		// LDRSH Rd,[odd]  -->  LDRSB Rd,[odd]         ;sign-expand BYTE value
		// On ARM9 aka ARMv5 aka NDS9:
		// LDRSH Rd,[odd]  -->  LDRSH Rd,[odd-1]       ;forced align

		if misaligned := pre&1 == 1; misaligned {
			// sign-expand BYTE value
			unexpanded := int16(cpu.Gba.Mem.Read16(pre))
			expanded := uint32(unexpanded)

			if int8(unexpanded) < 0 {
				expanded |= (0xFFFFFF << 8)
			} else {
				expanded &= 0xFF
			}

			r[half.Rd] = expanded
		} else {

			// sign-expand half value
			unexpanded := int16(cpu.Gba.Mem.Read16(pre &^ 0b1))
			expanded := uint32(unexpanded)

			if unexpanded < 0 {
				expanded |= (0xFFFF << 16)
			}

			r[half.Rd] = expanded
		}
	} else {
		addr := pre &^ 0b1
		cpu.Gba.Mem.Write16(addr, uint16(int16(half.RdValue)))
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

	if sram := addr >= 0xE00_0000 && addr < 0x1000_0000; sram {
		addr = pre
	}

	if half.Load {
		v := uint32(cpu.Gba.Mem.Read16(addr))
		is := (pre & 0b1) << 3
        v = utils.RorSimple(v, is)
		//v, _, _ = utils.Ror(v, is, false, false, false)
		r[half.Rd] = v
	} else {
		cpu.Gba.Mem.Write16(addr, uint16(half.RdValue))
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

type Block struct {
	Opcode, Rn, RnValue, Rlist uint32
	Pre, Up, PSR, Writeback, Load    bool
}

func (c *Cpu) Block(opcode uint32) {

	r := &c.Reg.R

	block := &Block{
		Opcode:    opcode,
		Pre:       utils.BitEnabled(opcode, 24),
		Up:        utils.BitEnabled(opcode, 23),
		PSR:       utils.BitEnabled(opcode, 22),
		Writeback: utils.BitEnabled(opcode, 21),
		Load:      utils.BitEnabled(opcode, 20),
		Rn:        utils.GetByte(opcode, 16),
		Rlist:     utils.GetVarData(opcode, 0, 15),
	}

	block.RnValue = c.Reg.R[block.Rn]

	incPc := true

	mode := c.Reg.getMode()

	if forceUser := block.PSR; forceUser && mode != MODE_USR {
		c.Reg.setMode(mode, MODE_USR)
	}

	if block.Load {
		incPc = c.ldm(block)
	} else {
		c.stm(block)
	}

	if !block.Writeback {
		r[block.Rn] = block.RnValue
	}

	if forceUser := block.PSR; forceUser && mode != MODE_USR {
		c.Reg.setMode(MODE_USR, mode)
	}

	if utils.BitEnabled(block.Opcode, 15) && block.PSR {
		panic("LDM WITH R15 AND SET USED")
		return
	}

	if incPc {
		c.Reg.R[PC] += 4
	}
}

func (c *Cpu) ldm(block *Block) bool {

	incPC := true
	r := &c.Reg.R

	ib := block.Pre && block.Up
	ia := !block.Pre && block.Up
	db := block.Pre && !block.Up
	da := !block.Pre && !block.Up

	if block.Rlist == 0 {

		c.Reg.R[PC] += 16 // i believe this is short cut for {} => {r15} behavior

		if block.Up {
			r[block.Rn] += 0x40
			return false
		}

		r[block.Rn] -= 0x40
		return false
	}

	regCount := utils.CountBits(block.Rlist)

	if (block.Rlist>>block.Rn)&1 == 1 {
		regCount--
		block.Writeback = false
	}

	addr := r[block.Rn] &^ 0b11

	reg := uint32(0)
	for reg = range 16 {

		regBitEnabled := utils.BitEnabled(block.Rlist, uint8(reg))
		decRegBitEnabled := utils.BitEnabled(block.Rlist, uint8(15-reg))

		switch {
		case ib && regBitEnabled:

			addr += 4
			r[reg] = c.Gba.Mem.Read32(addr)

			if reg == PC {
				incPC = incPC && false
			}
			if reg == block.Rn {
				block.RnValue = r[block.Rn]
			}

		case ia && regBitEnabled:

			r[reg] = c.Gba.Mem.Read32(addr)

			if reg == PC {
				incPC = incPC && false
			}
			if reg == block.Rn {
				block.RnValue = r[block.Rn]
			}

			addr += 4

		case db && decRegBitEnabled: // pop

			addr -= 4
			r[15-reg] = c.Gba.Mem.Read32(addr)

			if 15-reg == PC {
				incPC = incPC && false
			}
			if 15-reg == block.Rn {
				block.RnValue = r[block.Rn]
			}

		case da && decRegBitEnabled:

			r[15-reg] = c.Gba.Mem.Read32(addr)
			addr -= 4

			if 15-reg == PC {
				incPC = incPC && false
			}
			if 15-reg == block.Rn {
				block.RnValue = r[block.Rn]
			}
		}
	}

	if block.Up {
		r[block.Rn] += (regCount * 4)
	} else {
		r[block.Rn] -= (regCount * 4)
	}

	return incPC
}

func (c *Cpu) stm(block *Block) {

	r := &c.Reg.R

	ib := block.Pre && block.Up
	ia := !block.Pre && block.Up
	db := block.Pre && !block.Up
	da := !block.Pre && !block.Up

	if block.Rlist == 0 {
		// stm {} => {PC}
		switch {
		case ib:
			addr := r[block.Rn] + 4
			r[block.Rn] += 0x40
			c.Gba.Mem.Write32(addr, r[PC]+12)
		case ia:
			c.Gba.Mem.Write32(r[block.Rn], r[PC]+12)
			r[block.Rn] += 0x40
		case db:
			r[block.Rn] -= 0x40
			c.Gba.Mem.Write32(r[block.Rn], r[PC]+12)
		case da:
			r[block.Rn] -= 0x40
			c.Gba.Mem.Write32(r[block.Rn]+4, r[PC]+12)
		}

		return
	}

	regCount := utils.CountBits(block.Rlist)

	smallest := (block.Rlist & -block.Rlist) == 1<<block.Rn
	matchingRn := (block.Rlist>>block.Rn)&1 == 1
	matchingValue := uint32(0)
	matchingAddr := uint32(0) // rn during regs

	addr := r[block.Rn] &^ 0b11

	count := uint32(0)
	rnIdx := uint32(0)
	for reg := range 16 {

		regBitEnabled := utils.BitEnabled(block.Rlist, uint8(reg))
		decRegBitEnabled := utils.BitEnabled(block.Rlist, uint8(15-reg))

		switch {
		case ib && regBitEnabled:

			count++

			r[block.Rn] += 4
			addr += 4

			if reg == int(block.Rn) {
				c.Gba.Mem.Write32(addr, r[reg]-4)
				matchingValue = r[reg]
				matchingAddr = addr
				rnIdx = regCount - count
				continue
			}

			if reg == PC {
				c.Gba.Mem.Write32(addr, r[reg]+12)
				continue
			}

			c.Gba.Mem.Write32(addr, r[reg])

		case ia && regBitEnabled:

			count++

			if reg == int(block.Rn) {

				c.Gba.Mem.Write32(addr, r[reg])
				matchingValue = r[reg] + 4
				matchingAddr = addr
				rnIdx = regCount - count
				r[block.Rn] += 4
				addr += 4
				continue
			}

			if reg == PC {
				c.Gba.Mem.Write32(addr, r[reg]+12)
				continue
			}

			c.Gba.Mem.Write32(addr, r[reg])

			r[block.Rn] += 4
			addr += 4

		case db && decRegBitEnabled: // push
			count++

			r[block.Rn] -= 4
			addr -= 4

			if 15-reg == int(block.Rn) {
				matchingValue = r[15-reg]
				matchingAddr = addr
				rnIdx = regCount - count // regCount only for 15 - reg
			}
			if 15-reg == PC {
				c.Gba.Mem.Write32(addr, r[15-reg]+12)
				continue
			}

			c.Gba.Mem.Write32(addr, r[15-reg])

		case da && decRegBitEnabled:

			count++

			decReg := 15 - reg

			if decReg == int(block.Rn) {
				c.Gba.Mem.Write32(addr, r[decReg]+(count-1)*4)
				matchingValue = r[decReg] - 4 // -4 offsets above +4 when matching Value (not first smallest)
				matchingAddr = addr
				rnIdx = regCount - count
				r[block.Rn] -= 4
				addr -= 4
				continue
			}

			if decReg == PC {
				c.Gba.Mem.Write32(addr, r[decReg]+12)
				continue
			}

			c.Gba.Mem.Write32(addr, r[decReg])

			r[block.Rn] -= 4
			addr -= 4
		}
	}

	if block.Writeback && smallest {

		v := c.Gba.Mem.Read32(addr)

		if block.Up {
			c.Gba.Mem.Write32(r[block.Rn], v-(regCount*4))
			return
		}
		c.Gba.Mem.Write32(r[block.Rn], v+(regCount*4))
		return
	}

	if block.Writeback && matchingRn {
		if block.Up {
			c.Gba.Mem.Write32(matchingAddr, matchingValue+(rnIdx*4))
			return
		}

		c.Gba.Mem.Write32(matchingAddr, matchingValue-(rnIdx*4))
		return
	}
}

type PSR struct {
	Opcode, Rd, Rm, Shift, Imm uint32
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
		mode := cpu.Reg.getMode()
		r[psr.Rd] = uint32(cpu.Reg.SPSR[BANK_ID[mode]])
		return
	}

	mask := PRIV_MASK
	if cpu.Reg.getMode() == MODE_USR {
		mask = USR_MASK
	}

	r[psr.Rd] = uint32(cpu.Reg.CPSR) & mask
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
	curr := cpu.Reg.getMode()
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
			spsr = uint32(reg.CPSR) &^ mask
		} else {
			spsr = uint32(reg.SPSR[BANK_ID[curr]]) &^ mask
		}

		spsr |= v & mask
		reg.SPSR[BANK_ID[curr]] = Cond(spsr)

		return
	}

	next := v & 0b11111
	cpsr := uint32(reg.CPSR) &^ mask

	cpsr |= v & mask

	reg.CPSR = Cond(cpsr)
    reg.isThumb = reg.CPSR.GetFlag(FLAG_T)

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

	aligned := rnValue

	var rnMemValue uint32
	if isByte {
		rnMemValue = cpu.Gba.Mem.Read8(rnValue)
		r[rd] = rnMemValue
		cpu.Gba.Mem.Write8(rnValue, uint8(rmValue))
		r[PC] += 4
		return

	} else {
		aligned = rnValue &^ 0b11
		rnMemValue = cpu.Gba.Mem.Read32(aligned)
		is := (rnValue & 0b11) << 3
		//rnMemValue, _, _ = utils.Ror(rnMemValue, is, false, false, false)
        rnMemValue = utils.RorSimple(rnMemValue, is)
	}

	r[rd] = rnMemValue
	cpu.Gba.Mem.Write32(aligned, rmValue)
	r[PC] += 4
}
