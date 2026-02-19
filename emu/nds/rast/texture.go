package rast

import (
	"github.com/aabalke/guac/emu/nds/rast/gl"
	"github.com/aabalke/guac/emu/nds/utils"
)

const (
	TEX_FMT_NONE = iota
	TEX_FMT_A3I5
	TEX_FMT_4_PAL
	TEX_FMT_16_PAL
	TEX_FMT_256_PAL
	TEX_FMT_4X4
	TEX_FMT_A5I3
	TEX_FMT_DIRECT
)

type Texture struct {
	Sv, Tv             float32
	S, T               float32
	VramOffset         uint32
	RepeatS, RepeatT   bool
	FlipS, FlipT       bool
	SizeS, SizeT       uint32
	Format             uint32
	TransparentZero    bool
	TransformationMode uint32
	PaletteBaseAddr    uint32
	PitchShift         uint32
	param              uint32
}

func (tex *Texture) WriteCoord(v uint32, g *GeoEngine) {
	tex.Sv = utils.Convert16ToFloat(uint16(v), 4)
	tex.Tv = utils.Convert16ToFloat(uint16(v>>16), 4)

	if tex.TransformationMode != 1 {
		tex.S = tex.Sv
		tex.T = tex.Tv
		return
	}

	textureVertex := gl.VectorW{
		X: tex.Sv,
		Y: tex.Tv,
		Z: 1.0 / 16,
		W: 1.0 / 16,
	}

	mtx := &g.MtxStacks.Stacks[3].CurrMtx
	tex.S = textureVertex.Dot(mtx.Col(0))
	tex.T = textureVertex.Dot(mtx.Col(1))
}

func (tex *Texture) WriteParam(v uint32) {
	tex.param = v
	tex.VramOffset = (v & 0xFFFF) * 8
	tex.RepeatS = (v>>16)&1 != 0
	tex.RepeatT = (v>>17)&1 != 0
	tex.FlipS = (v>>18)&1 != 0
	tex.FlipT = (v>>19)&1 != 0

	tex.PitchShift = ((v >> 20) & 0b111) + 3
	tex.SizeS = 8 << ((v >> 20) & 0b111)
	tex.SizeT = 8 << ((v >> 23) & 0b111)
	tex.Format = (v >> 26) & 0b111
	tex.TransparentZero = (v>>29)&1 != 0
	tex.TransformationMode = (v >> 30) & 0b11

	if tex.TransformationMode == 3 {
		//panic("VTX TEXT MODE WHICH I THINK IS GOOD BUT YOU SHOULD CHECK")
	}
}

func (text *Texture) WritePalBase(v uint32) {
	text.PaletteBaseAddr = (v & 0x1FFF)
}
