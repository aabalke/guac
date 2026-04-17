package cartridge

import (
	"fmt"
)

type Mbc5 struct {
	Cartridge *Cartridge

	// registers
	RamEnabled, AdvMode bool
	Bank1, Bank2, Bank3 uint8

	RomBase, RomBase2, RamBase uint32
}

func NewMbc5(c *Cartridge) *Mbc5 {

	fmt.Printf("Cartridge MBC5\n")

    m := &Mbc5{
        Cartridge: c,
        Bank1: 1,
    }

    m.UpdateAddrs()

    return m
}

func (m *Mbc5) Read(addr uint16) uint8 {
	switch {
	case addr < 0x4000:
		return m.Cartridge.Data[(m.RomBase|uint32(addr)) & m.Cartridge.RomMask]
	case addr < 0x8000:
		return m.Cartridge.Data[(m.RomBase2|uint32(addr-0x4000)) & m.Cartridge.RomMask]
	case m.RamEnabled:
		return m.Cartridge.RamData[(m.RamBase|uint32(addr-0xA000))& m.Cartridge.RamMask]
	default:
		return 0xFF
	}
}

func (m *Mbc5) Write(addr uint16, v uint8) {

	switch {
	case addr < 0x2000:
		m.RamEnabled = v & 0xF == 0xA

	case addr < 0x3000:
		m.Bank1 = v
		m.UpdateAddrs()

	case addr < 0x4000:
		m.Bank2 = v & 1
		m.UpdateAddrs()

	case addr < 0x6000:
        m.Bank3 = v & 0xF
		m.UpdateAddrs()

    case addr < 0x8000:
        return

    case m.RamEnabled:
        m.Cartridge.RamData[(m.RamBase|uint32(addr-0xA000))&m.Cartridge.RamMask] = v
	}
}

func (m *Mbc5) UpdateAddrs() {

	m.RomBase2 = (uint32(m.Bank2) << (14 + 8)) | (uint32(m.Bank1) << 14)
    m.RamBase = uint32(m.Bank3) << 13
}

func (m *Mbc5) Save() {}
