package gba

import "github.com/aabalke/guac/emu/scheduler"

const (
	IRQ_VBL  = 0
	IRQ_HBL  = 1
	IRQ_VCT  = 2
	IRQ_TMR0 = 3
	IRQ_TMR1 = 4
	IRQ_TMR2 = 5
	IRQ_TMR3 = 6
	IRQ_SER  = 7
	IRQ_DMA0 = 8
	IRQ_DMA1 = 9
	IRQ_DMA2 = 10
	IRQ_DMA3 = 11
	IRQ_KEY  = 12
	IRQ_GBA  = 13
)

type Irq struct {
	gba                  *GBA
	sch                  *scheduler.Scheduler
	pendingIF, pendingIE uint32
	IF, IE               uint32
	IdleIrq              uint32
	IME                  bool
	pendingIME           bool
	IrqAvailable         bool
	IrqLine              bool

	CpuIrqLine *bool

	events struct {
		OnWrite       scheduler.EventIdx
		UpdateIEAndIF scheduler.EventIdx
		UpdateIRQLine scheduler.EventIdx
	}
}

func NewIrq(gba *GBA, s *scheduler.Scheduler) *Irq {
	i := &Irq{}

	i.gba = gba
	i.sch = s
	i.events = struct {
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
	byte := addr & 1
	switch addr {
	case 0x200, 0x201:
		return uint8(i.IE >> (byte << 3))
	case 0x202, 0x203:
		return uint8(i.IF >> (byte << 3))
	case 0x208:
		if i.IME {
			return 1
		}

		return 0
	}

	return 0
}

func (i *Irq) Write8(addr uint32, v uint8) {
	byte := addr & 1

	switch addr {
	case 0x200, 0x201:
		i.pendingIE &^= 0xFF << (byte << 3)
		i.pendingIE |= (uint32(v) << (byte << 3))
		i.pendingIE &= 0x3FFF
	case 0x202, 0x203:
		i.pendingIF &^= uint32(v) << (byte << 3)
	case 0x208:
		i.pendingIME = v&1 != 0
	}

	i.sch.Schedule(i.events.OnWrite, 1, nil)
}

func (i *Irq) Write16(addr uint32, v uint16) {
	switch addr {
	case 0x200:
		i.pendingIE = uint32(v & 0x3FFF)
	case 0x202:
		i.pendingIF &^= uint32(v)
	case 0x208:
		i.pendingIME = v&1 != 0
	}

	i.sch.Schedule(i.events.OnWrite, 1, nil)
}

func (i *Irq) SetIRQ(irq uint32) {
	i.pendingIF |= (1 << irq)
	i.sch.Schedule(i.events.OnWrite, 1, nil)
}

func (i *Irq) OnWrite(late int64, argz any) {
	i.IF = i.pendingIF
	i.IE = i.pendingIE
	i.IME = i.pendingIME

	irqAvailableNew := i.IF&i.IE != 0

	if i.IrqAvailable != irqAvailableNew {
		i.sch.Schedule(i.events.UpdateIEAndIF, 1, irqAvailableNew)
	}

	irqLineNew := i.IME && irqAvailableNew

	if i.IrqLine != irqLineNew {
		i.sch.Schedule(i.events.UpdateIRQLine, 2, irqLineNew)
		i.IrqLine = irqLineNew
	}
}

func (i *Irq) UpdateIEAndIF(late int64, argz any) {
	i.IrqAvailable = argz.(bool)
}

func (i *Irq) UpdateIRQLine(late int64, argz any) {
	*i.CpuIrqLine = argz.(bool)
}
