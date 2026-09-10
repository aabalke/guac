package nds

import "github.com/aabalke/guac/emu/cpu/arm7"

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

func (nds *Nds) Idle9(cycles int64) {
	if nds.dma9.IsRunning() {
		nds.dma9.CheckDmas()
	}

	c := nds.arm9
	// idle 9 is provided as 66mhz cycles
	c.IdleCycles += int64(cycles)
}

func (nds *Nds) Cycles9(addr, width, seq uint32, inst bool) {
	if nds.dma9.IsRunning() {
		nds.dma9.CheckDmas()
	}

	c := nds.arm9

	if inst {
		if c.Itcm.Readable(addr, width) {
			c.InstCycles++
		} else {
			c.InstCycles += c.CyclesPerInst
		}
		return
	}

	// >> 12 4KB pages, uint64 bitmask, 1bit per 4kb page
	idx := (addr >> 12) / 64
	bit := (addr >> 12) & 63

	if c.Cp15.ProtectionUnit.DataCache.Pages[idx]&(1<<bit) != 0 {

		if seq == arm7.SEQ {
			c.DataCycles++
		} else {
			c.DataCycles += 3 // this is rough estimate
		}
		return
	}

	region := addr >> 24
	cycles := int64(c.Timings[region][((width>>2)<<1)|seq]) << 1

	if penalty := region != 2 && seq == arm7.NONSEQ; penalty {
		cycles += 6
	}

	c.DataCycles += cycles
}

func (nds *Nds) CyclesDma9(addr, width, seq uint32) {
	c := nds.arm9

	// >> 12 4KB pages, uint64 bitmask, 1bit per 4kb page
	idx := (addr >> 12) / 64
	bit := (addr >> 12) & 63

	if c.Cp15.ProtectionUnit.DataCache.Pages[idx]&(1<<bit) != 0 {

		if seq == arm7.SEQ {
			nds.Tick9(1)
		} else {
			nds.Tick9(3)
		}
		return
	}

	region := addr >> 24
	cycles := int64(c.Timings[region][((width>>2)<<1)|seq]) << 1

	if penalty := region != 2 && seq == arm7.NONSEQ; penalty {
		cycles += 6
	}

	nds.Tick9(cycles)
}

func (nds *Nds) CyclesDma7(addr, width, seq uint32) {
	nds.Tick7(1)
}
