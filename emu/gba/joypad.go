package gba

type Key struct {
	Input uint16
	Cnt   uint16
}

func (k *Key) Read(addr uint32) uint8 {
	switch addr & 3 {
	case 0:
		return uint8(k.Input)
	case 1:
		return uint8(k.Input >> 8)
	case 2:
		return uint8(k.Cnt)
	case 3:
		return uint8(k.Cnt >> 8)
	default:
		return 0
	}
}

func (k *Key) Write(addr uint32, v uint8) {
	switch addr & 3 {
	case 2:
		k.Cnt = (k.Cnt &^ 0xFF) | uint16(v)
	case 3:
		k.Cnt = (k.Cnt & 0xFF) | (uint16(v) << 8)
	}
}

func (k *Key) keyIRQ() bool {
	if disabled := (k.Cnt>>14)&1 == 0; disabled {
		return false
	}

	andFlag := k.Cnt&0x80 != 0

	if or := !andFlag && ^k.Cnt&k.Input != 0; or {
		return true
	}

	if and := andFlag && ^k.Cnt&^k.Input == 0; and {
		return true
	}

	return false
}
