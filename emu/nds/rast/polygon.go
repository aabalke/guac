package rast

import (
	"image/color"

	"github.com/aabalke/guac/emu/nds/rast/gl"
	"github.com/aabalke/guac/emu/nds/utils"
)

type Polygon struct {
	Lights                 [4]bool
	Mode                   uint8
	RenderBack             bool
	RenderFront            bool
	SetNewTranslucentDepth bool
	RenderFarPlanePolygons bool
	RenderBehind1Dot       bool
	DrawEqualDepthPixels   bool
	FogEnabled             bool
	Alpha                  uint32
	Id                     uint32

	PrimitiveType uint8
	Vertices      []gl.Vertex
    Color gl.Color
}

func (p *Polygon) WriteAttrs(v uint32) {
    p.Lights[0] = utils.BitEnabled(v, 0)
    p.Lights[1] = utils.BitEnabled(v, 1)
    p.Lights[2] = utils.BitEnabled(v, 2)
    p.Lights[3] = utils.BitEnabled(v, 3)
    p.Mode = uint8(utils.GetVarData(v, 4, 5))
    p.RenderBack = utils.BitEnabled(v, 6)
    p.RenderFront = utils.BitEnabled(v, 7)
    p.SetNewTranslucentDepth = utils.BitEnabled(v, 11)
    p.RenderFarPlanePolygons = utils.BitEnabled(v, 12)
    p.RenderBehind1Dot = utils.BitEnabled(v, 13)
    p.DrawEqualDepthPixels = utils.BitEnabled(v, 14)
    p.FogEnabled = utils.BitEnabled(v, 15)
    p.Alpha = utils.GetVarData(v, 16, 20)
    p.Id = utils.GetVarData(v, 24, 29)
}

func (p *Polygon) WriteColor(v uint32) {

	r := uint8((v) & 0b11111)
	g := uint8((v >> 5) & 0b11111)
	b := uint8((v >> 10) & 0b11111)

	r = (r << 3) | (r >> 2)
	g = (g << 3) | (g >> 2)
	b = (b << 3) | (b >> 2)

    c := color.RGBA{
        R: r,
        G: g,
        B: b,
    }

    p.Color = gl.MakeColor(c)
}

func (p *Polygon) WriteVtx16(data []uint32, transfromMatrix *gl.Matrix) {

    x := utils.Convert16ToFloat(uint16(data[1]), 12)
    y := utils.Convert16ToFloat(uint16(data[1] >> 16), 12)
    z := utils.Convert16ToFloat(uint16(data[2]), 12)
    w := float64(1.0)

    // clip mask is posMtx * perMtx
    // we just apply posMtx here since perMtx is applied in shader

    vw := transfromMatrix.MulVectorW(gl.VectorW{X: x, Y: y, Z: z, W: w})

    v := gl.Vertex{
        Position: gl.Vector{ X: vw.X, Y: vw.Y, Z: vw.Z },
        Color: p.Color,
        W: vw.W,
    }

    p.Vertices = append(p.Vertices, v)
}

