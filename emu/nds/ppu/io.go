package ppu

type MasterBright struct {
	Factor uint32
	Mode   uint8
}

const (
	MB_NONE = 0
	MB_UP   = 1
	MB_DOWN = 2
)

func (m *MasterBright) Write(v, b uint8) {
	switch b {
	case 0:
		m.Factor = min(16, uint32(v)&0b11111)

	case 1:
		m.Mode = v >> 6
	}
}

func (m *MasterBright) Read(b uint8) uint8 {
	switch b {
	case 0:
		return uint8(m.Factor)

	case 1:
		return m.Mode << 6
	default:
		return 0
	}
}
