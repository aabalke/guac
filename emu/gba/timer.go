package gba

var (
	freqs      = [...]uint32{1, 64, 256, 1024}
	freqShifts = [...]uint32{0, 6, 8, 10}
)

type Timer struct {
	Gba               *GBA
	Idx               int
	Cnt               uint8
	D                 uint32
	SavedInitialValue uint32
	SavedCycles       uint32
	Elapsed           uint32

	Enabled     bool
	OverflowIRQ bool
	Cascade     bool
	Freq        uint32
	FreqShift   uint32
}

func NewTimer(gba *GBA, idx int) *Timer {
	return &Timer{
		Gba: gba,
		Idx: idx,
	}
}

func (gba *GBA) UpdateTimers(cycles uint32) {
	overflow := false
	for i := range 4 {
		if gba.Timers[i].Enabled {
			overflow = gba.Timers[i].Update(overflow, cycles)
		}
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

	if notOverflow := total <= 0xFFFF; notOverflow {
		t.D = total
		return false
	}

	//t.D = t.SavedInitialValue + (total & 0xFFFF)

	// if aTick := (t.Gba.Mem.IO[0x83]>>2)&1 == uint8(t.Idx); aTick {

	//	fifo := &t.Gba.Apu.FifoA

	//	fifo.Load()

	//	if refill := fifo.Length <= 0x10; refill {
	//		t.Gba.Dma[1].transferFifo()
	//	}
	//}

	// if bTick := (t.Gba.Mem.IO[0x83]>>6)&1 == uint8(t.Idx); bTick {

	//	fifo := &t.Gba.Apu.FifoB

	//	fifo.Load()

	//	if refill := fifo.Length <= 0x10; refill {
	//		t.Gba.Dma[2].transferFifo()
	//	}
	//}

	//if t.OverflowIRQ {
	//	println("irq")
	//	t.Gba.Irq.SetIRQ(3 + uint32(t.Idx))
	//}

	return true
}

func (t *Timer) ReadCnt() uint8 {
	return t.Cnt
}

func (t *Timer) WriteCnt(v uint8) {
	oldValue := t.Cnt & 0xC7
	t.Cnt = v & 0xC7
	t.Cascade = (t.Cnt>>2)&1 != 0
	t.OverflowIRQ = (t.Cnt>>6)&1 != 0
	t.Enabled = (t.Cnt>>7)&1 != 0
	t.Freq = freqs[t.Cnt&3]
	t.FreqShift = freqShifts[t.Cnt&3]

	if setEnabled := (v>>7)&1 != 0 && (oldValue>>7) == 0; setEnabled {
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
