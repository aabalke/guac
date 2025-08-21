package spi

import "github.com/aabalke/guac/emu/nds/utils"

const (
    REG_POWERMG = 0
    REG_BATTERY = 1
    REG_MICCTRL = 2
    REG_MICGAIN = 3
    REG_BACKLIT = 4
)

type Pmd struct {
	isIdxSet bool
	RegIdx   uint8
	isRead   bool

    RegPowermg uint8
    RegBattery uint8
    RegMicctrl uint8
    RegMicgain uint8
    RegBacklit uint8
}

func (p *Pmd) Write(v uint8)  {

	if !p.isIdxSet {
		p.isRead = utils.BitEnabled(uint32(v), 7)
        p.RegIdx &= 0b111_1111
        p.isIdxSet = true
		return
	}

    p.isIdxSet = false

    switch p.RegIdx {
    case REG_POWERMG:
        p.RegPowermg = v & 0b111_1111
    case REG_BATTERY:
        return
    case REG_MICCTRL:
        p.RegMicctrl = v & 0b1
    case REG_MICGAIN:
        p.RegMicgain = v & 0b11
    default: panic("UNKNOWN SPI POWER MANAGEMENT REG IDX")
    }
}

func(p *Pmd) Read() uint8 {

	if !p.isIdxSet {
        panic("IDX IS NOT SET FOR READ")
	}

    p.isIdxSet = false

    switch p.RegIdx {
    case REG_POWERMG:
        return p.RegPowermg
    case REG_BATTERY:
        return p.RegBattery
    case REG_MICCTRL:
        return p.RegMicctrl
    case REG_MICGAIN:
        return p.RegMicgain
    default: panic("UNKNOWN SPI POWER MANAGEMENT REG IDX")
    }
}
