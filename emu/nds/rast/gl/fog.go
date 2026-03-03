package gl

import (
	"math"
)

type Fog struct {
	Enabled bool

	// gets set for each polygon when rendering
	PolygonFogEnabled bool
	DepthBufferW      bool

	AlphaOnly bool

	Offset   uint16
	Step     uint16
	Color    Color
	Density  [32]uint8
	Boundary [32]float32
}

func (f *Fog) ApplyFog(c Color, depth float64) Color {

	region := 0
	for region < 32 && depth > float64(f.Boundary[region]) {
		region++
	}

	var d0, d1, b0, b1 float64
	if region-1 < 0 {
		d0 = float64(f.Density[0])
		b0 = 0
	} else {
		d0 = float64(f.Density[region-1])
		b0 = float64(f.Boundary[region-1])
	}

	if region < 31 {
		d1 = float64(f.Density[region])
		b1 = float64(f.Boundary[region])
	} else {
		d1 = float64(f.Density[31])
		b1 = float64(f.Boundary[31])
	}

	diff := (depth - b0) / (b1 - b0)

    switch {
    case math.IsInf(diff, 0):
        diff = 1
    case math.IsNaN(diff):
        diff = 0
    }

	var den float64
	if atThreshold := math.IsNaN(diff); atThreshold {
		den = d0 / 0x7F
	} else {
		den = (d0*(1-diff) + d1*diff) / 0x7F
	}

	den = max(0, min(1, den))

    den32 := float32(den)

	if !f.AlphaOnly {
		c.R = (f.Color.R*den32 + c.R*(1-den32))
		c.G = (f.Color.G*den32 + c.G*(1-den32))
		c.B = (f.Color.B*den32 + c.B*(1-den32))

        c.R = min(1, max(0, c.R))
        c.G = min(1, max(0, c.G))
        c.B = min(1, max(0, c.B))
	}

	c.A = (f.Color.A*den32 + c.A*(1-den32))
    c.A = min(1, max(0, c.A))

	return c
}

func (f *Fog) UpdateBoundaries() {
	for i := range len(f.Boundary) {
		f.Boundary[i] = float32(f.Offset) + float32(f.Step)*float32(i+1)
	}
}
