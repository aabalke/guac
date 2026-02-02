package mem

type Dispstat struct {
	V bool // shared
	H bool // shared

	A9VC    bool
	A9VIrq  bool
	A9HIrq  bool
	A9VCIrq bool
	A9LYC   uint32

	A7VC    bool
	A7VIrq  bool
	A7HIrq  bool
	A7VCIrq bool
	A7LYC   uint32
}

func (d *Dispstat) Read(hi, arm9 bool) uint8 {

	if hi {
		if arm9 {
			return uint8(d.A9LYC)
		}

		return uint8(d.A7LYC)
	}

	v := uint8(0)

	if d.V {
		v |= 1 << 0
	}

	if d.H {
		v |= 1 << 1
	}

	if arm9 {
		if d.A9VC {
			v |= 1 << 2
		}

		if d.A9VIrq {
			v |= 1 << 3
		}

		if d.A9HIrq {
			v |= 1 << 4
		}

		if d.A9VCIrq {
			v |= 1 << 5
		}

		v |= uint8(d.A9LYC>>8) << 7

		return v
	}

	if d.A7VC {
		v |= 1 << 2
	}

	if d.A7VIrq {
		v |= 1 << 3
	}

	if d.A7HIrq {
		v |= 1 << 4
	}

	if d.A7VCIrq {
		v |= 1 << 5
	}

	v |= uint8(d.A7LYC>>8) << 7

	return v
}

func (d *Dispstat) Write7(v uint8, hi bool) {

	if hi {
		d.A7LYC = (d.A7LYC &^ 0xFF) | uint32(v)
		return
	}

	d.V = v&(1<<0) != 0
	d.H = v&(1<<1) != 0

	d.A7VC = v&(1<<2) != 0
	d.A7VIrq = v&(1<<3) != 0
	d.A7HIrq = v&(1<<4) != 0
	d.A7VCIrq = v&(1<<5) != 0
	d.A7LYC = (d.A7LYC & 0xFF) | (uint32(v&0x80) << 8)
}

func (d *Dispstat) Write9(v uint8, hi bool) {

	if hi {
		d.A9LYC = (d.A9LYC &^ 0xFF) | uint32(v)
		return
	}

	d.V = v&(1<<0) != 0
	d.H = v&(1<<1) != 0

	d.A9VC = v&(1<<2) != 0
	d.A9VIrq = v&(1<<3) != 0
	d.A9HIrq = v&(1<<4) != 0
	d.A9VCIrq = v&(1<<5) != 0
	d.A9LYC = (d.A9LYC & 0xFF) | (uint32(v&0x80) << 8)
}
