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

func (gba *GBA) SysCall(opcode uint32) {
    switch inst := utils.GetVarData(opcode, 16, 23); inst {
    case SYS_SoftReset: SoftReset(gba)
    case SYS_Div: Div(gba, false)
    case SYS_DivArm: Div(gba, true)
    case SYS_Sqrt: Sqrt(gba)
    default:panic(fmt.Sprintf("EXCEPTION OR UNHANDLED SYS CALL %b\n", utils.GetVarData(opcode, 16, 23)))
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
