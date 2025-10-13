package mem

import (
	"github.com/aabalke/guac/emu/gba/utils"
)

// these timers should run at 33 MHZ, GBA was just 16MHZ, not sure what I need to do

type Timer struct {
    IsArm9 bool
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

func (t *Timer) UpdateSingle(overflow bool) (bool, bool) {

    // this only works if you update 1 cycle per run

	increment := uint32(0)

    if !t.Cascade {
        t.Elapsed++

        if t.Elapsed >= t.Freq {
            increment = t.Elapsed >> t.FreqShift
            t.Elapsed -= increment << t.FreqShift
        }
    } else if overflow {
        increment = 1
    }

	total := t.D + increment

	if notOverflow := !(total > 0xFFFF); notOverflow {
		t.D = total
		return false, false
	}

	t.D = t.SavedInitialValue + (total & 0xFFFF)

	return true, t.OverflowIRQ
}

func (t *Timer) Update(overflow bool, cycles uint32) (bool, bool)  {

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
			//t.Elapsed -= increment * t.Freq // %= freq
		}
	}

	total := t.D + increment

	if notOverflow := !(total > 0xFFFF); notOverflow {
		t.D = total
		return false, false
	}

	t.D = t.SavedInitialValue + (total & 0xFFFF)

    // in nds function
	//if t.OverflowIRQ {
	//	t.Gba.Irq.setIRQ(3 + uint32(t.Idx))
	//}

	return true, t.OverflowIRQ
}

func (t *Timer) ReadCnt(hi bool) uint8 {

	if hi {
		return uint8(t.CNT >> 8)
	}

	return uint8(t.CNT)
}

func (t *Timer) WriteCnt(v uint8, hi bool) {

	if hi {
		return
	}

	oldValue := t.CNT & 0xC7
	t.CNT = uint32(v) & 0xC7
	t.Cascade = utils.BitEnabled(t.CNT, 2)
	t.OverflowIRQ = utils.BitEnabled(t.CNT, 6)
	t.Enabled = utils.BitEnabled(t.CNT, 7)
	t.Freq = t.getFreq()
	t.FreqShift = t.getFreqShift()

	if setEnabled := utils.BitEnabled(uint32(v), 7) && !utils.BitEnabled(oldValue, 7); setEnabled {
		//if setEnabled := utils.BitEnabled(uint32(v), 7); setEnabled {
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

	freq := utils.GetVarData(t.CNT, 0, 1)
	switch freq {
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

	freq := utils.GetVarData(t.CNT, 0, 1)
	switch freq {
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
