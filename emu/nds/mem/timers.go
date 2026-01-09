package mem

// these timers should run at 33 MHZ, GBA was just 16MHZ, not sure what I need to do

type Timer struct {
	IsArm9            bool
	Idx               int
	CNT, D            uint32
	SavedInitialValue uint32
	SavedCycles       uint32
	Elapsed           uint32

	Enabled     bool
	OverflowIRQ bool
	Cascade     bool
	Freq        uint32
	FreqShift   uint32
}

func (t *Timer) Update(overflow bool, cycles uint32) (bool, bool) {

	increment := uint32(0)
	if t.Cascade {
		if overflow {
			increment = 1
		}
	} else {

		t.Elapsed += cycles

		if t.Elapsed >= t.Freq {
			increment = t.Elapsed >> t.FreqShift
			t.Elapsed -= increment << t.FreqShift
		}
	}

    t.D += increment

    if notOverflow := t.D <= 0xFFFF; notOverflow {
        return false, false
    }

    t.D = (t.D & 0xFFFF) + t.SavedInitialValue
	return true, t.OverflowIRQ
}

func (t *Timer) ReadCnt(hi bool) uint8 {

	if hi {
		return uint8(t.CNT >> 8)
	}

	return uint8(t.CNT)
}

func (t *Timer) WriteCnt(v uint8) {

    wasEnabled := t.Enabled
	t.CNT = uint32(v) & 0xC7
	t.Cascade = (t.CNT>>2)&1 != 0
	t.OverflowIRQ = (t.CNT>>6)&1 != 0
	t.Enabled = (t.CNT>>7)&1 != 0
	t.Freq = t.getFreq()
	t.FreqShift = t.getFreqShift()

    if t.Enabled && !wasEnabled {
		t.D = t.SavedInitialValue
		t.Elapsed = 0
    }
}

func (t *Timer) ReadD(hi bool) uint8 {

	if hi {
		return uint8(t.D >> 8)
	}

	return uint8(t.D)
}

func (t *Timer) WriteD(v uint8, hi bool) {

	if hi {
		t.SavedInitialValue = (t.SavedInitialValue & 0xFF) | (uint32(v) << 8)
		return
	}

	t.SavedInitialValue = (t.SavedInitialValue & 0xFF00) | uint32(v)
}

func (t *Timer) getFreq() uint32 {
	switch freq := (t.CNT & 0b11); freq {
	case 0:
		return 1
	case 1:
		return 64
	case 2:
		return 256
	case 3:
		return 1024
	}

	return 1
}

func (t *Timer) getFreqShift() uint32 {

	switch freq := (t.CNT & 0b11); freq {
	case 0:
		return 0
	case 1:
		return 6
	case 2:
		return 8
	case 3:
		return 10
	}

	return 0
}
