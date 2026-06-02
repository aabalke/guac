package gba

type Event int

const (
	EVENT_VBK = iota
	EVENT_HBK
	EVENT_DRW
	EVENT_END_FRAME
	EVENT_END_SCANLINE
	EVENT_SND_SAMPLE_GEN
	EVENT_SND_WAVE_CLOCK
	EVENT_SND_FRAME_SEQ
)

type Scheduler struct {
	Events       [32]ScheduledEvent
	Cnt          int
	CurrentCycle int64
}

type ScheduledEvent struct {
	Event     Event
	InitCycle int64
}

func NewScheduler() *Scheduler {
	return &Scheduler{}
}

func (s *Scheduler) schedule(e Event, cyclesUntil int64) {
	s.scheduleAt(e, s.CurrentCycle+cyclesUntil)
}

func (s *Scheduler) scheduleAt(e Event, initCycle int64) {
	if s.Cnt >= 32 {
		panic("gb/gbc: scheduler reached hard limit")
	}

	es := ScheduledEvent{Event: e, InitCycle: initCycle}
	for i := range s.Cnt {
		if es.InitCycle < s.Events[i].InitCycle {
			copy(s.Events[i+1:s.Cnt+1], s.Events[i:s.Cnt])
			s.Events[i] = es
			s.Cnt++
			return
		}
	}
	s.Events[s.Cnt] = es
	s.Cnt++
}

func (s *Scheduler) popNext() ScheduledEvent {
	next := s.Events[0]
	copy(s.Events[0:s.Cnt-1], s.Events[1:s.Cnt])
	s.Cnt--
	return next
}

func (s *Scheduler) endFrame() {
	s.CurrentCycle -= CYCLES_FRAME
	for i := range s.Cnt {
		s.Events[i].InitCycle -= CYCLES_FRAME
	}
}

func (s *Scheduler) cancel(e Event) {
	for i := range s.Cnt {
		if s.Events[i].Event == e {
			copy(s.Events[i:s.Cnt-1], s.Events[i+1:s.Cnt])
			s.Cnt--
			return
		}
	}
}

func (s *Scheduler) penalize(e Event, cycles int64) {
	for i := range s.Cnt {
		if s.Events[i].Event == e {
			s.Events[i].InitCycle += cycles
			return
		}
	}
}
