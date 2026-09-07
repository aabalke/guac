package mem

type Dispstat struct {
	LYC   uint32
	V     bool // shared
	H     bool // shared
	VC    bool
	VIrq  bool
	HIrq  bool
	VCIrq bool
}

func (d *Dispstat) Read(b uint32) uint8 {
	switch b {
	case 0:

		v := uint8(d.LYC>>8) << 7

		if d.V {
			v |= 1 << 0
		}

		if d.H {
			v |= 1 << 1
		}

		if d.VC {
			v |= 1 << 2
		}

		if d.VIrq {
			v |= 1 << 3
		}

		if d.HIrq {
			v |= 1 << 4
		}

		if d.VCIrq {
			v |= 1 << 5
		}

		return v

	case 1:
		return uint8(d.LYC)
	default:
		panic("not possible")
	}
}

func (d *Dispstat) Write(b uint32, v uint8) {
	switch b {
	case 0:
		d.V = v&(1<<0) != 0
		d.H = v&(1<<1) != 0
		d.VC = v&(1<<2) != 0
		d.VIrq = v&(1<<3) != 0
		d.HIrq = v&(1<<4) != 0
		d.VCIrq = v&(1<<5) != 0
		d.LYC = (d.LYC & 0xFF) | (uint32(v&0x80) << 8)
	case 1:
		d.LYC = (d.LYC &^ 0xFF) | uint32(v)
	}
}
