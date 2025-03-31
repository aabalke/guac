package cartridge

type Mbc5 struct {
    RamEnabled      bool
    RomBank         uint32
    RamBank         uint32
}


func (m *Mbc5) Read(c Cartridge, addr uint16) uint8 {
    return c.Data[addr]
}

func (m *Mbc5) ReadRom(c Cartridge, addr uint16) uint8 {
    return c.Data[uint32(addr-0x4000)+(m.RomBank*0x4000)]
}

func (m *Mbc5) ReadRam(c Cartridge, addr uint16) uint8 {
    idx := (0x2000*m.RamBank)+uint32(addr-0xA000)
    return c.RamData[idx]
}

func (m *Mbc5) Handle(addr uint16, v uint8) {
    switch {
    case addr < 0x2000: m.enableRam(v)
    case addr < 0x3000: m.RomBank = (m.RomBank & 0x100) | uint32(v)
    case addr < 0x4000: m.RomBank = (m.RomBank & 0xFF) | uint32(v&0x1)<<8
    case addr < 0x6000: m.RamBank = uint32(v & 0xF)
    }
}

func (m *Mbc5) WriteRam(c Cartridge, addr uint16, data uint8) {
    if !m.RamEnabled {
        return
    }

    idx := (0x2000*m.RamBank)+uint32(addr-0xA000)
    c.RamData[idx] = data
}

func (m *Mbc5) enableRam(v uint8) {

    switch v & 0xF {
    case 0xA: m.RamEnabled = true
    case 0x0: m.RamEnabled = false
    }
}
