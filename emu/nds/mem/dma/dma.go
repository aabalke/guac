package dma

import (
	"github.com/aabalke/guac/emu/gba/utils"
	"github.com/aabalke/guac/emu/nds/cpu"
)

const (
	ARM9_DMA_MODE_IMM = 0
	ARM9_DMA_MODE_VBL = 1
	ARM9_DMA_MODE_HBL = 2
	ARM9_DMA_MODE_STA = 3
	ARM9_DMA_MODE_DSC = 5

    //unsetup
	ARM9_DMA_MODE_MAI = 4
	ARM9_DMA_MODE_GBA = 6
	ARM9_DMA_MODE_GEO = 7

	ARM7_DMA_MODE_IMM = 0
	ARM7_DMA_MODE_VBL = 1
	ARM7_DMA_MODE_DSC = 2

	ARM7_DMA_MODE_WIF = 3
	ARM7_DMA_MODE_GBA = 3

	DMA_ADJ_INC = 0
	DMA_ADJ_DEC = 1
	DMA_ADJ_NON = 2
	DMA_ADJ_RES = 3

	IRQ_DMA_0 = 8
	IRQ_DMA_1 = 9
	IRQ_DMA_2 = 10
	IRQ_DMA_3 = 11
)

type DMA struct {
	Idx int

    //Cpu *arm9.Cpu
    mem MemoryInterface
    irq *cpu.Irq

	Src     uint32
	Dst     uint32
	InitSrc uint32
	InitDst uint32

	Control   uint32
	WordCount uint32

    DefaultCount uint32

	DstAdj  uint32
	SrcAdj  uint32
	Repeat  bool
	isWord  bool
	DRQ     bool
	Mode    uint32
	IRQ     bool
	Enabled bool

	Value uint32
}

func (dma *DMA) Init(idx int, mem MemoryInterface, irq *cpu.Irq, arm9 bool) {
    dma.Idx = idx
    dma.mem = mem
    dma.irq = irq

    switch {
    case arm9:
        dma.DefaultCount = 0x200000
    case idx == 3:
        dma.DefaultCount = 0x10000
    default:
        dma.DefaultCount = 0x4000
    }
}

func (dma *DMA) ReadControl(hi bool) uint8 {
	if hi {
		return uint8(dma.Control>>8)
	}
	return uint8(dma.Control)
}

func (dma *DMA) MaskAddr(v uint32, src bool) uint32 {
	return v // & 0xFFF_FFFE
}

func (dma *DMA) WriteSrc(v uint8, byte uint32) {
	dma.Src = utils.ReplaceByte(dma.Src, uint32(v), byte)
	dma.InitSrc = dma.Src
}

func (dma *DMA) WriteDst(v uint8, byte uint32) {
	dma.Dst = utils.ReplaceByte(dma.Dst, uint32(v), byte)
	dma.InitDst = dma.Dst
}

func (dma *DMA) WriteCount(v uint8, hi bool) {

	if hi {
		dma.WordCount = (dma.WordCount & 0xFF) | (uint32(v) << 8)
		return
	}

	dma.WordCount = (dma.WordCount &^ 0xFF) | uint32(v)
}

func (dma *DMA) WriteControl(v uint8, hi bool) {

	if hi {
		a := uint32(v)
		wasDisabled := !dma.Enabled
		dma.Control = (dma.Control & 0b1111_1111) | (a << 8)
		dma.SrcAdj = (dma.SrcAdj & 1) | (a&1)<<1
		dma.Repeat = utils.BitEnabled(a, 1)
		dma.isWord = utils.BitEnabled(a, 2)
		dma.Mode = utils.GetVarData(a, 3, 5)
		dma.IRQ = utils.BitEnabled(a, 6)
		dma.Enabled = utils.BitEnabled(a, 7)

		if wasDisabled && dma.Enabled {
			dma.Src = dma.InitSrc
			dma.Dst = dma.InitDst
		}

		if isImmediate := wasDisabled && dma.CheckMode(ARM9_DMA_MODE_IMM); isImmediate {
			dma.Transfer()
		}
		return
	}

	a := uint32(v) & 0xE0
	dma.Control = (dma.Control &^ 0b1111_1111) | a
	dma.DstAdj = (uint32(a) >> 5) & 0b11
	dma.SrcAdj = (dma.SrcAdj &^ 1) | ((uint32(a) >> 7) & 1)
}

func (dma *DMA) disable() {
	dma.Enabled = false
	dma.Control &^= 0b1000_0000_0000_0000
}

func (dma *DMA) Transfer() {

	mem := dma.mem
	count := dma.WordCount

    if count == 0 {
        count = dma.DefaultCount
    }

	//if dma.Mode == DMA_MODE_HBL {
	//	srcInLimited := dma.Src >= 0x600_0000 && dma.Src < 0x800_0000
	//	dstInLimited := dma.Dst >= 0x600_0000 && dma.Dst < 0x800_0000
	//	allowed := utils.BitEnabled(uint32(dma.Gba.Mem.IO[0]), 5)

	//	if (srcInLimited || dstInLimited) && !allowed {
	//		return
	//	}
	//}

	dstOffset := int64(0)
	srcOffset := int64(0)
	tmpDst := dma.Dst
	tmpSrc := dma.Src

	if dma.isWord {
		tmpDst &^= 0b11
		tmpSrc &^= 0b11
	} else {
		tmpDst &^= 0b1
		tmpSrc &^= 0b1
	}

	//rom := tmpSrc >= 0x800_0000 && tmpSrc < 0xE00_0000
	//if rom && dma.Idx == 0 {
	//	tmpSrc &= 0x7FF_FFFF
	//}

	//if rom {
	//	dma.SrcAdj = DMA_ADJ_INC
	//}

	ofs := int64(2)
	if dma.isWord {
		ofs = 4
	}

	switch dma.DstAdj {
	case DMA_ADJ_INC, DMA_ADJ_RES:
		dstOffset = ofs
	case DMA_ADJ_DEC:
		dstOffset = -ofs
	}

	switch dma.SrcAdj {
	case DMA_ADJ_INC:
		srcOffset = ofs
	case DMA_ADJ_DEC:
		srcOffset = -ofs
	case DMA_ADJ_RES:
		panic("DMA SRC SET TO PROHIBITTED")
	}

	for i := uint32(0); i < count; i++ {

		//if eeprom := CheckEeprom(dma.Gba, tmpDst); eeprom {
		//	dstRom := tmpDst >= 0x800_0000 && tmpDst < 0xE00_0000
		//	srcRom := tmpSrc >= 0x800_0000 && tmpSrc < 0xE00_0000

		//	if count == 9 || count == 73 {
		//		cart.EepromWidth = 6
		//	} else if count == 17 || count == 81 {
		//		cart.EepromWidth = 14
		//	}

		//	if srcRom && dstRom {
		//		panic("EEPROM HAS BOTH SRC AND DST ROM ADDR")
		//	}

		//	// do not continue this., do not put this outside loop
		//}

		//badAddr := tmpSrc < 0x200_0000
		//sram := tmpSrc >= 0xE00_0000 && tmpSrc < 0x1000_0000

		if dma.isWord {
			switch {
			//case badAddr:
			//	mem.Write32(tmpDst&^3, dma.Value, true)
			//case sram && dma.Idx == 0:
			//	dma.Value = 0
			//	mem.Write32(tmpDst&^3, dma.Value, true)
			default:
				dma.Value = mem.Read32(tmpSrc &^ 3, true)
				mem.Write32(tmpDst&^3, dma.Value, true)
			}

		} else {

			switch {
			//case badAddr:
			//	mem.Write16(tmpDst&^1, uint16(dma.Value), true)
			//case sram && dma.Idx == 0:
			//	dma.Value = 0
			//	mem.Write16(tmpDst&^1, uint16(dma.Value), true)
			default:
				dma.Value = mem.Read16(tmpSrc &^ 1, true)
				dma.Value |= (dma.Value << 16)
				mem.Write16(tmpDst&^1, uint16(dma.Value), true)
			}
		}

		tmpDst = uint32(int64(tmpDst) + dstOffset)
		tmpSrc = uint32(int64(tmpSrc) + srcOffset)
	}

	//DMA_FINISHED = DMA_ACTIVE
	//DMA_ACTIVE = prevActive

	if dma.IRQ {
		dma.irq.SetIRQ(8 + uint32(dma.Idx))
	}

	if !dma.Repeat {
		// DO NOT WRITEBACK DST AND SRC UNLESS REPEAT
		dma.disable()
		return
	}

	if dma.DstAdj == DMA_ADJ_RES {
		// dma.Dst stays the same
		dma.Dst = dma.InitDst
		dma.Src = dma.MaskAddr(tmpSrc, true)
		return
	}

	dma.Src = dma.MaskAddr(tmpSrc, true)
	dma.Dst = dma.MaskAddr(tmpDst, false)
}

func (dma *DMA) CheckMode(mode uint32) bool {
	return mode == dma.Mode && dma.Enabled
}

func (dma *DMA) CheckGamecart(arm9 bool) {

    //After sending a command, data can be read from this register manually (when the DRQ bit is set), or by DMA (with DMASAD=4100010h, Fixed Source Address, Length=1, Size=32bit, Repeat=On, Mode=DS Gamecard)

    if !dma.Enabled {
        return
    }

    if !(dma.Src == 0x4100010 && dma.SrcAdj == DMA_ADJ_NON && dma.WordCount == 1 && dma.isWord && dma.Repeat) {
        return
    }

    if !(arm9 && dma.Mode == ARM9_DMA_MODE_DSC) {
        return
    }
    if !(!arm9 && dma.Mode == ARM7_DMA_MODE_DSC) {
        return
    }

    dma.Transfer()
}

type MemoryInterface interface {
    Write8(addr uint32, v uint8, arm9 bool)
    Write16(addr uint32, v uint16, arm9 bool)
    Write32(addr uint32, v uint32, arm9 bool)

    Read8(addr uint32, arm9 bool) uint32
    Read16(addr uint32, arm9 bool) uint32
    Read32(addr uint32, arm9 bool) uint32
}
