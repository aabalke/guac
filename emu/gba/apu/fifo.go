package apu

type Fifo struct {
    Buffer [0x20]int8
    Length uint8
    Sample int8
}

func (f *Fifo) Copy(v uint32) {

    if fifoFull := f.Length > 28; fifoFull {
        f.Length -= 28
    }

    for i := range 4 {
        f.Buffer[f.Length] = int8(v >> (8 * i))
        f.Length++
    }
}

func (f *Fifo) Load() {

    if f.Length == 0 {
        return
    }

    f.Sample = f.Buffer[0]
    f.Length--

    for i := range f.Length {
        f.Buffer[i] = f.Buffer[i+1]
    }
}
