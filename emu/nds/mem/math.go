package mem

type Div struct {
	CNT, NUM, DEN, RES, REM uint64
}

func (d *Div) Write(addr uint32, v uint8) {

	if addr >= 0x283 && addr < 0x290 {
		return
	}

	w := func(dst uint64, v uint8, b uint32) uint64 {
		dst &^= (0xFF << (8 * b))
		dst |= uint64(v) << (8 * b)
		return dst
	}

	switch addr {
	case 0x280:
		d.CNT = w(d.CNT, v, 0)
	case 0x281:
		d.CNT = w(d.CNT, v, 1)
	case 0x282:
		d.CNT = w(d.CNT, v, 2)
	case 0x283:
		d.CNT = w(d.CNT, v, 3)

	case 0x290:
		d.NUM = w(d.NUM, v, 0)
	case 0x291:
		d.NUM = w(d.NUM, v, 1)
	case 0x292:
		d.NUM = w(d.NUM, v, 2)
	case 0x293:
		d.NUM = w(d.NUM, v, 3)
	case 0x294:
		d.NUM = w(d.NUM, v, 4)
	case 0x295:
		d.NUM = w(d.NUM, v, 5)
	case 0x296:
		d.NUM = w(d.NUM, v, 6)
	case 0x297:
		d.NUM = w(d.NUM, v, 7)
	case 0x298:
		d.DEN = w(d.DEN, v, 0)
	case 0x299:
		d.DEN = w(d.DEN, v, 1)
	case 0x29A:
		d.DEN = w(d.DEN, v, 2)
	case 0x29B:
		d.DEN = w(d.DEN, v, 3)
	case 0x29C:
		d.DEN = w(d.DEN, v, 4)
	case 0x29D:
		d.DEN = w(d.DEN, v, 5)
	case 0x29E:
		d.DEN = w(d.DEN, v, 6)
	case 0x29F:
		d.DEN = w(d.DEN, v, 7)
	}

	d.Calc()
}

func (d *Div) Read(addr uint32) uint8 {

	r := func(dst, b uint64) uint8 {
		return uint8(dst >> (8 * b))
	}

	switch addr {
	case 0x280:
		return r(d.CNT, 0)
	case 0x281:
		return r(d.CNT, 1)
	case 0x282:
		return r(d.CNT, 2)
	case 0x283:
		return r(d.CNT, 3)
	case 0x284:
		return 0
	case 0x285:
		return 0
	case 0x286:
		return 0
	case 0x287:
		return 0
	case 0x288:
		return 0
	case 0x289:
		return 0
	case 0x28A:
		return 0
	case 0x28B:
		return 0
	case 0x28C:
		return 0
	case 0x28D:
		return 0
	case 0x28E:
		return 0
	case 0x28F:
		return 0

	case 0x290:
		return r(d.NUM, 0)
	case 0x291:
		return r(d.NUM, 1)
	case 0x292:
		return r(d.NUM, 2)
	case 0x293:
		return r(d.NUM, 3)
	case 0x294:
		return r(d.NUM, 4)
	case 0x295:
		return r(d.NUM, 5)
	case 0x296:
		return r(d.NUM, 6)
	case 0x297:
		return r(d.NUM, 7)
	case 0x298:
		return r(d.DEN, 0)
	case 0x299:
		return r(d.DEN, 1)
	case 0x29A:
		return r(d.DEN, 2)
	case 0x29B:
		return r(d.DEN, 3)
	case 0x29C:
		return r(d.DEN, 4)
	case 0x29D:
		return r(d.DEN, 5)
	case 0x29E:
		return r(d.DEN, 6)
	case 0x29F:
		return r(d.DEN, 7)
	case 0x2A0:
		return r(d.RES, 0)
	case 0x2A1:
		return r(d.RES, 1)
	case 0x2A2:
		return r(d.RES, 2)
	case 0x2A3:
		return r(d.RES, 3)
	case 0x2A4:
		return r(d.RES, 4)
	case 0x2A5:
		return r(d.RES, 5)
	case 0x2A6:
		return r(d.RES, 6)
	case 0x2A7:
		return r(d.RES, 7)
	case 0x2A8:
		return r(d.REM, 0)
	case 0x2A9:
		return r(d.REM, 1)
	case 0x2AA:
		return r(d.REM, 2)
	case 0x2AB:
		return r(d.REM, 3)
	case 0x2AC:
		return r(d.REM, 4)
	case 0x2AD:
		return r(d.REM, 5)
	case 0x2AE:
		return r(d.REM, 6)
	case 0x2AF:
		return r(d.REM, 7)
	}

	panic("UNKNOWN DIV READ")
}

func (d *Div) Calc() {

	d.CNT &^= 1 << 14

	if d.DEN == 0 {
		d.CNT |= 1 << 14
	}

	switch mode := d.CNT & 0b11; mode {
	case 1:

		if uint32(d.DEN) == 0 {
			d.REM = d.NUM
			if d.NUM&(0x8000_0000_0000_0000) == 0 {
				d.RES = uint64(0xFFFF_FFFF_FFFF_FFFF)
			} else {
				d.RES = 1
			}

			return
		}

		numerator := int64(d.NUM)
		denominator := int32(d.DEN)

		res := numerator / int64(denominator)
		rem := numerator % int64(denominator)

		d.RES = uint64(res)
		d.REM = uint64((int32(rem)))

	case 2:

		if d.DEN == 0 {
			d.REM = d.NUM
			if d.NUM&(0x8000_0000_0000_0000) == 0 {
				d.RES = uint64(0xFFFF_FFFF_FFFF_FFFF)
			} else {
				d.RES = 1
			}

			return
		}

		res := int64(d.NUM) / int64(d.DEN)
		rem := int64(d.NUM) % int64(d.DEN)

		d.RES = uint64(res)
		d.REM = uint64(rem)

	default:

		if uint32(d.DEN) == 0 {
			d.REM = d.NUM
			if d.NUM&(0x0000_0000_8000_0000) == 0 {
				d.RES = uint64(0xFFFF_FFFF_FFFF_FFFF)
			} else {
				d.RES = 1
				d.REM |= 0xFFFF_FFFF_0000_0000
			}

			d.RES ^= 0xFFFF_FFFF_0000_0000
			return
		}

		numerator := int32(d.NUM)
		denominator := int32(d.DEN)

		quotient := numerator / int32(denominator)
		remainder := numerator % int32(denominator)

		d.RES = uint64(int32(quotient))
		d.REM = uint64(int32(remainder))

		if uint32(d.NUM) == 0x8000_0000 && int32(d.DEN) == -1 {
			d.RES ^= 0xFFFF_FFFF_0000_0000
		}
	}
}

type Sqrt struct {
	is64  bool
	PARAM uint64
	RES   uint32
}

func (s *Sqrt) Write(addr uint32, v uint8) {

	w := func(dst uint64, v uint8, b uint32) uint64 {
		dst &^= (0xFF << (8 * b))
		dst |= uint64(v) << (8 * b)
		return dst
	}

	// ignoring busy bit since instant

	switch addr {
	case 0x2B0:
		s.is64 = v&1 == 1
	case 0x2B8:
		s.PARAM = w(s.PARAM, v, 0)
	case 0x2B9:
		s.PARAM = w(s.PARAM, v, 1)
	case 0x2BA:
		s.PARAM = w(s.PARAM, v, 2)
	case 0x2BB:
		s.PARAM = w(s.PARAM, v, 3)
	case 0x2BC:
		s.PARAM = w(s.PARAM, v, 4)
	case 0x2BD:
		s.PARAM = w(s.PARAM, v, 5)
	case 0x2BE:
		s.PARAM = w(s.PARAM, v, 6)
	case 0x2BF:
		s.PARAM = w(s.PARAM, v, 7)
	}

	if !(addr >= 0x2B4 && addr < 0x2B8) {
		s.Calc()
	}
}

func (s *Sqrt) Read(addr uint32) uint8 {

	r := func(dst, b uint64) uint8 {
		return uint8(dst >> (8 * b))
	}

	switch addr {
	case 0x2B0:

		if s.is64 {
			return 1
		}

		return 0

	case 0x2B4:
		return r(uint64(s.RES), 0)
	case 0x2B5:
		return r(uint64(s.RES), 1)
	case 0x2B6:
		return r(uint64(s.RES), 2)
	case 0x2B7:
		return r(uint64(s.RES), 3)
	case 0x2B8:
		return r(s.PARAM, 0)
	case 0x2B9:
		return r(s.PARAM, 1)
	case 0x2BA:
		return r(s.PARAM, 2)
	case 0x2BB:
		return r(s.PARAM, 3)
	case 0x2BC:
		return r(s.PARAM, 4)
	case 0x2BD:
		return r(s.PARAM, 5)
	case 0x2BE:
		return r(s.PARAM, 6)
	case 0x2BF:
		return r(s.PARAM, 7)
	}

	return 0
}

func (s *Sqrt) Calc() {

	if s.is64 {
		s.RES = uint32(sqrt(s.PARAM))
		return
	}
	s.RES = uint32(sqrt(s.PARAM & 0xFFFF_FFFF))
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
