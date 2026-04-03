package apu

// source https://nightshade256.github.io/2021/03/27/gb-sound-emulation.html

var DutyLookUp = [4]float64{0.125, 0.25, 0.5, 0.75}
var DutyLookUpi = [4]float64{0.875, 0.75, 0.5, 0.25}

type ToneChannel struct {
	Apu *Apu
	Idx uint32

	phase       bool
	InFirstHalf bool
	Duty        uint8

	samples float64

	SweepPace     uint8
	SweepDecrease bool
	SweepStep     uint8
    SweepEnabled  bool
	SweepTimer    uint8
	Shadow uint16

    NegateLatch bool

	Period uint16

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

func (ch *ToneChannel) LengthTrigger() {

	if ch.LengthCounter == 0 {
		return
	}

	if ch.Apu.fsStep&1 != 0 {
		ch.clockLength()
	}
}

func (ch *ToneChannel) Trigger() {

	if ch.LengthCounter == 0 {
		ch.ResetLength(0)
		ch.LengthTrigger()
	}

	if !ch.DACEnabled {
		return
	}

	ch.phase = false
	ch.samples = 0

	ch.Shadow = ch.Period
    ch.SweepTimer = ch.SweepPace
    if ch.SweepTimer == 0 {
        ch.SweepTimer = 8
    }
	ch.SweepEnabled = ch.SweepStep != 0 || ch.SweepPace != 0

	ch.EnvTimer  = ch.EnvPace
	ch.EnvVolume = ch.InitVolume
	ch.ChannelEnabled = true
    ch.NegateLatch = false

	if ch.SweepStep != 0 {
		ch.calcFreq()
	}
}

func (ch *ToneChannel) clockSweep() {

    if ch.SweepTimer > 0 {
        ch.SweepTimer -= 1
    }

    if ch.SweepTimer != 0 {
        return
    }

    if ch.SweepPace != 0 {
        ch.SweepTimer = ch.SweepPace
    } else {
        ch.SweepTimer = 8
    }

    if !ch.SweepEnabled {
        return
    }

    if  ch.SweepPace == 0 {
        return
    }

    newPeriod := ch.calcFreq()
    if newPeriod <= 2047 && ch.SweepStep > 0 {
        ch.Period = newPeriod
        ch.Shadow = newPeriod
        ch.calcFreq()
    }
}

func (ch *ToneChannel) calcFreq() uint16 {

    newPeriod := ch.Shadow >> ch.SweepStep

    if ch.SweepDecrease {
        newPeriod = ch.Shadow - newPeriod
        ch.NegateLatch = true

    } else {
        newPeriod = ch.Shadow + newPeriod
    }

    if newPeriod > 2047 {
        ch.ChannelEnabled = false
    }

    return newPeriod
}

func (ch *ToneChannel) clockLength() {

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

func (ch *ToneChannel) ResetLength(initLength uint8) {
	ch.LengthCounter = 64 - initLength
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

func (ch *ToneChannel) GetSample(doubleSpeed bool) int8 {

	freq := 131072 / float64(2048-ch.Shadow)
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

	vol <<= 3 // original range 0...15, need 0..127 for int8

	if ch.phase {
		return int8(vol)
	}
	return -int8(vol)
}
