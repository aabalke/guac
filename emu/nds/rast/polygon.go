package rast

import (
	"github.com/aabalke/guac/emu/nds/rast/gl"
	"github.com/aabalke/guac/emu/nds/utils"
)

const (
	PRIM_SEP_TRI    = 0
	PRIM_SEP_QUAD   = 1
	PRIM_TRI_STRIP  = 2
	PRIM_QUAD_STRIP = 3
)

type Polygon struct {
	v                      uint32
	LightsEnabled          [4]bool
	Mode                   uint8
	RenderBack             bool
	RenderFront            bool
	SetNewTranslucentDepth bool
	RenderFarPlanePolygons bool
	RenderBehind1Dot       bool
	DrawEqualDepthPixels   bool
	FogEnabled             bool
	Alpha                  float32
	AlphaV                 uint32
	Id                     uint32

	PrimitiveType uint8
	Vertices      []gl.Vertex

	Texture Texture
}

func (p *Polygon) WriteAttrs(v uint32) {
	p.v = v
	p.LightsEnabled[0] = (v>>0)&1 != 0
	p.LightsEnabled[1] = (v>>1)&1 != 0
	p.LightsEnabled[2] = (v>>2)&1 != 0
	p.LightsEnabled[3] = (v>>3)&1 != 0
	p.Mode = uint8(v>>4) & 0b11
	p.RenderBack = (v>>6)&1 != 0
	p.RenderFront = (v>>7)&1 != 0
	p.SetNewTranslucentDepth = (v>>11)&1 != 0
	p.RenderFarPlanePolygons = (v>>12)&1 != 0
	p.RenderBehind1Dot = (v>>13)&1 != 0
	p.DrawEqualDepthPixels = (v>>14)&1 != 0
	p.FogEnabled = (v>>15)&1 != 0
	p.Alpha = (float32((v >> 16) & 0x1F)) / 31 // 0 is wireframe
	p.AlphaV = (v >> 16) & 0x1F
	p.Id = (v >> 24) & 0x3F
}

const (
	V_16 = 0
	V_10 = 1
	V_XY = 2
	V_XZ = 3
	V_YZ = 4
	V_DF = 5
)

var coordFuncs = [...]func(data []uint32, prev *gl.Vertex) (float32, float32, float32){
	//V_16:
	func(data []uint32, _ *gl.Vertex) (float32, float32, float32) {
		x := utils.Convert16ToFloat(uint16(data[1]), 12)
		y := utils.Convert16ToFloat(uint16(data[1]>>16), 12)
		z := utils.Convert16ToFloat(uint16(data[2]), 12)
		return x, y, z
	},
	//V_10:
	func(data []uint32, _ *gl.Vertex) (float32, float32, float32) {
		x := utils.Convert10ToFloat(uint16(data[1]), 6)
		y := utils.Convert10ToFloat(uint16(data[1]>>10), 6)
		z := utils.Convert10ToFloat(uint16(data[1]>>20), 6)
		return x, y, z
	},
	//V_XY:
	func(data []uint32, prev *gl.Vertex) (float32, float32, float32) {
		x := utils.Convert16ToFloat(uint16(data[1]), 12)
		y := utils.Convert16ToFloat(uint16(data[1]>>16), 12)
		z := prev.Position.Z
		return x, y, z
	},
	//V_XZ:
	func(data []uint32, prev *gl.Vertex) (float32, float32, float32) {
		x := utils.Convert16ToFloat(uint16(data[1]), 12)
		y := prev.Position.Y
		z := utils.Convert16ToFloat(uint16(data[1]>>16), 12)
		return x, y, z
	},
	//V_YZ:
	func(data []uint32, prev *gl.Vertex) (float32, float32, float32) {
		x := prev.Position.X
		y := utils.Convert16ToFloat(uint16(data[1]), 12)
		z := utils.Convert16ToFloat(uint16(data[1]>>16), 12)
		return x, y, z
	},
	//V_DF:
	func(data []uint32, prev *gl.Vertex) (float32, float32, float32) {
		convert := func(v uint32) float32 {
			v &= 0x3FF
			sext := int16(v<<6) >> 6
			f := float32(sext) / (1 << 9)
			return f / 8.0
		}

		x := convert(data[1]) + prev.Position.X
		y := convert(data[1]>>10) + prev.Position.Y
		z := convert(data[1]>>20) + prev.Position.Z
		return x, y, z
	},
}

func (p *Polygon) WriteVertex(data []uint32, g *GeoEngine, method uint8) *gl.Vertex {
	x, y, z := coordFuncs[method](data, g.Vertex)

	if tex := &g.Texture; tex.TransformationMode == 3 {

		vtx := gl.VectorW{
			X: x,
			Y: y,
			Z: z,
		}

		mtx := &g.MtxStacks.Stacks[3].CurrMtx
		tex.S = tex.Sv + vtx.Dot3(mtx.Col(0))
		tex.T = tex.Tv + vtx.Dot3(mtx.Col(1))
	}

	v := p.GetVertex(g, x, y, z)

	p.Vertices = append(p.Vertices, v)
	return &v
}

func (p *Polygon) GetVertex(g *GeoEngine, x, y, z float32) gl.Vertex {
	pos := gl.VectorW{X: x, Y: y, Z: z, W: 1.0}
	output := g.ClipMatrix.MulVectorW(pos)
	clr := g.Color
	clr.A = p.Alpha

	return gl.Vertex{
		Position:   pos,
		Color:      clr,
		S:          g.Texture.S,
		T:          g.Texture.T,
		Output:     output,
		NdsTexture: p.GetTexture(g),
	}
}

func (p *Polygon) GetTexture(g *GeoEngine) *gl.Texture {

	// texture has to be copy
	t := g.Texture

	if t.Format == TEX_FMT_NONE {
		return &gl.Texture{
			Mode:        p.Mode,
			ToonTbl:     &g.ToonTbl,
			IsHighlight: g.Disp3dCnt.HighlightShading,
		}
	}

	cache := &g.TextureCache
	vram := g.Vram

	return &gl.Texture{
		Width:         int(t.SizeS),
		Height:        int(t.SizeT),
		RepeatS:       t.RepeatS,
		RepeatT:       t.RepeatT,
		FlipS:         t.FlipS,
		FlipT:         t.FlipT,
		CachedTexture: cache.Get(vram, &t),
		Mode:          p.Mode,
		ToonTbl:       &g.ToonTbl,
		IsHighlight:   g.Disp3dCnt.HighlightShading,
	}
}

func (p *Polygon) valid1DotDepth(depth float32) bool {

	if p.RenderBehind1Dot {
		return true
	}

	for i := range len(p.Vertices) {
		if p.Vertices[i].Output.W <= depth {
			return true
		}
	}

	return false

}

func (p *Polygon) isAlpha() bool {
	// Check if the texture contains an alpha value (and we're using it)

	if (p.Texture.Format == TEX_FMT_A3I5 || p.Texture.Format == TEX_FMT_A5I3) &&
		(p.Mode == 0 || p.Mode == 2) {
		return true
	}

	return p.Alpha > 0 && p.Alpha < 1
}
