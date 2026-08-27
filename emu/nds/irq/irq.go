package irq

import (
	"github.com/aabalke/guac/emu/gba/irq"
	"github.com/aabalke/guac/emu/scheduler"
)

const (
	IRQ_VBL  = 0
	IRQ_HBL  = 1
	IRQ_VCT  = 2
	IRQ_TMR0 = 3
	IRQ_TMR1 = 4
	IRQ_TMR2 = 5
	IRQ_TMR3 = 6
	IRQ_RTC  = 7 // arm7 only
	IRQ_DMA0 = 8
	IRQ_DMA1 = 9
	IRQ_DMA2 = 10
	IRQ_DMA3 = 11
	IRQ_KEY  = 12
	IRQ_GBA  = 13

	IRQ_IPC_SYNC            = 16
	IRQ_IPC_SEND_FIFO       = 17
	IRQ_IPC_RECV_FIFO       = 18
	IRQ_CARD_TRANS_COMPLETE = 19
	IRQ_CARD_IREQ_MC        = 20
	IRQ_GEO_CMD_FIFO        = 21 // arm9 only
	IRQ_SCREEN_UNFOLDING    = 22 // arm7 only
	IRQ_SPI_BUS             = 23 // arm7 only
	IRQ_WIFI                = 24 // arm7 only
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

	i.OnWrite(0, nil)
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

	i.OnWrite(0, nil)
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

	i.OnWrite(0, nil)
}

func (i *Irq) SetIRQ(irq uint32) {
	i.PendingIF |= (1 << irq)
	i.OnWrite(0, nil)
}

func (i *Irq) OnWrite(late int64, argz any) {
	i.IF = i.PendingIF
	i.IE = i.PendingIE
	i.IME = i.PendingIME

	irqAvailableNew := i.IF&i.IE != 0

	if i.IrqAvailable != irqAvailableNew {
		i.IrqAvailable = irqAvailableNew
	}

	irqLineNew := i.IME && irqAvailableNew

	if *i.IrqLine != irqLineNew {
		*i.IrqLine = irqLineNew
	}
}
