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
        *d = Dispstat(uint16(v) << 8)
        return
    }

    v &^= 0b111
    *d = Dispstat(uint16(v))
}

func (d *Dispstat) SetVBlank(v bool) {

    if v {
        *d = Dispstat((uint16(*d) &^ 1) | 1)
        return
    }

    *d = Dispstat((uint16(*d) &^ 1))
}

func (d *Dispstat) SetHBlank(v bool) {

    if v {
        *d = Dispstat((uint16(*d) &^ 10) | 10)
        return
    }

    *d = Dispstat((uint16(*d) &^ 10))
}

func (d *Dispstat) SetVCounter(scanline int) {

    lyc := int(*d >> 8)

    if scanline == lyc {
        *d = Dispstat((uint16(*d) &^ 100) | 100)
        return

    }

    *d = Dispstat((uint16(*d) &^ 100))
}
