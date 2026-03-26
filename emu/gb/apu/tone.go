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

	phase bool
	samples float64

    LengthCounter uint8
    EnvTimer      uint8
    EnvVolume     uint8


    SweepPace uint8
    SweepDecrease bool
    SweepStep uint8
    SweepTimer uint8

    Period uint16

    InitVolume     uint8
    EnvPace      uint8
    EnvIncrement   bool

    Duty           uint8 

    DACEnabled     bool
    SweepEnabled   bool
    EnvEnabled     bool
    LenEnabled     bool
    ChannelEnabled bool
}

func (ch *ToneChannel) Trigger() {

    if !ch.DACEnabled { 
        return
    }

    if ch.LengthCounter <= 0 {
        //ch.ResetLength(0, false)
    }

    ch.phase = false
    ch.samples = 0

    ch.SweepTimer = ch.SweepPace
    ch.SweepEnabled = ch.SweepStep != 0 || ch.SweepPace != 0
    ch.EnvTimer = ch.EnvPace
    ch.EnvVolume = ch.InitVolume
    ch.ChannelEnabled = true

    ch.clockSweep()
}

func (ch *ToneChannel) clockLength() {
    //if ch.LenEnabled && ch.LengthCounter < 64 {
    //    ch.LengthCounter++
    //    if ch.LengthCounter >= 64 {
    //        ch.ChannelEnabled = false
    //    }
    //}
}

func (ch *ToneChannel) clockEnvelope() {

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

func (ch *ToneChannel) clockSweep() {

    if !ch.ChannelEnabled {
        return
    }

    if !ch.SweepEnabled {
        return
    }

    ch.SweepTimer--

    if ch.SweepTimer != 0 {
        return
    }

    ch.SweepTimer = ch.SweepPace

    // X(t) = X(t-1) ± X(t-1)/2^n
    disp := ch.Period >> ch.SweepStep // X(t-1)/2^n
    if ch.SweepDecrease {
        ch.Period -= disp
    } else {
        ch.Period += disp
    }

    if ch.Period >= 0x7FF {
        ch.Period = 0
        ch.ChannelEnabled = false
    }
}

func (ch *ToneChannel) GetSample(doubleSpeed bool) int8 {

    if !ch.ChannelEnabled {
		return 0
	}

	freq := 131072 / float64(2048-ch.Period)
	cycleSamples := float64(ch.Apu.sndFrequency) / freq

    ch.samples++
    if ch.phase {
        if ch.samples > cycleSamples*DutyLookUp[ch.Duty] {
            ch.samples -= cycleSamples * DutyLookUp[ch.Duty]
            ch.phase = false
        }
    } else {
        if ch.samples > cycleSamples*DutyLookUpi[ch.Duty] {
            ch.samples -= cycleSamples * DutyLookUpi[ch.Duty]
            ch.phase = true
        }
    }

    vol := uint8(ch.InitVolume)
    if ch.EnvEnabled {
        vol = ch.EnvVolume
    }

	if ch.phase {
		return int8(float64(vol) * PSG_MAX / 15)
	}
	return int8(float64(vol) * PSG_MIN / 15)
}
