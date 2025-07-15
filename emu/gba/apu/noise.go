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

    if !ApuInstance.isSoundChanEnable(uint8(ch.Idx)) {
        return 0
    }

    soundLength := utils.GetVarData(uint32(ch.CntL), 0, 5)
	length := (64 - float64(soundLength)) / 256

    if stopAtLength := utils.BitEnabled(uint32(ch.CntH), 14); stopAtLength {
		ch.lengthTime += SAMPLE_TIME
        if stop := ch.lengthTime >= length; stop {
            ApuInstance.enableSoundChan(int(ch.Idx), false)
			return 0
		}
	}

    envStep := float64(utils.GetVarData(uint32(ch.CntL), 8, 10))
    envelope := uint16(utils.GetVarData(uint32(ch.CntL), 12, 15))

	if envStep != 0 {
		ch.envTime += SAMPLE_TIME
        envelopeInterval := envStep / 64

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

    r := float64(utils.GetVarData(uint32(ch.CntH), 0, 2))
    s := float64(utils.GetVarData(uint32(ch.CntH), 4, 7))

	if r == 0 {
		r = 0.5
	}

	frequency := (524288 / r) / math.Pow(2, s+1)
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
		return int8((float64(envelope) / 15) * PSG_MAX) // Out=HIGH
	}
	return int8((float64(envelope) / 15) * PSG_MIN) // Out=LOW
}
