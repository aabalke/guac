package gba

import (
	"encoding/binary"
	"time"
	"unsafe"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/bios"
	"github.com/aabalke/guac/utils"
)

type Memory struct {
	GBA   *GBA
	BIOS  *[]uint8
	WRAM1 [0x40000]uint8
	WRAM2 [0x8000]uint8

	PRAM [0x200]uint16
	VRAM [0x18001]uint8
	OAM  [0x400]uint8
	IO   [0x400]uint8

	ProtectedValue uint32
	Dispstat       Dispstat

	Waitstate Waitstate
	Prefetch  *Prefetch

	readRegions  [0x100]func(m *Memory, addr uint32) uint8
	writeRegions [0x100]func(m *Memory, addr uint32, v uint8, byteWrite bool)
}

func NewMemory(gba *GBA) *Memory {
	m := &Memory{GBA: gba}
	m.ProtectedValue = 0xE129F000

	m.Prefetch = NewPrefetch(&m.Waitstate)
	m.Waitstate.Prefetch = m.Prefetch

	m.initReadRegions()
	m.initWriteRegions()

	m.Write32(0x4000000, 0x80)
	m.Write32(0x4000134, 0x800F) // IR requires bit 3 on. I believe this is auth check (sonic adv)

	m.Write32(0x4000204, 0x0000)
	m.Write32(0x4000088, 0x0200)

	m.InitSaveLoop()

	return m
}

func (m *Memory) LoadBios() {
	if p := config.Conf.Gba.Bios.Path; p != "" {
		buf, _, _ := utils.ReadFile(p)
		m.BIOS = &buf
	} else {
		m.BIOS = &bios.BiosGba
	}
}

func (m *Memory) InitSaveLoop() {
	saveTicker := time.Tick(time.Second)

	go func() {
		for range saveTicker {

			if config.Conf.General.DisableSaves {
				continue
			}

			if m.GBA.Save {
				m.GBA.Cartridge.Save()
				m.GBA.Save = false
			}
		}
	}()
}

func (m *Memory) initWriteRegions() {
	for i := range len(m.writeRegions) {
		m.writeRegions[i] = func(m *Memory, addr uint32, v uint8, byteWrite bool) {
		}
	}

	m.writeRegions[0x2] = func(m *Memory, addr uint32, v uint8, byteWrite bool) {
		m.WRAM1[addr&0x3_FFFF] = v
	}

	m.writeRegions[0x3] = func(m *Memory, addr uint32, v uint8, byteWrite bool) {
		m.WRAM2[addr&0x7FFF] = v
	}

	m.writeRegions[0x4] = func(m *Memory, addr uint32, v uint8, byteWrite bool) {
		if addr < 0x0400_0400 {
			m.WriteIO(addr&0x3FF, v)
		}
	}

	m.writeRegions[0x5] = func(m *Memory, addr uint32, v uint8, byteWrite bool) {
		relative := addr & 0x3FF

		if relative&1 == 1 {
			m.PRAM[relative>>1] &= 0xFF
			m.PRAM[relative>>1] |= uint16(v) << 8

			return
		}

		if byteWrite {
			m.PRAM[relative>>1] &^= 0xFF
			m.PRAM[relative>>1] |= uint16(v)

			m.PRAM[relative>>1] &= 0xFF
			m.PRAM[relative>>1] |= uint16(v) << 8
			return
		}

		m.PRAM[relative>>1] &^= 0xFF
		m.PRAM[relative>>1] |= uint16(v)
	}

	m.writeRegions[0x6] = func(m *Memory, addr uint32, v uint8, byteWrite bool) {
		addr &= 0x1FFFF
		if addr >= 0x1_8000 {
			addr -= 0x8000 // 32k internal mirror
		}

		if !byteWrite {
			m.VRAM[addr] = v
			return
		}

		if bgVRAM := addr < 0x1_0000; bgVRAM {

			m.VRAM[addr] = v

			if addr+1 >= uint32(len(m.VRAM)) {
				return
			}

			m.VRAM[addr+1] = v

			return
		}
	}

	m.writeRegions[0x7] = func(m *Memory, addr uint32, v uint8, byteWrite bool) {
		if byteWrite {
			return
		}
		rel := addr & 0x3FF
		m.OAM[rel] = v
		m.GBA.PPU.UpdateOAM(rel)
	}

	m.writeRegions[0xE] = func(m *Memory, addr uint32, v uint8, byteWrite bool) {
		m.GBA.Save = true

		cartridge := m.GBA.Cartridge
		relative := addr & 0xFFFF

		cartridge.Write(relative, v)
	}

	m.writeRegions[0xF] = func(m *Memory, addr uint32, v uint8, byteWrite bool) {
		m.GBA.Save = true

		cartridge := m.GBA.Cartridge
		relative := addr & 0xFFFF

		cartridge.Write(relative, v)
	}
}

func (m *Memory) initReadRegions() {
	for i := range len(m.readRegions) {
		m.readRegions[i] = func(m *Memory, addr uint32) uint8 {
			if m.GBA.Cpu.Reg.R[PC] < 0x4000 {
				return 0
			}
			return m.ReadOpenBus(addr)
		}
	}

	m.readRegions[0x0] = func(m *Memory, addr uint32) uint8 {
		if addr < 0x4000 {

			pc := m.GBA.Cpu.Reg.R[15]

			if pc >= 0x4000 {
				return m.ReadBios(addr)
			}

			if pc == 0xDC || pc == 0x134 || pc == 0x13C || pc == 0x188 {
				m.ProtectedValue = binary.LittleEndian.Uint32((*m.BIOS)[pc+8:])
			}

			return (*m.BIOS)[addr]
		}

		return m.ReadOpenBus(addr)
	}

	m.readRegions[0x2] = func(m *Memory, addr uint32) uint8 {
		return m.WRAM1[addr&0x3FFFF]
	}

	m.readRegions[0x3] = func(m *Memory, addr uint32) uint8 {
		return m.WRAM2[addr&0x7FFF]
	}

	m.readRegions[0x4] = func(m *Memory, addr uint32) uint8 {
		if addr < 0x0400_0400 {
			return m.ReadIO(addr & 0x3FF)
		}
		return m.ReadOpenBus(addr)
	}

	m.readRegions[0x5] = func(m *Memory, addr uint32) uint8 {
		if addr&1 == 1 {
			return uint8(m.PRAM[addr&0x3FF>>1] >> 8)
		}

		return uint8(m.PRAM[addr&0x3FF>>1])
	}

	m.readRegions[0x6] = func(m *Memory, addr uint32) uint8 {
		addr &= 0x1FFFF
		if addr >= 0x18000 {
			addr -= 0x8000
		}
		return m.VRAM[addr]
	}

	m.readRegions[0x7] = func(m *Memory, addr uint32) uint8 {
		return m.OAM[addr&0x3FF]
	}

	for i := 0x8; i < 0xE; i++ {
		m.readRegions[i] = func(m *Memory, addr uint32) uint8 {
			return m.GBA.Cartridge.Rom[addr&0x1FFFFFF]
		}
	}

	m.readRegions[0xE] = func(m *Memory, addr uint32) uint8 {
		return m.GBA.Cartridge.Read(addr & 0xFFFF)
	}

	m.readRegions[0xF] = func(m *Memory, addr uint32) uint8 {
		return m.GBA.Cartridge.Read(addr & 0xFFFF)
	}
}

func (m *Memory) ReadPtr(addr uint32) (unsafe.Pointer, bool) {
	return nil, false
	switch regions := addr >> 24; regions {
	case 0x2:
		return unsafe.Add(
			unsafe.Pointer(&m.WRAM1), addr&0x3FFFF,
		), true
	case 0x3:
		return unsafe.Add(
			unsafe.Pointer(&m.WRAM2), addr&0x7FFF,
		), true
	case 0x6:
		addr &= 0x1FFFF
		if addr >= 0x18000 {
			addr -= 0x8000
		}
		return unsafe.Add(
			unsafe.Pointer(&m.VRAM), addr,
		), true

	case 0x7:
		return unsafe.Add(
			unsafe.Pointer(&m.OAM), addr&0x3FF,
		), true

	case 0x8, 0x9, 0xA, 0xB, 0xC, 0xD:
		return unsafe.Add(
			unsafe.Pointer(&m.GBA.Cartridge.Rom), addr&0x1FF_FFFF,
		), true
	default:
		return nil, false
	}
}

func (m *Memory) Read(addr uint32) uint8 {
	return m.readRegions[addr>>24](m, addr)
}

func (m *Memory) ReadBios(addr uint32) uint8 {
	return uint8(m.ProtectedValue >> ((addr & 3) << 3))
}

func (m *Memory) ReadOpenBus(addr uint32) uint8 {
	pc := m.GBA.Cpu.Reg.R[15]

	if m.GBA.Cpu.Reg.CPSR.T {
		// does pipeline impliment region based thumb mode?
		return uint8(m.Read16(pc) >> ((addr & 1) << 3))
	}

	return uint8(m.Read32(pc) >> ((addr & 3) << 3))
}

func (m *Memory) ReadIO(addr uint32) uint8 {
	switch {
	case addr >= 0x10 && addr < 0x48,
		addr >= 0x4C && addr < 0x50,
		addr >= 0x54 && addr < 0x60,
		addr >= 0xB0 && addr < 0xB8,
		addr >= 0xBC && addr < 0xC4,
		addr >= 0xC8 && addr < 0xD0,
		addr >= 0xD4 && addr < 0xDC,
		addr >= 0xE0 && addr < 0x100:
		return m.ReadOpenBus(addr)
	case addr >= 0x60 && addr < 0xB0:
		return m.ReadSoundIO(addr)
	case addr >= 0xB0 && addr < 0xE0:

		addr -= 0xB0
		i := addr / 12
		addr %= 12
		return m.GBA.Dma[i].Read(addr)

	case addr >= 0x100 && addr < 0x110:

		addr -= 0x100
		i := addr / 4
		idx := addr & 3
		return m.GBA.Timers[i].Read(int(idx))

	}

	switch addr {
	case 0x0004:
		return uint8(m.Dispstat)
	case 0x0005:
		return uint8(m.Dispstat >> 8)
	case 0x0007:
		return 0

	case 0x130:
		return m.GBA.Keypad.readINPUT(false)
	case 0x131:
		return m.GBA.Keypad.readINPUT(true)
	case 0x132:
		return m.GBA.Keypad.readCNT(false)
	case 0x133:
		return m.GBA.Keypad.readCNT(true)

	case 0x136, 0x137, 0x138, 0x139, 0x142, 0x143, 0x15A, 0x15B:
		return 0

	case 0x200:
		return uint8(m.GBA.Irq.IE)
	case 0x201:
		return uint8(m.GBA.Irq.IE >> 8)
	case 0x202:
		return uint8(m.GBA.Irq.IF)
	case 0x203:
		return uint8(m.GBA.Irq.IF >> 8)

	case 0x204:
		return m.Waitstate.Read(0)
	case 0x205:
		return m.Waitstate.Read(1)
	case 0x206, 0x207:
		return 0

	case 0x208:
		return m.GBA.Irq.ReadIME()
	case 0x209, 0x20A, 0x20B, 0x301, 0x302, 0x303, 0x304:
		return 0
	}

	return m.IO[addr]
}

func (m *Memory) Read8(addr uint32) uint32 {
	if badRom := addr >= 0x800_0000 && addr < 0xE00_0000; badRom {
		if addr&0x1FF_FFFF >= m.GBA.Cartridge.RomLength {
			return m.ReadBadRom(addr, 1)
		}
	}

	return uint32(m.Read(addr))
}

func (m *Memory) Read16(addr uint32) uint32 {
	if addr >= 0xE00_0000 {
		v := uint32(m.Read(addr))
		return v | (v << 8)
	}

	addr &^= 1

	switch {
	case addr >= 0xD00_0000:
		if ok := CheckEeprom(m.GBA, addr); ok {
			return uint32(m.GBA.Cartridge.EepromRead())
		}

		offset := (addr - 0x800_0000) & (0x200_0000 - 1)
		if offset >= m.GBA.Cartridge.RomLength {
			return m.ReadBadRom(addr, 2)
		}

	case addr >= 0x800_0000:
		if addr&0x1FF_FFFF >= m.GBA.Cartridge.RomLength {
			return m.ReadBadRom(addr, 2)
		}
	}

	//if ptr, ok := m.ReadPtr(addr); ok {
	//	return uint32(binary.LittleEndian.Uint16((*[4]uint8)(ptr)[:]))
	//}

	v := uint32(m.Read(addr + 0))
	v |= uint32(m.Read(addr+1)) << 8

	return v
}

func (m *Memory) Read32(addr uint32) uint32 {
	if addr >= 0xE00_0000 {
		v := uint32(m.Read(addr))
		return v | (v << 8) | (v << 16) | (v << 24)
	}

	addr &^= 3

	if addr >= 0x800_0000 && addr&0x1FF_FFFF >= m.GBA.Cartridge.RomLength {
		return m.ReadBadRom(addr, 4)
	}

	//if ptr, ok := m.ReadPtr(addr); ok {
	//	return binary.LittleEndian.Uint32((*[4]uint8)(ptr)[:])
	//}

	v := uint32(m.Read(addr + 0))
	v |= uint32(m.Read(addr+1)) << 8
	v |= uint32(m.Read(addr+2)) << 16
	v |= uint32(m.Read(addr+3)) << 24
	return v
}

func (m *Memory) ReadBadRom(addr, size uint32) uint32 {
	switch size {
	case 1:
		return ((addr >> 1) >> ((addr & 1) << 3)) & 0xFF

	case 2:

		if addr&1 != 0 {
			return ((addr >> 1) >> ((addr & 1) << 3)) & 0xFF
		}

		return (addr >> 1) & 0xFFFF

	case 4:
		return (((addr &^ 3) >> 1) & 0xFFFF) | ((((addr &^ 3) + 2) >> 1) << 16)
	default:
		panic("BAD ROM READ USING BYTES READ NOT VALID (1, 2, 4)")
	}
}

func (m *Memory) Write(addr uint32, v uint8, byteWrite bool) {
	m.writeRegions[addr>>24](m, addr, v, byteWrite)
}

func (m *Memory) WriteIO(addr uint32, v uint8) {
	// this addr should be relative. - 0x4000000
	// do not make bg control addrs special, unless you know what the f you are doing
	// VCOUNT is not writable, no touchy

	switch {
	case addr >= 0x60 && addr < 0xB0:
		WriteSound(addr, v, m.GBA.Apu)
		return
	case addr >= 0xB0 && addr < 0xE0:

		addr -= 0xB0
		i := addr / 12
		addr %= 12
		m.GBA.Dma[i].Write(addr, v)
		return

	case addr >= 0x100 && addr < 0x110:

		addr -= 0x100
		i := addr / 4
		idx := addr & 3
		m.GBA.Timers[i].Write(int(idx), v)
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
	case 0x0009:
		m.IO[addr] = v &^ 0b0010_0000 // BG0CNT mask
	case 0x000B:
		m.IO[addr] = v &^ 0b0010_0000 // BG1CNT mask

	case 0x0011:
		m.IO[addr] = v &^ 0b1111_1110 // BG0HOFS mask
	case 0x0013:
		m.IO[addr] = v &^ 0b1111_1110 // BG0VOFS mask
	case 0x0015:
		m.IO[addr] = v &^ 0b1111_1110 // BG1HOFS mask
	case 0x0017:
		m.IO[addr] = v &^ 0b1111_1110 // BG1VOFS mask
	case 0x0019:
		m.IO[addr] = v &^ 0b1111_1110 // BG2HOFS mask
	case 0x001B:
		m.IO[addr] = v &^ 0b1111_1110 // BG2VOFS mask
	case 0x001D:
		m.IO[addr] = v &^ 0b1111_1110 // BG3HOFS mask
	case 0x001F:
		m.IO[addr] = v &^ 0b1111_1110 // BG3VOFS mask

	case 0x0048:
		m.IO[addr] = v & 0x3F // winin
	case 0x0049:
		m.IO[addr] = v & 0x3F // winin
	case 0x004A:
		m.IO[addr] = v & 0x3F // winout
	case 0x004B:
		m.IO[addr] = v & 0x3F // winout

	case 0x0050:
		m.IO[addr] = v // bldcnt
	case 0x0051:
		m.IO[addr] = v &^ 0b1100_0000 // bldcnt
	case 0x0052:
		m.IO[addr] = v &^ 0b1110_0000 // bldalpha
	case 0x0053:
		m.IO[addr] = v &^ 0b1110_0000 // bldalpha

	case 0x130:
		return
	case 0x131:
		return
	case 0x132:
		m.GBA.Keypad.writeCNT(v, false)
	case 0x133:
		m.GBA.Keypad.writeCNT(v, true)

	case 0x200:
		m.GBA.Irq.WriteIE(v, 0)
	case 0x201:
		m.GBA.Irq.WriteIE(v, 1)
	case 0x202:
		m.GBA.Irq.WriteIF(v, 0)
	case 0x203:
		m.GBA.Irq.WriteIF(v, 1)
	case 0x204:
		m.Waitstate.Write(0, v)
	case 0x205:
		m.Waitstate.Write(1, v)
	case 0x206:
		m.Waitstate.Write(2, v)
	case 0x207:
		m.Waitstate.Write(3, v)

	// IME
	case 0x208:
		m.GBA.Irq.WriteIME(v)
		m.IO[addr] = v
	case 0x209:
		return
	case 0x20A:
		return
	case 0x20B:
		return

	case 0x301:
		m.IO[addr] = v & 0x80
		m.GBA.Cpu.Halted = true

	default:
		m.IO[addr] = v
	}

	if addr == 0x0 || addr == 0x1 || (addr >= 0x8 && addr < 0x55) {
		m.GBA.PPU.UpdatePPU(addr, uint32(v))
	}
}

func (m *Memory) Write8(addr uint32, v uint8) {
	m.Write(addr, v, true)
}

func (m *Memory) Write16(addr uint32, v uint16) {
	if addr >= 0xE00_0000 {
		v = v >> ((addr & 1) << 3)
		m.Write(addr, uint8(v), false)
		return
	}

	addr &^= 1

	if addr >= 0xD00_0000 {
		if ok := CheckEeprom(m.GBA, addr); ok {
			m.GBA.Save = true
			m.GBA.Cartridge.EepromWrite(v)
			return
		}
	}

	m.Write(addr+0, uint8(v), false)
	m.Write(addr+1, uint8(v>>8), false)
}

func (m *Memory) Write32(addr uint32, v uint32) {
	if addr >= 0xE00_0000 {
		v = v >> ((addr & 3) << 3)
		m.Write(addr, uint8(v), false)
		return
	}

	addr &^= 3

	m.Write(addr+0, uint8(v), false)
	m.Write(addr+1, uint8(v>>8), false)
	m.Write(addr+2, uint8(v>>16), false)
	m.Write(addr+3, uint8(v>>24), false)
}

func CheckEeprom(gba *GBA, addr uint32) bool {
	if gba.Cartridge.Id != 1 {
		return false
	}

	if addr < 0xD00_0000 || addr >= 0xE00_0000 {
		return false
	}

	if gba.Cartridge.RomLength > 0x1000_0000 && addr < 0xDFF_FF00 {
		return false
	}

	return true
}

func (m *Memory) ReadSoundIO(addr uint32) uint8 {
	switch addr &^ 0b1 {
	case 0x8C:
		return m.ReadOpenBus(addr)
	case 0x8E:
		return m.ReadOpenBus(addr)
	case 0xA0:
		return m.ReadOpenBus(addr)
	case 0xA2:
		return m.ReadOpenBus(addr)
	case 0xA4:
		return m.ReadOpenBus(addr)
	case 0xA6:
		return m.ReadOpenBus(addr)
	case 0xA8:
		return m.ReadOpenBus(addr)
	case 0xAA:
		return m.ReadOpenBus(addr)
	case 0xAC:
		return m.ReadOpenBus(addr)
	case 0xAE:
		return m.ReadOpenBus(addr)
	default:
		return ReadSound(addr, m.GBA.Apu)
	}
}

func (m *Memory) ReadIODirect(addr uint32, size uint32) uint32 {
	switch size {
	case 1:
		return m.ReadIODirectByte(addr)

	case 2:
		return m.ReadIODirectByte(addr+1)<<8 | m.ReadIODirectByte(addr)
	case 4:
		a := m.ReadIODirectByte(addr+3)<<8 | m.ReadIODirectByte(addr+2)
		b := m.ReadIODirectByte(addr+1)<<8 | m.ReadIODirectByte(addr)

		return (a << 16) | b

	default:
		panic("UNKOWN READ IO DIRECT SIZE")
	}
}

func (m *Memory) ReadIODirectByte(addr uint32) uint32 {
	switch addr {
	case 0x4:
		return uint32(m.Dispstat)
	case 0x5:
		return uint32(m.Dispstat >> 8)
	default:
		return uint32(m.IO[addr])
	}
}

func (m *Memory) WritePtr(addr uint32) (unsafe.Pointer, bool) {
	switch regions := addr >> 24; regions {
	case 0x2:
		return unsafe.Add(
			unsafe.Pointer(&m.WRAM1), addr&0x3FFFF,
		), true
	case 0x3:
		return unsafe.Add(
			unsafe.Pointer(&m.WRAM2), addr&0x7FFF,
		), true
	case 0x6:
		addr &= 0x1FFFF
		if addr >= 0x18000 {
			addr -= 0x8000
		}
		return unsafe.Add(
			unsafe.Pointer(&m.VRAM), addr,
		), true

	default:
		return nil, false
	}
}
