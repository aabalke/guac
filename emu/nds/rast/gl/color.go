package gl

var (
	Discard     = Color{}
	Transparent = Color{}
	Black       = Color{0, 0, 0, 1}
	White       = Color{1, 1, 1, 1}
)

// Color uses range 0..1

type Color struct {
	R, G, B, A float64
}

func MakeColorFrom15Bit(r, g, b uint8) Color {

	r = (r << 3) | (r >> 2)
	g = (g << 3) | (g >> 2)
	b = (b << 3) | (b >> 2)

	const d = 0xff
	return Color{float64(r) / d, float64(g) / d, float64(b) / d, 1}
}

func (a Color) Add(b Color) Color {
	return Color{a.R + b.R, a.G + b.G, a.B + b.B, a.A + b.A}
}

func (a Color) MulScalar(b float64) Color {
	return Color{a.R * b, a.G * b, a.B * b, a.A * b}
}
