package apu

type WaveChannel struct {
	Apu *Apu
	Idx uint32

	Ram [0x10]uint8

	OutputLevel uint8

	WavePosition  uint8
	LengthCounter uint16
	Period        uint16
	ActivePeriod  uint16

	LastReadCycle uint32
	Sample        uint8
	SampleByte    uint8

	cyclesPerSample uint32
	accCycles       uint32

	DACEnabled     bool
	EnvEnabled     bool
	LenEnabled     bool
	ChannelEnabled bool
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

	// bank
	ch.WavePosition = 0
	ch.ChannelEnabled = true
	ch.ActivePeriod = ch.Period
	ch.accCycles = 0

	//fmt.Printf("Trigger. Active Period is %04d tcycles\n", (2048-ch.ActivePeriod)<<1)
	//debug.B[3] = true
}

func (ch *WaveChannel) clockLength() {

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

func (ch *WaveChannel) ResetLength(initLength uint8) {
	ch.LengthCounter = 256 - uint16(initLength)
}

// wave channel period divider is 1/2 cpu speed (2097152hz)
// relative to cpu cycles, clocked at CPU_CYCLE / (2 * (2048 - period))
func (ch *WaveChannel) ClockWave(tCycles, frameCycles uint32) {

	if !ch.ChannelEnabled {
		return
	}

	ch.cyclesPerSample = uint32(2048-ch.ActivePeriod) << 1
	ch.accCycles += tCycles

	for i := 0; ch.accCycles >= ch.cyclesPerSample; i++ {
		// need to set bank as well
		ch.accCycles -= ch.cyclesPerSample

		ch.WavePosition = (ch.WavePosition + 1) & 0x1F
		//ch.ReadLatch = ch.accCycles == 0

		// instead of read latch, have read latch cycle cnt

		ch.LastReadCycle = frameCycles - ch.accCycles

		//if debug.B[3] {
		//	fmt.Printf("Enabling Latch Acc Cycles %04d FrameCycle %08d lastRead %08d New Wave Position %02d BYTE VALUE %02X\n", ch.accCycles, frameCycles, ch.LastReadCycle, ch.WavePosition, ch.SampleByte)
		//}

		if ch.WavePosition&1 == 0 {
			ch.Sample = ch.SampleByte >> 4
		} else {
			ch.ActivePeriod = ch.Period
			ch.cyclesPerSample = uint32(2048-ch.ActivePeriod) << 1
			b := ch.Ram[ch.WavePosition>>1]
			ch.SampleByte = b
			ch.Sample = ch.SampleByte & 0xF
		}
	}
}

func (ch *WaveChannel) GetSample() int8 {

	// -8 changes the wave to be signed 0...15 to -8...7
	//vol := int8(ch.Buffer[ch.WavePosition & 0x1F]) - 8
	vol := int8(ch.Sample) - 8

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

//func (ch *WaveChannel) Reset() {
//
//	//if twoBanks := (ch.CntL >> 5) & 1 != 0; twoBanks {
//	//	ch.WavePosition = 0
//	//	ch.WaveSamples = 64
//	//	return
//	//}
//
//	//bankIdx := (ch.CntL >> 6) & 0b1
//	bankIdx := 0
//	ch.WavePosition = uint8(32 * bankIdx)
//	ch.WaveSamples = 32
//}
