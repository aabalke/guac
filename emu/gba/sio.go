package gba

var sioClock = [...]int64{512, 64}

type Sio struct {
	Cnt uint16
	irq *Irq
	sch *Scheduler
}

func NewSio(irq *Irq, sch *Scheduler) *Sio {
	return &Sio{
		irq: irq,
		sch: sch,
	}
}

func (s *Sio) Write(idx uint32, v uint8) {
	value := uint16(v) << (idx << 3)

	s.Cnt = (s.Cnt & 0x80) | (value &^ 0x80)

	if (s.Cnt&0x80) == 0 && value&0x80 != 0 {
		s.Cnt |= 0x80

		cycles := sioClock[(s.Cnt>>1)&1]
		bytes := ((s.Cnt >> 12) & 1) << 2
		cycles <<= bytes

		s.sch.schedule(EVENT_SIO, 3, cycles, s.CompleteSioTransfer, nil)
	}
}

func (s *Sio) Read(idx uint32) uint8 {
	if idx == 0 {
		return uint8(s.Cnt)
	} else {
		return uint8(s.Cnt >> 8)
	}
}

func (s *Sio) CompleteSioTransfer(_ int64, _ any) {
	if s.Cnt&0x4000 != 0 {
		s.irq.SetIRQ(7)
	}

	s.Cnt &^= 0x80
}
