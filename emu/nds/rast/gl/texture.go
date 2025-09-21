package gl

import (
	"image"
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
	v = 1 - v
	u -= math.Floor(u)
	v -= math.Floor(v)
	x := u * float64(t.Width-1)
	y := v * float64(t.Height-1)
	x0 := int(x)
	y0 := int(y)
	x1 := x0 + 1
	y1 := y0 + 1
	x -= float64(x0)
	y -= float64(y0)
	c00 := MakeColor(t.Image.At(x0, y0))
	c01 := MakeColor(t.Image.At(x0, y1))
	c10 := MakeColor(t.Image.At(x1, y0))
	c11 := MakeColor(t.Image.At(x1, y1))
	c := Color{}
	c = c.Add(c00.MulScalar((1 - x) * (1 - y)))
	c = c.Add(c10.MulScalar(x * (1 - y)))
	c = c.Add(c01.MulScalar((1 - x) * y))
	c = c.Add(c11.MulScalar(x * y))
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
	v = 1 - v
	u -= math.Floor(u)
	v -= math.Floor(v)
	x := u * float64(t.Width-1)
	y := v * float64(t.Height-1)
	x0 := int(x)
	y0 := int(y)
	x1 := x0 + 1
	y1 := y0 + 1
	x -= float64(x0)
	y -= float64(y0)
	c00 := t.getColor(x0, y0)
	c01 := t.getColor(x0, y1)
	c10 := t.getColor(x1, y0)
	c11 := t.getColor(x1, y1)

	c := Color{}
	c = c.Add(c00.MulScalar((1 - x) * (1 - y)))
	c = c.Add(c10.MulScalar(x * (1 - y)))
	c = c.Add(c01.MulScalar((1 - x) * y))
	c = c.Add(c11.MulScalar(x * y))
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
	v = 1 - v
	u -= math.Floor(u)
	v -= math.Floor(v)
	x := u * float64(t.Width-1)
	y := v * float64(t.Height-1)
	x0 := int(x)
	y0 := int(y)
	x1 := x0 + 1
	y1 := y0 + 1
	x -= float64(x0)
	y -= float64(y0)
	c00 := t.getColor(x0, y0)
	c01 := t.getColor(x0, y1)
	c10 := t.getColor(x1, y0)
	c11 := t.getColor(x1, y1)

	c := Color{}
	c = c.Add(c00.MulScalar((1 - x) * (1 - y)))
	c = c.Add(c10.MulScalar(x * (1 - y)))
	c = c.Add(c01.MulScalar((1 - x) * y))
	c = c.Add(c11.MulScalar(x * y))
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
	v = 1 - v
	u -= math.Floor(u)
	v -= math.Floor(v)
	x := u * float64(t.Width-1)
	y := v * float64(t.Height-1)
	x0 := int(x)
	y0 := int(y)
	x1 := x0 + 1
	y1 := y0 + 1
	x -= float64(x0)
	y -= float64(y0)
	c00 := t.getColor(x0, y0)
	c01 := t.getColor(x0, y1)
	c10 := t.getColor(x1, y0)
	c11 := t.getColor(x1, y1)

	c := Color{}
	c = c.Add(c00.MulScalar((1 - x) * (1 - y)))
	c = c.Add(c10.MulScalar(x * (1 - y)))
	c = c.Add(c01.MulScalar((1 - x) * y))
	c = c.Add(c11.MulScalar(x * y))
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
