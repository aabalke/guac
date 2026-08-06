package utils

import (
	"encoding/binary"
	"sync"
	"time"
)

type Stream struct {
	mu      sync.Mutex
	notFull *sync.Cond
	buf     []int16
	head    int
	cnt     int
	cap     int
}

// NewStream creates a stream sized to hold bufferSize worth of audio at the
// given sampleRate (frames/sec), stereo interleaved.
// e.g. NewStream(20*time.Millisecond, 48000) -> capacity of 1920 int16s
// (48000 * 0.020 * 2 channels)
// notFull will block thread until used up - in turn syncing emulation to audio
func NewStream(bufferSize time.Duration, sampleRate int) *Stream {
	frames := int(bufferSize.Seconds() * float64(sampleRate))

	if frames <= 0 {
		panic("audio stream frame cnt is <= 0")
	}

	samples := frames * 2 // stereo

	s := &Stream{
		buf: make([]int16, samples),
		cap: samples,
	}
	s.notFull = sync.NewCond(&s.mu)
	return s
}

func (s *Stream) Write(v ...int16) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sample := range v {
		for s.cnt == s.cap {
			s.notFull.Wait()
		}
		idx := (s.head + s.cnt) % s.cap
		s.buf[idx] = sample
		s.cnt++
	}
}

func (s *Stream) Read(buf []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range len(buf) / 4 {
		var l, r int16
		if s.cnt >= 2 {
			l = s.buf[s.head]
			r = s.buf[(s.head+1)%s.cap]
			s.head = (s.head + 2) % s.cap
			s.cnt -= 2
		}

		binary.LittleEndian.PutUint16(buf[4*i+0:], uint16(l))
		binary.LittleEndian.PutUint16(buf[4*i+2:], uint16(r))
	}

	s.notFull.Signal()
	return len(buf), nil
}
