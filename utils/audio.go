package utils

import (
	"encoding/binary"
	"sync"
)

// need to make atomic, also can s.buf memmoves be replaced with ring buffer or other option

type Stream struct {
	// mutex needed since ebiten audio goroutine uses read and guac emu uses write at the same time
	mu  sync.Mutex
	buf []int16
}

func (s *Stream) Write(v ...int16) {
	s.mu.Lock()
	s.buf = append(s.buf, v...)
	s.mu.Unlock()
}

func (s *Stream) Read(buf []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range len(buf) / 4 {
		var l, r int16
		if len(s.buf) >= 2 {
			l, r = s.buf[0], s.buf[1]
			s.buf = s.buf[2:]
		}
		binary.LittleEndian.PutUint16(buf[4*i+0:], uint16(l))
		binary.LittleEndian.PutUint16(buf[4*i+2:], uint16(r))
	}
	return len(buf), nil
}
