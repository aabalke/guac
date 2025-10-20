package gl

import (
	"image"
	"image/color"
	"math"
	"sync"
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

type RasterizeInfo struct {
	TotalPixels   uint64
	UpdatedPixels uint64
}

func (info RasterizeInfo) Add(other RasterizeInfo) RasterizeInfo {
	return RasterizeInfo{
		info.TotalPixels + other.TotalPixels,
		info.UpdatedPixels + other.UpdatedPixels,
	}
}

type Context struct {
	Width        int
	Height       int
	ColorBuffer  *image.NRGBA
	DepthBuffer  []float64
	ClearColor   Color
	Shader       *Shader
	ReadDepth    bool
	WriteDepth   bool
	WriteColor   bool
	AlphaBlend   bool
	Wireframe    bool
	FrontFace    Face
	Cull         Cull
	LineWidth    float64
	DepthBias    float64
	screenMatrix Matrix
	locks        []sync.Mutex
}

func NewContext(width, height int) *Context {
	dc := &Context{}
	dc.Width = width
	dc.Height = height
	dc.ColorBuffer = image.NewNRGBA(image.Rect(0, 0, width, height))
	dc.DepthBuffer = make([]float64, width*height)
	dc.ClearColor = Transparent
	//dc.Shader = NewSolidColorShader(Identity(), Color{1, 0, 1, 1})
	dc.ReadDepth = true
	dc.WriteDepth = true
	dc.WriteColor = true
	dc.AlphaBlend = true
	dc.Wireframe = false
	dc.FrontFace = FaceCCW
	dc.Cull = CullBack
	dc.LineWidth = 2
	dc.DepthBias = 0
	dc.screenMatrix = Screen(width, height)
	dc.locks = make([]sync.Mutex, 256)
	dc.ClearDepthBuffer()
	return dc
}

func (dc *Context) Image() image.Image {
	return dc.ColorBuffer
}

func (dc *Context) DepthImage() image.Image {
	lo := math.MaxFloat64
	hi := -math.MaxFloat64
	for _, d := range dc.DepthBuffer {
		if d == math.MaxFloat64 {
			continue
		}
		if d < lo {
			lo = d
		}
		if d > hi {
			hi = d
		}
	}

	im := image.NewGray16(image.Rect(0, 0, dc.Width, dc.Height))
	var i int
	for y := 0; y < dc.Height; y++ {
		for x := 0; x < dc.Width; x++ {
			d := dc.DepthBuffer[i]
			t := (d - lo) / (hi - lo)
			if d == math.MaxFloat64 {
				t = 1
			}
			c := color.Gray16{uint16(t * 0xffff)}
			im.SetGray16(x, y, c)
			i++
		}
	}
	return im
}

func (dc *Context) ClearColorBufferWith(color Color) {
	c := color.NRGBA()
	for y := 0; y < dc.Height; y++ {
		i := dc.ColorBuffer.PixOffset(0, y)
		for x := 0; x < dc.Width; x++ {
			dc.ColorBuffer.Pix[i+0] = c.R
			dc.ColorBuffer.Pix[i+1] = c.G
			dc.ColorBuffer.Pix[i+2] = c.B
			dc.ColorBuffer.Pix[i+3] = c.A
			i += 4
		}
	}
}

func (dc *Context) ClearColorBuffer() {
	dc.ClearColorBufferWith(dc.ClearColor)
}

func (dc *Context) ClearDepthBufferWith(value float64) {
	for i := range dc.DepthBuffer {
		dc.DepthBuffer[i] = value
	}
}

func (dc *Context) ClearDepthBuffer() {
	dc.ClearDepthBufferWith(math.MaxFloat64)
}

func edge(a, b, c Vector) float64 {
	return (b.X-c.X)*(a.Y-c.Y) - (b.Y-c.Y)*(a.X-c.X)
}

// these variables remove reallocations
var vert Vertex

func (dc *Context) rasterize(v0, v1, v2 Vertex, s0, s1, s2 Vector) RasterizeInfo {
	var info RasterizeInfo

	// integer bounding box
	minValue := s0.Min(s1.Min(s2)).Floor()
	maxValue := s0.Max(s1.Max(s2)).Ceil()
	x0 := int(minValue.X)
	x1 := int(maxValue.X)
	y0 := int(minValue.Y)
	y1 := int(maxValue.Y)

	// forward differencing variables
	p := Vector{float64(x0) + 0.5, float64(y0) + 0.5, 0}
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

	// iterate over all pixels in bounding box
	for y := y0; y <= y1; y++ {
		var d float64
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
		d = float64(int(d))
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
			info.TotalPixels++
			z := b0*s0.Z + b1*s1.Z + b2*s2.Z
			bz := z + dc.DepthBias
			if dc.ReadDepth && bz > dc.DepthBuffer[i] { // safe w/out lock?
				continue
			}
			// perspective-correct interpolation of vertex data
			b := VectorW{b0 * r0, b1 * r1, b2 * r2, 0}
			b.W = 1 / (b.X + b.Y + b.Z)
			//v := InterpolateVertexes(v0, v1, v2, b)
			vert.InterpolateVertexes(v0, v1, v2, b)
			// invoke fragment shader
			//color := dc.Shader.Fragment(v)
			dc.Shader.Fragment(&vert)
            color := &vert.Color

			if *color == Discard {
				continue
			}
			// update buffers atomically
			//lock := &dc.locks[(x+y)&255]
			//lock.Lock()
			// check depth buffer again
			if bz <= dc.DepthBuffer[i] || !dc.ReadDepth {
				info.UpdatedPixels++
				if dc.WriteDepth {
					// update depth buffer
					dc.DepthBuffer[i] = z
				}
				if dc.WriteColor {
					// update color buffer
					if dc.AlphaBlend && color.A < 1 {
						sr, sg, sb, sa := color.NRGBA().RGBA()
						a := (0xffff - sa) * 0x101
						j := dc.ColorBuffer.PixOffset(x, y)
						dr := &dc.ColorBuffer.Pix[j+0]
						dg := &dc.ColorBuffer.Pix[j+1]
						db := &dc.ColorBuffer.Pix[j+2]
						da := &dc.ColorBuffer.Pix[j+3]
						*dr = uint8((uint32(*dr)*a/0xffff + sr) >> 8)
						*dg = uint8((uint32(*dg)*a/0xffff + sg) >> 8)
						*db = uint8((uint32(*db)*a/0xffff + sb) >> 8)
						*da = uint8((uint32(*da)*a/0xffff + sa) >> 8)
					} else {
						dc.ColorBuffer.SetNRGBA(x, y, color.NRGBA())
					}
				}
			}
			//lock.Unlock()
		}
		w00 += b12
		w01 += b20
		w02 += b01
	}

	return info
}

func (dc *Context) drawClippedTriangle(v0, v1, v2 Vertex) RasterizeInfo {
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
		return RasterizeInfo{}
	}

	// screen coordinates
	s0 := dc.screenMatrix.MulPosition(ndc0)
	s1 := dc.screenMatrix.MulPosition(ndc1)
	s2 := dc.screenMatrix.MulPosition(ndc2)
    return dc.rasterize(v0, v1, v2, s0, s1, s2)
}

func (dc *Context) DrawTriangle(t *Triangle) RasterizeInfo {
	v1 := t.V1
	v2 := t.V2
	v3 := t.V3

	if v1.Outside() || v2.Outside() || v3.Outside() {

		// clip to viewing volume
		triangles := ClipTriangle(NewTriangle(v1, v2, v3))
		var result RasterizeInfo
		for _, t := range triangles {
			info := dc.drawClippedTriangle(t.V1, t.V2, t.V3)
			result = result.Add(info)
		}
		return result
	} else {
		// no need to clip
		return dc.drawClippedTriangle(v1, v2, v3)
	}
}

func (dc *Context) DrawQuad(q *Quad) RasterizeInfo {
	v1 := q.V1
	v2 := q.V2
	v3 := q.V3
	v4 := q.V4

    var result RasterizeInfo

    if v1.Outside() || v2.Outside() || v3.Outside() {

        // clip to viewing volume
        triangles := ClipTriangle(NewTriangle(v1, v2, v3))
        for _, t := range triangles {
            info := dc.drawClippedTriangle(t.V1, t.V2, t.V3)
            result = result.Add(info)
        }

    } else {
        // no need to clip
        result = result.Add(dc.drawClippedTriangle(v1, v2, v3))
    }

    if v1.Outside() || v3.Outside() || v4.Outside() {

        // clip to viewing volume
        triangles := ClipTriangle(NewTriangle(v1, v3, v4))
        var result RasterizeInfo
        for _, t := range triangles {
            //info := dc.drawClippedTriangle(t.V1, t.V2, t.V3)
            info := dc.drawClippedTriangle(t.V1, t.V2, t.V3)
            result = result.Add(info)
        }
    } else {
        // no need to clip
        result = result.Add(dc.drawClippedTriangle(v1, v3, v4))
    }

    return result

}
