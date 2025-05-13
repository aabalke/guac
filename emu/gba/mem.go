package gba

type Memory struct {
	GBA   *GBA
	BIOS  [0x4000]uint8
	WRAM1 [0x40000]uint8
	WRAM2 [0x8000]uint8

	PRAM [0x400]uint8
	VRAM [0x18000]uint8
	OAM  [0x400]uint8

	IO [0x400]uint8 // THIS IS TEMP

	BIOS_MODE uint32
}

func NewMemory(gba *GBA) *Memory {
	m := &Memory{GBA: gba}

	m.Write(0x4000000, 0x80)

    //m.Write(0x4000130, 0xFF) // KEY INPUT

    m.GBA.Joypad = 0x3FF

	m.BIOS_MODE = BIOS_STARTUP

	return m
}

func (m *Memory) Read(addr uint32) uint8 {

	switch {
	case addr < 0x0000_4000:
		return m.BIOS[addr]
	case addr < 0x0200_0000:
		//return m.BIOS[addr % 0x0000_4000]
        return 0
	case addr < 0x0204_0000:
		return m.WRAM1[addr-0x0200_0000]
	case addr < 0x0300_0000:
		return m.WRAM1[(addr-0x0204_0000) % 0x4_000]
	case addr < 0x0300_8000:
		return m.WRAM2[addr-0x0300_0000]
	case addr < 0x0400_0000:
		return m.WRAM2[(addr-0x0300_8000) % 0x8_000]
	case addr < 0x0400_0400:
		return m.ReadIO(addr - 0x0400_0000)
	case addr < 0x0500_0000:
		return 0
	case addr < 0x0500_0400:
		return m.PRAM[addr-0x0500_0000]
	case addr < 0x0600_0000:
		return m.PRAM[(addr-0x0500_0400) % 0x400]
	case addr < 0x0601_8000:
		return m.VRAM[addr-0x0600_0000]
	case addr < 0x0700_0000:
		return m.VRAM[(addr-0x0601_8000) % 128]
	case addr < 0x0700_0400:
		return m.OAM[addr-0x0700_0000]
	case addr < 0x0800_0000:
		return m.OAM[(addr-0x0700_0400) % 0x400]
	case addr < 0x0A00_0000:
		return m.GBA.Cartridge.Data[addr-0x0800_0000]
	case addr < 0x0E00_0000:
		return m.GBA.Cartridge.Data[(addr-0x0A00_0000) % 0x0200_0000]
    case addr < 0x1000_0000:
		return m.GBA.Cartridge.SRAM[(addr-0x0E00_0000) % 0x1_0000]
	default:
		return 0
	}
}

func (m *Memory) ReadBios(addr uint32) uint32 {

	addr, ok := BIOS_ADDR[m.BIOS_MODE]
	if !ok {
		return 0xE129F000
	}

	return addr
}

func (m *Memory) ReadIO(addr uint32) uint8 {

	// this addr should be relative. - 0x400000


    switch addr {
    //case 0x00: return 0x04
    //case 0x01: return 0x04

    //case 0x04:
    //    // this is temp for testing cpu
    //    if VCOUNT >= 160 && VCOUNT < 227 {
    //        return 0x1
    //    }

    //    return 0x00

    case 0x05:          return 0x00
    case 0x06:          return VCOUNT
    case 0x0204:        panic("CHANGE CART WAIT STATE")
    case KEYINPUT:      return m.GBA.getJoypad(false)
    case KEYINPUT+1:    return m.GBA.getJoypad(true)
    case BG0CNT:        return uint8(Bg0Control)
    case BG0CNT+1:      return uint8(Bg0Control>>8)
    case BG1CNT:        return uint8(Bg1Control)
    case BG1CNT+1:      return uint8(Bg1Control>>8)
    case BG2CNT:        return uint8(Bg2Control)
    case BG2CNT+1:      return uint8(Bg2Control>>8)
    case BG3CNT:        return uint8(Bg3Control)
    case BG3CNT+1:      return uint8(Bg3Control>>8)
    }

	v := m.IO[addr]
    return v

	switch {
	case addr < 0x060:
		println("IO:Read - LCD")
	case addr < 0x0B0:
		println("IO:Read - SOUND")
	case addr < 0x100:
		println("IO:Read - DMA")
	case addr < 0x120:
		println("IO:Read - TIMER")
	case addr < 0x130:
		println("IO:Read - SERIAL1")
	case addr < 0x134:
		println("IO:Read - KEYPAD")
	case addr < 0x200:
		println("IO:Read - SERIAL2")
	default:
		println("IO:Read - OTHER")
	}

	return v
}

func (m *Memory) Read8(addr uint32) uint32 {
	return uint32(m.Read(addr))
}

// Accessing SRAM Area by 16bit/32bit
// Reading retrieves 8bit value from specified address, multiplied by 0101h (LDRH) or by 01010101h (LDR). Writing changes the 8bit value at the specified address only, being set to LSB of (source_data ROR (address*8)).
func (m *Memory) Read16(addr uint32) uint32 {

    if sram := addr > 0xE00_0000 && addr < 0x1000_0000; sram {
        return uint32(m.Read(addr)) * 0x0101
    }

	return uint32(m.Read(addr+1))<<8 | uint32(m.Read(addr))
}

func (m *Memory) Read32(addr uint32) uint32 {
	if addr == 0x0 {
		return m.ReadBios(addr)
	}

    if sram := addr > 0xE00_0000 && addr < 0x1000_0000; sram {
        return uint32(m.Read(addr)) * 0x01010101
    }

	return uint32(m.Read16(addr+2))<<16 | uint32(m.Read16(addr))
}

func (m *Memory) Write(addr uint32, v uint8) {

	switch {
	case addr < 0x0000_4000:
		m.BIOS[addr] = v
	case addr < 0x0200_0000:
		return
	case addr < 0x0204_0000:
		m.WRAM1[addr-0x0200_0000] = v
	case addr < 0x0300_0000:
		return
	case addr < 0x0300_8000:
		m.WRAM2[addr-0x0300_0000] = v
	case addr < 0x0400_0000:
		return
	case addr < 0x0400_0400:
		m.WriteIO(addr-0x0400_0000, v)
	case addr < 0x0500_0000:
		return
	case addr < 0x0500_0400:
		m.PRAM[addr-0x0500_0000] = v
	case addr < 0x0600_0000:
		return
	case addr < 0x0601_8000:
		m.VRAM[addr-0x0600_0000] = v
	case addr < 0x0700_0000:
		return
	case addr < 0x0700_0400:
		m.OAM[addr-0x0700_0000] = v
	case addr < 0x0800_0000:
		return
	case addr < 0x0A00_0000:
		m.GBA.Cartridge.Data[addr-0x0800_0000] = v
	case addr < 0x0E00_0000:
		m.GBA.Cartridge.Data[(addr-0x0A00_0000) % 0x0200_0000] = v
    case addr < 0x1000_0000:
		m.GBA.Cartridge.SRAM[(addr-0x0E00_0000) % 0x1_0000] = v
	default:
		return
    }
	//case addr < 0x1000_0000:
	//	m.GBA.Cartridge.Data[addr-0x0800_0000] = v
	//default:
	//	return
	//}
}

func (m *Memory) WriteIO(addr uint32, v uint8) {

	// this addr should be relative. - 0x400000

    switch addr {
    //case 0x06: VCOUNT = v // not writable
    case BG0CNT: Bg0Control = BgControl((uint32(Bg0Control) &^ 0b11) | uint32(v))
    case BG0CNT+1: Bg0Control = BgControl((uint32(Bg0Control) &^ 0b1100) | (uint32(v) << 8))
    case BG1CNT: Bg1Control = BgControl((uint32(Bg1Control) &^ 0b11) | uint32(v))
    case BG1CNT+1: Bg1Control = BgControl((uint32(Bg1Control) &^ 0b1100) | (uint32(v) << 8))
    case BG2CNT: Bg2Control = BgControl((uint32(Bg2Control) &^ 0b11) | uint32(v))
    case BG2CNT+1: Bg2Control = BgControl((uint32(Bg2Control) &^ 0b1100) | (uint32(v) << 8))
    case BG3CNT: Bg3Control = BgControl((uint32(Bg3Control) &^ 0b11) | uint32(v))
    case BG3CNT+1: Bg3Control = BgControl((uint32(Bg3Control) &^ 0b1100) | (uint32(v) << 8))
    default: m.IO[addr] = v
    }
}

func (m *Memory) Write8(addr uint32, v uint8) {
	m.Write(addr, v)
}

func (m *Memory) Write16(addr uint32, v uint16) {
    //if sram := addr > 0xE00_0000 && addr < 0x1000_0000; sram {
    //    v, _, _ := utils.Ror(uint32(v), (addr & 0xFFFFFF) * 8, false, false, false)
    //    m.Write8(addr, uint8(v))
    //    //m.Write8(addr+1, uint8(v>>8))
    //    return
    //}

	m.Write8(addr, uint8(v))
	m.Write8(addr+1, uint8(v>>8))
}

func (m *Memory) Write32(addr uint32, v uint32) {
	m.Write16(addr, uint16(v))
	m.Write16(addr+2, uint16(v>>16))
}
