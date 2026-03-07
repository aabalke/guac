package rast

import (
	"fmt"
	"math"
	"sort"

	"sync"

	"github.com/aabalke/guac/emu/nds/rast/gl"
)

const (
	WIDTH  = 256
	HEIGHT = 192
)

type Render struct {
	Rasterizer *Rasterizer
	Pixels     Pixels
	Context    *gl.Context
	Buffers    *Buffers
	RearPlane  *RearPlane
	lock       sync.Mutex
}

type Pixels struct {
	PalettesA []uint32
	PalettesB []uint32
	AlphaA    []uint32
	AlphaB    []uint32
}

func (p *Pixels) InitPixels() {
	p.PalettesA = make([]uint32, WIDTH*HEIGHT)
	p.PalettesB = make([]uint32, WIDTH*HEIGHT)
	p.AlphaA = make([]uint32, WIDTH*HEIGHT)
	p.AlphaB = make([]uint32, WIDTH*HEIGHT)
}

func NewRender(rast *Rasterizer, buffers *Buffers, rp *RearPlane) *Render {

	r := &Render{
		Rasterizer: rast,
		Buffers:    buffers,
		Context:    gl.NewContext(WIDTH, HEIGHT),
		RearPlane:  rp,
	}

	r.Pixels.InitPixels()
	r.Context.Shader = gl.NewShader()

	return r
}

func (r *Render) ResetRasterizer() {

	r.Context.AlphaBlending = r.Rasterizer.GeoEngine.Disp3dCnt.AlphaBlending
	r.Context.EdgeEnabled = r.Rasterizer.GeoEngine.Disp3dCnt.EdgeMarking
	r.Context.ClearColor = gl.Transparent

	if !r.RearPlane.Enabled {
		r.Context.EdgeClearId = 0xFF
		r.Context.ClearDepth = uint32(gl.MAX_DEPTH)

		r.Context.ClearBuffers(
			gl.Transparent,
			gl.MAX_DEPTH,
			false,
			0xFF,
			false,
		)
		return
	}

	r.Context.EdgeClearId = r.RearPlane.Id

	if bitmap := r.Rasterizer.GeoEngine.Disp3dCnt.RearPlaneBitmapEnabled; bitmap {
		r.Context.ClearDepth = r.RearPlane.ClearDepth
		r.Context.ClearBuffers(
			gl.Transparent,
			gl.MAX_DEPTH,
			false,
			r.RearPlane.Id,
			false,
		)

		r.ClearBitmapPlane()
		return
	}

	r.Context.ClearDepth = r.RearPlane.ClearDepth
	r.Context.ClearBuffers(
		r.RearPlane.ClearColor,
		float32(r.RearPlane.ClearDepth),
		false,
		r.RearPlane.Id,
		r.RearPlane.FogEnabled,
	)
}

func (r *Render) ClearBitmapPlane() {

	const (
		WIDTH = 256
	)

	dc := r.Context
	rp := &r.Rasterizer.RearPlane

	// could dma copy to buffers if zero offset

	for y := range dc.Height {
		for x := range dc.Width {

			xIdx := (x) //+ int(rp.OffsetX)) & 255
			yIdx := (y) //+ int(rp.OffsetY)) & 255

			i := xIdx + yIdx*WIDTH

			c := rp.Color[i]
			d := float32(rp.Depth[i])
			f := rp.Fog[i]
			dc.ClearBuffersPixel(x, y, c, d, f)
		}
	}
}

func (r *Render) UpdateRender() {

	r.ResetRasterizer()

	buffer := r.Buffers.GetBuffer()
	polygons := buffer.Polys

	r.Context.DepthW = buffer.DepthBufferW

	// solid polygons sorted first, followed by translucent (unless manual sort alpha)

	sort.SliceStable(polygons, func(i, j int) bool {

		isolid := !polygons[i].isAlpha()
		jsolid := !polygons[j].isAlpha()
		if isolid && !jsolid {
			return true
		} else if jsolid && !isolid {
			return false
		}
		// Polygons have the same translucent / solid status

		// Second: sort by "y" solid polygons (or all polygons if
		// alphaYSort is true)
		minY := func(poly *Polygon) float32 {
			m := float32(math.MaxFloat32)
			for i := range len(poly.Vertices) {
                m = min(poly.Vertices[i].Position.Y, m)
			}

			return m
		}

		if isolid || buffer.ManualSort {

			iy := minY(&polygons[i])
			jy := minY(&polygons[j])
			switch {
			case iy < jy:
				return true
			case jy < iy:
				return false
			}
		}
		// Now polygons have the same y
		// FIXME: sort left-to-right? For now, ignore.

		// Polygons have the same sorting properties.
		// Return false so that stable sorting keeps them in the right order
		return false

	})

	for i := range len(polygons) {

		// 1 dot check seems unneeded
		//if !p.valid1DotDepth(r.Rasterizer.Disp1Dot.V) {
		//    return
		//}

		r.RenderPolygon(&polygons[i])
	}

	if r.Rasterizer.GeoEngine.Disp3dCnt.EdgeMarking {
		r.ApplyEdge(buffer.DepthBufferW)
	}


	if r.Rasterizer.GeoEngine.Fog.Enabled {
		r.ApplyFog(buffer.DepthBufferW)
	}

	r.ImageToPixels(*r.Context.Image())
}

func (r *Render) ApplyFog(depthW bool) {

	fog := &r.Rasterizer.GeoEngine.Fog

	for y := range r.Context.Height {
		for x := range r.Context.Width {

			i := x + y*r.Context.Width

			if !r.Context.FogEnabledBuffer[i] {
				continue
			}

			c := (*r.Context.Image())[x+y*WIDTH]

			var depth float32

			// depth is z normalized range 0...1, w is 4096 (W far, W near is 0)
			if depthW {
				depth = r.Context.DepthBufferW[i] * 8
			} else {
				depth = r.Context.DepthBuffer[i] * 0x7FFF
			}

			depth = max(0, min(depth, 0x7FFF))

			r.Context.SetColor(x, y, fog.ApplyFog(c, float64(depth)))

		}
	}
}

func (r *Render) ApplyEdge(depthW bool) {

	for y := range r.Context.Height {
		for x := range r.Context.Width {
			id, ok := r.Context.EdgeId(x, y, depthW)
			if !ok {
				//r.Context.SetColor(x, y, color.White)
				continue
			}

			i := id / 8
			c := r.Rasterizer.Edge.Color[i]
			r.Context.SetColor(x, y, c)
		}
	}
}

func (r *Render) RenderPolygon(p *Polygon) {

	if len(p.Vertices) == 0 {
		return
	}

    if p.Cull == 0 {
        return
    }

    r.Context.Cull = p.Cull
	r.Context.PolygonFogEnabled = p.FogEnabled
	r.Context.DepthEqual = p.DrawEqualDepthPixels
	r.Context.NewTranslucentDepth = p.SetNewTranslucentDepth
	r.Context.PolygonId = p.Id
	r.Context.PolygonOpaque = p.AlphaV == 0x1F || p.AlphaV == 0
	//r.Context.Wireframe = p.AlphaV == 0

	switch p.PrimitiveType {
	case PRIM_SEP_TRI:

		if invalidCnt := len(p.Vertices)%3 != 0; invalidCnt {
			fmt.Printf("Separate Tri Polygon has invalid vert count.\n")
		}

		for i := 0; i < len(p.Vertices); i += 3 {
			r.Context.Shader.SetTexture(p.Vertices[i].NdsTexture)
			if p.Vertices[i].NdsTexture != nil {
				tW := p.Vertices[i].NdsTexture.Width
				tH := p.Vertices[i].NdsTexture.Height

				p.Vertices[i+2].CalcTextureVector(tW, tH)
				p.Vertices[i+1].CalcTextureVector(tW, tH)
				p.Vertices[i+0].CalcTextureVector(tW, tH)
			}

			tri := gl.NewTriangle(
				p.Vertices[i+2],
				p.Vertices[i+1],
				p.Vertices[i+0])

			r.Context.DrawTriangle(tri)
		}

	case PRIM_SEP_QUAD:


		if invalidCnt := len(p.Vertices)%4 != 0; invalidCnt {
			fmt.Printf("Separate Quad Polygon has invalid vert count.\n")
		}

		for i := 0; i < len(p.Vertices); i += 4 {

			r.Context.Shader.SetTexture(p.Vertices[i].NdsTexture)
			if p.Vertices[i].NdsTexture != nil {
				tW := p.Vertices[i].NdsTexture.Width
				tH := p.Vertices[i].NdsTexture.Height

				p.Vertices[i+3].CalcTextureVector(tW, tH)
				p.Vertices[i+2].CalcTextureVector(tW, tH)
				p.Vertices[i+1].CalcTextureVector(tW, tH)
				p.Vertices[i+0].CalcTextureVector(tW, tH)
			}

			quad := gl.NewQuad(
				p.Vertices[i+3],
				p.Vertices[i+2],
				p.Vertices[i+1],
				p.Vertices[i+0])

			r.Context.DrawQuad(quad)
		}

	case PRIM_TRI_STRIP:

		for i := 2; i < len(p.Vertices); i++ {

			r.Context.Shader.SetTexture(p.Vertices[i].NdsTexture)
			if p.Vertices[i].NdsTexture != nil {
				tW := p.Vertices[i].NdsTexture.Width
				tH := p.Vertices[i].NdsTexture.Height

				p.Vertices[i-2].CalcTextureVector(tW, tH)
				p.Vertices[i-1].CalcTextureVector(tW, tH)
				p.Vertices[i-0].CalcTextureVector(tW, tH)
			}

			if clockwise := i&1 == 1; clockwise {
				tri := gl.NewTriangle(
					p.Vertices[i-2],
					p.Vertices[i-1],
					p.Vertices[i-0])

				r.Context.DrawTriangle(tri)
				continue
			}

			tri := gl.NewTriangle(
				p.Vertices[i-0],
				p.Vertices[i-1],
				p.Vertices[i-2])

			r.Context.DrawTriangle(tri)
		}

	case PRIM_QUAD_STRIP:

		for i := 2; i+1 < len(p.Vertices); i += 2 {

			r.Context.Shader.SetTexture(p.Vertices[i].NdsTexture)
			if p.Vertices[i].NdsTexture != nil {
				tW := p.Vertices[i].NdsTexture.Width
				tH := p.Vertices[i].NdsTexture.Height

				p.Vertices[i-2].CalcTextureVector(tW, tH)
				p.Vertices[i-1].CalcTextureVector(tW, tH)
				p.Vertices[i+1].CalcTextureVector(tW, tH)
				p.Vertices[i+0].CalcTextureVector(tW, tH)
			}

			quad := gl.NewQuad(
				p.Vertices[i+0],
				p.Vertices[i+1],
				p.Vertices[i-1],
				p.Vertices[i-2],
			)

			r.Context.DrawQuad(quad)
		}
	}
}

func (r *Render) ImageToPixels(img []gl.Color) {
	//r.lock.Lock()

	for y := range HEIGHT {
		for x := range WIDTH {
			i := x + y*WIDTH
			c := img[i]
			r5 := uint32(min(0x1F, max(0, c.R*0x1F)))
			g5 := uint32(min(0x1F, max(0, c.G*0x1F)))
			b5 := uint32(min(0x1F, max(0, c.B*0x1F)))
			a5 := uint32(min(0x1F, max(0, c.A*0x1F)))

			v := r5 | g5<<5 | b5<<10

			if r.Rasterizer.Buffers.BisRendering {
				r.Pixels.PalettesB[i] = v
				r.Pixels.AlphaB[i] = a5
			} else {
				r.Pixels.PalettesA[i] = v
				r.Pixels.AlphaA[i] = a5
			}
		}
	}

	//r.lock.Unlock()
}
