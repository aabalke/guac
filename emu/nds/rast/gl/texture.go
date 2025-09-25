package gl

import (
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
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
    NearestNeightborSample(u , v float64) Color
}

type BilinearCoords struct {
    X, Y float64
    X0, Y0 int
    X1, Y1 int
}

func getBilinearCoords(w, h, u, v float64) BilinearCoords {
	v = 1 - v
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

func (t *ImageTexture) NearestNeightborSample(u, v float64) Color {
    v = 1 - v

    u -= math.Floor(u)
    v -= math.Floor(v)

    x := int(math.Round(u * float64(t.Width-1)))
    y := int(math.Round(v * float64(t.Height-1)))

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
	v = 1 - v
	u -= math.Floor(u)
	v -= math.Floor(v)
	x := int(u * float64(t.Width))
	y := int(v * float64(t.Height))
    return t.getColor(x, y)
}

func (t *DirectColorTexture) NearestNeightborSample(u, v float64) Color {
    v = 1 - v

    u -= math.Floor(u)
    v -= math.Floor(v)

    x := int(math.Round(u * float64(t.Width-1)))
    y := int(math.Round(v * float64(t.Height-1)))

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
}

func (t *PalColorTexture) Sample(u, v float64) Color {
	v = 1 - v
	u -= math.Floor(u)
	v -= math.Floor(v)
	x := int(u * float64(t.Width))
	y := int(v * float64(t.Height))
    return t.getColor(x, y)
}

func (t *PalColorTexture) NearestNeightborSample(u, v float64) Color {
    v = 1 - v

    u -= math.Floor(u)
    v -= math.Floor(v)

    x := int(math.Round(u * float64(t.Width-1)))
    y := int(math.Round(v * float64(t.Height-1)))

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
	v = 1 - v
	u -= math.Floor(u)
	v -= math.Floor(v)
	x := int(u * float64(t.Width))
	y := int(v * float64(t.Height))
    return t.getColor(x, y)
}

func (t *TranslucentTexture) NearestNeightborSample(u, v float64) Color {
    v = 1 - v

    u -= math.Floor(u)
    v -= math.Floor(v)

    x := int(math.Round(u * float64(t.Width-1)))
    y := int(math.Round(v * float64(t.Height-1)))

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

    log.Printf("TRANSLUCENT TEXTURE. NEED ALPHA SETUP")

    colorIdx *= 2

    data := uint32(t.Vram.ReadPalTexture(t.PalBase + colorIdx))
    data |= uint32(t.Vram.ReadPalTexture(t.PalBase + colorIdx + 1)) << 8

    return MakeColorFrom15Bit(
        uint8(data & 0b11111),
        uint8(data >> 5) & 0b11111,
        uint8(data >> 10) & 0b11111,
    )
}

type CompressedTexture struct {
    Width, Height int
    Vram VRAM
    VramBase uint32
    PalBase uint32
}

func (t *CompressedTexture) Sample(u, v float64) Color {
    return MakeColor(color.White)
}

func (t *CompressedTexture) NearestNeightborSample(u, v float64) Color {
    return MakeColor(color.White)
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

    var colors [4]uint16

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

    transparent := uint16(0b11111) // red

    switch mode {
    case 0:
        colors[0] = color0
        colors[1] = color1
        colors[2] = color2
        colors[3] = transparent

    case 1:
        colors[0] = color0
        colors[1] = color1
        colors[3] = blendMode1(color0, color1)
        colors[3] = transparent

    case 2:
        colors[0] = color0
        colors[1] = color1
        colors[2] = color2
        colors[3] = color3
    case 3:
        colors[0] = color0
        colors[1] = color1
        colors[2] = blendMode3(color0, color1)
        colors[3] = blendMode3(color1, color0)
    default:
        return MakeColorFrom15Bit(
            0b11111,
            0,
            0,
        )
    }

    return MakeColorFrom15Bit(
        uint8(colors[texelVal] & 0b11111),
        uint8(colors[texelVal] >> 5) & 0b11111,
        uint8(colors[texelVal] >> 10) & 0b11111,
    )

    //var colors [4]uint16
    //switch mode {
    //case 0:
    //    colors[0] = color0
    //    colors[1] = color1
    //    colors[2] = ReadPalColor(palBase + palOffset + 4)
    //    colors[3] = Transparent
    //case 1:
    //    colors[0] = color0
    //    colors[1] = color1
    //    colors[2] = Blend(color0, color1, 0.5)
    //    colors[3] = Transparent
    //case 2:
    //    colors[0] = color0
    //    colors[1] = color1
    //    colors[2] = ReadPalColor(palBase + palOffset + 4)
    //    colors[3] = ReadPalColor(palBase + palOffset + 6)
    //case 3:
    //    colors[0] = color0
    //    colors[1] = color1
    //    //colors[2] = Blend(color0, color1, 5.0/8.0)
    //    //colors[3] = Blend(color0, color1, 3.0/8.0)
    //}

    //return colors[texelVal]
}
