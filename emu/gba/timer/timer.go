package timer

import (
	"github.com/aabalke/guac/emu/scheduler"
)

var freqShifts = [...]uint8{0, 6, 8, 10}

type Timer struct {
	Next            *Timer
	scheduler       *scheduler.Scheduler
	OnTimerOverflow func(t *Timer, late int64)
	Idx             int
	From            int64
	Counter         uint32
	Reload          uint16

	FreqShift uint8
	Cnt       uint8
	Irq       bool
	Cascade   bool
	Enabled   bool
	Running   bool

	IsArm9 bool

	events struct {
		Reload   scheduler.EventIdx
		ReloadHi scheduler.EventIdx
		ReloadLo scheduler.EventIdx
		Control  scheduler.EventIdx
		Overflow scheduler.EventIdx
	}
}

func NewTimer(sch *scheduler.Scheduler, OnTimerOverflow func(t *Timer, late int64), idx int) *Timer {
	t := &Timer{
		scheduler:       sch,
		OnTimerOverflow: OnTimerOverflow,
		Idx:             idx,
	}

	t.events = struct {
		Reload   scheduler.EventIdx
		ReloadHi scheduler.EventIdx
		ReloadLo scheduler.EventIdx
		Control  scheduler.EventIdx
		Overflow scheduler.EventIdx
	}{
		sch.Register(t.ReloadEvent, 1),
		sch.Register(t.ReloadEventHi, 1),
		sch.Register(t.ReloadEventLo, 1),
		sch.Register(t.ControlEvent, 2),
		sch.Register(t.Overflow, 0),
	}

	return t
}

func (t *Timer) Delta(late int64) uint32 {
	return uint32((t.scheduler.Now()-late)-t.From) >> t.FreqShift
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
	default:
		return 0
	}
}

func (t *Timer) Read16(idx int) uint16 {
	switch idx {
	case 0:
		return uint16(t.GetCounter())
	case 2:
		return uint16(t.Cnt)
	default:
		return 0
	}
}

// separate writes required for ngba-suite timer/reload.gba

func (t *Timer) Write(idx int, v uint8) {
	switch idx {
	case 0:
		t.scheduler.Schedule(t.events.ReloadLo, 1, v)
	case 1:
		t.scheduler.Schedule(t.events.ReloadHi, 1, v)
	case 2:
		t.scheduler.Schedule(t.events.Control, 1, v)
	}
}

func (t *Timer) Write16(idx uint32, v uint16) {
	if idx == 2 {
		t.scheduler.Schedule(t.events.Control, 1, uint8(v))
		return
	}

	t.scheduler.Schedule(t.events.Reload, 1, v)
}

func (t *Timer) Write32(v uint32) {
	t.scheduler.Schedule(t.events.Reload, 1, uint16(v))
	t.scheduler.Schedule(t.events.Control, 1, uint8(v>>16))
}

func (t *Timer) ReloadEvent(late int64, a any) {
	t.ReloadEventLo(late, uint8(a.(uint16)))
	t.ReloadEventHi(late, uint8(a.(uint16)>>8))
}

func (t *Timer) ReloadEventLo(_ int64, a any) {
	t.Reload = (t.Reload &^ 0xFF) | uint16(a.(uint8))
}

func (t *Timer) ReloadEventHi(_ int64, a any) {
	t.Reload = (t.Reload & 0xFF) | (uint16(a.(uint8)) << 8)
}

func (t *Timer) ControlEvent(late int64, a any) {
	v := a.(uint8)

	if t.Running {
		t.Stop(late)
	}

	if t.Idx == 0 {
		v &^= 0x4
	}

	prevEnabled := t.Cnt&0x80 != 0
	t.Cnt = v &^ 0b11_1000

	t.FreqShift = freqShifts[t.Cnt&3]
	t.Cascade = t.Cnt&0x4 != 0
	t.Irq = t.Cnt&0x40 != 0
	t.Enabled = t.Cnt&0x80 != 0

	if !t.Enabled {
		return
	}

	offset := (t.scheduler.Now() - late) & ((int64(1) << t.FreqShift) - 1)

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
}

func (t *Timer) Start(cycles int64) {
	t.Running = true
	t.From = t.scheduler.Now() - cycles
	until := int64((0x10000-t.Counter)<<t.FreqShift) - cycles
	t.scheduler.Schedule(t.events.Overflow, until, nil)
}

func (t *Timer) Stop(late int64) {
	t.Counter += t.Delta(late)
	if t.Counter >= 0x10000 {
		t.OverflowHandle(late)
	}

	t.scheduler.Cancel(t.events.Overflow)

	t.Running = false
}

func (t *Timer) Overflow(late int64, _ any) {
	t.OverflowHandle(late)
	t.Start(late)
}

func (t *Timer) OverflowHandle(late int64) {
	t.Counter = uint32(t.Reload)
	t.OnTimerOverflow(t, late)
}
