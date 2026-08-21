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

func (p *PPU) ReadPram(addr uint32) uint8 {
	addr &= 0x7FF

	var bank *[0x100]uint16
	switch bankIdx := addr >> 9; bankIdx {
	case PRAM_A_BG:
		bank = &p.EngineA.Pram.Bg
		if !p.EngineA2D {
			return 0
		}
	case PRAM_A_OBJ:
		bank = &p.EngineA.Pram.Obj
		if !p.EngineA2D {
			return 0
		}
	case PRAM_B_BG:
		bank = &p.EngineB.Pram.Bg
		if !p.EngineB2D {
			return 0
		}
	case PRAM_B_OBJ:
		bank = &p.EngineB.Pram.Obj
		if !p.EngineB2D {
			return 0
		}
	}

	hi := addr&1 != 0

	addr &= 0x1FF
	addr >>= 1

	if hi {
		return uint8(bank[addr] >> 8)
	}

	return uint8(bank[addr])
}

func (p *PPU) WritePram(addr uint32, v uint8) {
	addr &= 0x7FF

	var bank *[0x100]uint16
	switch bankIdx := addr >> 9; bankIdx {
	case PRAM_A_BG:
		bank = &p.EngineA.Pram.Bg
		if !p.EngineA2D {
			return
		}
	case PRAM_A_OBJ:
		bank = &p.EngineA.Pram.Obj
		if !p.EngineA2D {
			return
		}
	case PRAM_B_BG:
		bank = &p.EngineB.Pram.Bg
		if !p.EngineB2D {
			return
		}
	case PRAM_B_OBJ:
		bank = &p.EngineB.Pram.Obj
		if !p.EngineB2D {
			return
		}
	}

	hi := addr&1 != 0
	addr &= 0x1FF
	addr >>= 1

	if hi {
		bank[addr] = (bank[addr] & 0x00FF) | (uint16(v) << 8)
		return
	}

	bank[addr] = (bank[addr] & 0xFF00) | (uint16(v) << 0)
}
