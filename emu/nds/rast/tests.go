package rast

import (
	"github.com/aabalke/guac/emu/nds/rast/gl"
	"github.com/aabalke/guac/emu/nds/utils"
)

func (g *GeoEngine) BoxTest(data []uint32, clipMtx *gl.Matrix) {

	x := utils.Convert16ToFloat(uint16(data[1]), 12)
	y := utils.Convert16ToFloat(uint16(data[1]>>16), 12)
	z := utils.Convert16ToFloat(uint16(data[2]), 12)
	w := utils.Convert16ToFloat(uint16(data[2]>>16), 12)
	h := utils.Convert16ToFloat(uint16(data[3]), 12)
	d := utils.Convert16ToFloat(uint16(data[3]>>16), 12)

    vw0 := clipMtx.MulVectorW(gl.VectorW{X: x,   Y: y, Z: z, W: 1.0})
    vw1 := clipMtx.MulVectorW(gl.VectorW{X: x,   Y: y+h, Z: z, W: 1.0})
    vw2 := clipMtx.MulVectorW(gl.VectorW{X: x+w, Y: y, Z: z, W: 1.0})
    vw3 := clipMtx.MulVectorW(gl.VectorW{X: x+w, Y: y+h, Z: z, W: 1.0})
    vw4 := clipMtx.MulVectorW(gl.VectorW{X: x,   Y: y, Z: z+d, W: 1.0})
    vw5 := clipMtx.MulVectorW(gl.VectorW{X: x,   Y: y+h, Z: z+d, W: 1.0})
    vw6 := clipMtx.MulVectorW(gl.VectorW{X: x+w, Y: y, Z: z+d, W: 1.0})
    vw7 := clipMtx.MulVectorW(gl.VectorW{X: x+w, Y: y+h, Z: z+d, W: 1.0})

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
