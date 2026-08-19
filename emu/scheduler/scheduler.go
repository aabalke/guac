package scheduler

type EventIdx int

type Scheduler struct {
	Registered    []RegisteredEvent
	RegisteredCnt EventIdx

	Events       [64]ScheduledEvent
	Cnt          int
	CurrentCycle int64
}

type RegisteredEvent struct {
	Func     func(late int64, args any)
	Idx      EventIdx
	Priority int
}

type ScheduledEvent struct {
	RegisteredEvent *RegisteredEvent
	InitCycle       int64
	Args            any
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		Registered: []RegisteredEvent{},
	}
}

func (s *Scheduler) Now() int64 {
	return s.CurrentCycle
}

func (s *Scheduler) Register(f func(int64, any), priority int) EventIdx {
	eventIdx := s.RegisteredCnt

	s.Registered = append(s.Registered, RegisteredEvent{
		Func:     f,
		Idx:      eventIdx,
		Priority: priority,
	})

	s.RegisteredCnt++

	return eventIdx
}

func (s *Scheduler) Schedule(idx EventIdx, cyclesUntil int64, args any) {
	// is this needed?
	cyclesUntil = max(cyclesUntil, 0)

	s.ScheduleAt(idx, s.Now()+cyclesUntil, args)
}

func (s *Scheduler) ScheduleAt(idx EventIdx, initCycle int64, args any) {
	if s.Cnt >= len(s.Events) {
		panic("scheduler reached hard limit")
	}

	es := ScheduledEvent{
		RegisteredEvent: &s.Registered[idx],
		InitCycle:       initCycle,
		Args:            args,
	}

	var i int
	for ; i < s.Cnt; i++ {

		if es.InitCycle < s.Events[i].InitCycle {
			copy(s.Events[i+1:s.Cnt+1], s.Events[i:s.Cnt])
			break
		}

		if es.InitCycle == s.Events[i].InitCycle && es.RegisteredEvent.Priority <= s.Events[i].RegisteredEvent.Priority {
			copy(s.Events[i+1:s.Cnt+1], s.Events[i:s.Cnt])
			break
		}
	}

	s.Events[i] = es
	s.Cnt++
}

func (s *Scheduler) peekNext() *ScheduledEvent {
	if s.Cnt == 0 {
		return nil
	}
	return &s.Events[0]
}

func (s *Scheduler) popNext() ScheduledEvent {
	next := s.Events[0]

	copy(s.Events[0:s.Cnt-1], s.Events[1:s.Cnt])
	s.Cnt--
	return next
}

func (s *Scheduler) Cancel(idx EventIdx) {
	for i := range s.Cnt {
		if s.Events[i].RegisteredEvent.Idx == idx {
			copy(s.Events[i:s.Cnt-1], s.Events[i+1:s.Cnt])
			s.Cnt--
			return
		}
	}
}

// nanoboy rewinds curr cycle for event handling
// would not require explicit "late" amount but otherwise I do not think it matters
//func (s *Scheduler) Add(cycles int64) {
//	nextCycle := s.CurrentCycle + cycles
//
//	for {
//		if next := s.peekNext(); next == nil || next.InitCycle > nextCycle {
//			break
//		}
//
//		event := s.popNext()
//		late := int64(0)
//		s.CurrentCycle = event.InitCycle
//		event.Func(late, event.Args)
//	}
//
//	s.CurrentCycle = nextCycle
//}

func (s *Scheduler) Add(cycles int64) {
	s.CurrentCycle += cycles

	for {
		if next := s.peekNext(); next == nil || next.InitCycle > s.CurrentCycle {
			break
		}

		event := s.popNext()
		late := s.CurrentCycle - event.InitCycle
		event.RegisteredEvent.Func(late, event.Args)
	}
}

func (s *Scheduler) GetRemaining() int64 {
	return s.Events[0].InitCycle - s.Now()
}
