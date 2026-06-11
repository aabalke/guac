package gba

var (
	NonSeqWait = [4]uint8{4, 3, 2, 8}
	SeqWait    = [3][2]uint8{
		{2, 1},
		{4, 1},
		{8, 1},
	}
)

type Waitstate struct {
	V uint16

	N [4]uint8
	S [4]uint8

	Prefetch bool
}

func (w *Waitstate) Read(b uint8) uint8 {
	switch b {
	case 0:
		return uint8(w.V)
	case 1:
		return uint8(w.V >> 8)
	}

	return 0
}

func (w *Waitstate) Write(b, v uint8) {
	switch b {
	case 0:
		w.V = (w.V &^ 0xFF) | uint16(v)

		// 0: gp0, 1: gp1, 2: gp2, 3: sram

		w.N[3] = NonSeqWait[v&3] + 1
		w.S[3] = NonSeqWait[v&3] + 1

		w.N[0] = NonSeqWait[(v>>2)&3] + 1
		w.S[0] = SeqWait[0][(v>>4)&1] + 1
		w.N[1] = NonSeqWait[(v>>5)&3] + 1
		w.S[1] = SeqWait[1][(v>>7)&1] + 1

	case 1:
		w.V = (w.V & 0xFF) | (uint16(v) << 8)

		w.N[2] = NonSeqWait[v&3] + 1
		w.S[2] = SeqWait[2][(v>>2)&1] + 1

		w.Prefetch = v&0x40 != 0
	}
}

func (w *Waitstate) Get(width, addr uint32, seq bool) int64 {
	region := (addr >> 25) & 3

	if width == 4 {

		if seq {
			return int64(w.S[region]) + int64(w.S[region])
		}

		return int64(w.N[region]) + int64(w.S[region])
	}

	if seq {
		return int64(w.S[region])
	}

	return int64(w.N[region])
}
