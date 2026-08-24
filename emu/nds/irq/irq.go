package irq

import (
	"github.com/aabalke/guac/emu/gba/irq"
	"github.com/aabalke/guac/emu/scheduler"
)

type Irq struct {
	*irq.Irq
}

func NewIrq(s *scheduler.Scheduler, irqLine *bool) *Irq {
	return &Irq{
		irq.NewIrq(s, irqLine),
	}
}

func (i *Irq) Read(addr uint32) uint8 {
	switch addr & 0xFFFF {
	case 0x208:
		if i.IME {
			return 1
		}

		return 0

	case 0x210, 0x211, 0x212, 0x213:
		return uint8(i.IE >> ((addr & 3) * 8))
	case 0x214, 0x215, 0x216, 0x217:
		return uint8(i.IF >> ((addr & 3) * 8))
	}

	return 0
}

func (i *Irq) Write8(addr uint32, v uint8) {
	switch addr & 0xFFFF {
	case 0x208:
		i.PendingIME = v&1 != 0
	case 0x210, 0x211, 0x212, 0x213:
		offset := (addr & 3) * 8
		i.PendingIE = ((i.PendingIE &^ (0xFF << offset)) | (uint32(v) << offset))
	case 0x214, 0x215, 0x216, 0x217:
		i.PendingIF &^= uint32(v) << ((addr & 3) * 8)
	}

	i.Sch.Schedule(i.Events.OnWrite, 1, nil)
}

func (i *Irq) Write16(addr uint32, v uint16) {
	switch addr & 0xFFFF {
	case 0x208:
		i.PendingIME = v&1 != 0
	case 0x210, 0x212:
		offset := (addr & 3) * 8
		i.PendingIE = ((i.PendingIE &^ (0xFFFF << offset)) | (uint32(v) << offset))
	case 0x214, 0x216:
		i.PendingIF &^= uint32(v) << ((addr & 3) * 8)
	}

	i.Sch.Schedule(i.Events.OnWrite, 1, nil)
}

func (i *Irq) Write32(addr uint32, v uint32) {
	switch addr & 0xFFFF {
	case 0x208:
		i.PendingIME = v&1 != 0
	case 0x210:
		i.PendingIE = v
	case 0x214:
		i.PendingIF &^= v
	}

	i.Sch.Schedule(i.Events.OnWrite, 1, nil)
}
