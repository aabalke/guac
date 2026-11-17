package gpu

import (
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/nds/gpu/gl"
)

type Rasterizer struct {
	Viewport   Viewport
	GeoEngine  *GeoEngine
	Buffers    Buffers
	Render     *Render
	ClearColor gl.Color
	RearPlane  RearPlane
	Vram       Vram
	Disp1Dot   Disp1Dot
	Edge       Edge
	Export     *Export
}

type Irq interface {
	SetIRQ(irq uint32)
}

func NewRasterizer(irq Irq) *Rasterizer {
	r := &Rasterizer{}
	r.GeoEngine = NewGeoEngine(&r.Buffers, irq, &r.Vram)
	r.Render = NewRender(r, &r.Buffers, &r.RearPlane)
	r.RearPlane.VRAM = &r.Vram

	for i := range len(r.Edge.Color) {
		r.Edge.Color[i] = gl.Color{A: 1}
	}

	r.Export = NewExport(
		config.Conf.Nds.Export.Directory,
		config.Conf.Nds.Export.ShadowPolys,
		r,
	)

	return r
}

type Disp3dCnt struct {
	Fog                    *gl.Fog
	TextureMapping         bool
	HighlightShading       bool
	AlphaTesting           bool
	AlphaBlending          bool
	AntiAliasing           bool
	EdgeMarking            bool
	ColorRdlinesOverflow   bool
	PolygonRamOverflow     bool
	RearPlaneBitmapEnabled bool
	v                      uint16
}

func (d *Disp3dCnt) Read(b uint8) uint8 {
	return uint8(d.v >> (8 * b))
}

func (d *Disp3dCnt) Write(v, b uint8) {
	if b == 0 {

		d.TextureMapping = (v>>0)&1 != 0
		d.HighlightShading = (v>>1)&1 != 0
		d.AlphaTesting = (v>>2)&1 != 0
		d.AlphaBlending = (v>>3)&1 != 0
		d.AntiAliasing = (v>>4)&1 != 0
		d.EdgeMarking = (v>>5)&1 != 0
		d.Fog.AlphaOnly = (v>>6)&1 != 0
		d.Fog.Enabled = (v>>7)&1 != 0

		d.v &^= 0xFF
		d.v |= uint16(v)
		return
	}

	d.v &^= 0b0100_1111 << 8
	d.v |= (uint16(v&0b0100_1111) << 8)

	d.Fog.Step = 0x400 >> (v & 0xF)
	d.Fog.UpdateBoundaries()

	if (v>>4)&1 != 0 {
		d.ColorRdlinesOverflow = false
		d.v &^= 0b1_0000 << 8
	}

	if (v>>5)&1 != 0 {
		d.PolygonRamOverflow = false
		d.v &^= 0b10_0000 << 8
	}

	d.RearPlaneBitmapEnabled = (v>>6)&1 != 0
}

type Viewport struct {
	X1, Y1, X2, Y2 uint8
}

type Disp1Dot struct {
	param uint16
	V     float64
}

type Edge struct {
	V     [8]uint16
	Color [8]gl.Color
}

func (e *Edge) Write(addr uint32, v uint8) {
	addr -= 0x330

	i := addr / 2
	hi := addr&1 != 0

	c := e.Color[i]

	e.Color[i] = Convert15BitByte(c, v, hi)

	//discard := color.RGBA{}

	//if e.Color[i] != discard {
	//    r, g, b, a := e.Color[i].RGBA()
	//    fmt.Printf("%d COLOR %d %d %d %d\n", i, r, g, b, a)
	//}

	if hi {
		e.V[i] = (e.V[i] & 0x00FF) | (uint16(v) << 8)
	} else {
		e.V[i] = (e.V[i] & 0xFF00) | uint16(v)
	}
}

func (e *Edge) Read(addr uint32) uint8 {
	addr -= 0x330
	return uint8(e.V[addr/2] >> (addr & 1))
}
