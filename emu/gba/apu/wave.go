package apu

type WaveChannel struct {
	Apu *Apu
	Idx uint32

	Ram [0x20]uint8

	OutputLevel    uint8
	WavePosition   uint8
	LengthCounter  uint16
	Period         uint16
	ActivePeriod   uint16
	Sample         uint8
	SampleByte     uint8
	DACEnabled     bool
	EnvEnabled     bool
	LenEnabled     bool
	ChannelEnabled bool
	BankIdx        uint8
	DoubleSize     bool
}

func (ch *WaveChannel) LengthTrigger() {
	if ch.LengthCounter == 0 {
		return
	}

	if ch.Apu.fsStep&1 != 0 {
		ch.clockLength()
	}
}

func (ch *WaveChannel) Trigger() {
	if ch.LengthCounter == 0 {
		ch.ResetLength(0)
		ch.LengthTrigger()
	}

	if !ch.DACEnabled {
		return
	}

	// bank
	ch.WavePosition = 0 | (ch.BankIdx << 5)
	ch.ChannelEnabled = true
	ch.ActivePeriod = ch.Period
}

func (ch *WaveChannel) clockLength() {
	if !ch.LenEnabled {
		return
	}

	if ch.LengthCounter == 0 {
		return
	}

	ch.LengthCounter--

	if ch.LengthCounter != 0 {
		return
	}

	ch.ChannelEnabled = false
}

func (ch *WaveChannel) ResetLength(initLength uint8) {
	ch.LengthCounter = 256 - uint16(initLength)
}

func (ch *WaveChannel) GetSample() int8 {
	// -8 changes the wave to be signed 0...15 to -8...7
	// vol := int8(ch.Buffer[ch.WavePosition & 0x1F]) - 8
	vol := int8(ch.Sample) - 8

	switch ch.OutputLevel {
	case 0:
		//vol >>= 4
		vol = 0
	case 1:
		//vol >>= 0
	case 2:
		vol >>= 1
	case 3:
		vol >>= 2
	default:
		// 75 % on gba
		vol = (vol >> 2) + (vol >> 1)
	}

	vol <<= 3

	return vol
}

func (ch *WaveChannel) ClockRam() {
	if !ch.ChannelEnabled {
		return
	}

	// wave position has range 0...31 or 0...63 depending on double size

	if ch.DoubleSize {
		ch.WavePosition = ((ch.WavePosition + 1) & 0x3F)
	} else {
		ch.WavePosition = ((ch.WavePosition + 1) & 0x1F) | (ch.BankIdx << 5)
	}

	if ch.WavePosition&1 == 0 {
		ch.Sample = ch.SampleByte >> 4
	} else {
		ch.ActivePeriod = ch.Period
		b := ch.Ram[ch.WavePosition>>1]
		ch.SampleByte = b
		ch.Sample = ch.SampleByte & 0xF
	}
}

//func (ch *WaveChannel) Reset() {
//
//	//if twoBanks := (ch.CntL >> 5) & 1 != 0; twoBanks {
//	//	ch.WavePosition = 0
//	//	ch.WaveSamples = 64
//	//	return
//	//}
//
//	//bankIdx := (ch.CntL >> 6) & 0b1
//	bankIdx := 0
//	ch.WavePosition = uint8(32 * bankIdx)
//	ch.WaveSamples = 32
//}
