package gba

import (
	"github.com/aabalke/guac/emu/gba/cart"
)

var EVENTS = []Event{EVENT_DMA0, EVENT_DMA1, EVENT_DMA2, EVENT_DMA3}

var isDmas uint8

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
	Gba  *GBA
	Tick func(cycles int)
	Idx  int

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
		Gba:  gba,
		Idx:  idx,
		Tick: gba.Tick,
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

		if !prev && dma.Enabled && dma.Mode == 0 {
			dma.Gba.Scheduler.schedule(EVENTS[dma.Idx], 0, 2, dma.Start, nil)
			return
		}

		if prev && !dma.Enabled {
			dma.disable()
			dma.Gba.Scheduler.cancel(EVENTS[dma.Idx])
		}
	}
}

func (dma *Dma) Start(late int64, _ any) {
	var (
		src       = dma.Src
		dst       = dma.Dst
		cnt       = dma.Cnt
		dstOffset = 2
		srcOffset = 2
	)

	dma.Active = true
	isDmas |= 1 << dma.Idx

	if dma.isWord {
		dst &^= 3
		src &^= 3
		dstOffset = 4
		srcOffset = 4
	} else {
		dst &^= 1
		src &^= 1
	}

	if cnt == 0 {
		if dma.Idx == 3 {
			cnt = 0x10000
		} else {
			cnt = 0x4000
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
	dma.latched.cnt = cnt
	dma.latched.srcOffset = srcOffset
	dma.latched.dstOffset = dstOffset

	dma.EepromDma(dma.latched.cnt, dst, src)
}

func (dma *Dma) disable() {
	dma.Enabled = false
	dma.Control &^= 0x8000
}

func (dma *Dma) transfer() {
	var (
		mem       = dma.Gba.Mem
		src       = dma.latched.src
		dst       = dma.latched.dst
		accessRom = false
	)

	for range dma.latched.cnt {

		srcSeq := true
		dstSeq := true

		if !accessRom {
			if src >= 0x800_0000 {
				srcSeq = false
				accessRom = true
			} else if dst >= 0x800_0000 {
				dstSeq = false
				accessRom = true
			}
		}

		if dma.isWord {

			if src < 0x200_0000 {
				dma.Tick(1)
			} else {
				dma.Gba.Cpu.Cycles(src, 4, true, srcSeq, false)
				dma.Value = mem.Read32(src)
			}

			dma.Gba.Cpu.Cycles(dst, 4, true, dstSeq, false)
			mem.Write32(dst, dma.Value)

		} else {

			v := uint32(0)

			if src < 0x200_0000 {

				// required for ngba-suite/latch.gba
				if dst&2 != 0 {
					v = dma.Value >> 16
				} else {
					v = dma.Value & 0xFFFF
				}

				dma.Tick(1)
			} else {

				dma.Gba.Cpu.Cycles(src, 2, true, srcSeq, false)
				v = mem.Read16(src)
				dma.Value = v | (v << 16)
			}

			dma.Gba.Cpu.Cycles(dst, 2, true, dstSeq, false)
			mem.Write16(dst, uint16(v))
		}

		dst = uint32(int(dst) + dma.latched.dstOffset)
		src = uint32(int(src) + dma.latched.srcOffset)
	}

	if dma.IRQ {
		dma.Gba.Irq.SetIRQ(8 + uint32(dma.Idx))
	}

	dma.Active = false
	isDmas &^= 1 << dma.Idx
	dma.latched.src = src
	dma.latched.dst = dst

	if !dma.Repeat {
		dma.disable()
		return
	}

	dma.Src = src

	if dma.DstAdj != DMA_ADJ_REL {
		dma.Dst = dst
	}
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
			dma.Gba.Scheduler.schedule(EVENT_DMA3, 0, 2, dma.Start, nil)
		}
	}
}

func (gba *GBA) checkDmas(mode uint8) {
	for i := range 4 {
		dma := gba.Dma[i]
		if ok := dma.Enabled && dma.Mode == mode; ok {
			dma.Gba.Scheduler.schedule(EVENTS[dma.Idx], 0, 2, dma.Start, nil)
		}
	}
}

func (gba *GBA) CheckDmas() {
	if isDmas == 0 {
		return
	}

	gba.Tick(1)

	for i := range 4 {
		if dma := gba.Dma[i]; dma.Active {
			dma.transfer()
		}
	}

	gba.Tick(1)
}

func (dma *Dma) EepromDma(count, dst, src uint32) {
	if !CheckEeprom(dma.Gba, dst) {
		return
	}

	switch count {
	case 9, 73:
		cart.EepromWidth = 6
	case 17, 81:
		cart.EepromWidth = 14
	}

	dstRom := dst >= 0x800_0000 && dst < 0xE00_0000
	srcRom := src >= 0x800_0000 && src < 0xE00_0000
	if srcRom && dstRom {
		panic("EEPROM HAS BOTH SRC AND DST ROM ADDR")
	}
}
