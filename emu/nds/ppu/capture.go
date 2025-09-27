package ppu

import (
	"encoding/binary"

	"github.com/aabalke/guac/emu/nds/utils"
)

type Capture struct {
	EVA, EVB    uint8
	WriteBlock  uint8
	WriteOffset uint32
	ReadBlock  *uint32
	ReadOffset  uint8

	Size                uint8
	SrcA3D, SrcBMemFifo bool

	Src     uint8
	Enabled bool

    VramBlocks [4]*[0x2_0000]uint8

    TopA *bool
    Top, Bottom *[]uint8
}

func (c *Capture) Init(vram *VRAM, ppu *PPU, rdBlk *uint32, b, t *[]uint8) {

    c.VramBlocks[0] = &vram.A
    c.VramBlocks[1] = &vram.B
    c.VramBlocks[2] = &vram.C
    c.VramBlocks[3] = &vram.D
    c.TopA = &ppu.TopA
    c.ReadBlock = rdBlk
    c.Bottom = b
    c.Top = t
}

func (c *Capture) Write(addr uint32, v uint8) {

    addr &= 0xFF

	switch addr {
	case 0x64:
		c.EVA = uint8(utils.GetVarData(uint32(v), 0, 4))
	case 0x65:
		c.EVB = uint8(utils.GetVarData(uint32(v), 0, 4))
	case 0x66:
        c.WriteBlock = v & 0b11
        c.WriteOffset = uint32(v >> 2) & 0b11
        c.Size = (v >> 4) & 0b11

	case 0x67:
        c.SrcA3D = utils.BitEnabled(uint32(v), 0)
        c.SrcBMemFifo = utils.BitEnabled(uint32(v), 1)
        c.ReadOffset = (v >> 2) & 0b11

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

        var v uint8
        v |= c.WriteBlock
        v |= uint8(c.WriteOffset << 2)
        v |= (c.Size << 4)

        return v


	case 0x67:

        var v uint8

        if c.SrcA3D {
            v |= 0b1
        }

        if c.SrcBMemFifo {
            v |= 0b10
        }

        v |= (c.ReadOffset << 2)
        v |= (c.Src << 6)

        if c.Enabled {
            v |= 1 << 7
        }

        return v
	}

    panic("UNKNOWN CAPTURE ADDRESS")
}

func (c *Capture) TempLimiter() {

    if c.EVA != 0 && c.EVA != 16 {
        panic("UNSETUP CAPTURE SETTING")
    }

    if c.EVB != 0 {
        // need read block from dispcnt
        panic("UNSETUP CAPTURE SETTING")
    }

    if c.Size != 3 && c.Size != 0 {
        panic("UNSETUP CAPTURE SETTING SIZE")
    }

    //if c.SrcA3D {
    //    panic("UNSETUP CAPTURE SETTING 3d only")
    //}

    if c.SrcBMemFifo {
        panic("UNSETUP CAPTURE SETTING fifo")
    }

    if c.Src >= 2 {
        panic("UNSETUP CAPTURE SETTING src")
    }
}

func (c *Capture) StartCapture() {

    if !c.Enabled {
        return
    }

    screen := c.Bottom

    if *c.TopA {
        screen = c.Top
    }

    j := uint32(0)
    v := uint16(0)
    block := c.VramBlocks[c.WriteBlock]

    for i := 0; i < len(*screen); i += 4 {

        v = Convert24to15(
            (*screen)[i+0],
            (*screen)[i+1],
            (*screen)[i+2],
        )

        v |= 0x8000

        binary.LittleEndian.PutUint16(block[j+c.WriteOffset:], v)

        j += 2
    }
}

func (c *Capture) EndCapture() {
    c.Enabled = false
}

func Convert24to15(r, g, b uint8) uint16 {
    r5 := uint16(r >> 3)
    g5 := uint16(g >> 3)
    b5 := uint16(b >> 3)
    return (r5) | (g5 << 5) | (b5 << 10)
}
