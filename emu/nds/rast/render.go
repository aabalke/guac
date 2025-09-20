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
	Context *gl.Context
	ProjectionMatrix  *gl.Matrix
    Buffers *Buffers
    RearPlane *RearPlane
}

func NewRender(rast *Rasterizer, buffers *Buffers, projectMatrix *gl.Matrix, rp *RearPlane) *Render {

    r := &Render{
        Rasterizer: rast,
        Buffers: buffers,
        ProjectionMatrix: projectMatrix,
        Context: gl.NewContext(WIDTH, HEIGHT),
        PixelPalettes: make([]uint32, WIDTH*HEIGHT),
        RearPlane: rp,

    }

    r.Context.Cull = gl.CullNone

    return r
}

func (r *Render) UpdateRender() {

    //r.Rasterizer.DebugTexture(7)
    //return

    r.Context.ClearColor = r.RearPlane.ClearColor
    r.Context.ClearColorBuffer()
    r.Context.ClearDepthBuffer()

    r.Context.Shader = gl.NewNdsShader(*r.ProjectionMatrix)

    for _, p := range r.Buffers.GetPolygons() {
        //r.Context.Shader.SetTexture(*r.Texture)
        r.Context.Shader.SetTexture(p.GetTexture(r.Rasterizer.VRAM))
        r.RenderPolygon(&p)
    }

	image := r.Context.Image()

    r.ImageToPixels(image)
}

func (r *Render) RenderPolygon(p *Polygon) {


    tW := int(p.Texture.SizeS)
    tH := int(p.Texture.SizeT)

    switch p.PrimitiveType {
    case PRIM_SEP_TRI:

        if invalidCnt := len(p.Vertices) % 3 != 0; invalidCnt {
            fmt.Printf("Separate Tri Polygon has invalid vert count.\n")
        }

        for i := 0; i < len(p.Vertices); i += 3 {

            //if i + 2 > len(p.Vertices) {
            //    continue
            //}

            p.Vertices[i+0].CalcTextureVector(tW, tH)
            p.Vertices[i+1].CalcTextureVector(tW, tH)
            p.Vertices[i+2].CalcTextureVector(tW, tH)

            tri := gl.NewTriangle(
                p.Vertices[i+0],
                p.Vertices[i+1],
                p.Vertices[i+2])

            r.Context.DrawTriangle(tri)
        }

    case PRIM_SEP_QUAD:

        if invalidCnt := len(p.Vertices) % 4 != 0; invalidCnt {
            fmt.Printf("Separate Quad Polygon has invalid vert count.\n")
        }

        for i := 0; i < len(p.Vertices); i += 4 {

            //if i + 3 > len(p.Vertices) {
            //    continue
            //}

            p.Vertices[i+0].CalcTextureVector(tW, tH)
            p.Vertices[i+1].CalcTextureVector(tW, tH)
            p.Vertices[i+2].CalcTextureVector(tW, tH)
            p.Vertices[i+3].CalcTextureVector(tW, tH)

            quad := gl.NewQuad(
                p.Vertices[i+0],
                p.Vertices[i+1],
                p.Vertices[i+2],
                p.Vertices[i+3])

            r.Context.DrawQuad(quad)
        }

    case PRIM_TRI_STRIP:

        //if invalidCnt := len(p.Vertices) % 4 != 0; invalidCnt {
        //    fmt.Printf("Separate Quad Polygon has invalid vert count.\n")
        //}

        for i := 0; i < len(p.Vertices); i++ {

            p.Vertices[i+0].CalcTextureVector(tW, tH)
            p.Vertices[i+1].CalcTextureVector(tW, tH)
            p.Vertices[i+2].CalcTextureVector(tW, tH)

            tri := gl.NewTriangle(
                p.Vertices[i+0],
                p.Vertices[i+1],
                p.Vertices[i+2])

            r.Context.DrawTriangle(tri)
        }

    case PRIM_QUAD_STRIP:

        //if invalidCnt := len(p.Vertices) % 4 != 0; invalidCnt {
        //    fmt.Printf("Separate Quad Polygon has invalid vert count.\n")
        //}

        for i := 0; i < len(p.Vertices); i += 2 {

            //if i + 3 > len(p.Vertices) {
            //    continue
            //}
            p.Vertices[i+0].CalcTextureVector(tW, tH)
            p.Vertices[i+1].CalcTextureVector(tW, tH)
            p.Vertices[i+2].CalcTextureVector(tW, tH)
            p.Vertices[i+3].CalcTextureVector(tW, tH)

            quad := gl.NewQuad(
                p.Vertices[i+0],
                p.Vertices[i+1],
                p.Vertices[i+2],
                p.Vertices[i+3])

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
