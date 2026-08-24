package irq

import "github.com/aabalke/guac/emu/scheduler"

type Irq struct {
	Sch                  *scheduler.Scheduler
	IrqLine              *bool
	PendingIF, PendingIE uint32
	IF, IE               uint32
	IdleIrq              uint32
	IME                  bool
	PendingIME           bool
	IrqAvailable         bool

	Events struct {
		OnWrite       scheduler.EventIdx
		UpdateIEAndIF scheduler.EventIdx
		UpdateIRQLine scheduler.EventIdx
	}
}

func NewIrq(s *scheduler.Scheduler, irqLine *bool) *Irq {
	i := &Irq{
		IME:     true,
		Sch:     s,
		IrqLine: irqLine,
	}

	i.Events = struct {
		OnWrite       scheduler.EventIdx
		UpdateIEAndIF scheduler.EventIdx
		UpdateIRQLine scheduler.EventIdx
	}{
		s.Register(i.OnWrite, 0),
		s.Register(i.UpdateIEAndIF, 0),
		s.Register(i.UpdateIRQLine, 0),
	}

	return i
}

func (i *Irq) Read(addr uint32) uint8 {
	switch addr {
	case 0x200, 0x201:
		return uint8(i.IE >> ((addr & 1) * 8))
	case 0x202, 0x203:
		return uint8(i.IF >> ((addr & 1) * 8))
	case 0x208:
		if i.IME {
			return 1
		}

		return 0
	}

	return 0
}

func (i *Irq) Write8(addr uint32, v uint8) {
	switch addr {
	case 0x200, 0x201:
		offset := (addr & 1) * 8
		i.PendingIE = ((i.PendingIE &^ (0xFF << offset)) | (uint32(v) << offset)) & 0x3FFF
	case 0x202, 0x203:
		i.PendingIF &^= uint32(v) << ((addr & 1) * 8)
	case 0x208:
		i.PendingIME = v&1 != 0
	}

	i.Sch.Schedule(i.Events.OnWrite, 1, nil)
}

func (i *Irq) Write16(addr uint32, v uint16) {
	switch addr {
	case 0x200:
		i.PendingIE = uint32(v & 0x3FFF)
	case 0x202:
		i.PendingIF &^= uint32(v)
	case 0x208:
		i.PendingIME = v&1 != 0
	}

	i.Sch.Schedule(i.Events.OnWrite, 1, nil)
}

func (i *Irq) SetIRQ(irq uint32) {
	i.PendingIF |= (1 << irq)
	i.Sch.Schedule(i.Events.OnWrite, 1, nil)
}

func (i *Irq) OnWrite(late int64, argz any) {
	i.IF = i.PendingIF
	i.IE = i.PendingIE
	i.IME = i.PendingIME

	irqAvailableNew := i.IF&i.IE != 0

	if i.IrqAvailable != irqAvailableNew {
		i.Sch.Schedule(i.Events.UpdateIEAndIF, 1, irqAvailableNew)
	}

	irqLineNew := i.IME && irqAvailableNew

	if *i.IrqLine != irqLineNew {
		i.Sch.Schedule(i.Events.UpdateIRQLine, 2, irqLineNew)
	}
}

func (i *Irq) UpdateIEAndIF(late int64, argz any) {
	i.IrqAvailable = argz.(bool)
}

func (i *Irq) UpdateIRQLine(late int64, argz any) {
	*i.IrqLine = argz.(bool)
}
