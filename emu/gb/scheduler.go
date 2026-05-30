package gb

type Event int

const (
	EVENT_VBK = iota
	EVENT_HBK
	EVENT_DRW
	EVENT_END_FRAME
	EVENT_END_SCANLINE
	EVENT_SND_SAMPLE_GEN
)

type Scheduler struct {
	Events       []ScheduledEvent
	CurrentCycle int64
}

type ScheduledEvent struct {
	Event     Event
	InitCycle int64
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		Events: make([]ScheduledEvent, 0, 32),
	}
}

func (s *Scheduler) schedule(e Event, cyclesUntil int64) {
	s.scheduleAt(e, s.CurrentCycle+cyclesUntil)
}

func (s *Scheduler) scheduleAt(e Event, initCycle int64) {
	es := ScheduledEvent{Event: e, InitCycle: initCycle}

	for i, existing := range s.Events {
		if es.InitCycle < existing.InitCycle {
			s.Events = append(s.Events, ScheduledEvent{})
			copy(s.Events[i+1:], s.Events[i:])
			s.Events[i] = es
			return
		}
	}

	s.Events = append(s.Events, es)
}

func (s *Scheduler) popNext() ScheduledEvent {
	next := s.Events[0]
	s.Events = s.Events[1:]
	return next
}

func (s *Scheduler) endFrame() {
	framecycles := int64(CYCLES_PER_FRAME)
	s.CurrentCycle -= framecycles
	for i := range s.Events {
		s.Events[i].InitCycle -= framecycles
	}
}
