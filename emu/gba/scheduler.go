package gba

type Event string

const (
	EVENT_VBK            = "event vblank"
	EVENT_HBK            = "event hblank"
	EVENT_DRW            = "event draw"
	EVENT_END_FRAME      = "event end frame"
	EVENT_END_SCANLINE   = "event end scanline"
	EVENT_SND_SAMPLE_GEN = "event gen sample"
	EVENT_TIMER_RELOAD   = "event timer reload"
	EVENT_TIMER_OVERFLOW = "event timer overflow"
	EVENT_TIMER_CONTROL  = "event timer control"
)

type Scheduler struct {
	Events       [64]ScheduledEvent
	Cnt          int
	CurrentCycle int64
}

type ScheduledEvent struct {
	Event     Event
	InitCycle int64
	Func      func(int64, any) bool
	Args      any
}

func NewScheduler() *Scheduler {
	return &Scheduler{}
}

func (s *Scheduler) schedule(e Event, cyclesUntil int64, f func(int64, any) bool, args any) {
	s.scheduleAt(e, s.CurrentCycle+cyclesUntil, f, args)
}

func (s *Scheduler) scheduleAt(e Event, initCycle int64, f func(int64, any) bool, args any) {
	if s.Cnt >= len(s.Events) {
		panic("gba: scheduler reached hard limit")
	}

	es := ScheduledEvent{Event: e, InitCycle: initCycle, Func: f, Args: args}
	for i := range s.Cnt {
		//if es.InitCycle == s.Events[i].InitCycle && es.Event != EVENT_SND_SAMPLE_GEN {
		//	fmt.Printf("Clash: %32s\t%32s\n", es.Event, s.Events[i].Event)
		//}

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
