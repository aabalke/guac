package gba

import (
	"fmt"
    "time"
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
    //IRQFlags IRQFlags

	GamePak0                     [0x200_0000]byte
	SRAM                         [0x1_0000]byte
	Flash                        [0x2_0000]byte // multiple banks
	flashMode                    FlashMode
	flashBank                    uint32
	flashIDMode                  bool
	HasFlash                     bool
}

func NewMemory(gba *GBA) *Memory {
	m := &Memory{GBA: gba}

	m.Write32(0x4000000, 0x80)

	//m.Write(0x4000130, 0xFF) // KEY INPUT

	m.GBA.Joypad = 0x3FF

	m.BIOS_MODE = BIOS_STARTUP

    m.InitSaveLoop()

	return m
}

func (m *Memory) SaveSRAM() {
    if m.GBA.Save {
        m.GBA.Cartridge.save()
        m.GBA.Save = false
    }
}

func (m *Memory) InitSaveLoop() {

    saveTicker := time.Tick(time.Second)

    go func()  {
        for range saveTicker {
            m.SaveSRAM()
        }
    }()
}

func (m *Memory) Read(addr uint32, byteRead bool) uint8 {

	switch {
	case addr < 0x0000_4000:
        //panic(fmt.Sprintf("READING BIOS CURR %d, PC %08X ADDR %08X", CURR_INST, m.GBA.Cpu.Reg.R[PC], addr))
		return m.BIOS[addr]
	case addr < 0x0200_0000:
		//return m.BIOS[addr % 0x0000_4000]
        //fmt.Printf("READING FROM UNUSED MEMORY ADDR %08X\n", addr)
		return 0
	case addr < 0x0300_0000:
		return m.WRAM1[(addr-0x0200_0000)%0x4_0000]
	case addr < 0x0400_0000:
		return m.WRAM2[(addr-0x0300_0000)%0x8_000]
	case addr < 0x0400_0400:
		return m.ReadIO(addr - 0x0400_0000)
	case addr < 0x0500_0000:
        //fmt.Printf("READING FROM UNUSED MEMORY ADDR %08X\n", addr)
		return 0
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

	case addr < 0x0A00_0000:
		//return m.GBA.Cartridge.Data[addr-0x0800_0000]
		return m.GamePak0[addr-0x0800_0000]
	case addr < 0x0E00_0000:

        offset := (addr - 0x0A00_0000) % 0x200_0000 // should be rom length?
        //offset := (addr - 0x0A00_0000) % m.GBA.Cartridge.RomLength // should be rom length?
		//return m.GBA.Cartridge.Data[offset]
		return m.GamePak0[offset]

//	case addr < 0x0E00_0000:
//        // do not make rom length
//        //offset := (addr - 0x0800_0000) % 0x200_0000
//        offset := (addr - 0x0800_0000) % 0x200_0000
//		return m.GBA.Cartridge.Data[offset]

	case addr < 0x1000_0000:

		//if device := addr == 0x0E00_0001; device {
		//	return 0xD4 // temp for mario kart SST
		//	//return 0x09 // temp for pokemon fire Macronix
		//}

		//if manufacturer := addr == 0xE00_0000; manufacturer {
		//	return 0xBF // temp for mario kart SST
		//	//return 0xC2 // temp for pokemon fire Macronix
		//}

		//return m.GBA.Cartridge.SRAM[(addr-0x0E00_0000)%0x1_0000]

        return m.GBA.Mem.FlashRead(addr)
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

	// do not touch the damn bg control regs

	switch addr {
	case 0x0004:
		return uint8(m.Dispstat)
	case 0x0005:
		return uint8(m.Dispstat >> 8)
	case 0x0006:
		return uint8(m.GBA.VCOUNT)

	case 0x0007:
		return 0x0
	//case 0x0204:     panic("CHANGE CART WAIT STATE")
	case KEYINPUT:
		return m.GBA.getJoypad(false)
	case KEYINPUT + 1:
		return m.GBA.getJoypad(true)

    //case 0x0088: return 0x00 // temp sound bias value for ruby
    //case 0x0089: return 0x42 // temp sound bias value for ruby

    //case 0x0050: return 0x44
    //case 0x0051: return 0x3B
    //case 0x0052: return 0x10
    //case 0x0053: return 0x00

    case 0x00B0: return m.GBA.Dma[0].ReadSrc(0)
    case 0x00B1: return m.GBA.Dma[0].ReadSrc(1)
    case 0x00B2: return m.GBA.Dma[0].ReadSrc(2)
    case 0x00B3: return m.GBA.Dma[0].ReadSrc(3)
    case 0x00B4: return m.GBA.Dma[0].ReadDst(0)
    case 0x00B5: return m.GBA.Dma[0].ReadDst(1)
    case 0x00B6: return m.GBA.Dma[0].ReadDst(2)
    case 0x00B7: return m.GBA.Dma[0].ReadDst(3)
    case 0x00B8: return m.GBA.Dma[0].ReadCount(false)
    case 0x00B9: return m.GBA.Dma[0].ReadCount(true)
    case 0x00BA: return m.GBA.Dma[0].ReadControl(false)
    case 0x00BB: return m.GBA.Dma[0].ReadControl(true)
    case 0x00BC: return m.GBA.Dma[1].ReadSrc(0)
    case 0x00BD: return m.GBA.Dma[1].ReadSrc(1)
    case 0x00BE: return m.GBA.Dma[1].ReadSrc(2)
    case 0x00BF: return m.GBA.Dma[1].ReadSrc(3)
    case 0x00C0: return m.GBA.Dma[1].ReadDst(0)
    case 0x00C1: return m.GBA.Dma[1].ReadDst(1)
    case 0x00C2: return m.GBA.Dma[1].ReadDst(2)
    case 0x00C3: return m.GBA.Dma[1].ReadDst(3)
    case 0x00C4: return m.GBA.Dma[1].ReadCount(false)
    case 0x00C5: return m.GBA.Dma[1].ReadCount(true)
    case 0x00C6: return m.GBA.Dma[1].ReadControl(false)
    case 0x00C7: return m.GBA.Dma[1].ReadControl(true)
    case 0x00C8: return m.GBA.Dma[2].ReadSrc(0)
    case 0x00C9: return m.GBA.Dma[2].ReadSrc(1)
    case 0x00CA: return m.GBA.Dma[2].ReadSrc(2)
    case 0x00CB: return m.GBA.Dma[2].ReadSrc(3)
    case 0x00CC: return m.GBA.Dma[2].ReadDst(0)
    case 0x00CD: return m.GBA.Dma[2].ReadDst(1)
    case 0x00CE: return m.GBA.Dma[2].ReadDst(2)
    case 0x00CF: return m.GBA.Dma[2].ReadDst(3)
    case 0x00D0: return m.GBA.Dma[2].ReadCount(false)
    case 0x00D1: return m.GBA.Dma[2].ReadCount(true)
    case 0x00D2: return m.GBA.Dma[2].ReadControl(false)
    case 0x00D3: return m.GBA.Dma[2].ReadControl(true)
    case 0x00D4: return m.GBA.Dma[3].ReadSrc(0)
    case 0x00D5: return m.GBA.Dma[3].ReadSrc(1)
    case 0x00D6: return m.GBA.Dma[3].ReadSrc(2)
    case 0x00D7: return m.GBA.Dma[3].ReadSrc(3)
    case 0x00D8: return m.GBA.Dma[3].ReadDst(0)
    case 0x00D9: return m.GBA.Dma[3].ReadDst(1)
    case 0x00DA: return m.GBA.Dma[3].ReadDst(2)
    case 0x00DB: return m.GBA.Dma[3].ReadDst(3)
    case 0x00DC: return m.GBA.Dma[3].ReadCount(false)
    case 0x00DD: return m.GBA.Dma[3].ReadCount(true)
    case 0x00DE: return m.GBA.Dma[3].ReadControl(false)
    case 0x00DF: return m.GBA.Dma[3].ReadControl(true)

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
	}

	v := m.IO[addr]
	return v
}

func (m *Memory) Read8(addr uint32) uint32 {

	return uint32(m.Read(addr, true))
}

// Accessing SRAM Area by 16bit/32bit
// Reading retrieves 8bit value from specified address, multiplied by 0101h (LDRH) or by 01010101h (LDR). Writing changes the 8bit value at the specified address only, being set to LSB of (source_data ROR (address*8)).
func (m *Memory) Read16(addr uint32) uint32 {

	if sram := addr > 0xE00_0000 && addr < 0x1000_0000; sram {
		return uint32(m.Read(addr, false)) * 0x0101
	}

	return uint32(m.Read(addr+1, false)) <<8 | uint32(m.Read(addr, false))
}

func (m *Memory) Read32(addr uint32) uint32 {

	if addr == 0x0 {
		return m.ReadBios(addr)
	}

	if sram := addr > 0xE00_0000 && addr < 0x1000_0000; sram {
		return uint32(m.Read(addr, false)) * 0x01010101
	}

	return m.Read16(addr+2)<<16 | m.Read16(addr)
}

func (m *Memory) Write(addr uint32, v uint8, byteWrite bool) {

	switch {
	case addr < 0x0000_4000:
		m.BIOS[addr] = v
	case addr < 0x0200_0000:
        //fmt.Printf("COULD NOT WRITE %08X TO ADDR %08X\n", v, addr)
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
        //fmt.Printf("COULD NOT WRITE %08X TO ADDR %08X\n", v, addr)
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

        mode := m.Read(0x400_0000, false) & 0b111
        if bitmap := mode > 2; bitmap && byteWrite && mirrorAddr > 0x1_0000 {
            return
        }

        if byteWrite {
            m.VRAM[mirrorAddr] = v

            if mirrorAddr + 1 >= uint32(len(m.VRAM)) {
                return
            }

            m.VRAM[mirrorAddr + 1] = v

            return
        }

		m.VRAM[mirrorAddr] = v
        return

	case addr < 0x0800_0000:
        if byteWrite {
            return
        }
		m.OAM[(addr-0x0700_0000)%0x400] = v
        return
	case addr < 0x0A00_0000:
        //fmt.Printf("COULD NOT WRITE %08X TO ADDR %08X\n", v, addr)
        return
	case addr < 0x0E00_0000:
        //fmt.Printf("COULD NOT WRITE %08X TO ADDR %08X\n", v, addr)
        return
	case addr < 0x1000_0000:

        m.GBA.Save = true

        //relative := (addr - 0xE00_0000) % 0x1_000

        //if byteWrite {
        //    m.GBA.Cartridge.SRAM[relative] = v
        //    m.GBA.Cartridge.SRAM[relative + 1] = v
        //    return
        //}

		//m.GBA.Cartridge.SRAM[relative] = v

        m.FlashWrite(addr, v)

        return
	default:
        //fmt.Printf("COULD NOT WRITE %08X TO ADDR %08X\n", v, addr)
		return
	}
}

func (m *Memory) WriteIO(addr uint32, v uint8) {

	// this addr should be relative. - 0x400000
	// do not make bg control addrs special, unless you know what the f you are doing
	// VCOUNT is not writable, no touchy

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


    case 0x200, 0x201:
		m.IO[addr] = v
        //fmt.Printf("IRQ EXCEPTION CHECK AT MEM IE\n")
        //m.GBA.checkIRQ()
    case 0x208, 0x209:
		m.IO[addr] = v
        //fmt.Printf("IRQ EXCEPTION CHECK AT MEM IME\n")
        //m.GBA.checkIRQ()

    // manual clear IF by writing 1
    case 0x202: m.IO[addr] &^= v
    case 0x203: m.IO[addr] &^= v

    case 0x301: m.GBA.Halted = true

	default:
		m.IO[addr] = v
	}
}

func (m *Memory) Write8(addr uint32, v uint8) {
	m.Write(addr, v, true)
}

func (m *Memory) Write16(addr uint32, v uint16) {

	//if sram := addr >= 0xE00_0000; sram {
    //    //fmt.Printf("ADDR  %08X V %08X\n", addr, v)
    //    //_, _, _ = utils.Ror(uint32(v), (addr+1) * 8, false, false, false)
    //    //fmt.Printf("ADDR2 %08X V %08X\n", addr, v)
	//    //m.Write(addr, uint8(v), false)
	//    ////m.Write(addr+1, uint8(v>>8), false)
	//    return
	//}

	m.Write(addr, uint8(v), false)
	m.Write(addr+1, uint8(v>>8), false)
}

func (m *Memory) Write32(addr uint32, v uint32) {
	m.Write16(addr, uint16(v))
	m.Write16(addr+2, uint16(v>>16))
}
