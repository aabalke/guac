package arm9

type (
	TimingAttr [4]uint8 // 0 n16, 1 s16, 2 n32, 3 s32
	Timings    [0x100]TimingAttr
)

func NewTimings() *Timings {
	t := &Timings{}

	for i := range t {
		t.setMemTimings(i, 32, 1, 1)
	}

	// PSRAM
	for i := 0x02; i < 0x03; i++ {
		t.setMemTimings(i, 16, 8, 1)
	}

	// Palette, VRAM
	for i := 0x05; i < 0x07; i++ {
		t.setMemTimings(i, 16, 1, 1)
	}

	return t
}

func (t *Timings) setMemTimings(region int, busSize, n, s uint8) {
	t[region][0] = n
	t[region][1] = s

	// on 16 bit bus a 32 bit request requires multiple reads / writes
	if busSize == 16 {
		t[region][2] = n + s
		t[region][3] = s + s
	} else {
		t[region][2] = n
		t[region][3] = s
	}
}
