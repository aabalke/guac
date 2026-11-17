package gpu

import (
	"github.com/aabalke/guac/emu/nds/gpu/gl"
)

const (
	PRIM_SEP_TRI = iota
	PRIM_SEP_QUAD
	PRIM_TRI_STRIP
	PRIM_QUAD_STRIP
)

const (
	RENDER_NONE = iota
	RENDER_BACK
	RENDER_FRNT
	RENDER_BOTH
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
	Cull                   gl.Cull
	PrimitiveType          uint8
	Vertices               []gl.Vertex
	Texture                Texture
}

func buildVertex(g *GeoEngine, x, y, z float32, mode uint8) gl.Vertex {
	// have to make a copy instead of ref if multi threaded in old version
	t := &g.Texture

	if t.TransformationMode == 3 {

		vtx := gl.VectorW{
			X: x,
			Y: y,
			Z: z,
		}

		mtx := &g.MtxStacks.Stacks[3].Curr
		t.S = t.Sv + vtx.Dot3(mtx.Col(0))
		t.T = t.Tv + vtx.Dot3(mtx.Col(1))
	}

	pos := gl.VectorW{X: x, Y: y, Z: z, W: 1.0}
	v := gl.Vertex{
		Position:      pos,
		Color:         gl.Color{R: g.Color.R, G: g.Color.G, B: g.Color.B, A: 1},
		S:             t.S,
		T:             t.T,
		Output:        g.ClipMatrix.MulVectorW(pos),
		WorldPosition: g.WorldMatrix.MulVectorW(pos),
		NdsTexture: &gl.Texture{
			Mode:        mode,
			ToonTbl:     &g.ToonTbl,
			IsHighlight: g.Disp3dCnt.HighlightShading,
			Param:       t.param,
		},
	}

	if t.Format != TEX_FMT_NONE {
		v.NdsTexture.Width = int(t.SizeS)
		v.NdsTexture.Height = int(t.SizeT)
		v.NdsTexture.RepeatS = t.RepeatS
		v.NdsTexture.RepeatT = t.RepeatT
		v.NdsTexture.FlipS = t.FlipS
		v.NdsTexture.FlipT = t.FlipT
		v.NdsTexture.CachedTexture = g.TextureCache.Get(g.Vram, t)
	}

	return v
}

func (p *Polygon) isAlpha() bool {
	// Check if the texture contains an alpha value (and we're using it)

	if (p.Texture.Format == TEX_FMT_A3I5 || p.Texture.Format == TEX_FMT_A5I3) &&
		(p.Mode == 0 || p.Mode == 2) {
		return true
	}

	return p.Alpha > 0 && p.Alpha < 1
}
