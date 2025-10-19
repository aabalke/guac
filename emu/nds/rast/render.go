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
    PixelPalettes []uint32
    Alphas []float64
	Context *gl.Context
    Buffers *Buffers
    RearPlane *RearPlane

	lock        sync.Mutex
}

func NewRender(rast *Rasterizer, buffers *Buffers, rp *RearPlane) *Render {

    r := &Render{
        Rasterizer: rast,
        Buffers: buffers,
        Context: gl.NewContext(WIDTH, HEIGHT),
        PixelPalettes: make([]uint32, WIDTH*HEIGHT),
        Alphas: make([]float64, WIDTH*HEIGHT),
        RearPlane: rp,
    }

    r.Context.Cull = gl.CullNone
    r.Context.Shader = gl.NewShader()

    return r
}

func (r *Render) UpdateRender() {

    r.Context.ClearColor = r.RearPlane.ClearColor
    r.Context.ClearColorBuffer()
    r.Context.ClearDepthBuffer()

    polygons := r.Buffers.GetPolygons()

    for _, p := range polygons {
        r.RenderPolygon(&p)
    }

    r.ImageToPixels(r.Context.Image())
}

func (r *Render) RenderPolygon(p *Polygon) {

    if len(p.Vertices) == 0 {
        return
    }

    if shadow := p.Mode == 3; shadow {
        for i := range p.Vertices {
            p.Vertices[i].Color = gl.Transparent
            p.Vertices[i].NdsTexture = nil
            //p.Vertices[i].Color = gl.Color{A: 1, R: 1}
        }
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
            r.PixelPalettes[i] = uint32(RGB24ToRGB15(c.R, c.G, c.B))
            r.Alphas[i] = float64(c.A) / 0xFF
            i++
        }
    }

	r.lock.Unlock()
}

func RGB24ToRGB15(r, g, b uint8) uint16 {
    r5 := uint16(r >> 3)
    g5 := uint16(g >> 3)
    b5 := uint16(b >> 3)

    return (b5 << 10) | (g5 << 5) | r5
}

func RGB15ToRGB24(r, g, b uint8) (uint8, uint8, uint8){
	r = (r << 3) | (r >> 2)
	g = (g << 3) | (g >> 2)
	b = (b << 3) | (b >> 2)
    return r, g, b
}
