package spi

import (
	"fmt"

	"github.com/aabalke/guac/emu/nds/utils"
)

const (
	DEV_POWER = 0
	DEV_FIRMW = 1
	DEV_TOUCH = 2
)

type Spi struct {
	CNT                uint16
	Device             uint8
	Hold, Irq, Enabled bool

    Data uint8

    Pmd Pmd
    Firmware Firmware
}

func (s *Spi) WriteCNT(b, v uint8) {

	switch b {
	case 0:

		v &= 0b1000_0011

		s.CNT &^= 0xFF
		s.CNT |= uint16(v)

	case 1:

		v &= 0b1100_1111

		s.CNT &= 0xFF
		s.CNT |= uint16(v) << 8

		s.Device = v & 0b11
        hold := utils.BitEnabled(uint32(v), 3)

        if s.Hold && !hold {
            s.Firmware.Reset()
        }

        s.Hold = hold


        s.Irq = utils.BitEnabled(uint32(v), 6)
        s.Enabled = utils.BitEnabled(uint32(v), 7)
	}
}

func (s *Spi) ReadCNT(b uint8) uint8 {
	return uint8(s.CNT >> (8 * b))
}

func (s *Spi) WriteData(v uint8) {

    // need to get data from spi device

    switch s.Device {
    case DEV_POWER:
        if !s.Pmd.isIdxSet {
            s.Pmd.Write(v)
            return
        }

        if s.Pmd.isRead {
            s.Data = s.Pmd.Read()
            return
        }

        s.Pmd.Write(v)

    case DEV_FIRMW:

        // temp

        //fmt.Printf("WRITING TO FIRM %02X CNT %04X\n", v, s.CNT)

        s.Data = s.Firmware.Read()
        s.Firmware.Write(v)

    default: panic(fmt.Sprintf("UNSETUP SPI IN WRITE DATA %d", s.Device))
    }
}

func (s *Spi) ReadData() uint8 {

    //fmt.Printf("READING DATA\n")

    return s.Data
}
