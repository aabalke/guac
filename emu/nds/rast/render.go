package rast

import (
	"image"
	"image/color"

	"github.com/aabalke/guac/emu/nds/rast/gl"
)

const (
    WIDTH = 256
    HEIGHT = 192

	scale  = 1    // optional supersampling
	//fovy   = 30   // vertical field of view in degrees
	//near   = 0.5    // near clipping plane
	//far    = 1000   // far clipping plane
)

var (
    //eye = gl.V(0,0,0)
    //center = gl.V(0,0,0)
    //up = gl.V(0,0,1)
    //light = gl.V(-1, 1, 0.5).Normalize()

	//aspect = float64(WIDTH) / float64(HEIGHT)
)

type Render struct {
    PixelPalettes []uint32
	Context *gl.Context
	ProjectionMatrix  *gl.Matrix
    Buffers *Buffers
}

func NewRender(buffers *Buffers, projectMatrix *gl.Matrix) Render {

    r := Render{
        Buffers: buffers,
        ProjectionMatrix: projectMatrix,
    }

	context := gl.NewContext(WIDTH*scale, HEIGHT*scale)
    //context.ClearColor = gl.HexColor("#FF0000") //White
    context.ClearColor = gl.Gray(0.5)

	context.ClearColorBuffer()

    r.PixelPalettes = make([]uint32, WIDTH*HEIGHT)
    r.Context = context
    //r.Matrix = matrix

    return r
}

func (r *Render) UpdateRender() {

	r.Context.ClearColorBuffer()
    r.Context.ClearDepthBuffer()

    for _, v := range r.Buffers.GetPolygons() {

        //fmt.Printf("MATRIX % f\n", v.Vertices[0].Output.W)
        tri := gl.NewTriangle(v.Vertices[0], v.Vertices[1], v.Vertices[2])

        shader := gl.NewSolidColorShader(*r.ProjectionMatrix, gl.HexColorLiteral(0xFF0000))

        //shader := gl.NewPhongShader(*r.ProjectionMatrix, light, eye)
        //shader.Texture = *v.Texture
        //shader.ObjectColor = gl.HexColor("#468966")
        r.Context.Shader = shader

        r.Context.DrawTriangle(tri)
    }

	image := r.Context.Image()

    r.ImageToPixels(image)
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

