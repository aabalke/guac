package mem

import "unsafe"

type Tcm struct {
	Itcm [0x8000]uint8
	Dtcm [0x4000]uint8

	ItcmSize uint32
	DtcmSize uint32
	DtcmBase uint32

	ItcmEnabled  bool
	ItcmLoadMode bool
	DtcmEnabled  bool
	DtcmLoadMode bool
}

func (t *Tcm) ReadDtcm(addr uint32) (uint8, bool) {

	if t.DtcmLoadMode || !t.DtcmEnabled {
		return 0, false
	}

	return t.Dtcm[(addr-t.DtcmBase)&0x3FFF], true
}

func (t *Tcm) Read(addr uint32) (uint8, bool) {

	if addr < t.ItcmSize {

		if t.ItcmLoadMode || !t.ItcmEnabled {
			return 0, false
		}

		return t.Itcm[addr&0x7FFF], true

	}

    if addr >= t.DtcmBase && addr < t.DtcmBase+t.DtcmSize {
		return t.ReadDtcm(addr)
	}

	return 0, false
}

func (t *Tcm) ReadTcmWindow(addr uint32) (uint8, bool) {
	if notDtcm := !(addr >= t.DtcmBase && addr < t.DtcmBase+t.DtcmSize); notDtcm {
		return 0, false
	}

	return t.Read(addr)
}

func (t *Tcm) ReadTcmWindowPtr(addr uint32) (unsafe.Pointer, bool) {
	if notDtcm := !(addr >= t.DtcmBase && addr < t.DtcmBase+t.DtcmSize); notDtcm {
		return nil, false
	}

	return t.ReadPtr(addr)
}

func (t *Tcm) ReadPtr(addr uint32) (unsafe.Pointer, bool) {

	if addr < t.ItcmSize {

		if t.ItcmLoadMode || !t.ItcmEnabled {
			return nil, false
		}

		return unsafe.Add(unsafe.Pointer(&t.Itcm), addr&0x7FFF), true

	} else if addr >= t.DtcmBase && addr < t.DtcmBase+t.DtcmSize {
		return t.ReadDtcmPtr(addr)
	}

	return nil, false
}

func (t *Tcm) ReadDtcmPtr(addr uint32) (unsafe.Pointer, bool) {

	if t.DtcmLoadMode || !t.DtcmEnabled {
		return nil, false
	}

	return unsafe.Add(unsafe.Pointer(&t.Dtcm), (addr-t.DtcmBase)&0x3FFF), true
}

func (t *Tcm) WriteDtcm(addr uint32, v uint8) bool {

	if !t.DtcmEnabled {
		return false
	}

	t.Dtcm[(addr-t.DtcmBase)&0x3FFF] = v
	return true
}

func (t *Tcm) Write(addr uint32, v uint8) bool {

	if addr < t.ItcmSize {

		if !t.ItcmEnabled {
			return false
		}

		t.Itcm[addr&0x7FFF] = v
		return true
	}

	if addr >= t.DtcmBase && addr < t.DtcmBase+t.DtcmSize {
		return t.WriteDtcm(addr, v)
	}

	return false
}

func (t *Tcm) WriteTcmWindow(addr uint32, v uint8) bool {
	if notDtcm := !(addr >= t.DtcmBase && addr < t.DtcmBase+t.DtcmSize); notDtcm {
		return false
	}

	return t.Write(addr, v)
}
