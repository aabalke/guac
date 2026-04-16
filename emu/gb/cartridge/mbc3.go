package cartridge

type Mbc3 struct {
	Cartridge  *Cartridge
	RamEnabled bool
	RomBank    uint8
	RamBank    uint8
	Rtc        Rtc
}

type Rtc struct {
	Rtc        []uint8
	RtcEnabled bool
	Temp       []uint8
}

func (m *Mbc3) Read(addr uint16) uint8 {
	switch {
	case addr < 0x4000:
		return m.Cartridge.Data[addr]
	case addr < 0x8000:

		if m.RomBank == 0 {
			panic("ROM BANK 0")
		}

		newAddr := uint32(addr - 0x4000)
		a := newAddr + uint32(m.RomBank)*0x4000
		return m.Cartridge.Data[a]

	default:

		switch {
		case m.RamBank >= 0x4 && m.Rtc.RtcEnabled:
			return m.Rtc.Temp[m.RamBank]
		case m.RamBank >= 0x4 && !m.Rtc.RtcEnabled:
			return m.Rtc.Rtc[m.RamBank]
		}

		newAddr := uint32(addr - 0xA000)
		a := newAddr + (uint32(m.RamBank) * 0x2000)
		return m.Cartridge.RamData[a]
	}
}

func (m *Mbc3) Write(addr uint16, v uint8) {
	switch {
	case addr < 0x8000:
		switch {
		case addr < 0x2000:
			m.enableRam(v)
		case addr < 0x4000:
			m.setRomBank1(v)
		case addr < 0x6000:
			m.setRamBank(v)
		case addr < 0x8000:
			m.setAdvBanking(v)
		}
	default:

		if !m.RamEnabled {
			return
		}

		if m.RamBank >= 0x4 {
			m.Rtc.Rtc[m.RamBank] = v
			return
		}

		newAddr := uint32(addr - 0xA000)
		a := newAddr + (uint32(m.RamBank) * 0x2000)
		m.Cartridge.RamData[a] = v
	}
}

func (m *Mbc3) enableRam(v uint8) {
	switch v & 0xF {
	case 0xA:
		m.RamEnabled = true
	case 0x0:
		m.RamEnabled = false
	}
}

func (m *Mbc3) setRomBank1(v uint8) {
	m.RomBank = v & 0b1111111
	if m.RomBank == 0 {
		m.RomBank = 1
	}
}
func (m *Mbc3) setRamBank(v uint8) {
	m.RamBank = v
}
func (m *Mbc3) setAdvBanking(v uint8) {

	if v == 0 {
		m.Rtc.RtcEnabled = true
		copy(m.Rtc.Rtc, m.Rtc.Temp)
		return
	}

	m.Rtc.RtcEnabled = false
}
