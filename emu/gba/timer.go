package gba

import (
	"github.com/aabalke/guac/emu/gba/utils"
)

type Timer struct {
	Gba               *GBA
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

func (gba *GBA) UpdateTimers(cycles uint32) {

	overflow := false

	if gba.Timers[0].Enabled {
		overflow = gba.Timers[0].Update(overflow, cycles)
	}
	if gba.Timers[1].Enabled {
		overflow = gba.Timers[1].Update(overflow, cycles)
	}
	if gba.Timers[2].Enabled {
		overflow = gba.Timers[2].Update(overflow, cycles)
	}
	if gba.Timers[3].Enabled {
		overflow = gba.Timers[3].Update(overflow, cycles)
	}
}

func (t *Timer) Update(overflow bool, cycles uint32) bool {

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
		return false
	}

	t.D = t.SavedInitialValue + (total & 0xFFFF)

	if aTick := (t.Gba.Mem.IO[0x83]>>2)&1 == uint8(t.Idx); aTick {

		fifo := &t.Gba.Apu.FifoA

		fifo.Load()

		if refill := fifo.Length <= 0x10; refill {
			t.Gba.Dma[1].transferFifo()
		}
	}

	if bTick := (t.Gba.Mem.IO[0x83]>>6)&1 == uint8(t.Idx); bTick {

		fifo := &t.Gba.Apu.FifoB

		fifo.Load()

		if refill := fifo.Length <= 0x10; refill {
			t.Gba.Dma[2].transferFifo()
		}
	}

	if t.OverflowIRQ {
		t.Gba.Irq.setIRQ(3 + uint32(t.Idx))
	}

	return true
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
