package gameboy

func (gb *GameBoy) getJoypad() uint8 {

	joyp := gb.MemoryBus.Memory[0xFF00]

	// 1 is off, and 0 is on, need to negate
	joyp = joyp ^ 0xFF

	if dpad := !(joyp&0b10000 == 0b10000); dpad {
		top := gb.Joypad >> 4
		top |= 0xF0
		return joyp & top
    }

	if ssba := !(joyp&0b100000 == 0b100000); ssba {
		bottom := gb.Joypad & 0xF
		bottom |= 0xF0
		return joyp & bottom
	}

	return joyp
}
