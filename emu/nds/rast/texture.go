package rast

import (
	"fmt"

	"github.com/aabalke/guac/emu/nds/utils"
)

const (
    TEX_FMT_NONE = 0
    TEX_FMT_A3I5 = 1
    TEX_FMT_4_PAL = 2
    TEX_FMT_16_PAL = 3
    TEX_FMT_256_PAL = 4
    TEX_FMT_4X4 = 5
    TEX_FMT_A5I3 = 6
    TEX_FMT_DIRECT = 7
)

type Texture struct {
	S, T               float64
	VramOffset         uint32
	RepeatS, RepeatT   bool
	FlipS, FlipT       bool
	SizeS, SizeT       uint32
	Format             uint32
	TransparentZero    bool
	TransformationMode uint32
	PaletteBaseAddr    uint32
    PitchShift uint32
}

func (tex *Texture) WriteCoord(v uint32) {
    tex.S = utils.Convert16ToFloat(uint16(v), 4)
    tex.T = utils.Convert16ToFloat(uint16(v >> 16), 4)
}

func (tex *Texture) WriteParam(v uint32) {
    tex.VramOffset = (v & 0xFFFF) * 8
    tex.RepeatS = utils.BitEnabled(v, 16)
    tex.RepeatT = utils.BitEnabled(v, 17)
    tex.FlipS = utils.BitEnabled(v, 18)
    tex.FlipT = utils.BitEnabled(v, 19)
    tex.PitchShift = utils.GetVarData(v, 20, 22) + 3
    tex.SizeS = 8 << utils.GetVarData(v, 20, 22)
    tex.SizeT = 8 << utils.GetVarData(v, 23, 25)

    tex.Format = utils.GetVarData(v, 26, 28)
    tex.TransparentZero = utils.BitEnabled(v, 29)
    tex.TransformationMode = utils.GetVarData(v, 30, 31)

    if tex.TransformationMode > 1 {
        fmt.Printf("MODE %02d\n", tex.TransformationMode)
    }
}

func (text *Texture) WritePalBase(v uint32) {
    text.PaletteBaseAddr = utils.GetVarData(v, 0, 12)
}
