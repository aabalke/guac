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
	LightsEnabled          [4]bool
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
    p.LightsEnabled[0] = utils.BitEnabled(v, 0)
    p.LightsEnabled[1] = utils.BitEnabled(v, 1)
    p.LightsEnabled[2] = utils.BitEnabled(v, 2)
    p.LightsEnabled[3] = utils.BitEnabled(v, 3)
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

    //fmt.Printf("LIGHTS % v\n", p.LightsEnabled)
}

const (

    V_16 = 0
    V_10 = 1
    V_XY = 2
    V_XZ = 3
    V_YZ = 4
    V_DF = 5

)

func (p *Polygon) WriteVertex(data []uint32, g *GeoEngine, method uint8) *gl.Vertex {

    var S, T float64
    S = g.Texture.S
    T = g.Texture.T
    switch g.Texture.TransformationMode { // textrans
    case 0: // continue
    case 1:

        textureVertex := gl.VectorW{
            X: S,
            Y: T,
            Z: 1.0/16,
            W: 1.0/16,
        }

        mtx := &g.MtxStacks.Stacks[3].CurrMtx

        S = textureVertex.Dot(mtx.Col(0))
        T = textureVertex.Dot(mtx.Col(1))
        //S = textureVertex.Dot(mtx.Row(0))
        //T = textureVertex.Dot(mtx.Row(1))

    default:
        //panic("UNSETUP TEXTURE TRANSFORMATION MODE")
    }

    var x, y, z float64

    switch method {
    case V_16:
        x = utils.Convert16ToFloat(uint16(data[1]), 12)
        y = utils.Convert16ToFloat(uint16(data[1] >> 16), 12)
        z = utils.Convert16ToFloat(uint16(data[2]), 12)
    case V_10:
        x = utils.Convert10ToFloat(uint16(data[1]), 6)
        y = utils.Convert10ToFloat(uint16(data[1] >> 10), 6)
        z = utils.Convert10ToFloat(uint16(data[1] >> 20), 6)
    case V_XY:
        prev := g.Vertex
        x = utils.Convert16ToFloat(uint16(data[1]), 12)
        y = utils.Convert16ToFloat(uint16(data[1] >> 16), 12)
        z = prev.Position.Z
    case V_XZ:
        prev := g.Vertex
        x = utils.Convert16ToFloat(uint16(data[1]), 12)
        y = prev.Position.Y
        z = utils.Convert16ToFloat(uint16(data[1] >> 16), 12)
    case V_YZ:
        prev := g.Vertex
        x = prev.Position.X
        y = utils.Convert16ToFloat(uint16(data[1]), 12)
        z = utils.Convert16ToFloat(uint16(data[1] >> 16), 12)
    case V_DF:
        prev := g.Vertex
        convert := func(v uint16) float64 {
            raw := int32(v & 0x3FF)
            if raw&0x200 != 0 {
                raw |= ^0x3FF
            }

            f := float64(raw) / (1 << 9)

            return f / 8.0
        }

        x = convert(uint16(data[1]))     + prev.Position.X
        y = convert(uint16(data[1]>>10)) + prev.Position.Y
        z = convert(uint16(data[1]>>20)) + prev.Position.Z
    }

    v := p.GetVertex(x, y, z, &g.ClipMatrix, g.Color, S, T, &g.StoredNormal)
    p.Vertices = append(p.Vertices, v)
    return &v
}

func (p *Polygon) GetVertex(x, y, z float64, clipMtx *gl.Matrix, color gl.Color, S, T float64, normal *gl.Vector) gl.Vertex {

    vert := gl.VectorW{X: x,Y: y,Z: z,W: 1.0}
    output := clipMtx.MulVectorW(vert)

    v := gl.Vertex{
        Normal: *normal,
        Position: gl.Vector{ X: x, Y: y, Z: z },
        Color: color,
        W: 1.0,
        S: S,
        T: T,
        Output: output,
    }

    return v
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
            TransparentZero: t.TransparentZero,
            RepeatS: t.RepeatS,
            RepeatT: t.RepeatT,
            FlipS: t.FlipS,
            FlipT: t.FlipT,
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
            TransparentZero: t.TransparentZero,
            RepeatS: t.RepeatS,
            RepeatT: t.RepeatT,
            FlipS: t.FlipS,
            FlipT: t.FlipT,
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
            TransparentZero: t.TransparentZero,
            RepeatS: t.RepeatS,
            RepeatT: t.RepeatT,
            FlipS: t.FlipS,
            FlipT: t.FlipT,
        }

    case TEX_FMT_DIRECT: 

        return &gl.DirectColorTexture{
            Width: int(t.SizeS),
            Height: int(t.SizeT),
            Vram: vram,
            VramBase: t.VramOffset,
            RepeatS: t.RepeatS,
            RepeatT: t.RepeatT,
            FlipS: t.FlipS,
            FlipT: t.FlipT,
        }

    case TEX_FMT_A3I5:

        return &gl.TranslucentTexture{
            Width: int(t.SizeS),
            Height: int(t.SizeT),
            Vram: vram,
            VramBase: t.VramOffset,
            PalBase: t.PaletteBaseAddr * 0x10,
            ColorIdxBits: 5,
            RepeatS: t.RepeatS,
            RepeatT: t.RepeatT,
            FlipS: t.FlipS,
            FlipT: t.FlipT,
        }

    case TEX_FMT_A5I3:

        return &gl.TranslucentTexture{
            Width: int(t.SizeS),
            Height: int(t.SizeT),
            Vram: vram,
            VramBase: t.VramOffset,
            PalBase: t.PaletteBaseAddr * 0x10,
            ColorIdxBits: 3,
            RepeatS: t.RepeatS,
            RepeatT: t.RepeatT,
            FlipS: t.FlipS,
            FlipT: t.FlipT,
        }

    case TEX_FMT_4X4:

        return &gl.CompressedTexture{
            Width: int(t.SizeS),
            Height: int(t.SizeT),
            Vram: vram,
            VramBase: t.VramOffset,
            PalBase: t.PaletteBaseAddr * 0x10,
            RepeatS: t.RepeatS,
            RepeatT: t.RepeatT,
            FlipS: t.FlipS,
            FlipT: t.FlipT,
            PitchShift: t.PitchShift,
            CachedTexture: get4x4(vram, &p.Texture),
        }

        //panic("UNSETUP TEXTURE FMT 4X4")

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

func get4x4(vram VRAM, tex *Texture) []uint8 {

	off := tex.VramOffset
	out := make([]uint8, (tex.SizeS)*(tex.SizeT)*2)

    const SLOT_SIZE = 128 * 1024

	var xtraoff uint32
	switch slot := off / SLOT_SIZE; slot {
	case 0:
		xtraoff = SLOT_SIZE + off/2
	case 2:
		xtraoff = SLOT_SIZE + (off-2*SLOT_SIZE)/2 + 0x10000
	default:
        return []uint8{}
		//panic("compressed texture in wrong slot?")
	}

	for y := uint32(0); y < tex.SizeT; y += 4 {
		for x := uint32(0); x < tex.SizeS; x += 4 {
			xtra := (
                uint32(vram.ReadTexture(xtraoff+0)) |
				uint32(vram.ReadTexture(xtraoff+1)) << 8)

			xtraoff += 2
			mode := xtra >> 14
			paloff := uint32(xtra & 0x3FFF)

			palAddr := (tex.PaletteBaseAddr * 0x10) + paloff*4

			var colors [4]uint16
			colors[0] = (uint16(vram.ReadPalTexture(palAddr+0)) |
				uint16(vram.ReadPalTexture(palAddr+1))<<8)
			colors[1] = (uint16(vram.ReadPalTexture(palAddr+2)) |
				uint16(vram.ReadPalTexture(palAddr+3))<<8)
			colors2 := (uint16(vram.ReadPalTexture(palAddr+4)) |
				uint16(vram.ReadPalTexture(palAddr+5))<<8)
			colors3 := (uint16(vram.ReadPalTexture(palAddr+6)) |
				uint16(vram.ReadPalTexture(palAddr+7))<<8)

			switch mode {
			case 0:
				colors[2] = colors2
			case 1:
				colors[2] = blendMode1(colors[0], colors[1])
			case 2:
				colors[2] = colors2
				colors[3] = colors3
			case 3:
				colors[2] = blendMode3(colors[0], colors[1])
				colors[3] = blendMode3(colors[1], colors[0])
			}

			for j := range uint32(4) {
				pack := vram.ReadTexture(off)
				off++
				for i := range uint32(4) {
                    k := ((y+j)<<tex.PitchShift+(x+i))*2
					tex := (pack >> uint(i*2)) & 3

					out[k] = uint8(colors[tex])
					out[k+1] = uint8(colors[tex] >> 8)
				}
			}
		}
	}

	return out
}

func blendMode1(a, b uint16) uint16 {

    aR := uint16(a) & 0b11111
    aG := uint16(a>>5) & 0b11111
    aB := uint16(a>>10) & 0b11111

    bR := uint16(b) & 0b11111
    bG := uint16(b>>5) & 0b11111
    bB := uint16(b>>10) & 0b11111

    oR := (((aR + bR) / 2) & 0b11111)
    oG := (((aG + bG) / 2) & 0b11111) << 5
    oB := (((aB + bB) / 2) & 0b11111) << 10

    return oR | oG | oB
}

func blendMode3(a, b uint16) uint16 {

    aR := uint16(a) & 0b11111
    aG := uint16(a>>5) & 0b11111
    aB := uint16(a>>10) & 0b11111

    bR := uint16(b) & 0b11111
    bG := uint16(b>>5) & 0b11111
    bB := uint16(b>>10) & 0b11111

    oR := (((aR*5 + bR*3) / 8) & 0b11111)
    oG := (((aG*5 + bG*3) / 8) & 0b11111) << 5
    oB := (((aB*5 + bB*3) / 8) & 0b11111) << 10

    return oR | oG | oB
}
