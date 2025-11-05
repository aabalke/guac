//package nds
//
//import (
//	"fmt"
//	"os"
//
//	"github.com/aabalke/guac/emu/nds/debug"
//)
//
//var _ = os.Args
//
//type Debugger struct {
//	nds *Nds
//}
//
//func (d *Debugger) PrintLine(arm9 bool) {
//
//	if !arm9 {
//
//        pc := d.nds.arm7.Reg.R[15]
//        r := &d.nds.arm7.Reg.R
//        cpsr := d.nds.arm7.Reg.CPSR
//        opcode := d.nds.mem.Read32(pc, false)
//        mem := &d.nds.mem
//
//        _ = r[15]
//        _ = mem.Read(0, true)
//        _ = cpsr
//
//        var desc, op string
//
//        if d.nds.arm7.Reg.IsThumb {
//            opcode &= 0xFFFF
//            desc = d.DecodeThumb(uint16(opcode))
//            op = fmt.Sprintf("OP     %04X", opcode)
//        } else {
//            desc = d.DecodeArm(opcode)
//            op = fmt.Sprintf("OP %08X", opcode)
//        }
//
//    //fmt.Printf("CURR %5d ARM7: PC %08X %12s %-12s R0 %08X R1 %08X R2 %08X R3 %08X R4 %08X R5 %08X R6 %08X R7 %08X R8 %08X R9 %08X R10 %08X R11 %08X R12 %08X SP %08X LR %08X 0x180 %08X 0x184 %08X FIFO %08X CPSR %08X\n", CURR_INST, pc, op, desc, r[0], r[1], r[2], r[3], r[4], r[5], r[6], r[7], r[8], r[9], r[10], r[11], r[12], r[13], r[14], d.nds.mem.Read32(0x400_0180, false), d.nds.mem.Read32(0x400_0184, false), d.nds.mem.Ipc.Fifo9.Value, cpsr)
//        fmt.Printf("CURR %5d ARM7: PC %08X %12s %-12s R0 %08X R1 %08X R2 %08X R3 %08X R4 %08X R5 %08X R7 %08X R11 %08X R12 %08X SP %08X LR %08X CPSR %08X\n", debug.CURR_INST, pc, op, desc, r[0], r[1], r[2], r[3], r[4], r[5], r[7], r[11], r[12], r[13], r[14], cpsr)
//
//        return
//	}
//
//	pc := d.nds.arm9.Reg.R[15]
//	r := &d.nds.arm9.Reg.R
//	cpsr := d.nds.arm9.Reg.CPSR
//	opcode := d.nds.mem.Read32(pc, true)
//	mem := &d.nds.mem
//
//	_ = r[15]
//	_ = mem.Read(0, true)
//	_ = cpsr
//
//	var desc, op string
//
//	if d.nds.arm9.Reg.IsThumb {
//		opcode &= 0xFFFF
//		desc = d.DecodeThumb(uint16(opcode))
//		op = fmt.Sprintf("OP     %04X", opcode)
//	} else {
//		desc = d.DecodeArm(opcode)
//		op = fmt.Sprintf("OP %08X", opcode)
//	}
//    
//    fmt.Printf("CURR %5d ARM9: PC %08X %12s %-12s R0 %08X R1 %08X R2 %08X R3 %08X R4 %08X R5 %08X R6 %08X R7 %08X R11 %08X R12 %08X SP %08X LR %08X CPSR %08X\n", debug.CURR_INST, pc, op, desc, r[0], r[1], r[2], r[3], r[4], r[5], r[6], r[7], r[11], r[12], r[13], r[14], cpsr)
//
//}
//
//
//func (d *Debugger) print(i int) {
//	reg := &d.nds.arm9.Reg
//	p := func(a string, b uint32) { fmt.Printf("% 8s: % 9X\n", a, b) }
//	s := func(a string) { fmt.Printf("%s\n", a) }
//
//	s("--------  --------")
//	fmt.Printf("inst dec %d\n", uint32(i))
//	p("inst", uint32(i))
//
//	if d.nds.arm9.Reg.IsThumb {
//		p("opcode", d.nds.mem.Read16(reg.R[15], true))
//	} else {
//		p("opcode", d.nds.mem.Read32(reg.R[15], true))
//	}
//	//mode := d.Gba.Cpu.Reg.getMode()
//	//s("--------  --------")
//	//p("r00", reg.R[0])
//	//p("r01", reg.R[1])
//	//p("r02", reg.R[2])
//	//p("r03", reg.R[3])
//	//p("r04", reg.R[4])
//	//p("r05", reg.R[5])
//	//p("r06", reg.R[6])
//	//p("r07", reg.R[7])
//	//p("r08", reg.R[8])
//	//p("r09", reg.R[9])
//	//p("r10", reg.R[10])
//	//p("r11", reg.R[11])
//	//p("r12", reg.R[12])
//	//p("sp/r13", reg.R[13])
//	//p("lr/r14", reg.R[14])
//	//p("pc/r15", reg.R[15])
//	//s("--------  --------")
//	//p("cpsr", uint32(reg.CPSR))
//	//p("8FF8", d.nds.mem.Read32(0x0, true))
//	//p("8FFC", d.nds.mem.Read32(0x4, true))
//	//p("9000", d.nds.mem.Read32(0x8, true))
//	//p("spsr", uint32(reg.SPSR[BANK_ID[mode]]))
//	//p("MODE", BANK_ID[mode])
//
//	s("--------  --------")
//	//start := 0x6200000
//	//count := 0x70
//	//for i := start; i < start + (count * 4); i += 16 {
//
//    //    s := fmt.Sprintf("%08X: ", i)
//
//    //    for j := range 16 {
//    //        s += fmt.Sprintf("%02X ", d.nds.mem.Read(uint32(i + j), true))
//    //    }
//    //    fmt.Printf("%s\n", s)
//	//}
//
//    //for i := range 0x100_0000 / 4 {
//
//    //    if a := d.nds.mem.Read32(0x600_0000 + (uint32(i) * 4), true); a == 0xFBDE {
//    //        panic(fmt.Sprintf("ADDR %08X", a))
//    //    }
//
//    //}
//
//    start := 0x2171CD0
//	count := 16
//	//for i := start; i <= start - (count * 4); i -= 4 {
//	for i := start; i >= start - (count * 4); i -= 4 {
//	    p(fmt.Sprintf("ADDR %X", i), d.nds.mem.Read32(uint32(i), true))
//	}
//	s("------")
//}
//
//func (d *Debugger) DecodeArm(opcode uint32) string {
//	switch {
//	case isSWI(opcode):
//		return "SWI"
//	case isB(opcode):
//		return "B"
//	case isBX(opcode):
//		return "BX"
//	case isSDT(opcode):
//		return "SDT"
//	case isBlock(opcode):
//		return "BLOCK"
//	case isHalf(opcode):
//		return "HALF"
//	case isUD(opcode):
//		return "UNDEFINED"
//	case isPSR(opcode):
//		return "PSR"
//	case isSWP(opcode):
//		return "SWP"
//	case isM(opcode):
//		return "MULTI"
//	case isALU(opcode):
//		return "ALU"
//	case isCoDataReg(opcode):
//		return "CO DATA REG"
//	case isCoDataTrans(opcode):
//		return "CO DATA TRANS"
//	}
//	return "UNKNOWN ARM"
//}
//
//func (d *Debugger) DecodeThumb(opcode uint16) string {
//
//	switch {
//	case isthumbSWI(opcode):
//		return "SWI"
//	case isThumbAddSub(opcode):
//		return "ADD/SUB"
//	case isThumbShift(opcode):
//		return "SHIFT"
//	case isThumbImm(opcode):
//		return "IMM"
//	case isThumbAlu(opcode):
//		return "ALU"
//	case isThumbHiReg(opcode):
//		return "HI REG"
//	case isLSHalf(opcode):
//		return "LSHALF"
//	case isLSSigned(opcode):
//		return "LSSIGNED"
//	case isLPC(opcode):
//		return "LPC"
//	case isLSR(opcode):
//		return "LSR"
//	case isLSImm(opcode):
//		return "LSIMM"
//	case isPushPop(opcode):
//		return "PUSHPOP"
//	case isRelative(opcode):
//		return "RELATIVE"
//	case isThumbB(opcode):
//		return "B"
//	case isJumpCall(opcode):
//		return "JUMP"
//	case isStack(opcode):
//		return "STACK"
//	case isLongBranch(opcode):
//		return "BL"
//	case isShortLongBranch(opcode):
//		return "BL SHORT"
//	case isLSSP(opcode):
//		return "LSSP"
//	case isMulti(opcode):
//		return "MULTI"
//	}
//	return "UNKNOWN THUMB"
//}
//
//func isOpcodeFormat(opcode, mask, format uint32) bool {
//	return opcode&mask == format
//}
//
//func isCoDataTrans(opcode uint32) bool {
//	return isOpcodeFormat(opcode,
//		0b0000_1110_0000_0000_0000_0000_0000_0000,
//		0b0000_1100_0000_0000_0000_0000_0000_0000,
//	)
//}
//
//func isCoDataReg(opcode uint32) bool {
//	return isOpcodeFormat(opcode,
//		0b0000_1111_0000_0000_0000_0000_0000_0000,
//		0b0000_1110_0000_0000_0000_0000_0000_0000,
//	)
//}
//
//func isSWP(opcode uint32) bool {
//
//	return isOpcodeFormat(opcode,
//		0b0000_1111_1011_0000_0000_1111_1111_0000,
//		0b0000_0001_0000_0000_0000_0000_1001_0000,
//	)
//}
//
//func isBlock(opcode uint32) bool {
//
//	is := false
//
//	is = is || isOpcodeFormat(opcode,
//		0b0000_1110_0001_0000_0000_0000_0000_0000,
//		0b0000_1000_0001_0000_0000_0000_0000_0000,
//	)
//
//	is = is || isOpcodeFormat(opcode,
//		0b0000_1110_0001_0000_0000_0000_0000_0000,
//		0b0000_1000_0000_0000_0000_0000_0000_0000,
//	)
//
//	return is
//}
//
//func isHalf(opcode uint32) bool {
//
//	is := false
//
//	// LDRH
//	is = is || isOpcodeFormat(opcode,
//		0b0000_1110_0001_0000_0000_0000_1111_0000,
//		0b0000_0000_0001_0000_0000_0000_1011_0000,
//	)
//
//	// LDRSB
//	is = is || isOpcodeFormat(opcode,
//		0b0000_1110_0001_0000_0000_0000_1111_0000,
//		0b0000_0000_0001_0000_0000_0000_1101_0000,
//	)
//
//	// LDRSH
//	is = is || isOpcodeFormat(opcode,
//		0b0000_1110_0001_0000_0000_0000_1111_0000,
//		0b0000_0000_0001_0000_0000_0000_1111_0000,
//	)
//
//	// STRH
//	is = is || isOpcodeFormat(opcode,
//		0b0000_1110_0001_0000_0000_0000_1111_0000,
//		0b0000_0000_0000_0000_0000_0000_1011_0000,
//	)
//
//	return is
//}
//
//func isALU(opcode uint32) bool {
//	return isOpcodeFormat(opcode,
//		0b0000_1100_0000_0000_0000_0000_0000_0000,
//		0b0000_0000_0000_0000_0000_0000_0000_0000,
//	)
//}
//
//func isBX(opcode uint32) bool {
//	return isOpcodeFormat(opcode,
//		0b0000_1111_1111_1111_1111_1111_1101_0000,
//		0b0000_0001_0010_1111_1111_1111_0001_0000,
//	)
//}
//
//func isB(opcode uint32) bool {
//	return isOpcodeFormat(opcode,
//		0b0000_1110_0000_0000_0000_0000_0000_0000,
//		0b0000_1010_0000_0000_0000_0000_0000_0000,
//	)
//}
//
//func isM(opcode uint32) bool {
//
//	is := false
//
//	is = is || isOpcodeFormat(opcode,
//		0b0000_1110_1000_0000_0000_0000_1111_0000,
//		0b0000_0000_0000_0000_0000_0000_1001_0000,
//	)
//	is = is || isOpcodeFormat(opcode,
//		0b0000_1110_1000_0000_0000_0000_1111_0000,
//		0b0000_0000_1000_0000_0000_0000_1001_0000,
//	)
//
//	return is
//}
//
//func isSWI(opcode uint32) bool {
//	return isOpcodeFormat(
//		opcode,
//		0b0000_1111_0000_0000_0000_0000_0000_0000,
//		0b0000_1111_0000_0000_0000_0000_0000_0000,
//	)
//}
//
//func isUD(opcode uint32) bool {
//	return isOpcodeFormat(
//		opcode,
//		0b0000_1110_0000_0000_0000_0000_0000_0000,
//		0b0000_0110_0000_0000_0000_0000_0000_0000,
//	)
//}
//
//func isSDT(opcode uint32) bool {
//	is := false
//	is = is || isOpcodeFormat(
//		opcode,
//		0b0000_1100_0001_0000_0000_0000_0000_0000,
//		0b0000_0100_0001_0000_0000_0000_0000_0000,
//	)
//	is = is || isOpcodeFormat(
//		opcode,
//		0b0000_1100_0001_0000_0000_0000_0000_0000,
//		0b0000_0100_0000_0000_0000_0000_0000_0000,
//	)
//
//	return is
//}
//
//func isPSR(opcode uint32) bool {
//
//	is := false
//
//	is = is || isOpcodeFormat(
//		opcode,
//		0b0000_1111_1011_1111_0000_1111_1111_1111,
//		0b0000_0001_0000_1111_0000_0000_0000_0000,
//	)
//
//	is = is || isOpcodeFormat(
//		opcode,
//		0b0000_1101_1011_0000_1111_0000_0000_0000,
//		0b0000_0001_0010_0000_1111_0000_0000_0000,
//	)
//
//	return is
//}
//
//func isThumbOpcodeFormat(opcode, mask, format uint16) bool {
//	return opcode&mask == format
//}
//
//func isThumbShift(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1110_0000_0000_0000,
//		0b0000_0000_0000_0000,
//	)
//}
//
//func isThumbAddSub(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1111_1000_0000_0000,
//		0b0001_1000_0000_0000,
//	)
//}
//
//func isThumbImm(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1110_0000_0000_0000,
//		0b0010_0000_0000_0000,
//	)
//}
//
//func isThumbAlu(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1111_1100_0000_0000,
//		0b0100_0000_0000_0000,
//	)
//}
//
//func isThumbHiReg(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1111_1100_0000_0000,
//		0b0100_0100_0000_0000,
//	)
//}
//
//func isLSHalf(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1111_0000_0000_0000,
//		0b1000_0000_0000_0000,
//	)
//}
//
//func isLSSigned(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1111_0010_0000_0000,
//		0b0101_0010_0000_0000,
//	)
//}
//
//func isLPC(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1111_1000_0000_0000,
//		0b0100_1000_0000_0000,
//	)
//}
//
//func isLSR(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1111_0010_0000_0000,
//		0b0101_0000_0000_0000,
//	)
//}
//
//func isLSImm(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1110_0000_0000_0000,
//		0b0110_0000_0000_0000,
//	)
//}
//
//func isPushPop(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1111_0110_0000_0000,
//		0b1011_0100_0000_0000,
//	)
//}
//
//func isRelative(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1111_0000_0000_0000,
//		0b1010_0000_0000_0000,
//	)
//}
//
//func isJumpCall(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1111_0000_0000_0000,
//		0b1101_0000_0000_0000,
//	)
//}
//
//func isThumbB(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1111_1000_0000_0000,
//		0b1110_0000_0000_0000,
//	)
//}
//
//func isStack(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1111_1111_0000_0000,
//		0b1011_0000_0000_0000,
//	)
//}
//
//func isLongBranch(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1111_1000_0000_0000,
//		0b1111_0000_0000_0000,
//	)
//}
//
//func isShortLongBranch(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1111_1000_0000_0000,
//		0b1111_1000_0000_0000,
//	)
//}
//
//func isLSSP(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1111_0000_0000_0000,
//		0b1001_0000_0000_0000,
//	)
//}
//
//func isMulti(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1111_0000_0000_0000,
//		0b1100_0000_0000_0000,
//	)
//}
//
//func isthumbSWI(opcode uint16) bool {
//	return isThumbOpcodeFormat(opcode,
//		0b1111_1111_0000_0000,
//		0b1101_1111_0000_0000,
//	)
//}
