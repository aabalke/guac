package rast

import (
	"image"
	"image/color"

	"github.com/aabalke/guac/emu/nds/rast/gl"
	"github.com/aabalke/guac/emu/nds/utils"
)

const(
    PRIM_SEP_TRI  = 0
    PRIM_SEP_QUAD = 1
    PRIM_TRI_STRIP = 2
    PRIM_QUAD_STRIP = 3
)

type Polygon struct {
	Lights                 [4]bool
	Mode                   uint8
	RenderBack             bool
	RenderFront            bool
	SetNewTranslucentDepth bool
	RenderFarPlanePolygons bool
	RenderBehind1Dot       bool
	DrawEqualDepthPixels   bool
	FogEnabled             bool
	Alpha                  uint32
	Id                     uint32

	PrimitiveType uint8
	Vertices      []gl.Vertex

    Texture Texture
}

func (p *Polygon) WriteAttrs(v uint32) {
    p.Lights[0] = utils.BitEnabled(v, 0)
    p.Lights[1] = utils.BitEnabled(v, 1)
    p.Lights[2] = utils.BitEnabled(v, 2)
    p.Lights[3] = utils.BitEnabled(v, 3)
    p.Mode = uint8(utils.GetVarData(v, 4, 5))
    p.RenderBack = utils.BitEnabled(v, 6)
    p.RenderFront = utils.BitEnabled(v, 7)
    p.SetNewTranslucentDepth = utils.BitEnabled(v, 11)
    p.RenderFarPlanePolygons = utils.BitEnabled(v, 12)
    p.RenderBehind1Dot = utils.BitEnabled(v, 13)
    p.DrawEqualDepthPixels = utils.BitEnabled(v, 14)
    p.FogEnabled = utils.BitEnabled(v, 15)
    p.Alpha = utils.GetVarData(v, 16, 20)
    p.Id = utils.GetVarData(v, 24, 29)
}

func (p *Polygon) WriteVtx16(data []uint32, transfromMatrix *gl.Matrix, color gl.Color, S, T float64) {

    x := utils.Convert16ToFloat(uint16(data[1]), 12)
    y := utils.Convert16ToFloat(uint16(data[1] >> 16), 12)
    z := utils.Convert16ToFloat(uint16(data[2]), 12)
    w := float64(1.0)

    // clip mask is posMtx * perMtx
    // we just apply posMtx here since perMtx is applied in shader

    //fmt.Printf("S %.2f T %.2f\n", S, T)

    vw := transfromMatrix.MulVectorW(gl.VectorW{X: x, Y: y, Z: z, W: w})

    v := gl.Vertex{
        Position: gl.Vector{ X: vw.X, Y: vw.Y, Z: vw.Z },
        Color: color,
        W: vw.W,
        S: S,
        T: T,
    }

    p.Vertices = append(p.Vertices, v)
}

// this is temp using imagetexture, would be best to get texture more effectently
func (p *Polygon) GetTexture(vram VRAM) *gl.ImageTexture {

    t := p.Texture

    img := image.NewRGBA(image.Rect(0,0,int(t.SizeS), int(t.SizeT)))
    switch t.Format {
    case 7: 
        for i := uint32(0); i < t.SizeS * t.SizeT * 2; i += 2 {  

            data := uint32(vram.ReadTexture(i+0))
            data |= uint32(vram.ReadTexture(i+1)) << 8

            x := int((i >> 1) % t.SizeS)
            y := int((i >> 1) / t.SizeS)

            r, g, b := RGB15ToRGB24(
                uint8(data & 0b11111),
                uint8(data >> 5) & 0b11111,
                uint8(data >> 10) & 0b11111,
            )

            img.Set(x, y, color.RGBA{r, g, b, 0xFF})
        }

    }

    return &gl.ImageTexture{
        Width: int(t.SizeS),
        Height: int(t.SizeT),
        Image: img,
    }
}
