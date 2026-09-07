package arm7

type Bus7 struct {
	c *Cpu
}

//go:nosplit
func (b *Bus7) Write8(addr uint32, v uint8) {
	b.c.Cycles(addr, 1, NONSEQ, false)
	b.c.Mem.Write8(addr, v)
	b.c.Seq = NONSEQ
	b.c.LastWasDma = false
}

//go:nosplit
func (b *Bus7) Write16(addr uint32, v uint16) {
	b.c.Cycles(addr, 2, NONSEQ, false)
	b.c.Mem.Write16(addr, v)
	b.c.Seq = NONSEQ
	b.c.LastWasDma = false
}

//go:nosplit
func (b *Bus7) Write32(addr, v uint32) {
	b.Write32Block(addr, v, NONSEQ)
}

//go:nosplit
func (b *Bus7) Write32Block(addr, v, seq uint32) {
	b.c.Cycles(addr, 4, seq, false)
	b.c.Mem.Write32(addr, v)
	b.c.Seq = NONSEQ
	b.c.LastWasDma = false
}

//go:nosplit
func (b *Bus7) Read8(addr uint32) uint32 {
	b.c.Cycles(addr, 1, NONSEQ, false)
	v := b.c.Mem.Read8(addr)
	b.c.Idle(1)
	b.c.LastWasDma = false
	return v
}

//go:nosplit
func (b *Bus7) Read16(addr uint32) uint32 {
	b.c.Cycles(addr, 2, NONSEQ, false)
	v := b.c.Mem.Read16(addr)
	b.c.Idle(1)
	b.c.LastWasDma = false
	return v
}

//go:nosplit
func (b *Bus7) Read32(addr uint32) uint32 {
	b.c.Cycles(addr, 4, NONSEQ, false)
	v := b.c.Mem.Read32(addr)
	b.c.Idle(1)
	b.c.LastWasDma = false
	return v
}

//go:nosplit
func (b *Bus7) Read32Block(addr, seq uint32) uint32 {
	b.c.Cycles(addr, 4, seq, false)
	v := b.c.Mem.Read32(addr)
	b.c.LastWasDma = false
	return v
}
