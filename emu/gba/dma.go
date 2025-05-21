package gba

import (
	//"fmt"

	"github.com/aabalke33/guac/emu/gba/utils"
)

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

type DMA struct {

    Gba *GBA
    Idx int

	Src uint32
	Dst uint32

	Control       uint32
	singleByteSet bool
	WordCount     uint32

	DstAdj  uint32
	SrcAdj  uint32
	Repeat  bool
	isWord  bool
    DRQ     bool
	Mode    uint32
	IRQ     bool
	Enabled bool
}

func (dma *DMA) ReadSrc(byte uint32) uint8 {
    bitOffset := 8 * byte
    return uint8(dma.Src >> bitOffset) & 0xF
}

func (dma *DMA) ReadDst(byte uint32) uint8 {
    bitOffset := 8 * byte
    return uint8(dma.Dst >> bitOffset) & 0xF
}

func (dma *DMA) ReadCount(hi bool) uint8 {
    if hi { return uint8(dma.WordCount >> 8) & 0xFF }
    return uint8(dma.WordCount) & 0xFF
}

func (dma *DMA) ReadControl(hi bool) uint8 {
    if hi { return uint8(dma.Control >> 8) & 0xFF }
    return uint8(dma.Control) & 0xFF
}

func (dma *DMA) WriteSrc(v uint8, byte uint32) {
    dma.Src = utils.ReplaceByte(dma.Src, uint32(v), byte)
}

func (dma *DMA) WriteDst(v uint8, byte uint32) {
    dma.Dst = utils.ReplaceByte(dma.Dst, uint32(v), byte)
}

func (dma *DMA) WriteCount(v uint8, hi bool) {

    if hi {
        dma.WordCount = (dma.WordCount & 0b1111_1111) |  (uint32(v) << 8)
        return
    }

	dma.WordCount = (dma.WordCount &^ 0b1111_1111) |  uint32(v)
}

func (dma *DMA) WriteControl(v uint8, hi bool) {

	if hi {
		dma.Control = (dma.Control & 0b1111_1111) | (uint32(v) << 8)
		dma.SrcAdj = (dma.SrcAdj &^ 0b10) | (uint32(v) & 1) << 1
		dma.Repeat = utils.BitEnabled(uint32(v), 1)
		dma.isWord = utils.BitEnabled(uint32(v), 2)
        dma.DRQ = utils.BitEnabled(uint32(v), 3)
        dma.Mode = utils.GetVarData(uint32(v), 4, 5)
		dma.IRQ = utils.BitEnabled(uint32(v), 6)
		dma.Enabled = utils.BitEnabled(uint32(v), 7)

	} else {
		dma.Control = (dma.Control &^ 0b1111_1111) | uint32(v)
		dma.DstAdj = (uint32(v) >> 5) & 0b11
		dma.SrcAdj = (dma.SrcAdj &^ 0b1) | ((uint32(v) >> 6) & 1)
	}

	if dma.Mode >= 0b100 {
		panic("DMA MODE HAS BEEN SET OVER b11")
	}

    // need to make sure entire 16 bit of control is written before transfer
	dma.singleByteSet = !dma.singleByteSet
	if dma.singleByteSet {
		return
	}

	// immediate should be 2 cycles after enabling
	if isImmediate := dma.checkMode(DMA_MODE_IMM); isImmediate {
		dma.transfer()
	}
}

func (dma *DMA) disable() {
    dma.Enabled = false
    dma.Control &^= 0b1000_0000_0000_0000
}

func (dma *DMA) transfer() {

    mem := dma.Gba.Mem

    count := dma.WordCount

    //if sram := (dma.Dst >= 0xE00_0000 || dma.Src >= 0xE00_0000); sram {
    //    return
    //}

    //dstInPak := (dma.Dst >= 0x800_0000 && dma.Dst < 0xE00_0000)
    //srcInPak := (dma.Src >= 0x800_0000 && dma.Src < 0xE00_0000)
    //if srcInPak && dma.Idx != 3 {
    //    // not sure how to handle drq
    //    dma.disable()
    //    return
    //}

    dstOffset := int64(0)
    srcOffset := int64(0)

    if dma.isWord {
        dma.Dst &^= 0b11
        dma.Src &^= 0b11
    } else {
        dma.Dst &^= 0b1
        dma.Src &^= 0b1
    }

    switch {
    case dma.isWord && dma.DstAdj  == DMA_ADJ_INC: dstOffset = 4
    case !dma.isWord && dma.DstAdj == DMA_ADJ_INC: dstOffset = 2
    case dma.isWord && dma.DstAdj  == DMA_ADJ_DEC: dstOffset = -4
    case !dma.isWord && dma.DstAdj == DMA_ADJ_DEC: dstOffset = -2
    }

    switch {
    case dma.isWord && dma.SrcAdj  == DMA_ADJ_INC: srcOffset = 4
    case !dma.isWord && dma.SrcAdj == DMA_ADJ_INC: srcOffset = 2
    case dma.isWord && dma.SrcAdj  == DMA_ADJ_DEC: srcOffset = -4
    case !dma.isWord && dma.SrcAdj == DMA_ADJ_DEC: srcOffset = -2
    }

    for i := uint32(0); i < count; i++ {

        // not sure about this
        if dma.Dst >= 0x800_0000 && dma.Idx != 3 {
            continue
        }

        if dma.isWord {
            v := mem.Read32(dma.Src)
            mem.Write32(dma.Dst, v)
        } else {
            v := mem.Read16(dma.Src)
            mem.Write16(dma.Dst, uint16(v))
        }

        dma.Dst = uint32(int64(dma.Dst) + dstOffset)
        dma.Src = uint32(int64(dma.Src) + srcOffset)

    }


    dma.disable()
    
    return

    if !dma.Repeat {
        return
    }

    if dma.IRQ {
        dma.Gba.triggerIRQ(0x8 + uint32(dma.Idx))
    }

    //dma.count = ch.wordCount()
    if dma.DstAdj == DMA_ADJ_RES {
        panic("DMA IDK")
        //dma.dst = util.LE32(ch.io[4:])
    }
}

func (dma *DMA) checkMode(mode uint32) bool {
	return mode == dma.Mode && dma.Enabled
}
