package gl

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

var (
    _ = fmt.Sprintf("")
)

const (
    MAX_DEPTH = float64(0x7FFF)
    EDGE_THRES = 1
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
	AlphaBlending   bool
	Wireframe    bool
	FrontFace    Face
	Cull         Cull
	LineWidth    float64
	screenMatrix Matrix

    DepthBufferW []float64
    FogEnabledBuffer []bool // bools for if polygon has fog enabled
    PolygonFogEnabled bool
    NewTranslucentDepth bool

    EdgeBuffer []bool
    PolyIdBuffer []uint32
    PolygonId uint32
    EdgeEnabled bool
    PolygonOpaque bool

    EdgeClearId uint32
    ClearDepth uint32

    DepthW bool
}

func NewContext(width, height int) *Context {
	dc := &Context{}
	dc.Width = width
	dc.Height = height
	dc.ColorBuffer = image.NewNRGBA(image.Rect(0, 0, width, height))
	dc.DepthBuffer = make([]float64, width*height)
	dc.DepthBufferW = make([]float64, width*height)
	dc.FogEnabledBuffer = make([]bool, width*height)
	dc.EdgeBuffer = make([]bool, width*height)
	dc.PolyIdBuffer = make([]uint32, width*height)
	dc.ClearColor = Transparent
	dc.ReadDepth = true
	dc.WriteDepth = true
	dc.WriteColor = true
	dc.Wireframe = false
	dc.FrontFace = FaceCCW
	dc.Cull = CullNone
	dc.LineWidth = 2
	dc.screenMatrix = Screen(width, height)
	dc.ClearDepthBufferWith(MAX_DEPTH)
	return dc
}

func (dc *Context) Image() image.Image {
	return dc.ColorBuffer
}

func (dc *Context) SetColor(x, y int, color color.Color) {
    dc.ColorBuffer.Set(x, y, color)
}

func (dc *Context) EdgeId(x, y int, depthW bool) (uint32, bool) {

    i := x + y * dc.Width
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
        i-1       ,
        i+1       ,
        i-dc.Width,
        i+dc.Width,
    }

    for j, n := range neighbors {

        if screenOut := (
            n < 0 ||
            n >= len(dc.PolyIdBuffer) ||
            (j == 0 && n % dc.Width == dc.Width - 1) || 
            (j == 1 && n % dc.Width == 0)); screenOut {

            if nid := dc.EdgeClearId; nid == id {
                continue
            }

            if depth < float64(dc.ClearDepth) / MAX_DEPTH {
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

func (dc *Context) BoolImage() image.Image {

	im := image.NewGray16(image.Rect(0, 0, dc.Width, dc.Height))
	var i int
	for y := 0; y < dc.Height; y++ {
		for x := 0; x < dc.Width; x++ {
			d := dc.EdgeBuffer[i]
            t := 0
			if d {
				t = 1
			}
			c := color.Gray16{uint16(t * 0xffff)}
			im.SetGray16(x, y, c)
			i++
		}
	}
	return im
}


func (dc *Context) BufferImage() image.Image {
	im := image.NewGray16(image.Rect(0, 0, dc.Width, dc.Height))
	var i int
	for y := 0; y < dc.Height; y++ {
		for x := 0; x < dc.Width; x++ {
			d := dc.PolyIdBuffer[i]

            t := 0 
            if d == 8 {
                t = 1
            }

			c := color.Gray16{uint16(t * 0xffff)}
			im.SetGray16(x, y, c)
			i++
		}
	}
	return im
}

func (dc *Context) DepthImage() image.Image {
	lo := MAX_DEPTH
	hi := -MAX_DEPTH
	for _, d := range dc.DepthBuffer {
		if d == MAX_DEPTH {
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
			if d == MAX_DEPTH {
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
		dc.DepthBufferW[i] = value
	}
}

func (dc *Context) ClearEdgeBufferWith(isEdge bool, defaultPolygonId uint32) {
	for i := range len(dc.EdgeBuffer) {
		dc.EdgeBuffer[i] = isEdge
		dc.PolyIdBuffer[i] = defaultPolygonId // invalid value (0 is valid)
	}
}

func (dc *Context) ClearFogBufferWith(value bool) {
	for i := range len(dc.FogEnabledBuffer) {
		dc.FogEnabledBuffer[i] = value
	}
}

func (dc *Context) SetClearBuffers(x, y int, color Color, depth float64, fog bool) {

    i := int(x + y * dc.Width)

    dc.FogEnabledBuffer[i] = fog

    // depth is z normalized range 0...1, w is 4096 (W far, W near is 0)
    dc.DepthBuffer[i]      = depth / 0xFF_FFFF
    dc.DepthBufferW[i]     = depth / 0x1000

	c := color.NRGBA()
    i = dc.ColorBuffer.PixOffset(x, y)
    dc.ColorBuffer.Pix[i+0] = c.R
    dc.ColorBuffer.Pix[i+1] = c.G
    dc.ColorBuffer.Pix[i+2] = c.B
    dc.ColorBuffer.Pix[i+3] = c.A

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

        grad0 := math.Hypot(a12, b12) * ra // for b0
        grad1 := math.Hypot(a20, b20) * ra // for b1
        grad2 := math.Hypot(a01, b01) * ra // for b2
        edgeThickness0 := EDGE_THRES * grad0
        edgeThickness1 := EDGE_THRES * grad1
        edgeThickness2 := EDGE_THRES * grad2

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
			// perspective-correct interpolation of vertex data
			b := VectorW{b0 * r0, b1 * r1, b2 * r2, 0}
			b.W = 1 / (b.X + b.Y + b.Z)

            depthBuffer := &dc.DepthBuffer
            depth := z
            if dc.DepthW {
                depthBuffer = &dc.DepthBufferW
                depth = b.W
            }

			if dc.ReadDepth && depth > (*depthBuffer)[i] {
				continue
			}

			vert.InterpolateVertexes(v0, v1, v2, b)
			dc.Shader.Fragment(&vert)

			if vert.Color == Discard {
				continue
			}

            color := &vert.Color

			if depth <= (*depthBuffer)[i] || !dc.ReadDepth {
				info.UpdatedPixels++

                if dc.PolygonOpaque {

                    if edge := (
                    b0 < edgeThickness0 ||
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

                // this will need to be fixed
                if !(dc.AlphaBlending && color.A < 0.999) {

                //if !dc.AlphaBlending || (dc.AlphaBlending && color.A > 0 && dc.NewTranslucentDepth) {
                    (*depthBuffer)[i] = depth
                }

                if !dc.AlphaBlending || color.A >= 1 {
                    dc.ColorBuffer.SetNRGBA(x, y, color.NRGBA())
                    continue
                }

                sr, sg, sb, sa := color.NRGBA().RGBA()
                a := (0xffff - sa) * 0x101
                j := dc.ColorBuffer.PixOffset(x, y)
                da := &dc.ColorBuffer.Pix[j+3]

                if *da == 0 {
                    dc.ColorBuffer.SetNRGBA(x, y, color.NRGBA())
                    continue
                }

                dr := &dc.ColorBuffer.Pix[j+0]
                dg := &dc.ColorBuffer.Pix[j+1]
                db := &dc.ColorBuffer.Pix[j+2]


                *dr = uint8((uint32(*dr)*a/0xffff + sr) >> 8)
                *dg = uint8((uint32(*dg)*a/0xffff + sg) >> 8)
                *db = uint8((uint32(*db)*a/0xffff + sb) >> 8)
                *da = max(*da, uint8(float64(sa) / 0xFFFF * 0xFF))
                //*da = uint8((uint32(*da)*a/0xffff + sa) >> 8)
            }
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
