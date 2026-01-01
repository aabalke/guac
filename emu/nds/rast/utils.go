package rast

import "github.com/aabalke/guac/emu/nds/rast/gl"

func RGB24ToRGB15(r, g, b uint8) uint16 {
	r5 := uint16(r >> 3)
	g5 := uint16(g >> 3)
	b5 := uint16(b >> 3)
	return (b5 << 10) | (g5 << 5) | r5
}

func RGB15ToRGB24(r, g, b uint8) (uint8, uint8, uint8) {
	r = (r << 3) | (r >> 2)
	g = (g << 3) | (g >> 2)
	b = (b << 3) | (b >> 2)
	return r, g, b
}

func Convert15BitByte(c gl.Color, v uint8, hi bool) gl.Color {
	//1111_1111|1111_1111
	//abbb bbgg|gggr rrrr

	a := c.A

	if hi {
		r := uint8(c.R * 0x1F)

		g := uint8(c.G * 0x1F)
		g &= 0b111
		g |= (uint8(v) & 0b11) << 3

		b := uint8(v>>2) & 0x1F

		c = gl.MakeColorFrom15Bit(r, g, b)
		c.A = a
		return c
	}
	r := v & 0x1F

	g := uint8(c.G * 0x1F)
	g &^= 0b111
	g |= uint8(v>>5) & 0x1F

	b := uint8(c.B * 0x1F)

	c = gl.MakeColorFrom15Bit(r, g, b)
	c.A = a
	return c
}
