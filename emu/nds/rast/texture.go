package rast

import (
	"fmt"

	"github.com/aabalke/guac/emu/nds/utils"
)

type Texture struct {
	S, T               float64
	VramOffset         uint32
	RepeatS, RepeatT   bool
	FlipS, FlipT       bool
	SizeS, SizeT       uint32
	Format             uint32
	TransparentZero    bool
	TransformationMode uint8
	PaletteBaseAddr    uint32
}

func (tex *Texture) WriteCoord(v uint32) {

	negS := v&0x80 != 0
	s := utils.ConvertToFloat((v & 0xFFFF) &^ 0x80, 4)
    if negS {
        s *= -1
    }

    tex.S = s

	negT := (v>>16)&0x80 != 0
	t := utils.ConvertToFloat((v >> 16) &^ 0x80, 4)
    if negT {
        t *= -1
    }

    tex.T = t
}

func (tex *Texture) WriteParam(v uint32) {
    tex.VramOffset = (v & 0xFFFF) * 8
    tex.RepeatS = utils.BitEnabled(v, 16)
    tex.RepeatT = utils.BitEnabled(v, 17)
    tex.FlipS = utils.BitEnabled(v, 18)
    tex.FlipT = utils.BitEnabled(v, 19)
    tex.SizeS = 8 << utils.GetVarData(v, 20, 22)
    tex.SizeT = 8 << utils.GetVarData(v, 23, 25)
    tex.Format = utils.GetVarData(v, 26, 28)
    tex.TransparentZero = utils.BitEnabled(v, 29)
    tex.PaletteBaseAddr = utils.GetVarData(v, 30, 31)

    if tex.Format != 0 && tex.Format != 7 {panic(fmt.Sprintf("Unsetup texture format %d", tex.Format))}

}

func (text *Texture) WritePalBase(v uint32) {
    text.PaletteBaseAddr = utils.GetVarData(v, 0, 12)
}
