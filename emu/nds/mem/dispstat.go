package mem

type Dispstat struct {
	A7, A9 uint16
}

func (d *Dispstat) Write(v uint8, hi bool, arm9 bool) {

	r := &d.A7
	if arm9 {
		r = &d.A9
	}

	if hi {
		*r = (uint16(*r) & 0xFF) | (uint16(v) << 8)
		return
	}

	v &^= 0b111
	*r = (uint16(*r) &^ 0b0011_1000) | uint16(v)
}

func (d *Dispstat) SetVBlank(v bool) {

	if v {
		d.A7 |= 0b1
		d.A9 |= 0b1
		return
	}

	d.A7 &^= 0b1
	d.A9 &^= 0b1
}

func (d *Dispstat) SetHBlank(v bool) {

	if v {
		d.A7 |= 0b10
		d.A9 |= 0b10
		return
	}

	d.A7 &^= 0b10
	d.A9 &^= 0b10
}

func (d *Dispstat) SetVCFlag(v, arm9 bool) {

	r := &d.A7
	if arm9 {
		r = &d.A9
	}

	if v {
		*r |= 0b100
		*r |= 0b100
		return
	}

	*r &^= 0b100
	*r &^= 0b100
}

func (d *Dispstat) GetLYC(arm9 bool) uint32 {

	r := &d.A7
	if arm9 {
		r = &d.A9
	}

	return uint32(*r>>8) + ((uint32(*r>>7) & 1) << 8)
}
