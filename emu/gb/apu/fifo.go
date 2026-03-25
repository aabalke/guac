package apu

type Fifo struct {
	Buffer [0x20]int8
	Length uint8
	Sample int8

	Head, Tail uint8
}

//func (f *Fifo) Copy(v uint32) {
//
//	if fifoFull := f.Length > 28; fifoFull {
//		f.Length -= 28
//	}
//
//	for i := range 4 {
//		f.Buffer[f.Length] = int8(v >> (i << 3))
//		f.Length++
//	}
//}
//
//func (f *Fifo) Load() {
//
//	if f.Length == 0 {
//		return
//	}
//
//	f.Sample = f.Buffer[0]
//	f.Length--
//
//	for i := range f.Length {
//		f.Buffer[i] = f.Buffer[i+1]
//	}
//}

// Adds 4 bytes from the uint32 `v` into the buffer
func (f *Fifo) Copy(v uint32) {
	// Prevent overflow: drop oldest data if needed
	if f.Length > 28 {
		f.Head = (f.Head + 28) & 0x1F // wrap at 0x20
		f.Length -= 28
	}

	for i := range 4 {
		f.Buffer[f.Tail] = int8(v >> (i << 3))
		f.Tail = (f.Tail + 1) & 0x1F // wrap around
		f.Length++
	}
}

// Loads the next sample from the buffer
func (f *Fifo) Load() {
	if f.Length == 0 {
		return
	}

	f.Sample = f.Buffer[f.Head]
	f.Head = (f.Head + 1) & 0x1F // wrap around
	f.Length--
}
