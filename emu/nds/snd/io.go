package snd

import (
	"github.com/aabalke/guac/emu/nds/utils"
)

func (s *Snd) Write(addr uint32, v uint8) {

	addr &= 0xFFFF

    switch {
    case addr < 0x400:
        return
    case addr < 0x500:
        i := (addr & 0xF0) >> 4
        s.Channels[i].Write(addr, v)

    case addr < 0x600:

		switch addr {
		case 0x500:

			s.VolMaster = utils.GetVarData(uint32(v), 0, 6)

		case 0x501:

            s.LOut = (v & 0b11) >> 0
            s.ROut = (v & 0b11) >> 2
            s.NoOutCh1 = utils.BitEnabled(uint32(v), 4)
            s.NoOutCh3 = utils.BitEnabled(uint32(v), 5)
            s.Enabled = utils.BitEnabled(uint32(v), 7)

		case 0x504:

            s.Bias &^= 0xFF
            s.Bias |= uint32(v)

		case 0x505:

            s.Bias &= 0xFF
            s.Bias |= uint32(v & 0b11) << 8

		}

		return
	}

}

func (c *Channel) Write(addr uint32, v uint8) {

    addr &= 0xF

    switch addr {
    case 0x0:
        c.VolMul = uint32(v & 0b111_1111)
    case 0x1:
        c.VolDiv = uint32(v & 0b11)
        c.Hold = utils.BitEnabled(uint32(v), 7)
    case 0x2:
        c.Panning = uint32(v & 0b111_1111)
    case 0x3:

        c.Duty = uint32(v & 0b111)
        c.RepeatMode = uint32(v >> 3) & 0b11
        c.Format = uint32(v >> 5) & 0b11
        busy := utils.BitEnabled(uint32(v), 7)

        if busy {
            c.Start = true
        } else {
            c.Start = false
            c.Playing = false
        }

    case 0x4:

        c.SrcAddr &^= 0xFF
        c.SrcAddr |= uint32(v &^ 0b11)

    case 0x5:

        c.SrcAddr &^= 0xFF << 8
        c.SrcAddr |= uint32(v) << 8

    case 0x6:

        c.SrcAddr &^= 0xFF << 16
        c.SrcAddr |= uint32(v) << 16

    case 0x7:

        c.SrcAddr &^= 0xFF << 24
        c.SrcAddr |= uint32(v & 0b111) << 24

    case 0x8:

        c.TimerValue &^= 0xFF
        c.TimerValue |= uint16(v)

    case 0x9:

        c.TimerValue &^= 0xFF << 8
        c.TimerValue |= uint16(v) << 8
 
    case 0xA:

        c.StartPosition &^= 0xFF
        c.StartPosition |= uint16(v)

    case 0xB:

        c.StartPosition &^= 0xFF << 8
        c.StartPosition |= uint16(v) << 8

    case 0xC:

        c.SndLength &^= 0xFF
        c.SndLength |= uint32(v)

    case 0xD:

        c.SndLength &^= 0xFF << 8
        c.SndLength |= uint32(v) << 8

    case 0xE:

        c.SndLength &^= 0xFF << 16
        c.SndLength |= uint32(v & 0b11_1111) << 16
    }
}

func (s *Snd) Read(addr uint32) uint8 {

	addr &= 0xFFF

	if addr >= 0x500 {

		switch addr {
		case 0x500:

            return uint8(s.VolMaster)

		case 0x501:

            v := s.LOut
            v |= s.ROut << 2

            if s.NoOutCh1 {
                v |= 1 << 4
            }

            if s.NoOutCh3 {
                v |= 1 << 5
            }

            if s.Enabled {
                v |= 1 << 7
            }

            return v

		case 0x504:

            return uint8(s.Bias)

		case 0x505:

            return uint8(s.Bias >> 8)
		}

		return 0
	}

    i := (addr & 0xF0) >> 4

    return s.Channels[i].Read(addr)
}

func (c *Channel) Read(addr uint32) uint8 {

    addr &= 0xF

    switch addr {
    case 0x0:

        return uint8(c.VolMul)

    case 0x1:

        v := uint8(c.VolDiv)

        if c.Hold {
            v |= 1 << 7
        }

        return v

    case 0x2:

        return uint8(c.Panning)

    case 0x3:

        v := c.Duty
        v |= c.RepeatMode << 3
        v |= c.Format << 5

        if c.Playing {
            v |= 1 << 7
        }

        return uint8(v)
    }

    return 0
}
