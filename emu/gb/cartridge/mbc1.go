package cartridge

import (
	"fmt"
	"unsafe"
)

type Mbc1 struct {
	Cartridge *Cartridge

	// registers
	RamEnabled, AdvMode bool
	Bank1, Bank2        uint8

	RomBase, RomBase2, RamBase uint32
}

func NewMbc1(c *Cartridge) *Mbc1 {

	fmt.Printf("Cartridge MBC1\n")

	m := &Mbc1{
		Cartridge: c,
		Bank1:     1,
	}

	m.UpdateAddrs()

	return m
}

func (m *Mbc1) Read(addr uint16) uint8 {
	switch {
	case addr < 0x4000:
		return m.Cartridge.Data[(m.RomBase|uint32(addr))&m.Cartridge.RomMask]
	case addr < 0x8000:
		return m.Cartridge.Data[(m.RomBase2|uint32(addr-0x4000))&m.Cartridge.RomMask]
	case m.RamEnabled:
		return m.Cartridge.RamData[(m.RamBase|uint32(addr-0xA000))&m.Cartridge.RamMask]
	default:
		return 0xFF
	}
}

func (m *Mbc1) ReadPtr(addr uint16) unsafe.Pointer {
	switch {
	case addr < 0x4000:
		return unsafe.Pointer(&m.Cartridge.Data[(m.RomBase|uint32(addr))&m.Cartridge.RomMask])
	case addr < 0x8000:
		return unsafe.Pointer(&m.Cartridge.Data[(m.RomBase2|uint32(addr-0x4000))&m.Cartridge.RomMask])
	case m.RamEnabled:
		return unsafe.Pointer(&m.Cartridge.RamData[(m.RamBase|uint32(addr-0xA000))&m.Cartridge.RamMask])
	default:
		return nil
	}
}

func (m *Mbc1) Write(addr uint16, v uint8) {

	switch {
	case addr < 0x2000:
		m.RamEnabled = v&0xF == 0xA

	case addr < 0x4000:
		m.Bank1 = max(1, v&0x1F)
		m.UpdateAddrs()

	case addr < 0x6000:
		m.Bank2 = v & 0x3
		m.UpdateAddrs()

	case addr < 0x8000:
		m.AdvMode = v&1 != 0
		m.UpdateAddrs()

	case m.RamEnabled:
		m.Cartridge.RamData[(m.RamBase|uint32(addr-0xA000))&m.Cartridge.RamMask] = v
	}
}

func (m *Mbc1) UpdateAddrs() {

	m.RomBase2 = (uint32(m.Bank2) << 19) | (uint32(m.Bank1) << 14)

	if m.AdvMode {
		m.RomBase = uint32(m.Bank2) << 19
		m.RamBase = uint32(m.Bank2) << 13
	} else {
		m.RomBase = 0
		m.RamBase = 0
	}
}
func (m *Mbc1) Save() {}
