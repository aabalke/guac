package gameboy

type Flags struct {
    Z, S, H, C bool
}

const (
	AF = iota
	BC
	DE
	HL
)

func (c *Cpu) setHl(v uint16) {
	c.h = uint8(v>>8)
	c.l = uint8(v)
}

func (c *Cpu) setDe(v uint16) {
	c.d = uint8(v>>8)
	c.e = uint8(v)
}

func (c *Cpu) setBc(v uint16) {
	c.b = uint8(v>>8)
	c.c = uint8(v)
}

func (c *Cpu) setAf(v uint16) {
	c.a = uint8(v>>8)
	c.f.Set(uint8(v))
}

func (c *Cpu) af() uint16 {
	return uint16(c.a)<<8 | uint16(c.f.Get())
}

func (c *Cpu) bc() uint16 {
	return uint16(c.b)<<8 | uint16(c.c)
}

func (c *Cpu) de() uint16 {
	return uint16(c.d)<<8 | uint16(c.e)
}

func (c *Cpu) hl() uint16 {
	return uint16(c.h)<<8 | uint16(c.l)
}

func (c *Cpu) setCombinedRegister(v uint16, register int) {

	switch register {
	case AF:
		c.setAf(v)
	case BC:
		c.setBc(v)
	case DE:
		c.setDe(v)
	case HL:
		c.setHl(v)
	}
}

func (c *Cpu) getCombinedRegister(register int) uint16 {
	switch register {
	case AF:
		return c.af()
	case BC:
		return c.bc()
	case DE:
		return c.de()
	case HL:
		return c.hl()
	}
	return 0
}

func (f *Flags) Get() uint8 {

	var v uint8

	if f.Z {
		v |= 1 << 7
	}

	if f.S {
		v |= 1 << 6
	}

	if f.H {
		v |= 1 << 5
	}

	if f.C {
		v |= 1 << 4
	}

	return v
}

func (f *Flags) Set(v uint8) {
	f.Z = (v>>7)&1 != 0
	f.S = (v>>6)&1 != 0
	f.H = (v>>5)&1 != 0
    f.C = (v>>4)&1 != 0
}
