package apu

var DutyLookUp = [4]float64{0.125, 0.25, 0.5, 0.75}
var DutyLookUpi = [4]float64{0.875, 0.75, 0.5, 0.25}

const (
	PSG_MAX = 0x7f
	PSG_MIN = -0x80
)

type ToneChannel struct {
	Apu              *Apu
	Idx              uint32
	CntL uint16

	phase                                   bool
	samples, lengthTime, sweepTime, envTime float64

    Period uint16

    VolumeRegister uint8
    Duty           uint8 
    DACEnabled     bool
    LenEnabled     bool
    ChannelEnabled bool
}

func (ch *ToneChannel) Trigger() {

    if !ch.DACEnabled { 
        return
    }

    if ch.lengthTime <= 0 {
        ch.ResetLength(0, false)
    }

    ch.phase = false
    ch.samples = 0
    ch.sweepTime = 0
    ch.envTime = 0
    ch.ChannelEnabled = true
}

func (ch *ToneChannel) ResetLength(initLength uint8, doubleSpeed bool) {
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

func (ch *ToneChannel) GetSample(doubleSpeed bool) int8 {

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

	frequency := 131072 / float64(2048-ch.Period)
	cycleSamples := float64(ch.Apu.sndFrequency) / frequency

    if ch.Idx == 0 {
        ch.Sweep(ch.Period)
    }

	initVol := ch.VolumeRegister >> 4

	if envStep := ch.VolumeRegister & 7; envStep > 0 {

        envelopeInterval := float64(envStep) / float64(64)
		ch.envTime += ch.Apu.sampleTime

		if ch.envTime >= envelopeInterval {
			ch.envTime -= envelopeInterval

			if increment := (ch.VolumeRegister >> 3) & 1 != 0; increment {
				if initVol < 0xf {
					initVol++
				}
			} else {
				if initVol > 0 {
					initVol--
				}
			}
		}
	}

	ch.samples++
	if ch.phase {
		phaseChange := cycleSamples * DutyLookUp[ch.Duty]
		if ch.samples > phaseChange {
			ch.samples -= phaseChange
			ch.phase = false
		}
	} else {
		phaseChange := cycleSamples * DutyLookUpi[ch.Duty]
		if ch.samples > phaseChange {
			ch.samples -= phaseChange
			ch.phase = true
		}
	}

	if ch.phase {
		return int8(float64(initVol) * PSG_MAX / 15)
	}
	return int8(float64(initVol) * PSG_MIN / 15)
}

func (ch *ToneChannel) Sweep(period uint16) {
    sweepTime := (ch.CntL >> 4) & 7
    sweepInterval := 0.0078 * float64(sweepTime+1)

    ch.sweepTime += ch.Apu.sampleTime

    if ch.sweepTime < sweepInterval {
        return
    }

    ch.sweepTime -= sweepInterval

    sweepShift := ch.CntL & 7

    if sweepShift == 0 {
        return
    }

    // X(t) = X(t-1) ± X(t-1)/2^n
    disp := period >> sweepShift // X(t-1)/2^n
    if decrease := (ch.CntL >> 3) & 1 != 0; decrease {
        period -= disp
    } else {
        period += disp
    }

    if period >= 0x7FF {
        ch.ChannelEnabled = false
    }
}
