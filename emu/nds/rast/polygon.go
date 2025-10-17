package rast

import (
	//"image"
	//"image/color"


	"github.com/aabalke/guac/emu/nds/rast/gl"
	"github.com/aabalke/guac/emu/nds/utils"
)

const(
    PRIM_SEP_TRI  = 0
    PRIM_SEP_QUAD = 1
    PRIM_TRI_STRIP = 2
    PRIM_QUAD_STRIP = 3
)

type Polygon struct {
	LightsEnabled          [4]bool
	Mode                   uint8
	RenderBack             bool
	RenderFront            bool
	SetNewTranslucentDepth bool
	RenderFarPlanePolygons bool
	RenderBehind1Dot       bool
	DrawEqualDepthPixels   bool
	FogEnabled             bool
	Alpha                  float64
	Id                     uint32

	PrimitiveType uint8
	Vertices      []gl.Vertex

    Texture Texture
}

func (p *Polygon) WriteAttrs(v uint32) {
    p.LightsEnabled[0] = utils.BitEnabled(v, 0)
    p.LightsEnabled[1] = utils.BitEnabled(v, 1)
    p.LightsEnabled[2] = utils.BitEnabled(v, 2)
    p.LightsEnabled[3] = utils.BitEnabled(v, 3)
    p.Mode = uint8(utils.GetVarData(v, 4, 5))
    p.RenderBack = utils.BitEnabled(v, 6)
    p.RenderFront = utils.BitEnabled(v, 7)
    p.SetNewTranslucentDepth = utils.BitEnabled(v, 11)
    p.RenderFarPlanePolygons = utils.BitEnabled(v, 12)
    p.RenderBehind1Dot = utils.BitEnabled(v, 13)
    p.DrawEqualDepthPixels = utils.BitEnabled(v, 14)
    p.FogEnabled = utils.BitEnabled(v, 15)
    p.Alpha = float64(utils.GetVarData(v, 16, 20)) / 31
    p.Id = utils.GetVarData(v, 24, 29)

    // some 3d examples set alpha to zero, but display solid (Mixed 3d text example)
    if p.Alpha == 0 {
        p.Alpha = 1
    }

    //fmt.Printf("LIGHTS % v\n", p.LightsEnabled)
}

const (
    V_16 = 0
    V_10 = 1
    V_XY = 2
    V_XZ = 3
    V_YZ = 4
    V_DF = 5
)

func (p *Polygon) WriteVertex(data []uint32, g *GeoEngine, method uint8) *gl.Vertex {

    c := g.Color
    //c.A = g.ActivePoly.Alpha

    S := g.Texture.S
    T := g.Texture.T

    var x, y, z float64

    switch method {
    case V_16:
        x = utils.Convert16ToFloat(uint16(data[1]), 12)
        y = utils.Convert16ToFloat(uint16(data[1] >> 16), 12)
        z = utils.Convert16ToFloat(uint16(data[2]), 12)
    case V_10:
        x = utils.Convert10ToFloat(uint16(data[1]), 6)
        y = utils.Convert10ToFloat(uint16(data[1] >> 10), 6)
        z = utils.Convert10ToFloat(uint16(data[1] >> 20), 6)
    case V_XY:
        prev := g.Vertex
        x = utils.Convert16ToFloat(uint16(data[1]), 12)
        y = utils.Convert16ToFloat(uint16(data[1] >> 16), 12)
        z = prev.Position.Z
    case V_XZ:
        prev := g.Vertex
        x = utils.Convert16ToFloat(uint16(data[1]), 12)
        y = prev.Position.Y
        z = utils.Convert16ToFloat(uint16(data[1] >> 16), 12)
    case V_YZ:
        prev := g.Vertex
        x = prev.Position.X
        y = utils.Convert16ToFloat(uint16(data[1]), 12)
        z = utils.Convert16ToFloat(uint16(data[1] >> 16), 12)
    case V_DF:
        prev := g.Vertex
        convert := func(v uint32) float64 {
            v &= 0x3FF
            sext := int32(v << 22) >> 22
            f := float64(sext) / (1 << 9)
            return f / 8.0
        }

        x = convert(data[1])     + prev.Position.X
        y = convert(data[1]>>10) + prev.Position.Y
        z = convert(data[1]>>20) + prev.Position.Z
    }

    v := p.GetVertex(x, y, z, &g.ClipMatrix, c, S, T, &g.StoredNormal)
    v.NdsTexture = p.GetTexture(g.Vram, &g.TextureCache)
    p.Vertices = append(p.Vertices, v)
    return &v
}

func (p *Polygon) GetVertex(x, y, z float64, clipMtx *gl.Matrix, color gl.Color, S, T float64, normal *gl.Vector) gl.Vertex {

    vert := gl.VectorW{X: x,Y: y,Z: z,W: 1.0}
    output := clipMtx.MulVectorW(vert)

    v := gl.Vertex{
        Normal: *normal,
        Position: gl.Vector{ X: x, Y: y, Z: z },
        Color: color,
        W: 1.0,
        S: S,
        T: T,
        Output: output,
    }

    return v
}

func (p *Polygon) GetTexture(vram VRAM, cache *TextureCache) *gl.Texture {

    t := &p.Texture

    if t.Format == TEX_FMT_NONE {
        return nil
    }

    return &gl.Texture{
        Width: int(t.SizeS),
        Height: int(t.SizeT),
        RepeatS: t.RepeatS,
        RepeatT: t.RepeatT,
        FlipS: t.FlipS,
        FlipT: t.FlipT,
        CachedTexture: cache.Get(vram, t),
    }
}
