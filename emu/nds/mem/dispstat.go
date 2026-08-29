package mem

type Dispstat struct {
	A9LYC uint32
	A7LYC uint32

	V bool // shared
	H bool // shared

	A9VC    bool
	A9VIrq  bool
	A9HIrq  bool
	A9VCIrq bool

	A7VC    bool
	A7VIrq  bool
	A7HIrq  bool
	A7VCIrq bool
}

func (d *Dispstat) Read9(b uint32) uint8 {
	switch b {
	case 0:

		v := uint8(d.A9LYC>>8) << 7

		if d.V {
			v |= 1 << 0
		}

		if d.H {
			v |= 1 << 1
		}

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

		return v

	case 1:
		return uint8(d.A9LYC)
	default:
		panic("not possible")
	}
}

func (d *Dispstat) Read7(b uint32) uint8 {
	switch b {
	case 0:

		v := uint8(d.A7LYC>>8) << 7

		if d.V {
			v |= 1 << 0
		}

		if d.H {
			v |= 1 << 1
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

		return v

	case 1:
		return uint8(d.A7LYC)
	default:
		panic("not possible")
	}
}

func (d *Dispstat) Write7(b uint32, v uint8) {
	switch b {
	case 0:
		d.V = v&(1<<0) != 0
		d.H = v&(1<<1) != 0

		d.A7VC = v&(1<<2) != 0
		d.A7VIrq = v&(1<<3) != 0
		d.A7HIrq = v&(1<<4) != 0
		d.A7VCIrq = v&(1<<5) != 0
		d.A7LYC = (d.A7LYC & 0xFF) | (uint32(v&0x80) << 8)
	case 1:
		d.A7LYC = (d.A7LYC &^ 0xFF) | uint32(v)
	}
}

func (d *Dispstat) Write9(b uint32, v uint8) {
	switch b {
	case 0:

		d.V = v&(1<<0) != 0
		d.H = v&(1<<1) != 0

		d.A9VC = v&(1<<2) != 0
		d.A9VIrq = v&(1<<3) != 0
		d.A9HIrq = v&(1<<4) != 0
		d.A9VCIrq = v&(1<<5) != 0
		d.A9LYC = (d.A9LYC & 0xFF) | (uint32(v&0x80) << 8)
	case 1:
		d.A9LYC = (d.A9LYC &^ 0xFF) | uint32(v)

	}
}
