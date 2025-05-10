package gba

func (gba *GBA) getJoypad(secondByte bool) uint8 {

	if secondByte {
        return uint8(gba.Joypad >> 8)
	}

    return uint8(gba.Joypad)
}
