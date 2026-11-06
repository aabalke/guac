package ui

import (
	"sync"
	"time"
)

type EventType int

const (
	EVENT_UNKNOWN = iota
	EVENT_FAST_FOCUS_NORTH
	EVENT_FAST_FOCUS_EAST
	EVENT_FAST_FOCUS_WEST
	EVENT_FAST_FOCUS_SOUTH
	EVENT_TOAST
)

type TimedEventBus struct {
	mu     sync.Mutex
	Events []*TimedEvent
}

type TimedEvent struct {
	eventType               EventType
	started, dirty          bool
	TicksLeft               int // assumes 60tps
	OnStart, OnTick, OnDone func()
}

func NewTimedEventBus() *TimedEventBus {
	b := &TimedEventBus{}
	go func() {
		t := time.NewTicker(time.Millisecond * 16)

		for range t.C {
			b.Update()
		}
	}()

	return b
}

func (b *TimedEventBus) Update() {
	b.mu.Lock()
	for _, event := range b.Events {

		if event.dirty {
			continue
		}

		if !event.started {
			event.started = true
			if event.OnStart != nil {
				event.OnStart()
			}
		}

		if event.TicksLeft > 0 {
			if event.OnTick != nil {
				event.OnTick()
			}
			event.TicksLeft--
		} else {
			if event.OnDone != nil {
				event.OnDone()
			}
			event.dirty = true
		}
	}
	b.mu.Unlock()
}

func (b *TimedEventBus) AddEvent(t *TimedEvent) {
	b.mu.Lock()

	// delete dirty events before adding new one
	tmp := b.Events
	b.Events = make([]*TimedEvent, 0, len(b.Events))

	for _, event := range tmp {
		if !event.dirty {
			b.Events = append(b.Events, event)
		}
	}

	b.Events = append(b.Events, t)

	b.mu.Unlock()
}

func (b *TimedEventBus) CancelEventType(eventType EventType) {
	b.mu.Lock()

	for _, event := range b.Events {
		if event.eventType == eventType {
			event.dirty = true
		}
	}

	b.mu.Unlock()
}
