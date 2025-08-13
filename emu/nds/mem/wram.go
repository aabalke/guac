package mem

type WRAM struct {
	Wram [0x8000]uint8
	CNT  uint8
}

func (w *WRAM) WriteCNT(v uint8) {
	w.CNT = v & 0b11
}

func (w *WRAM) ReadCNT() uint8 {
	return w.CNT
}

func (w *WRAM) Write(addr uint32, v uint8, arm9 bool) {

	addr -= 0x300_0000

	if arm9 {

		switch w.CNT {
		case 0:
			w.Wram[addr%0x8000] = v
		case 1:
			w.Wram[0x4000+(addr%0x4000)] = v
		case 2:
			w.Wram[addr%0x4000] = v
			//case 3: return 0 // should this clear ram?
		}

		return
	}

	switch w.CNT {
	//case 0: return 0 // should return 0x380_0000 +
	case 1:
		w.Wram[addr%0x4000] = v
	case 2:
		w.Wram[0x4000+(addr%0x4000)] = v
	case 3:
		w.Wram[addr%0x8000] = v
	}

	return
}

func (w *WRAM) Read(addr uint32, arm9 bool) uint8 {

	if arm9 {

		switch w.CNT {
		case 0:
			return w.Wram[addr%0x8000]
		case 1:
			return w.Wram[0x4000+(addr%0x4000)]
		case 2:
			return w.Wram[addr%0x4000]
		case 3:
			return 0 // should this clear ram?
		}

		return 0
	}

	switch w.CNT {
	case 0:
		return 0 // should return 0x380_0000 +
	case 1:
		return w.Wram[addr%0x4000]
	case 2:
		return w.Wram[0x4000+(addr%0x4000)]
	case 3:
		return w.Wram[addr%0x8000]
	}

	return 0
}
