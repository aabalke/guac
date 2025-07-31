package gba

import (
	"github.com/aabalke/guac/emu/gba/utils"
)

type Timers [4]Timer

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
}

func (tt *Timers) Update(cycles uint32) {

	overflow := false

	for i := range tt {
		t := &tt[i]

		if t.Enabled {
			overflow = t.Update(overflow, cycles)
		}
	}
}

func (t *Timer) Update(overflow bool, cycles uint32) bool {

	increment := uint32(0)
	if t.Cascade && overflow {
		increment = 1
	}

	if !t.Cascade {

		t.Elapsed += cycles

		if freq := t.Freq; t.Elapsed >= freq {
            increment = t.Elapsed / freq
            t.Elapsed = t.Elapsed - increment * freq // %= freq
		}
	}

	overflow = incrementTimer(t, increment)

	if !overflow {
		return false
	}

	if aTick := (t.Gba.Mem.IO[0x83]>>2)&1 == uint8(t.Idx); aTick {

		fifo := t.Gba.Apu.FifoA

		fifo.Load()

		if refill := fifo.Length <= 0x10; refill {
			t.Gba.Dma[1].transferFifo()
		}
	}

	if bTick := (t.Gba.Mem.IO[0x83]>>6)&1 == uint8(t.Idx); bTick {

		fifo := t.Gba.Apu.FifoB

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

func incrementTimer(t *Timer, increment uint32) bool {

	overflow := false

	if t.D+increment < 0xFFFF {
		t.D += increment
		return overflow
	}

	for range increment {
		tmp := t.D + 1
		if tmp > 0xFFFF {
			t.D = t.SavedInitialValue
			overflow = true
			continue
		}
		t.D = tmp
	}

	return overflow
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
	t.Freq = t.getCycles()

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

func (t *Timer) getCycles() uint32 {

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
