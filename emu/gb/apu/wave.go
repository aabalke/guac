package apu

type WaveChannel struct {
	Apu *Apu
	Idx uint32

	CntH uint16
	WaveRam          [0x20]uint8

	samples, LengthTime float64

	WaveSamples, WavePosition uint8

    Period     uint16
    DACEnabled bool
    LenEnabled bool
    ChannelEnabled bool
}

func (ch *WaveChannel) Trigger() {

    if !ch.DACEnabled { 
        return
    }

    if ch.LengthTime <= 0 {
        ch.ResetLength(0, false)
    }

    ch.samples = 0
    ch.Reset()
    ch.ChannelEnabled = true
}

func (ch *WaveChannel) ResetLength(initLength uint8, doubleSpeed bool) {

    multipler := uint16(1)
    if doubleSpeed {
        multipler = 2
    }

    maxTimer := 256.0 * float64(multipler)
    divApuRate := float64(multipler) / 256.0

    if initLength == 0 {
        ch.LengthTime = (maxTimer + 1) * divApuRate
        return
    }

    ch.LengthTime = (maxTimer - float64(initLength)) * divApuRate
}

func (ch *WaveChannel) GetSample(doubleSpeed bool) int8 {

    if ch.LenEnabled && ch.LengthTime <= 0 && ch.ChannelEnabled {
        ch.ChannelEnabled = false
        return 0
    }

    if ch.LenEnabled {
        ch.LengthTime -= ch.Apu.sampleTime
        if ch.LengthTime <= 0 {

            ch.ChannelEnabled = false
            return 0
        }
    }

    if !ch.ChannelEnabled {
		return 0
	}

	freq := 2097152 / (2048 - float64(ch.Period))
	cycleSamples := float64(ch.Apu.sndFrequency) / freq

	ch.samples++
	if ch.samples >= cycleSamples {
		ch.samples -= cycleSamples

		ch.WaveSamples--
		if ch.WaveSamples != 0 {
			ch.WavePosition = (ch.WavePosition + 1) & 0x3F
		} else {
			ch.Reset()
		}
	}

	wavedata := ch.WaveRam[(uint32(ch.WavePosition)>>1)&0x1f]
	sample := (float64((wavedata>>((ch.WavePosition&1)<<2))&0xf) - 0x8) / 8

	if forceVolume := (ch.CntH >> 15) & 1 != 0; forceVolume {

		sample *= 0.75
	} else {
		switch vol := GetVarData(uint32(ch.CntH), 13, 14); vol {
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
