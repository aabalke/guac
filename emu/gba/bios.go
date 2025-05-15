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

    for i:= range len(buf) {
        gba.Mem.BIOS[i] = uint8(buf[i])
    }
}

func (gba *GBA) SysCall(inst uint32) {
    switch inst {
    //switch inst := utils.GetVarData(opcode, 16, 23); inst {
    case SYS_SoftReset: SoftReset(gba)
    case SYS_RegisterRamReset: RegisterRamReset(gba)
    case SYS_Div: Div(gba, false)
    case SYS_DivArm: Div(gba, true)
    case SYS_Sqrt: Sqrt(gba)
    case SYS_CpuSet: CpuSet(gba)
    case SYS_BitUnPack: BitUnPack(gba)
    default:panic(fmt.Sprintf("EXCEPTION OR UNHANDLED SYS CALL TYPE 0x%X\n", inst))
    }
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
        for i := range (0x8000 - 0x200) {
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

	r := &gba.Cpu.Reg.R

    nu := int32(r[0])
    de := int32(r[1])

    if arm {
        tmp := nu
        nu = de
        de = tmp
    }

    if de == 0 {
        panic("SYS CALL DIV BY 0. WHAT HAPPENED")
    }

    if de == -1 && nu == -MAX {
        r[0] = MAX
        r[1] = 0
        r[2] = MAX
        return
    }

    res := uint32(nu / de)
    mod := uint32(nu % de)
    abs := uint32(math.Abs(float64(res)))

    r[0] = res
    r[1] = mod
    r[2] = abs
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

func ArcTan()  { panic("ARCTAN IS NOT FUNCTIONAL") }
func ArcTan2() { panic("ARCTAN2 IS NOT FUNCTIONAL") }

func BitUnPack(gba *GBA) {

    mem := gba.Mem
    r := &gba.Cpu.Reg.R
    rs := r[0]
    rd := utils.WordAlign(r[1])

    pointer := r[2]

    length := mem.Read16(pointer)
    sBitWidth := mem.Read8(pointer+2)
    dBitWidth := mem.Read8(pointer+3)
    s := mem.Read32(pointer+4)

    offset := s & 0b0111_1111_1111_1111_1111_1111_1111_1111
    zeroFlag := (s >> 31) & 1 == 1

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
        v := mem.Read32(rs+i)
        src = append(src, (v >> 0) & 0b1111)
        src = append(src, (v >> 4) & 0b1111)
        src = append(src, (v >> 8) & 0b1111)
        src = append(src, (v >> 12) & 0b1111)
        src = append(src, (v >> 16) & 0b1111)
        src = append(src, (v >> 20) & 0b1111)
        src = append(src, (v >> 24) & 0b1111)
        src = append(src, (v >> 28) & 0b1111)
    }

    for i := 0; i < len(src); i += 2 {

        lo := uint32(0)
        hi := uint32(0)

        a := src[i]
        b := src[i+1]
        for j := range 8 {

            if (a >> j) & 1 == 1 {
                lo |= (1 << (j * 4))
            }

            if (b >> j) & 1 == 1 {
                hi |= (1 << (j * 4))
            }
        }

        dst = append(dst, (hi << 16) | lo)
    }

    for i, v := range dst {
        mem.Write32(rd + uint32(i * 4), v)
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

        rs &= 0xfffffffc
        rd &= 0xfffffffc

        word := mem.Read32(rs)
        for i := range wordCount {
            mem.Write32(rd+(i<<2), word)
        }

    case fill && !isWord:

        rs &= 0xfffffffe
        rd &= 0xfffffffe

        word := mem.Read16(rs)
        for i := range wordCount {
            mem.Write16(rd+(i<<1), uint16(word))
        }

    case !fill && isWord:

        rs &= 0xfffffffc
        rd &= 0xfffffffc

        for i := range wordCount {
            word := mem.Read32(rs + (i << 2))
            mem.Write32(rd+(i<<2), word)
        }

    case !fill && !isWord:

        rs &= 0xfffffffe
        rd &= 0xfffffffe

        for i := range wordCount {
            word := mem.Read16(rs + (i << 1))
            mem.Write16(rd+(i<<1), uint16(word))
        }
    }
}
