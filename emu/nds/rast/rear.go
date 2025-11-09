package rast

import (
	"github.com/aabalke/guac/emu/nds/rast/gl"
	"github.com/aabalke/guac/emu/nds/utils"
)

type RearPlane struct {
	paramActive      [8]bool
	Enabled          bool
	ClearColor       gl.Color
	FogEnabled       bool
	Id               uint32
	ClearDepth       uint32
	OffsetX, OffsetY uint32
    VRAM VRAM

    Color [256 * 256]gl.Color
    Depth [256 * 256]float64
    Fog   [256 * 256]bool
}

func (r *RearPlane) Write(addr uint32, v uint8) {

	// disable rearplane if all zero
	r.paramActive[addr-0x350] = v != 0
	r.Enabled = false
	for _, v := range r.paramActive {
		if v {
			r.Enabled = true
		}
	}

	switch addr {
	case 0x350:

		r.ClearColor = Convert15BitByte(r.ClearColor, v, false)

	case 0x351:

		r.ClearColor = Convert15BitByte(r.ClearColor, v, true)
		r.FogEnabled = utils.BitEnabled(uint32(v), 7)

	case 0x352:

		r.ClearColor.A = float64(v&0x1F) / 0x1F

	case 0x353:
		r.Id = uint32(v & 0b11_1111)

	case 0x354:
		r.ClearDepth &^= 0xFF
		r.ClearDepth |= uint32(v)

	case 0x355:
		r.ClearDepth &^= 0xFF << 8
		r.ClearDepth |= uint32(v&^0x80) << 8

	case 0x356:
		r.OffsetX = uint32(v)

	case 0x357:
		r.OffsetY = uint32(v)
	}
}

func (r *RearPlane) Cache() {

    vram := r.VRAM

    const (
        SLOT2_BASE = 0x40000
        SLOT3_BASE = 0x60000
        WIDTH  = 256
        HEIGHT = 256
    )

    for y := range HEIGHT {
        for x := range WIDTH {

            i := x + y * WIDTH

            addr := uint32(i) * 2
            colorData := uint16(vram.ReadTexture(SLOT2_BASE + addr))
            colorData |= uint16(vram.ReadTexture(SLOT2_BASE + addr + 1)) << 8

            c := gl.MakeColorFrom15Bit(
                uint8(colorData) & 0x1F,
                uint8(colorData >> 5) & 0x1F,
                uint8(colorData >> 10) & 0x1F,
            )

            if transparent := colorData & 0x8000 == 0; transparent {
                c.A = 0
            }

            depthData := uint16(vram.ReadTexture(SLOT3_BASE + addr))
            depthData |= uint16(vram.ReadTexture(SLOT3_BASE + addr + 1)) << 8

            depth := uint32(depthData &^ 0x8000)
            //depth = (depth * 0x200) + ((depth + 1)/ 0x8000) * 0x1FF //gbatek wrong
            depth = (depth * 0x200) + 0x1FF // desmume / ndsemu say this correct

            fog := depthData & 0x8000 != 0

            r.Color[i] = c
            r.Depth[i] = float64(depth)
            r.Fog[i] = fog
        }
    }
}
