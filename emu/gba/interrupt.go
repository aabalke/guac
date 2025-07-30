package gba

type Irq struct {
	Gba    *GBA
	IF, IE uint16
	IME    bool
}

func (s *Irq) WriteIME(v uint8) {
	s.IME = v&1 == 1
}

func (s *Irq) ReadIME() uint8 {

	if s.IME {
		return 1
	}

	return 0
}

func (s *Irq) WriteIE(v uint8, hi bool) {

	if hi {
		s.IE = (s.IE &^ 0xFF00) | (uint16(v&0xFF) << 8)
		return
	}

	s.IE = (s.IE &^ 0xFF) | uint16(v)
}

func (s *Irq) WriteIF(v uint8, hi bool) {

	if hi {
		s.IF &^= (uint16(v) << 8)
		return
	}

	s.IF &^= uint16(v)
}

func (s *Irq) setIRQ(irq uint32) {
	s.IF |= (1 << irq)
}

func (s *Irq) checkIRQ() {

	interruptEnabled := !s.Gba.Cpu.Reg.CPSR.GetFlag(FLAG_I)
	ime := s.IME
	interrupts := s.IF&s.IE != 0

	if interrupts {
		s.Gba.Halted = false
	}

	if interruptEnabled && ime && interrupts {
		s.Gba.exception(VEC_IRQ, MODE_IRQ)
	}
}
