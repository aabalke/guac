package rast

import (
	"fmt"
	"image"
	"image/color"

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
    //PixelPalettes []uint32
    //Alphas []float64
	Context *gl.Context
    Buffers *Buffers
    RearPlane *RearPlane

	lock        sync.Mutex
}

type Pixels struct {
    PalettesA []uint32
    PalettesB []uint32
    AlphaA    []float64
    AlphaB    []float64
    WritingB  bool
}

func (p *Pixels) InitPixels() {
    p.PalettesA = make([]uint32, WIDTH*HEIGHT)
    p.PalettesB = make([]uint32, WIDTH*HEIGHT)
    p.AlphaA = make([]float64, WIDTH*HEIGHT)
    p.AlphaB = make([]float64, WIDTH*HEIGHT)
}

func NewRender(rast *Rasterizer, buffers *Buffers, rp *RearPlane) *Render {

    r := &Render{
        Rasterizer: rast,
        Buffers: buffers,
        Context: gl.NewContext(WIDTH, HEIGHT),
        RearPlane: rp,
    }

    r.Pixels.InitPixels()

    r.Context.Cull = gl.CullNone
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

    polygons := r.Buffers.GetPolygons()

    //sort.Slice(polygons, func(i, j int) bool {

    //    average := func(poly Polygon) float64{

    //        a := float64(0)
    //        for i := range len(poly.Vertices) {
    //            a += poly.Vertices[i].Output.Z
    //        }
    //        a /= float64(len(poly.Vertices))
    //        return a
    //    }

    //    zi := average(polygons[i])
    //    zj := average(polygons[j])
    //    return zi > zj
    //})

    for _, p := range polygons {
        r.RenderPolygon(&p)
    }

    r.ImageToPixels(r.Context.Image())
}

func (r *Render) RenderPolygon(p *Polygon) {

    if len(p.Vertices) == 0 {
        return
    }

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
                p.Vertices[i+0].CalcTextureVector(tW, tH)
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
            if r.Pixels.WritingB {
                r.Pixels.PalettesB[i] = uint32(RGB24ToRGB15(c.R, c.G, c.B))
                r.Pixels.AlphaB[i] = float64(c.A) / 0xFF
            } else {
                r.Pixels.PalettesA[i] = uint32(RGB24ToRGB15(c.R, c.G, c.B))
                r.Pixels.AlphaA[i] = float64(c.A) / 0xFF
            }
            i++
        }
    }

	r.lock.Unlock()
}
