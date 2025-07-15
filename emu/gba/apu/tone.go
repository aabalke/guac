package apu

import (
	"github.com/aabalke33/guac/emu/gba/utils"
)

var dutyLookUp = [4]float64{0.125, 0.25, 0.5, 0.75}
var dutyLookUpi = [4]float64{0.875, 0.75, 0.5, 0.25}

const (
	PSG_MAX = 0x7f
	PSG_MIN = -0x80
)

type ToneChannel struct {
    Idx uint32
    CntL, CntH, CntX uint16

    phase bool
    samples, lengthTime, sweepTime, envTime float64
}

func (ch *ToneChannel) GetSample() int8 {

    if !ApuInstance.isSoundChanEnable(uint8(ch.Idx)) {
        return 0
    }

	//toneAddr := uint32(ch.CntH)
	freqHz := ch.CntX & 0b0111_1111_1111
	frequency := 131072 / float64(2048-freqHz)

	// Full length of the generated wave (if enabled) in seconds
	soundLen := ch.CntH & 0b0011_1111
	length := float64(64-soundLen) / 256

	// Envelope volume change interval in seconds
	envStep := ch.CntH >> 8 & 0b111
	envelopeInterval := float64(envStep) / 64

	cycleSamples := SND_FREQUENCY / frequency // Numbers of samples that a single cycle (wave phase change 1 -> 0) takes at output sample rate

	// Length reached check (if so, just disable the channel and return silence)

    if lenFlag := utils.BitEnabled(uint32(ch.CntX), 14); lenFlag {
		ch.lengthTime += SAMPLE_TIME
		if ch.lengthTime >= length {
            ApuInstance.enableSoundChan(int(ch.Idx), false)
			return 0
		}
	}

	// Frequency sweep (Square 1 channel only)
    if ch.Idx == 0 {
        sweepTime := (ch.CntL >> 4) & 0b111 // 0-7 (0=7.8ms, 7=54.7ms)
        sweepInterval := 0.0078 * float64(sweepTime+1)    // Frquency sweep change interval in seconds

        ch.sweepTime += SAMPLE_TIME
        if ch.sweepTime >= sweepInterval {
            ch.sweepTime -= sweepInterval

            // A Sweep Shift of 0 means that Sweep is disabled
            sweepShift := byte(ch.CntL & 0b111)

            if sweepShift != 0 {
                // X(t) = X(t-1) ± X(t-1)/2^n
                disp := freqHz >> sweepShift // X(t-1)/2^n
                if decrease := utils.BitEnabled(uint32(ch.CntL), 3); decrease {
                    freqHz -= disp
                } else {
                    freqHz += disp
                }

                if freqHz < 0x7ff {
                    // update frequency
                    cntx := (ch.CntX & ^uint16(0x7ff)) | uint16(freqHz)
                    ch.CntX = cntx

                } else {
                    ApuInstance.enableSoundChan(int(ch.Idx), false)
                }
            }
        }
    }

	// Envelope volume
	envelope := (ch.CntH >> 12) & 0xf
	if envStep > 0 {
		ch.envTime += SAMPLE_TIME

		if ch.envTime >= envelopeInterval {
			ch.envTime -= envelopeInterval

            if increment := utils.BitEnabled(uint32(ch.CntH), 11); increment {
				if envelope < 0xf {
					envelope++
				}
			} else {
				if envelope > 0 {
					envelope--
				}
			}

            ch.CntH = (ch.CntH & ^uint16(0xf000)) | (envelope << 12)
		}
	}

	// Phase change (when the wave goes from Low to High or High to Low, the Square Wave pattern)
	duty := (ch.CntH >> 6) & 0b11
	ch.samples++
	if ch.phase {
		// 1 -> 0 -_
		phaseChange := cycleSamples * dutyLookUp[duty]
		if ch.samples > phaseChange {
			ch.samples -= phaseChange
			ch.phase = false
		}
	} else {
		// 0 -> 1 _-
		phaseChange := cycleSamples * dutyLookUpi[duty]
		if ch.samples > phaseChange {
			ch.samples -= phaseChange
			ch.phase = true
		}
	}

	if ch.phase {
		return int8(float64(envelope) * PSG_MAX / 15)
	}
	return int8(float64(envelope) * PSG_MIN / 15)

}
