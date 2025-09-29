package gl

import (
	"encoding/binary"
	"math"
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

func getTextureCoords(u, v float64, w, h int, repeatT, repeatS, flipT, flipS bool) (int, int) {

	x := int(u * float64(w))
	y := int(v * float64(h))

    if repeatT {

        flip := flipT && int(v) & 1 == 1 
        v -= math.Floor(v)
        tmp := int(v * float64(h))

        if flip {
            y = h - tmp
        } else {
            y = tmp
        }

    } else {
        y = min(h-1, y)
        y = max(y, 0)
    }

    if repeatS {
        flip := flipS && int(u) & 1 == 1 
        u -= math.Floor(u)
        tmp := int(u * float64(w))

        if flip {
            x = w - tmp
        } else {
            x = tmp
        }

    } else {
        x = min(w-1, x)
        x = max(x, 0)
    }

    return x, y
}

type NdsTexture struct {
    Width, Height int
	RepeatS, RepeatT   bool
	FlipS, FlipT       bool
    CachedTexture *[]uint8
}

func (t *NdsTexture) Sample(u, v float64) Color {
    x, y := getTextureCoords(
        u, v,
        t.Width, t.Height,
        t.RepeatT, t.RepeatS,
        t.FlipT, t.FlipS,
    )
    return t.getColor(x, y)
}

func (t *NdsTexture) BilinearSample(u, v float64) Color {
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

func (t *NdsTexture) getColor(x, y int) Color {

    idx := (x + y * t.Width) * 2

    if idx + 1 >= len(*t.CachedTexture) {
        return Black
    }

    data := binary.LittleEndian.Uint16((*t.CachedTexture)[idx:])

    return MakeColorFrom15Bit(
        uint8(data & 0b11111),
        uint8(data >> 5) & 0b11111,
        uint8(data >> 10) & 0b11111,
    )
}
