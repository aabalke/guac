package gpu

import (
	"math"

	"github.com/aabalke/guac/emu/nds/gpu/gl"
	"github.com/aabalke/guac/utils"
)

var commands = [0x100]func(g *GeoEngine, params []uint32){
	0x00: NOP,
	0x10: MTX_MODE,
	0x11: MTX_PUSH,
	0x12: MTX_POP,
	0x13: MTX_STORE,
	0x14: MTX_RESTORE,
	0x15: MTX_IDENTITY,
	0x16: MTX_LOAD_4x4,
	0x17: MTX_LOAD_4x3,
	0x18: MTX_MULT_4x4,
	0x19: MTX_MULT_4x3,
	0x1A: MTX_MULT_3x3,
	0x1B: MTX_SCALE,
	0x1C: MTX_TRANS,
	0x20: COLOR,
	0x21: NORMAL,
	0x22: TEXCOORD,
	0x23: VTX_16,
	0x24: VTX_10,
	0x25: VTX_XY,
	0x26: VTX_XZ,
	0x27: VTX_YZ,
	0x28: VTX_DIFF,
	0x29: POLYGON_ATTR,
	0x2A: TEXIMAGE_PARAM,
	0x2B: PLTT_BASE,
	0x30: DIF_AMB,
	0x31: SPE_EMI,
	0x32: LIGHT_VECTOR,
	0x33: LIGHT_COLOR,
	0x34: SHININESS,
	0x40: BEGIN_VTXS,
	0x41: END_VTXS,
	0x50: SWAP_BUFFERS,
	0x60: VIEWPORT,
	0x70: BOX_TEST,
	0x71: POS_TEST,
	0x72: VEC_TEST,
}

func NOP(g *GeoEngine, params []uint32) {}

func MTX_MODE(g *GeoEngine, params []uint32) {
	g.Mode = params[0] & 3
}

func MTX_PUSH(g *GeoEngine, _ []uint32) {
	switch g.Mode {
	case 0:
		if g.psp > 0 {
			g.StackOverflow = true
		}

		g.Stacks[0].Mtxs[0] = g.Stacks[0].Curr
		g.psp = (g.psp + 1) & 1

	case 1, 2:
		if g.csp > 30 {
			g.StackOverflow = true
		}

		g.Stacks[1].Mtxs[g.csp&31] = g.Stacks[1].Curr
		g.Stacks[2].Mtxs[g.csp&31] = g.Stacks[2].Curr
		g.csp = (g.csp + 1) & 63

	case 3:
		if g.csp > 0 {
			g.StackOverflow = true
		}

		g.Stacks[3].Mtxs[0] = g.Stacks[3].Curr
		g.tsp = (g.tsp + 1) & 1
	}
}

func MTX_POP(g *GeoEngine, params []uint32) {
	switch g.Mode {
	case 0:
		if g.csp == 0 {
			g.StackOverflow = true
		}
		g.Stacks[0].Curr = g.Stacks[0].Mtxs[0]
		g.psp = (g.psp - 1) & 1
		g.UpdateClipMtx = true

	case 1, 2:

		g.csp = int(uint8(int8(g.csp)-(int8(params[0]<<2)>>2)) & 63)
		if g.csp > 30 {
			g.StackOverflow = true
		}

		g.Stacks[1].Curr = g.Stacks[1].Mtxs[g.csp&31]
		g.Stacks[2].Curr = g.Stacks[2].Mtxs[g.csp&31]
		g.UpdateClipMtx = true

	default:
		if g.csp == 0 {
			g.StackOverflow = true
		}
		g.Stacks[3].Curr = g.Stacks[3].Mtxs[0]
		g.tsp = (g.tsp - 1) & 1
	}
}

func MTX_STORE(g *GeoEngine, params []uint32) {
	switch g.Mode {
	case 0:
		g.Stacks[0].Mtxs[0] = g.Stacks[0].Curr
	case 1, 2:

		idx := params[0] & 0x1F
		if idx > 30 {
			g.StackOverflow = true
		}
		g.Stacks[1].Mtxs[idx] = g.Stacks[1].Curr
		g.Stacks[2].Mtxs[idx] = g.Stacks[2].Curr

	case 3:
		g.Stacks[3].Mtxs[0] = g.Stacks[3].Curr
	}
}

func MTX_RESTORE(g *GeoEngine, params []uint32) {
	switch g.Mode {
	case 0:
		g.Stacks[0].Curr = g.Stacks[0].Mtxs[0]
		g.UpdateClipMtx = true

	case 1, 2:

		idx := params[0] & 0x1F
		if idx > 30 {
			g.StackOverflow = true
		}
		g.Stacks[1].Curr = g.Stacks[1].Mtxs[idx]
		g.Stacks[2].Curr = g.Stacks[2].Mtxs[idx]
		g.UpdateClipMtx = true

	case 3:
		g.Stacks[3].Curr = g.Stacks[3].Mtxs[0]
	}
}

func MTX_IDENTITY(g *GeoEngine, _ []uint32) {
	mode := g.Mode

	g.Stacks[mode].Curr = gl.Identity()
	if mode == 2 {
		g.Stacks[1].Curr = gl.Identity()
	}
	if mode != 3 {
		g.UpdateClipMtx = true
	}
}

func MTX_LOAD_4x4(g *GeoEngine, params []uint32) {
	mode := g.Mode
	m := gl.Matrix{}
	arr := gl.AsArray(&m)

	for i := range 16 {
		arr[i] = utils.ConvertToFloat(params[i], 12)
	}

	g.Stacks[mode].Curr = m
	if mode == 2 {
		g.Stacks[1].Curr = m
	}

	if mode != 3 {
		g.UpdateClipMtx = true
	}
}

func MTX_LOAD_4x3(g *GeoEngine, params []uint32) {
	mode := g.Mode
	m := gl.Matrix{X33: 1}
	arr := gl.AsArray(&m)

	j := 0
	for i := range 16 {
		if notLastCol := i&3 != 3; notLastCol {
			arr[i] = utils.ConvertToFloat(params[j], 12)
			j++
		}
	}

	g.Stacks[mode].Curr = m
	if mode == 2 {
		g.Stacks[1].Curr = m
	}

	if mode != 3 {
		g.UpdateClipMtx = true
	}
}

func MTX_MULT_4x4(g *GeoEngine, params []uint32) {
	mode := g.Mode
	m := gl.Matrix{}
	arr := gl.AsArray(&m)

	for i := range 16 {
		arr[i] = utils.ConvertToFloat(params[i], 12)
	}

	g.Stacks[mode].Curr = m.Mul(g.Stacks[mode].Curr)
	if mode == 2 {
		g.Stacks[1].Curr = m.Mul(g.Stacks[1].Curr)
	}

	if mode != 3 {
		g.UpdateClipMtx = true
	}
}

func MTX_MULT_4x3(g *GeoEngine, params []uint32) {
	mode := g.Mode
	m := gl.Matrix{X33: 1}
	arr := gl.AsArray(&m)

	j := 0
	for i := range 16 {
		if notLastCol := i&3 != 3; notLastCol {
			arr[i] = utils.ConvertToFloat(params[j], 12)
			j++
		}
	}

	g.Stacks[mode].Curr = m.Mul(g.Stacks[mode].Curr)
	if mode == 2 {
		g.Stacks[1].Curr = m.Mul(g.Stacks[1].Curr)
	}

	if mode != 3 {
		g.UpdateClipMtx = true
	}
}

func MTX_MULT_3x3(g *GeoEngine, params []uint32) {
	mode := g.Mode
	m := gl.Matrix{X33: 1}
	arr := gl.AsArray(&m)

	j := 0
	for i := range 12 {
		if notLastCol := i&3 != 3; notLastCol {
			arr[i] = utils.ConvertToFloat(params[j], 12)
			j++
		}
	}

	g.Stacks[mode].Curr = m.Mul(g.Stacks[mode].Curr)
	if mode == 2 {
		g.Stacks[1].Curr = m.Mul(g.Stacks[1].Curr)
	}

	if mode != 3 {
		g.UpdateClipMtx = true
	}
}

func MTX_SCALE(g *GeoEngine, params []uint32) {
	mode := g.Mode
	v := gl.Vector{
		X: utils.ConvertToFloat(params[0], 12),
		Y: utils.ConvertToFloat(params[1], 12),
		Z: utils.ConvertToFloat(params[2], 12),
	}

	// no effect on vector matrix - keeps light vector length intact
	if mode == 2 {
		g.Stacks[1].Curr = g.Stacks[1].Curr.Scale(v)
	} else {
		g.Stacks[mode].Curr = g.Stacks[mode].Curr.Scale(v)
	}

	if mode != 3 {
		g.UpdateClipMtx = true
	}
}

func MTX_TRANS(g *GeoEngine, params []uint32) {
	mode := g.Mode
	v := gl.Vector{
		X: utils.ConvertToFloat(params[0], 12),
		Y: utils.ConvertToFloat(params[1], 12),
		Z: utils.ConvertToFloat(params[2], 12),
	}

	g.Stacks[mode].Curr = g.Stacks[mode].Curr.Translate(v)
	if mode == 2 {
		g.Stacks[1].Curr = g.Stacks[1].Curr.Translate(v)
	}

	if mode != 3 {
		g.UpdateClipMtx = true
	}
}

func COLOR(g *GeoEngine, params []uint32) {
	g.Color = Write15BitColor(params[0])
}

func NORMAL(g *GeoEngine, params []uint32) {
	x := utils.Convert10ToFloat(uint16(params[0]), 9)
	y := utils.Convert10ToFloat(uint16(params[0]>>10), 9)
	z := utils.Convert10ToFloat(uint16(params[0]>>20), 9)

	if tex := &g.Texture; tex.TransformationMode == 2 {

		// divide 16 fixes fixed point scaling (normal .9, mtx .12)

		vtx := gl.VectorW{
			X: x / 16,
			Y: y / 16,
			Z: z / 16,
		}

		mtx := &g.Stacks[3].Curr
		tex.S = tex.Sv + vtx.Dot3(mtx.Col(0))
		tex.T = tex.Tv + vtx.Dot3(mtx.Col(1))
	}

	directionalMtx := g.Stacks[2].Curr
	v := gl.Vector{
		X: x,
		Y: y,
		Z: z,
	}

	n := &g.LightData.Normal
	*n = directionalMtx.VecMul3x3(v)

	ld := &g.LightData
	g.Color = ld.EmissionColor

	for i, v := range ld.Lights {

		if !g.ActivePoly.LightsEnabled[i] {
			continue
		}

		diffuseLevel := max(0, -v.Vector.Dot(*n))
		shininessLevel := float32(math.Pow(float64(max(0, -v.HalfVector.Dot(*n))), 2))

		if ld.UseSpecularTbl {
			shininessLevel = ld.ShininessTbl[uint32(shininessLevel)]
		}

		g.Color.R += ld.SpecularColor.R * v.Color.R * shininessLevel
		g.Color.R += ld.DiffuseColor.R * v.Color.R * diffuseLevel
		g.Color.R += ld.AmbientColor.R * v.Color.R

		g.Color.G += ld.SpecularColor.G * v.Color.G * shininessLevel
		g.Color.G += ld.DiffuseColor.G * v.Color.G * diffuseLevel
		g.Color.G += ld.AmbientColor.G * v.Color.G

		g.Color.B += ld.SpecularColor.B * v.Color.B * shininessLevel
		g.Color.B += ld.DiffuseColor.B * v.Color.B * diffuseLevel
		g.Color.B += ld.AmbientColor.B * v.Color.B
	}

	g.Color.R = min(0.99, g.Color.R)
	g.Color.G = min(0.99, g.Color.G)
	g.Color.B = min(0.99, g.Color.B)
	//g.Color.A = 1
}

func TEXCOORD(g *GeoEngine, params []uint32) {
	tex := &g.Texture
	v := params[0]

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

	mtx := &g.Stacks[3].Curr
	tex.S = textureVertex.Dot(mtx.Col(0))
	tex.T = textureVertex.Dot(mtx.Col(1))
}

func VTX_16(g *GeoEngine, params []uint32) {
	x := utils.Convert16ToFloat(uint16(params[0]), 12)
	y := utils.Convert16ToFloat(uint16(params[0]>>16), 12)
	z := utils.Convert16ToFloat(uint16(params[1]), 12)
	v := buildVertex(g, x, y, z, g.ActivePoly.Mode)
	g.ActivePoly.Vertices = append(g.ActivePoly.Vertices, v)
	g.lastVertexPosition = v.Position
}

func VTX_10(g *GeoEngine, params []uint32) {
	x := utils.Convert10ToFloat(uint16(params[0]), 6)
	y := utils.Convert10ToFloat(uint16(params[0]>>10), 6)
	z := utils.Convert10ToFloat(uint16(params[0]>>20), 6)
	v := buildVertex(g, x, y, z, g.ActivePoly.Mode)
	g.ActivePoly.Vertices = append(g.ActivePoly.Vertices, v)
	g.lastVertexPosition = v.Position
}

func VTX_XY(g *GeoEngine, params []uint32) {
	x := utils.Convert16ToFloat(uint16(params[0]), 12)
	y := utils.Convert16ToFloat(uint16(params[0]>>16), 12)
	z := g.lastVertexPosition.Z
	v := buildVertex(g, x, y, z, g.ActivePoly.Mode)
	g.ActivePoly.Vertices = append(g.ActivePoly.Vertices, v)
	g.lastVertexPosition = v.Position
}

func VTX_XZ(g *GeoEngine, params []uint32) {
	x := utils.Convert16ToFloat(uint16(params[0]), 12)
	y := g.lastVertexPosition.Y
	z := utils.Convert16ToFloat(uint16(params[0]>>16), 12)
	v := buildVertex(g, x, y, z, g.ActivePoly.Mode)
	g.ActivePoly.Vertices = append(g.ActivePoly.Vertices, v)
	g.lastVertexPosition = v.Position
}

func VTX_YZ(g *GeoEngine, params []uint32) {
	x := g.lastVertexPosition.X
	y := utils.Convert16ToFloat(uint16(params[0]), 12)
	z := utils.Convert16ToFloat(uint16(params[0]>>16), 12)
	v := buildVertex(g, x, y, z, g.ActivePoly.Mode)
	g.ActivePoly.Vertices = append(g.ActivePoly.Vertices, v)
	g.lastVertexPosition = v.Position
}

func VTX_DIFF(g *GeoEngine, params []uint32) {
	convert := func(v uint32) float32 {
		v &= 0x3FF
		sext := int16(v<<6) >> 6
		f := float32(sext) / (1 << 9)
		return f / 8.0
	}

	x := convert(params[0]) + g.lastVertexPosition.X
	y := convert(params[0]>>10) + g.lastVertexPosition.Y
	z := convert(params[0]>>20) + g.lastVertexPosition.Z

	v := buildVertex(g, x, y, z, g.ActivePoly.Mode)
	g.ActivePoly.Vertices = append(g.ActivePoly.Vertices, v)
	g.lastVertexPosition = v.Position
}

func POLYGON_ATTR(g *GeoEngine, params []uint32) {
	g.pendingPolyAttr = params[0]
}

func TEXIMAGE_PARAM(g *GeoEngine, params []uint32) {
	v := params[0]
	g.Texture.param = v
	g.Texture.VramOffset = (v & 0xFFFF) * 8
	g.Texture.RepeatS = (v>>16)&1 != 0
	g.Texture.RepeatT = (v>>17)&1 != 0
	g.Texture.FlipS = (v>>18)&1 != 0
	g.Texture.FlipT = (v>>19)&1 != 0
	g.Texture.PitchShift = ((v >> 20) & 7) + 3
	g.Texture.SizeS = 8 << ((v >> 20) & 7)
	g.Texture.SizeT = 8 << ((v >> 23) & 7)
	g.Texture.Format = (v >> 26) & 7
	g.Texture.TransparentZero = (v>>29)&1 != 0
	g.Texture.TransformationMode = (v >> 30) & 3
}

func PLTT_BASE(g *GeoEngine, params []uint32) {
	g.Texture.PaletteBaseAddr = (params[0] & 0x1FFF)
}

func DIF_AMB(g *GeoEngine, params []uint32) {
	g.LightData.DiffuseColor = Write15BitColor(params[0])
	g.LightData.AmbientColor = Write15BitColor(params[0] >> 16)

	if setVertex := params[0]&0x8000 != 0; setVertex {
		g.Color = g.LightData.DiffuseColor
	}
}

func SPE_EMI(g *GeoEngine, params []uint32) {
	g.LightData.SpecularColor = Write15BitColor(params[0])
	g.LightData.EmissionColor = Write15BitColor(params[0] >> 16)
	g.LightData.UseSpecularTbl = params[0]&0x8000 != 0
}

func LIGHT_VECTOR(g *GeoEngine, params []uint32) {
	idx := params[0] >> 30
	light := &g.LightData.Lights[idx]
	light.Vector = g.Stacks[2].Curr.VecMul3x3(gl.Vector{
		X: utils.Convert10ToFloat(uint16(params[0]), 9),
		Y: utils.Convert10ToFloat(uint16(params[0]>>10), 9),
		Z: utils.Convert10ToFloat(uint16(params[0]>>20), 9),
	})

	// line of sight vector
	light.HalfVector = light.Vector
	light.HalfVector.Z -= 1
	light.HalfVector.X /= 2
	light.HalfVector.Y /= 2
	light.HalfVector.Z /= 2
}

func LIGHT_COLOR(g *GeoEngine, params []uint32) {
	g.LightData.Lights[params[0]>>30].Color = Write15BitColor(params[0])
}

func SHININESS(g *GeoEngine, params []uint32) {
	sTbl := &g.LightData.ShininessTbl

	var i uint32
	for _, v := range params {
		sTbl[i+0] = float32(v&0xFF) / 0xFF
		sTbl[i+1] = float32((v>>8)&0xFF) / 0xFF
		sTbl[i+2] = float32((v>>16)&0xFF) / 0xFF
		sTbl[i+3] = float32((v>>24)&0xFF) / 0xFF
		i += 4
	}
}

func BEGIN_VTXS(g *GeoEngine, params []uint32) {
	if endPoly := len(g.ActivePoly.Vertices) != 0; endPoly {
		g.ActivePoly.Texture = g.Texture

		if g.ActivePoly.Cull != 0 {
			g.Buffers.bufGeo.Polys = append(g.Buffers.bufGeo.Polys, g.ActivePoly)
		}
		g.ActivePoly.Vertices = []gl.Vertex{}
	}

	v := g.pendingPolyAttr

	g.ActivePoly.v = v
	g.ActivePoly.LightsEnabled[0] = (v>>0)&1 != 0
	g.ActivePoly.LightsEnabled[1] = (v>>1)&1 != 0
	g.ActivePoly.LightsEnabled[2] = (v>>2)&1 != 0
	g.ActivePoly.LightsEnabled[3] = (v>>3)&1 != 0
	g.ActivePoly.Mode = uint8(v>>4) & 3
	g.ActivePoly.RenderBack = (v>>6)&1 != 0
	g.ActivePoly.RenderFront = (v>>7)&1 != 0
	g.ActivePoly.SetNewTranslucentDepth = (v>>11)&1 != 0
	g.ActivePoly.RenderFarPlanePolygons = (v>>12)&1 != 0
	g.ActivePoly.RenderBehind1Dot = (v>>13)&1 != 0
	g.ActivePoly.DrawEqualDepthPixels = (v>>14)&1 != 0
	g.ActivePoly.FogEnabled = (v>>15)&1 != 0
	g.ActivePoly.Alpha = float32((v>>16)&0x1F) / 0x1F // 0 is wireframe
	g.ActivePoly.AlphaV = (v >> 16) & 0x1F
	g.ActivePoly.Id = (v >> 24) & 0x3F
	g.ActivePoly.PrimitiveType = uint8(params[0] & 3)

	switch render := (v >> 6) & 3; render {
	case RENDER_NONE:
		g.ActivePoly.Cull = 0
	case RENDER_BACK:
		g.ActivePoly.Cull = gl.CullFront
	case RENDER_FRNT:
		g.ActivePoly.Cull = gl.CullBack
	case RENDER_BOTH:
		g.ActivePoly.Cull = gl.CullNone
	}
}

func END_VTXS(g *GeoEngine, params []uint32) {}

func SWAP_BUFFERS(g *GeoEngine, params []uint32) {
	g.Buffers.bufGeo.ManualSort = params[0]&1 != 0
	g.Buffers.bufGeo.DepthBufferW = params[0]&(1<<1) != 0
	g.Buffers.SwapSet = true
}

func VIEWPORT(g *GeoEngine, params []uint32) {
	g.Viewport.X1 = uint8(params[0])
	g.Viewport.Y1 = uint8(params[0] >> 8)
	g.Viewport.X2 = uint8(params[0] >> 16)
	g.Viewport.Y2 = uint8(params[0] >> 24)
}

func BOX_TEST(g *GeoEngine, params []uint32) {
	// returns true if any face is within view volume
	// cur do not have "return false if whole volume is located in box" may need
	var (
		x     = utils.Convert16ToFloat(uint16(params[0]), 12)
		y     = utils.Convert16ToFloat(uint16(params[0]>>16), 12)
		z     = utils.Convert16ToFloat(uint16(params[1]), 12)
		w     = x + utils.Convert16ToFloat(uint16(params[1]>>16), 12)
		h     = y + utils.Convert16ToFloat(uint16(params[2]), 12)
		d     = z + utils.Convert16ToFloat(uint16(params[2]>>16), 12)
		verts = []gl.VectorW{
			g.ClipMatrix.MulVectorW(gl.VectorW{X: x, Y: y, Z: z, W: 1}),
			g.ClipMatrix.MulVectorW(gl.VectorW{X: x, Y: h, Z: z, W: 1}),
			g.ClipMatrix.MulVectorW(gl.VectorW{X: w, Y: y, Z: z, W: 1}),
			g.ClipMatrix.MulVectorW(gl.VectorW{X: w, Y: h, Z: z, W: 1}),
			g.ClipMatrix.MulVectorW(gl.VectorW{X: x, Y: y, Z: d, W: 1}),
			g.ClipMatrix.MulVectorW(gl.VectorW{X: x, Y: h, Z: d, W: 1}),
			g.ClipMatrix.MulVectorW(gl.VectorW{X: w, Y: y, Z: d, W: 1}),
			g.ClipMatrix.MulVectorW(gl.VectorW{X: w, Y: h, Z: d, W: 1}),
		}
	)

	g.TestInView = frustumPlaneTest(verts)
}

func POS_TEST(g *GeoEngine, params []uint32) {
	vw := g.ClipMatrix.MulVectorW(gl.VectorW{
		X: utils.Convert16ToFloat(uint16(params[0]), 12),
		Y: utils.Convert16ToFloat(uint16(params[0]>>16), 12),
		Z: utils.Convert16ToFloat(uint16(params[1]), 12),
		W: 1.0,
	})

	g.PosTestData = [4]uint32{
		utils.ConvertFromFloat(vw.X, 12),
		utils.ConvertFromFloat(vw.Y, 12),
		utils.ConvertFromFloat(vw.Z, 12),
		utils.ConvertFromFloat(vw.W, 12),
	}
}

func VEC_TEST(g *GeoEngine, params []uint32) {
	v := g.Stacks[2].Curr.VecMul3x3(gl.Vector{
		X: utils.Convert10ToFloat(uint16(params[0]), 9),
		Y: utils.Convert10ToFloat(uint16(params[0]>>10), 9),
		Z: utils.Convert10ToFloat(uint16(params[0]>>20), 9),
	})
	g.VecTestData = [3]uint16{
		utils.ConvertFromFloat4_0_12(v.X),
		utils.ConvertFromFloat4_0_12(v.Y),
		utils.ConvertFromFloat4_0_12(v.Z),
	}
}
