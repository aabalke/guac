package snd

func (s *Snd) Write(addr uint32, v uint8) {
	addr &= 0xFFF

	if addr < 0x500 {
		s.Channels[(addr/16)&0xF].Write(addr, v)
		return
	}

	c0 := &s.Capture[0]
	c1 := &s.Capture[1]

	switch addr {
	case 0x500:
		s.VolMaster = float64(v&0x7F) / 127

	case 0x501:

		s.LOut = (v & 3) >> 0
		s.ROut = (v & 3) >> 2
		s.NoOutCh1 = (v>>4)&1 != 0
		s.NoOutCh3 = (v>>5)&1 != 0
		s.Enabled = v&0x80 != 0

	case 0x504, 0x505:
		offset := (addr & 1) * 8
		s.Bias = ((s.Bias &^ (0xFF << offset)) | (uint16(v) << offset))

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

	case 0x510, 0x511, 0x512, 0x513:
		offset := (addr & 3) * 8
		c0.Dest = ((c0.Dest &^ (0xFF << offset)) | (uint32(v) << offset)) & 0x07FFFFFC

	case 0x514, 0x515:
		offset := (addr & 1) * 8
		c0.Len = ((c0.Len &^ (0xFF << offset)) | (uint16(v) << offset))

	case 0x518, 0x519, 0x51A, 0x51B:
		offset := (addr & 3) * 8
		c1.Dest = ((c1.Dest &^ (0xFF << offset)) | (uint32(v) << offset)) & 0x07FFFFFC

	case 0x51C, 0x51D:
		offset := (addr & 1) * 8
		c1.Len = ((c1.Len &^ (0xFF << offset)) | (uint16(v) << offset))
	}
}

func (c *Channel) Write(addr uint32, v uint8) {
	switch addr & 0xF {
	case 0x0:
		c.VolMul = v & 0x7F
	case 0x1:
		c.VolDiv = v & 3
		c.Hold = v&0x80 != 0
	case 0x2:
		c.Panning = v & 0x7F
	case 0x3:

		c.Duty = v & 7
		c.RepeatMode = (v >> 3) & 3
		c.Format = (v >> 5) & 3
		c.Start = v&0x80 != 0

		if !c.Start {
			c.Playing = false
		}

	case 0x4, 0x5, 0x6, 0x7:
		offset := (addr & 3) * 8
		c.SrcAddr = ((c.SrcAddr &^ (0xFF << offset)) | (uint32(v) << offset)) & 0x07FFFFFC

	case 0x8, 0x9:
		offset := (addr & 1) * 8
		c.TimerValue = ((c.TimerValue &^ (0xFF << offset)) | (uint16(v) << offset))

	case 0xA, 0xB:
		offset := (addr & 1) * 8
		c.StartPosition = ((c.StartPosition &^ (0xFF << offset)) | (uint16(v) << offset))

	case 0xC, 0xD, 0xE:
		offset := (addr & 3) * 8
		c.SndLength = ((c.SndLength &^ (0xFF << offset)) | (uint32(v) << offset)) & 0x003FFFFF
	}
}

func (s *Snd) Read(addr uint32) uint8 {
	addr &= 0xFFF

	if addr < 0x500 {
		return s.Channels[(addr/16)&0xF].Read(addr)
	}

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

	case 0x504, 0x505:
		return uint8(s.Bias >> ((addr & 1) * 8))

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

	case 0x510, 0x511, 0x512, 0x513:
		return uint8(c0.Dest >> ((addr & 3) * 8))
	case 0x514, 0x515:
		return uint8(c0.Len >> ((addr & 1) * 8))
	case 0x518, 0x519, 0x51A, 0x51B:
		return uint8(c1.Dest >> ((addr & 3) * 8))
	case 0x51C, 0x51D:
		return uint8(c1.Len >> ((addr & 1) * 8))
	}
	return 0
}

func (c *Channel) Read(addr uint32) uint8 {
	switch addr & 0xF {
	case 0x0:
		return c.VolMul

	case 0x1:

		v := c.VolDiv

		if c.Hold {
			v |= 1 << 7
		}

		return v

	case 0x2:
		return c.Panning

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
