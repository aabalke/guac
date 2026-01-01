package nds

import (
	"fmt"

	"github.com/aabalke/guac/emu/nds/debug"
)

var validModes = map[uint32]bool{
	0b10000: true,
	0b10001: true,
	0b10010: true,
	0b10011: true,
	0b10111: true,
	0b11011: true,
	0b11111: true,
}

func (nds *Nds) checkBadPc() {

	reg9 := &nds.arm9.Reg
	reg7 := &nds.arm7.Reg

	if reg9.R[15] > 0x400_0000 && reg9.R[15] < 0xFFFF_0000 {
		panic(fmt.Sprintf("BAD ARM9 PC %08X CPSR %08X CURR %d\n", reg9.R[15], reg9.CPSR, debug.CURR_INST))
	}
	if (reg7.R[15] > 0x400_0000 && reg7.R[15] < 0x600_0000) || reg7.R[15] >= 0x700_0000 {
		panic(fmt.Sprintf("BAD ARM7 PC %08X CPSR %08X CURR %d\n", reg7.R[15], reg7.CPSR, debug.CURR_INST))
	}

	// should probably check proper vramwram for arm7

	switch {
	case reg9.CPSR.T && reg9.R[15]&0b1 != 0:
		panic(fmt.Sprintf("BAD ARM9 THUMB PC %08X CPSR %08X CURR %d\n", reg9.R[15], reg9.CPSR, debug.CURR_INST))
	case !reg9.CPSR.T && reg9.R[15]&0b11 != 0:
		panic(fmt.Sprintf("BAD ARM9 ARM   PC %08X CPSR %08X CURR %d\n", reg9.R[15], reg9.CPSR, debug.CURR_INST))
	case reg7.CPSR.T && reg7.R[15]&0b1 != 0:
		//uhh.PrintPcs()
		panic(fmt.Sprintf("BAD ARM7 THUMB PC %08X CPSR %08X CURR %d\n", reg7.R[15], reg7.CPSR, debug.CURR_INST))
	case !reg7.CPSR.T && reg7.R[15]&0b11 != 0:
		//uhh.PrintPcs()
		panic(fmt.Sprintf("BAD ARM7 ARM   PC %08X CPSR %08X CURR %d\n", reg7.R[15], reg7.CPSR, debug.CURR_INST))
	}

	zeroWordcnt := 0x100

	if nds.mem.Read32(reg9.R[15], true) == 0x0 {

		zeros := true

		for i := uint32(0); i < uint32(zeroWordcnt); i += 4 {
			if nds.mem.Read32(reg9.R[15]+i, true) != 0x0 {
				zeros = false
				break
			}
		}

		if zeros {
			panic(fmt.Sprintf("BAD ARM9 PC %08X (ZEROS) CPSR %08X CURR %d\n", reg9.R[15], reg9.CPSR, debug.CURR_INST))
		}
	}

	if nds.mem.Read32(reg7.R[15], false) == 0x0 {

		zeros := true

		for i := uint32(0); i < uint32(zeroWordcnt); i += 4 {
			if nds.mem.Read32(reg7.R[15]+i, false) != 0x0 {
				zeros = false
				break
			}
		}

		if zeros {
			panic(fmt.Sprintf("BAD ARM7 PC %08X (ZEROS) CPSR %08X CURR %d\n", reg7.R[15], reg7.CPSR, debug.CURR_INST))
		}
	}

	if reg9.R[15] < 0x30 && !nds.arm9.LowVector {
		panic(fmt.Sprintf("BAD ARM9 PC %08X (LOW WHEN HIGH) CPSR %08X CURR %d\n", reg9.R[15], reg9.CPSR, debug.CURR_INST))
	}
}

func (nds *Nds) checkMode(arm9 bool) {

	if arm9 {

		m9 := nds.arm9.Reg.CPSR.Mode
		_, valid9 := validModes[m9]
		if !valid9 {
			panic(fmt.Sprintf("ARM9 MODE INVALID %02X CURR %d\n", m9, debug.CURR_INST))
		}

		return
	}

	m7 := nds.arm7.Reg.CPSR.Mode & 0x1F
	_, valid7 := validModes[m7]
	if !valid7 {
		panic(fmt.Sprintf("ARM7 MODE INVALID %02X CURR %d\n", m7, debug.CURR_INST))
	}
}
