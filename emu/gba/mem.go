package gba

import (
	"encoding/binary"
	"time"
	"unsafe"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/bios"
	"github.com/aabalke/guac/emu/gba/gpio"
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
	Sio            *Sio
	Gpio           *gpio.Gpio
	Dispstat       Dispstat
	postflg        uint8

	Timings *Timings

	readRegions  [0x100]func(m *Memory, addr uint32) uint8
	writeRegions [0x100]func(m *Memory, addr uint32, v uint8, byteWrite bool)
}

func NewMemory(gba *GBA) *Memory {
	m := &Memory{GBA: gba}
	m.ProtectedValue = 0xE129F000
	m.Timings = NewTimings(gba.Tick)
	m.Sio = NewSio(gba.Irq, gba.Scheduler)

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
		m.writeRegions[i] = func(_ *Memory, _ uint32, _ uint8, _ bool) {
		}
	}

	m.writeRegions[0x2] = func(m *Memory, addr uint32, v uint8, _ bool) {
		m.WRAM1[addr&0x3_FFFF] = v
	}

	m.writeRegions[0x3] = func(m *Memory, addr uint32, v uint8, _ bool) {
		m.WRAM2[addr&0x7FFF] = v
	}

	m.writeRegions[0x4] = func(m *Memory, addr uint32, v uint8, _ bool) {
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

			if addr+1 < uint32(len(m.VRAM)) {
				m.VRAM[addr+1] = v
			}

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
	m.writeRegions[0x8] = func(m *Memory, addr uint32, v uint8, byteWrite bool) {
		if m.Gpio != nil && addr >= 0x800_00C4 && addr <= 0x800_00C8 {
			m.Gpio.Write(addr, v)
		}
	}

	for i := 0xE; i < 0x10; i++ {
		m.writeRegions[i] = func(m *Memory, addr uint32, v uint8, _ bool) {
			m.GBA.Save = true
			m.GBA.Cartridge.Write(addr&0xFFFF, v)
		}
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
				return uint8(m.ProtectedValue >> ((addr & 3) << 3))
			}

			if pc8 := pc - 8; pc8 == 0xDC || pc8 == 0x134 || pc8 == 0x13C || pc8 == 0x188 {
				m.ProtectedValue = binary.LittleEndian.Uint32((*m.BIOS)[pc:])
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
		return uint8(m.PRAM[addr&0x3FF>>1] >> ((addr & 1) << 3))
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
			return (*m.GBA.Cartridge.Rom)[addr&0x1FFFFFF]
		}
	}

	for i := 0xE; i < 0x10; i++ {
		m.readRegions[i] = func(m *Memory, addr uint32) uint8 {
			return m.GBA.Cartridge.Read(addr & 0xFFFF)
		}
	}
}

func (m *Memory) ReadPtr(addr uint32) unsafe.Pointer {
	switch addr >> 24 {
	case 0:

		// need to avoid protected latch value, skip anything near latches
		if addr >= 0x4000 || addr < 0x200 {
			return nil
		}

		if m.GBA.Cpu.Reg.R[15] >= 0x4000 {
			return nil
		}

		return unsafe.Add(unsafe.Pointer(&(*m.BIOS)[0]), addr&0x3FFF)
	case 2:
		return unsafe.Add(unsafe.Pointer(&m.WRAM1), addr&0x3FFFF)
	case 3:
		return unsafe.Add(unsafe.Pointer(&m.WRAM2), addr&0x7FFF)
	case 5:
		return unsafe.Add(unsafe.Pointer(&m.PRAM), addr&0x3FF)
	case 6:
		addr &= 0x1FFFF
		if addr >= 0x18000 {
			addr -= 0x8000
		}
		return unsafe.Add(unsafe.Pointer(&m.VRAM), addr)

	case 7:
		return unsafe.Add(unsafe.Pointer(&m.OAM), addr&0x3FF)

	case 8, 9, 0xA, 0xB, 0xC, 0xD:
		if addr&0x1FF_FFFF >= uint32(len(*m.GBA.Cartridge.Rom)) {
			return nil
		}
		return unsafe.Add(unsafe.Pointer(&(*m.GBA.Cartridge.Rom)[0]), addr&0x1FF_FFFF)
	}

	return nil
}

func (m *Memory) WritePtr(addr uint32) unsafe.Pointer {
	switch addr >> 24 {
	case 2:
		return unsafe.Add(unsafe.Pointer(&m.WRAM1), addr&0x3FFFF)
	case 3:
		return unsafe.Add(unsafe.Pointer(&m.WRAM2), addr&0x7FFF)
	case 5:
		return unsafe.Add(unsafe.Pointer(&m.PRAM), addr&0x3FF)
	case 6:
		addr &= 0x1FFFF
		if addr >= 0x18000 {
			addr -= 0x8000
		}
		return unsafe.Add(unsafe.Pointer(&m.VRAM), addr)

		// cannot case 7 (oam) rn since need to update oam on every write
	}

	return nil
}

func (m *Memory) Read(addr uint32) uint8 {
	return m.readRegions[addr>>24](m, addr)
}

func (m *Memory) Read8(addr uint32) uint32 {
	if addr < 0x800_0000 || addr >= 0xE00_0000 {
		return uint32(m.Read(addr))
	}
	if addr&0x1FF_FFFF >= uint32(len(*m.GBA.Cartridge.Rom)) {
		return m.ReadBadRom(addr, 1)
	}

	return uint32(m.Read(addr))
}

func (m *Memory) Read16(addr uint32) uint32 {
	if addr >= 0xE00_0000 {
		v := uint32(m.Read(addr))
		return v | (v << 8)
	}

	addr &^= 1

	if addr >= 0xD00_0000 && CheckEeprom(m.GBA, addr) {
		return uint32(m.GBA.Cartridge.EepromRead())
	}

	if addr >= 0x800_0000 && addr&0x1FF_FFFF >= uint32(len(*m.GBA.Cartridge.Rom)) {
		return m.ReadBadRom(addr, 2)
	}

	switch addr {
	case 0x400_0100, 0x400_0102:
		return uint32(m.GBA.Timers[0].Read16(int(addr & 3)))
	case 0x400_0104, 0x400_0106:
		return uint32(m.GBA.Timers[1].Read16(int(addr & 3)))
	case 0x400_0108, 0x400_010A:
		return uint32(m.GBA.Timers[2].Read16(int(addr & 3)))
	case 0x400_010c, 0x400_010E:
		return uint32(m.GBA.Timers[3].Read16(int(addr & 3)))
	case 0x800_00C4, 0x800_00C6, 0x800_00C8:
		if m.Gpio != nil {
			return uint32(m.Gpio.Read(addr))
		}
	}

	if ptr := m.ReadPtr(addr); ptr != nil {
		return uint32(*(*uint16)(ptr))
	}

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

	if addr >= 0x800_0000 && addr&0x1FF_FFFF >= uint32(len(*m.GBA.Cartridge.Rom)) {
		return m.ReadBadRom(addr, 4)
	}

	if ptr := m.ReadPtr(addr); ptr != nil {
		return *(*uint32)(ptr)
	}

	v := uint32(m.Read16(addr + 0))
	v |= uint32(m.Read16(addr+2)) << 16

	return v
}

var openBusDepth int

func (m *Memory) ReadOpenBus(addr uint32) uint8 {
	if m.GBA.Cpu.LastWasDma {
		return uint8(m.GBA.Dma.LatchValue >> ((addr & 3) << 3))
	}

	openBusDepth++

	defer func() {
		openBusDepth--
	}()

	if openBusDepth > 100 {
		panic("open bus depth >= 100")
	}

	if m.GBA.Cpu.Reg.CPSR.T {
		// does pipeline impliment region based thumb mode?
		return uint8(m.Read16(m.GBA.Cpu.Reg.R[15]) >> ((addr & 1) << 3))
	}

	return uint8(m.Read32(m.GBA.Cpu.Reg.R[15]) >> ((addr & 3) << 3))
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

func (m *Memory) ReadIO(addr uint32) uint8 {
	switch {
	case
		addr >= 0x10 && addr < 0x48,
		addr >= 0x4C && addr < 0x50,
		addr >= 0x54 && addr < 0x60,
		addr >= 0xB0 && addr < 0xB8,
		addr >= 0xBC && addr < 0xC4,
		addr >= 0xC8 && addr < 0xD0,
		addr >= 0xD4 && addr < 0xDC,
		addr >= 0xE0 && addr < 0x100:
		return m.ReadOpenBus(addr)
	case addr >= 0x60 && addr < 0xB0:
		switch addr &^ 1 {
		case 0x8C, 0x8E, 0xA0, 0xA2, 0xA4, 0xA6, 0xA8, 0xAA, 0xAC, 0xAE:
			return m.ReadOpenBus(addr)
		default:
			return ReadSound(addr, m.GBA.Apu)
		}
	case addr >= 0xB0 && addr < 0xE0:

		addr -= 0xB0
		i := addr / 12
		addr %= 12
		return m.GBA.Dma.Chs[i].Read(addr)

	case addr >= 0x100 && addr < 0x110:

		addr -= 0x100
		i := addr / 4
		addr &= 3

		return m.GBA.Timers[i].Read(int(addr))
	}

	switch addr {
	case 0x04:
		return uint8(m.Dispstat)
	case 0x05:
		return uint8(m.Dispstat >> 8)
	case 0x07:
		return 0

	case 0x128, 0x129:
		return m.Sio.Read(addr & 1)

	case 0x130, 0x131, 0x132, 0x133:
		return m.GBA.Keypad.Read(addr & 3)

	case 0x136, 0x137, 0x138, 0x139, 0x142, 0x143, 0x15A, 0x15B:
		return 0

	case 0x200, 0x201, 0x202, 0x203, 0x208:
		return m.GBA.Irq.Read(addr)

	case 0x204, 0x205, 0x206, 0x207:
		return m.Timings.ReadWaitstate(addr)

	case 0x300:
		return m.postflg

	case 0x209, 0x20A, 0x20B, 0x301, 0x302, 0x303, 0x304:
		return 0
	}

	return m.IO[addr]
}

func (m *Memory) Write(addr uint32, v uint8, byteWrite bool) {
	m.writeRegions[addr>>24](m, addr, v, byteWrite)
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

	switch addr {
	case 0x400_0100, 0x400_0102:
		m.GBA.Timers[0].Write16(addr&3, v)
		return
	case 0x400_0104, 0x400_0106:
		m.GBA.Timers[1].Write16(addr&3, v)
		return
	case 0x400_0108, 0x400_010A:
		m.GBA.Timers[2].Write16(addr&3, v)
		return
	case 0x400_010C, 0x400_010E:
		m.GBA.Timers[3].Write16(addr&3, v)
		return

	case 0x400_0200, 0x400_0202, 0x400_0208:
		m.GBA.Irq.Write16(addr&0xFFFF, v)
		return
	}

	if ptr := m.WritePtr(addr); ptr != nil {
		*(*uint16)(ptr) = v
		return
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

	switch addr {
	case 0x400_0100:
		m.GBA.Timers[0].Write32(v)
		return
	case 0x400_0104:
		m.GBA.Timers[1].Write32(v)
		return
	case 0x400_0108:
		m.GBA.Timers[2].Write32(v)
		return
	case 0x400_010c:
		m.GBA.Timers[3].Write32(v)
		return
	case 0x400_00A0:
		m.GBA.Apu.FifoA.Copy(v)
		return
	case 0x400_00A4:
		m.GBA.Apu.FifoB.Copy(v)
		return
	}

	if ptr := m.WritePtr(addr); ptr != nil {
		*(*uint32)(ptr) = v
		return
	}

	m.Write16(addr+0, uint16(v))
	m.Write16(addr+2, uint16(v>>16))
}

func CheckEeprom(gba *GBA, addr uint32) bool {
	if notEeprom := gba.Cartridge.Id != 1; notEeprom {
		return false
	}

	if len(*gba.Cartridge.Rom) <= 0x100_0000 {
		return addr >= 0xD00_0000
	}

	return addr >= 0xDFFFF00
}

func (m *Memory) WriteIO(addr uint32, v uint8) {
	switch {
	case addr >= 0x60 && addr < 0xB0:
		WriteSound(addr, v, m.GBA.Apu)
		return
	case addr >= 0xB0 && addr < 0xE0:

		addr -= 0xB0
		i := addr / 12
		addr %= 12

		m.GBA.Dma.Chs[i].Write(addr, v)
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

	case 0x128, 0x129:
		m.Sio.Write(addr&1, v)

	case 0x130, 0x131, 0x132, 0x133:
		m.GBA.Keypad.Write(addr&3, v)
		return

	case 0x200, 0x201, 0x202, 0x203, 0x208:
		m.GBA.Irq.Write8(addr, v)

	case 0x204, 0x205, 0x206, 0x207:
		m.Timings.WriteWaitstate(addr, v)

	case 0x209, 0x20A, 0x20B:
		return

	case 0x300:
		m.postflg |= v & 1

	case 0x301:

		if m.GBA.Cpu.Reg.R[15] < 0x4000 {
			if halt := v&0x80 == 0; halt {
				m.GBA.Cpu.Halted = true
				m.GBA.Tick(1)

			}
		}

	default:
		m.IO[addr] = v
	}

	if addr == 0x0 || addr == 0x1 || (addr >= 0x8 && addr < 0x55) {
		m.GBA.PPU.UpdatePPU(addr, uint32(v))
	}
}
