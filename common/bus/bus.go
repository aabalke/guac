package bus

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

type EventType int

const (
	INPUT EventType = iota
	PAUSE
	MUTE
	SET_FPS
	LEN_EVENT_TYPE
)

type Event struct {
	Type EventType
	Data any
}

type EventBus struct {
	mu          sync.Mutex
	subscribers [LEN_EVENT_TYPE][]chan Event
}

func NewEventBus() *EventBus {
	return &EventBus{}
}

func (b *EventBus) Subscribe(e EventType, bufsize int) (<-chan Event, func()) {
	ch := make(chan Event, bufsize)
	b.mu.Lock()
	b.subscribers[e] = append(b.subscribers[e], ch)
	b.mu.Unlock()
	cancel := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		list := b.subscribers[e]
		for i, c := range list {
			if c == ch {
				b.subscribers[e] = append(list[:i], list[i+1:]...)
				close(ch)
				break
			}
		}
	}
	return ch, cancel
}

func (b *EventBus) Publish(e Event) {
	b.mu.Lock()
	chans := b.subscribers[e.Type]
	b.mu.Unlock()
	for _, ch := range chans {
		select {
		case ch <- e:
		default:
		}
	}
}

type InputData struct {
	Keys, JustKeys       []ebiten.Key
	Buttons, JustButtons []ebiten.StandardGamepadButton
}
