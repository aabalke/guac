package cartridge

type Mbc1 struct {
	RamEnabled     bool
	RomBank        uint8
	RamBank        uint8
	AdvBankingMode bool
	Latched        bool
}

func (m *Mbc1) ReadRom(c Cartridge, addr uint16) uint8 {

	if m.RomBank == 0 {
		panic("ROM BANK 0")
	}

	newAddr := uint32(addr - 0x4000)
	a := newAddr + uint32(m.RomBank)*0x4000
	return c.Data[a]
}

func (m *Mbc1) ReadRam(c Cartridge, addr uint16) uint8 {

	if !m.RamEnabled {
		return 0xFF
	}

	newAddr := uint32(addr - 0xA000)
	a := newAddr + (uint32(m.RamBank) * 0x2000)
	return c.RamData[a]
}

func (m *Mbc1) WriteRam(c Cartridge, addr uint16, data uint8) {
	if !m.RamEnabled {
		return
	}

	newAddr := uint32(addr - 0xA000)
	a := newAddr + (uint32(m.RamBank) * 0x2000)
	c.RamData[a] = data
}

func (m *Mbc1) Read(c Cartridge, addr uint16) uint8 {
	return c.Data[addr]
}

func (m *Mbc1) Handle(addr uint16, v uint8) {
	switch {
	case addr < 0x2000:
		m.enableRam(v)
	case addr < 0x4000:
		m.setRomBank1(v)
	case addr < 0x6000:
		if m.AdvBankingMode {
			m.setRamBank(v)
			return
		}

		m.setRomBank2(v)

	case addr < 0x8000:
		m.setAdvBanking(v)
	}
}

func (m *Mbc1) enableRam(v uint8) {

	switch v & 0xF {
	case 0xA:
		m.RamEnabled = true
	case 0x0:
		m.RamEnabled = false
	}
}

func (m *Mbc1) setRomBank1(v uint8) {
	m.RomBank &= 0b11100000
	m.RomBank |= (v & 0b11111)
	if m.RomBank == 0 {
		m.RomBank++
	}
}
func (m *Mbc1) setRomBank2(v uint8) {
	m.RomBank &= 0b11111
	m.RomBank |= (v & 0b11100000)
	if m.RomBank == 0 {
		m.RomBank++
	}
}
func (m *Mbc1) setRamBank(v uint8) {
	m.RamBank = v & 0b11
}
func (m *Mbc1) setAdvBanking(v uint8) {

	if adv := v&0b1 == 1; adv {
		m.AdvBankingMode = true
		return
	}

	m.AdvBankingMode = false
	m.RamBank = 0
}
