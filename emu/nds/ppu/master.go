package ppu

type MasterBright struct {
	Factor uint16
	Mode   uint8
	LUT    [0x8000]uint32
}

const (
	MB_NONE = 0
	MB_UP   = 1
	MB_DOWN = 2
)

func (m *MasterBright) Write(v, b uint8) {

	oldFactor := m.Factor
	oldMode := m.Mode
	switch b {
	case 0:
		m.Factor = uint16(min(16, v&0x1F))

	case 1:
		m.Mode = v >> 6
	}

	if m.Factor != oldFactor || m.Mode != oldMode {
		m.RebuildLUT()
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

func (m *MasterBright) RebuildLUT() {
	for v := range 0x8000 {

		r := uint32(v & 0x1F)
		g := uint32((v >> 5) & 0x1F)
		b := uint32((v >> 10) & 0x1F)

		switch m.Mode {
		case MB_UP:
			r += (31 - r) * uint32(m.Factor) >> 4
			g += (31 - g) * uint32(m.Factor) >> 4
			b += (31 - b) * uint32(m.Factor) >> 4

		case MB_DOWN:
			r -= r * uint32(m.Factor) >> 4
			g -= g * uint32(m.Factor) >> 4
			b -= b * uint32(m.Factor) >> 4
		}

		R := bit15tobit24lut[r]
		G := bit15tobit24lut[g]
		B := bit15tobit24lut[b]

		m.LUT[v] =
			uint32(R) |
				uint32(G)<<8 |
				uint32(B)<<16 |
				0xFF00_0000
	}
}

var bit15tobit24lut = [32]uint8{
	0, 8, 16, 24, 32, 41, 49, 57,
	65, 74, 82, 90, 98, 106, 115, 123,
	131, 139, 148, 156, 164, 172, 180, 189,
	197, 205, 213, 222, 230, 238, 246, 255,
}
