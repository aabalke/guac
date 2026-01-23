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

func (m *MasterBright) Apply(v uint32) (uint8, uint8, uint8) {

	//r = (r << 3) | (r >> 2)
	//g = (g << 3) | (g >> 2)
	//b = (b << 3) | (b >> 2)

	// takes in 15bit and returns 24bit

	r := ((v) & 0x1F)
	g := ((v >> 5) & 0x1F)
	b := ((v >> 10) & 0x1F)

	switch m.Mode {
	case MB_NONE:
		return bit15tobit24lut[r], bit15tobit24lut[g], bit15tobit24lut[b]

	case MB_UP:
		r += (31 - r) * m.Factor >> 4
		g += (31 - g) * m.Factor >> 4
		b += (31 - b) * m.Factor >> 4

	case MB_DOWN:
		r -= r * m.Factor >> 4
		g -= g * m.Factor >> 4
		b -= b * m.Factor >> 4
	}

	return bit15tobit24lut[r], bit15tobit24lut[g], bit15tobit24lut[b]
}

var bit15tobit24lut = [32]uint8{
	0, 8, 16, 24, 32, 41, 49, 57,
	65, 74, 82, 90, 98, 106, 115, 123,
	131, 139, 148, 156, 164, 172, 180, 189,
	197, 205, 213, 222, 230, 238, 246, 255,
}
