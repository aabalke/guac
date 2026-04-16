package cartridge

type Mbc0 struct {
    Cartridge      *Cartridge
}

func (m *Mbc0) Read(addr uint16) uint8 {
    switch {
    case addr < 0x8000:
        return m.Cartridge.Data[addr]
    default:
        return m.Cartridge.RamData[addr - 0xA000]
    }
}

func (m *Mbc0) Write(addr uint16, v uint8) {}
