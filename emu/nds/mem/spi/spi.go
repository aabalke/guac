package spi

import (

	"github.com/aabalke/guac/emu/nds/utils"
)

const (
	DEV_POWER = 0
	DEV_FIRMW = 1
	DEV_TOUCH = 2

    STAT_CONT = 1
    STAT_DONE = 2



    UNDEFINIED_DEV = 10
)

type Spi struct {
	CNT                uint16
	Device             uint8
	Hold, Irq, Enabled bool

    Pmd Pmd
    Firmware Firmware
    Tsc Tsc

    TransferDevice uint8


    Value uint8
    Req, Res []uint8
}

func (s *Spi) Init() {
    s.Pmd.RegPowermg = 0b1101
    s.Pmd.RegMicgain = 0b1

    s.TransferDevice = UNDEFINIED_DEV
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

        s.Hold = utils.BitEnabled(uint32(v), 3)

        s.Irq = utils.BitEnabled(uint32(v), 6)
        s.Enabled = utils.BitEnabled(uint32(v), 7)

	}
}

func (s *Spi) ReadCNT(b uint8) uint8 {
	return uint8(s.CNT >> (8 * b))
}

func (s *Spi) WriteData(v uint8) {

    if s.Enabled {

        if s.TransferDevice != s.Device {
            if s.Device == DEV_FIRMW {
                s.Firmware.Addr = 0
                s.Firmware.WriteBuffer = nil
            }

            if s.Req == nil {
                s.Req = make([]uint8, 16)
            }
            s.Req = s.Req[:0]
            s.Res = nil
        }

        s.TransferDevice = s.Device
    }

    var value uint8

    if len(s.Res) > 0 {
        value = s.Res[0]
        s.Res = s.Res[1:]
    }

    if len(s.Res) == 0 {
        var stat uint8
        s.Req = append(s.Req, v)

        switch s.Device {
        case DEV_POWER:
            s.Res, stat = s.Pmd.Transfer(s.Req)
        case DEV_FIRMW:
            s.Res, stat = s.Firmware.Transfer(s.Req)
        case DEV_TOUCH:
            s.Res, stat = s.Tsc.Transfer(s.Req)
        }

        if stat == STAT_DONE {
            s.Req = s.Req[:0]
        }
    }

    s.Value = value

    if !s.Hold {
        if s.Device == DEV_FIRMW {
            s.Firmware.Write()
        }

        s.TransferDevice = UNDEFINIED_DEV
    }
}

func (s *Spi) ReadData() uint8 {
    return s.Value
}
