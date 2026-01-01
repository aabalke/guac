package snd

//import (
//	"github.com/aabalke/guac/emu/nds/utils"
//)

func (s *Snd) Write(addr uint32, v uint8) {

	addr &= 0xFFFF

	switch {
	case addr < 0x400:
		return
	case addr < 0x500:
		i := (addr & 0xF0) >> 4
		s.Channels[i].Write(addr, v)

	case addr < 0x600:

		c0 := &s.Capture[0]
		c1 := &s.Capture[1]

		switch addr {
		case 0x500:

			s.VolMaster = float64(v&0b111_1111) / 127

		case 0x501:

			s.LOut = (v & 0b11) >> 0
			s.ROut = (v & 0b11) >> 2
			s.NoOutCh1 = (v>>4)&1 != 0
			s.NoOutCh3 = (v>>5)&1 != 0
			s.Enabled = (v>>7)&1 != 0

		case 0x504:

			s.Bias &^= 0xFF
			s.Bias |= uint32(v)

		case 0x505:

			s.Bias &= 0xFF
			s.Bias |= uint32(v&0b11) << 8

		case 0x508:
			c0.Add = v&(1<<0) != 0
			c0.ChanSrc = v&(1<<1) != 0
			c0.OneShot = v&(1<<2) != 0
			c0.PCM8 = v&(1<<3) != 0
			busy := v&(1<<7) != 0

			c0.Start = busy

			if !busy {
				c0.Playing = false
			}

			if c0.Add {
				panic("UNSETUP SND CAP 0 ADD")
			}

		case 0x509:
			c1.Add = v&(1<<0) != 0
			c1.ChanSrc = v&(1<<1) != 0
			c1.OneShot = v&(1<<2) != 0
			c1.PCM8 = v&(1<<3) != 0
			busy := v&(1<<7) != 0

			c1.Start = busy

			if !busy {
				c1.Playing = false
			}

			if c1.Add {
				panic("UNSETUP SND CAP 1 ADD")
			}
		case 0x510:
			c0.Dest &^= 0xFF
			c0.Dest |= uint32(v &^ 0b11)
		case 0x511:
			c0.Dest &^= 0xFF << 8
			c0.Dest |= uint32(v) << 8
		case 0x512:
			c0.Dest &^= 0xFF << 16
			c0.Dest |= uint32(v) << 16
		case 0x513:
			c0.Dest &^= 0xFF << 24
			c0.Dest |= uint32(v&0b111) << 24
		case 0x514:
			c0.Len &^= 0xFF
			c0.Len |= uint16(v)
		case 0x515:
			c0.Len &^= 0xFF << 8
			c0.Len |= uint16(v) << 8
		case 0x518:
			c1.Dest &^= 0xFF
			c1.Dest |= uint32(v &^ 0b11)
		case 0x519:
			c1.Dest &^= 0xFF << 8
			c1.Dest |= uint32(v) << 8
		case 0x51A:
			c1.Dest &^= 0xFF << 16
			c1.Dest |= uint32(v) << 16
		case 0x51B:
			c1.Dest &^= 0xFF << 24
			c1.Dest |= uint32(v&0b111) << 24
		case 0x51C:
			c1.Len &^= 0xFF
			c1.Len |= uint16(v)
		case 0x51D:
			c1.Len &^= 0xFF << 8
			c1.Len |= uint16(v) << 8
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
		c.Hold = (v>>7)&1 != 0
	case 0x2:
		c.Panning = uint32(v & 0b111_1111)
	case 0x3:

		c.Duty = uint32(v & 0b111)
		c.RepeatMode = uint32(v>>3) & 0b11
		c.Format = uint32(v>>5) & 0b11
		busy := (v>>7)&1 != 0

		c.Start = busy

		if !busy {
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
		c.SrcAddr |= uint32(v&0b111) << 24

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
		c.SndLength |= uint32(v&0b11_1111) << 16
	}
}

func (s *Snd) Read(addr uint32) uint8 {

	addr &= 0xFFF

	if addr >= 0x500 {

		c0 := &s.Capture[0]
		c1 := &s.Capture[1]

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

		case 0x508:

			var v uint8

			if c0.Add {
				v |= (1 << 0)
			}
			if c0.ChanSrc {
				v |= (1 << 1)
			}
			if c0.OneShot {
				v |= (1 << 2)
			}
			if c0.PCM8 {
				v |= (1 << 3)
			}
			if c0.Playing {
				v |= (1 << 7)
			}

			return v

		case 0x509:

			var v uint8

			if c1.Add {
				v |= (1 << 0)
			}
			if c1.ChanSrc {
				v |= (1 << 1)
			}
			if c1.OneShot {
				v |= (1 << 2)
			}
			if c1.PCM8 {
				v |= (1 << 3)
			}
			if c1.Playing {
				v |= (1 << 7)
			}

			return v

		case 0x510:
			return uint8(c0.Dest)
		case 0x511:
			return uint8(c0.Dest >> 8)
		case 0x512:
			return uint8(c0.Dest >> 16)
		case 0x513:
			return uint8(c0.Dest >> 24)
		case 0x514:
			return uint8(c0.Len)
		case 0x515:
			return uint8(c0.Len >> 8)
		case 0x518:
			return uint8(c1.Dest)
		case 0x519:
			return uint8(c1.Dest >> 8)
		case 0x51A:
			return uint8(c1.Dest >> 16)
		case 0x51B:
			return uint8(c1.Dest >> 24)
		case 0x51C:
			return uint8(c1.Len)
		case 0x51D:
			return uint8(c1.Len >> 8)
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
