package gba

import (
	"fmt"
	"time"

	"github.com/aabalke33/guac/emu/gba/apu"
	"github.com/aabalke33/guac/emu/gba/utils"
)

var (
    _ = fmt.Sprintf("")
)

type Memory struct {
	GBA   *GBA
	BIOS  [0x4000]uint8
	WRAM1 [0x40000]uint8
	WRAM2 [0x8000]uint8

	PRAM [0x400]uint8
	VRAM [0x18000]uint8
	OAM  [0x400]uint8
	IO [0x400]uint8

	BIOS_MODE uint32
	Dispstat Dispstat
}

func NewMemory(gba *GBA) *Memory {
	m := &Memory{GBA: gba}

	m.Write32(0x4000000, 0x80)
	m.Write32(0x4000134, 0x800F) // IR requires bit 3 on. I believe this is auth check (sonic adv)

	m.BIOS_MODE = BIOS_STARTUP

    m.InitSaveLoop()

	return m
}

func (m *Memory) InitSaveLoop() {

    saveTicker := time.Tick(time.Second)

    go func() {
        for range saveTicker {
            if m.GBA.Save && false {
                m.GBA.Cartridge.Save()
                m.GBA.Save = false
            }
        }
    }()
}

func (m *Memory) Read(addr uint32, byteRead bool) uint8 {

	switch {
	case addr < 0x0000_4000:

        if m.GBA.Cpu.Reg.R[PC] >= 0x4000 {
            return m.ReadBios(addr)
        }

        return m.BIOS[addr]

	case addr < 0x0200_0000:
        return m.ReadOpenBus(addr)
	case addr < 0x0300_0000:
		return m.WRAM1[(addr-0x0200_0000)%0x4_0000]
	case addr < 0x0400_0000:
		return m.WRAM2[(addr-0x0300_0000)%0x8_000]
	case addr < 0x0400_0400:
		return m.ReadIO(addr - 0x0400_0000)
	case addr < 0x0500_0000:
        return m.ReadOpenBus(addr & 0b1)
	case addr < 0x0600_0000:
		return m.PRAM[(addr-0x0500_0000)%0x400]
    case addr < 0x700_0000:

        mirrorAddr := (addr - 0x600_0000) % 0x2_0000
        if mirrorAddr >= 0x1_8000 {
            mirrorAddr -= 0x8000 // 32k internal mirror
        }

		return m.VRAM[mirrorAddr]

	case addr < 0x0800_0000:
		return m.OAM[(addr-0x0700_0000)%0x400]

    case addr < 0xE00_0000:
        offset := (addr - 0x0800_0000) % 0x200_0000
        return m.GBA.Cartridge.Rom[offset]

	case addr < 0x1000_0000:
        relative := (addr - 0xE00_0000) % 0x1_0000
        return m.GBA.Cartridge.Read(relative)

	default:
        return m.ReadOpenBus(addr)
	}
}

func (m *Memory) ReadBios(addr uint32) uint8 {

	nAddr, ok := BIOS_ADDR[m.BIOS_MODE]
	if !ok {
		nAddr = 0xE129F000
	}

    //nAddr = BIOS_ADDR[BIOS_SWI]

    switch addr & 0b11 {
    case 0: return uint8(nAddr)
    case 1: return uint8(nAddr >> 8)
    case 2: return uint8(nAddr >> 16)
    case 3: return uint8(nAddr >> 24)
    default: panic("THIS IS IMPOSSIBLE")
    }

}

func (m *Memory) ReadOpenBus(addr uint32) uint8 {

    //r := &m.GBA.Cpu.Reg.R

    //thumb := m.GBA.Cpu.Reg.CPSR.GetFlag(FLAG_T)

    //switch {
    //case DMA_ACTIVE != -1:
    //    return uint8(m.GBA.Dma[DMA_ACTIVE].Value)
    //case thumb && r[PC] - DMA_PC == 2, !thumb && r[PC] - DMA_PC == 4:
    //    return uint8(m.GBA.Dma[DMA_FINISHED].Value)
    //}

    switch addr & 0b11 {
    case 0: return uint8(m.GBA.OpenBusOpcode)
    case 1: return uint8(m.GBA.OpenBusOpcode >> 8)
    case 2: return uint8(m.GBA.OpenBusOpcode >> 16)
    case 3: return uint8(m.GBA.OpenBusOpcode >> 24)
    default: panic("THIS IS IMPOSSIBLE")
    }
}

func (m *Memory) ReadIO(addr uint32) uint8 {

	// this addr should be relative. - 0x400000

	// do not touch the damn bg control regs
    if addr >= 0x60 && addr < 0xB0 {
        return m.ReadSoundIO(addr)
    }

	switch addr {
	case 0x0004:
		return uint8(m.Dispstat)
	case 0x0005:
		return uint8(m.Dispstat >> 8)
    case 0x0006:
        return m.IO[addr]
	case 0x0007:
		return 0x0

    case 0x0010: return m.ReadOpenBus(addr)
    case 0x0011: return m.ReadOpenBus(addr)
    case 0x0012: return m.ReadOpenBus(addr - 2)
    case 0x0013: return m.ReadOpenBus(addr - 2)
    case 0x0014: return m.ReadOpenBus(addr)
    case 0x0015: return m.ReadOpenBus(addr)
    case 0x0016: return m.ReadOpenBus(addr - 2)
    case 0x0017: return m.ReadOpenBus(addr - 2)
    case 0x0018: return m.ReadOpenBus(addr)
    case 0x0019: return m.ReadOpenBus(addr)
    case 0x001A: return m.ReadOpenBus(addr - 2)
    case 0x001B: return m.ReadOpenBus(addr - 2)
    case 0x001C: return m.ReadOpenBus(addr)
    case 0x001D: return m.ReadOpenBus(addr)
    case 0x001E: return m.ReadOpenBus(addr - 2)
    case 0x001F: return m.ReadOpenBus(addr - 2)

    case 0x0020: return m.ReadOpenBus(addr)
    case 0x0021: return m.ReadOpenBus(addr)
    case 0x0022: return m.ReadOpenBus(addr - 2)
    case 0x0023: return m.ReadOpenBus(addr - 2)
    case 0x0024: return m.ReadOpenBus(addr)
    case 0x0025: return m.ReadOpenBus(addr)
    case 0x0026: return m.ReadOpenBus(addr - 2)
    case 0x0027: return m.ReadOpenBus(addr - 2)
    case 0x0028: return m.ReadOpenBus(addr)
    case 0x0029: return m.ReadOpenBus(addr)
    case 0x002A: return m.ReadOpenBus(addr - 2)
    case 0x002B: return m.ReadOpenBus(addr - 2)
    case 0x002C: return m.ReadOpenBus(addr)
    case 0x002D: return m.ReadOpenBus(addr)
    case 0x002E: return m.ReadOpenBus(addr - 2)
    case 0x002F: return m.ReadOpenBus(addr - 2)

    case 0x0030: return m.ReadOpenBus(addr)
    case 0x0031: return m.ReadOpenBus(addr)
    case 0x0032: return m.ReadOpenBus(addr - 2)
    case 0x0033: return m.ReadOpenBus(addr - 2)
    case 0x0034: return m.ReadOpenBus(addr)
    case 0x0035: return m.ReadOpenBus(addr)
    case 0x0036: return m.ReadOpenBus(addr - 2)
    case 0x0037: return m.ReadOpenBus(addr - 2)
    case 0x0038: return m.ReadOpenBus(addr)
    case 0x0039: return m.ReadOpenBus(addr)
    case 0x003A: return m.ReadOpenBus(addr - 2)
    case 0x003B: return m.ReadOpenBus(addr - 2)
    case 0x003C: return m.ReadOpenBus(addr)
    case 0x003D: return m.ReadOpenBus(addr)
    case 0x003E: return m.ReadOpenBus(addr - 2)
    case 0x003F: return m.ReadOpenBus(addr - 2)

    case 0x0040: return m.ReadOpenBus(addr)
    case 0x0041: return m.ReadOpenBus(addr)
    case 0x0042: return m.ReadOpenBus(addr - 2)
    case 0x0043: return m.ReadOpenBus(addr - 2)
    case 0x0044: return m.ReadOpenBus(addr)
    case 0x0045: return m.ReadOpenBus(addr)
    case 0x0046: return m.ReadOpenBus(addr - 2)
    case 0x0047: return m.ReadOpenBus(addr - 2)
    case 0x0048: return m.IO[addr]
    case 0x0049: return m.IO[addr]
    case 0x004A: return m.IO[addr]
    case 0x004B: return m.IO[addr]
    case 0x004C: return m.ReadOpenBus(addr)
    case 0x004D: return m.ReadOpenBus(addr)
    case 0x004E: return m.ReadOpenBus(addr - 2)
    case 0x004F: return m.ReadOpenBus(addr - 2)

    case 0x0050: return m.IO[addr]
    case 0x0051: return m.IO[addr]
    case 0x0052: return m.IO[addr]
    case 0x0053: return m.IO[addr]
    case 0x0054: return m.ReadOpenBus(addr)
    case 0x0055: return m.ReadOpenBus(addr)
    case 0x0056: return m.ReadOpenBus(addr - 2)
    case 0x0057: return m.ReadOpenBus(addr - 2)
    case 0x0058: return m.ReadOpenBus(addr)
    case 0x0059: return m.ReadOpenBus(addr)
    case 0x005A: return m.ReadOpenBus(addr - 2)
    case 0x005B: return m.ReadOpenBus(addr - 2)
    case 0x005C: return m.ReadOpenBus(addr)
    case 0x005D: return m.ReadOpenBus(addr)
    case 0x005E: return m.ReadOpenBus(addr - 2)
    case 0x005F: return m.ReadOpenBus(addr - 2)

    case 0x00B0: return m.ReadOpenBus(addr)
    case 0x00B1: return m.ReadOpenBus(addr)
    case 0x00B2: return m.ReadOpenBus(addr - 2)
    case 0x00B3: return m.ReadOpenBus(addr - 2)
    case 0x00B4: return m.ReadOpenBus(addr)
    case 0x00B5: return m.ReadOpenBus(addr)
    case 0x00B6: return m.ReadOpenBus(addr - 2)
    case 0x00B7: return m.ReadOpenBus(addr - 2)
    case 0x00B8: return 0
    case 0x00B9: return 0
    case 0x00BA: return m.GBA.Dma[0].ReadControl(false)
    case 0x00BB: return m.GBA.Dma[0].ReadControl(true)
    case 0x00BC: return m.ReadOpenBus(addr)
    case 0x00BD: return m.ReadOpenBus(addr)
    case 0x00BE: return m.ReadOpenBus(addr - 2)
    case 0x00BF: return m.ReadOpenBus(addr - 2)
    case 0x00C0: return m.ReadOpenBus(addr)
    case 0x00C1: return m.ReadOpenBus(addr)
    case 0x00C2: return m.ReadOpenBus(addr - 2)
    case 0x00C3: return m.ReadOpenBus(addr - 2)
    case 0x00C4: return 0
    case 0x00C5: return 0
    case 0x00C6: return m.GBA.Dma[1].ReadControl(false)
    case 0x00C7: return m.GBA.Dma[1].ReadControl(true)
    case 0x00C8: return m.ReadOpenBus(addr)
    case 0x00C9: return m.ReadOpenBus(addr)
    case 0x00CA: return m.ReadOpenBus(addr - 2)
    case 0x00CB: return m.ReadOpenBus(addr - 2)
    case 0x00CC: return m.ReadOpenBus(addr)
    case 0x00CD: return m.ReadOpenBus(addr)
    case 0x00CE: return m.ReadOpenBus(addr - 2)
    case 0x00CF: return m.ReadOpenBus(addr - 2)
    case 0x00D0: return 0
    case 0x00D1: return 0
    case 0x00D2: return m.GBA.Dma[2].ReadControl(false)
    case 0x00D3: return m.GBA.Dma[2].ReadControl(true)
    case 0x00D4: return m.ReadOpenBus(addr)
    case 0x00D5: return m.ReadOpenBus(addr)
    case 0x00D6: return m.ReadOpenBus(addr - 2)
    case 0x00D7: return m.ReadOpenBus(addr - 2)
    case 0x00D8: return m.ReadOpenBus(addr)
    case 0x00D9: return m.ReadOpenBus(addr)
    case 0x00DA: return m.ReadOpenBus(addr - 2)
    case 0x00DB: return m.ReadOpenBus(addr - 2)
    case 0x00DC: return 0
    case 0x00DD: return 0
    case 0x00DE: return m.GBA.Dma[3].ReadControl(false)
    case 0x00DF: return m.GBA.Dma[3].ReadControl(true)

    case 0x00E0: return m.ReadOpenBus(addr)
    case 0x00E1: return m.ReadOpenBus(addr)
    case 0x00E2: return m.ReadOpenBus(addr - 2)
    case 0x00E3: return m.ReadOpenBus(addr - 2)
    case 0x00E4: return m.ReadOpenBus(addr)
    case 0x00E5: return m.ReadOpenBus(addr)
    case 0x00E6: return m.ReadOpenBus(addr - 2)
    case 0x00E7: return m.ReadOpenBus(addr - 2)
    case 0x00E8: return m.ReadOpenBus(addr)
    case 0x00E9: return m.ReadOpenBus(addr)
    case 0x00EA: return m.ReadOpenBus(addr - 2)
    case 0x00EB: return m.ReadOpenBus(addr - 2)
    case 0x00EC: return m.ReadOpenBus(addr)
    case 0x00ED: return m.ReadOpenBus(addr)
    case 0x00EE: return m.ReadOpenBus(addr - 2)
    case 0x00EF: return m.ReadOpenBus(addr - 2)

    case 0x00F0: return m.ReadOpenBus(addr)
    case 0x00F1: return m.ReadOpenBus(addr)
    case 0x00F2: return m.ReadOpenBus(addr - 2)
    case 0x00F3: return m.ReadOpenBus(addr - 2)
    case 0x00F4: return m.ReadOpenBus(addr)
    case 0x00F5: return m.ReadOpenBus(addr)
    case 0x00F6: return m.ReadOpenBus(addr - 2)
    case 0x00F7: return m.ReadOpenBus(addr - 2)
    case 0x00F8: return m.ReadOpenBus(addr)
    case 0x00F9: return m.ReadOpenBus(addr)
    case 0x00FA: return m.ReadOpenBus(addr - 2)
    case 0x00FB: return m.ReadOpenBus(addr - 2)
    case 0x00FC: return m.ReadOpenBus(addr)
    case 0x00FD: return m.ReadOpenBus(addr)
    case 0x00FE: return m.ReadOpenBus(addr - 2)
    case 0x00FF: return m.ReadOpenBus(addr - 2)

	case 0x100:
		return m.GBA.Timers[0].ReadD(false)
	case 0x101:
		return m.GBA.Timers[0].ReadD(true)
	case 0x102:
		return m.GBA.Timers[0].ReadCnt(false)
	case 0x103:
		return m.GBA.Timers[0].ReadCnt(true)
	case 0x104:
		return m.GBA.Timers[1].ReadD(false)
	case 0x105:
		return m.GBA.Timers[1].ReadD(true)
	case 0x106:
		return m.GBA.Timers[1].ReadCnt(false)
	case 0x107:
		return m.GBA.Timers[1].ReadCnt(true)
	case 0x108:
		return m.GBA.Timers[2].ReadD(false)
	case 0x109:
		return m.GBA.Timers[2].ReadD(true)
	case 0x10A:
		return m.GBA.Timers[2].ReadCnt(false)
	case 0x10B:
		return m.GBA.Timers[2].ReadCnt(true)
	case 0x10C:
		return m.GBA.Timers[3].ReadD(false)
	case 0x10D:
		return m.GBA.Timers[3].ReadD(true)
	case 0x10E:
		return m.GBA.Timers[3].ReadCnt(false)
	case 0x10F:
		return m.GBA.Timers[3].ReadCnt(true)

	case 0x130:
		return m.GBA.Keypad.readINPUT(false)
	case 0x131:
		return m.GBA.Keypad.readINPUT(true)
	case 0x132:
		return m.GBA.Keypad.readCNT(false)
	case 0x133:
		return m.GBA.Keypad.readCNT(true)

    case 0x136: return 0
    case 0x137: return 0
    case 0x138: return 0
    case 0x139: return 0

    case 0x142: return 0
    case 0x143: return 0

    case 0x15A: return 0
    case 0x15B: return 0

    case 0x200: return uint8(m.GBA.Irq.IE)
    case 0x201: return uint8(m.GBA.Irq.IE >> 8)
    case 0x202: return uint8(m.GBA.Irq.IF)
    case 0x203: return uint8(m.GBA.Irq.IF >> 8)

    case 0x204: return m.IO[addr]
    case 0x205: return m.IO[addr]
    case 0x206: return 0
    case 0x207: return 0
    case 0x208: return m.GBA.Irq.ReadIME()
    case 0x209: return 0

    case 0x20A: return 0
    case 0x20B: return 0

    case 0x301: return 0
    case 0x302: return 0
    case 0x303: return 0
    case 0x304: return 0
	}

	return m.IO[addr]
}

func (m *Memory) Read8(addr uint32) uint32 {
    //SEQ = addr == prevAddr + 1
    //prevAddr = addr

    if v, ok := m.ReadBadRom(addr, 1); ok {
        return v
    }

	return uint32(m.Read(addr, true))
}

// Accessing SRAM Area by 16bit/32bit
// Reading retrieves 8bit value from specified address, multiplied by 0101h (LDRH) or by 01010101h (LDR). Writing changes the 8bit value at the specified address only, being set to LSB of (source_data ROR (address*8)).
func (m *Memory) Read16(addr uint32) uint32 {
    //SEQ = addr == prevAddr + 2
    //prevAddr = addr

    if ok := CheckEeprom(m.GBA, addr); ok {
        return uint32(m.GBA.Cartridge.EepromRead())
    }

	if sram := addr >= 0xE00_0000 && addr < 0x1000_0000; sram {
		return uint32(m.Read(addr, false)) * 0x0101
	}

    if v, ok := m.ReadBadRom(addr, 2); ok {
        return v
    }

	return uint32(m.Read(addr+1, false)) <<8 | uint32(m.Read(addr, false))
}

func (m *Memory) Read32(addr uint32) uint32 {

    if v, ok := m.ReadBadRom(addr, 4); ok {
        return v
    }

	if sram := addr >= 0xE00_0000 && addr < 0x1000_0000; sram {

        //if (addr - 0xE00_0000) & 0b11 != 0 {
        //    return uint32(m.Read(addr + 3, false)) * 0x01010101
        //}

		return uint32(m.Read(addr, false)) * 0x01010101
	}

	return m.Read16(addr+2)<<16 | m.Read16(addr)
}

func (m *Memory) ReadBadRom(addr uint32, bytesRead uint8) (uint32, bool) {

    if addr < 0x800_0000 || addr >= 0xE00_0000 {
        return 0, false
    }

    offset := (addr - 0x800_0000) % 0x200_0000

    if offset >= m.GBA.Cartridge.RomLength {

        switch bytesRead {
        case 1:
            v := ((addr >> 1) >> ((addr & 1) * 8)) & 0xFF
            return uint32(uint8(v)), true
        case 2:

            v := (addr >> 1) & 0xFFFF
            if addr & 1 == 1 {
                v = ((addr >> 1) >> ((addr & 1) * 8)) & 0xFF
            }

            return uint32(uint16(v)), true

        case 4:
            v := ((addr &^ 3) >> 1) & 0xFFFF
            v |= (((addr &^ 3) + 2) >> 1) << 16
            return uint32(v), true
        default:
            panic("BAD ROM READ USING BYTES READ NOT VALID (1, 2, 4)")
        }
    }

    return 0, false
}

func (m *Memory) Write(addr uint32, v uint8, byteWrite bool) {

	switch {
	case addr < 0x0000_4000:
        return
	case addr < 0x0200_0000:
		return
	case addr < 0x0300_0000:
		m.WRAM1[(addr-0x0200_0000)%0x4_0000] = v
        return
	case addr < 0x0400_0000:
		m.WRAM2[(addr-0x0300_0000)%0x8_000] = v
        return
	case addr < 0x0400_0400:
		m.WriteIO(addr-0x0400_0000, v)
        return
	case addr < 0x0500_0000:
		return
	case addr < 0x0600_0000:

        relative := (addr-0x0500_0000)%0x400

        if byteWrite {
            m.PRAM[relative] = v
            if relative + 1 >= uint32(len(m.PRAM)) {
                return
            }

            m.PRAM[relative + 1] = v
            return
        }
		m.PRAM[relative] = v

        return
	case addr < 0x0700_0000:
        /*
             0x16000, 0x8000, 0x8000 | 24_000
            | 64k, 32k 32k (mirror) | mirror of block |
        */
        mirrorAddr := (addr - 0x600_0000) % 0x2_0000
        if mirrorAddr >= 0x1_8000 {
            mirrorAddr -= 0x8000 // 32k internal mirror
        }

        mode := m.ReadIODirect(0x0, 1) & 0b111
        if bitmap := mode > 2; bitmap && byteWrite && mirrorAddr > 0x1_0000 {
            return
        }

        if bgVRAM := mirrorAddr < 0x1_0000; byteWrite && bgVRAM {

            m.VRAM[mirrorAddr] = v

            if mirrorAddr + 1 >= uint32(len(m.VRAM)) {
                return
            }

            m.VRAM[mirrorAddr + 1] = v

            return
        }

        if objVRAM := mirrorAddr >= 0x1_0000; byteWrite && objVRAM {
            return
        }

		m.VRAM[mirrorAddr] = v
        return

	case addr < 0x0800_0000:
        if byteWrite {
            return
        }
        rel := (addr-0x0700_0000)%0x400
		m.OAM[rel] = v
        m.GBA.PPU.UpdateOAM(rel)
        return
	case addr < 0x0E00_0000:
        return

	case addr < 0x1000_0000:

        m.GBA.Save = true

        cartridge := m.GBA.Cartridge
        relative := (addr - 0xE00_0000) % 0x1_0000
        cartridge.Write(relative, v)
        return

	default:
		return
	}
}

func (m *Memory) WriteIO(addr uint32, v uint8) {

	// this addr should be relative. - 0x400000
	// do not make bg control addrs special, unless you know what the f you are doing
	// VCOUNT is not writable, no touchy
    if sound := addr >= 0x60 && addr < 0xB0; sound {
        m.WriteSoundIO(addr, v)
        return
    }

	switch addr {
	case 0x004:
		m.Dispstat.Write(v, false)
	case 0x005:
		m.Dispstat.Write(v, true)
    case 0x006:
        return
    case 0x007:
        return
    case 0x0009: m.IO[addr] = v &^ 0b0010_0000 // BG0CNT mask
    case 0x000B: m.IO[addr] = v &^ 0b0010_0000 // BG1CNT mask

    case 0x0011: m.IO[addr] = v &^ 0b1111_1110 // BG0HOFS mask
    case 0x0013: m.IO[addr] = v &^ 0b1111_1110 // BG0VOFS mask
    case 0x0015: m.IO[addr] = v &^ 0b1111_1110 // BG1HOFS mask
    case 0x0017: m.IO[addr] = v &^ 0b1111_1110 // BG1VOFS mask
    case 0x0019: m.IO[addr] = v &^ 0b1111_1110 // BG2HOFS mask
    case 0x001B: m.IO[addr] = v &^ 0b1111_1110 // BG2VOFS mask
    case 0x001D: m.IO[addr] = v &^ 0b1111_1110 // BG3HOFS mask
    case 0x001F: m.IO[addr] = v &^ 0b1111_1110 // BG3VOFS mask

    case 0x0048: m.IO[addr] = v & 0x3F //winin
    case 0x0049: m.IO[addr] = v & 0x3F //winin
    case 0x004A: m.IO[addr] = v & 0x3F //winout
    case 0x004B: m.IO[addr] = v & 0x3F //winout

    case 0x0050: m.IO[addr] = v// bldcnt
    case 0x0051: m.IO[addr] = v &^ 0b1100_0000 // bldcnt
    case 0x0052: m.IO[addr] = v &^ 0b1110_0000 // bldalpha
    case 0x0053: m.IO[addr] = v &^ 0b1110_0000 // bldalpha

    case 0x00B0: m.GBA.Dma[0].WriteSrc(v, 0)
    case 0x00B1: m.GBA.Dma[0].WriteSrc(v, 1)
    case 0x00B2: m.GBA.Dma[0].WriteSrc(v, 2)
    case 0x00B3: m.GBA.Dma[0].WriteSrc(v, 3)
    case 0x00B4: m.GBA.Dma[0].WriteDst(v, 0)
    case 0x00B5: m.GBA.Dma[0].WriteDst(v, 1)
    case 0x00B6: m.GBA.Dma[0].WriteDst(v, 2)
    case 0x00B7: m.GBA.Dma[0].WriteDst(v, 3)
    case 0x00B8: m.GBA.Dma[0].WriteCount(v, false)
    case 0x00B9: m.GBA.Dma[0].WriteCount(v, true)
    case 0x00BA: m.GBA.Dma[0].WriteControl(v, false)
    case 0x00BB: m.GBA.Dma[0].WriteControl(v, true)
    case 0x00BC: m.GBA.Dma[1].WriteSrc(v, 0)
    case 0x00BD: m.GBA.Dma[1].WriteSrc(v, 1)
    case 0x00BE: m.GBA.Dma[1].WriteSrc(v, 2)
    case 0x00BF: m.GBA.Dma[1].WriteSrc(v, 3)
    case 0x00C0: m.GBA.Dma[1].WriteDst(v, 0)
    case 0x00C1: m.GBA.Dma[1].WriteDst(v, 1)
    case 0x00C2: m.GBA.Dma[1].WriteDst(v, 2)
    case 0x00C3: m.GBA.Dma[1].WriteDst(v, 3)
    case 0x00C4: m.GBA.Dma[1].WriteCount(v, false)
    case 0x00C5: m.GBA.Dma[1].WriteCount(v, true)
    case 0x00C6: m.GBA.Dma[1].WriteControl(v, false)
    case 0x00C7: m.GBA.Dma[1].WriteControl(v, true)
    case 0x00C8: m.GBA.Dma[2].WriteSrc(v, 0)
    case 0x00C9: m.GBA.Dma[2].WriteSrc(v, 1)
    case 0x00CA: m.GBA.Dma[2].WriteSrc(v, 2)
    case 0x00CB: m.GBA.Dma[2].WriteSrc(v, 3)
    case 0x00CC: m.GBA.Dma[2].WriteDst(v, 0)
    case 0x00CD: m.GBA.Dma[2].WriteDst(v, 1)
    case 0x00CE: m.GBA.Dma[2].WriteDst(v, 2)
    case 0x00CF: m.GBA.Dma[2].WriteDst(v, 3)
    case 0x00D0: m.GBA.Dma[2].WriteCount(v, false)
    case 0x00D1: m.GBA.Dma[2].WriteCount(v, true)
    case 0x00D2: m.GBA.Dma[2].WriteControl(v, false)
    case 0x00D3: m.GBA.Dma[2].WriteControl(v, true)
    case 0x00D4: m.GBA.Dma[3].WriteSrc(v, 0)
    case 0x00D5: m.GBA.Dma[3].WriteSrc(v, 1)
    case 0x00D6: m.GBA.Dma[3].WriteSrc(v, 2)
    case 0x00D7: m.GBA.Dma[3].WriteSrc(v, 3)
    case 0x00D8: m.GBA.Dma[3].WriteDst(v, 0)
    case 0x00D9: m.GBA.Dma[3].WriteDst(v, 1)
    case 0x00DA: m.GBA.Dma[3].WriteDst(v, 2)
    case 0x00DB: m.GBA.Dma[3].WriteDst(v, 3)
    case 0x00DC: m.GBA.Dma[3].WriteCount(v, false)
    case 0x00DD: m.GBA.Dma[3].WriteCount(v, true)
    case 0x00DE: m.GBA.Dma[3].WriteControl(v, false)
    case 0x00DF: m.GBA.Dma[3].WriteControl(v, true)

	case 0x100:
		m.GBA.Timers[0].WriteD(v, false)
	case 0x101:
		m.GBA.Timers[0].WriteD(v, true)
	case 0x102:
		m.GBA.Timers[0].WriteCnt(v, false)
	case 0x103:
		m.GBA.Timers[0].WriteCnt(v, true)
	case 0x104:
		m.GBA.Timers[1].WriteD(v, false)
	case 0x105:
		m.GBA.Timers[1].WriteD(v, true)
	case 0x106:
		m.GBA.Timers[1].WriteCnt(v, false)
	case 0x107:
		m.GBA.Timers[1].WriteCnt(v, true)
	case 0x108:
		m.GBA.Timers[2].WriteD(v, false)
	case 0x109:
		m.GBA.Timers[2].WriteD(v, true)
	case 0x10A:
		m.GBA.Timers[2].WriteCnt(v, false)
	case 0x10B:
		m.GBA.Timers[2].WriteCnt(v, true)
	case 0x10C:
		m.GBA.Timers[3].WriteD(v, false)
	case 0x10D:
		m.GBA.Timers[3].WriteD(v, true)
	case 0x10E:
		m.GBA.Timers[3].WriteCnt(v, false)
	case 0x10F:
		m.GBA.Timers[3].WriteCnt(v, true)

	case 0x130:
        return
	case 0x131:
        return
	case 0x132:
		m.GBA.Keypad.writeCNT(v, false)
	case 0x133:
		m.GBA.Keypad.writeCNT(v, true)


    case 0x200: m.GBA.Irq.WriteIE(v, false)
    case 0x201: m.GBA.Irq.WriteIE(v, true)
    case 0x202: m.GBA.Irq.WriteIF(v, false)
    case 0x203: m.GBA.Irq.WriteIF(v, true)

    case 0x204: m.IO[addr] = v
    case 0x205: m.IO[addr] = (m.IO[addr] & 0x80) | (v & 0x5F)
    case 0x206: return
    case 0x207: return

    // IME
    case 0x208: m.GBA.Irq.WriteIME(v); m.IO[addr] = v
    case 0x209: return
    case 0x20A: return
    case 0x20B: return

    case 0x301:
        m.IO[addr] = v & 0x80
        m.GBA.Halted = true

	default:
		m.IO[addr] = v
	}

    if addr == 0x0 || addr == 0x1 || addr >= 0x50 && addr < 0x55 {
        m.GBA.PPU.UpdatePPU(addr, uint32(v))
    }

    if window := addr >= 0x8 && addr < 0x4C; window {
        m.GBA.PPU.UpdatePPU(addr, uint32(v))
    }
}

func (m *Memory) Write8(addr uint32, v uint8) {
	m.Write(addr, v, true)
}

func (m *Memory) Write16(addr uint32, v uint16) {

	if sram := addr >= 0xE00_0000 && addr < 0x1000_0000; sram {
        if addr & 1 == 1 {
            v >>= 8
        }

        m.Write8(addr, uint8(v))
        return
    }

    if ok := CheckEeprom(m.GBA, addr); ok {
        m.GBA.Save = true
        m.GBA.Cartridge.EepromWrite(v)
        return
    }

    m.Write(addr, uint8(v), false)
    m.Write(addr+1, uint8(v>>8), false)
}

func (m *Memory) Write32(addr uint32, v uint32) {

	if sram := addr >= 0xE00_0000 && addr < 0x1000_0000; sram {
        is := addr * 8
        v, _, _ = utils.Ror(v, is, false, false ,false)
        m.Write(addr, uint8(v), false)
        return
    }

    m.Write16(addr, uint16(v))
    m.Write16(addr+2, uint16(v>>16))
}

func CheckEeprom(gba *GBA, addr uint32) bool {

    if addr < 0xD00_0000 || addr >= 0xE00_0000 {
        return false
    }

    if gba.Cartridge.Id != 1 {
        return false
    }

    if gba.Cartridge.RomLength > 0x1000_0000 && addr < 0xDFF_FF00 {
        return false
    }

    return true
}

func (m *Memory) ReadIODirect(addr uint32, size uint32) uint32 {

    switch size {
    case 1:
        return m.ReadIODirectByte(addr)
    case 2:
        return m.ReadIODirectByte(addr + 1) << 8 | m.ReadIODirectByte(addr)
    case 4:
        a := m.ReadIODirectByte(addr + 3) << 8 | m.ReadIODirectByte(addr + 2)
        b := m.ReadIODirectByte(addr + 1) << 8 | m.ReadIODirectByte(addr)
        return (a << 16) | b

    default:
        panic("UNKOWN READ IO DIRECT SIZE")
    }
}

func (m *Memory) ReadIODirectByte(addr uint32) uint32 {
    switch addr {
    case 0x4: return uint32(m.Dispstat)
	case 0x5: return uint32(m.Dispstat >> 8)
    default: return uint32(m.IO[addr])
    }
}

func (m *Memory) ReadSoundIO(addr uint32) uint8 {


    switch addr &^ 0b1 {
    case 0x8C: return m.ReadOpenBus(addr)
    case 0x8E: return m.ReadOpenBus(addr - 2)
    case 0xA0: return m.ReadOpenBus(addr)
    case 0xA2: return m.ReadOpenBus(addr - 2)
    case 0xA4: return m.ReadOpenBus(addr)
    case 0xA6: return m.ReadOpenBus(addr - 2)
    case 0xA8: return m.ReadOpenBus(addr)
    case 0xAA: return m.ReadOpenBus(addr - 2)
    case 0xAC: return m.ReadOpenBus(addr)
    case 0xAE: return m.ReadOpenBus(addr - 2)
    default:
        a := &apu.ApuInstance
        return a.Read(addr)
    }
}

func (m *Memory) WriteSoundIO(addr uint32, v uint8) {
    a := &apu.ApuInstance
    a.Write(addr, v)
}
