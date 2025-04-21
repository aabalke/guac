package gba

type Memory struct {
    GBA *GBA
    BIOS [0x4000]uint8
    WRAM1 [0x40000]uint8
    WRAM2 [0x8000]uint8

    PRAM [0x400]uint8
    VRAM [0x18000]uint8
    OAM  [0x400]uint8
}

func (m *Memory) Read(addr uint32) uint8 {

    switch {
    case addr < 0x00004000: return m.BIOS[addr]
    case addr < 0x02000000: return 0
    case addr < 0x02040000: return m.WRAM1[addr - 0x02000000]
    case addr < 0x03000000: return 0
    case addr < 0x03008000: return m.WRAM2[addr - 0x03000000]
    case addr < 0x04000000: return 0
    case addr < 0x04000400: println("IO"); return 0
    case addr < 0x05000000: return 0
    case addr < 0x05000400: return m.PRAM[addr - 0x050000000]
    case addr < 0x06000000: return 0
    case addr < 0x06018000: return m.VRAM[addr - 0x060000000]
    case addr < 0x07000000: return 0
    case addr < 0x07000400: return m.OAM[addr -  0x070000000]
    case addr < 0x08000000: return 0
    case addr < 0x10000000: println("GAME PAK"); return 0
    default: return 0
    }
}

func (m *Memory) ReadIO(addr uint32) uint8 {

    // this addr should be relative. - 0x400000

    switch {
    case addr < 0x060: println("IO - LCD")
    case addr < 0x0B0: println("IO - SOUND")
    case addr < 0x100: println("IO - DMA")
    case addr < 0x120: println("IO - TIMER")
    case addr < 0x130: println("IO - SERIAL1")
    case addr < 0x134: println("IO - KEYPAD")
    case addr < 0x200: println("IO - SERIAL2")
    default: println("IO - OTHER")
    }

    return 0
}
