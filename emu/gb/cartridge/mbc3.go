package cartridge

import (
	"encoding/binary"
	"fmt"
	"time"
	"unsafe"
)

// rtc uses https://bgb.bircd.org/rtcsave.html
// version VBA-M (48 byte)

type Mbc3 struct {
	Cartridge *Cartridge

	// registers
	RamTimerEnabled bool
	Bank1, Bank2    uint8

	RomBase, RomBase2, RamBase uint32

	MBC30 bool

	time        [5]uint32
	latchedtime [5]uint32

	last int64 // unix milli for sub second accuracy

	PendingLatch bool
	Latched      bool
}

var bitMask = [...]uint8{0x3F, 0x3F, 0x1F, 0xFF, 0xFF}

func NewMbc3(c *Cartridge) *Mbc3 {

	fmt.Printf("Cartridge MBC3 %02X\n", c.Type)

	m := &Mbc3{
		Cartridge: c,
		Bank1:     1,
		MBC30:     c.RomSize > 1<<21 || c.RamSize > 1<<15,
	}

	if buf, err := ReadRam(c.RomPath + ".rtc"); err != nil {
		m.last = time.Now().Unix()
	} else {
		m.Parse(buf)
	}

	m.UpdateAddrs()

	return m
}

func (m *Mbc3) Parse(buf []byte) {

	m.time[0] = uint32(buf[0])
	m.time[1] = uint32(buf[4])
	m.time[2] = uint32(buf[8])
	m.time[3] = uint32(buf[12])
	m.time[4] = uint32(buf[16])
	m.latchedtime[0] = uint32(buf[20])
	m.latchedtime[1] = uint32(buf[24])
	m.latchedtime[2] = uint32(buf[28])
	m.latchedtime[3] = uint32(buf[32])
	m.latchedtime[4] = uint32(buf[36])
	m.last = int64(binary.LittleEndian.Uint64(buf[40:48]))

	m.UpdateSince()
}

func (m *Mbc3) Save() {

	buf := make([]uint8, 48)

	m.UpdateSince()

	buf[0] = uint8(m.time[0])
	buf[4] = uint8(m.time[1])
	buf[8] = uint8(m.time[2])
	buf[12] = uint8(m.time[3])
	buf[16] = uint8(m.time[4])
	buf[20] = uint8(m.latchedtime[0])
	buf[24] = uint8(m.latchedtime[1])
	buf[28] = uint8(m.latchedtime[2])
	buf[32] = uint8(m.latchedtime[3])
	buf[36] = uint8(m.latchedtime[4])
	binary.LittleEndian.PutUint32(buf[40:], uint32(m.last))

	WriteRam(m.Cartridge.RomPath+".rtc", buf)
}

func (m *Mbc3) Read(addr uint16) uint8 {
	switch {
	case addr < 0x4000:
		return m.Cartridge.Data[(m.RomBase|uint32(addr))&m.Cartridge.RomMask]
	case addr < 0x8000:
		return m.Cartridge.Data[(m.RomBase2|uint32(addr-0x4000))&m.Cartridge.RomMask]
	case m.RamTimerEnabled:

		if m.Bank2 >= 8 {

			m.UpdateSince()

			if m.Latched {
				return uint8(m.latchedtime[m.Bank2-8]) & bitMask[m.Bank2-8]
			}
			return uint8(m.time[m.Bank2-8]) & bitMask[m.Bank2-8]
		}

		return m.Cartridge.RamData[(m.RamBase|uint32(addr-0xA000))&m.Cartridge.RamMask]
	default:
		return 0xFF
	}
}

func (m *Mbc3) ReadPtr(addr uint16) unsafe.Pointer {
	switch {
	case addr < 0x4000:
		return unsafe.Pointer(&m.Cartridge.Data[(m.RomBase|uint32(addr))&m.Cartridge.RomMask])
	case addr < 0x8000:
		return unsafe.Pointer(&m.Cartridge.Data[(m.RomBase2|uint32(addr-0x4000))&m.Cartridge.RomMask])
	default:
		return nil
	}
}

func (m *Mbc3) Write(addr uint16, v uint8) {

	switch {
	case addr < 0x2000:
		m.RamTimerEnabled = v&0xF == 0xA

	case addr < 0x4000:

		if m.MBC30 {
			m.Bank1 = max(1, v&0xFF)
		} else {
			m.Bank1 = max(1, v&0x7F)
		}

		m.UpdateAddrs()

	case addr < 0x6000:

		if m.MBC30 {
			m.Bank2 = v & 0xFF
		} else {
			m.Bank2 = v & 0xF
		}

		m.UpdateAddrs()

	case addr < 0x8000:

		if m.PendingLatch && v == 1 {
			m.UpdateSince()
			m.Latched = !m.Latched
			if m.Latched {
				m.latchedtime = m.time
			}
		}

		m.PendingLatch = v == 0

	case m.RamTimerEnabled:

		if m.Bank2 >= 0x08 {
			m.UpdateSince()
			m.time[m.Bank2-8] = uint32(v & bitMask[m.Bank2-8])
			return
		}

		m.Cartridge.RamData[(m.RamBase|uint32(addr-0xA000))&m.Cartridge.RamMask] = v
	}
}

func (m *Mbc3) UpdateAddrs() {
	m.RomBase2 = uint32(m.Bank1) << 14
	m.RamBase = uint32(m.Bank2) << 13
}

func (m *Mbc3) UpdateSince() {
	last := m.last
	now := time.Now().Unix()
	delta := (now - last)
	m.last = now
	m.UpdateTime(uint32(delta))
}

func (m *Mbc3) UpdateTime(cnt uint32) {

	if halted := (m.time[4]>>6)&1 != 0; halted {
		return
	}

	m.time[0] += cnt

	if m.time[0] < 60 {
		return
	}

	m.time[1] += m.time[0] / 60
	m.time[0] %= 60

	if m.time[1] < 60 {
		return
	}

	m.time[2] += m.time[1] / 60
	m.time[1] %= 60

	if m.time[2] < 24 {
		return
	}

	day := uint32(m.time[3]) | (uint32(m.time[4]&1) << 8) + (uint32(m.time[2] / 24))
	m.time[2] %= 24

	if day >= 512 {
		m.time[4] |= 1 << 7
	}

	m.time[3] = day & 0xFF
	m.time[4] &^= 1
	m.time[4] |= (day >> 8) & 1
}
