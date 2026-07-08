package gba

import (
	"unsafe"
)

var EVENTS = []Event{EVENT_DMA0, EVENT_DMA1, EVENT_DMA2, EVENT_DMA3}

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
	Gba           *GBA
	Tick          func(cycles int64)
	Chs           [4]Channel
	Runnable      uint8
	ActiveDma     int
	ShouldReEnter bool
}

type Channel struct {
	dma     *Dma
	Idx     int
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
		srcPtr    unsafe.Pointer
		dstPtr    unsafe.Pointer
		src       uint32
		dst       uint32
		cnt       uint32
		dstOffset int
		srcOffset int
	}
}

func NewDma(gba *GBA) *Dma {
	d := &Dma{
		Gba:  gba,
		Tick: gba.Tick,
	}

	d.Chs[0].Idx = 0
	d.Chs[1].Idx = 1
	d.Chs[2].Idx = 2
	d.Chs[3].Idx = 3

	d.Chs[0].dma = d
	d.Chs[1].dma = d
	d.Chs[2].dma = d
	d.Chs[3].dma = d

	return d
}

func (d *Dma) Read(idx, addr uint32) uint8 {
	switch addr {
	case 10:
		return uint8(d.Chs[idx].Control)
	case 11:
		return uint8(d.Chs[idx].Control >> 8)
	default:
		return 0
	}
}

func (d *Dma) Write(idx, addr uint32, v uint8) {
	ch := &d.Chs[idx]

	switch addr {
	case 0, 1, 2, 3:

		if addr == 3 {
			if ch.Idx == 0 {
				v &= 0x7
			} else {
				v &= 0xF
			}
		}

		ch.Src = (ch.Src &^ (0xFF << (addr << 3))) | (uint32(v) << (addr << 3))

	case 4, 5, 6, 7:

		addr -= 4

		if addr == 3 {
			if ch.Idx == 3 {
				v &= 0xF
			} else {
				v &= 0x7
			}
		}

		ch.Dst = (ch.Dst &^ (0xFF << (addr << 3))) | (uint32(v) << (addr << 3))

	case 8:
		ch.Cnt = (ch.Cnt &^ 0xFF) | uint32(v)

	case 9:

		if ch.Idx != 3 {
			v &= 0x3F
		}

		ch.Cnt = (ch.Cnt & 0xFF) | (uint32(v) << 8)

	case 10:

		v &= 0xE0
		ch.Control = (ch.Control &^ 0xFF) | uint32(v)
		ch.DstAdj = uint32(v>>5) & 3
		ch.SrcAdj = (ch.SrcAdj &^ 1) | (uint32(v) >> 7)

	case 11:

		if ch.Idx != 3 {
			v &= 0xF7
		}

		prev := ch.Enabled

		ch.Repeat = (v>>1)&1 != 0
		ch.isWord = (v>>2)&1 != 0
		ch.DRQ = (v>>3)&1 != 0
		ch.Mode = (v >> 4) & 3
		ch.IRQ = (v>>6)&1 != 0
		ch.Enabled = (v>>7)&1 != 0

		ch.Control = (ch.Control & 0xFF) | (uint32(v) << 8)
		ch.SrcAdj = (ch.SrcAdj & 1) | (uint32(v)&1)<<1

		//if B[15] && !prev && ch.Enabled && ch.Mode == 2 {
		//	fmt.Printf("Enabling Hblank Dma\n")
		//}

		if !prev && ch.Enabled && ch.Mode == 0 {
			d.Gba.Scheduler.schedule(EVENTS[ch.Idx], 0, 2, ch.Start, nil)
			return
		}

		if prev && !ch.Enabled {
			ch.disable()
			d.Gba.Scheduler.cancel(EVENTS[ch.Idx])
			return
		}
	}
}

func (ch *Channel) Start(late int64, _ any) {
	var (
		src       = ch.Src
		dst       = ch.Dst
		cnt       = ch.Cnt
		dstOffset = 2
		srcOffset = 2
	)

	ch.Active = true

	if ch.isWord {
		dst &^= 3
		src &^= 3
		dstOffset = 4
		srcOffset = 4
	} else {
		dst &^= 1
		src &^= 1
	}

	if cnt == 0 {
		if ch.Idx == 3 {
			cnt = 0x10000
		} else {
			cnt = 0x4000
		}
	}

	if rom := src >= 0x800_0000 && src < 0xE00_0000; !rom {
		switch ch.SrcAdj {
		case DMA_ADJ_NON:
			srcOffset = 0
		case DMA_ADJ_DEC:
			srcOffset = -srcOffset
		case DMA_ADJ_REL:
			panic("invalid ch.src method")
		}
	}

	if rom := dst >= 0x800_0000 && dst < 0xE00_0000; !rom {
		switch ch.DstAdj {
		case DMA_ADJ_NON:
			dstOffset = 0
		case DMA_ADJ_DEC:
			dstOffset = -dstOffset
		}
	}

	ch.latched.src = src
	ch.latched.dst = dst
	ch.latched.cnt = cnt
	ch.latched.srcOffset = srcOffset
	ch.latched.dstOffset = dstOffset

	ch.dma.EepromDma(ch.latched.cnt, dst)

	switch {
	case ch.dma.Runnable == 0:
		ch.dma.ActiveDma = ch.Idx
	case ch.Idx < ch.dma.ActiveDma:
		//fmt.Printf("New Ch %d is < ActiveDma %b\n", ch.Idx, ch.dma.ActiveDma)
		ch.dma.ActiveDma = ch.Idx
		ch.dma.ShouldReEnter = true
	}

	ch.dma.Runnable |= 1 << ch.Idx
}

func (ch *Channel) disable() {
	ch.Enabled = false
	ch.Control &^= 0x8000
	ch.Active = false
	ch.dma.Runnable &^= 1 << ch.Idx
}

func (ch *Channel) transfer() {
	var (
		mem       = ch.dma.Gba.Mem
		src       = ch.latched.src
		dst       = ch.latched.dst
		accessRom = false
	)

	ch.latched.srcPtr = ch.dma.Gba.Mem.ReadPtr(src)
	ch.latched.dstPtr = ch.dma.Gba.Mem.WritePtr(dst)

	for ch.latched.cnt > 0 {

		if ch.dma.ShouldReEnter {
			ch.dma.ShouldReEnter = false
			return
		}

		src = ch.latched.src
		dst = ch.latched.dst

		srcSeq := uint32(SEQ)
		dstSeq := uint32(SEQ)

		if !accessRom {
			if src >= 0x800_0000 {
				srcSeq = NONSEQ
				accessRom = true
			} else if dst >= 0x800_0000 {
				dstSeq = NONSEQ
				accessRom = true
			}
		}

		if ch.isWord {

			if src < 0x200_0000 {
				ch.dma.Tick(1)
			} else {
				ch.dma.Gba.Cpu.CyclesDma(src, 4, srcSeq)
				if ch.latched.srcPtr == nil {
					ch.Value = mem.Read32(src)
				} else {
					ch.Value = *(*uint32)(ch.latched.srcPtr)
				}
			}

			ch.dma.Gba.Cpu.CyclesDma(dst, 4, dstSeq)
			if ch.latched.dstPtr == nil {
				mem.Write32(dst, ch.Value)
			} else {
				*(*uint32)(ch.latched.dstPtr) = ch.Value
			}

		} else {

			v := uint32(0)

			if src < 0x200_0000 {

				// required for ngba-suite/latch.gba
				if dst&2 != 0 {
					v = ch.Value >> 16
				} else {
					v = ch.Value & 0xFFFF
				}

				ch.dma.Tick(1)
			} else {

				ch.dma.Gba.Cpu.CyclesDma(src, 2, srcSeq)

				if ch.latched.srcPtr == nil {
					v = mem.Read16(src)
				} else {
					v = *(*uint32)(ch.latched.srcPtr) & 0xFFFF
				}
				ch.Value = v | (v << 16)
			}

			ch.dma.Gba.Cpu.CyclesDma(dst, 2, dstSeq)
			if ch.latched.dstPtr == nil {
				mem.Write16(dst, uint16(v))
			} else {
				*(*uint16)(ch.latched.dstPtr) = uint16(v)
			}
		}

		ch.latched.src = uint32(int(ch.latched.src) + ch.latched.srcOffset)
		ch.latched.dst = uint32(int(ch.latched.dst) + ch.latched.dstOffset)
		if ch.latched.srcPtr != nil {
			ch.latched.srcPtr = unsafe.Add(ch.latched.srcPtr, ch.latched.srcOffset)
		}
		if ch.latched.dstPtr != nil {
			ch.latched.dstPtr = unsafe.Add(ch.latched.dstPtr, ch.latched.dstOffset)
		}

		ch.latched.cnt--
	}

	if ch.IRQ {
		ch.dma.Gba.Irq.SetIRQ(8 + uint32(ch.Idx))
	}

	ch.Active = false
	ch.dma.Runnable &^= 1 << ch.Idx
	ch.dma.SelectNextChannel(ch.Idx)

	if !ch.Repeat {
		ch.disable()
		return
	}

	ch.Src = ch.latched.src

	if ch.DstAdj != DMA_ADJ_REL {
		ch.Dst = ch.latched.dst
	}
}

func (d *Dma) SelectNextChannel(curr int) {
	if curr == 3 || d.Runnable == 0 {
		return
	}

	for i := curr; i < 4; i++ {
		if d.Runnable&(1<<i) != 0 {
			d.ActiveDma = i
		}
	}
}

func (d *Dma) videoDma(vcount uint8) {
	ch := &d.Chs[3]

	if ok := ch.Enabled && ch.Mode == DMA_MODE_SPE; ok {

		if vcount == 2 {
			ch.InVideoMode = true
		}

		if vcount == 162 {
			ch.disable()
			ch.InVideoMode = false
		}

		if ch.InVideoMode {
			d.Gba.Scheduler.schedule(EVENT_DMA3, 0, 2, ch.Start, nil)
		}
	}
}

func (d *Dma) IsRunning() bool {
	return d.Runnable != 0
}

func (gba *GBA) checkDmas(mode uint8) {
	for i := range 4 {
		ch := &gba.Dma.Chs[i]
		if ok := ch.Enabled && ch.Mode == mode; ok {
			gba.Scheduler.schedule(EVENTS[ch.Idx], 0, 2, ch.Start, nil)
		}
	}
}

func (gba *GBA) CheckDmas() uint32 {
	start := gba.Scheduler.CurrentCycle

	gba.Tick(1)

	for gba.Dma.IsRunning() {

		ch := &gba.Dma.Chs[gba.Dma.ActiveDma]
		ch.transfer()
	}

	gba.Tick(1)

	return uint32(gba.Scheduler.CurrentCycle - start)
}

func (d *Dma) EepromDma(count, dst uint32) {
	if !CheckEeprom(d.Gba, dst) {
		return
	}

	switch count {
	case 9, 73:
		d.Gba.Cartridge.EepromWidth = 6
	case 17, 81:
		d.Gba.Cartridge.EepromWidth = 14
	}
}
