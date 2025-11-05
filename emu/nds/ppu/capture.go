package ppu

import (
	"encoding/binary"

	"github.com/aabalke/guac/emu/nds/rast"
	"github.com/aabalke/guac/emu/nds/utils"
)

type Capture struct {
	EVA, EVB    uint8

	WriteBlock  uint8
    WriteOffset uint32
	WriteOffsetV uint8
	ReadBlock  *uint32
    ReadOffset uint32
	ReadOffsetV  uint8

	Size                uint8
	SrcA3D, SrcBMemFifo bool

	Src     uint8
	Enabled bool
    ActiveCapture bool

    VramBlocks [4]*[0x2_0000]uint8

    Pixels *[]uint8
    Pixels3d *rast.Pixels

    arm7CBlock, arm7DBlock *bool

    ActiveData ActiveData
}

type ActiveData struct {
	WriteBlock  uint8
    WriteOffset uint32
	WriteOffsetV uint8
	ReadBlock  *uint32
    ReadOffset uint32
	ReadOffsetV  uint8
}

func (c *Capture) Init(vram *VRAM, ppu *PPU, rdBlk *uint32, pixels *[]uint8, pixels3d *rast.Pixels) {
    c.VramBlocks[0] = &vram.A
    c.VramBlocks[1] = &vram.B
    c.VramBlocks[2] = &vram.C
    c.VramBlocks[3] = &vram.D
    c.ReadBlock = rdBlk
    c.Pixels = pixels
    c.Pixels3d = pixels3d
    c.arm7CBlock = &vram.isCArm7
    c.arm7DBlock = &vram.isDArm7
}

func (c *Capture) Write(addr uint32, v uint8) {

    addr &= 0xFF

	switch addr {
	case 0x64:
		c.EVA = min(uint8(utils.GetVarData(uint32(v), 0, 4)), 16)
	case 0x65:
		c.EVB = min(uint8(utils.GetVarData(uint32(v), 0, 4)), 16)
	case 0x66:
        c.WriteBlock = v & 0b11
        c.WriteOffsetV = (v >> 2) & 0b11
        c.WriteOffset = (uint32(v >> 2) & 0b11) * 0x8000
        c.Size = (v >> 4) & 0b11
	case 0x67:
        c.SrcA3D = utils.BitEnabled(uint32(v), 0)
        c.SrcBMemFifo = utils.BitEnabled(uint32(v), 1)
        c.ReadOffsetV = (v >> 2) & 0b11
        c.ReadOffset = (uint32(v >> 2) & 0b11) * 0x8000

        c.Src = (v >> 5) & 0b11
        c.Enabled = utils.BitEnabled(uint32(v), 7)
	}

    c.TempLimiter()
}

func (c *Capture) Read(addr uint32) uint8 {


    addr &= 0xFF

	switch addr {
	case 0x64:
		return c.EVA
	case 0x65:
		return c.EVB
	case 0x66:

        v := c.WriteBlock
        v |= c.WriteOffsetV << 2
        v |= c.Size << 4

        return v

	case 0x67:

        var v uint8

        if c.SrcA3D {
            v |= 0b1
        }

        if c.SrcBMemFifo {
            v |= 0b10
        }

        v |= c.ReadOffsetV << 2
        v |= c.Src << 6

        if c.Enabled {
            v |= 1 << 7
        }

        return v
	}

    panic("UNKNOWN CAPTURE ADDRESS")
}

func (c *Capture) TempLimiter() {

    if c.EVA != 0 && c.EVA != 16 {
        panic("UNSETUP CAPTURE SETTING BLEND A")
    }

    if c.EVB != 0 {
        // need read block from dispcnt
        panic("UNSETUP CAPTURE SETTING BLEND B")
    }

    if c.Size != 3 && c.Size != 0 {
        panic("UNSETUP CAPTURE SETTING SIZE")
    }

    if c.SrcBMemFifo {
        panic("UNSETUP CAPTURE SETTING fifo")
    }

    if c.Src >= 2 {
        panic("UNSETUP CAPTURE SETTING src")
    }
}

func (c *Capture) StartCapture() {

    //if !c.Enabled {
    //    return
    //}

    c.ActiveData = ActiveData{
        WriteBlock: c.WriteBlock,
        WriteOffset: c.WriteOffset,
	    ReadBlock: c.ReadBlock,
        ReadOffset: c.ReadOffset,
    }

    if c.ActiveCapture {
        panic("ACTIVE CAPTURE ON START CAPTURE")
    }

    //uhh.V = uint32(c.WriteBlock)

    c.ActiveCapture = true
}

func (c *Capture) CaptureLine(y uint32, isRenderingB bool) {

    //return

    //if !c.ActiveCapture {
    //    return
    //}

    block := c.VramBlocks[c.ActiveData.WriteBlock]

    // these smooth over timings with arm7 usage after capture
    // ex. mario kart transitions
    if blockCArm7 := c.ActiveData.WriteBlock == 2 && *c.arm7CBlock; blockCArm7 {
        return
    }

    if blockDArm7 := c.ActiveData.WriteBlock == 3 && *c.arm7DBlock; blockDArm7 {
        return
    }

    for x := range uint32(SCREEN_WIDTH) {
        j := (x + (y * SCREEN_WIDTH)) * 2

        if c.SrcA3D {
            i := (x + (y * SCREEN_WIDTH))
            v, alpha := uint32(0), float32(0)
            pixels := c.Pixels3d
            if isRenderingB {
                v, alpha = pixels.PalettesA[i], pixels.AlphaA[i]
            } else {
                v, alpha = pixels.PalettesB[i], pixels.AlphaB[i]
            }

            if alpha > 0 {
                v |= 0x8000
            }

            binary.LittleEndian.PutUint16(block[j+c.ActiveData.WriteOffset:], uint16(v))

            continue
        }

        i := (x + (y * SCREEN_WIDTH)) * 4

        v := Convert24to15(
            (*c.Pixels)[i+0],
            (*c.Pixels)[i+1],
            (*c.Pixels)[i+2],
        )

        v |= 0x8000

        binary.LittleEndian.PutUint16(block[j+c.ActiveData.WriteOffset:], v)
    }
}

func (c *Capture) EndCapture() {
    c.ActiveCapture = false
    c.Enabled = false
}

func Convert24to15(r, g, b uint8) uint16 {
    r5 := uint16(r >> 3)
    g5 := uint16(g >> 3)
    b5 := uint16(b >> 3)
    return (r5) | (g5 << 5) | (b5 << 10)
}
