package gameboy

type Serial struct {
	sb uint8
	sc uint8
}

func (s *Serial) WriteSb(v uint8) {
	s.sc = v & 0x81
}

func (s *Serial) ReadSb() uint8 {
	return s.sc | 0b01111110
}
