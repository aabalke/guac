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
    InFirstHalf    bool
    Duty           uint8 

	samples float64

    SweepPace uint8
    SweepDecrease bool
    SweepStep uint8

    SweepStoredPace uint8

    SweepTimer uint8

    Period uint16
    Shadow uint16

    LengthCounter uint8
    EnvTimer      uint8
    EnvVolume     uint8

    InitVolume     uint8
    EnvPace      uint8
    EnvIncrement   bool

    DACEnabled     bool
    SweepEnabled   bool
    EnvEnabled     bool
    LenEnabled     bool
    ChannelEnabled bool
}

func (ch *ToneChannel) LengthTrigger() {

    if ch.LengthCounter == 0 {
        return
    }

    if ch.Apu.fsStep & 1 != 0 {
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

    ch.SweepTimer = ch.SweepPace
    ch.SweepStoredPace = ch.SweepPace
    ch.SweepEnabled = ch.SweepStep != 0 || ch.SweepPace != 0
    ch.EnvTimer = ch.EnvPace
    ch.EnvVolume = ch.InitVolume
    ch.ChannelEnabled = true
    ch.Shadow = ch.Period

    if ch.SweepStep > 0 && ch.SweepEnabled {

        ch.sweepCalculate(false)
    }
}

func (ch *ToneChannel) clockLength() {

    if !ch.LenEnabled {
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

func (ch *ToneChannel) sweepCalculate(writeback bool) {
    disp := ch.Shadow >> ch.SweepStep
    var newPeriod uint16
    if ch.SweepDecrease {
        newPeriod = ch.Shadow - disp
    } else {
        newPeriod = ch.Shadow + disp
    }
    if newPeriod > 0x7FF {
        ch.ChannelEnabled = false
        return
    }
    if writeback {

        ch.Shadow = newPeriod
        ch.Period = newPeriod

        disp2 := ch.Shadow >> ch.SweepStep
        var newPeriod2 uint16
        if ch.SweepDecrease {
            newPeriod2 = ch.Shadow - disp2
        } else {
            newPeriod2 = ch.Shadow + disp2
        }
        if newPeriod2 > 0x7FF {
            ch.ChannelEnabled = false
        }
    }
}

func (ch *ToneChannel) clockSweep() {
    //if !ch.ChannelEnabled {
    //    return
    //}
    if !ch.SweepEnabled {
        return
    }
    if ch.SweepPace == 0 {
        return
    }

    if ch.SweepTimer != 0 {
        ch.SweepTimer--
        if ch.SweepTimer != 0 {
            return
        }
    }

    ch.SweepStoredPace = ch.SweepPace
    ch.SweepTimer = ch.SweepStoredPace

    if ch.SweepTimer == 0 {
        ch.SweepTimer = 8
    }

    ch.sweepCalculate(ch.SweepStep != 0)
}

//func (ch *ToneChannel) clockSweep() {
//
//    if !ch.ChannelEnabled {
//        return
//    }
//
//    if !ch.SweepEnabled {
//        return
//    }
//
//    // X(t) = X(t-1) ± X(t-1)/2^n
//    disp := ch.Shadow >> ch.SweepStep
//    newPeriod := uint16(0)
//    if ch.SweepDecrease {
//        newPeriod = ch.Shadow - disp
//    } else {
//        newPeriod = ch.Shadow + disp
//    }
//
//    if debug.B[0] {
//        fmt.Printf("triggered. newPeriod %03X Shadow %03X PeriodReg %03X\n", newPeriod, ch.Shadow, ch.Period)
//    }
//
//    if ch.SweepStoredPace != 0 {
//        ch.Shadow = newPeriod
//        ch.Period = newPeriod
//    }
//
//    if newPeriod > 0x7FF {
//        if debug.B[0] { fmt.Printf("triggered. disabling call\n")}
//        ch.Shadow = 0
//        ch.Period = 0
//        ch.ChannelEnabled = false
//        ch.SweepEnabled = false
//    }
//}
//
//func (ch *ToneChannel) clockSweep2() {
//
//    if !ch.SweepEnabled {
//        return
//    }
//
//    if ch.SweepPace == 0 {
//        return
//    }
//
//    if ch.SweepTimer != 0 { // handled if pace init 0 but clocked on trigger
//        ch.SweepTimer--
//        if ch.SweepTimer != 0 {
//            return
//        }
//    }
//
//    ch.SweepTimer = ch.SweepStoredPace
//    if ch.SweepTimer == 0 {
//        fmt.Printf("HERE\n")
//        ch.SweepTimer = 8
//    }
//
//    if debug.B[0] {
//        fmt.Printf("clocked. timer=%d pace=%d storedpace=%d\n", ch.SweepTimer, ch.SweepPace, ch.SweepStoredPace)
//    }
//
//
//    // X(t) = X(t-1) ± X(t-1)/2^n
//    disp := ch.Shadow >> ch.SweepStep
//    newPeriod := uint16(0)
//    if ch.SweepDecrease {
//        newPeriod = ch.Shadow - disp
//    } else {
//        newPeriod = ch.Shadow + disp
//    }
//
//    if debug.B[0] {
//        fmt.Printf("clocked. newPeriod %03X Shadow %03X PeriodReg %03X\n", newPeriod, ch.Shadow, ch.Period)
//    }
//
//    if newPeriod > 0x7FF {
//        ch.ChannelEnabled = false
//        if debug.B[0] { fmt.Printf("clocked. disabling 1st call\n")}
//    } else {
//        if ch.SweepStep != 0 {
//            ch.Shadow = newPeriod
//            ch.Period = newPeriod
//
//            // X(t) = X(t-1) ± X(t-1)/2^n
//            disp := ch.Shadow >> ch.SweepStep
//            newPeriod := uint16(0)
//            if ch.SweepDecrease {
//                newPeriod = ch.Shadow - disp
//            } else {
//                newPeriod = ch.Shadow + disp
//            }
//
//            if newPeriod > 0x7FF {
//                ch.ChannelEnabled = false
//                if debug.B[0] { fmt.Printf("clocked. disabling 2nd call\n")}
//            }
//        }
//    }
//}

func (ch *ToneChannel) GetSample(doubleSpeed bool) int8 {

    if !ch.ChannelEnabled {
		return 0
	}

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
