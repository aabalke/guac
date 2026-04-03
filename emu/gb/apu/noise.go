package apu

import (
	"math"
)

type NoiseChannel struct {
	Apu *Apu
	Idx uint32

	lfsr    uint16
	samples float64

	S, R   uint8
	Width7 bool

	LengthCounter uint8
	EnvTimer      uint8
	EnvVolume     uint8

	InitVolume   uint8
	EnvPace      uint8
	EnvIncrement bool

	DACEnabled     bool
	EnvEnabled     bool
	LenEnabled     bool
	ChannelEnabled bool
}

func (ch *NoiseChannel) LengthTrigger() {

	if ch.LengthCounter == 0 {
		return
	}

	if ch.Apu.fsStep&1 != 0 {
		ch.clockLength()
	}
}

func (ch *NoiseChannel) Trigger() {

	if ch.LengthCounter == 0 {
		ch.ResetLength(0)
		ch.LengthTrigger()
	}

	if !ch.DACEnabled {
		return
	}

	if ch.Width7 {
		ch.lfsr ^= 0x60
	} else {
		ch.lfsr ^= 0x6000
	}
	ch.samples = 0
	ch.ChannelEnabled = true
	ch.EnvTimer = ch.EnvPace
	ch.EnvVolume = ch.InitVolume
}

func (ch *NoiseChannel) clockLength() {

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

func (ch *NoiseChannel) ResetLength(initLength uint8) {
	ch.LengthCounter = 64 - initLength
}

func (ch *NoiseChannel) clockEnvelope() {

	if !ch.ChannelEnabled {
		return
	}

	if !ch.EnvEnabled {
		return
	}

	ch.EnvTimer--

	if ch.EnvTimer != 0 {
		return
	}

	ch.EnvTimer = ch.EnvPace
	if ch.EnvIncrement && ch.EnvVolume < 15 {
		ch.EnvVolume++
	} else if !ch.EnvIncrement && ch.EnvVolume > 0 {
		ch.EnvVolume--
	}
}

func (ch *NoiseChannel) GetSample(doubleSpeed bool) int8 {

	r := float64(ch.R)
	if r == 0 {
		r = 0.5
	}

	frequency := (524288 / r) / math.Pow(2, float64(ch.S)+1)
	cycleSamples := float64(ch.Apu.sndFrequency) / frequency

	carry := ch.lfsr&1 != 0
	ch.samples++
	if ch.samples >= cycleSamples {
		ch.samples -= cycleSamples
		ch.lfsr >>= 1

		if carry {
			if ch.Width7 {
				ch.lfsr ^= 0x60
			} else {
				ch.lfsr ^= 0x6000
			}
		}
	}

	vol := ch.InitVolume
	if ch.EnvEnabled {
		vol = ch.EnvVolume
	}

	vol <<= 3 // original range 0...15, need 0..127 for int8

	if carry {
		return int8(vol)
	}
	return -int8(vol)
}
