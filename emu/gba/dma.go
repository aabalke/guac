package gba

import (
	"github.com/aabalke33/guac/emu/gba/utils"
	"github.com/aabalke33/guac/emu/gba/cart"
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
	SingleByteSet bool
	WordCount     uint32

	DstAdj  uint32
	SrcAdj  uint32
	Repeat  bool
	isWord  bool
    DRQ     bool
	Mode    uint32
	IRQ     bool
	Enabled bool

    EepromWidth uint32
}

func (dma *DMA) ReadControl(hi bool) uint8 {
    if hi { return uint8(dma.Control >> 8) & 0xFF }
    return uint8(dma.Control) & 0xFF
}

func (dma *DMA) MaskAddr(v uint32, src bool) uint32 {

    if dma.Idx == 0 || !src && dma.Idx == 1 || !src && dma.Idx == 2 {
        return v & 0x7FF_FFFF
    }

    return v & 0xFFF_FFFF
}

func (dma *DMA) WriteSrc(v uint8, byte uint32) {

    if byte == 3 {
    switch dma.Idx {
    case 0: v &= 0x07
    case 1, 2, 3: v &= 0x0F
    }
    }

    dma.Src = utils.ReplaceByte(dma.Src, uint32(v), byte)
}

func (dma *DMA) WriteDst(v uint8, byte uint32) {

    if byte == 3 {
    switch dma.Idx {
    case 0, 1, 2: v &= 0x07
    case 3: v &= 0x0F
    }
    }

    dma.Dst = utils.ReplaceByte(dma.Dst, uint32(v), byte)
}

func (dma *DMA) WriteCount(v uint8, hi bool) {

    if hi {
        mask := uint32(0x3F)
        if dma.Idx == 3 {
            mask = 0xFF
        }
        dma.WordCount = (dma.WordCount & 0xFF) | ((uint32(v) & mask) << 8)
        return
    }

	dma.WordCount = (dma.WordCount &^ 0xFF) |  uint32(v)
}

func (dma *DMA) WriteControl(v uint8, hi bool) {

	if hi {
        a := uint32(v) & 0xF7
        if dma.Idx == 3 {
            a = uint32(v) & 0xFF
        }

        wasDisabled := !dma.Enabled
		dma.Control = (dma.Control & 0b1111_1111) | (a << 8)
		dma.SrcAdj = (dma.SrcAdj & 1) | (a & 1) << 1
		dma.Repeat = utils.BitEnabled(a, 1)
		dma.isWord = utils.BitEnabled(a, 2)
        dma.DRQ = utils.BitEnabled(a, 3)
        dma.Mode = utils.GetVarData(a, 4, 5)
		dma.IRQ = utils.BitEnabled(a, 6)
		dma.Enabled = utils.BitEnabled(a, 7)

        // immediate should be 2 cycles after enabling
        if isImmediate := wasDisabled && dma.checkMode(DMA_MODE_IMM); isImmediate {
            dma.transfer()
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

func (dma *DMA) transfer() {
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
        srcInLimited := dma.Src >= 0x600_0000 && dma.Src < 0x800_0000
        dstInLimited := dma.Dst >= 0x600_0000 && dma.Dst < 0x800_0000
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
    case dma.isWord && dma.DstAdj == DMA_ADJ_RES: dstOffset = 4
    case dma.isWord && dma.DstAdj == DMA_ADJ_RES: dstOffset = 2
    }

    switch {
    case dma.isWord && dma.SrcAdj  == DMA_ADJ_INC: srcOffset = 4
    case !dma.isWord && dma.SrcAdj == DMA_ADJ_INC: srcOffset = 2
    case dma.isWord && dma.SrcAdj  == DMA_ADJ_DEC: srcOffset = -4
    case !dma.isWord && dma.SrcAdj == DMA_ADJ_DEC: srcOffset = -2
    case dma.SrcAdj == DMA_ADJ_RES: panic("DMA SRC SET TO PROHIBITTED")

    }

    if fifo := (dma.Idx == 1 || dma.Idx == 2) && dma.Mode == DMA_MODE_REF; fifo {

        //if dma.isWord {
        //    count = 1
        //} else {
        //    count = 2
        //}

        dma.isWord = true
        dma.DstAdj = DMA_ADJ_NON
        count = 4

        dstOffset = 0
        //srcOffset = 4

        if !dma.Repeat || (dma.Dst != 0x400_00A0 && dma.Dst != 0x400_00A4) {
            panic("INVALID FIFO DMA")
        }
    }


    dstRom := tmpDst >= 0x800_0000 && tmpDst < 0xE00_0000
    //srcRom := tmpSrc >= 0x800_0000 && tmpSrc < 0xE00_0000

    dstSram := tmpDst >= 0xE00_0000 && tmpDst < 0x1000_0000
    srcSram := tmpSrc >= 0xE00_0000 && tmpSrc < 0x1000_0000

    if sram := (dstSram || srcSram) && dma.Gba.Cartridge.FlashType == 2; sram {
        return
    } else if rom := (dstRom || dstSram) && dma.Idx != 3; rom {
        return
    }

    if (dstSram || srcSram) && dma.Gba.Cartridge.FlashType == 2 {
        return
    }

    for i := uint32(0); i < count; i++ {

        if eeprom := CheckEeprom(dma.Gba, tmpDst); eeprom {
            dstRom := tmpDst >= 0x800_0000 && tmpDst < 0xE00_0000
            srcRom := tmpSrc >= 0x800_0000 && tmpSrc < 0xE00_0000

            if count == 9 || count == 73 {
                cart.EepromWidth = 6
            } else if count == 17 || count == 81 {
                cart.EepromWidth = 14
            }

            if srcRom && dstRom {
                panic("EEPROM HAS BOTH SRC AND DST ROM ADDR")
            }

            // do not continue this., do not put this outside loop
        }

        if dma.isWord {

            if badAddr := tmpSrc < 0x200_0000; badAddr {
                mem.Write32(tmpDst, 0)
            } else {
                v := mem.Read32(tmpSrc)
                mem.Write32(tmpDst, v)
            }

        } else {

            if badAddr := tmpSrc < 0x200_0000; badAddr {
                mem.Write16(tmpDst, 0)
            } else {
                v := mem.Read16(tmpSrc)
                mem.Write16(tmpDst, uint16(v))
            }
        }

        tmpDst = uint32(int64(tmpDst) + dstOffset)
        tmpSrc = uint32(int64(tmpSrc) + srcOffset)
    }

    if dma.IRQ {
        dma.Gba.InterruptStack.setIRQ(8 + uint32(dma.Idx))
    }

    if !dma.Repeat {
        // DO NOT WRITEBACK DST AND SRC UNLESS REPEAT
        dma.disable()
        return
    }

    if dma.DstAdj == DMA_ADJ_RES {
        // dma.Dst stays the same
        dma.Src = dma.MaskAddr(tmpSrc, true)
        return
    }

    dma.Src = dma.MaskAddr(tmpSrc, true)
    dma.Dst = dma.MaskAddr(tmpDst, false)

    return
}

func (dma *DMA) transferVideo(vcount uint32) {

    if dma.Idx != 3 || dma.Mode != 3 || !dma.Enabled || (dma.Dst >= 0x600_0000 || dma.Dst <= 0x700_0000) {
        return
    }

    if vcount >= 2 && vcount <= 162 {
        dma.transfer()
    }
}

func (dma *DMA) transferFifo() {

    if (dma.Idx != 1 && dma.Idx != 2) || dma.Mode != DMA_MODE_REF || !dma.Enabled {
        return
    }

    if !dma.Repeat {
        return
    }

    if !dma.Gba.Apu.ChannelA.Refill && dma.Dst != 0x400_00A0 {
        return
    }

    if !dma.Gba.Apu.ChannelB.Refill && dma.Dst != 0x400_00A4 {
        return
    }

    dma.transfer()
}

func (dma *DMA) checkMode(mode uint32) bool {
	return mode == dma.Mode && dma.Enabled
}
