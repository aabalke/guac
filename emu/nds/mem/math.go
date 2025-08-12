package mem

import (
	"fmt"

	"github.com/aabalke/guac/emu/nds/utils"
)

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
        fmt.Printf("READING\n")
        fmt.Printf("DIV NU %d %d, DE %d %d, RES %d %d, REM %d %d DIV0 %t\n",
        uint32(d.NUM), uint32(d.NUM >> 32), uint32(d.DEN), uint32(d.DEN >> 32),
        uint32(d.RES), uint32(d.RES >> 32), uint32(d.REM), uint32(d.REM >> 32), utils.BitEnabled(uint32(d.CNT), 14))
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

    if d.DEN == 0 {
        d.REM = d.NUM
        // is this supposed to change to int32 in int32 version?
        if d.NUM & (0x8000_0000_0000_0000) == 0 {
            d.RES = uint64(0xFFFF_FFFF_FFFF_FFFF)
        } else {
            d.RES = 1
        }

        d.CNT |= 1 << 14
        return
    }

    d.CNT &^= 1 << 14

	switch mode := d.CNT & 0b11; mode {
	case 1:
	case 2:

        res := int64(d.NUM) / int64(d.DEN)
        rem := int64(d.NUM) % int64(d.DEN)

		d.RES = uint64(res)
		d.REM = uint64(rem)

	default:
	}
    //u32 numLo, numHi, denomLo, denomHi, resLo, resHi, remLo, remHi;


	//fmt.Printf("DIV 64 64 CNT %d NUM %d, DEN %d, RES %d, REM %d\n", int64(d.CNT), int64(d.DEN), int64(d.NUM), int64(d.RES), int64(d.REM))
}
