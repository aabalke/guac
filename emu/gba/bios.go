package gba

import (
	"fmt"
	"math"
	"os"

	"github.com/aabalke33/guac/emu/gba/utils"
)

func (gba *GBA) LoadBios(path string) {

	buf, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	if len(buf) > len(gba.Mem.BIOS) {
		panic(fmt.Sprintf("GBA Bios of 0x%X > Allowed size 0x%X", len(buf), len(gba.Mem.BIOS)))
	}

	for i := range len(buf) {
		gba.Mem.BIOS[i] = uint8(buf[i])
	}
}

func (gba *GBA) SysCall(inst uint32) int {

	cycles := 0

	switch inst {
	case SYS_SoftReset:
		SoftReset(gba)
		cycles += 200 // approx
	case SYS_RegisterRamReset:
		RegisterRamReset(gba)
		cycles += 30 // approx
    case SYS_Halt:
        gba.Halted = true
	case SYS_IntrWait:
		IntrWait(gba)
	case SYS_VBlankIntrWait:
		VBlankIntrWait(gba)
	case SYS_Div:
		Div(gba, false)
		cycles += 36 // approx
	case SYS_DivArm:
		Div(gba, true)
		cycles += 36 // approx
	case SYS_Sqrt:
		Sqrt(gba)
        cycles += 17
	case SYS_ArcTan:
		ArcTan(gba)
        cycles += 5
	case SYS_ArcTan2:
		ArcTan2(gba)
		cycles += 520 // approx
	case SYS_CpuSet:
		CpuSet(gba)
	case SYS_CpuFastSet:
		CpuFastSet(gba)
	case SYS_BitUnPack:
		BitUnPack(gba)
    case SYS_HuffUnCompReadNormal:
        cycles += HuffUnCompReadNormal(gba)
	case SYS_LZ77UnCompReadNormalWrite8bit:
		cycles += LZ77UnCompReadNormalWrite8bit(gba)
	case SYS_LZ77UnCompReadNormalWrite16bit:
		cycles += LZ77UnCompReadNormalWrite16bit(gba)
	case SYS_RLUnCompReadNormalWrite8bit:
		cycles += RLUUnCompReadNormalWrite8bit(gba)
	case SYS_RLUnCompReadNormalWrite16bit:
		cycles += RLUUnCompReadNormalWrite16bit(gba)
    case SYS_Diff16bitUnFilter:
        cycles += DecompressDiff16bit(gba, gba.Cpu.Reg.R[0], gba.Cpu.Reg.R[1])
    case SYS_Diff8bitUnFilterWrite8bit:
        cycles += DecompressDiff8bit(gba, gba.Cpu.Reg.R[0], gba.Cpu.Reg.R[1])
    case SYS_Diff8bitUnFilterWrite16bit:
        cycles += DecompressDiff8bit(gba, gba.Cpu.Reg.R[0], gba.Cpu.Reg.R[1])
	case SYS_ObjAffineSet:
		cycles += ObjAffineSet(gba)
	case SYS_BgAffineSet:
		cycles += BGAffineSet(gba)
	case SYS_GetBiosChecksum:
		GetBiosChecksum(gba)
        cycles += 168948
    //default :gba.exception(SWI_VEC, MODE_SWI)
	default:
		panic(fmt.Sprintf("EXCEPTION OR UNHANDLED SYS CALL TYPE 0x%X\n", inst))
	}

    cycles += 6

	return cycles
}

func BGAffineSet(gba *GBA) int {

	r := &gba.Cpu.Reg.R
	mem := gba.Mem

	i := r[2]
	var ox, oy float64
	var cx, cy float64
	var sx, sy float64
	var theta float64
	offset, destination := r[0], r[1]
	var a, b, c, d float64
	var rx, ry float64
	for ; i > 0; i-- {
		// [ sx   0  0 ]   [ cos(theta)  -sin(theta)  0 ]   [ 1  0  cx - ox ]   [ A B rx ]
		// [  0  sy  0 ] * [ sin(theta)   cos(theta)  0 ] * [ 0  1  cy - oy ] = [ C D ry ]
		// [  0   0  1 ]   [     0            0       1 ]   [ 0  0     1    ]   [ 0 0  1 ]

		ox = float64(mem.Read32(offset)) / 256
		oy = float64(mem.Read32(offset+4)) / 256
		cx = float64(uint16(mem.Read16(offset + 8)))
		cy = float64(uint16(mem.Read16(offset + 10)))
		sx = float64(uint16(mem.Read16(offset+12))) / 256
		sy = float64(uint16(mem.Read16(offset+14))) / 256
		theta = (float64(mem.Read16(offset+16)>>8) / 128) * math.Pi
		offset += 20

		// Rotation
		a = math.Cos(theta)
		d = a
		b = math.Sin(theta)
		c = b

		// Scale
		a *= sx
		b *= -sx
		c *= sy
		d *= sy

		// Translate
		rx = ox - (a*cx + b*cy)
		ry = oy - (c*cx + d*cy)

		mem.Write16(destination, uint16(a*256))
		mem.Write16(destination+2, uint16(b*256))
		mem.Write16(destination+4, uint16(c*256))
		mem.Write16(destination+6, uint16(d*256))
		mem.Write32(destination+8, uint32(rx*256))
		mem.Write32(destination+12, uint32(ry*256))
		destination += 16
	}

    return 36 + (int(i) * 19)
}

func ObjAffineSet(gba *GBA) int {

	r := &gba.Cpu.Reg.R
	mem := gba.Mem

	i := r[2]
	var sx, sy float64
	var theta float64
	offset := r[0]
	destination := r[1]
	diff := r[3]
	var a, b, c, d float64
	for ; i > 0; i-- {
		// [ sx   0 ]   [ cos(theta)  -sin(theta) ]   [ A B ]
		// [  0  sy ] * [ sin(theta)   cos(theta) ] = [ C D ]
		sx = float64(uint16(mem.Read16(offset))) / 256
		sy = float64(uint16(mem.Read16(offset+2))) / 256
		theta = (float64(uint16(mem.Read16(offset+4))>>8) / 128) * math.Pi
		offset += 6

		// Rotation
		a = math.Cos(theta)
		d = a
		b = math.Sin(theta)
		c = b

		// Scale
		a *= sx
		b *= -sx
		c *= sy
		d *= sy

		mem.Write16(destination, uint16(a*256))
		mem.Write16(destination+diff, uint16(b*256))
		mem.Write16(destination+diff*2, uint16(c*256))
		mem.Write16(destination+diff*3, uint16(d*256))
		destination += diff * 4
	}

    return 13 + (int(i) * 18)
}

func IntrWait(gba *GBA) {
	//fmt.Printf("IntrWait is called, but is not completely setup\n")
}

func VBlankIntrWait(gba *GBA) {
	r := &gba.Cpu.Reg.R

	r[0] = 1
	r[1] = 1

	IntrWait(gba)
}

func SoftReset(gba *GBA) {

	// untested

	/*
	   clears 0x200 of ram
	   set r0-r12, LR_svc, SPSR_svc, LR_irq, SPSR_irq all to zero
	   enters sys mode
	   Host  sp_svc    sp_irq    sp_sys    zerofilled area       return address
	   GBA   3007FE0h  3007FA0h  3007F00h  [3007E00h..3007FFFh]  Flag[3007FFAh]
	*/

	reg := &gba.Cpu.Reg

	const (
		RETURN_ADDR = 0x0300_7FFA
		ZERO_FILL   = 0x0300_7E00
	)

	flag := gba.Mem.Read(RETURN_ADDR)

	i := uint32(0)
	for i = range 0x200 {
		gba.Mem.Write8(ZERO_FILL+i, 0)
	}

	reg.CPSR.SetMode(MODE_SWI)
	reg.R[SP] = 0x0300_7FE0
	reg.CPSR.SetMode(MODE_IRQ)
	reg.R[SP] = 0x0300_7FA0
	reg.CPSR.SetMode(MODE_SYS)
	reg.R[SP] = 0x0300_7F00

	reg.R[LR] = 0x0200_0000
	if flag > 0 {
		reg.R[LR] = 0x0800_0000
	}

	reg.CPSR.SetFlag(FLAG_T, false)

	reg.R[PC] = reg.R[LR]

	// pipelining
}

func RegisterRamReset(gba *GBA) {

	mem := gba.Mem
	r := &gba.Cpu.Reg.R
	flags := r[0]

	if clearWRAM1 := utils.BitEnabled(flags, 0); clearWRAM1 {
		mem.WRAM1 = [0x40000]uint8{}
	}

	if clearWRAM2 := utils.BitEnabled(flags, 1); clearWRAM2 {

		// need to exclude last 0x200
		for i := range 0x8000 - 0x200 {
			mem.WRAM2[i] = 0x0
		}
	}

	if clearPRAM := utils.BitEnabled(flags, 2); clearPRAM {
		mem.PRAM = [0x400]uint8{}
	}

	if clearVRAM := utils.BitEnabled(flags, 3); clearVRAM {
		mem.VRAM = [0x18000]uint8{}
	}

	if clearOAM := utils.BitEnabled(flags, 4); clearOAM {
		mem.OAM = [0x400]uint8{}
	}

	if clearSIO := utils.BitEnabled(flags, 5); clearSIO {

		for i := 0x120; i <= 0x12C; i++ {
			mem.IO[i] = 0x0
		}

		for i := 0x134; i <= 0x154; i++ {
			mem.IO[i] = 0x0
		}
	}

	if clearSound := utils.BitEnabled(flags, 6); clearSound {

		for i := 0x60; i <= 0xA8; i++ {
			mem.IO[i] = 0x0
		}
	}

	if clearOther := utils.BitEnabled(flags, 7); clearOther {
		for i := range 0x400 {

			sio1 := i >= 0x120 && i <= 0x12C
			sio2 := i >= 0x134 && i <= 0x154
			sound := i >= 0x60 && i <= 0xA8

			if sio1 || sio2 || sound {
				continue
			}

			mem.IO[i] = 0x0
		}
	}
}

func Div(gba *GBA, arm bool) {

	const MAX = 0x8000_0000
    const I32_MIN = -2147483647 - 1

	r := &gba.Cpu.Reg.R

	nu := int32(r[0])
	de := int32(r[1])

	if arm {
		tmp := nu
		nu = de
		de = tmp
	}

    switch {
    case de == 0 && nu < 0:
        r[0] = 0xFFFF_FFFF
        r[1] = uint32(nu)
        r[3] = 1
        return
    case de == 0:
        r[0] = 1
        r[1] = uint32(nu)
        r[3] = 1
        return
    case de == -1 && nu == I32_MIN:
		r[0] = MAX
		r[1] = 0
		r[3] = MAX
		return
	}

	res := uint32(nu / de)
	mod := uint32(nu % de)
	abs := uint32(math.Abs(float64(res)))

	r[0] = res
	r[1] = mod
	r[3] = abs
}

func Sqrt(gba *GBA) {

	reg := &gba.Cpu.Reg

	input := reg.R[0]

	if input == 0 {
		reg.R[0] = 0
		return
	}

	lo, hi, bound := uint32(0), input, uint32(1)

	for bound < hi {
		hi >>= 1
		bound <<= 1
	}

	for {
		hi = input
		acc := uint32(0)
		lo = bound

		for {
			oldLower := lo
			if lo <= hi>>1 {
				lo <<= 1
			}
			if oldLower >= hi>>1 {
				break
			}
		}

		for {
			acc <<= 1
			if hi >= lo {
				acc++
				hi -= lo
			}
			if lo == bound {
				break
			}
			lo >>= 1
		}

		oldBound := bound
		bound += acc
		bound >>= 1
		if bound >= oldBound {
			bound = oldBound
			break
		}
	}

	reg.R[0] = bound
}

func ArcTan(gba *GBA) {

	r := &gba.Cpu.Reg.R
	r[0], r[1], r[3] = _ArcTan(int32(r[0]))
}

func ArcTan2(gba *GBA) {

	r := &gba.Cpu.Reg.R

	x := int32(r[0])
	y := int32(r[1])

	outX := uint32(0)
	outY := uint32(0)

	switch {
	case y == 0:
		if x < 0 {
			outX = 0x8000
		}
	case x == 0:
		if y >= 0 {
			outX = 0x4000
			outY = uint32(y)
		} else {
			outX = 0xC000
			outY = uint32(y)
		}
	case y >= 0:
		if x >= 0 && x >= y {
			outX, outY, _ = _ArcTan((y << 14) / x)
		} else if -x >= y {
			outX, outY, _ = _ArcTan((y << 14) / x)
			outX += 0x8000
		} else {
			outX, outY, _ = _ArcTan((x << 14) / y)
			outX = 0x4000 - outX
		}
	case y < 0:
		if x <= 0 && -x > -y {
			outX, outY, _ = _ArcTan((y << 14) / x)
			outX += 0x8000
		} else if x >= -y {
			outX, outY, _ = _ArcTan((y << 14) / x)
			outX += 0x10000
		} else {
			outX, outY, _ = _ArcTan((x << 14) / y)
			outX = 0xC000 - outX
		}
	}

	r[0] = outX
	r[1] = outY
	r[3] = 0x170
}

func _ArcTan(src int32) (uint32, uint32, uint32) {

	a := -((src * src) >> 14)
	b := (int32(0xA9*a) >> 14) + 0x390
	b = ((b * a) >> 14) + 0x91C
	b = ((b * a) >> 14) + 0xFB6
	b = ((b * a) >> 14) + 0x16AA
	b = ((b * a) >> 14) + 0x2081
	b = ((b * a) >> 14) + 0x3651
	b = ((b * a) >> 14) + 0xA2F9

	return uint32((int32(src) * b) >> 16), uint32(a), uint32(b)
}

func BitUnPack(gba *GBA) {

	mem := gba.Mem
	r := &gba.Cpu.Reg.R
	rs := r[0]
	rd := utils.WordAlign(r[1])

	pointer := r[2]

	length := mem.Read16(pointer)
	sBitWidth := mem.Read8(pointer + 2)
	dBitWidth := mem.Read8(pointer + 3)
	s := mem.Read32(pointer + 4)

	offset := s & 0b0111_1111_1111_1111_1111_1111_1111_1111
	zeroFlag := (s>>31)&1 == 1

	if length > 0xFFFF {
		panic("bitunpack length failed")
	}

	//fmt.Printf("rs %X, rd %X, pointer %X\n", rs, rd, pointer)
	//fmt.Printf("length %X, sWidth %X, dWidth %X, s %X\n", length, sBitWidth, dBitWidth, s)

	if sBitWidth != 1 || dBitWidth != 4 || offset != 0 || zeroFlag {
		panic("LIMITED UNPACK SUPPORT")
	}

	src := []uint32{}
	dst := []uint32{}

	for i := uint32(0); i < length; i += 4 {
		v := mem.Read32(rs + i)
		src = append(src, (v>>0)&0b1111)
		src = append(src, (v>>4)&0b1111)
		src = append(src, (v>>8)&0b1111)
		src = append(src, (v>>12)&0b1111)
		src = append(src, (v>>16)&0b1111)
		src = append(src, (v>>20)&0b1111)
		src = append(src, (v>>24)&0b1111)
		src = append(src, (v>>28)&0b1111)
	}

	for i := 0; i < len(src); i += 2 {

		lo := uint32(0)
		hi := uint32(0)

		a := src[i]
		b := src[i+1]
		for j := range 8 {

			if (a>>j)&1 == 1 {
				lo |= (1 << (j * 4))
			}

			if (b>>j)&1 == 1 {
				hi |= (1 << (j * 4))
			}
		}

		dst = append(dst, (hi<<16)|lo)
	}

	for i, v := range dst {
		mem.Write32(rd+uint32(i*4), v)
	}

	return
}

func CpuSet(gba *GBA) {

	mem := gba.Mem
	r := &gba.Cpu.Reg.R

	rs := r[0]
	rd := r[1]
	info := r[2]

	wordCount := utils.GetVarData(info, 0, 20)
	fill := utils.BitEnabled(info, 24)
	isWord := utils.BitEnabled(info, 26)

	switch {
	case fill && isWord:

		rs &^= 0b11
		rd &^= 0b11

		word := mem.Read32(rs)
		for i := range wordCount {
			mem.Write32(rd+(i<<2), word)
		}

        r[0] += 4
        r[1] += wordCount * 4

	case fill && !isWord:

		rd &^= 0b1

        srcAddr := (rs)
		word := mem.Read16(srcAddr)
        if unaligned := srcAddr & 1 == 1; unaligned {
            word = mem.Read8(srcAddr)
        }

		for i := range wordCount {
            addr := rd + (i << 1)
			mem.Write16(addr, uint16(word))
		}

        r[0] += 4
        r[1] += wordCount * 2

	case !fill && isWord:

		rs &^= 0b11
		rd &^= 0b11

		for i := range wordCount {
			word := mem.Read32(rs + (i << 2))
			mem.Write32(rd+(i<<2), word)
		}

        r[0] += 4
        r[1] += wordCount * 4

	case !fill && !isWord:

		rd &^= 0b1

		for i := range wordCount {

            srcAddr := (rs + (i<<1))
			word := mem.Read16(srcAddr)
            if unaligned := srcAddr & 1 == 1; unaligned {
                word = mem.Read8(srcAddr)
            }

            dstAddr := (rd + (i<<1))
			mem.Write16(dstAddr, uint16(word))
		}

        r[0] += 4
        r[1] += wordCount * 2
	}

    r[3] = 0x170 // offical bios clobbers r3
}

func CpuFastSet(gba *GBA) {

	r := &gba.Cpu.Reg.R
	mem := gba.Mem

	src := r[0] & 0xFFFF_FFFC
	dst := r[1] & 0xFFFF_FFFC
	mode := r[2]

	count := ((mode&0x000f_ffff + 7) >> 3) << 3
	fill := utils.BitEnabled(mode, 24)
	if fill {
		word := mem.Read32(src)
		for i := uint32(0); i < count; i++ {
			mem.Write32(dst+(i<<2), word)
		}

	} else {
		for i := uint32(0); i < count; i++ {
			word := mem.Read32(src + (i << 2))
			mem.Write32(dst+(i<<2), word)
		}
	}
}

func LZ77UnCompReadNormalWrite8bit(gba *GBA) int {

	r := &gba.Cpu.Reg.R
	src := r[0]
	dst := r[1]

    bytesOutputted := DecompressLZ77(gba, src, dst)
    return bytesOutputted * 5
}

func LZ77UnCompReadNormalWrite16bit(gba *GBA) int {

	r := &gba.Cpu.Reg.R
	src := r[0]
	dst := r[1]

    bytesOutputted := DecompressLZ77(gba, src, dst)
    return bytesOutputted * 4
}

func DecompressLZ77(gba *GBA, src, dst uint32) int {

	// need to align half and pad 16bit?

	mem := gba.Mem

	header := mem.Read32(src)
	decompressedSize := int(header >> 8)
	src += 4

	end := int(dst) + decompressedSize

    bytesOutputted := 0

	for int(dst) < end {

		flagByte := mem.Read8(src)
		src++

		for i := range 8 {
			if int(dst) >= end {
				break
			}

			flag := (flagByte >> (7 - i)) & 1
			if flag == 0 {
				// Uncompressed
				mem.Write8(dst, uint8(mem.Read8(src)))
                bytesOutputted++
				dst++
				src++

			} else {
				// Compressed
				first := mem.Read8(src)
				second := mem.Read8(src + 1)

				src += 2

				length := int((first >> 4) + 3)
				disp := int(((int(first) & 0xF) << 8) | int(second))
				copyFrom := int(dst) - (disp + 1)

				for j := range length {
					mem.Write8(dst, uint8(mem.Read8(uint32(copyFrom+j))))
                    bytesOutputted++
					dst++
				}
			}
		}
	}

    return bytesOutputted
}

func RLUUnCompReadNormalWrite8bit(gba *GBA) int {

	r := &gba.Cpu.Reg.R
	src := r[0]
	dst := r[1]

    bytesOutputted := DecompressRLU(gba, src, dst)
    return bytesOutputted * 3
}

func RLUUnCompReadNormalWrite16bit(gba *GBA) int {

	r := &gba.Cpu.Reg.R
	src := r[0]
	dst := r[1]

    bytesOutputted := DecompressRLU(gba, src, dst)
    return bytesOutputted * 2
}

func DecompressRLU(gba *GBA, src, dst uint32) int {

	// need to align half and pad 16bit?

	mem := gba.Mem

	header := mem.Read32(src)
	decompressedSize := int(header >> 8)
	src += 4

	end := int(dst) + decompressedSize

    bytesOutputted := 0
	for int(dst) < end {
		flag := mem.Read8(src)
		src++

		if (flag & 0x80) == 0 {
			// Uncompressed block: copy (flag + 1) bytes
			count := int(flag&0x7F) + 1
			for range count {
				b := mem.Read8(src)
				mem.Write8(dst, uint8(b))
                bytesOutputted++
				src++
				dst++
			}
		} else {
			// Compressed block: repeat 1 byte for (flag & 0x7F) + 3 times
			count := int(flag&0x7F) + 3
			value := mem.Read8(src)
			src++
			for range count {
				mem.Write8(dst, uint8(value))
                bytesOutputted++
				dst++
			}
		}
	}

    return bytesOutputted
}


func HuffUnCompReadNormal(gba *GBA) int {

	r := &gba.Cpu.Reg.R
	src := r[0]
	dst := r[1]

    bytesOutputted := DecompressHuff(gba, src, dst)
    return bytesOutputted * 2
}
func DecompressHuff(gba *GBA, src, dst uint32) int {

	// need to align half and pad 16bit?

	mem := gba.Mem

	header := mem.Read32(src)
	decompressedSize := int(header >> 8)
	src += 4

	end := int(dst) + decompressedSize

    bytesOutputted := 0
	for int(dst) < end {
		flag := mem.Read8(src)
		src++

		if (flag & 0x80) == 0 {
			// Uncompressed block: copy (flag + 1) bytes
			count := int(flag&0x7F) + 1
			for range count {
				b := mem.Read8(src)
				mem.Write8(dst, uint8(b))
                bytesOutputted++
				src++
				dst++
			}
		} else {
			// Compressed block: repeat 1 byte for (flag & 0x7F) + 3 times
			count := int(flag&0x7F) + 3
			value := mem.Read8(src)
			src++
			for range count {
				mem.Write8(dst, uint8(value))
                bytesOutputted++
				dst++
			}
		}
	}

    return bytesOutputted
}

func DecompressDiff8bit(gba *GBA, src, dst uint32) int {
	mem := gba.Mem

	header := mem.Read32(src)
	dataSize := int(header >> 8)
	src += 4

	end := dst + uint32(dataSize)
	if dataSize <= 0 {
		return 0
	}

	// First byte is raw
	prev := mem.Read8(src)
	mem.Write8(dst, uint8(prev))
	src++
	dst++

	// Remaining bytes are differences
	for dst < end {
		diff := int8(mem.Read8(src))
		val := uint8(int(prev) + int(diff))
		mem.Write8(dst, val)
		prev = uint32(val)
		src++
		dst++
	}

	return dataSize
}

func DecompressDiff16bit(gba *GBA, src, dst uint32) int {
	mem := gba.Mem

	header := mem.Read32(src)
	dataSize := int(header >> 8)
	src += 4

	end := dst + uint32(dataSize)
	if dataSize <= 0 || dataSize%2 != 0 {
		return 0 // Must be even number of bytes for 16-bit data
	}

	// First 16-bit unit is raw
	prev := mem.Read16(src)
	mem.Write16(dst, uint16(prev))
	src += 2
	dst += 2

	for dst < end {
		diff := int16(mem.Read16(src))
		val := uint16(int(prev) + int(diff))
		mem.Write16(dst, val)
		prev = uint32(val)
		src += 2
		dst += 2
	}

	return dataSize
}

func GetBiosChecksum(gba *GBA) {
	r := &gba.Cpu.Reg.R
	r[0] = 0xBAAE_187F
	r[1] = 1
	r[3] = 0x0000_4000
}
