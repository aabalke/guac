package apu

import (
	"math"

)

type NoiseChannel struct {
    Apu *Apu
    Idx uint32
    CntL, CntH uint16

    lfsr uint16
    samples, lengthTime, envTime float64
}

func (ch *NoiseChannel) GetSample(doubleSpeed bool) int8 {

    if !ch.Apu.isSoundChanEnable(uint8(ch.Idx)) {
        return 0
    }

    multipler := uint16(1)
    if doubleSpeed {
        multipler = 2
    }
    maxTimer := 64.0 * float64(multipler)
    divApuRate := float64(multipler) / 256.0

    soundLength := GetVarData(uint32(ch.CntL), 0, 5)
    length := (maxTimer - float64(soundLength)) * divApuRate

    if stopAtLength := BitEnabled(uint32(ch.CntH), 14); stopAtLength {

		ch.lengthTime += ch.Apu.sampleTime

        if stop := ch.lengthTime >= length; stop {
            ch.Apu.enableSoundChan(int(ch.Idx), false)
			return 0
		}
	}

    envStep := float64(GetVarData(uint32(ch.CntL), 8, 10))
    envelope := uint16(GetVarData(uint32(ch.CntL), 12, 15))

	if envStep != 0 {
		ch.envTime += ch.Apu.sampleTime
        envelopeInterval := envStep / 64

		if ch.envTime >= envelopeInterval {
			ch.envTime -= envelopeInterval


			if BitEnabled(uint32(ch.CntL), 11) {
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

    r := float64(GetVarData(uint32(ch.CntH), 0, 2))
    s := float64(GetVarData(uint32(ch.CntH), 4, 7))

	if r == 0 {
		r = 0.5
	}

	frequency := (524288 / r) / math.Pow(2, s+1)
	cycleSamples := float64(ch.Apu.sndFrequency) / frequency

	carry := byte(ch.lfsr & 0b1)
	ch.samples++
	if ch.samples >= cycleSamples {
		ch.samples -= cycleSamples
		ch.lfsr >>= 1

		if carry > 0 {
			if BitEnabled(uint32(ch.CntH), 3) { // R/W Counter Step/Width
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
