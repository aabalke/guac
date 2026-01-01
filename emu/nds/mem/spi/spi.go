package spi

const (
	DEV_POWER = 0
	DEV_FIRMW = 1
	DEV_TOUCH = 2

	STAT_CONT = 1
	STAT_DONE = 2
)

type Spi struct {
	CNT                uint16
	Device             uint8
	Hold, Irq, Enabled bool

	Pmd      *Pmd
	Firmware Firmware
	Tsc      Tsc

	// pointer in order to nil when not used, not sure if better method
	TransferDevice *uint8

	Value    uint8
	Req, Res []uint8
}

func (s *Spi) Init() {

	s.Pmd = &Pmd{}
	s.Pmd.Init()

	s.TransferDevice = nil
	//FirmwareConfig()
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

		s.Hold = (v>>3)&1 != 0
		s.Irq = (v>>6)&1 != 0
		s.Enabled = (v>>7)&1 != 0
	}
}

func (s *Spi) ReadCNT(b uint8) uint8 {
	return uint8(s.CNT >> (8 * b))
}

func (s *Spi) WriteData(v uint8) {

	//fmt.Printf("SPI WRITE DATA % 02X\n", v)

	if s.Enabled {

		if s.TransferDevice == nil || *s.TransferDevice != s.Device {
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

		d := s.Device

		s.TransferDevice = &d
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

		s.TransferDevice = nil
	}
}

func (s *Spi) ReadData() uint8 {
	//fmt.Printf("READING SPI %02X\n", s.Value)
	return s.Value
}
