package gba

type Key struct {
	Irq     *Irq
	Input   uint16
	Cnt     uint16
	Enabled bool
	AndMode bool
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
		k.Cnt = (k.Cnt & 0xFF) | (uint16(v&3) << 8)

		k.Enabled = v&0x40 != 0
		k.AndMode = v&0x80 != 0
	}

	k.keyIRQ()
}

func (k *Key) keyIRQ() {
	if !k.Enabled {
		return
	}

	if k.AndMode {
		if (^k.Cnt & 0x3FF) == k.Input {
			k.Irq.SetIRQ(12)
		}

		return
	}

	if (^k.Cnt&k.Input)&0x3FF != 0 {
		k.Irq.SetIRQ(12)
	}
}
