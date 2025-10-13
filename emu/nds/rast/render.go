package rast

import (
	"fmt"
	"image"
	"image/color"

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

    return r
}

func (r *Render) UpdateRender() {

    r.Context.ClearColor = r.RearPlane.ClearColor
    r.Context.ClearColorBuffer()
    r.Context.ClearDepthBuffer()

    r.Context.Shader = gl.NewNdsShader(r.Rasterizer.GeoEngine.Lights)

    for _, p := range r.Buffers.GetPolygons() {
        //r.Context.Shader.SetTexture(*r.Texture)
        //r.Context.Shader.SetTexture(
        //    p.GetTexture(
        //        r.Rasterizer.VRAM,
        //        &r.Rasterizer.GeoEngine.TextureCache))
        r.Context.Shader.(*gl.NdsShader).LightEnabled = p.LightsEnabled
        r.RenderPolygon(&p)
    }

	image := r.Context.Image()

    r.ImageToPixels(image)
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


            if p.Vertices[i].NdsTexture != nil {
                r.Context.Shader.SetTexture(p.Vertices[i].NdsTexture)
                tW := int(p.Vertices[i].NdsTexture.Width)
                tH := int(p.Vertices[i].NdsTexture.Height)
                p.Vertices[i+2].CalcTextureVector(tW, tH)
                p.Vertices[i+1].CalcTextureVector(tW, tH)
                p.Vertices[i+0].CalcTextureVector(tW, tH)
            }
            if p.Mode == 3 {
                r.Context.Shader.SetTexture(nil)
                //p.Vertices[i+2].Color.A = 0
                //p.Vertices[i+1].Color.A = 0
                //p.Vertices[i+0].Color.A = 0
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


            if p.Vertices[i].NdsTexture != nil {
                r.Context.Shader.SetTexture(p.Vertices[i].NdsTexture)
                tW := int(p.Vertices[i].NdsTexture.Width)
                tH := int(p.Vertices[i].NdsTexture.Height)
                p.Vertices[i+3].CalcTextureVector(tW, tH)
                p.Vertices[i+2].CalcTextureVector(tW, tH)
                p.Vertices[i+1].CalcTextureVector(tW, tH)
                p.Vertices[i+0].CalcTextureVector(tW, tH)
            }
            if p.Mode == 3 {
                r.Context.Shader.SetTexture(nil)
                //p.Vertices[i+3].Color.A = 0
                //p.Vertices[i+2].Color.A = 0
                //p.Vertices[i+1].Color.A = 0
                //p.Vertices[i+0].Color.A = 0
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


            if p.Vertices[i].NdsTexture != nil {
                r.Context.Shader.SetTexture(p.Vertices[i].NdsTexture)
                tW := int(p.Vertices[i].NdsTexture.Width)
                tH := int(p.Vertices[i].NdsTexture.Height)
                p.Vertices[i-2].CalcTextureVector(tW, tH)
                p.Vertices[i-1].CalcTextureVector(tW, tH)
                p.Vertices[i-0].CalcTextureVector(tW, tH)
            }

            if p.Mode == 3 {
                r.Context.Shader.SetTexture(nil)
                //p.Vertices[i-2].Color.A = 0
                //p.Vertices[i-1].Color.A = 0
                //p.Vertices[i-0].Color.A = 0
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

            if p.Vertices[i].NdsTexture != nil {
                r.Context.Shader.SetTexture(p.Vertices[i].NdsTexture)
                tW := p.Vertices[i].NdsTexture.Width
                tH := p.Vertices[i].NdsTexture.Height
                p.Vertices[i-2].CalcTextureVector(tW, tH)
                p.Vertices[i-1].CalcTextureVector(tW, tH)
                p.Vertices[i+1].CalcTextureVector(tW, tH)
                p.Vertices[i+0].CalcTextureVector(tW, tH)
            }

            if p.Mode == 3 {
                r.Context.Shader.SetTexture(nil)
                //p.Vertices[i-2].Color.A = 0
                //p.Vertices[i-1].Color.A = 0
                //p.Vertices[i+1].Color.A = 0
                //p.Vertices[i+0].Color.A = 0
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
    i := 0
    for y := range HEIGHT {
        for x := range WIDTH {
            c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
            r.PixelPalettes[i] = uint32(RGB24ToRGB15(c.R, c.G, c.B))
            r.Alphas[i] = float64(c.A) / 0xFF
            i++
        }
    }
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
