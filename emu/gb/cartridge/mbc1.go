package cartridge

type Mbc1 struct {
    Cartridge      *Cartridge

	RamEnabled, AdvMode    bool
    Bank1, Bank2 uint8
    RomBase, RomBase2, RamBase uint32
}

func (m *Mbc1) Read(addr uint16) uint8 {
    switch {
    case addr < 0x4000:
        return m.Cartridge.Data[m.RomBase + uint32(addr)]
    case addr < 0x8000:
        return m.Cartridge.Data[m.RomBase2 + uint32(addr - 0x4000)]

    default:

        if !m.RamEnabled {
            return 0xFF
        }

        return m.Cartridge.RamData[m.RamBase + uint32(addr - 0xA000)]
    }
}

func (m *Mbc1) Write(addr uint16, v uint8) {
    switch {
    case addr < 0x2000:
        m.RamEnabled = v == 0xA

    case addr < 0x4000:
        m.Bank1 = v & 0x1F
        if m.Bank1 == 0 {
            m.Bank1 = 1
        }
        m.UpdateAddrs()

    case addr < 0x6000:
        m.Bank2 = v & 0x3
        m.UpdateAddrs()

    case addr < 0x8000:
        m.AdvMode = v & 1 != 0
        m.UpdateAddrs()

    default:
        if !m.RamEnabled {
            return
        }

        m.Cartridge.RamData[m.RamBase + uint32(addr - 0xA000)] = v
    }
}

func (m *Mbc1) UpdateAddrs() {

    m.RomBase2 = (uint32(m.Bank2)<<19) | (uint32(m.Bank1) << 14)

    if m.AdvMode {
        m.RomBase = uint32(m.Bank2)<<19
        m.RamBase = uint32(m.Bank2)<<13
    } else {
        m.RomBase = 0
        m.RamBase = 0
    }
}
