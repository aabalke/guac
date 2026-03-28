package apu

type WaveChannel struct {
	Apu *Apu
	Idx uint32

	CntH    uint16
	WaveRam [0x20]uint8

	samples float64

	WaveSamples, WavePosition uint8

    LengthCounter uint16

    Period uint16

    InitVolume   uint8
    EnvPace      uint8
    EnvIncrement bool

    DACEnabled     bool
    EnvEnabled     bool
    LenEnabled     bool
    ChannelEnabled bool
}

func (ch *WaveChannel) LengthTrigger() {

    if ch.LengthCounter == 0 {
        return
    }

    if ch.Apu.fsStep & 1 != 0 {
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

    ch.samples = 0
    ch.Reset()
    ch.ChannelEnabled = true
}

func (ch *WaveChannel) clockLength() {

    if !ch.LenEnabled {
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

func (ch *WaveChannel) GetSample(doubleSpeed bool) int8 {

    if !ch.ChannelEnabled {
		return 0
	}

	freq := 2097152 / float64(2048 - ch.Period)
	cycleSamples := float64(ch.Apu.sndFrequency) / freq

	ch.samples++
	if ch.samples >= cycleSamples {
		ch.samples -= cycleSamples

		ch.WaveSamples--
		if ch.WaveSamples > 0 {
			ch.WavePosition = (ch.WavePosition + 1) & 0x3F
		} else {
			ch.Reset()
		}
	}

	wavedata := ch.WaveRam[(uint32(ch.WavePosition)>>1)&0x1F]
	sample := (float64((wavedata>>((ch.WavePosition&1)<<2))&0xF) - 0x8) / 8

	if forceVolume := (ch.CntH >> 15) & 1 != 0; forceVolume {

		sample *= 0.75
	} else {
        switch vol := (ch.CntH >> 13) & 3; vol {
		case 0:
			sample = 0
		case 1:
		case 2:
			sample *= 0.5
		case 3:
			sample *= 0.25
		}
	}

	if sample >= 0 {
		return int8(sample / 7 * PSG_MAX)
	}
	return int8(sample / (-8) * PSG_MIN)
}

func (ch *WaveChannel) Reset() {

	//if twoBanks := (ch.CntL >> 5) & 1 != 0; twoBanks {
	//	ch.WavePosition = 0
	//	ch.WaveSamples = 64
	//	return
	//}

	//bankIdx := (ch.CntL >> 6) & 0b1
    bankIdx := 0
	ch.WavePosition = uint8(32 * bankIdx)
	ch.WaveSamples = 32
}
