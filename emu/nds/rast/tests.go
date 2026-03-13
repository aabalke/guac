package rast

import (
	"github.com/aabalke/guac/emu/nds/rast/gl"
	"github.com/aabalke/guac/emu/nds/utils"
)

func (g *GeoEngine) BoxTest(data []uint32, clipMtx *gl.Matrix) {

	// returns true if any face is within view volume
	// cur do not have "return false if whole volume is located in box" may need
	var (
		x     = utils.Convert16ToFloat(uint16(data[1]), 12)
		y     = utils.Convert16ToFloat(uint16(data[1]>>16), 12)
		z     = utils.Convert16ToFloat(uint16(data[2]), 12)
		w     = x + utils.Convert16ToFloat(uint16(data[2]>>16), 12)
		h     = y + utils.Convert16ToFloat(uint16(data[3]), 12)
		d     = z + utils.Convert16ToFloat(uint16(data[3]>>16), 12)
		verts = []gl.VectorW{
			clipMtx.MulVectorW(gl.VectorW{X: x, Y: y, Z: z, W: 1}),
			clipMtx.MulVectorW(gl.VectorW{X: x, Y: h, Z: z, W: 1}),
			clipMtx.MulVectorW(gl.VectorW{X: w, Y: y, Z: z, W: 1}),
			clipMtx.MulVectorW(gl.VectorW{X: w, Y: h, Z: z, W: 1}),
			clipMtx.MulVectorW(gl.VectorW{X: x, Y: y, Z: d, W: 1}),
			clipMtx.MulVectorW(gl.VectorW{X: x, Y: h, Z: d, W: 1}),
			clipMtx.MulVectorW(gl.VectorW{X: w, Y: y, Z: d, W: 1}),
			clipMtx.MulVectorW(gl.VectorW{X: w, Y: h, Z: d, W: 1}),
		}
	)

	g.GxStat.TestInView = frustumPlaneTest(verts)
}

//go:inline
func frustumPlaneTest(verts []gl.VectorW) bool {

	for plane := range 6 {
		allOutside := true

		for _, v := range verts {
			if insidePlane(v, plane) {
				allOutside = false
				break
			}
		}

		if allOutside {
			return false
		}
	}

	return true

}

//go:inline
func insidePlane(v gl.VectorW, plane int) bool {
	switch plane {
	case 0:
		return v.X >= -v.W // left
	case 1:
		return v.X <= v.W // right
	case 2:
		return v.Y >= -v.W // bottom
	case 3:
		return v.Y <= v.W // top
	case 4:
		return v.Z >= -v.W // near
	case 5:
		return v.Z <= v.W // far
	default:
		return false
	}
}

func (g *GeoEngine) PosTest(data []uint32, clipMtx *gl.Matrix) [4]uint32 {

	x := utils.Convert16ToFloat(uint16(data[1]), 12)
	y := utils.Convert16ToFloat(uint16(data[1]>>16), 12)
	z := utils.Convert16ToFloat(uint16(data[2]), 12)

	vw := clipMtx.MulVectorW(gl.VectorW{X: x, Y: y, Z: z, W: 1.0})

	return [4]uint32{
		utils.ConvertFromFloat(vw.X, 12),
		utils.ConvertFromFloat(vw.Y, 12),
		utils.ConvertFromFloat(vw.Z, 12),
		utils.ConvertFromFloat(vw.W, 12),
	}
}

func (g *GeoEngine) VecTest(data []uint32, dirMtx *gl.Matrix) [3]uint16 {

	x := utils.Convert10ToFloat(uint16(data[1]), 9)
	y := utils.Convert10ToFloat(uint16(data[1]>>10), 9)
	z := utils.Convert10ToFloat(uint16(data[1]>>20), 9)
	v := dirMtx.VecMul3x3(gl.Vector{X: x, Y: y, Z: z})

	return [3]uint16{
		utils.ConvertFromFloat4_0_12(v.X),
		utils.ConvertFromFloat4_0_12(v.Y),
		utils.ConvertFromFloat4_0_12(v.Z),
	}
}
