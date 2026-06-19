package gba

import (
	"github.com/aabalke/guac/emu/gba/cart"
)

//var DMA_ACTIVE = -1
//var DMA_FINISHED = -1
//var DMA_PC = uint32(0)

const (
	DMA_MODE_IMM = 0
	DMA_MODE_VBL = 1
	DMA_MODE_HBL = 2
	DMA_MODE_REF = 3

	DMA_ADJ_INC = 0
	DMA_ADJ_DEC = 1
	DMA_ADJ_NON = 2
	DMA_ADJ_RES = 3

	IRQ_DMA_0 = 8
	IRQ_DMA_1 = 9
	IRQ_DMA_2 = 10
	IRQ_DMA_3 = 11
)

type Dma struct {
	Gba *GBA
	Idx int

	Src     uint32
	Dst     uint32
	InitSrc uint32
	InitDst uint32

	Control   uint32
	WordCount uint32

	DstAdj  uint32
	SrcAdj  uint32
	Repeat  bool
	isWord  bool
	DRQ     bool
	Mode    uint8
	IRQ     bool
	Enabled bool
	Active  bool

	Value uint32
}

func NewDma(gba *GBA, idx int) *Dma {
	return &Dma{
		Gba: gba,
		Idx: idx,
	}
}

func (dma *Dma) Read(addr uint32) uint8 {
	switch addr {
	case 10:
		return uint8(dma.Control)
	case 11:
		return uint8(dma.Control >> 8)
	default:
		return 0
	}
}

func (dma *Dma) Write(addr uint32, v uint8) {
	switch addr {
	case 0, 1, 2, 3:

		if addr == 3 {
			if dma.Idx == 0 {
				v &= 0x7
			} else {
				v &= 0xF
			}
		}

		dma.Src = (dma.Src &^ (0xFF << (addr << 3))) | (uint32(v) << (addr << 3))
		dma.InitSrc = dma.Src

	case 4, 5, 6, 7:

		byte := addr - 4

		if byte == 3 {
			if dma.Idx != 0 {
				v &= 0xF
			} else {
				v &= 0x7
			}
		}

		dma.Dst = (dma.Dst &^ (0xFF << (byte << 3))) | (uint32(v) << (byte << 3))
		dma.InitDst = dma.Dst

	case 8:
		dma.WordCount = (dma.WordCount &^ 0xFF) | uint32(v)

	case 9:

		if dma.Idx != 3 {
			v &= 0x3F
		}

		dma.WordCount = (dma.WordCount & 0xFF) | (uint32(v) << 8)

	case 10:

		v &= 0xE0
		dma.Control = (dma.Control &^ 0xFF) | uint32(v)
		dma.DstAdj = uint32(v>>5) & 3
		dma.SrcAdj = (dma.SrcAdj &^ 1) | (uint32(v) >> 7)

	case 11:

		if dma.Idx != 3 {
			v &= 0xF7
		}

		prev := dma.Enabled

		dma.Repeat = (v>>1)&1 != 0
		dma.isWord = (v>>2)&1 != 0
		dma.DRQ = (v>>3)&1 != 0
		dma.Mode = (v >> 4) & 3
		dma.IRQ = (v>>6)&1 != 0
		dma.Enabled = (v>>7)&1 != 0

		dma.Control = (dma.Control & 0xFF) | (uint32(v) << 8)
		dma.SrcAdj = (dma.SrcAdj & 1) | (uint32(v)&1)<<1

		//fmt.Printf("new Dma Idx %d Mode %d Repeat %t V %02X\n", dma.Idx, dma.Mode, dma.Repeat, v)

		if !prev && dma.Enabled {
			dma.Src = dma.InitSrc
			dma.Dst = dma.InitDst

			if dma.Mode == 0 {
				switch dma.Idx {
				case 0:
					dma.Gba.Scheduler.schedule(EVENT_DMA0, 2, 1, dma.Start, nil)
				case 1:
					dma.Gba.Scheduler.schedule(EVENT_DMA1, 2, 1, dma.Start, nil)
				case 2:
					dma.Gba.Scheduler.schedule(EVENT_DMA2, 2, 1, dma.Start, nil)
				case 3:
					dma.Gba.Scheduler.schedule(EVENT_DMA3, 2, 1, dma.Start, nil)
				}
			}
		}

		if prev && !dma.Enabled {

			dma.disable()
			//fmt.Printf("disabling Dma Idx %d Mode %d Repeat %t\n", dma.Idx, dma.Mode, dma.Repeat)

			switch dma.Idx {
			case 0:
				dma.Gba.Scheduler.cancel(EVENT_DMA0)
			case 1:
				dma.Gba.Scheduler.cancel(EVENT_DMA1)
			case 2:
				dma.Gba.Scheduler.cancel(EVENT_DMA2)
			case 3:
				dma.Gba.Scheduler.cancel(EVENT_DMA3)
			}

		}
	}
}

func (dma *Dma) Start(_ int64, _ any) bool {
	dma.Active = true
	return false
}

func (dma *Dma) disable() {
	dma.Active = false
	dma.Enabled = false
	dma.Control &^= 0x8000
}

func (dma *Dma) transfer() int {
	if dma.Mode == DMA_MODE_REF {
		return 0
	}

	var (
		mem       = dma.Gba.Mem
		dstOffset = 0
		srcOffset = 0
		offset    = 2
		tmpDst    = dma.Dst
		tmpSrc    = dma.Src
	)

	//if dma.Dst >= 0xE00_0000 || dma.Src >= 0xE00_0000 {
	//	dma.disable()
	//	return 0
	//}

	if dma.isWord {
		tmpDst &^= 3
		tmpSrc &^= 3
		offset = 4
	} else {
		tmpDst &^= 1
		tmpSrc &^= 1
	}

	bios := tmpSrc < 0x200_0000

	if rom := tmpSrc >= 0x800_0000 && tmpSrc < 0xE00_0000; rom {
		dma.SrcAdj = DMA_ADJ_INC
	}

	switch dma.DstAdj {
	case DMA_ADJ_INC, DMA_ADJ_RES:
		dstOffset = offset
	case DMA_ADJ_DEC:
		dstOffset = -offset
	}

	switch dma.SrcAdj {
	case DMA_ADJ_INC:
		srcOffset = offset
	case DMA_ADJ_DEC:
		srcOffset = -offset
	case DMA_ADJ_RES:
		panic("invalid dma src method")
	}

	count := dma.WordCount
	if count == 0 {
		if dma.Idx == 3 {
			count = 0x10000
		} else {
			count = 0x4000
		}
	}

	cycles := 2

	if dma.isWord {
		for i := range count {

			dma.EepromDma(count, tmpDst, tmpSrc)

			switch {
			case bios:
				cycles++

			default:
				cycles += dma.Gba.Cpu.CycleCounterDma(tmpSrc, 4, i != 0)
				dma.Value = mem.Read32(tmpSrc)
			}

			cycles += dma.Gba.Cpu.CycleCounterDma(tmpDst, 4, i != 0)
			mem.Write32(tmpDst, dma.Value)

			tmpDst = uint32(int(tmpDst) + dstOffset)
			tmpSrc = uint32(int(tmpSrc) + srcOffset)
		}
	} else {
		for i := range count {

			dma.EepromDma(count, tmpDst, tmpSrc)

			switch {
			case bios:
				cycles++

			default:
				cycles += dma.Gba.Cpu.CycleCounterDma(tmpSrc, 2, i != 0)
				v := mem.Read16(tmpSrc)
				dma.Value = v | (v << 16)
			}

			cycles += dma.Gba.Cpu.CycleCounterDma(tmpDst, 2, i != 0)
			mem.Write16(tmpDst, uint16(dma.Value))

			tmpDst = uint32(int(tmpDst) + dstOffset)
			tmpSrc = uint32(int(tmpSrc) + srcOffset)
		}
	}

	if dma.IRQ {
		dma.Gba.Irq.SetIRQ(8 + uint32(dma.Idx))
	}

	if !dma.Repeat {
		dma.disable()
		return cycles
	}

	dma.Src = tmpSrc

	if dma.DstAdj != DMA_ADJ_RES {
		dma.Dst = tmpDst
	}

	//fmt.Printf("SRC %08X DST %08X\n", dma.Src, dma.Dst)

	return cycles
}

func (gba *GBA) checkDmas(mode uint8) {
	for i := range 4 {
		dma := gba.Dma[i]
		if ok := dma.Enabled && dma.Mode == mode; ok {
			switch i {
			case 0:
				gba.Scheduler.schedule(EVENT_DMA0, 2, 1, gba.Dma[i].Start, nil)
			case 1:
				gba.Scheduler.schedule(EVENT_DMA1, 2, 1, gba.Dma[i].Start, nil)
			case 2:
				gba.Scheduler.schedule(EVENT_DMA2, 2, 1, gba.Dma[i].Start, nil)
			case 3:
				gba.Scheduler.schedule(EVENT_DMA3, 2, 1, gba.Dma[i].Start, nil)
			}
		}
	}
}

func (gba *GBA) CheckDmas() bool {
	for i := range 4 {
		if gba.Dma[i].Active {
			gba.Tick(gba.Dma[i].transfer())
			return true
		}
	}

	return false
}

func (dma *Dma) EepromDma(count, tmpDst, tmpSrc uint32) {
	if eeprom := CheckEeprom(dma.Gba, tmpDst); eeprom {
		dstRom := tmpDst >= 0x800_0000 && tmpDst < 0xE00_0000
		srcRom := tmpSrc >= 0x800_0000 && tmpSrc < 0xE00_0000

		switch count {
		case 9, 73:
			cart.EepromWidth = 6
		case 17, 81:
			cart.EepromWidth = 14
		}

		if srcRom && dstRom {
			panic("EEPROM HAS BOTH SRC AND DST ROM ADDR")
		}

		// do not continue this., do not put this outside loop
	}
}
