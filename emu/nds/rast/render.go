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

func (r *Render) UpdateRender() {

    r.Context.ClearColor = gl.Transparent
    if !r.Rasterizer.GeoEngine.Disp3dCnt.RearPlaneBitmapEnabled {
        r.Context.ClearColor = r.RearPlane.ClearColor
    }

    r.Context.ClearColorBuffer()
    r.Context.ClearDepthBuffer()

    r.Context.AlphaBlending = r.Rasterizer.GeoEngine.Disp3dCnt.AlphaBlending

    polygons, manualSort, depthW := r.Buffers.GetPolygons()

    if !manualSort {
        sort.Slice(polygons, func(i, j int) bool {

            average := func(poly Polygon) float64{
                sum := float64(0)
                for i := range len(poly.Vertices) {
                    sum += poly.Vertices[i].Output.Y
                }
                return sum / float64(len(poly.Vertices))
            }

            zi := average(polygons[i])
            zj := average(polygons[j])
            return zi > zj
        })
    }

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

    r.ImageToPixels(r.Context.Image())
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

            if depthW {
                depth = r.Context.DepthBufferW[i] * 8
            } else {
                depth = r.Context.DepthBuffer[i] * 0x7FFF
            }

            ca := gl.MakeColorColor(fog.ApplyFog(c, depth))
            r.Context.SetColor(x, y, ca)
        }
    }
}

func (r *Render) RenderPolygon(p *Polygon) {

    if len(p.Vertices) == 0 {
        return
    }

    r.Context.PolygonFogEnabled = p.FogEnabled
    r.Context.NewTranslucentDepth = p.SetNewTranslucentDepth

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
