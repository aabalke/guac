package rast

import "github.com/aabalke/guac/emu/nds/rast/gl"

func WriteToonTbl(t *[32]gl.Color, addr uint32, v uint8) {
	addr -= 0x380
	idx := addr / 2
	hi := addr%2 == 1
	t[idx] = Convert15BitByte(t[idx], v, hi)
}

func Convert15BitByte(c gl.Color, v uint8, hi bool) gl.Color {
	//1111_1111|1111_1111
	//abbb bbgg|gggr rrrr

	if hi {
		r := uint8(c.R * 0x1F)

		g := uint8(c.G * 0x1F)
		g &= 0b111
		g |= (uint8(v) & 0b11) << 3

		b := uint8(v>>2) & 0x1F

		return gl.MakeColorFrom15Bit(r, g, b)
	}
	r := v & 0x1F

	g := uint8(c.G * 0x1F)
	g &^= 0b111
	g |= uint8(v >> 5) & 0x1F

	b := uint8(c.B * 0x1F)

	return gl.MakeColorFrom15Bit(r, g, b)
}
