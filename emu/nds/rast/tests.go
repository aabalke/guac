package rast

import (
	"github.com/aabalke/guac/emu/nds/rast/gl"
	"github.com/aabalke/guac/emu/nds/utils"
)

func (g *GeoEngine) BoxTest(data []uint32) {

	x := utils.Convert16ToFloat(uint16(data[1]), 12)
	y := utils.Convert16ToFloat(uint16(data[1]>>16), 12)
	z := utils.Convert16ToFloat(uint16(data[2]), 12)
	w := utils.Convert16ToFloat(uint16(data[2]>>16), 12)
	h := utils.Convert16ToFloat(uint16(data[3]), 12)
	d := utils.Convert16ToFloat(uint16(data[3]>>16), 12)

    posMtx := &g.MtxStacks.Stacks[1].CurrMtx
    vw0 := posMtx.MulVectorW(gl.VectorW{X: x,   Y: y, Z: z, W: 1.0})
    vw1 := posMtx.MulVectorW(gl.VectorW{X: x,   Y: y+h, Z: z, W: 1.0})
    vw2 := posMtx.MulVectorW(gl.VectorW{X: x+w, Y: y, Z: z, W: 1.0})
    vw3 := posMtx.MulVectorW(gl.VectorW{X: x+w, Y: y+h, Z: z, W: 1.0})
    vw4 := posMtx.MulVectorW(gl.VectorW{X: x,   Y: y, Z: z+d, W: 1.0})
    vw5 := posMtx.MulVectorW(gl.VectorW{X: x,   Y: y+h, Z: z+d, W: 1.0})
    vw6 := posMtx.MulVectorW(gl.VectorW{X: x+w, Y: y, Z: z+d, W: 1.0})
    vw7 := posMtx.MulVectorW(gl.VectorW{X: x+w, Y: y+h, Z: z+d, W: 1.0})


    perMtx := &g.MtxStacks.Stacks[0].CurrMtx
    vw0 = perMtx.MulVectorW(vw0)
    vw1 = perMtx.MulVectorW(vw1)
    vw2 = perMtx.MulVectorW(vw2)
    vw3 = perMtx.MulVectorW(vw3)
    vw4 = perMtx.MulVectorW(vw4)
    vw5 = perMtx.MulVectorW(vw5)
    vw6 = perMtx.MulVectorW(vw6)
    vw7 = perMtx.MulVectorW(vw7)

    g.GxStat.TestInView = !(
        vw0.Outside() && 
        vw1.Outside() && 
        vw2.Outside() && 
        vw3.Outside() && 
        vw4.Outside() && 
        vw5.Outside() && 
        vw6.Outside() && 
        vw7.Outside())
}
