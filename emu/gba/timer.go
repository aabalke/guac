package gba

var freqShifts = [...]uint8{0, 6, 8, 10}

type Timer struct {
	Gba     *GBA
	Idx     int
	From    int64
	Counter uint32
	Reload  uint16

	FreqShift uint8
	Cnt       uint8
	Irq       bool
	Cascade   bool
	Enabled   bool
	Running   bool
}

func NewTimer(gba *GBA, idx int) *Timer {
	return &Timer{
		Gba: gba,
		Idx: idx,
	}
}

func (t *Timer) Delta(late int64) uint32 {
	return uint32((t.Gba.Scheduler.Now()-late)-t.From) >> t.FreqShift
}

func (t *Timer) GetCounter() uint32 {
	counter := t.Counter
	if t.Running {
		counter += t.Delta(0)
	}

	return counter
}

func (t *Timer) Read(idx int) uint8 {
	switch idx {
	case 0:
		return uint8(t.GetCounter())
	case 1:
		return uint8(t.GetCounter() >> 8)
	case 2:
		return t.Cnt
	}

	return 0
}

// separate writes required for ngba-suite timer/reload.gba

func (t *Timer) Write(idx int, v uint8) {
	switch idx {
	case 0:
		t.Gba.Scheduler.schedule(EVENT_TIMER_RELOAD, 1, 1, t.ReloadEventLo, v)
	case 1:
		t.Gba.Scheduler.schedule(EVENT_TIMER_RELOAD, 1, 1, t.ReloadEventHi, v)
	case 2:
		t.Gba.Scheduler.schedule(EVENT_TIMER_CONTROL, 2, 1, t.ControlEvent, v)
	}
}

func (t *Timer) Write16(v uint16) {
	// write16 Control is identical to Write8

	t.Gba.Scheduler.schedule(EVENT_TIMER_RELOAD, 1, 1, t.WriteEvent16, v)
}

func (t *Timer) WriteEvent16(late int64, argz any) bool {
	v := argz.(uint16)

	t.ReloadEventLo(late, uint8(v))
	t.ReloadEventHi(late, uint8(v>>8))
	return false
}

func (t *Timer) Write32(v uint32) {
	t.Gba.Scheduler.schedule(EVENT_TIMER_CONTROL, 1, 1, t.WriteEvent32, v)
}

func (t *Timer) WriteEvent32(late int64, argz any) bool {
	v := argz.(uint32)

	t.ReloadEventLo(late, uint8(v))
	t.ReloadEventHi(late, uint8(v>>8))
	t.ControlEvent(late, uint8(v>>16))
	return false
}

func (t *Timer) ReloadEventLo(_ int64, argz any) bool {
	t.Reload = (t.Reload &^ 0xFF) | uint16(argz.(uint8))
	return false
}

func (t *Timer) ReloadEventHi(_ int64, argz any) bool {
	t.Reload = (t.Reload & 0xFF) | (uint16(argz.(uint8)) << 8)
	return false
}

func (t *Timer) ControlEvent(late int64, argz any) bool {
	v := argz.(uint8)

	if t.Running {
		t.Stop(late)
	}

	if t.Idx == 0 {
		v &^= 0x4
	}

	prevEnabled := t.Cnt&0x80 != 0
	t.Cnt = v

	t.FreqShift = freqShifts[t.Cnt&3]
	t.Cascade = t.Cnt&0x4 != 0
	t.Irq = t.Cnt&0x40 != 0
	t.Enabled = t.Cnt&0x80 != 0

	if !t.Enabled {
		return false
	}

	offset := (t.Gba.Scheduler.Now() - late) & ((int64(1) << t.FreqShift) - 1)

	if prevEnabled {
		if t.Cnt&0x4 == 0 {
			t.Start(offset + late)
		}
	} else {
		switch {
		case t.Cnt&0x4 != 0:
			t.Counter = uint32(t.Reload)
		case t.Counter == 0xFFFF && offset == 0:
			t.Start(late)
		default:
			t.Counter = uint32(t.Reload)
			t.Start(offset + late - 1)
		}
	}
	return false
}

func (t *Timer) Start(cycles int64) {
	t.Running = true
	t.From = t.Gba.Scheduler.Now() - cycles
	until := int64((0x10000-t.Counter)<<t.FreqShift) - cycles

	switch t.Idx {
	case 0:
		t.Gba.Scheduler.schedule(EVENT_TIMER_OVERFLOW0, 0, until, t.Overflow, nil)
	case 1:
		t.Gba.Scheduler.schedule(EVENT_TIMER_OVERFLOW1, 0, until, t.Overflow, nil)
	case 2:
		t.Gba.Scheduler.schedule(EVENT_TIMER_OVERFLOW2, 0, until, t.Overflow, nil)
	case 3:
		t.Gba.Scheduler.schedule(EVENT_TIMER_OVERFLOW3, 0, until, t.Overflow, nil)

	}
}

func (t *Timer) Stop(late int64) {
	t.Counter += t.Delta(late)
	if t.Counter >= 0x10000 {
		t._Overflow(late)
	}

	switch t.Idx {
	case 0:
		t.Gba.Scheduler.cancel(EVENT_TIMER_OVERFLOW0)
	case 1:
		t.Gba.Scheduler.cancel(EVENT_TIMER_OVERFLOW1)
	case 2:
		t.Gba.Scheduler.cancel(EVENT_TIMER_OVERFLOW2)
	case 3:
		t.Gba.Scheduler.cancel(EVENT_TIMER_OVERFLOW3)
	}

	t.Running = false
}

func (t *Timer) Overflow(late int64, _ any) bool {
	t._Overflow(late)
	t.Start(late)
	return false
}

func (t *Timer) _Overflow(late int64) {
	t.Counter = uint32(t.Reload)
	t.OnTimerOverflow(late)
}

func (t *Timer) OnTimerOverflow(late int64) {
	if t.Irq {
		t.Gba.Irq.SetIRQ(3 + uint32(t.Idx))
	}

	//if t.Idx < 2 {
	//	for ch := uint8(0); ch < 2; ch++ { // 0: PCM-A, 1: PCM-B
	//		if c.apu.StepPCM(ch, t.Idx) {
	//			c.RequestSoundDMA(ch + 1) // PCM-A: DMA1, PCM-B: DMA2
	//		}
	//	}
	//}

	if t.Idx != 3 {
		if next := t.Gba.Timers[t.Idx+1]; next.Enabled && next.Cascade {
			next.Counter++
			if next.Counter >= 0x10000 {
				next._Overflow(late)
			}
		}
	}
}
