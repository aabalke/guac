package rast

import (
	"fmt"
	"image"
	"image/color"
	"sort"

	"sync"

	"github.com/aabalke/guac/emu/nds/rast/gl"
)

const (
    WIDTH = 256
    HEIGHT = 192
)

type Render struct {
    Rasterizer *Rasterizer
    Pixels Pixels
	Context *gl.Context
    Buffers *Buffers
    RearPlane *RearPlane
	lock        sync.Mutex
}

type Pixels struct {
    PalettesA []uint32
    PalettesB []uint32
    AlphaA    []float32
    AlphaB    []float32
}

func (p *Pixels) InitPixels() {
    p.PalettesA = make([]uint32, WIDTH*HEIGHT)
    p.PalettesB = make([]uint32, WIDTH*HEIGHT)
    p.AlphaA = make([]float32, WIDTH*HEIGHT)
    p.AlphaB = make([]float32, WIDTH*HEIGHT)
}

func NewRender(rast *Rasterizer, buffers *Buffers, rp *RearPlane) *Render {

    r := &Render{
        Rasterizer: rast,
        Buffers: buffers,
        Context: gl.NewContext(WIDTH, HEIGHT),
        RearPlane: rp,
    }

    r.Pixels.InitPixels()

    //r.Context.Cull = gl.CullFront
    //r.Context.Cull = gl.CullNone
    r.Context.Shader = gl.NewShader()

    return r
}

func (r *Render) ResetRasterizer() {

    r.Context.AlphaBlending = r.Rasterizer.GeoEngine.Disp3dCnt.AlphaBlending
    r.Context.EdgeEnabled   = r.Rasterizer.GeoEngine.Disp3dCnt.EdgeMarking
    r.Context.ClearColor = gl.Transparent

    if !r.RearPlane.Enabled {
        r.Context.EdgeClearId = 0xFF
        r.Context.ClearDepth = uint32(gl.MAX_DEPTH)
        r.Context.ClearColorBufferWith(gl.Transparent)
        r.Context.ClearDepthBufferWith(gl.MAX_DEPTH)
        r.Context.ClearEdgeBufferWith(false, 0xFF)
        r.Context.ClearFogBufferWith(false)
        return
    }

    r.Context.EdgeClearId = r.RearPlane.Id

    if bitmap := r.Rasterizer.GeoEngine.Disp3dCnt.RearPlaneBitmapEnabled; bitmap {
        r.Context.ClearDepth = r.RearPlane.ClearDepth
        r.Context.ClearColorBufferWith(gl.Transparent)
        r.Context.ClearDepthBufferWith(gl.MAX_DEPTH)
        r.Context.ClearEdgeBufferWith(false, r.RearPlane.Id)
        r.Context.ClearFogBufferWith(false)

        r.ClearBitmapPlane()
        return
    }


    r.Context.ClearDepth = r.RearPlane.ClearDepth
    r.Context.ClearColorBufferWith(r.RearPlane.ClearColor)
    r.Context.ClearDepthBufferWith(float64(r.RearPlane.ClearDepth))
    r.Context.ClearEdgeBufferWith(false, r.RearPlane.Id)
    r.Context.ClearFogBufferWith(r.RearPlane.FogEnabled)
}

func (r *Render) ClearBitmapPlane() {

    const (
        SLOT2_BASE = 0x40000
        SLOT3_BASE = 0x60000
    )

    dc := r.Context
    vram := r.Rasterizer.GeoEngine.Vram

    for y := range dc.Height {
        for x := range dc.Width {

            xIdx := (x + int(r.Rasterizer.RearPlane.OffsetX)) //& 255
            yIdx := (y + int(r.Rasterizer.RearPlane.OffsetY)) //& 255

            addr := uint32(xIdx+(yIdx * dc.Width)) * 2
            colorData := uint16(vram.ReadTexture(SLOT2_BASE + addr))
            colorData |= uint16(vram.ReadTexture(SLOT2_BASE + addr + 1)) << 8

            c := gl.MakeColorFrom15Bit(
                uint8(colorData) & 0x1F,
                uint8(colorData >> 5) & 0x1F,
                uint8(colorData >> 10) & 0x1F,
            )

            if transparent := colorData & 0x8000 == 0; transparent {
                c.A = 0
            }

            depthData := uint16(vram.ReadTexture(SLOT3_BASE + addr))
            depthData |= uint16(vram.ReadTexture(SLOT3_BASE + addr + 1)) << 8

            depth := uint32(depthData &^ 0x8000)
            //depth = (depth * 0x200) + ((depth + 1)/ 0x8000) * 0x1FF //gbatek wrong
            depth = (depth * 0x200) + 0x1FF // desmume / ndsemu say this correct

            // debug
            //c.R = float64(depth) / 0xFF_FFFF
            //c.G = float64(depth) / 0xFF_FFFF
            //c.B = float64(depth) / 0xFF_FFFF
            //c.A = 1

            fog := depthData & 0x8000 != 0

            dc.SetClearBuffers(x, y, c, float64(depth), fog)
        }
    }
}

func (r *Render) UpdateRender() {

    r.ResetRasterizer()

    polygons, manualSortAlpha, depthW := r.Buffers.GetPolygons()

    r.Context.DepthW = depthW

    sort.SliceStable(polygons, func(i, j int) bool {

        pi := polygons[i]
        pj := polygons[i]

		isolid := pi.Alpha != 1
		jsolid := pj.Alpha != 1
		if isolid && !jsolid {
			return true
		} else if jsolid && !isolid {
			return false
		}
		// Polygons have the same translucent / solid status

		// Second: sort by "y" solid polygons (or all polygons if
		// alphaYSort is true)
        average := func(poly Polygon) float64{
            sum := float64(0)
            for i := range len(poly.Vertices) {
                sum += poly.Vertices[i].Output.Y
            }
            return sum / float64(len(poly.Vertices))
        }

		if isolid || manualSortAlpha {
			iy := average(pi)
			jy := average(pj)

			if iy < jy {
				return true
			} else if jy < iy {
				return false
			}
		}
		// Now polygons have the same y
		// FIXME: sort left-to-right? For now, ignore.

		// Polygons have the same sorting properties.
		// Return false so that stable sorting keeps them in the right order
		return false

    })

    for _, p := range polygons {

        // 1 dot check seems unneeded
        //if !p.valid1DotDepth(r.Rasterizer.Disp1Dot.V) {
        //    return
        //}

        r.RenderPolygon(&p)
    }

    if r.Rasterizer.GeoEngine.Fog.Enabled {
        r.ApplyFog(depthW)
    }

    if r.Rasterizer.GeoEngine.Disp3dCnt.EdgeMarking {
        r.ApplyEdge(depthW)
    }

    r.ImageToPixels(r.Context.Image())
    //r.ImageToPixels(r.Context.BufferImage())
}

func (r *Render) ApplyFog(depthW bool) {

    fog := &r.Rasterizer.GeoEngine.Fog

    for y := range r.Context.Height {
        for x := range r.Context.Width {

            i := x + y * r.Context.Width

            if !r.Context.FogEnabledBuffer[i] {
                continue
            }

            c := gl.MakeColor(r.Context.Image().At(x, y))

            var depth float64

            // depth is z normalized range 0...1, w is 4096 (W far, W near is 0)
            if depthW {
                depth = r.Context.DepthBufferW[i] * 8
            } else {
                depth = r.Context.DepthBuffer[i] * 0x7FFF
            }

            depth = max(0, min(depth, 0x7FFF))

            ca := gl.MakeColorColor(fog.ApplyFog(c, depth))
            r.Context.SetColor(x, y, ca)
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

    r.Context.PolygonFogEnabled = p.FogEnabled
    r.Context.NewTranslucentDepth = p.SetNewTranslucentDepth
    r.Context.PolygonId = p.Id
    r.Context.PolygonOpaque = p.AlphaV == 0x1F || p.AlphaV == 0
    r.Context.Wireframe = p.AlphaV == 0

    //switch {
    //case p.RenderFront && p.RenderBack:
    //    r.Context.Cull = gl.CullNone
    //case p.RenderFront && !p.RenderBack:
    //    r.Context.Cull = gl.CullFront
    //case !p.RenderFront && p.RenderBack:
    //    r.Context.Cull = gl.CullBack
    //default:
    //    return
    //}

    switch p.PrimitiveType {
    case PRIM_SEP_TRI:

        if invalidCnt := len(p.Vertices) % 3 != 0; invalidCnt {
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

        if invalidCnt := len(p.Vertices) % 4 != 0; invalidCnt {
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

            if clockwise := i & 1 == 1; clockwise {
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

        for i := 2; i + 1 < len(p.Vertices); i += 2 {

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
                p.Vertices[i-2],
                p.Vertices[i-1],
                p.Vertices[i+1],
                p.Vertices[i+0],
            )

            r.Context.DrawQuad(quad)
        }
    }
}

func (r *Render) ImageToPixels(img image.Image) {
    r.lock.Lock()

    i := 0
    for y := range HEIGHT {
        for x := range WIDTH {
            c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)

            // limit to 5 bit, maybe 6??? rounds things nicely
            // ex. if value is .99, then will still be slightly not visible on screen - its jarring
            alpha := (c.A >> 3)
            alpha = (alpha << 3) | (alpha >> 2)

            if r.Rasterizer.Buffers.BisRendering {
                r.Pixels.PalettesB[i] = uint32(RGB24ToRGB15(c.R, c.G, c.B))
                r.Pixels.AlphaB[i] = float32(alpha) / 0xFF
            } else {
                r.Pixels.PalettesA[i] = uint32(RGB24ToRGB15(c.R, c.G, c.B))
                r.Pixels.AlphaA[i] = float32(alpha) / 0xFF
            }
            i++
        }
    }

	r.lock.Unlock()
}
