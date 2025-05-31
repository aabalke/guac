package gba

type FlashMode uint32

const (
	Idle        FlashMode = 0x00
	EraseEntire FlashMode = 0x10
	Erase       FlashMode = 0x80
	EnterID     FlashMode = 0x90
	Write       FlashMode = 0xa0
	BankSwitch  FlashMode = 0xb0
	TerminateID FlashMode = 0xf0
)

func (m *Memory) FlashRead(addr uint32) byte {
	if m.flashIDMode { // This is the Flash ROM ID, we return Sanyo ID code
		switch addr {
		case 0x0e000000:
			return 0x62
		case 0x0e000001:
			return 0x13
		}
	} else if m.HasFlash {
		return m.Flash[m.flashBank|(addr&0xffff)]
	} else {
		return m.SRAM[addr&0xffff]
	}
	return 0
}

func (m *Memory) FlashWrite(addr uint32, b byte) {
	switch {
	case m.flashMode == Write:
		m.Flash[m.flashBank|(addr&0xffff)] = b
		m.flashMode = Idle
	case m.flashMode == BankSwitch && addr == 0x0e000000:
		m.flashBank = uint32(b&1) << 16
		m.flashMode = Idle
	case m.SRAM[0x5555] == 0xaa && m.SRAM[0x2aaa] == 0x55:
		if addr == 0xe005555 { // Command for Flash ROM
			switch FlashMode(b) {
			case EraseEntire:
				if m.flashMode == Erase {
					for idx := 0; idx < 0x20000; idx++ {
						m.Flash[idx] = 0xff
					}
					m.flashMode = Idle
				}
			case Erase:
				m.flashMode = Erase
			case EnterID:
				m.flashIDMode = true
			case Write:
				m.flashMode = Write
			case BankSwitch:
				m.flashMode = BankSwitch
			case TerminateID:
				m.flashIDMode = false
			}

			if m.flashMode != Idle || m.flashIDMode {
				m.HasFlash = true
			}
		} else if m.flashMode == Erase && b == 0x30 {
			bankStart := addr & 0xf000
			bankEnd := bankStart + 0x1000
			for idx := bankStart; idx < bankEnd; idx++ {
				m.Flash[m.flashBank|idx] = 0xff
			}
			m.flashMode = Idle
		}
	}
	m.SRAM[addr&0xffff] = b
}
