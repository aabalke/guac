package mem

import "unsafe"

type WRAM struct {
	wram  [0x8000]uint8
	wram7 [0x10000]uint8
	cnt   uint8
}

func (w *WRAM) WriteCNT(v uint8) {
	w.cnt = v & 3
}

func (w *WRAM) ReadCNT() uint8 {
	return w.cnt
}

func (w *WRAM) Write9(addr uint32, v uint8) {
	switch w.cnt {
	case 0:
		w.wram[addr&0x7FFF] = v
	case 1:
		w.wram[0x4000+(addr&0x3FFF)] = v
	case 2:
		w.wram[addr&0x3FFF] = v
	}
}

func (w *WRAM) Write7(addr uint32, v uint8) {
	if addr >= 0x380_0000 {
		w.wram7[addr&0xFFFF] = v
		return
	}

	switch w.cnt {
	case 0:
		w.wram7[addr&0xFFFF] = v
	case 1:
		w.wram[addr&0x3FFF] = v
	case 2:
		w.wram[0x4000+(addr&0x3FFF)] = v
	case 3:
		w.wram[addr&0x7FFF] = v
	}
}

func (w *WRAM) Read9(addr uint32) uint8 {
	switch w.cnt {
	case 0:
		return w.wram[addr&0x7FFF]
	case 1:
		return w.wram[0x4000+(addr&0x3FFF)]
	case 2:
		return w.wram[addr&0x3FFF]
	case 3:
		return 0 // should this clear ram?
	}
	return 0
}

func (w *WRAM) Read7(addr uint32) uint8 {
	if addr >= 0x380_0000 {
		return w.wram7[addr&0xFFFF]
	}

	switch w.cnt {
	case 0:
		return w.wram7[addr&0xFFFF]
	case 1:
		return w.wram[addr&0x3FFF]
	case 2:
		return w.wram[0x4000+(addr&0x3FFF)]
	case 3:
		return w.wram[addr&0x7FFF]
	}

	return 0
}

func (w *WRAM) ReadPtr9(addr uint32) unsafe.Pointer {
	switch w.cnt {
	case 0:
		return unsafe.Add(unsafe.Pointer(&w.wram), addr&0x7FFF)
	case 1:
		return unsafe.Add(unsafe.Pointer(&w.wram), 0x4000+(addr&0x3FFF))
	case 2:
		return unsafe.Add(unsafe.Pointer(&w.wram), addr&0x3FFF)
	case 3:
		return nil
	}

	return nil
}

func (w *WRAM) ReadPtr7(addr uint32) unsafe.Pointer {
	switch {
	case addr >= 0x380_0000:
		return unsafe.Add(unsafe.Pointer(&w.wram7), addr&0xFFFF)
	case addr >= 0x380_0000-0x20:
		// sonic brotherhood has arm7 use wram at 0x37F_FFFA -> 0x380_0000. Need to cancel read ptr near 0x380_0000
		return nil
	}

	switch w.cnt {
	case 0:
		return unsafe.Add(unsafe.Pointer(&w.wram7), addr&0xFFFF)
	case 1:
		return unsafe.Add(unsafe.Pointer(&w.wram), addr&0x3FFF)
	case 2:
		return unsafe.Add(unsafe.Pointer(&w.wram), 0x4000+(addr&0x3FFF))
	case 3:
		return unsafe.Add(unsafe.Pointer(&w.wram), addr&0x7FFF)
	}

	return nil
}
