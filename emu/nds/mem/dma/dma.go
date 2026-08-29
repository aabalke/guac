package dma

import (
	"unsafe"

	"github.com/aabalke/guac/emu/scheduler"
)

type Irq interface {
	SetIRQ(irq uint32)
}

const (
	DMA_MODE_IMM = 0
	DMA_MODE_VBL = 1

	ARM9_DMA_MODE_HBL = 2
	ARM9_DMA_MODE_STA = 3
	ARM9_DMA_MODE_DSC = 5
	ARM9_DMA_MODE_MAI = 4
	ARM9_DMA_MODE_GBA = 6
	ARM9_DMA_MODE_GEO = 7

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

	sch *scheduler.Scheduler
	mem MemoryInterface
	irq Irq

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

	InitialGc bool
	GcDst     uint32

	gxTransferEvent scheduler.EventIdx
}

//go:inline
func ReplaceByte(value uint32, newByte uint32, byteOffset uint32) uint32 {
	bitOffset := 8 * byteOffset
	return (value &^ (0xFF << bitOffset)) | (newByte << bitOffset)
}

func (dma *DMA) Init(idx int, mem MemoryInterface, scheduler *scheduler.Scheduler, irq Irq, arm9 bool) {
	dma.Idx = idx
	dma.mem = mem
	dma.irq = irq

	dma.sch = scheduler

	dma.gxTransferEvent = scheduler.Register(dma.GxTransfer, 1)

	switch {
	case arm9:
		dma.DefaultCount = 0x200000
	case idx == 3:
		dma.DefaultCount = 0x10000
	default:
		dma.DefaultCount = 0x4000
	}
}

func (dma *DMA) Read(addr uint32) uint8 {
	switch addr {
	case 10:
		return uint8(dma.Control)
	case 11:
		return uint8(dma.Control >> 8)
	default:
		return 0
	}
}

func (dma *DMA) Write(addr uint32, v uint8) {
	switch addr {
	case 0, 1, 2, 3:
		dma.Src = (dma.Src &^ (0xFF << (addr << 3))) | (uint32(v) << (addr << 3))
		dma.InitSrc = dma.Src

	case 4, 5, 6, 7:
		addr -= 4
		dma.Dst = (dma.Dst &^ (0xFF << (addr << 3))) | (uint32(v) << (addr << 3))
		dma.InitDst = dma.Dst
		dma.GcDst = dma.Dst

	case 8:
		dma.WordCount = (dma.WordCount &^ 0xFF) | uint32(v)

	case 9:
		dma.WordCount = (dma.WordCount & 0xFF) | (uint32(v) << 8)

	case 10:

		v &= 0xE0
		dma.Control = (dma.Control &^ 0xFF) | uint32(v)
		dma.DstAdj = uint32(v>>5) & 3
		dma.SrcAdj = (dma.SrcAdj &^ 1) | (uint32(v) >> 7)

	case 11:
		wasDisabled := !dma.Enabled
		dma.Control = (dma.Control & 0xFF) | uint32(v)<<8
		dma.SrcAdj = (dma.SrcAdj & 1) | uint32(v&1)<<1
		dma.Repeat = (v>>1)&1 != 0

		dma.isWord = (v>>2)&1 != 0
		dma.Mode = uint32(v>>3) & 7
		dma.IRQ = (v>>6)&1 != 0
		dma.Enabled = (v>>7)&1 != 0

		if wasDisabled && dma.Enabled {
			dma.Src = dma.InitSrc
			dma.Dst = dma.InitDst
		}

		if isImmediate := wasDisabled && dma.CheckMode(DMA_MODE_IMM); isImmediate {
			dma.Transfer()
		}

		if wasDisabled && dma.Enabled && dma.Mode == ARM9_DMA_MODE_GEO {
			dma.sch.Schedule(dma.gxTransferEvent, 1, nil)
		}
	}
}

func (dma *DMA) disable() {
	dma.Enabled = false
	dma.Control &^= 0x8000
}

func (dma *DMA) Transfer() {
	var (
		mem       = dma.mem
		count     = dma.WordCount
		dstOffset int
		srcOffset int
		tmpDst    = dma.Dst
		tmpSrc    = dma.Src
		ofs       int
	)

	if count == 0 {
		count = dma.DefaultCount
	}

	if dma.isWord {
		tmpDst &^= 0b11
		tmpSrc &^= 0b11
		ofs = 4
	} else {
		tmpDst &^= 0b1
		tmpSrc &^= 0b1
		ofs = 2
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

	srcPtr := mem.ReadPtr(tmpSrc)
	if srcPtr != nil {
		top := uint32(int(tmpSrc) + srcOffset*int(count))
		if srcTop := mem.ReadPtr(top); srcTop == nil {
			srcPtr = nil
		}
	}

	dstPtr := mem.WritePtr(tmpDst)
	if dstPtr != nil {
		top := uint32(int(tmpDst) + dstOffset*int(count))
		if dstTop := mem.WritePtr(top); dstTop == nil {
			dstPtr = nil
		}
	}

	for range uint32(count) {
		if dma.isWord {
			if srcPtr == nil {
				dma.Value = mem.Read32(tmpSrc &^ 3)
			} else {
				dma.Value = *(*uint32)(srcPtr)
			}

			if dstPtr == nil {
				mem.Write32(tmpDst&^3, dma.Value)
			} else {
				*(*uint32)(dstPtr) = dma.Value
			}

		} else {
			if srcPtr == nil {
				dma.Value = mem.Read16(tmpSrc &^ 1)
			} else {
				dma.Value = uint32(*(*uint16)(srcPtr))
			}

			dma.Value |= (dma.Value << 16)

			if dstPtr == nil {
				mem.Write16(tmpDst&^1, uint16(dma.Value))
			} else {
				*(*uint16)(dstPtr) = uint16(dma.Value)
			}

			dma.Value = mem.Read16(tmpSrc &^ 1)
			dma.Value |= (dma.Value << 16)
			mem.Write16(tmpDst&^1, uint16(dma.Value))
		}

		tmpDst = uint32(int(tmpDst) + dstOffset)
		tmpSrc = uint32(int(tmpSrc) + srcOffset)

		if srcPtr != nil {
			srcPtr = unsafe.Add(srcPtr, srcOffset)
		}
		if dstPtr != nil {
			dstPtr = unsafe.Add(dstPtr, dstOffset)
		}
	}

	if dma.IRQ {
		dma.irq.SetIRQ(8 + uint32(dma.Idx))
	}

	if !dma.Repeat {
		// DO NOT WRITEBACK DST AND SRC UNLESS REPEAT
		dma.disable()
		return
	}

	if dma.DstAdj == DMA_ADJ_RES {
		dma.Dst = dma.InitDst
		dma.Src = tmpSrc
		return
	}

	dma.Src = tmpSrc
	dma.Dst = tmpDst
}

func (dma *DMA) CheckMode(mode uint32) bool {
	return mode == dma.Mode && dma.Enabled
}

func (dma *DMA) GamecartTransfer(arm9, initial bool) {
	const GC_SRC = 0x4100010

	if !dma.Enabled {
		return
	}

	if arm9 && dma.Mode != ARM9_DMA_MODE_DSC {
		return
	}
	if !arm9 && dma.Mode != ARM7_DMA_MODE_DSC {
		return
	}

	if notGamecart := (dma.Src == GC_SRC &&
		dma.SrcAdj == DMA_ADJ_NON &&
		dma.WordCount == 1 &&
		dma.isWord &&
		dma.Repeat); !notGamecart {
		return
	}

	mem := dma.mem

	// gamecard transfer requires recursive access.
	// Therefore, GcDst is incremented before access to not cause same dst loop

	if initial {
		dma.GcDst = dma.Dst &^ 0b11
	} else {
		dma.GcDst += 4
	}

	tmpDst := dma.GcDst &^ 0b11

	dstOffset := 4
	switch dma.DstAdj {
	case DMA_ADJ_NON:
		dstOffset = 0
	case DMA_ADJ_DEC:
		dstOffset = -4
	}

	v := mem.Read32(GC_SRC)
	mem.Write32(tmpDst, v)

	dma.Dst = uint32(int(tmpDst) + dstOffset)

	if dma.IRQ {
		dma.irq.SetIRQ(8 + uint32(dma.Idx))
	}
}

func (dma *DMA) GxTransfer(late int64, _ any) {
	if dma.Dst != 0x400_0400 || dma.DstAdj != DMA_ADJ_NON || !dma.isWord {
		return
	}

	count := dma.WordCount
	if count == 0 {
		count = dma.DefaultCount
	}

	ofs := int(2)
	if dma.isWord {
		ofs = 4
	}

	srcOffset := int(0)
	switch dma.SrcAdj {
	case DMA_ADJ_INC:
		srcOffset = ofs
	case DMA_ADJ_DEC:
		srcOffset = -ofs
	}

	mem := dma.mem
	tmpSrc := int(dma.Src &^ 3)

	ptr := mem.ReadPtr(uint32(tmpSrc))
	if ptr == nil {
		for range count {
			mem.WriteGXFIFO(mem.Read32(uint32(tmpSrc)))
			tmpSrc += srcOffset
		}
	} else {
		for range count {
			mem.WriteGXFIFO(*(*uint32)(ptr))
			ptr = unsafe.Add(ptr, srcOffset)
		}

		tmpSrc += srcOffset * int(count)
	}

	if dma.IRQ {
		dma.irq.SetIRQ(8 + uint32(dma.Idx))
	}

	if !dma.Repeat {
		dma.disable()
		return
	}

	dma.Src = uint32(tmpSrc)
	dma.sch.Schedule(dma.gxTransferEvent, 1, nil)
}

type MemoryInterface interface {
	Write8(addr uint32, v uint8)
	Write16(addr uint32, v uint16)
	Write32(addr uint32, v uint32)
	WritePtr(addr uint32) unsafe.Pointer
	WriteGXFIFO(v uint32)

	Read8(addr uint32) uint32
	Read16(addr uint32) uint32
	Read32(addr uint32) uint32
	ReadPtr(addr uint32) unsafe.Pointer
}
