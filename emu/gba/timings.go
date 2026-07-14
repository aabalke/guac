package gba

var (
	NonSeqWait = [4]uint8{4, 3, 2, 8}
	SeqWait    = [3][2]uint8{
		{2, 1},
		{4, 1},
		{8, 1},
	}
)

type Timings struct {
	Tick func(cycles int64)

	// prefetch
	AccessTimeShift       int
	AccessTime, Countdown int64

	Head, Addr uint32
	Width      uint32
	Capacity   uint32
	Opcodes    uint32
	Enabled    bool
	Disabled   bool
	Active     bool

	// waitstate
	V       uint16
	Timings [2][2][0x10]uint8
}

func NewTimings(Tick func(int64)) *Timings {
	return &Timings{
		AccessTime: 5,
		Tick:       Tick,
		// [16/32][non/seq][region]
		Timings: [2][2][0x10]uint8{
			{
				//..1..2..3..4..5..6..7..8..9..a..b..c..d..e..f
				{1, 1, 3, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				{1, 1, 3, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			}, {
				//..1..2..3..4..5..6..7..8..9..a..b..c..d..e..f
				{1, 1, 6, 1, 1, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				{1, 1, 6, 1, 1, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			},
		},
	}
}

func (t *Timings) Cancel(r15 uint32) {
	t.Active = false

	if r15 < 0x800_0000 || r15 >= 0xE00_0000 {
		return
	}

	halfPlusOne := (t.AccessTime >> 1) + 1
	if t.Countdown == 1 || (t.Width == 4 && t.Countdown == halfPlusOne) {
		t.Tick(1)
	}
}

//go:inline
func (t *Timings) Step(cycles int64) {
	t.Countdown -= cycles
	if !t.Enabled {
		t.Opcodes++
		return
	}

	if t.Countdown > 0 {
		return
	}

	if t.Opcodes >= t.Capacity {
		t.Opcodes++
		return
	}

	var need uint32
	if notShiftable := t.AccessTimeShift < 0; notShiftable {
		need = uint32((-t.Countdown)/t.AccessTime) + 1
	} else {
		need = uint32(-t.Countdown>>t.AccessTimeShift) + 1
	}

	if avail := t.Capacity - t.Opcodes; need > avail {
		need = avail
		t.Opcodes++
	}

	t.Opcodes += need
	t.Addr += t.Width * need
	t.Countdown += t.AccessTime * int64(need)
}

func (t *Timings) ReadWaitstate(addr uint32) uint8 {
	switch addr & 3 {
	case 0:
		return uint8(t.V)
	case 1:
		return uint8(t.V >> 8)
	default:
		return 0
	}
}

func (t *Timings) WriteWaitstate(addr uint32, v uint8) {
	switch addr & 3 {
	case 0:
		t.V = (t.V &^ 0xFF) | uint16(v)

		// 0: gp0, 1: gp1, 2: gp2, 3: sram

		sram := NonSeqWait[t.V&3] + 1
		t.Timings[0][0][0xE] = sram
		t.Timings[0][0][0xE] = sram
		t.Timings[0][1][0xE] = sram
		t.Timings[0][1][0xE] = sram
		t.Timings[1][0][0xF] = sram
		t.Timings[1][0][0xF] = sram
		t.Timings[1][1][0xF] = sram
		t.Timings[1][1][0xF] = sram

		gp0Non := NonSeqWait[(t.V>>2)&3] + 1
		gp0Seq := SeqWait[0][(t.V>>4)&1] + 1

		t.Timings[0][0][0x8] = gp0Non
		t.Timings[0][1][0x8] = gp0Seq
		t.Timings[1][0][0x8] = gp0Non + gp0Seq
		t.Timings[1][1][0x8] = gp0Seq + gp0Seq
		t.Timings[0][0][0x9] = gp0Non
		t.Timings[0][1][0x9] = gp0Seq
		t.Timings[1][0][0x9] = gp0Non + gp0Seq
		t.Timings[1][1][0x9] = gp0Seq + gp0Seq

		gp1Non := NonSeqWait[(t.V>>5)&3] + 1
		gp1Seq := SeqWait[1][(t.V>>7)&1] + 1

		t.Timings[0][0][0xA] = gp1Non
		t.Timings[0][1][0xA] = gp1Seq
		t.Timings[1][0][0xA] = gp1Non + gp1Seq
		t.Timings[1][1][0xA] = gp1Seq + gp1Seq
		t.Timings[0][0][0xB] = gp1Non
		t.Timings[0][1][0xB] = gp1Seq
		t.Timings[1][0][0xB] = gp1Non + gp1Seq
		t.Timings[1][1][0xB] = gp1Seq + gp1Seq

	case 1:
		t.V = (t.V & 0xFF) | (uint16(v&0x5F) << 8)

		gp2Non := NonSeqWait[(t.V>>8)&3] + 1
		gp2Seq := SeqWait[2][(t.V>>10)&1] + 1

		t.Timings[0][0][0xC] = gp2Non
		t.Timings[0][1][0xC] = gp2Seq
		t.Timings[1][0][0xC] = gp2Non + gp2Seq
		t.Timings[1][1][0xC] = gp2Seq + gp2Seq
		t.Timings[0][0][0xD] = gp2Non
		t.Timings[0][1][0xD] = gp2Seq
		t.Timings[1][0][0xD] = gp2Non + gp2Seq
		t.Timings[1][1][0xD] = gp2Seq + gp2Seq

		old := t.Enabled
		t.Enabled = (t.V>>14)&1 != 0

		if old && !t.Enabled {
			t.Disabled = true
		}
	}
}
