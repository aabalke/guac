package ppu

type DisplayFifo struct {
	Pixels []uint8
	i      uint32
}

func (d *DisplayFifo) FifoWrite(data uint16) {

	r := uint8((data) & 0b11111)
	g := uint8((data >> 5) & 0b11111)
	b := uint8((data >> 10) & 0b11111)

	r = (r << 3) | (r >> 2)
	g = (g << 3) | (g >> 2)
	b = (b << 3) | (b >> 2)

	d.Pixels[d.i+0] = r
	d.Pixels[d.i+1] = g
	d.Pixels[d.i+2] = b
	d.Pixels[d.i+3] = 0xFF
}
