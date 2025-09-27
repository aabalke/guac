package gl

import (
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
)

type VRAM interface {
    ReadTexture(uint32) uint8
    ReadPalTexture(uint32) uint8
}

type Texture interface {
	Sample(u, v float64) Color
	BilinearSample(u, v float64) Color
}

type BilinearCoords struct {
    X, Y float64
    X0, Y0 int
    X1, Y1 int
}

func getBilinearCoords(w, h, u, v float64) BilinearCoords {
	u -= math.Floor(u)
	v -= math.Floor(v)
	x := u * float64(w-1)
	y := v * float64(h-1)
	x0 := int(x)
	y0 := int(y)
	x1 := x0 + 1
	y1 := y0 + 1
	x -= float64(x0)
	y -= float64(y0)

    return BilinearCoords{
        X: x,
        Y: y,
        X0: x0,
        Y0: y0,
        X1: x1,
        Y1: y1,
    }
}

//func LoadTexture(path string) (Texture, error) {
//	im, err := LoadImage(path)
//	if err != nil {
//		return nil, err
//	}
//	return NewImageTexture(im), nil
//}
//
type ImageTexture struct {
	Width  int
	Height int
	Image  image.Image
}

func NewImageTexture(im image.Image) Texture {
	size := im.Bounds().Max
	return &ImageTexture{size.X, size.Y, im}
}

func (t *ImageTexture) Sample(u, v float64) Color {
	v = 1 - v
	u -= math.Floor(u)
	v -= math.Floor(v)
	x := int(u * float64(t.Width))
	y := int(v * float64(t.Height))
	return MakeColor(t.Image.At(x, y))
}

func (t *ImageTexture) BilinearSample(u, v float64) Color {

    coords := getBilinearCoords(float64(t.Width), float64(t.Height), u, v)

	c00 := MakeColor(t.Image.At(coords.X0, coords.Y0))
	c01 := MakeColor(t.Image.At(coords.X0, coords.Y1))
	c10 := MakeColor(t.Image.At(coords.X1, coords.Y0))
	c11 := MakeColor(t.Image.At(coords.X1, coords.Y1))
	c := Color{}
	c = c.Add(c00.MulScalar((1 - coords.X) * (1 - coords.Y)))
	c = c.Add(c10.MulScalar(coords.X * (1 - coords.Y)))
	c = c.Add(c01.MulScalar((1 - coords.X) * coords.Y))
	c = c.Add(c11.MulScalar(coords.X * coords.Y))
	return c
}

func LoadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	im, _, err := image.Decode(file)
	return im, err
}

type DirectColorTexture struct {
    Width, Height int
    Vram VRAM
    VramBase uint32
}

func (t *DirectColorTexture) Sample(u, v float64) Color {
	u -= math.Floor(u)
	v -= math.Floor(v)
	x := int(u * float64(t.Width))
	y := int(v * float64(t.Height))
    return t.getColor(x, y)
}

func (t *DirectColorTexture) BilinearSample(u, v float64) Color {
    coords := getBilinearCoords(float64(t.Width), float64(t.Height), u, v)
	c00 := t.getColor(coords.X0, coords.Y0)
	c01 := t.getColor(coords.X0, coords.Y1)
	c10 := t.getColor(coords.X1, coords.Y0)
	c11 := t.getColor(coords.X1, coords.Y1)

	c := Color{}
	c = c.Add(c00.MulScalar((1 - coords.X) * (1 - coords.Y)))
	c = c.Add(c10.MulScalar(coords.X * (1 - coords.Y)))
	c = c.Add(c01.MulScalar((1 - coords.X) * coords.Y))
	c = c.Add(c11.MulScalar(coords.X * coords.Y))
	return c
}

func (t *DirectColorTexture) getColor(x, y int) Color {

    i := uint32(x + (y * t.Width)) * 2

    data := uint32(t.Vram.ReadTexture(t.VramBase+i+0))
    data |= uint32(t.Vram.ReadTexture(t.VramBase+i+1)) << 8

    return MakeColorFrom15Bit(
        uint8(data & 0b11111),
        uint8(data >> 5) & 0b11111,
        uint8(data >> 10) & 0b11111,
    )
}

type PalColorTexture struct {
    Width, Height int
    Vram VRAM
    VramBase uint32
    PalBase uint32
    BitsPerTexel uint32
    BitsPerTexelShift uint32
    TransparentZero bool

}

func (t *PalColorTexture) Sample(u, v float64) Color {
	u -= math.Floor(u)
	v -= math.Floor(v)
	x := int(u * float64(t.Width))
	y := int(v * float64(t.Height))
    return t.getColor(x, y)
}

func (t *PalColorTexture) BilinearSample(u, v float64) Color {
    coords := getBilinearCoords(float64(t.Width), float64(t.Height), u, v)
	c00 := t.getColor(coords.X0, coords.Y0)
	c01 := t.getColor(coords.X0, coords.Y1)
	c10 := t.getColor(coords.X1, coords.Y0)
	c11 := t.getColor(coords.X1, coords.Y1)

	c := Color{}
	c = c.Add(c00.MulScalar((1 - coords.X) * (1 - coords.Y)))
	c = c.Add(c10.MulScalar(coords.X * (1 - coords.Y)))
	c = c.Add(c01.MulScalar((1 - coords.X) * coords.Y))
	c = c.Add(c11.MulScalar(coords.X * coords.Y))
	return c
}

func (t *PalColorTexture) getColor(x, y int) Color {

    i := uint32(x + (y * t.Width))

    palIdx := uint32(t.Vram.ReadTexture(t.VramBase + (i >> t.BitsPerTexelShift)))

    switch t.BitsPerTexel {
    case 2: palIdx = (palIdx >> ((i & 0b11) * t.BitsPerTexel)) & 0b11
    case 4: palIdx = (palIdx >> ((i & 0b1) * t.BitsPerTexel)) & 0b1111
    case 8: palIdx = (palIdx >> ((i & 0b0) * t.BitsPerTexel)) & 0b1111_1111
    }

    if palIdx == 0 && t.TransparentZero {
        return MakeColor(color.Transparent)
    }

    // palettes take up 2 bytes each
    palIdx *= 2

    data := uint32(t.Vram.ReadPalTexture(t.PalBase + palIdx))
    data |= uint32(t.Vram.ReadPalTexture(t.PalBase + palIdx + 1)) << 8

    return MakeColorFrom15Bit(
        uint8(data & 0b11111),
        uint8(data >> 5) & 0b11111,
        uint8(data >> 10) & 0b11111,
    )
}

type TranslucentTexture struct {
    Width, Height int
    Vram VRAM
    VramBase uint32
    PalBase uint32
    ColorIdxBits uint32
}

func (t *TranslucentTexture) Sample(u, v float64) Color {
	u -= math.Floor(u)
	v -= math.Floor(v)
	x := int(u * float64(t.Width))
	y := int(v * float64(t.Height))
    return t.getColor(x, y)
}

func (t *TranslucentTexture) BilinearSample(u, v float64) Color {

    coords := getBilinearCoords(float64(t.Width), float64(t.Height), u, v)
	c00 := t.getColor(coords.X0, coords.Y0)
	c01 := t.getColor(coords.X0, coords.Y1)
	c10 := t.getColor(coords.X1, coords.Y0)
	c11 := t.getColor(coords.X1, coords.Y1)

	c := Color{}
	c = c.Add(c00.MulScalar((1 - coords.X) * (1 - coords.Y)))
	c = c.Add(c10.MulScalar(coords.X * (1 - coords.Y)))
	c = c.Add(c01.MulScalar((1 - coords.X) * coords.Y))
	c = c.Add(c11.MulScalar(coords.X * coords.Y))
	return c
}

func (t *TranslucentTexture) getColor(x, y int) Color {

    i := uint32(x + (y * t.Width))

    palIdx := uint32(t.Vram.ReadTexture(t.VramBase + i))

    var colorIdx uint32
    switch t.ColorIdxBits {
    case 3:
        colorIdx = palIdx & 0b111
    case 5:
        colorIdx = palIdx & 0b11111
    }

    //log.Printf("TRANSLUCENT TEXTURE. NEED ALPHA SETUP")

    colorIdx *= 2

    data := uint32(t.Vram.ReadPalTexture(t.PalBase + colorIdx))
    data |= uint32(t.Vram.ReadPalTexture(t.PalBase + colorIdx + 1)) << 8

    c := MakeColorFrom15Bit(
        uint8(data & 0b11111),
        uint8(data >> 5) & 0b11111,
        uint8(data >> 10) & 0b11111,
    )

    switch t.ColorIdxBits {
    case 3:
        c.A = float64(palIdx >> 3) / 31
    case 5:
        c.A = float64(palIdx >> 5) / 7
    }

    return c
}

type CompressedTexture struct {
    Width, Height int
    Vram VRAM
    VramBase uint32
    PalBase uint32
}

func (t *CompressedTexture) Sample(u, v float64) Color {
	u -= math.Floor(u)
	v -= math.Floor(v)
	x := int(u * float64(t.Width))
	y := int(v * float64(t.Height))
    return t.getColor(x, y)
}

func (t *CompressedTexture) BilinearSample(u, v float64) Color {

    coords := getBilinearCoords(float64(t.Width), float64(t.Height), u, v)
	c00 := t.getColor(coords.X0, coords.Y0)
	c01 := t.getColor(coords.X0, coords.Y1)
	c10 := t.getColor(coords.X1, coords.Y0)
	c11 := t.getColor(coords.X1, coords.Y1)

	c := Color{}
	c = c.Add(c00.MulScalar((1 - coords.X) * (1 - coords.Y)))
	c = c.Add(c10.MulScalar(coords.X * (1 - coords.Y)))
	c = c.Add(c01.MulScalar((1 - coords.X) * coords.Y))
	c = c.Add(c11.MulScalar(coords.X * coords.Y))
	return c
}

func (t *CompressedTexture) getColor(x, y int) Color {

    slot1Base := uint32(0x2_0000)

    slot0Base := t.VramBase
    palBase := t.PalBase

    blockX, blockY := x   >> 2, y   >> 2
    texelX, texelY := x & 0b11, y & 0b11

    blockIdx := uint32(blockY*(t.Width/4) + blockX)

    blockData := uint32(t.Vram.ReadTexture(slot0Base + blockIdx * 4))
    blockData |= uint32(t.Vram.ReadTexture(slot0Base + blockIdx * 4 + 1)) << 8
    blockData |= uint32(t.Vram.ReadTexture(slot0Base + blockIdx * 4 + 2)) << 16
    blockData |= uint32(t.Vram.ReadTexture(slot0Base + blockIdx * 4 + 3)) << 24
    rowBits := (blockData >> (texelY*8)) & 0xFF
    texelVal := (rowBits >> (texelX*2)) & 0b11

    palInfo := uint32(t.Vram.ReadTexture(slot1Base + blockIdx*2))
    palInfo |= uint32(t.Vram.ReadTexture(slot1Base + blockIdx*2 + 1)) << 8
    palOffset := (palInfo & 0x3FFF) * 4
    mode := (palInfo >> 14) & 0b11

    color0 := uint16(t.Vram.ReadPalTexture(palBase + palOffset + 0))
    color0 |= uint16(t.Vram.ReadPalTexture(palBase + palOffset + 1)) << 8
    color1 := uint16(t.Vram.ReadPalTexture(palBase + palOffset + 2))
    color1 |= uint16(t.Vram.ReadPalTexture(palBase + palOffset + 3)) << 8
    color2 := uint16(t.Vram.ReadPalTexture(palBase + palOffset + 4))
    color2 |= uint16(t.Vram.ReadPalTexture(palBase + palOffset + 5)) << 8
    color3 := uint16(t.Vram.ReadPalTexture(palBase + palOffset + 6))
    color3 |= uint16(t.Vram.ReadPalTexture(palBase + palOffset + 7)) << 8

    blendMode1 := func (a, b uint16) uint16 {

        aR := uint16(a) & 0b11111
        aG := uint16(a >> 5) & 0b11111
        aB := uint16(a >> 10)& 0b11111

        bR := uint16(b) & 0b11111
        bG := uint16(b >> 5) & 0b11111
        bB := uint16(b >> 10)& 0b11111

        oR := (((aR + bR) / 2) & 0b11111)
        oG := (((aG + bG) / 2) & 0b11111) << 5
        oB := (((aB + bB) / 2) & 0b11111) << 10

        return oR | oG | oB
    }

    blendMode3 := func (a, b uint16) uint16 {

        aR := uint16(a) & 0b11111
        aG := uint16(a >> 5) & 0b11111
        aB := uint16(a >> 10)& 0b11111

        bR := uint16(b) & 0b11111
        bG := uint16(b >> 5) & 0b11111
        bB := uint16(b >> 10)& 0b11111

        oR := (((aR * 5 + bR * 3) / 8) & 0b11111)
        oG := (((aG * 5 + bG * 3) / 8) & 0b11111) << 5
        oB := (((aB * 5 + bB * 3) / 8) & 0b11111) << 10

        return oR | oG | oB
    }


    convert := func(v uint16) Color {
        return MakeColorFrom15Bit(
            uint8(v & 0b11111),
            uint8(v >> 5) & 0b11111,
            uint8(v >> 10) & 0b11111,
        )
    }

    switch texelVal {
    case 0:
        return convert(color0)
    case 1:
        return convert(color1)
    case 2:
        switch mode {
        case 1:
            return convert(blendMode1(color0, color1))
        case 3:
            return convert(blendMode3(color0, color1))
        default:
            return convert(color2)
        }

    case 3:
        switch mode {
        case 2:
            return convert(color3)
        case 3:
            return convert(blendMode3(color1, color0))
        default:
            return MakeColor(color.Transparent)
        }
    }

    panic("UNKNOWN TEXEL VALUE OR MODE")
}
