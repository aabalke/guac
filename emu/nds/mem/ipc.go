package mem

import (
	"github.com/aabalke/guac/emu/nds/irq"
)

type Ipc struct {
	Fifo7to9 Fifo
	Fifo9to7 Fifo
}

func NewIpc(irq7, irq9 *irq.Irq) *Ipc {
	ipc := &Ipc{}
	ipc.Fifo7to9.Irq = irq7
	ipc.Fifo9to7.Irq = irq9
	return ipc
}

type Fifo struct {
	Irq                *irq.Irq
	Buffer             [0x10]uint32
	Value              uint32
	Length, Head, Tail uint8
	Sync               uint16
	Error              bool
	IrqEmpty           bool
	IrqNotEmpty        bool
	wasEmptyIrq        bool
	wasNotEmptyIrq     bool
	Enabled            bool
}

func (f *Fifo) Empty() bool { return f.Length == 0 }
func (f *Fifo) Full() bool  { return f.Length == 0x10 }

func (f *Fifo) Write(v uint32) (ok bool) {
	if f.Full() {
		return false
	}

	f.Buffer[f.Tail] = v
	f.Tail = (f.Tail + 1) & 0xF
	f.Length++
	return true
}

func (f *Fifo) Read() (v uint32, ok bool) {
	if f.Empty() {
		return f.Value, false
	}

	f.Value = f.Buffer[f.Head]
	f.Head = (f.Head + 1) & 0xF
	f.Length--
	return f.Value, true
}

func (i *Ipc) WriteSync(v, b uint8, arm9 bool) {
	src := &i.Fifo9to7
	dst := &i.Fifo7to9

	if !arm9 {
		src = &i.Fifo7to9
		dst = &i.Fifo9to7
	}

	if b == 0 {
		src.Sync = (src.Sync &^ 0x00F0) | uint16(v&0x00F0)
		return
	}

	src.Sync = (src.Sync &^ 0x4F00) | ((uint16(v) << 8) & 0x4F00)
	dst.Sync = (dst.Sync &^ 0x000F) | uint16(v&0x000F)

	if (v>>5)&1 != 0 && (dst.Sync>>14)&1 != 0 {
		dst.Irq.SetIRQ(irq.IRQ_IPC_SYNC)
	}
}

func (i *Ipc) ReadSync(b uint8, arm9 bool) uint8 {
	if arm9 {
		return uint8(i.Fifo9to7.Sync >> (b * 8))
	}
	return uint8(i.Fifo7to9.Sync >> (b * 8))
}

func (i *Ipc) WriteCnt(v, b uint8, arm9 bool) {
	src := &i.Fifo9to7
	dst := &i.Fifo7to9

	if !arm9 {
		src = &i.Fifo7to9
		dst = &i.Fifo9to7
	}

	switch b {
	case 0:

		src.IrqEmpty = (v>>2)&1 != 0
		if flush := (v>>3)&1 != 0; flush {
			src.Value = 0
			src.Buffer = [0x10]uint32{}
			src.Length = 0
			src.Head = 0
			src.Tail = 0
		}
	case 1:
		dst.IrqNotEmpty = (v>>2)&1 != 0

		if ackErr := (v>>6)&1 != 0; ackErr {
			src.Error = false
		}

		src.Enabled = v&0x80 != 0
	}
	i.updateIRQs()
}

func (i *Ipc) ReadCnt(b uint8, arm9 bool) uint8 {
	src := &i.Fifo9to7
	dst := &i.Fifo7to9

	if !arm9 {
		src = &i.Fifo7to9
		dst = &i.Fifo9to7
	}

	v := uint8(0)

	if b == 0 {

		if src.Empty() {
			v |= 1 << 0
		}
		if src.Full() {
			v |= 1 << 1
		}
		if dst.IrqEmpty {
			v |= 1 << 2
		}
		return v
	}

	if dst.Empty() {
		v |= 1 << 0
	}
	if dst.Full() {
		v |= 1 << 1
	}
	if dst.IrqNotEmpty {
		v |= 1 << 2
	}
	if src.Error {
		v |= 1 << 6
	}
	if src.Enabled {
		v |= 1 << 7
	}

	return v
}

func (i *Ipc) WriteFifo(v uint32, arm9 bool) {
	src := &i.Fifo9to7
	if !arm9 {
		src = &i.Fifo7to9
	}

	defer i.updateIRQs()

	if !src.Enabled {
		return
	}

	if ok := src.Write(v); !ok {
		src.Error = true
	}
}

func (i *Ipc) ReadFifo(arm9 bool) uint32 {
	src := &i.Fifo9to7
	dst := &i.Fifo7to9

	if !arm9 {
		src = &i.Fifo7to9
		dst = &i.Fifo9to7
	}

	if !src.Enabled {
		return dst.Value
	}

	v, ok := dst.Read()
	if !ok {
		src.Error = true
		return dst.Value
	}

	i.updateIRQs()

	return v
}

func (i *Ipc) updateIRQs() {
	src := &i.Fifo9to7
	dst := &i.Fifo7to9

	for range 2 {

		isEmptyIrq := src.Empty() && src.IrqEmpty
		isNotEmptyIrq := !dst.Empty() && dst.IrqNotEmpty

		if !src.wasEmptyIrq && isEmptyIrq {
			src.Irq.SetIRQ(irq.IRQ_IPC_SEND_FIFO)
		}

		if !src.wasNotEmptyIrq && isNotEmptyIrq {
			src.Irq.SetIRQ(irq.IRQ_IPC_RECV_FIFO)
		}

		src.wasEmptyIrq = isEmptyIrq
		src.wasNotEmptyIrq = isNotEmptyIrq

		src = &i.Fifo7to9
		dst = &i.Fifo9to7
	}
}
