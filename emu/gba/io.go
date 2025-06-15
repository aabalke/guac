package gba

const (
    DISPCNT  = 0x0000
    DISPSTAT  = 0x0004
	KEYINPUT = 0x0130
	KEYCNT   = 0x0132
	BG0CNT   = 0x0008
	BG1CNT   = 0x000A
	BG2CNT   = 0x000C
	BG3CNT   = 0x000E
)

type Dispstat uint16

func (d *Dispstat) Write(v uint8, hi bool) {

    if hi {
        *d = Dispstat((uint16(*d) & 0b1111_1111) | (uint16(v) << 8))
        return
    }

    v &^= 0b111
    *d = Dispstat((uint16(*d) &^ 0b0011_1000) | uint16(v))
}

func (d *Dispstat) SetVBlank(v bool) {

    if v {
        *d = Dispstat(uint16(*d) | 0b1)
        return
    }

    *d = Dispstat((uint16(*d) &^ 0b1))
}

func (d *Dispstat) SetHBlank(v bool) {

    if v {
        *d = Dispstat(uint16(*d) | 0b10)
        return
    }

    *d = Dispstat((uint16(*d) &^ 0b10))
}

func (d *Dispstat) SetVCFlag(v bool) {

    if v {
        *d = Dispstat(uint16(*d) | 0b100)
        return
    }

    *d = Dispstat((uint16(*d) &^ 0b100))
}

//func (d *Dispstat) SetVCounter(scanline int) {
//
//    lyc := int(uint32(*d >> 8))
//
//    if scanline == lyc {
//        *d = Dispstat((uint16(*d) &^ 0b100) | 0b100)
//        return
//
//    }
//
//    *d = Dispstat((uint16(*d) &^ 0b100))
//}

//type IRQFlags uint32
//
//func (f *IRQFlags) write(v uint8, hi bool) {
//
//    if hi {
//        *f &^= IRQFlags(uint32(v) << 8)
//        return
//    }
//
//    *f &^= IRQFlags(uint32(v))
//}
//
//func (f *IRQFlags) read(hi bool) uint8 {
//
//    if hi {
//        return uint8(uint32(*f) >> 8)
//    }
//
//    return uint8(uint32(*f))
//}
