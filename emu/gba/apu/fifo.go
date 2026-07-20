package apu

const (
	BUF_LEN = 7
)

type Fifo struct {
	Count, RdPtr, WrPtr uint8
	Buffer              [BUF_LEN]uint32

	pipe struct {
		word uint32
		size uint8
	}

	Latched int8
}

func (f *Fifo) Reset() {
	f.Count = 0
	f.RdPtr = 0
	f.WrPtr = 0
	f.Buffer = [BUF_LEN]uint32{}
}

func (f *Fifo) Write32(v uint32) {
	if f.Count < BUF_LEN {
		f.Buffer[f.WrPtr] = v

		f.WrPtr++

		if f.WrPtr == BUF_LEN {
			f.WrPtr = 0
		}

		f.Count++
	} else {
		f.Reset()
	}
}

func (f *Fifo) Load() {
	if f.pipe.size == 0 && f.Count > 0 {
		f.pipe.word = f.Buffer[f.RdPtr]

		if f.Count > 0 {
			f.RdPtr++
			if f.RdPtr == BUF_LEN {
				f.RdPtr = 0
			}

			f.Count--
		}

		f.pipe.size = 4
	}

	sample := int8(uint8(f.pipe.word))

	if f.pipe.size > 0 {
		f.pipe.word >>= 8
		f.pipe.size--
	}

	f.Latched = sample
}

func (f *Fifo) Get() int8 {
	return f.Latched
}
