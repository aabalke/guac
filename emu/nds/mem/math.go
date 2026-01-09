package mem

import "math"

func w(d *uint64, v uint8, b uint32) {
	*d = (*d &^ (0xFF << (b << 3))) | (uint64(v) << (b << 3))
}

func r(d uint64, b uint32) uint8 {
	return uint8(d >> (b << 3))
}

type Div struct {
	cnt, num, den, res, rem uint64
}

func (d *Div) Write(addr uint32, v uint8) {

	switch {
	case addr >= 0x280 && addr < 0x284:
		w(&d.cnt, v, addr-0x280)

	case addr >= 0x290 && addr < 0x298:
		w(&d.num, v, addr-0x290)

	case addr >= 0x298 && addr < 0x2a0:
		w(&d.den, v, addr-0x298)

	default:
		return
	}

	d.Calc()
}

func (d *Div) Read(addr uint32) uint8 {
	switch {
	case addr >= 0x280 && addr < 0x284:
		return r(d.cnt, addr-0x280)

	case addr >= 0x290 && addr < 0x298:
		return r(d.num, addr-0x290)

	case addr >= 0x298 && addr < 0x2a0:
		return r(d.den, addr-0x298)

	case addr >= 0x2a0 && addr < 0x2a8:
		return r(d.res, addr-0x2a0)

	case addr >= 0x2a8 && addr < 0x2b0:
		return r(d.rem, addr-0x2a8)

	default:
		return 0
	}
}

func (d *Div) Calc() {

	d.cnt &^= 1 << 14

	if d.den == 0 {
		d.cnt |= 1 << 14
	}

	switch mode := d.cnt & 0b11; mode {
	case 1:

		if uint32(d.den) == 0 {
			d.rem = d.num

			if int64(d.num) < 0 {
				d.res = 1
			} else {
				d.res = ^uint64(0)
			}
			return
		}

		res := int64(d.num) / int64(int32(d.den))
		rem := int64(d.num) % int64(int32(d.den))

		d.res = uint64(res)
		d.rem = uint64((int32(rem)))

	case 2:

		if uint32(d.den) == 0 {
			d.rem = d.num

			if int64(d.num) < 0 {
				d.res = 1
			} else {
				d.res = ^uint64(0)
			}
			return
		}

		res := int64(d.num) / int64(d.den)
		rem := int64(d.num) % int64(d.den)

		d.res = uint64(res)
		d.rem = uint64(rem)

	default:

		if uint32(d.den) == 0 {
			d.rem = d.num

			if int32(d.num) < 0 {
				d.res = 1
				d.rem |= 0xffff_ffff_0000_0000
			} else {
				d.res = ^uint64(0)
			}

			d.res ^= 0xffff_ffff_0000_0000
			return
		}

		d.res = uint64(int32(d.num) / int32(d.den))
		d.rem = uint64(int32(d.num) % int32(d.den))

		if int32(d.num) == math.MinInt32 && int32(d.den) == -1 {
			d.res ^= 0xffff_ffff_0000_0000
		}
	}
}

type Sqrt struct {
	is64  bool
	param uint64
	res   uint32
}

func (s *Sqrt) Write(addr uint32, v uint8) {

	switch {
	case addr == 0x2B0:
		s.is64 = v&1 != 0
	case addr >= 0x2B8 && addr < 0x2C0:
		w(&s.param, v, addr-0x2B8)
	default:
		return
	}

	s.Calc()
}

func (s *Sqrt) Read(addr uint32) uint8 {

	switch {
	case addr == 0x2B0:

		if s.is64 {
			return 1
		}

		return 0

	case addr >= 0x2B4 && addr < 0x2B8:
		return r(uint64(s.res), addr-0x2B4)

	case addr >= 0x2B8 && addr < 0x2C0:
		return r(s.param, addr-0x2B8)
	}

	return 0
}

func (s *Sqrt) Calc() {

	if s.is64 {
		s.res = uint32(sqrt(s.param))
		return
	}
	s.res = uint32(sqrt(s.param & 0xFFFF_FFFF))
}

func sqrt(input uint64) uint64 {

	if input == 0 {
		return 0
	}

	lo, hi, bound := uint64(0), input, uint64(1)

	for bound < hi {
		hi >>= 1
		bound <<= 1
	}

	for {
		hi = input
		acc := uint64(0)
		lo = bound

		for {
			oldLower := lo
			if lo <= hi>>1 {
				lo <<= 1
			}
			if oldLower >= hi>>1 {
				break
			}
		}

		for {
			acc <<= 1
			if hi >= lo {
				acc++
				hi -= lo
			}
			if lo == bound {
				break
			}
			lo >>= 1
		}

		oldBound := bound
		bound += acc
		bound >>= 1
		if bound >= oldBound {
			bound = oldBound
			break
		}
	}

	return bound
}
