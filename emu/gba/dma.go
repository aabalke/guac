package gba

import (
	"github.com/aabalke33/guac/emu/gba/utils"
    "fmt"
)

var _ = fmt.Sprintf("")
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
    return uint8(dma.Src >> bitOffset) & 0xFF
}

func (dma *DMA) ReadDst(byte uint32) uint8 {
    bitOffset := 8 * byte
    return uint8(dma.Dst >> bitOffset) & 0xFF
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
    if dma.Idx == 0 {
        dma.Src &= 0x7FF_FFFF
        return
    }
    dma.Src &= 0xFFF_FFFF
}

func (dma *DMA) WriteDst(v uint8, byte uint32) {
    dma.Dst = utils.ReplaceByte(dma.Dst, uint32(v), byte)
    if dma.Idx == 0 {
        dma.Dst &= 0x7FF_FFFF
        return
    }
    dma.Dst &= 0xFFF_FFFF
}

func (dma *DMA) WriteCount(v uint8, hi bool) {

    if hi {
        dma.WordCount = (dma.WordCount & 0b1111_1111) |  (uint32(v) << 8)
        return
    }

    mask := 0x7F
    if dma.Idx == 3 {
        mask = 0x1FF
    }

	dma.WordCount = (dma.WordCount &^ 0b1111_1111) |  (uint32(v) & uint32(mask))
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
		dma.SrcAdj = (dma.SrcAdj &^ 0b1) | ((uint32(v) >> 7) & 1)
		//dma.SrcAdj = (dma.SrcAdj &^ 0b1) | ((uint32(v) >> 6) & 1)
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

    //fmt.Printf("DMA TRANSFER CURR %08d SRC %08X, DST %08X, WORD COUNT %08X 0x202EEC8 %08X TYPE %02b\n", CURR_INST, dma.Dst, dma.Src, dma.WordCount, dma.Gba.Mem.Read32(0x202EEC8), dma.Mode)

    mem := dma.Gba.Mem

    count := dma.WordCount

    switch {
    case count == 0 && dma.Idx == 3:
        count  = 0x10000
    case count == 0:
        count  = 0x4000
    }

    if dma.Idx == 0 && dma.Mode == 3 {
        return
    }

    if dma.Mode == DMA_MODE_HBL {
        srcInLimited := dma.Src >= 0x600_0000 && dma.Src < 0xA00_0000
        dstInLimited := dma.Dst >= 0x600_0000 && dma.Dst < 0xA00_0000
        allowed := utils.BitEnabled(dma.Gba.Mem.Read16(0x400_0000), 5)

        if (srcInLimited || dstInLimited) && !allowed {
            return
        }
    }

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
        dstRom := tmpDst >= 0x800_0000 && tmpDst < 0xE00_0000
        //srcRom := tmpSrc >= 0x800_0000 && tmpSrc < 0xE00_0000
        //dstSram := tmpDst >= 0xE00_0000 && tmpDst < 0x1000_0000
        //srcSram := tmpSrc >= 0xE00_0000 && tmpSrc < 0x1000_0000

        if (dstRom) && dma.Idx != 3 {
            continue
        }

        //if dstSram || srcSram {
        //    continue
        //}

        if dma.isWord {
            v := mem.Read32(tmpSrc)
            mem.Write32(tmpDst, v)
        } else {
            v := mem.Read16(tmpSrc)
            mem.Write16(tmpDst, uint16(v))

        }
        tmpDst = uint32(int64(tmpDst) + dstOffset)
        tmpSrc = uint32(int64(tmpSrc) + srcOffset)

    }

    dma.Src = tmpSrc
    dma.Dst = tmpDst

    if dma.IRQ {
        dma.Gba.triggerIRQ(0x8 + uint32(dma.Idx))
    }

    if !dma.Repeat {
        dma.disable()
        return
    }

    panic("REPEAT DMA")
    //dma.Dst = tmpDst
    //dma.Src = tmpSrc
    
    return

    //dma.count = ch.wordCount()
    if dma.DstAdj == DMA_ADJ_RES {
        panic("DMA IDK")
        //dma.dst = util.LE32(ch.io[4:])
    }
}

func (dma *DMA) checkMode(mode uint32) bool {
	return mode == dma.Mode && dma.Enabled
}
