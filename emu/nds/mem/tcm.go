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

func (t *Tcm) Read(addr uint32) (uint8, bool) {
	if t.ItcmEnabled && !t.ItcmLoadMode && addr < t.ItcmSize {
		return t.Itcm[addr&0x7FFF], true
	}

	if t.DtcmEnabled && !t.DtcmLoadMode && addr >= t.DtcmBase && addr < t.DtcmBase+t.DtcmSize {
		return t.Dtcm[(addr-t.DtcmBase)&0x3FFF], true
	}

	return 0, false
}

func (t *Tcm) ReadPtr(addr uint32) unsafe.Pointer {
	if t.ItcmEnabled && !t.ItcmLoadMode && addr < t.ItcmSize {
		return unsafe.Add(unsafe.Pointer(&t.Itcm), addr&0x7FFF)
	}

	if t.DtcmEnabled && !t.DtcmLoadMode && addr >= t.DtcmBase && addr < t.DtcmBase+t.DtcmSize {
		return unsafe.Add(unsafe.Pointer(&t.Dtcm), (addr-t.DtcmBase)&0x3FFF)
	}

	return nil
}

func (t *Tcm) Write(addr uint32, v uint8) bool {
	if t.ItcmEnabled && addr < t.ItcmSize {
		t.Itcm[addr&0x7FFF] = v
		return true
	}

	if t.DtcmEnabled && addr >= t.DtcmBase && addr < t.DtcmBase+t.DtcmSize {
		t.Dtcm[(addr-t.DtcmBase)&0x3FFF] = v
		return true
	}

	return false
}

func (t *Tcm) WritePtr(addr uint32) unsafe.Pointer {
	if t.ItcmEnabled && addr < t.ItcmSize {
		return unsafe.Add(unsafe.Pointer(&t.Itcm), addr&0x7FFF)
	}

	if t.DtcmEnabled && addr >= t.DtcmBase && addr < t.DtcmBase+t.DtcmSize {
		return unsafe.Add(unsafe.Pointer(&t.Dtcm), (addr-t.DtcmBase)&0x3FFF)
	}

	return nil
}
