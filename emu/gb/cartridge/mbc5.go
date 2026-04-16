package cartridge

import "unsafe"

type Mbc5 struct {
    Cartridge  *Cartridge
	RamEnabled bool
	RomBank    uint32
	RamBank    uint32
}

func (m *Mbc5) Read(addr uint16) uint8 {
    switch {
    case addr < 0x4000:
        return m.Cartridge.Data[addr]
    case addr < 0x8000:
        return m.Cartridge.Data[uint32(addr-0x4000)+(m.RomBank*0x4000)]
    default:
        idx := (0x2000 * m.RamBank) + uint32(addr-0xA000)
        return m.Cartridge.RamData[idx]
    }
}

func (m *Mbc5) Write(addr uint16, v uint8) {
    switch {
    case addr < 0x8000:
        switch {
        case addr < 0x2000:
            m.enableRam(v)
        case addr < 0x3000:
            m.RomBank = (m.RomBank & 0x100) | uint32(v)
        case addr < 0x4000:
            m.RomBank = (m.RomBank & 0xFF) | uint32(v&0x1)<<8
        case addr < 0x6000:
            m.RamBank = uint32(v & 0xF)
        }
    default:
        if !m.RamEnabled {
            return
        }

        idx := (0x2000 * m.RamBank) + uint32(addr-0xA000)
        m.Cartridge.RamData[idx] = v
    }
}

func (m *Mbc5) enableRam(v uint8) {

	switch v & 0xF {
	case 0xA:
		m.RamEnabled = true
	case 0x0:
		m.RamEnabled = false
	}
}

func (m *Mbc5) ReadPtr(c Cartridge, addr uint16) unsafe.Pointer {

	if uint64(addr)+2 >= uint64(len(c.Data)) {
		return nil
	}

	return unsafe.Pointer(&c.Data[addr])
}

func (m *Mbc5) ReadRomPtr(c Cartridge, addr uint16) unsafe.Pointer {

	if m.RomBank == 0 {
		panic("ROM BANK 0")
	}

	a := uint64(addr - 0x4000)
	a = a + uint64(m.RomBank)*0x4000

	if a+2 >= uint64(len(c.Data)) {
		return nil
	}

	return unsafe.Pointer(&c.Data[a])
}
