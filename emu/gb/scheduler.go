package gb

type Event int

const (
	EVENT_VBK = iota
	EVENT_HBK
	EVENT_DRW
	EVENT_END_FRAME
	EVENT_IRQ
	EVENT_DMA

	EVENT_END_SCANLINE
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
		Events: []ScheduledEvent{},
	}
}

func (s *Scheduler) schedule(e Event, cyclesUntil int64) {
	s.scheduleAt(e, s.CurrentCycle+cyclesUntil)
}

func (s *Scheduler) scheduleAt(e Event, initCycle int64) {
	//fmt.Printf("Scheduling Event %d cycles till %d\n", e, cyclesUntilEvent)

	es := ScheduledEvent{
		Event:     e,
		InitCycle: initCycle,
	}

	// Insert in sorted position (smallest cycles first)
	inserted := false
	for i, existing := range s.Events {
		if es.InitCycle < existing.InitCycle {
			// Insert at position i
			s.Events = append(s.Events[:i], append([]ScheduledEvent{es}, s.Events[i:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		s.Events = append(s.Events, es)
	}
}

func (s *Scheduler) popNext() ScheduledEvent {
	next := s.Events[0]
	s.Events = s.Events[1:]
	//fmt.Printf("Popped Event %d cycles till %d\n", next.Event, next.InitCycle)
	return next
}

func (s *Scheduler) endFrame() {
	framecycles := int64(CYCLES_PER_FRAME)
	s.CurrentCycle -= framecycles
	for i := range s.Events {
		s.Events[i].InitCycle -= framecycles
	}
}
