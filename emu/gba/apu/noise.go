package apu

import (
	"math"

	"github.com/aabalke33/guac/emu/gba/utils"
)

type NoiseChannel struct {
    Idx uint32
    CntL, CntH uint16

    lfsr uint16
    samples, lengthTime, envTime float64
}


func (ch *NoiseChannel) GetSample() int8 {

    //if disabled := !ApuInstance.isSoundChanEnable(uint8(ch.Idx)); disabled {
    //    return 0
    //}

	// Actual frequency in Hertz (524288 / r / 2^(s+1))
	r := float64(ch.CntH & 7)
    s := float64((ch.CntH>>4)&0xf)
	if r == 0 {
		r = 0.5
	}
	frequency := (524288 / r) / math.Pow(2, s+1)

	// Full length of the generated wave (if enabled) in seconds
	soundLen := ch.CntL & 0x3f
	length := (64 - float64(soundLen)) / 256

	// Length reached check (if so, just disable the channel and return silence)
	if utils.BitEnabled(uint32(ch.CntH), 14) {
		ch.lengthTime += SAMPLE_TIME
		if ch.lengthTime >= length {
            //ApuInstance.enableSoundChan(int(ch.Idx), false)
			return 0
		}
	}

	// Envelope volume change interval in seconds
	envStep := (ch.CntL >> 8) & 0x7
	envelopeInterval := float64(envStep) / 64

	// Envelope volume
	envelope := (ch.CntL >> 12) & 0xf
	if envStep != 0 {
		ch.envTime += SAMPLE_TIME
		if ch.envTime >= envelopeInterval {
			ch.envTime -= envelopeInterval

			if utils.BitEnabled(uint32(ch.CntL), 11) {
				if envelope < 0xf {
					envelope++
				}
			} else {
				if envelope > 0x0 {
					envelope--
				}
			}

			ch.CntL = (ch.CntL & ^uint16(0xf000)) | (envelope << 12)
		}
	}

	// Numbers of samples that a single cycle (pseudo-random noise value) takes at output sample rate
	cycleSamples := SND_FREQUENCY / frequency

	carry := byte(ch.lfsr & 0b1)
	ch.samples++
	if ch.samples >= cycleSamples {
		ch.samples -= cycleSamples
		ch.lfsr >>= 1

		if carry > 0 {
			if utils.BitEnabled(uint32(ch.CntH), 3) { // R/W Counter Step/Width
				ch.lfsr ^= 0x60 // 1: 7bits
			} else {
				ch.lfsr ^= 0x6000 // 0: 15bits
			}
		}
	}

	if carry != 0 {
		//return int8(float64(envelope) * PSG_MAX / 15)
		return int8((float64(envelope) / 15) * PSG_MAX) // Out=HIGH
	}
    //return int8(float64(envelope) * PSG_MIN / 15)
	return int8((float64(envelope) / 15) * PSG_MIN) // Out=LOW
}
