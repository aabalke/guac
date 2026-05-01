package gb

type ColorPalette struct {
	Palette  [0x40]uint8
	Unpacked [0x40 / 2]uint32
	Idx      uint8
	Inc      bool
}

func (p *ColorPalette) Init() {
	for i := range len(p.Palette) {
		p.Palette[i] = 0xFF
	}
}

func (p *ColorPalette) update(idx uint8) {

	idx &^= 1

	color := uint16(p.Palette[idx]) | uint16(p.Palette[idx+1])<<8

	p.Unpacked[idx/2] = (uint32(colArr[color&0x1F]) |
		uint32(colArr[(color>>5)&0x1F])<<8 |
		uint32(colArr[(color>>10)&0x1F]<<16))
}

// Mapping of the 5 bit colour value to a 8 bit value.
var colArr = [...]uint32{
	0x0,
	0x8,
	0x10,
	0x18,
	0x20,
	0x29,
	0x31,
	0x39,
	0x41,
	0x4a,
	0x52,
	0x5a,
	0x62,
	0x6a,
	0x73,
	0x7b,
	0x83,
	0x8b,
	0x94,
	0x9c,
	0xa4,
	0xac,
	0xb4,
	0xbd,
	0xc5,
	0xcd,
	0xd5,
	0xde,
	0xe6,
	0xee,
	0xf6,
	0xff,
}
