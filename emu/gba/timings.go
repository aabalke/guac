package gba

import "github.com/aabalke/guac/emu/cpu/arm7"

var (
	NonSeqWait = [4]uint8{4, 3, 2, 8}
	SeqWait    = [3][2]uint8{
		{2, 1},
		{4, 1},
		{8, 1},
	}
)

type Timings struct {
	// prefetch
	AccessTime, Countdown int64
	Head, Addr            uint32
	Width                 uint32
	Capacity              uint32
	Opcodes               uint32
	Enabled               bool
	Disabled              bool
	Active                bool

	// waitstate
	V       uint16
	Timings [2][2][0x10]uint8
}

func NewTimings() *Timings {
	return &Timings{
		AccessTime: 5,
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

func (t *Timings) Cancel(r15 uint32, tick func(int64)) {
	t.Active = false

	if r15 < 0x800_0000 || r15 >= 0xE00_0000 {
		return
	}

	halfPlusOne := (t.AccessTime >> 1) + 1
	if t.Countdown == 1 || (t.Width == 4 && t.Countdown == halfPlusOne) {
		tick(1)
	}
}

//go:inline
func (t *Timings) Step(cycles int64) {
	t.Countdown -= cycles

	if !t.Enabled {
		t.Opcodes++
		return
	}

	for t.Countdown <= 0 {
		if t.Opcodes >= t.Capacity {
			t.Opcodes++
			break
		}

		t.Opcodes++
		t.Addr += t.Width
		t.Countdown += t.AccessTime
	}
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

func (g *GBA) CyclesDma(addr, width, seq uint32) {
	g.Dma.ParallelDmaCycles = 0

	t := g.Mem.Timings
	region := addr >> 24
	flag32 := width >> 2

	switch {
	case region < 8:
		g.Tick(int64(t.Timings[flag32][0][region]))

	case region < 0x10:

		if region < 14 && addr&0x1FFFF == 0 {
			seq = arm7.NONSEQ
		}

		if t.Active {
			t.Cancel(g.Cpu.Reg.R[15], g.Tick)
		}

		g.Tick(int64(t.Timings[flag32][seq][region]))

	default:
		g.Tick(int64(t.Timings[flag32][0][0]))
	}
}

func (g *GBA) Cycles(addr, width, seq uint32, inst bool) {
	if g.Dma.IsRunning() {
		g.CheckDmas()
	}

	g.Dma.ParallelDmaCycles = 0

	t := g.Mem.Timings
	region := addr >> 24
	flag32 := width >> 2

	if region < 8 {
		g.Tick(int64(t.Timings[flag32][0][region]))
		return
	}

	if inst {

		if t.Active {
			if t.Opcodes != 0 && addr == t.Head {
				// requested addr is first entry in prefetch
				t.Head += width
				g.Scheduler.Add(1)
				t.Countdown--

				if !t.Enabled {
					return
				}

				t.Opcodes--

				for t.Countdown <= 0 {
					if t.Opcodes >= t.Capacity {
						t.Opcodes++
						break
					}

					t.Opcodes++
					t.Addr += width
					t.Countdown += t.AccessTime
				}

				return
			}

			if t.Countdown > 0 && addr == t.Addr {

				// requested addr is being prefetch

				g.Scheduler.Add(t.Countdown)

				if t.Enabled && t.Opcodes < t.Capacity {
					t.Addr += width
					t.Countdown = t.AccessTime
				} else {
					t.Countdown = 0
				}

				t.Head = t.Addr
				t.Opcodes = 0
				return
			}

			// cancel prefetch
			t.Active = false

			halfPlusOne := (t.AccessTime >> 1) + 1
			if t.Countdown == 1 || (t.Width == 4 && t.Countdown == halfPlusOne) {
				g.Scheduler.Add(1)
			}
		}

		if addr&0x1FFFF == 0 || g.Cpu.LastWasDma {
			seq = arm7.NONSEQ
		}

		cycles := t.Timings[flag32][seq][region]

		if t.Disabled {

			if cycles == t.Timings[flag32][arm7.SEQ][region] {
				cycles = t.Timings[flag32][arm7.NONSEQ][8]
			}
			t.Disabled = false
		}

		g.Scheduler.Add(int64(cycles))

		if t.Enabled {
			t.Active = true
			t.Opcodes = 0
			t.Width = width
			t.Addr = addr + width
			t.Head = addr + width
			timing := int64(t.Timings[flag32][1][region])
			t.AccessTime = timing
			t.Countdown = timing
			t.Capacity = (1 << ((width & 3) >> 1)) << 2
		}

		g.Cpu.LastWasDma = false

	} else {
		switch {
		case region < 14:

			if addr&0x1FFFF == 0 || g.Cpu.LastWasDma {
				seq = arm7.NONSEQ
			}

			if t.Active {
				t.Cancel(g.Cpu.Reg.R[15], g.Tick)
			}
			g.Tick(int64(t.Timings[flag32][seq][region]))

		case region < 0x10:
			if t.Active {
				t.Cancel(g.Cpu.Reg.R[15], g.Tick)
			}
			g.Tick(int64(t.Timings[flag32][0][region]))

		default:
			g.Tick(int64(t.Timings[flag32][0][0]))
		}
	}
}

func (gba *GBA) Idle(cycles int64) {
	if gba.Dma.IsRunning() {
		gba.Dma.ParallelDmaCycles = gba.CheckDmas()
	}

	if gba.Dma.ParallelDmaCycles == 0 {
		gba.Tick(cycles)
		gba.Cpu.Seq = arm7.NONSEQ
	} else {
		gba.Dma.ParallelDmaCycles--
	}
}
