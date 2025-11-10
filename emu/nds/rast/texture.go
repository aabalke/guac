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
    param uint32
}

func (tex *Texture) WriteCoord(v uint32, g *GeoEngine) {
    tex.Sv = utils.Convert16ToFloat(uint16(v), 4)
    tex.Tv = utils.Convert16ToFloat(uint16(v >> 16), 4)

    if tex.TransformationMode != 1 {
        tex.S = tex.Sv
        tex.T = tex.Tv
        return
    }

    textureVertex := gl.VectorW{
        X: tex.Sv,
        Y: tex.Tv,
        Z: 1.0/16,
        W: 1.0/16,
    }

    mtx := &g.MtxStacks.Stacks[3].CurrMtx
    tex.S = textureVertex.Dot(mtx.Col(0))
    tex.T = textureVertex.Dot(mtx.Col(1))
}

func (tex *Texture) WriteParam(v uint32) {
    tex.param = v
    tex.VramOffset = (v & 0xFFFF) * 8
    tex.RepeatS    = utils.BitEnabled(v, 16)
    tex.RepeatT    = utils.BitEnabled(v, 17)
    tex.FlipS      = utils.BitEnabled(v, 18)
    tex.FlipT      = utils.BitEnabled(v, 19)
    tex.PitchShift = utils.GetVarData(v, 20, 22) + 3
    tex.SizeS      = 8 << utils.GetVarData(v, 20, 22)
    tex.SizeT      = 8 << utils.GetVarData(v, 23, 25)
    tex.Format     = utils.GetVarData(v, 26, 28)
    tex.TransparentZero    = utils.BitEnabled(v, 29)
    tex.TransformationMode = utils.GetVarData(v, 30, 31)

    if tex.TransformationMode == 3 {
        panic("VTX TEXT MODE WHICH I THINK IS GOOD BUT YOU SHOULD CHECK")
    }
}

func (text *Texture) WritePalBase(v uint32) {
    text.PaletteBaseAddr = utils.GetVarData(v, 0, 12)
}
