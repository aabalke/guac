package gb

type Serial struct {
	out      uint8
	in       uint8
	IsMaster bool
	Enabled  bool
}

func (s *Serial) WriteSc(v uint8) {
	s.IsMaster = v&1 != 0
	s.Enabled = v&0x80 != 0
}

func (s *Serial) ReadSc() uint8 {
	v := uint8(0x7E)

	if s.IsMaster {
		v |= 1
	}

	if s.Enabled {
		v |= 0x80
	}

	return v
}

func (s *Serial) WriteSb(v uint8) {
	s.out = v
}

func (s *Serial) ReadSb() uint8 {
	return s.in
}
