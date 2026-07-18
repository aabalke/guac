package apu

const (
	BUF_LEN      = 0x20
	BUF_MASK     = BUF_LEN - 1
	BUF_OVERFLOW = 0x20 - 4
)

type Fifo struct {
	Length, Head, Tail uint8
	Buffer             [BUF_LEN]int8
	Sample             int8
}

func (f *Fifo) Reset() {
	f.Length = 0
	f.Head = 0
	f.Tail = 0
	f.Buffer = [BUF_LEN]int8{}
}

func (f *Fifo) Copy(v uint32) {
	if f.Length > BUF_OVERFLOW {
		f.Reset()
	}

	for i := range 4 {
		f.Buffer[f.Tail] = int8(v >> (i << 3))
		f.Tail = (f.Tail + 1) & BUF_MASK
		f.Length++
	}
}

func (f *Fifo) Load() {
	if f.Length != 0 {
		f.Sample = f.Buffer[f.Head]
		f.Head = (f.Head + 1) & BUF_MASK
		f.Length--
	}
}
