package cartridge

type Mbc0 struct {
	RamEnabled     bool
	RomBank        uint8
	RamBank        uint8
	AdvBankingMode bool
	Latched        bool
}

func (m *Mbc0) ReadRom(c Cartridge, addr uint16) uint8 {
	return c.Data[addr]
}

func (m *Mbc0) ReadRam(c Cartridge, addr uint16) uint8 {
	return uint8(0)
}

func (m *Mbc0) WriteRam(c Cartridge, addr uint16, data uint8) {
}

func (m *Mbc0) Read(c Cartridge, addr uint16) uint8 {
	return c.Data[addr]
}

func (m *Mbc0) Handle(addr uint16, v uint8) {
}
