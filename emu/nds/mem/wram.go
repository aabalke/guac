package mem

import "unsafe"

type WRAM struct {
	Wram [0x8000]uint8
	CNT  uint8

	WRAM7 [0x1_0000]uint8
}

func (w *WRAM) WriteCNT(v uint8) {
	w.CNT = v & 0b11
}

func (w *WRAM) ReadCNT() uint8 {
	return w.CNT
}

func (w *WRAM) Write(addr uint32, v uint8, arm9 bool) {

	if arm9 {

		switch w.CNT {
		case 0:
			w.Wram[addr&0x7FFF] = v
		case 1:
			w.Wram[0x4000+(addr&0x3FFF)] = v
		case 2:
			w.Wram[addr&0x3FFF] = v
		}

		return
	}

	if addr >= 0x380_0000 {
		w.WRAM7[(addr & 0xFFFF)] = v
		return
	}

	switch w.CNT {
	case 0:
		w.WRAM7[(addr & 0xFFFF)] = v
	case 1:
		w.Wram[addr&0x3FFF] = v
	case 2:
		w.Wram[0x4000+(addr&0x3FFF)] = v
	case 3:
		w.Wram[addr&0x7FFF] = v
	}
}

func (w *WRAM) Read(addr uint32, arm9 bool) uint8 {

	if arm9 {

		switch w.CNT {
		case 0:
            return w.Wram[addr&0x7FFF]
		case 1:
			return w.Wram[0x4000+(addr&0x3FFF)]
		case 2:
			return w.Wram[addr&0x3FFF]
		case 3:
			return 0 // should this clear ram?
		}

		return 0
	}

	if addr >= 0x380_0000 {
		return w.WRAM7[(addr & 0xFFFF)]
	}

	switch w.CNT {
	case 0:
		return w.WRAM7[(addr & 0xFFFF)]
	case 1:
		return w.Wram[addr&0x3FFF]
	case 2:
		return w.Wram[0x4000+(addr&0x3FFF)]
	case 3:
		return w.Wram[addr&0x7FFF]
	}

	return 0
}

func (w *WRAM) ReadPtr(addr uint32, arm9 bool) (unsafe.Pointer, bool) {


	if arm9 {
		switch w.CNT {
		case 0:
            // this fails tcm test rockwrestler
            return unsafe.Add(unsafe.Pointer(&w.Wram), addr & 0x7FFF), true
		case 1:
            return unsafe.Add(unsafe.Pointer(&w.Wram), 0x4000 + (addr & 0x3FFF)), true
		case 2:
            return unsafe.Add(unsafe.Pointer(&w.Wram), addr & 0x3FFF), true
		case 3:
            return nil, false
		}

        return nil, false
	}

	if addr >= 0x380_0000 {
        return unsafe.Add(unsafe.Pointer(&w.WRAM7), addr & 0xFFFF), true
	}

	switch w.CNT {
	case 0:
        return unsafe.Add(unsafe.Pointer(&w.WRAM7), addr & 0xFFFF), true
	case 1:
        return unsafe.Add(unsafe.Pointer(&w.Wram), addr & 0x3FFF), true
	case 2:
        return unsafe.Add(unsafe.Pointer(&w.Wram), 0x4000 + (addr &0x3FFF)), true
	case 3:
        return unsafe.Add(unsafe.Pointer(&w.Wram), addr & 0x7FFF), true
	}

	return nil, false
}
