package apu

import "github.com/aabalke33/guac/emu/gba/utils"

type WaveChannel struct {
    Apu *Apu
	Idx uint32

	CntL, CntH, CntX uint16
	WaveRam          [0x20]uint8

	samples, lengthTime float64

    WaveSamples, WavePosition uint8
}

func (ch *WaveChannel) GetSample(doubleSpeed bool) int8 {

    if !ch.Apu.isSoundChanEnable(uint8(ch.Idx)) {
        return 0
    }

    if !utils.BitEnabled(uint32(ch.CntL), 7) {
        return 0
    }

    multipler := uint16(1)
    if doubleSpeed {
        multipler = 2
    }
    maxTimer := 256.0 * float64(multipler)
    divApuRate := float64(multipler) / 256.0

    soundLength := utils.GetVarData(uint32(ch.CntH), 0, 7)
	length := (maxTimer - float64(soundLength)) / divApuRate

    if stopAtLength := utils.BitEnabled(uint32(ch.CntX), 14); stopAtLength {
		ch.lengthTime += ch.Apu.sampleTime
        if stop := ch.lengthTime >= length; stop {
            ch.Apu.enableSoundChan(int(ch.Idx), false)
			return 0
		}
	}

    rate := utils.GetVarData(uint32(ch.CntX), 0, 10)
	freq := 2097152 / (2048 - float64(rate))
	cycleSamples := float64(ch.Apu.sndFrequency) / freq

	ch.samples++
	if ch.samples >= cycleSamples {
		ch.samples -= cycleSamples

		ch.WaveSamples--
		if ch.WaveSamples != 0 {
			ch.WavePosition = (ch.WavePosition + 1) & 0b0011_1111
		} else {
            ch.Reset()
		}
	}

	wavedata := ch.WaveRam[(uint32(ch.WavePosition)>>1)&0x1f]
	sample := (float64((wavedata>>((ch.WavePosition&1)<<2))&0xf) - 0x8) / 8

    if forceVolume := utils.BitEnabled(uint32(ch.CntH), 15); forceVolume {
        sample *= 0.75
    } else {
        switch vol := utils.GetVarData(uint32(ch.CntH), 13, 14); vol {
        case 0: sample = 0
        case 1:
        case 2: sample *= 0.5
        case 3: sample *= 0.25
        }
    }

	if sample >= 0 {
		return int8(sample / 7 * PSG_MAX)
	}
	return int8(sample / (-8) * PSG_MIN)
}

func (ch *WaveChannel) Reset() {

    if twoBanks := utils.BitEnabled(uint32(ch.CntL), 5); twoBanks {
		ch.WavePosition = 0
        ch.WaveSamples = 64
		return
	}

    bankIdx := (ch.CntL >> 6) & 0b1
    ch.WavePosition = uint8(32 * bankIdx)
    ch.WaveSamples = 32
}
