package mem

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

func (t *Tcm) ReadDtcm(addr uint32) uint8 {

	if t.DtcmLoadMode || !t.DtcmEnabled {
		return 0
	}

	return t.Dtcm[(addr-t.DtcmBase)&0x3FFF]
}

func (t *Tcm) Read(addr uint32) uint8 {

	if addr < t.ItcmSize {

		if t.ItcmLoadMode || !t.ItcmEnabled {
			return 0
		}

		return t.Itcm[addr&0x7FFF]

	} else if addr >= t.DtcmBase && addr < t.DtcmBase+t.DtcmSize {
		return t.ReadDtcm(addr)
	}

	return 0
}

func (t *Tcm) WriteDtcm(addr uint32, v uint8) {

	if !t.DtcmEnabled {
		return
	}

	t.Dtcm[(addr-t.DtcmBase)&0x3FFF] = v
}

func (t *Tcm) Write(addr uint32, v uint8) {

	if addr < t.ItcmSize {

		if !t.ItcmEnabled {
			return
		}

		t.Itcm[addr&0x7FFF] = v
		return

	} else if addr >= t.DtcmBase && addr < t.DtcmBase+t.DtcmSize {
		t.WriteDtcm(addr, v)
	}
}
