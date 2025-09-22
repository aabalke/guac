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

func (p *Polygon) GetVertex(x, y, z float64, clipMtx *gl.Matrix, color gl.Color, S, T float64) gl.Vertex {

    vert := gl.VectorW{X: x,Y: y,Z: z,W: 1.0}
    output := clipMtx.MulVectorW(vert)

    v := gl.Vertex{
        Position: gl.Vector{ X: x, Y: y, Z: z },
        Color: color,
        W: 1.0,
        S: S,
        T: T,
        Output: output,
    }

    return v
}

func (p *Polygon) WriteVtx16(data []uint32, clipMtx *gl.Matrix, color gl.Color, S, T float64) *gl.Vertex {

    x := utils.Convert16ToFloat(uint16(data[1]), 12)
    y := utils.Convert16ToFloat(uint16(data[1] >> 16), 12)
    z := utils.Convert16ToFloat(uint16(data[2]), 12)

    v := p.GetVertex(x, y, z, clipMtx, color, S, T)
    p.Vertices = append(p.Vertices, v)
    return &v
}

func (p *Polygon) WriteVtx10(data []uint32, clipMtx *gl.Matrix, color gl.Color, S, T float64) *gl.Vertex {

    x := utils.Convert10ToFloat(uint16(data[1]), 6)
    y := utils.Convert10ToFloat(uint16(data[1] >> 10), 6)
    z := utils.Convert10ToFloat(uint16(data[1] >> 20), 6)

    v := p.GetVertex(x, y, z, clipMtx, color, S, T)


    p.Vertices = append(p.Vertices, v)
    return &v
}

const (
    REL_XY = 0
    REL_XZ = 1
    REL_YZ = 2
)

func (p *Polygon) WriteVtxRelative(data []uint32, clipMtx *gl.Matrix, color gl.Color, S, T float64, prev *gl.Vertex, set uint8) *gl.Vertex {

    var x, y, z float64
    switch set {
    case REL_XY:
        x = utils.Convert16ToFloat(uint16(data[1]), 12)
        y = utils.Convert16ToFloat(uint16(data[1] >> 10), 12)
        z = prev.Position.Z
    case REL_XZ:
        x = utils.Convert16ToFloat(uint16(data[1]), 12)
        y = prev.Position.Y
        z = utils.Convert16ToFloat(uint16(data[1] >> 10), 12)
    case REL_YZ:
        x = prev.Position.X
        y = utils.Convert16ToFloat(uint16(data[1]), 12)
        z = utils.Convert16ToFloat(uint16(data[1] >> 10), 12)
    }

    v := p.GetVertex(x, y, z, clipMtx, color, S, T)
    p.Vertices = append(p.Vertices, v)
    return &v
}

// this is temp using imagetexture, would be best to get texture more effectently
func (p *Polygon) GetTexture(vram VRAM) gl.Texture {

    t := p.Texture

    switch t.Format {

    case TEX_FMT_NONE:
        return nil

    case TEX_FMT_4_PAL:

        return &gl.PalColorTexture{
            Width: int(t.SizeS),
            Height: int(t.SizeT),
            Vram: vram,
            VramBase: t.VramOffset,
            PalBase: t.PaletteBaseAddr * 0x08,
            BitsPerTexel: 2,
            BitsPerTexelShift: 2,
        }

    case TEX_FMT_16_PAL:

        return &gl.PalColorTexture{
            Width: int(t.SizeS),
            Height: int(t.SizeT),
            Vram: vram,
            VramBase: t.VramOffset,
            PalBase: t.PaletteBaseAddr * 0x10,
            BitsPerTexel: 4,
            BitsPerTexelShift: 1,
        }

    case TEX_FMT_256_PAL:

        return &gl.PalColorTexture{
            Width: int(t.SizeS),
            Height: int(t.SizeT),
            Vram: vram,
            VramBase: t.VramOffset,
            PalBase: t.PaletteBaseAddr * 0x10,
            BitsPerTexel: 8,
            BitsPerTexelShift: 0,
        }

    case TEX_FMT_DIRECT: 

        return &gl.DirectColorTexture{
            Width: int(t.SizeS),
            Height: int(t.SizeT),
            Vram: vram,
            VramBase: t.VramOffset,
        }

    case TEX_FMT_A3I5:

        return &gl.TranslucentTexture{
            Width: int(t.SizeS),
            Height: int(t.SizeT),
            Vram: vram,
            VramBase: t.VramOffset,
            PalBase: t.PaletteBaseAddr * 0x10,
            ColorIdxBits: 5,
        }

    case TEX_FMT_A5I3:

        return &gl.TranslucentTexture{
            Width: int(t.SizeS),
            Height: int(t.SizeT),
            Vram: vram,
            VramBase: t.VramOffset,
            PalBase: t.PaletteBaseAddr * 0x10,
            ColorIdxBits: 3,
        }

    case TEX_FMT_4X4:

        panic("UNSETUP TEXTURE FMT 4X4")

    default: // this loads entire direct color texture in memory and sends it
        
        img := image.NewRGBA(image.Rect(0,0,int(t.SizeS), int(t.SizeT)))
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
        return &gl.ImageTexture{
            Width: int(t.SizeS),
            Height: int(t.SizeT),
            Image: img,
        }

    }

    panic("UNKNOWN TEXTURE TYPE")
}
