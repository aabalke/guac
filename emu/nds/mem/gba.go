package mem

func (m *Mem) ReadGbaSlot(addr uint32, arm9 bool) uint8 {

	if arm9 && m.Gamecard.ExMem.isGBAAccessArm7 {
		return 0
	}

	if !arm9 && !m.Gamecard.ExMem.isGBAAccessArm7 {
		return 0
	}

	if sram := addr >= 0xA00_0000; sram {
		return 0xFF
	}

	return uint8(((addr >> 1) & 0xFFFF) >> ((addr & 1) * 8))
}
