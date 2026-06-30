package gba

type Event string

const (
	EVENT_VBK             = "event vblank"
	EVENT_HBK             = "event hblank"
	EVENT_DRW             = "event draw"
	EVENT_END_FRAME       = "event end frame"
	EVENT_END_SCANLINE    = "event end scanline"
	EVENT_SND_SAMPLE_GEN  = "event gen sample"
	EVENT_TIMER_RELOAD    = "event timer reload"
	EVENT_TIMER_OVERFLOW0 = "event timer overflow 0"
	EVENT_TIMER_OVERFLOW1 = "event timer overflow 1"
	EVENT_TIMER_OVERFLOW2 = "event timer overflow 2"
	EVENT_TIMER_OVERFLOW3 = "event timer overflow 3"
	EVENT_TIMER_CONTROL   = "event timer control"
	EVENT_DMA0            = "event dma 0"
	EVENT_DMA1            = "event dma 1"
	EVENT_DMA2            = "event dma 2"
	EVENT_DMA3            = "event dma 3"
	EVENT_IRQ_SET         = "irq set"
)

type Scheduler struct {
	Events       [64]ScheduledEvent
	Cnt          int
	CurrentCycle int64
}

type ScheduledEvent struct {
	Event     Event
	Priority  int
	InitCycle int64
	Func      func(int64, any) bool
	Args      any
}

func NewScheduler() *Scheduler {
	return &Scheduler{}
}

func (s *Scheduler) Now() int64 {
	return s.CurrentCycle
}

func (s *Scheduler) schedule(e Event, priority int, cyclesUntil int64, f func(int64, any) bool, args any) {
	s.scheduleAt(e, priority, s.Now()+cyclesUntil, f, args)
}

func (s *Scheduler) scheduleAt(e Event, priority int, initCycle int64, f func(int64, any) bool, args any) {
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
	if len(s.Events) == 0 {
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

func (s *Scheduler) cancel(e Event) {
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
