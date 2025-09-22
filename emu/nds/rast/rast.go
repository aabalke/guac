package rast

import (
	"github.com/aabalke/guac/emu/nds/rast/gl"
	"github.com/aabalke/guac/emu/nds/utils"
)

const (
    MTX_PJT = 0
    MTX_POS = 1
    MTX_SIM = 2
    MTX_TEX = 3
)

type Rasterizer struct {

	Disp3dCnt Disp3dCnt
    Viewport Viewport

    GeoEngine *GeoEngine
    Buffers Buffers

    Render *Render
    ClearColor gl.Color
    RearPlane RearPlane
    VRAM VRAM
}

type VRAM interface {
    ReadTexture(uint32) uint8
    ReadPalTexture(uint32) uint8
}

func NewRasterizer(vram VRAM) *Rasterizer {
    r := &Rasterizer{}

    r.VRAM = vram

    r.GeoEngine = NewGeoEngine(&r.Buffers)

    r.Render = NewRender(r, &r.Buffers, &r.RearPlane)

    return r
}

type Disp3dCnt struct {
	TextureMapping         bool
	HighlightShading       bool
	AlphaTesting           bool
	AlphaBlending          bool
	AntiAliasing           bool
	EdgeMarking            bool
	FogAlpha               bool
	FogEnabled             bool
	FogDepth               uint8
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
		d.TextureMapping = utils.BitEnabled(uint32(v), 0)
		d.HighlightShading = utils.BitEnabled(uint32(v), 1)
		d.AlphaTesting = utils.BitEnabled(uint32(v), 2)
		d.AlphaBlending = utils.BitEnabled(uint32(v), 3)
		d.AntiAliasing = utils.BitEnabled(uint32(v), 4)
		d.EdgeMarking = utils.BitEnabled(uint32(v), 5)
		d.FogAlpha = utils.BitEnabled(uint32(v), 6)
		d.FogEnabled = utils.BitEnabled(uint32(v), 7)

        d.v &^= 0xFF
        d.v |= uint16(v)
		return
	}

    d.v &^= 0b1100_0000
	d.v |= (uint16(v &^ 0b11) << (8 * b))

	//d.FogDepth need to calc

    if utils.BitEnabled(uint32(v), 4) {
        d.ColorRdlinesOverflow = false
        d.v &^= 0b1_0000
    }

    if utils.BitEnabled(uint32(v), 5) {
        d.PolygonRamOverflow = false
        d.v &^= 0b10_0000
    }

	d.RearPlaneBitmapEnabled = utils.BitEnabled(uint32(v), 6)
}

type Viewport struct {
    X1, Y1, X2, Y2 uint8
}

const (
    IRQ_NEVER = 0
    IRQ_UNDHF = 1
    IRQ_EMPTY = 2
    IRQ_RESVD = 3
)

type GXSTAT struct {
    TestBusy bool
    TestInView bool

    PosStackLevel uint8
    ProjectionStackLevel uint8
    StackBusy bool
    StackError bool
    FifoEntries uint16
    GXBusy bool

    FifoIrq uint8
}

func (g *GXSTAT) Write(v, b uint8) {
    switch b {
    case 2:
        if errAck := utils.BitEnabled(uint32(v), 7); errAck {
            g.StackError = false
            g.ProjectionStackLevel = 0
        }
        return
    case 3:
        g.FifoIrq = v >> 6
        return
    }
}

func (g *GXSTAT) Read(b uint32) uint8 {

    var v uint8

    switch b {
    case 0:
        
        if g.TestBusy {
            v |= 1
        }

        if g.TestInView {
            v |= 1 << 1
        }

        return v

    case 1:

        v |= g.PosStackLevel
        v |= g.ProjectionStackLevel << 5

        if g.StackBusy {
            v |= 1 << 6
        }

        if g.StackError {
            v |= 1 << 7
        }

        return v

    case 2:

        v |= uint8(g.FifoEntries)
        return v

    case 3:

        v |= uint8(g.FifoEntries >> 8)

        if underHalf := g.FifoEntries < 128; underHalf {
            v |= 1 << 1
        }

        if empty := g.FifoEntries == 0; empty {
            v |= 1 << 2
        }

        if g.GXBusy {
            v |= 1 << 3
        }

        v |= g.FifoIrq << 5
        return v
    }

    return 0
}

type RearPlane struct {
    R, G, B uint8
    ClearColor gl.Color
    FogEnabled bool
    Alpha uint32
    Id uint32
    ClearDepth uint32
    OffsetX, OffsetY uint32
}

func (r *RearPlane) Write(addr uint32, v uint8) {
    switch addr {
    case 0x350:

        r.R = v & 0b11111
        r.G &^= 0b111
        r.G = (v >> 5) & 0b111

        r.CalcClearColor(r.R, r.G, r.B)

    case 0x351:

        r.G &= 0b111
        r.G |= (v & 0b11) << 3
        r.B = (v >> 3) & 0b11111
        r.CalcClearColor(r.R, r.G, r.B)

        r.FogEnabled = utils.BitEnabled(uint32(v), 7)
    case 0x352:
        r.Alpha = uint32(v & 0b1_1111)
    case 0x353:
        r.Id = uint32(v & 0b11_1111)

    case 0x354:
        r.ClearDepth &^= 0xFF
        r.ClearDepth |= uint32(v)

    case 0x355:
        r.ClearDepth &^= 0xFF << 8
        r.ClearDepth |= uint32(v &^ 0x80) << 8

    case 0x356:
        r.OffsetX = uint32(v)

    case 0x357:
        r.OffsetY = uint32(v)
    }
}

func (ra *RearPlane) CalcClearColor(r, g, b uint8) {

	r = (r << 3) | (r >> 2)
	g = (g << 3) | (g >> 2)
	b = (b << 3) | (b >> 2)

    ra.ClearColor = gl.Color{
        R: float64(r) / 0xFF,
        G: float64(g) / 0xFF,
        B: float64(b) / 0xFF,
        A: float64(0xFF) / 0xFF,
    }
}
