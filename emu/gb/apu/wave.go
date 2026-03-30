package apu

type WaveChannel struct {
	Apu *Apu
	Idx uint32

    Ram     [0x20]uint8
    Buffer  [0x40]uint8 // this is waveram but ind nibbles up

	OutputLevel uint8

	samples float64

	WaveSamples, WavePosition uint8
	LengthCounter             uint16
	Period                    uint16

	DACEnabled     bool
	EnvEnabled     bool
	LenEnabled     bool
	ChannelEnabled bool
}

func (ch *WaveChannel) UpdateCachedRam(addr uint32, v uint8) {
    addr -= 0x30
    addr <<= 1
    ch.Buffer[addr+0] = v & 0xF
    ch.Buffer[addr+1] = (v >> 4)
}

func (ch *WaveChannel) LengthTrigger() {

	if ch.LengthCounter == 0 {
		return
	}

	if ch.Apu.fsStep&1 != 0 {
		ch.clockLength()
	}
}

func (ch *WaveChannel) Trigger() {

	if ch.LengthCounter == 0 {
		ch.ResetLength(0)
		ch.LengthTrigger()
	}

	if !ch.DACEnabled {
		return
	}

	ch.samples = 0
	ch.Reset()
	ch.ChannelEnabled = true
}

func (ch *WaveChannel) clockLength() {

	if !ch.LenEnabled {
		return
	}

	ch.LengthCounter--

	if ch.LengthCounter != 0 {
		return
	}

	ch.ChannelEnabled = false
}

func (ch *WaveChannel) ResetLength(initLength uint8) {
	ch.LengthCounter = 256 - uint16(initLength)
}

func (ch *WaveChannel) GetSample(doubleSpeed bool) int8 {

	if !ch.ChannelEnabled {
		return 0
	}

	freq := 2097152 / float64(2048-ch.Period)
	cycleSamples := float64(ch.Apu.sndFrequency) / freq

	ch.samples++
	if ch.samples >= cycleSamples {
		ch.samples -= cycleSamples

		ch.WaveSamples--
		if ch.WaveSamples > 0 {
			ch.WavePosition++
		} else {
			ch.Reset()
		}
	}

    // -8 changes the wave to be signed 0...15 to -8...7
    vol := int8(ch.Buffer[ch.WavePosition & 0x1F]) - 8

	switch ch.OutputLevel {
	case 0:
		//vol >>= 4
        vol = 0
	case 1:
        //vol >>= 0
	case 2:
		vol >>= 1
	case 3:
		vol >>= 2
	}

    vol <<= 3

	return vol
}

func (ch *WaveChannel) Reset() {

	//if twoBanks := (ch.CntL >> 5) & 1 != 0; twoBanks {
	//	ch.WavePosition = 0
	//	ch.WaveSamples = 64
	//	return
	//}

	//bankIdx := (ch.CntL >> 6) & 0b1
	bankIdx := 0
	ch.WavePosition = uint8(32 * bankIdx)
	ch.WaveSamples = 32
}
