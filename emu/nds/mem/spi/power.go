package spi

import (
	"github.com/aabalke/guac/emu/nds/utils"
)

const (
    REG_POWERMG = 0
    REG_BATTERY = 1
    REG_MICCTRL = 2
    REG_MICGAIN = 3
    REG_BACKLIT = 4
)

type Pmd struct {
    RegPowermg uint8
    RegBattery uint8
    RegMicctrl uint8
    RegMicgain uint8
    RegBacklit uint8
}

func (p *Pmd) Transfer(data []uint8) (reply []uint8, stat uint8)  {

    idx := data[0]

    if read := utils.BitEnabled(uint32(idx), 7); read {

        if len(data) < 2 {
            return nil, STAT_CONT
        }

        v := data[1]

        switch idx & 0x7F {
        case REG_POWERMG:
            p.RegPowermg = v
        case REG_MICCTRL:
            p.RegMicctrl = v & 1
        case REG_MICGAIN:
            p.RegMicgain = v & 3
        }

        return nil, STAT_DONE
    }

    switch idx & 0x7F {
    case REG_POWERMG:
        return []uint8{p.RegPowermg}, STAT_DONE
    case REG_BATTERY:
        return []uint8{0}, STAT_DONE
    case REG_MICCTRL:
        return []uint8{p.RegMicctrl}, STAT_DONE

    default:
        return nil, STAT_DONE
    }
}
