package arm9

import (
	"github.com/aabalke/guac/emu/cpu/arm7"
)

type Bus9 struct {
	c *Cpu
}

//go:nosplit
func (b *Bus9) Write8(addr uint32, v uint8) {
	defer func() {
		b.c.Seq = arm7.NONSEQ
		b.c.LastWasDma = false
	}()

	if ok := b.c.Itcm.Write(addr, v); ok {
		b.c.DataCycles++
		return
	}

	if ok := b.c.Dtcm.Write(addr, v); ok {
		b.c.DataCycles++
		return
	}

	b.c.Cycles(addr, 1, arm7.NONSEQ, false)
	b.c.Mem.Write8(addr, v)
}

//go:nosplit
func (b *Bus9) Write16(addr uint32, v uint16) {
	defer func() {
		b.c.Seq = arm7.NONSEQ
		b.c.LastWasDma = false
	}()

	if ok := b.c.Itcm.Write16(addr, v); ok {
		b.c.DataCycles++
		return
	}

	if ok := b.c.Dtcm.Write16(addr, v); ok {
		b.c.DataCycles++
		return
	}
	b.c.Cycles(addr, 2, arm7.NONSEQ, false)
	b.c.Mem.Write16(addr, v)
}

//go:nosplit
func (b *Bus9) Write32(addr, v uint32) {
	b.c.Write32Block(addr, v, arm7.NONSEQ)
}

//go:nosplit
func (b *Bus9) Write32Block(addr, v, seq uint32) {
	defer func() {
		b.c.Seq = arm7.NONSEQ
		b.c.LastWasDma = false
	}()

	if ok := b.c.Itcm.Write32(addr, v); ok {
		b.c.DataCycles++
		return
	}

	if ok := b.c.Dtcm.Write32(addr, v); ok {
		b.c.DataCycles++
		return
	}
	b.c.Cycles(addr, 4, seq, false)
	b.c.Mem.Write32(addr, v)
}

//go:nosplit
func (b *Bus9) Read8(addr uint32) uint32 {
	if v, ok := b.c.Itcm.Read(addr); ok {
		b.c.DataCycles++
		b.c.LastWasDma = false
		return v
	}

	if v, ok := b.c.Dtcm.Read(addr); ok {
		b.c.DataCycles++
		b.c.LastWasDma = false
		return v
	}

	b.c.Cycles(addr, 1, arm7.NONSEQ, false)
	v := b.c.Mem.Read8(addr)
	b.c.LastWasDma = false
	return v
}

//go:nosplit
func (b *Bus9) Read16(addr uint32) uint32 {
	if v, ok := b.c.Itcm.Read16(addr); ok {
		b.c.DataCycles++
		b.c.LastWasDma = false
		return v
	}

	if v, ok := b.c.Dtcm.Read16(addr); ok {
		b.c.DataCycles++
		b.c.LastWasDma = false
		return v
	}

	b.c.Cycles(addr, 2, arm7.NONSEQ, false)
	v := b.c.Mem.Read16(addr)
	b.c.LastWasDma = false
	return v
}

//go:nosplit
func (b *Bus9) Read32(addr uint32) uint32 {
	return b.c.Read32Block(addr, arm7.NONSEQ)
}

//go:nosplit
func (b *Bus9) Read32Block(addr, seq uint32) uint32 {
	if v, ok := b.c.Itcm.Read32(addr); ok {
		b.c.DataCycles++
		b.c.LastWasDma = false
		return v
	}

	if v, ok := b.c.Dtcm.Read32(addr); ok {
		b.c.DataCycles++
		b.c.LastWasDma = false
		return v
	}

	b.c.Cycles(addr, 4, seq, false)
	v := b.c.Mem.Read32(addr)
	b.c.LastWasDma = false
	return v
}
