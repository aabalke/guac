package scheduler

type Event int

type Scheduler struct {
	Events       [64]ScheduledEvent
	Cnt          int
	CurrentCycle int64
}

type ScheduledEvent struct {
	Event     Event
	Priority  int
	InitCycle int64
	Func      func(int64, any)
	Args      any
}

func NewScheduler() *Scheduler {
	return &Scheduler{}
}

func (s *Scheduler) Now() int64 {
	return s.CurrentCycle
}

func (s *Scheduler) Schedule(e Event, priority int, cyclesUntil int64, f func(int64, any), args any) {
	// is this needed?
	cyclesUntil = max(cyclesUntil, 0)

	s.ScheduleAt(e, priority, s.Now()+cyclesUntil, f, args)
}

func (s *Scheduler) ScheduleAt(e Event, priority int, initCycle int64, f func(int64, any), args any) {
	if s.Cnt >= len(s.Events) {
		panic("gba: scheduler reached hard limit")
	}

	es := ScheduledEvent{Event: e, Priority: priority, InitCycle: initCycle, Func: f, Args: args}

	var i int
	for ; i < s.Cnt; i++ {

		if es.InitCycle < s.Events[i].InitCycle {
			copy(s.Events[i+1:s.Cnt+1], s.Events[i:s.Cnt])
			break
		}

		if es.InitCycle == s.Events[i].InitCycle && priority <= s.Events[i].Priority {
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

func (s *Scheduler) Cancel(e Event) {
	for i := range s.Cnt {
		if s.Events[i].Event == e {
			copy(s.Events[i:s.Cnt-1], s.Events[i+1:s.Cnt])
			s.Cnt--
			return
		}
	}
}

func (s *Scheduler) Add(cycles int64) {
	//nextCycle := s.CurrentCycle + cycles
	s.CurrentCycle += cycles

	for {
		// if next := s.peekNext(); next == nil || next.InitCycle > nextCycle {
		if next := s.peekNext(); next == nil || next.InitCycle > s.CurrentCycle {
			break
		}

		event := s.popNext()
		late := s.CurrentCycle - event.InitCycle
		//late := int64(0)
		//s.CurrentCycle = event.InitCycle
		event.Func(late, event.Args)
	}

	//s.CurrentCycle = nextCycle
}

func (s *Scheduler) GetRemaining() int64 {
	return s.Events[0].InitCycle - s.Now()
}
