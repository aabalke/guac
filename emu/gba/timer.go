package gba

var freqShifts = [...]uint32{0, 6, 8, 10}

type Timer struct {
	Gba       *GBA
	Idx       int
	Cnt       uint8
	Counter   uint16
	Reload    uint16
	Enabled   bool
	FreqShift uint32
	InitCycle int64
}

func NewTimer(gba *GBA, idx int) *Timer {
	return &Timer{
		Gba: gba,
		Idx: idx,
	}
}

func (t *Timer) Delta(late int64) uint32 {
	return uint32((t.Gba.Scheduler.CurrentCycle-late)-t.InitCycle) >> t.FreqShift
}

func (t *Timer) GetCounter() uint32 {
	if t.Enabled {
		return uint32(t.Counter) + t.Delta(0)
	}

	return uint32(t.Counter)
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

func (t *Timer) Write(idx int, v uint8) {
	switch idx {
	case 0:
		t.Gba.Scheduler.schedule(EVENT_TIMER_RELOAD, 1, t.ReloadEventLo, v)
		//t.Reload = (t.Reload &^ 0xFF) | uint16(v)
	case 1:
		t.Gba.Scheduler.schedule(EVENT_TIMER_RELOAD, 1, t.ReloadEventHi, v)
		//t.Reload = (t.Reload & 0xFF) | (uint16(v) << 8)
	case 2:
		t.Gba.Scheduler.schedule(EVENT_TIMER_CONTROL, 1, t.ControlEvent, v)
		//prevEnabled := t.Enabled
		//t.Cnt = v
		//t.Enabled = t.Cnt&0x80 != 0
		//t.FreqShift = freqShifts[t.Cnt&3]

		//if t.Enabled && !prevEnabled {
		//	t.Counter = t.Reload
		//	t.InitCycle = t.Gba.Scheduler.CurrentCycle
		//}
	}
}

func (t *Timer) ReloadEventLo(late int64, argz any) bool {
	t.Reload = (t.Reload &^ 0xFF) | uint16(argz.(uint8))
	return false
}

func (t *Timer) ReloadEventHi(late int64, argz any) bool {
	t.Reload = (t.Reload & 0xFF) | (uint16(argz.(uint8)) << 8)
	return false
}

func (t *Timer) ControlEvent(late int64, argz any) bool {
	v := argz.(uint8)
	prevEnabled := t.Enabled
	t.Cnt = v
	t.Enabled = t.Cnt&0x80 != 0
	t.FreqShift = freqShifts[t.Cnt&3]

	if t.Enabled && !prevEnabled {
		t.Counter = t.Reload
		t.InitCycle = t.Gba.Scheduler.CurrentCycle - late
	}
	return false
}
