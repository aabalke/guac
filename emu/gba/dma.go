package gba

import (
	"github.com/aabalke/guac/emu/gba/cart"
)

const (
	DMA_MODE_IMM = 0
	DMA_MODE_VBL = 1
	DMA_MODE_HBL = 2
	DMA_MODE_SPE = 3

	DMA_ADJ_INC = 0
	DMA_ADJ_DEC = 1
	DMA_ADJ_NON = 2
	DMA_ADJ_REL = 3
)

type Dma struct {
	Gba *GBA
	Idx int

	Src     uint32
	Dst     uint32
	Value   uint32
	Control uint32
	Cnt     uint32
	DstAdj  uint32
	SrcAdj  uint32
	Mode    uint8

	Repeat      bool
	isWord      bool
	DRQ         bool
	IRQ         bool
	Enabled     bool
	Active      bool
	InVideoMode bool

	latched struct {
		src       uint32
		dst       uint32
		cnt       uint32
		dstOffset int
		srcOffset int
	}
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

	case 4, 5, 6, 7:

		addr -= 4

		if addr == 3 {
			if dma.Idx == 3 {
				v &= 0xF
			} else {
				v &= 0x7
			}
		}

		dma.Dst = (dma.Dst &^ (0xFF << (addr << 3))) | (uint32(v) << (addr << 3))

	case 8:
		dma.Cnt = (dma.Cnt &^ 0xFF) | uint32(v)

	case 9:

		if dma.Idx != 3 {
			v &= 0x3F
		}

		dma.Cnt = (dma.Cnt & 0xFF) | (uint32(v) << 8)

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

		if !prev && dma.Enabled {
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
	var (
		src       = dma.Src
		dst       = dma.Dst
		count     = dma.Cnt
		dstOffset = 2
		srcOffset = 2
	)

	dma.Active = true
	if dma.isWord {
		dst &^= 3
		src &^= 3
		dstOffset = 4
		srcOffset = 4
	} else {
		dst &^= 1
		src &^= 1
	}

	if count == 0 {
		if dma.Idx == 3 {
			count = 0x10000
		} else {
			count = 0x4000
		}
	}

	if rom := src >= 0x800_0000 && src < 0xE00_0000; !rom {
		switch dma.SrcAdj {
		case DMA_ADJ_NON:
			srcOffset = 0
		case DMA_ADJ_DEC:
			srcOffset = -srcOffset
		case DMA_ADJ_REL:
			panic("invalid dma src method")
		}
	}

	if rom := dst >= 0x800_0000 && dst < 0xE00_0000; !rom {
		switch dma.DstAdj {
		case DMA_ADJ_NON:
			dstOffset = 0
		case DMA_ADJ_DEC:
			dstOffset = -dstOffset
		}
	}

	dma.latched.src = src
	dma.latched.dst = dst
	dma.latched.cnt = count
	dma.latched.srcOffset = srcOffset
	dma.latched.dstOffset = dstOffset

	return false
}

func (dma *Dma) disable() {
	dma.Enabled = false
	dma.Control &^= 0x8000
}

func (dma *Dma) transfer() int {
	var (
		mem = dma.Gba.Mem
		src = dma.latched.src
		dst = dma.latched.dst
	)

	bios := src < 0x200_0000

	cycles := 2

	if dma.isWord {
		for i := range dma.latched.cnt {

			dma.EepromDma(dma.latched.cnt, dst, src)

			switch {
			case bios:
				cycles++

			default:
				cycles += dma.Gba.Cpu.CycleCounterDma(src, 4, i != 0)
				dma.Value = mem.Read32(src)
			}

			cycles += dma.Gba.Cpu.CycleCounterDma(dst, 4, i != 0)
			mem.Write32(dst, dma.Value)

			dst = uint32(int(dst) + dma.latched.dstOffset)
			src = uint32(int(src) + dma.latched.srcOffset)
		}
	} else {
		for i := range dma.latched.cnt {

			dma.EepromDma(dma.latched.cnt, dst, src)

			switch {
			case bios:
				cycles++

			default:
				cycles += dma.Gba.Cpu.CycleCounterDma(src, 2, i != 0)
				v := mem.Read16(src)
				dma.Value = v | (v << 16)
			}

			cycles += dma.Gba.Cpu.CycleCounterDma(dst, 2, i != 0)
			mem.Write16(dst, uint16(dma.Value))

			dst = uint32(int(dst) + dma.latched.dstOffset)
			src = uint32(int(src) + dma.latched.srcOffset)
		}
	}

	if dma.IRQ {
		dma.Gba.Irq.SetIRQ(8 + uint32(dma.Idx))
	}

	dma.Active = false

	dma.latched.src = src
	dma.latched.dst = dst

	if !dma.Repeat {
		dma.disable()
		return cycles
	}

	dma.Src = src

	if dma.DstAdj != DMA_ADJ_REL {
		dma.Dst = dst
	}

	return cycles
}

func (dma *Dma) videoDma(vcount uint8) {
	if ok := dma.Enabled && dma.Mode == DMA_MODE_SPE; ok {

		if vcount == 2 {
			dma.InVideoMode = true
		}

		if vcount == 162 {
			dma.disable()
			dma.InVideoMode = false
		}

		if dma.InVideoMode {
			dma.Gba.Scheduler.schedule(EVENT_DMA3, 2, 1, dma.Start, nil)
		}
	}
}

func (gba *GBA) checkDmas(mode uint8) {
	for i := range 4 {
		dma := gba.Dma[i]
		if ok := dma.Enabled && dma.Mode == mode; ok {
			switch i {
			case 0:
				gba.Scheduler.schedule(EVENT_DMA0, 2, 1, dma.Start, nil)
			case 1:
				gba.Scheduler.schedule(EVENT_DMA1, 2, 1, dma.Start, nil)
			case 2:
				gba.Scheduler.schedule(EVENT_DMA2, 2, 1, dma.Start, nil)
			case 3:
				gba.Scheduler.schedule(EVENT_DMA3, 2, 1, dma.Start, nil)
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

		switch count {
		case 9, 73:
			cart.EepromWidth = 6
		case 17, 81:
			cart.EepromWidth = 14
		}

		dstRom := tmpDst >= 0x800_0000 && tmpDst < 0xE00_0000
		srcRom := tmpSrc >= 0x800_0000 && tmpSrc < 0xE00_0000
		if srcRom && dstRom {
			panic("EEPROM HAS BOTH SRC AND DST ROM ADDR")
		}

		// do not continue this., do not put this outside loop
	}
}
