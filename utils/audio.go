package utils

import "encoding/binary"

type Stream struct {
	buf []int16
}

func (s *Stream) Write(v ...int16) {
	s.buf = append(s.buf, v...)
}

func (s *Stream) Read(buf []byte) (int, error) {
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
