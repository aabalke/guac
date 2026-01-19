package ppu

type PRAM struct {
    Bg, Obj [0x100]uint16
}

const (
    PRAM_A_BG = iota
    PRAM_A_OBJ
    PRAM_B_BG
    PRAM_B_OBJ
)

func (p *PPU) ReadPram(addr uint32, ppu *PPU) uint8 {

	hi := addr&1 != 0
	addr &= 0x7FF

    bankIdx := addr >> 9

    var bank *[0x100]uint16

    switch bankIdx {
    case PRAM_A_BG:
        bank = &p.EngineA.Pram.Bg
        if !ppu.EngineA2D {
            return 0
        }
    case PRAM_A_OBJ:
        bank = &p.EngineA.Pram.Obj
        if !ppu.EngineA2D {
            return 0
        }
    case PRAM_B_BG:
        bank = &p.EngineB.Pram.Bg
        if !ppu.EngineB2D {
            return 0
        }
    case PRAM_B_OBJ:
        bank = &p.EngineB.Pram.Obj
        if !ppu.EngineB2D {
            return 0
        }
    }

    addr &= 0x1FF
	addr >>= 1

	if hi {
		return uint8(bank[addr] >> 8)
	}

	return uint8(bank[addr])
}

func (p *PPU) WritePram(addr uint32, v uint8, ppu *PPU) {

	hi := addr&1 != 0
	addr &= 0x7FF

    bankIdx := addr >> 9

    var bank *[0x100]uint16

    switch bankIdx {
    case PRAM_A_BG:
        bank = &p.EngineA.Pram.Bg
        if !ppu.EngineA2D {
            return
        }
    case PRAM_A_OBJ:
        bank = &p.EngineA.Pram.Obj
        if !ppu.EngineA2D {
            return
        }
    case PRAM_B_BG:
        bank = &p.EngineB.Pram.Bg
        if !ppu.EngineB2D {
            return
        }
    case PRAM_B_OBJ:
        bank = &p.EngineB.Pram.Obj
        if !ppu.EngineB2D {
            return
        }
    }

    addr &= 0x1FF
	addr >>= 1

	if hi {
		bank[addr] &= 0xFF
		bank[addr] |= uint16(v) << 8
        return
	}

    bank[addr] &^= 0xFF
    bank[addr] |= uint16(v)
}
