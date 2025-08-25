package ppu

type PRAM [0x400]uint16

func (p *PRAM) Read(addr uint32) uint8 {

	hi := addr&1 == 1
	addr &= 0x7FF
	addr >>= 1

	if hi {
		return uint8(p[addr] >> 8)
	}

	return uint8(p[addr])
}

func (p *PRAM) Write(addr uint32, v uint8) {


	hi := addr&1 == 1
	addr &= 0x7FF
	addr >>= 1

	if hi {
		p[addr] &= 0xFF
		p[addr] |= uint16(v) << 8
		return
	}

	p[addr] &^= 0xFF
	p[addr] |= uint16(v)
}
