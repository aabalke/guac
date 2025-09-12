package ppu

type PRAM [0x400]uint16

func (p *PRAM) Read(addr uint32, ppu *PPU) uint8 {

	hi := addr&1 == 1
	addr &= 0x7FF

    switch {
    case addr < 0x400 && !ppu.EngineA2D:
        return 0
    case addr >= 0x400 && !ppu.EngineB2D:
        return 0
    }

	addr >>= 1

	if hi {
		return uint8(p[addr] >> 8)
	}

	return uint8(p[addr])
}

func (p *PRAM) Write(addr uint32, v uint8, ppu *PPU) {

	hi := addr&1 == 1
	addr &= 0x7FF

    switch {
    case addr < 0x400 && !ppu.EngineA2D:
        return
    case addr >= 0x400 && !ppu.EngineB2D:
        return
    }

	addr >>= 1

	if hi {
		p[addr] &= 0xFF
		p[addr] |= uint16(v) << 8
		return
	}

	p[addr] &^= 0xFF
	p[addr] |= uint16(v)
}
