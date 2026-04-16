package cartridge

import (
	"fmt"
)

type Mbc2 struct {
	Cartridge *Cartridge

	RamEnabled bool
	Bank1 uint8
	RomBase2 uint32
}

func NewMbc2(c *Cartridge) *Mbc2 {

	fmt.Printf("Cartridge MBC2\n")

    m := &Mbc2{
        Cartridge: c,
        Bank1: 1,
    }

    m.UpdateAddrs()

    return m
}

func (m *Mbc2) Read(addr uint16) uint8 {
	switch {
	case addr < 0x4000:
		return m.Cartridge.Data[uint32(addr) & m.Cartridge.RomMask]
	case addr < 0x8000:
		return m.Cartridge.Data[(m.RomBase2|uint32(addr-0x4000)) & m.Cartridge.RomMask]
	case m.RamEnabled:
		return m.Cartridge.RamData[uint32(addr-0xA000)& m.Cartridge.RamMask & 0x1FF] | 0xF0
	default:
		return 0xFF
	}
}

func (m *Mbc2) Write(addr uint16, v uint8) {

	switch {
	case addr < 0x4000:

        if romb := addr & 0x100 != 0; romb {
            m.Bank1 = max(1, v&0xF)
            m.UpdateAddrs()

        } else {
            m.RamEnabled = v & 0xF == 0xA
        }

	case addr < 0x8000:
        return

    case m.RamEnabled:
        m.Cartridge.RamData[uint32(addr-0xA000)&m.Cartridge.RamMask&0x1FF] = v
	}
}

func (m *Mbc2) UpdateAddrs() {
	m.RomBase2 = uint32(m.Bank1) << 14
}
