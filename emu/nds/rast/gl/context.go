package gl

import (
	"fmt"
	"math"
	//"github.com/aabalke/guac/emu/nds/utils"
)

var (
	_ = fmt.Sprintf("")
)

const (
	MAX_DEPTH  = float32(0xFF_FFFF)
	EDGE_THRES = 1

	// higher number is less precise, important for proper ordering with overlapping polys
	// cooking mama requires ~ 0.01
	//DEPTH_PRECISION = 0.01
	// does not work, need to find out why
)

type Face int

const (
	_ Face = iota
	FaceCW
	FaceCCW
)

type Cull int

const (
	_ Cull = iota
	CullNone
	CullFront
	CullBack
)

type ShadowMode int

const (
	_ ShadowMode = iota
	ShadowBack
	ShadowFront
	ShadowRender
)

type Context struct {
	Width            int
	Height           int
	ColorBuffer      []Color
	DepthBuffer      []float32
	DepthBufferW     []float32
	FogEnabledBuffer []bool // bools for if polygon has fog enabled
	EdgeBuffer       []bool
	PolyIdBuffer     []uint32

	ClearColor    Color
	Shader        *Shader
	AlphaBlending bool
	FrontFace     Face
	Cull          Cull
	screenMatrix  Matrix

	PolygonFogEnabled   bool
	NewTranslucentDepth bool

	PolygonId     uint32
	EdgeEnabled   bool
	PolygonOpaque bool

	EdgeClearId uint32
	ClearDepth  uint32

	DepthW     bool
	DepthEqual bool // draw pixels with depth less vs less or equal

	StencilBuffer []bool
}

func NewContext(width, height int) *Context {
	dc := &Context{}
	dc.Width = width
	dc.Height = height
	dc.ColorBuffer = make([]Color, width*height)
	dc.DepthBuffer = make([]float32, width*height)
	dc.DepthBufferW = make([]float32, width*height)
	dc.StencilBuffer = make([]bool, width*height)
	dc.FogEnabledBuffer = make([]bool, width*height)
	dc.EdgeBuffer = make([]bool, width*height)
	dc.PolyIdBuffer = make([]uint32, width*height)
	dc.ClearColor = Transparent
	dc.FrontFace = FaceCCW
	dc.Cull = CullNone
	dc.screenMatrix = Screen(width, height)
	return dc
}

func (dc *Context) Image() *[]Color {
	return &dc.ColorBuffer
}

func (dc *Context) SetColor(x, y int, color Color) {
	dc.ColorBuffer[x+y*dc.Width] = color
}

func (dc *Context) EdgeId(x, y int, depthW bool) (uint32, bool) {

	i := x + y*dc.Width
	if !dc.EdgeBuffer[i] {
		return 0, false
	}

	depths := &dc.DepthBuffer
	if depthW {
		depths = &dc.DepthBufferW
	}

	depth := (*depths)[i]
	id := dc.PolyIdBuffer[i]

	neighbors := [4]int{
		i - 1,
		i + 1,
		i - dc.Width,
		i + dc.Width,
	}

	for j, n := range neighbors {

		if screenOut := (n < 0 ||
			n >= len(dc.PolyIdBuffer) ||
			(j == 0 && n%dc.Width == dc.Width-1) ||
			(j == 1 && n%dc.Width == 0)); screenOut {

			if nid := dc.EdgeClearId; nid == id {
				continue
			}

			if depth < float32(dc.ClearDepth)/MAX_DEPTH {
				return id, true
			}

		} else {

			if nid := dc.PolyIdBuffer[n]; nid == id {
				continue
			}

			if depth < (*depths)[n] {
				return id, true
			}
		}
	}

	return 0, false
}

func (dc *Context) ClearBuffers(c Color, depth float32, edge bool, polyId uint32, fog bool) {

	for i := range dc.ColorBuffer {
		dc.ColorBuffer[i] = c
		dc.DepthBuffer[i] = depth
		dc.DepthBufferW[i] = depth
		dc.EdgeBuffer[i] = edge
		dc.PolyIdBuffer[i] = polyId // invalid value (0 is valid)
		dc.FogEnabledBuffer[i] = fog
	}
}

func (dc *Context) ClearBuffersPixel(x, y int, color Color, depth float32, fog bool) {

	i := x + y*dc.Width

	dc.ColorBuffer[i] = color
	// depth is z normalized range 0...1, w is 4096 (W far, W near is 0)
	dc.DepthBuffer[i] = depth / 0xFF_FFFF
	dc.DepthBufferW[i] = depth / 0x1000
	dc.FogEnabledBuffer[i] = fog
}

func edge(a, b, c Vector) float32 {
	return (b.X-c.X)*(a.Y-c.Y) - (b.Y-c.Y)*(a.X-c.X)
}

// these variables remove reallocations
var vert Vertex

func (dc *Context) rasterize(v0, v1, v2 Vertex, s0, s1, s2 Vector) {

	// integer bounding box
	minValue := s0.Min(s1.Min(s2)).Floor()
	maxValue := s0.Max(s1.Max(s2)).Ceil()
	x0 := int(minValue.X)
	x1 := int(maxValue.X)
	y0 := int(minValue.Y)
	y1 := int(maxValue.Y)

	// forward differencing variables
	p := Vector{float32(x0) + 0.5, float32(y0) + 0.5, 0}
	w00 := edge(s1, s2, p)
	w01 := edge(s2, s0, p)
	w02 := edge(s0, s1, p)
	a01 := s1.Y - s0.Y
	b01 := s0.X - s1.X
	a12 := s2.Y - s1.Y
	b12 := s1.X - s2.X
	a20 := s0.Y - s2.Y
	b20 := s2.X - s0.X

	// reciprocals
	ra := 1 / edge(s0, s1, s2)
	r0 := 1 / v0.Output.W
	r1 := 1 / v1.Output.W
	r2 := 1 / v2.Output.W
	ra12 := 1 / a12
	ra20 := 1 / a20
	ra01 := 1 / a01

	grad0 := float32(math.Hypot(float64(a12), float64(b12))) * ra // for b0
	grad1 := float32(math.Hypot(float64(a20), float64(b20))) * ra // for b1
	grad2 := float32(math.Hypot(float64(a01), float64(b01))) * ra // for b2
	edgeThickness0 := EDGE_THRES * grad0
	edgeThickness1 := EDGE_THRES * grad1
	edgeThickness2 := EDGE_THRES * grad2

	// iterate over all pixels in bounding box
	for y := y0; y <= y1; y++ {
		var d float32
		d0 := -w00 * ra12
		d1 := -w01 * ra20
		d2 := -w02 * ra01
		if w00 < 0 && d0 > d {
			d = d0
		}
		if w01 < 0 && d1 > d {
			d = d1
		}
		if w02 < 0 && d2 > d {
			d = d2
		}
		d = float32(int(d))
		// occurs in pathological cases
		d = max(0, d)

		w0 := w00 + a12*d
		w1 := w01 + a20*d
		w2 := w02 + a01*d
		wasInside := false

		for x := x0 + int(d); x <= x1; x++ {
			b0 := w0 * ra
			b1 := w1 * ra
			b2 := w2 * ra
			w0 += a12
			w1 += a20
			w2 += a01
			// check if inside triangle
			if b0 < 0 || b1 < 0 || b2 < 0 {
				if wasInside {
					break
				}
				continue
			}
			wasInside = true
			// check depth buffer for early abort
			i := y*dc.Width + x
			if i < 0 || i >= len(dc.DepthBuffer) {
				// TODO: clipping roundoff error; fix
				// TODO: could also be from fat lines going off screen
				continue
			}

			z := b0*s0.Z + b1*s1.Z + b2*s2.Z
			// perspective-correct interpolation of vertex data
			b := VectorW{b0 * r0, b1 * r1, b2 * r2, 0}
			b.W = 1 / (b.X + b.Y + b.Z)

			depthBuffer := &dc.DepthBuffer
			depth := z
			if dc.DepthW {
				depthBuffer = &dc.DepthBufferW
				depth = b.W
			}

			if dc.DepthEqual && !(depth+0x200 >= (*depthBuffer)[i] && depth-0x200 < (*depthBuffer)[i]) {
				//if (dc.DepthEqual && depth != (*depthBuffer)[i]) {
				continue
			}

			//if !dc.DepthEqual && depth >= (*depthBuffer)[i]+DEPTH_PRECISION {
			if !dc.DepthEqual && depth >= (*depthBuffer)[i] {
				continue
			}

			vert.InterpolateVertexes(&v0, &v1, &v2, &b)
			dc.Shader.Fragment(&vert)

			if vert.Color == Discard {
				continue
			}

			color := &vert.Color

			if dc.PolygonOpaque {

				if edge := (b0 < edgeThickness0 ||
					b1 < edgeThickness1 ||
					b2 < edgeThickness2); edge {

					dc.EdgeBuffer[i] = dc.EdgeEnabled

					// Wireframe
					//vert.Color = Color{0, 255, 0, 255} // can use to apply wireframe
				} else {
					dc.EdgeBuffer[i] = false
				}
				dc.PolyIdBuffer[i] = dc.PolygonId
				dc.FogEnabledBuffer[i] = dc.PolygonFogEnabled
			} else {
				// When rendering translucent pixels, the old flag in the framebuffer gets ANDed with PolygonAttr.Bit15.
				dc.FogEnabledBuffer[i] = dc.FogEnabledBuffer[i] && dc.PolygonFogEnabled
			}

			if !dc.AlphaBlending ||
				dc.PolygonOpaque ||
				color.A >= 0.999 ||
				dc.NewTranslucentDepth {

				//(*depthBuffer)[i] = utils.FloatFloor(depth, DEPTH_PRECISION)
				(*depthBuffer)[i] = depth
			}

			buf := &dc.ColorBuffer[i]

			if !dc.AlphaBlending || color.A >= 1 || buf.A <= 0 {
                *buf = *color
				continue
			}

			buf.R = min(1, max(0, (color.R*color.A)+(buf.R*(1-color.A))))
			buf.G = min(1, max(0, (color.G*color.A)+(buf.G*(1-color.A))))
			buf.B = min(1, max(0, (color.B*color.A)+(buf.B*(1-color.A))))
			buf.A = min(1, max(0, max(buf.A, color.A)))
		}

		w00 += b12
		w01 += b20
		w02 += b01
	}
}

func (dc *Context) drawClippedTriangle(v0, v1, v2 Vertex) {
	// normalized device coordinates
	ndc0 := v0.Output.DivScalar(v0.Output.W).Vector()
	ndc1 := v1.Output.DivScalar(v1.Output.W).Vector()
	ndc2 := v2.Output.DivScalar(v2.Output.W).Vector()

	// back face culling
	a := (ndc1.X-ndc0.X)*(ndc2.Y-ndc0.Y) - (ndc2.X-ndc0.X)*(ndc1.Y-ndc0.Y)
	if a < 0 {
		v0, v1, v2 = v2, v1, v0
		ndc0, ndc1, ndc2 = ndc2, ndc1, ndc0
	}
	if dc.Cull == CullFront {
		a = -a
	}
	if dc.FrontFace == FaceCW {
		a = -a
	}
	if dc.Cull != CullNone && a <= 0 {
		return
	}

	// screen coordinates
	s0 := dc.screenMatrix.MulPosition(ndc0)
	s1 := dc.screenMatrix.MulPosition(ndc1)
	s2 := dc.screenMatrix.MulPosition(ndc2)
	dc.rasterize(v0, v1, v2, s0, s1, s2)
}

func (dc *Context) DrawTriangle(t *Triangle) {
	v1 := t.V1
	v2 := t.V2
	v3 := t.V3

	if v1.Outside() || v2.Outside() || v3.Outside() {

		// clip to viewing volume
		triangles := ClipTriangle(NewTriangle(v1, v2, v3))
		for _, t := range triangles {
			dc.drawClippedTriangle(t.V1, t.V2, t.V3)
		}
		return
	}

	// no need to clip
	dc.drawClippedTriangle(v1, v2, v3)
}

func (dc *Context) DrawQuad(q *Quad) {
	v1 := q.V1
	v2 := q.V2
	v3 := q.V3
	v4 := q.V4

	if v1.Outside() || v2.Outside() || v3.Outside() {

		// clip to viewing volume
		triangles := ClipTriangle(NewTriangle(v1, v2, v3))
		for _, t := range triangles {
			dc.drawClippedTriangle(t.V1, t.V2, t.V3)
			//result = result.Add(info)
		}

	} else {
		// no need to clip
		dc.drawClippedTriangle(v1, v2, v3)
	}

	if v1.Outside() || v3.Outside() || v4.Outside() {

		// clip to viewing volume
		triangles := ClipTriangle(NewTriangle(v1, v3, v4))
		for _, t := range triangles {
			dc.drawClippedTriangle(t.V1, t.V2, t.V3)
		}

		return
	}

	// no need to clip
	dc.drawClippedTriangle(v1, v3, v4)
}
