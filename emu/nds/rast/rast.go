package rast

import (
	"image/color"

	"github.com/aabalke/guac/emu/nds/cpu"
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
    Viewport Viewport
    GeoEngine *GeoEngine
    Buffers Buffers
    Render *Render
    ClearColor gl.Color
    RearPlane RearPlane
    VRAM VRAM
    Disp1Dot Disp1Dot
    Edge Edge
}

type VRAM interface {
    ReadTexture(uint32) uint8
    ReadPalTexture(uint32) uint8
}

func NewRasterizer(vram VRAM, irq *cpu.Irq) *Rasterizer {
    r := &Rasterizer{}
    r.VRAM = vram
    r.GeoEngine = NewGeoEngine(&r.Buffers, irq, vram)
    r.Render = NewRender(r, &r.Buffers, &r.RearPlane)
    r.RearPlane.VRAM = vram

    for i := range len(r.Edge.Color) {
        r.Edge.Color[i] = color.RGBA{A: 0xFF}
    }

    return r
}

type Disp3dCnt struct {
    Fog *gl.Fog
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
		d.TextureMapping = utils.BitEnabled(uint32(v), 0)
		d.HighlightShading = utils.BitEnabled(uint32(v), 1)
		d.AlphaTesting = utils.BitEnabled(uint32(v), 2)
		d.AlphaBlending = utils.BitEnabled(uint32(v), 3)
		d.AntiAliasing = utils.BitEnabled(uint32(v), 4)
		d.EdgeMarking = utils.BitEnabled(uint32(v), 5)
		d.Fog.AlphaOnly = utils.BitEnabled(uint32(v), 6)
		d.Fog.Enabled = utils.BitEnabled(uint32(v), 7)

        d.v &^= 0xFF
        d.v |= uint16(v)
		return
	}

    d.v &^= 0b0100_1111 << 8
	d.v |= (uint16(v & 0b0100_1111) << 8)

    d.Fog.Step = 0x400 >> (v & 0b1111)
    d.Fog.UpdateBoundaries()

    if utils.BitEnabled(uint32(v), 4) {
        d.ColorRdlinesOverflow = false
        d.v &^= 0b1_0000 << 8
    }

    if utils.BitEnabled(uint32(v), 5) {
        d.PolygonRamOverflow = false
        d.v &^= 0b10_0000 << 8
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
    GeoEngine *GeoEngine
    //TestBusy bool
    TestInView bool

    //StackBusy bool
    FifoEntries uint16
    //GXBusy bool

    FifoIrq uint8
}

func (g *GXSTAT) Write(v, b uint8) {
    switch b {
    case 2:
        if errAck := utils.BitEnabled(uint32(v), 7); errAck {
            g.GeoEngine.MtxStacks.Overflow = false
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
       
        // never?
        //if g.TestBusy {
        //    v |= 1
        //}

        if g.TestInView {
            v |= 1 << 1
        }

        return v

    case 1:

        v |= uint8(*g.GeoEngine.MtxStacks.Stacks[1].Pointer) & 0x1F
        v |= uint8(*g.GeoEngine.MtxStacks.Stacks[0].Pointer) << 5

        // never?
        //if g.StackBusy {
        //    v |= 1 << 6
        //}

        if g.GeoEngine.MtxStacks.Overflow {
            v |= 1 << 7
        }

        return v

    case 2:

        v |= uint8(g.FifoEntries)
        return v

    case 3:

        // I believe fifo entries always zero in emulated

        v |= uint8(g.FifoEntries >> 8)

        if underHalf := g.FifoEntries < 128; underHalf {
            v |= 1 << 1
        }

        if empty := g.FifoEntries == 0; empty {
            v |= 1 << 2
        }

        // never?
        //if g.GXBusy {
        //    v |= 1 << 3
        //}

        v |= g.FifoIrq << 5
        return v
    }

    return 0
}

type Disp1Dot struct {
    param uint16
    V float64
}

type Edge struct {
    V [8]uint16
    Color [8]color.Color
}

func (e *Edge) Write(addr uint32, v uint8) {

    addr -= 0x330

    i := addr / 2
    hi := addr & 1 == 1

    c := gl.MakeColor(e.Color[i])

    e.Color[i] = gl.MakeColorColor(Convert15BitByte(c, v, hi))

    //discard := color.RGBA{}

    //if e.Color[i] != discard {
    //    r, g, b, a := e.Color[i].RGBA()
    //    fmt.Printf("%d COLOR %d %d %d %d\n", i, r, g, b, a)
    //}

    if hi {
        e.V[i] &= 0xFF
        e.V[i] |= uint16(v) << 8

    } else {
        e.V[i] &^= 0xFF
        e.V[i] |= uint16(v)
    }
}

func (e *Edge) Read(addr uint32) uint8 {

    addr -= 0x330

    i := addr / 2
    hi := addr & 1 == 1

    if hi {
        return uint8(e.V[i] >> 8)
    }

    return uint8(e.V[i])
}
