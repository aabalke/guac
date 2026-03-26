package apu

import (
	"math"
)

type NoiseChannel struct {
	Apu        *Apu
	Idx        uint32

	lfsr                         uint16
	samples, lengthTime, envTime float64

    RandomRegister uint8
    VolumeRegister uint8
    LenEnabled bool
    DACEnabled bool
    ChannelEnabled bool
}

func (ch *NoiseChannel) Trigger() {

    if !ch.DACEnabled { 
        return
    }

    if ch.lengthTime <= 0 {
        ch.ResetLength(0, false)
    }

    ch.lfsr = 0//
    ch.samples = 0
    ch.envTime = 0
    ch.ChannelEnabled = true
}

func (ch *NoiseChannel) ResetLength(initLength uint8, doubleSpeed bool) {
    multipler := uint16(1)
    if doubleSpeed {
        multipler = 2
    }
    maxTimer := 64.0 * float64(multipler)
    divApuRate := float64(multipler) / 256.0

    if initLength == 0 {
        ch.lengthTime = maxTimer * divApuRate
        return
    }

    ch.lengthTime = (maxTimer - float64(initLength)) * divApuRate
}

func (ch *NoiseChannel) GetSample(doubleSpeed bool) int8 {

    if ch.LenEnabled {
        ch.lengthTime -= ch.Apu.sampleTime
        if ch.lengthTime <= 0 {
            ch.ChannelEnabled = false
            return 0
        }
    }

    if !ch.ChannelEnabled {
		return 0
	}

	vol := ch.VolumeRegister >> 4

	if envStep := ch.VolumeRegister & 7; envStep > 0 {

        envelopeInterval := float64(envStep) / float64(64)
		ch.envTime += ch.Apu.sampleTime

		if ch.envTime >= envelopeInterval {
			ch.envTime -= envelopeInterval

			if increment := (ch.VolumeRegister >> 3) & 1 != 0; increment {
				if vol < 0xf {
					vol++
				}
			} else {
				if vol > 0 {
					vol--
				}
			}
		}
	}

    r := float64(ch.RandomRegister & 7)
    s := float64((ch.RandomRegister >>4) & 0xF)

	if r == 0 {
		r = 0.5
	}

	frequency := (524288 / r) / math.Pow(2, s+1)
	cycleSamples := float64(ch.Apu.sndFrequency) / frequency

	carry := byte(ch.lfsr & 1)
	ch.samples++
	if ch.samples >= cycleSamples {
		ch.samples -= cycleSamples
		ch.lfsr >>= 1

		if carry > 0 {
			if (ch.RandomRegister >> 3) & 1 != 0 {
				ch.lfsr ^= 0x60 // 1: 7bits
			} else {
				ch.lfsr ^= 0x6000 // 0: 15bits
			}
		}
	}

	if carry != 0 {
		return int8((float64(vol) / 15) * PSG_MAX) // Out=HIGH
	}
	return int8((float64(vol) / 15) * PSG_MIN) // Out=LOW
}
