package gba

import (
	_ "embed"
)

//go:embed bios.bin
var biosFile []byte

const (
	INTRWAIT_NONE   = 0
	INTRWAIT_VBLANK = 1
)

const (
	SYS_SoftReset                      = 0x00
	SYS_RegisterRamReset               = 0x01
	SYS_Halt                           = 0x02
	SYS_StopSleep                      = 0x03
	SYS_IntrWait                       = 0x04
	SYS_VBlankIntrWait                 = 0x05
	SYS_Div                            = 0x06
	SYS_DivArm                         = 0x07
	SYS_Sqrt                           = 0x08
	SYS_ArcTan                         = 0x09
	SYS_ArcTan2                        = 0x0A
	SYS_CpuSet                         = 0x0B
	SYS_CpuFastSet                     = 0x0C
	SYS_GetBiosChecksum                = 0x0D
	SYS_BgAffineSet                    = 0x0E
	SYS_ObjAffineSet                   = 0x0F
	SYS_BitUnPack                      = 0x10
	SYS_LZ77UnCompReadNormalWrite8bit  = 0x11
	SYS_LZ77UnCompReadNormalWrite16bit = 0x12
	SYS_HuffUnCompReadNormal           = 0x13
	SYS_RLUnCompReadNormalWrite8bit    = 0x14
	SYS_RLUnCompReadNormalWrite16bit   = 0x15
	SYS_Diff8bitUnFilterWrite8bit      = 0x16
	SYS_Diff8bitUnFilterWrite16bit     = 0x17
	SYS_Diff16bitUnFilter              = 0x18
	SYS_SoundBias                      = 0x19
	SYS_SoundDriverInit                = 0x1A
	SYS_SoundDriverMode                = 0x1B
	SYS_SoundDriverMain                = 0x1C
	SYS_SoundDriverVSync               = 0x1D
	SYS_SoundChannelClear              = 0x1E
	SYS_MidiKey2Freq                   = 0x1F
	SYS_SoundWhatever0                 = 0x20
	SYS_SoundWhatever1                 = 0x21
	SYS_SoundWhatever2                 = 0x22
	SYS_SoundWhatever3                 = 0x23
	SYS_SoundWhatever4                 = 0x24
	SYS_MultiBoot                      = 0x25
	SYS_HardReset                      = 0x26
	SYS_CustomHalt                     = 0x27
	SYS_SoundDriverVSyncOff            = 0x28
	SYS_SoundDriverVSyncOn             = 0x29
	SYS_SoundGetJumpList               = 0x2A
)

func (gba *GBA) LoadBios() {

	for i := range len(biosFile) {
		gba.Mem.BIOS[i] = uint8(biosFile[i])
	}
}

//
//func (gba *GBA) SysCall(inst uint32) (int, bool) {
//
//    //savedIme := gba.InterruptStack.IME
//    //gba.InterruptStack.IME = false
//
//    panic("THIS SHOULD NOT BE CALLED")
//
//
//	cycles := 0
//
//    //fmt.Printf("SYS CALL %08X CURR %d\n", inst, CURR_INST)
//    gba.Mem.BIOS_MODE = BIOS_SWI
//
//    if inst > 0x2A {
//
//        r := &gba.Cpu.Reg.R
//
//        panic(fmt.Sprintf("INVALID SWI SYSCALL %08X PC %08X OPCODE %08X", inst, r[PC], gba.Mem.Read32(r[PC])))
//    }
//    //gba.exception(VEC_SWI, MODE_SWI)
//
//	switch inst {
//	case SYS_SoftReset:
//		SoftReset(gba)
//		cycles += 200 // approx
//	case SYS_RegisterRamReset:
//        fmt.Printf("SYS CALL %08X CURR %d\n", inst, CURR_INST)
//        PRINT_EM = false
//		RegisterRamReset(gba)
//		cycles += 30 // approx
//    case SYS_Halt:
//        //panic("HALT")
//        gba.Halted = true
//        gba.IntrWait = 0xFFFF
//	case SYS_IntrWait:
//		IntrWait(gba)
//	case SYS_VBlankIntrWait:
//		VBlankIntrWait(gba)
//	case SYS_Div:
//		Div(gba, false)
//		cycles += 370 // approx
//	case SYS_DivArm:
//		Div(gba, true)
//		cycles += 130 // approx
//	case SYS_Sqrt:
//		Sqrt(gba)
//        cycles += 130
//	case SYS_ArcTan:
//		ArcTan(gba)
//        cycles += 140
//	case SYS_ArcTan2:
//		ArcTan2(gba)
//		cycles += 520 // approx
//	case SYS_CpuSet:
//		cycles += CpuSet(gba)
//	case SYS_CpuFastSet:
//		cycles += CpuFastSet(gba)
//
//	case SYS_BitUnPack:
//		BitUnPack(gba)
//    case SYS_HuffUnCompReadNormal:
//        panic("HUFFMAN IS NOT IMPLIMENTED")
//
//        Huff(gba)
//
//        //cycles += HuffUnCompReadNormal(gba)
//	case SYS_LZ77UnCompReadNormalWrite8bit:
//		cycles += LZ77UnCompReadNormalWrite8bit(gba)
//	case SYS_LZ77UnCompReadNormalWrite16bit:
//		cycles += LZ77UnCompReadNormalWrite16bit(gba)
//	case SYS_RLUnCompReadNormalWrite8bit:
//		cycles += RLUUnCompReadNormalWrite8bit(gba)
//	case SYS_RLUnCompReadNormalWrite16bit:
//		cycles += RLUUnCompReadNormalWrite16bit(gba)
//    case SYS_Diff16bitUnFilter:
//        cycles += DecompressDiff16bit(gba, gba.Cpu.Reg.R[0], gba.Cpu.Reg.R[1])
//    case SYS_Diff8bitUnFilterWrite8bit:
//        cycles += DecompressDiff8bit(gba, gba.Cpu.Reg.R[0], gba.Cpu.Reg.R[1])
//    case SYS_Diff8bitUnFilterWrite16bit:
//        cycles += DecompressDiff8bit(gba, gba.Cpu.Reg.R[0], gba.Cpu.Reg.R[1])
//	case SYS_ObjAffineSet:
//		cycles += ObjAffineSet(gba)
//	case SYS_BgAffineSet:
//		cycles += BGAffineSet(gba)
//    case SYS_MidiKey2Freq:
//        MidiKey2Freq(gba)
//	case SYS_GetBiosChecksum:
//		GetBiosChecksum(gba)
//        cycles += 168948
//    //default :gba.exception(SWI_VEC, MODE_SWI)
//	default:
//
//    //    //fmt.Printf("SWI %04X\n", inst)
//
//        //gba.exception(SWI_VEC, MODE_SWI)
//    //    //return cycles, false // keeps from inc PC after setting in exception
//		panic(fmt.Sprintf("EXCEPTION OR UNHANDLED SYS CALL TYPE 0x%X\n", inst))
//	}
//
//    gba.Mem.BIOS_MODE = BIOS_SWI
//    //gba.InterruptStack.IME = savedIme
//
//    cycles += 6
//
//	return cycles, true
//}
//
//func BGAffineSet(gba *GBA) int {
//
//	r := &gba.Cpu.Reg.R
//	mem := gba.Mem
//
//	i := r[2]
//	var ox, oy float64
//	var cx, cy float64
//	var sx, sy float64
//	var theta float64
//	offset, destination := r[0], r[1]
//	var a, b, c, d float64
//	var rx, ry float64
//	for ; i > 0; i-- {
//		// [ sx   0  0 ]   [ cos(theta)  -sin(theta)  0 ]   [ 1  0  cx - ox ]   [ A B rx ]
//		// [  0  sy  0 ] * [ sin(theta)   cos(theta)  0 ] * [ 0  1  cy - oy ] = [ C D ry ]
//		// [  0   0  1 ]   [     0            0       1 ]   [ 0  0     1    ]   [ 0 0  1 ]
//
//		ox = float64(mem.Read32(offset)) / 256
//		oy = float64(mem.Read32(offset+4)) / 256
//		cx = float64(uint16(mem.Read16(offset + 8)))
//		cy = float64(uint16(mem.Read16(offset + 10)))
//		sx = float64(uint16(mem.Read16(offset+12))) / 256
//		sy = float64(uint16(mem.Read16(offset+14))) / 256
//		theta = (float64(mem.Read16(offset+16)>>8) / 128) * math.Pi
//		offset += 20
//
//		// Rotation
//		a = math.Cos(theta)
//		d = a
//		b = math.Sin(theta)
//		c = b
//
//		// Scale
//		a *= sx
//		b *= -sx
//		c *= sy
//		d *= sy
//
//		// Translate
//		rx = ox - (a*cx + b*cy)
//		ry = oy - (c*cx + d*cy)
//
//		mem.Write16(destination, uint16(a*256))
//		mem.Write16(destination+2, uint16(b*256))
//		mem.Write16(destination+4, uint16(c*256))
//		mem.Write16(destination+6, uint16(d*256))
//		mem.Write32(destination+8, uint32(rx*256))
//		mem.Write32(destination+12, uint32(ry*256))
//		destination += 16
//	}
//
//    return 36 + (int(i) * 19)
//}
//
//func ObjAffineSet(gba *GBA) int {
//
//	r := &gba.Cpu.Reg.R
//	mem := gba.Mem
//
//	i := r[2]
//	var sx, sy float64
//	var theta float64
//	offset := r[0]
//	destination := r[1]
//	diff := r[3]
//	var a, b, c, d float64
//	for ; i > 0; i-- {
//		// [ sx   0 ]   [ cos(theta)  -sin(theta) ]   [ A B ]
//		// [  0  sy ] * [ sin(theta)   cos(theta) ] = [ C D ]
//		sx = float64(uint16(mem.Read16(offset))) / 256
//		sy = float64(uint16(mem.Read16(offset+2))) / 256
//		theta = (float64(uint16(mem.Read16(offset+4))>>8) / 128) * math.Pi
//		offset += 6
//
//		// Rotation
//		a = math.Cos(theta)
//		d = a
//		b = math.Sin(theta)
//		c = b
//
//		// Scale
//		a *= sx
//		b *= -sx
//		c *= sy
//		d *= sy
//
//		mem.Write16(destination, uint16(a*256))
//		mem.Write16(destination+diff, uint16(b*256))
//		mem.Write16(destination+diff*2, uint16(c*256))
//		mem.Write16(destination+diff*3, uint16(d*256))
//		destination += diff * 4
//	}
//
//    return 13 + (int(i) * 18)
//}
//
//var IntrWaitReg Reg
//
//func IntrWait(gba *GBA) {
//
//	r := &gba.Cpu.Reg.R
//
//    IntrWaitReg = gba.Cpu.Reg
//
//	waitMode := r[0]
//
//    switch waitMode {
//    case 0:
//        gba.Halted = true
//        IntrWaitReturn(gba)
//
//    case 1:
//        // Discard old IF flags if waitMode == 1
//
//        // ldrh   r3, [r12, #-8]		set r3 to 0x3FF_FFF8 (0x300_7FF8)
//        r[3] = gba.Mem.Read16(0x3FF_FFF8)
//        // bic    r3, r1			    r3 &^= r1
//        r[3] &^= r[1]
//        // strh   r3, [r12, #-8]		set 0x3FF_FFF8 (0x300_7FF8) to r3
//        gba.Mem.Write16(0x3FF_FFF8, uint16(r[3]))
//
//        // strb   r0, [r12, #0x301]
//        gba.Halted = true
//
//    default:
//        panic("UNKNOWN INTRA WAIT MODE")
//    }
//}
//
//func IntrWaitReturn(gba *GBA) {
//
//    // IRQ USER HANDLER MUST |= Interrupts to 0x300_7FF8 (0x3FF_FFF8)
//
//    // @ Check which interrupts were acknowledged
//	r := &gba.Cpu.Reg.R
//    // strb   r0, [r12, #0x208]
//    gba.InterruptStack.IME = false
//    // ldrh   r3, [r12, #-8]
//    r[3] = gba.Mem.Read16(0x3FF_FFF8)
//    // ands   r3, r1
//    r[3] = r[3] & r[1]
//
//    // eorne  r3, r1
//    if r[3] != 0 {
//        r[3] ^= r[1]
//        // strneh r3, [r12, #-8]
//        gba.Mem.Write16(0x3FF_FFF8, uint16(r[3]))
//    }
//    // strb   r2, [r12, #0x208]
//    gba.InterruptStack.IME = true
//    // beq    0b
//    if r[3] == 0 {
//        return
//    }
//
//    fmt.Printf("LEAVING INTRA\n")
//    // ldmfd  sp!, {r2-r3, pc}
//    gba.Cpu.Reg = IntrWaitReg
//    gba.Halted = false
//}
//
//func VBlankIntrWait(gba *GBA) {
//	r := &gba.Cpu.Reg.R
//
//	r[0] = 1
//	r[1] = 1
//
//	IntrWait(gba)
//}
//
//func SoftReset(gba *GBA) {
//
//	// untested
//
//	/*
//	   clears 0x200 of ram
//	   set r0-r12, LR_svc, SPSR_svc, LR_irq, SPSR_irq all to zero
//	   enters sys mode
//	   Host  sp_svc    sp_irq    sp_sys    zerofilled area       return address
//	   GBA   3007FE0h  3007FA0h  3007F00h  [3007E00h..3007FFFh]  Flag[3007FFAh]
//	*/
//
//	reg := &gba.Cpu.Reg
//
//	const (
//		RETURN_ADDR = 0x0300_7FFA
//		ZERO_FILL   = 0x0300_7E00
//	)
//
//	flag := gba.Mem.Read(RETURN_ADDR, false)
//
//	i := uint32(0)
//	for i = range 0x200 {
//		gba.Mem.Write(ZERO_FILL+i, 0, false)
//	}
//
//	reg.CPSR.SetMode(MODE_SWI)
//	reg.R[SP] = 0x0300_7FE0
//	reg.CPSR.SetMode(MODE_IRQ)
//	reg.R[SP] = 0x0300_7FA0
//	reg.CPSR.SetMode(MODE_SYS)
//	reg.R[SP] = 0x0300_7F00
//
//	reg.R[LR] = 0x0200_0000
//	if flag == 0 {
//		reg.R[LR] = 0x0800_0000
//	}
//
//	reg.CPSR.SetFlag(FLAG_T, false)
//
//	reg.R[PC] = reg.R[LR]
//
//    gba.Mem.BIOS_MODE = BIOS_STARTUP
//
//	// pipelining
//}
//
//func RegisterRamReset(gba *GBA) {
//
//	mem := gba.Mem
//	r := &gba.Cpu.Reg.R
//	flags := r[0]
//
//    fmt.Printf("Flags %08b PC %08X\n", uint8(flags), r[PC])
//
//	if clearWRAM1 := utils.BitEnabled(flags, 0); clearWRAM1 {
//		mem.WRAM1 = [0x40000]uint8{}
//	}
//
//	if clearWRAM2 := utils.BitEnabled(flags, 1); clearWRAM2 {
//
//		// need to exclude last 0x200
//		for i := range 0x8000 - 0x200 {
//			mem.WRAM2[i] = 0x0
//		}
//	}
//
//	if clearPRAM := utils.BitEnabled(flags, 2); clearPRAM {
//		mem.PRAM = [0x400]uint8{}
//	}
//
//	if clearVRAM := utils.BitEnabled(flags, 3); clearVRAM {
//		mem.VRAM = [0x18000]uint8{}
//	}
//
//	if clearOAM := utils.BitEnabled(flags, 4); clearOAM {
//		mem.OAM = [0x400]uint8{}
//	}
//
//	if clearSIO := utils.BitEnabled(flags, 5); clearSIO {
//
//		for i := uint32(0x120); i <= 0x12C; i++ {
//            mem.Write8(0x400_0000 + i, 0)
//		}
//
//		for i := uint32(0x134); i <= 0x154; i++ {
//            mem.Write8(0x400_0000 + i, 0)
//		}
//	}
//
//	if clearSound := utils.BitEnabled(flags, 6); clearSound {
//
//		for i := uint32(0x60); i <= 0xA8; i++ {
//            mem.Write8(0x400_0000 + i, 0)
//		}
//	}
//
//	if clearOther := utils.BitEnabled(flags, 7); clearOther {
//		for i := range 0x400 {
//
//			sio1 := i >= 0x120 && i <= 0x12C
//			sio2 := i >= 0x134 && i <= 0x154
//			sound := i >= 0x60 && i <= 0xA8
//            //other := i >= 0x200 && i <= 0x20B
//
//			if sio1 || sio2 || sound {
//				continue
//			}
//		}
//
//        s := gba.InterruptStack
//        s.IF = 0
//        s.IE = 0
//        s.IME = false
//
//            //// default values pulled from ruby
//            //mem.IO[0x00] = 0x80
//            //mem.IO[0x0021] = 0x1
//            //mem.IO[0x0027] = 0x1
//            //mem.IO[0x0031] = 0x1
//            //mem.IO[0x0037] = 0x1
//
//            //mem.IO[0x0082] = 0xE
//            //mem.IO[0x0083] = 0x88
//            //mem.IO[0x0089] = 0x2
//
//            //mem.IO[0x0128] = 0x4
//            //mem.IO[0x0130] = 0xFF
//            //mem.IO[0x0131] = 0x3
//            //mem.IO[0x0134] = 0xF
//            //mem.IO[0x0135] = 0x80
//            //mem.IO[0x0300] = 0x1
//
//            ////mem.GBA.checkIRQ()
//	}
//
//    mem.Write8(0x80, 0)
//
//    r[3] = 0x170 // CLOBBER
//}
//
//func Div(gba *GBA, arm bool) {
//
//	const MAX = 0x8000_0000
//    const I32_MIN = -2147483647 - 1
//
//	r := &gba.Cpu.Reg.R
//
//	nu := int32(r[0])
//	de := int32(r[1])
//
//	if arm {
//		tmp := nu
//		nu = de
//		de = tmp
//	}
//
//    switch {
//    case de == 0 && nu < 0:
//        r[0] = 0xFFFF_FFFF
//        r[1] = uint32(nu)
//        r[3] = 1
//        return
//    case de == 0:
//        r[0] = 1
//        r[1] = uint32(nu)
//        r[3] = 1
//        return
//    case de == -1 && nu == I32_MIN:
//		r[0] = MAX
//		r[1] = 0
//		r[3] = MAX
//		return
//	}
//
//	res := uint32(nu / de)
//	mod := uint32(nu % de)
//	abs := uint32(math.Abs(float64(res)))
//
//	r[0] = res
//	r[1] = mod
//	r[3] = abs
//}
//
//func Sqrt(gba *GBA) {
//
//	reg := &gba.Cpu.Reg
//
//	input := reg.R[0]
//
//	if input == 0 {
//		reg.R[0] = 0
//		return
//	}
//
//	lo, hi, bound := uint32(0), input, uint32(1)
//
//	for bound < hi {
//		hi >>= 1
//		bound <<= 1
//	}
//
//	for {
//		hi = input
//		acc := uint32(0)
//		lo = bound
//
//		for {
//			oldLower := lo
//			if lo <= hi>>1 {
//				lo <<= 1
//			}
//			if oldLower >= hi>>1 {
//				break
//			}
//		}
//
//		for {
//			acc <<= 1
//			if hi >= lo {
//				acc++
//				hi -= lo
//			}
//			if lo == bound {
//				break
//			}
//			lo >>= 1
//		}
//
//		oldBound := bound
//		bound += acc
//		bound >>= 1
//		if bound >= oldBound {
//			bound = oldBound
//			break
//		}
//	}
//
//	reg.R[0] = bound
//}
//
//func ArcTan(gba *GBA) {
//
//	r := &gba.Cpu.Reg.R
//	r[0], r[1], r[3] = _ArcTan(int32(r[0]))
//}
//
//func ArcTan2(gba *GBA) {
//
//	r := &gba.Cpu.Reg.R
//
//	x := int32(r[0])
//	y := int32(r[1])
//
//	outX := uint32(0)
//	outY := uint32(0)
//
//	switch {
//	case y == 0:
//		if x < 0 {
//			outX = 0x8000
//		}
//	case x == 0:
//		if y >= 0 {
//			outX = 0x4000
//			outY = uint32(y)
//		} else {
//			outX = 0xC000
//			outY = uint32(y)
//		}
//	case y >= 0:
//		if x >= 0 && x >= y {
//			outX, outY, _ = _ArcTan((y << 14) / x)
//		} else if -x >= y {
//			outX, outY, _ = _ArcTan((y << 14) / x)
//			outX += 0x8000
//		} else {
//			outX, outY, _ = _ArcTan((x << 14) / y)
//			outX = 0x4000 - outX
//		}
//	case y < 0:
//		if x <= 0 && -x > -y {
//			outX, outY, _ = _ArcTan((y << 14) / x)
//			outX += 0x8000
//		} else if x >= -y {
//			outX, outY, _ = _ArcTan((y << 14) / x)
//			outX += 0x10000
//		} else {
//			outX, outY, _ = _ArcTan((x << 14) / y)
//			outX = 0xC000 - outX
//		}
//	}
//
//	r[0] = outX
//	r[1] = outY
//	r[3] = 0x170
//}
//
//func _ArcTan(src int32) (uint32, uint32, uint32) {
//
//	a := -((src * src) >> 14)
//	b := (int32(0xA9*a) >> 14) + 0x390
//	b = ((b * a) >> 14) + 0x91C
//	b = ((b * a) >> 14) + 0xFB6
//	b = ((b * a) >> 14) + 0x16AA
//	b = ((b * a) >> 14) + 0x2081
//	b = ((b * a) >> 14) + 0x3651
//	b = ((b * a) >> 14) + 0xA2F9
//
//	return uint32((int32(src) * b) >> 16), uint32(a), uint32(b)
//}
//
//func BitUnPack(gba *GBA) {
//
//	mem := gba.Mem
//	r := &gba.Cpu.Reg.R
//	rs := r[0]
//	rd := utils.WordAlign(r[1])
//
//	pointer := r[2]
//
//	length := mem.Read16(pointer)
//	sBitWidth := mem.Read8(pointer + 2)
//	dBitWidth := mem.Read8(pointer + 3)
//	s := mem.Read32(pointer + 4)
//
//	offset := s & 0b0111_1111_1111_1111_1111_1111_1111_1111
//	zeroFlag := (s>>31)&1 == 1
//
//	if length > 0xFFFF {
//		panic("bitunpack length failed")
//	}
//
//	//fmt.Printf("rs %X, rd %X, pointer %X\n", rs, rd, pointer)
//	//fmt.Printf("length %X, sWidth %X, dWidth %X, s %X\n", length, sBitWidth, dBitWidth, s)
//
//	if sBitWidth != 1 || dBitWidth != 4 || offset != 0 || zeroFlag {
//		panic("LIMITED UNPACK SUPPORT")
//	}
//
//	src := []uint32{}
//	dst := []uint32{}
//
//	for i := uint32(0); i < length; i += 4 {
//		v := mem.Read32(rs + i)
//		src = append(src, (v>>0)&0b1111)
//		src = append(src, (v>>4)&0b1111)
//		src = append(src, (v>>8)&0b1111)
//		src = append(src, (v>>12)&0b1111)
//		src = append(src, (v>>16)&0b1111)
//		src = append(src, (v>>20)&0b1111)
//		src = append(src, (v>>24)&0b1111)
//		src = append(src, (v>>28)&0b1111)
//	}
//
//	for i := 0; i < len(src); i += 2 {
//
//		lo := uint32(0)
//		hi := uint32(0)
//
//		a := src[i]
//		b := src[i+1]
//		for j := range 8 {
//
//			if (a>>j)&1 == 1 {
//				lo |= (1 << (j * 4))
//			}
//
//			if (b>>j)&1 == 1 {
//				hi |= (1 << (j * 4))
//			}
//		}
//
//		dst = append(dst, (hi<<16)|lo)
//	}
//
//	for i, v := range dst {
//		mem.Write32(rd+uint32(i*4), v)
//	}
//
//	return
//}
//
//func CpuSet(gba *GBA) int {
//
//	mem := gba.Mem
//	r := &gba.Cpu.Reg.R
//
//	rs := r[0]
//	rd := r[1]
//	info := r[2]
//
//	wordCount := utils.GetVarData(info, 0, 20)
//	fill := utils.BitEnabled(info, 24)
//	isWord := utils.BitEnabled(info, 26)
//
//	switch {
//	case fill && isWord:
//
//
//        rs &^= 0b11
//		rd &^= 0b11
//
//		word := mem.Read32(rs)
//
//        if rs <= 0x200_0000 {
//            word = 0
//        }
//
//		for i := range wordCount {
//
//			mem.Write32(rd+(i<<2), word)
//		}
//
//        r[0] += 4
//        r[1] += wordCount * 4
//
//	case fill && !isWord:
//
//		rd &^= 0b1
//
//        srcAddr := (rs)
//
//
//		word := mem.Read16(srcAddr)
//
//        if unaligned := srcAddr & 1 == 1; unaligned {
//            word = mem.Read8(srcAddr)
//        }
//
//        if srcAddr <= 0x200_0000 {
//            word = 0
//        }
//
//		for i := range wordCount {
//            addr := rd + (i << 1)
//			mem.Write16(addr, uint16(word))
//		}
//
//        r[0] += 2
//        r[1] += wordCount * 2
//
//	case !fill && isWord:
//
//        if notSram := !(rs >= 0xE00_0000 && rs < 0x1000_0000); notSram {
//            rs &^= 0b11
//        }
//        if notSram := !(rd >= 0xE00_0000 && rd < 0x1000_0000); notSram {
//            rd &^= 0b11
//        }
//
//		for i := range wordCount {
//			word := mem.Read32(rs + (i << 2))
//            if rs <= 0x200_0000 {
//                word = 0
//            }
//
//			mem.Write32(rd+(i<<2), word)
//		}
//
//        //r[0] += 4 // this does not match ruby
//        r[0] += wordCount * 4
//        r[1] += wordCount * 4
//
//	case !fill && !isWord:
//
//		rd &^= 0b1
//
//		for i := range wordCount {
//
//            srcAddr := (rs + (i<<1))
//			word := mem.Read16(srcAddr)
//            if unaligned := srcAddr & 1 == 1; unaligned {
//                word = mem.Read8(srcAddr)
//            }
//
//            if srcAddr <= 0x200_0000{
//                word = 0
//            }
//
//            dstAddr := (rd + (i<<1))
//			mem.Write16(dstAddr, uint16(word))
//		}
//
//        r[0] += wordCount * 2
//        r[1] += wordCount * 2
//	}
//
//    r[3] = 0x170 // offical bios clobbers r3
//
//    return int(wordCount) * 4
//}
//
//func CpuFastSet(gba *GBA) int {
//
//	r := &gba.Cpu.Reg.R
//	mem := gba.Mem
//
//    src := r[0]
//    dst := r[1]
//
//    if notSram := !(src >= 0xE00_0000 && src < 0x1000_0000); notSram {
//        src &^= 0b11
//    }
//
//    if notSram := !(dst >= 0xE00_0000 && dst < 0x1000_0000); notSram {
//        dst &^= 0b11
//    }
//
//	count := (utils.GetVarData(r[2], 0, 20) + 7) &^ 7 // round up 32 bytes (8 words)
//	fill := utils.BitEnabled(r[2], 24)
//
//    cycles := int(count) * 12
//
//	if fill {
//		word := mem.Read32(src)
//        if src <= 0x200_0000 {
//            word = 0
//        }
//
//		for i := uint32(0); i < count; i++ {
//			mem.Write32(dst+(i<<2), word)
//		}
//
//        r[1] += count * 4
//        r[3] = 0x0
//
//        return cycles
//    }
//
//    for i := uint32(0); i < count; i++ {
//
//        srcAddr := src + (i << 2)
//        word := mem.Read32(srcAddr)
//
//        if srcAddr <= 0x200_0000 {
//            word = 0
//        }
//
//        mem.Write32(dst+(i<<2), word)
//	}
//
//    // assuming r1 is incremented since fill does
//    r[1] += count * 4
//    r[3] = 0x0
//
//    return cycles
//
//}
//
//func LZ77UnCompReadNormalWrite8bit(gba *GBA) int {
//
//	r := &gba.Cpu.Reg.R
//	src := r[0]
//	dst := r[1]
//
//    bytesOutputted := DecompressLZ77(gba, src, dst, false)
//    return bytesOutputted * 5
//}
//
//func LZ77UnCompReadNormalWrite16bit(gba *GBA) int {
//
//	r := &gba.Cpu.Reg.R
//	src := r[0]
//	dst := r[1]
//
//    bytesOutputted := DecompressLZ77(gba, src, dst, true)
//    return bytesOutputted * 4
//}
//
//func DecompressLZ77(gba *GBA, src, dst uint32, half bool) int {
//
//	// need to align half and pad 16bit?
//    // I'm not sure how final r0 value is calculated, it does not match src
//
//	mem := gba.Mem
//
//	header := mem.Read32(utils.WordAlign(src))
//    //oSrc := src
//	decompressedSize := int(header >> 8)
//
//	src += 4
//
//	end := int(dst) + decompressedSize
//
//    bytesOutputted := 0
//
//	for int(dst) < end {
//
//		flagByte := mem.Read8(src)
//		src++
//
//		for i := range 8 {
//
//			if half && (int(dst) > end) || !half && (int(dst) >= end) {
//				break
//			}
//
//			flag := (flagByte >> (7 - i)) & 1
//			if flag == 0 {
//				// Uncompressed
//				mem.Write(dst, uint8(mem.Read8(src)), false)
//                bytesOutputted++
//				dst++
//				src++
//
//			} else {
//				// Compressed
//				first := mem.Read8(src)
//				second := mem.Read8(src + 1)
//
//				src += 2
//
//				length := int((first >> 4) + 3)
//				disp := int(((int(first) & 0xF) << 8) | int(second))
//				copyFrom := int(dst) - (disp + 1)
//
//				for j := range length {
//					mem.Write(dst, uint8(mem.Read8(uint32(copyFrom+j))), false)
//                    bytesOutputted++
//					dst++
//				}
//			}
//		}
//	}
//
//    gba.Cpu.Reg.R[0] = src
//    gba.Cpu.Reg.R[1] += uint32(decompressedSize)
//    //gba.Cpu.Reg.R[2] = 0x0
//    gba.Cpu.Reg.R[3] = 0x0 // CLOBBER
//
//    //fmt.Printf("srcSize %08X LEN %08X DCOMP %08X BYTES %08X\n", src, src - oSrc, decompressedSize, bytesOutputted)
//
//    return bytesOutputted
//}
//
//func RLUUnCompReadNormalWrite8bit(gba *GBA) int {
//
//	r := &gba.Cpu.Reg.R
//	src := r[0]
//	dst := r[1]
//
//    bytesOutputted := DecompressRLU(gba, src, dst)
//    return bytesOutputted * 3
//}
//
//func RLUUnCompReadNormalWrite16bit(gba *GBA) int {
//
//	r := &gba.Cpu.Reg.R
//	src := r[0]
//	dst := r[1]
//
//    bytesOutputted := DecompressRLU(gba, src, dst)
//    return bytesOutputted * 2
//}
//
//func DecompressRLU(gba *GBA, src, dst uint32) int {
//
//	// need to align half and pad 16bit?
//
//	mem := gba.Mem
//
//	header := mem.Read32(src)
//	decompressedSize := int(header >> 8)
//	src += 4
//
//	end := int(dst) + decompressedSize
//
//    bytesOutputted := 0
//	for int(dst) < end {
//		flag := mem.Read8(src)
//		src++
//
//		if (flag & 0x80) == 0 {
//			// Uncompressed block: copy (flag + 1) bytes
//			count := int(flag&0x7F) + 1
//			for range count {
//				b := mem.Read8(src)
//				mem.Write(dst, uint8(b), false)
//                bytesOutputted++
//				src++
//				dst++
//			}
//		} else {
//			// Compressed block: repeat 1 byte for (flag & 0x7F) + 3 times
//			count := int(flag&0x7F) + 3
//			value := mem.Read8(src)
//			src++
//			for range count {
//				mem.Write(dst, uint8(value), false)
//                bytesOutputted++
//				dst++
//			}
//		}
//	}
//
//    return bytesOutputted
//}
//
//
//func HuffUnCompReadNormal(gba *GBA) int {
//
//	r := &gba.Cpu.Reg.R
//	src := r[0]
//	dst := r[1]
//
//    bytesOutputted := DecompressHuff(gba, src, dst)
//    return bytesOutputted * 2
//}
//func DecompressHuff(gba *GBA, src, dst uint32) int {
//
//	// need to align half and pad 16bit?
//
//	mem := gba.Mem
//
//	header := mem.Read32(src)
//	decompressedSize := int(header >> 8)
//	src += 4
//
//	end := int(dst) + decompressedSize
//
//    bytesOutputted := 0
//	for int(dst) < end {
//		flag := mem.Read8(src)
//		src++
//
//		if (flag & 0x80) == 0 {
//			// Uncompressed block: copy (flag + 1) bytes
//			count := int(flag&0x7F) + 1
//			for range count {
//				b := mem.Read8(src)
//				mem.Write(dst, uint8(b), false)
//                bytesOutputted++
//				src++
//				dst++
//			}
//		} else {
//			// Compressed block: repeat 1 byte for (flag & 0x7F) + 3 times
//			count := int(flag&0x7F) + 3
//			value := mem.Read8(src)
//			src++
//			for range count {
//				mem.Write(dst, uint8(value), false)
//                bytesOutputted++
//				dst++
//			}
//		}
//	}
//
//    return bytesOutputted
//}
//
////func DecompressHuff(gba *GBA, srcAddr, dstAddr uint32) int {
////
////	mem := gba.Mem
////	header := mem.Read32(srcAddr)
////	srcAddr += 4
////
////	compType := (header >> 4) & 0xF
////	decompressedSize := header >> 8
////
////	if compType != 2 {
////		panic("Not Huffman compressed")
////	}
////
////	// --- Step 2: Tree size and read tree ---
////	treeSizeByte := mem.Read8(srcAddr)
////	srcAddr += 1
////
////	treeSize := uint32((int(treeSizeByte)+1)*2)
////	bitstreamStart := srcAddr + treeSize
////
////	tree := make([]uint32, treeSize)
////    for i := range treeSize {
////        tree[i] = mem.Read8(srcAddr)
////
////        srcAddr++
////    }
////
////	// --- Step 3: Bitstream reader ---
////	bitBuffer := uint32(0)
////	bitCount := 0
////	bitOffset := uint32(0)
////
////	getBit := func() int {
////		if bitCount == 0 {
////			bitBuffer = mem.Read32(bitstreamStart + bitOffset)
////			bitOffset += 4
////			bitCount = 32
////		}
////		bit := int((bitBuffer >> 31) & 1) // MSB first
////		bitBuffer <<= 1
////		bitCount--
////		return bit
////	}
////
////	// --- Step 4: Decode ---
////	var outBuf uint32
////	outOffset := 0
////	var written uint32
////
////for written < decompressedSize {
////	ptr := uint32(0) // start at root
////
////	for {
////		if ptr >= uint32(len(tree)) {
////			panic(fmt.Sprintf("tree pointer out of range: %d", ptr))
////		}
////
////		node := tree[ptr]
////		offset := uint32(node & 0x3F)
////		node1IsData := (node>>6)&1 != 0
////		node0IsData := (node>>7)&1 != 0
////
////		bit := getBit()
////
////		if bit == 0 {
////			if node0IsData {
////				dataAddr := (ptr &^ 1) + offset*2 + 2
////				if dataAddr >= uint32(len(tree)) {
////					panic("node0 data address out of range")
////				}
////				data := tree[dataAddr]
////				outBuf |= uint32(data) << (8 * outOffset)
////				outOffset++
////				if outOffset == 4 {
////					mem.Write32(dstAddr, outBuf)
////					dstAddr += 4
////					outBuf = 0
////					outOffset = 0
////				}
////				written++
////				break
////			} else {
////				ptr = (ptr &^ 1) + offset*2 + 2
////			}
////		} else {
////			if node1IsData {
////				dataAddr := (ptr &^ 1) + offset*2 + 3
////				if dataAddr >= uint32(len(tree)) {
////					panic("node1 data address out of range")
////				}
////				data := tree[dataAddr]
////				outBuf |= uint32(data) << (8 * outOffset)
////				outOffset++
////				if outOffset == 4 {
////					mem.Write32(dstAddr, outBuf)
////					dstAddr += 4
////					outBuf = 0
////					outOffset = 0
////				}
////				written++
////				break
////			} else {
////				ptr = (ptr &^ 1) + offset*2 + 3
////			}
////		}
////	}
////}
////
////	if outOffset > 0 {
////		mem.Write32(dstAddr, outBuf)
////	}
////    return 0
////}
//
//func DecompressDiff8bit(gba *GBA, src, dst uint32) int {
//	mem := gba.Mem
//
//	header := mem.Read32(src)
//	dataSize := int(header >> 8)
//	src += 4
//
//	end := dst + uint32(dataSize)
//	if dataSize <= 0 {
//		return 0
//	}
//
//	// First byte is raw
//	prev := mem.Read8(src)
//	mem.Write(dst, uint8(prev), false)
//	src++
//	dst++
//
//	// Remaining bytes are differences
//	for dst < end {
//		diff := int8(mem.Read8(src))
//		val := uint8(int(prev) + int(diff))
//		mem.Write(dst, val, false)
//		prev = uint32(val)
//		src++
//		dst++
//	}
//
//	return dataSize
//}
//
//func DecompressDiff16bit(gba *GBA, src, dst uint32) int {
//	mem := gba.Mem
//
//	header := mem.Read32(src)
//	dataSize := int(header >> 8)
//	src += 4
//
//	end := dst + uint32(dataSize)
//	if dataSize <= 0 || dataSize%2 != 0 {
//		return 0 // Must be even number of bytes for 16-bit data
//	}
//
//	// First 16-bit unit is raw
//	prev := mem.Read16(src)
//	mem.Write16(dst, uint16(prev))
//	src += 2
//	dst += 2
//
//	for dst < end {
//		diff := int16(mem.Read16(src))
//		val := uint16(int(prev) + int(diff))
//		mem.Write16(dst, val)
//		prev = uint32(val)
//		src += 2
//		dst += 2
//	}
//
//	return dataSize
//}
//
//func GetBiosChecksum(gba *GBA) {
//	r := &gba.Cpu.Reg.R
//	r[0] = 0xBAAE_187F
//	r[1] = 1
//	r[3] = 0x0000_4000
//}
//
//func MidiKey2Freq(gba *GBA) {
//
//    mem := gba.Mem
//	r := &gba.Cpu.Reg.R
//
//    key := float64(mem.Read32(r[0] + 4))
//    r[0] = uint32(key / math.Pow(2, (float64(180-r[1]-r[2])/256)/12))
//
//}
//
//func Huff(gba *GBA) int {
//
//    r := &gba.Cpu.Reg.R
//	mem := gba.Mem
//
//    src := r[0] &^ 0b11
//    dst := r[1]
//
//	header := mem.Read32(src)
//	dataSizeBits := int(header & 0xF)
//	if (header>>4)&0xF != 2 {
//		return 0 // Not Huffman
//	}
//	decompressedSize := int(header >> 8)
//	src += 4
//
//	treeSizeEntry := mem.Read8(src)
//	treeSize := uint32((treeSizeEntry + 1) * 2)
//	src += 1
//
//	// Load tree table
//	tree := make([]byte, treeSize)
//	for i := uint32(0); i < treeSize; i++ {
//		tree[i] = uint8(mem.Read8(src + i))
//	}
//	src += uint32(treeSize)
//
//	bitBuffer := uint32(0)
//	bitsLeft := 0
//	treeRoot := 0
//	out := dst
//	bitsPerUnit := dataSizeBits
//
//	for out < dst+uint32(decompressedSize) {
//		node := treeRoot
//		for {
//			if bitsLeft == 0 {
//				bitBuffer = mem.Read32(src)
//				src += 4
//				bitsLeft = 32
//			}
//			bit := (bitBuffer >> 31) & 1
//			bitBuffer <<= 1
//			bitsLeft--
//
//			nodeData := tree[node]
//			offset := int(nodeData & 0x3F)
//			isLeaf := false
//			if bit == 0 {
//				isLeaf = (nodeData & 0x80) != 0
//				node = (node &^ 1) + offset*2 + 2
//			} else {
//				isLeaf = (nodeData & 0x40) != 0
//				node = (node &^ 1) + offset*2 + 3
//			}
//
//			if isLeaf {
//				data := tree[node]
//				switch bitsPerUnit {
//				case 8:
//					mem.Write8(out, data)
//					out++
//				case 4:
//					// Store nibbles as packed bytes
//					if (out-dst)%2 == 0 {
//						mem.Write8(out, data&0x0F)
//					} else {
//						prev := uint8(mem.Read8(out - 1))
//						mem.Write8(out-1, prev|(data<<4))
//						out++
//					}
//				default:
//					// Unsupported size
//					return 0
//				}
//				break
//			}
//		}
//	}
//	return decompressedSize
//}
//
//
//func HuffA(gba *GBA) {
//
//    r := gba.Cpu.Reg.R
//    mem := gba.Mem
//
//    src := r[0] &^ 0b11
//    dst := r[1]
//    // is dst aligned???
//
//    // asserts to test tetris
//    assert(r[0] == 0x822305C, "R0")
//    assert(r[1] == 0x2005FEC, "R1")
//    assert(mem.Read32(src) == 0x50428, "MEM r0")
//    assert(mem.Read32(dst) == 0x20064F0, "MEM r1")
//
//    header := mem.Read32(src)
//    src += 4
//    treeHeader := mem.Read8(src)
//    src += 1
//
//    bitSize := utils.GetVarData(header, 0, 3)
//    compType := utils.GetVarData(header, 4, 7)
//    decompressedSize := utils.GetVarData(header, 8, 31)
//    treeTableSize := (treeHeader + 1) * 2
//
//    assert(compType == 2, "INVALID HUFF")
//
//    fmt.Printf("bitSize %d, decompsize %d, treetablesize %d compressedbitstream start %08X\n",
//        bitSize, decompressedSize, treeTableSize, treeTableSize + src)
//
//    assert(bitSize == 8, "BITSIZE NOT 8")
//
//    treeTable := map[uint32]*Node{}
//
//    for i := range treeTableSize {
//        treeTable[src + i] = &Node{
//            Data: mem.Read8(src + i),
//        }
//
//        //fmt.Printf("NODE ADDED %08X VALUE %02X\n", src + i, treeTable[src+ i].Data)
//    }
//
//    for nodeAddr, node := range treeTable {
//
//        if node.isLeaf {
//            continue
//        }
//
//        o := (utils.GetVarData(node.Data, 0, 5) * 2) + 2
//        base := uint32((nodeAddr) &^ 1)
//        offset := base + o
//
//        fmt.Printf("children of %08X at %08X and %08X, base %08X\n", nodeAddr, offset, offset + 1, base)
//
//        _, ok := treeTable[offset]
//        if !ok {
//            continue
//            panic(fmt.Sprintf("MISSING NODE A AT %08X", offset))
//        }
//
//        _, ok = treeTable[offset + 1]
//        if !ok {
//            continue
//            panic(fmt.Sprintf("MISSING NODE B AT %08X", offset + 1))
//        }
//
//        assert(!treeTable[offset].hasParent, "CHILD ALREADY HAS PARENT")
//        assert(!treeTable[offset + 1].hasParent, "CHILD ALREADY HAS PARENT")
//
//        node.A = treeTable[offset]
//        node.B = treeTable[offset+1]
//
//        treeTable[offset].hasParent = true
//        treeTable[offset + 1].hasParent = true
//
//        if childALeaf := utils.BitEnabled(node.Data, 6); childALeaf {
//            treeTable[offset].isLeaf = true
//        }
//
//        if childBLeaf := utils.BitEnabled(node.Data, 7); childBLeaf {
//            treeTable[offset+1].isLeaf = true
//        }
//    }
//}
//
//type Node struct {
//    hasParent bool
//    Addr uint32
//    Data uint32
//    isLeaf bool
//    A, B *Node
//}
