package gl

import (
	"math"
)

type VRAM interface {
    ReadTexture(uint32) uint8
    ReadPalTexture(uint32) uint8
}

type Texture struct {
    Width, Height       int
	RepeatS, RepeatT    bool
	FlipS, FlipT        bool
    CachedTexture       *[]Color
    Mode uint8
    TextureAlpha float64
}

func (t *Texture) Sample(u, v float64) Color {
    x, y := t.getTextureCoords(u, v)
    return t.getColor(x, y)
}

func (t *Texture) BilinearSample(u, v float64) Color {
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

func (t *Texture) getColor(x, y uint32) Color {

    idx := (x + y * uint32(t.Width))
    if idx >= uint32(len(*t.CachedTexture)) {
        return Transparent
    }

    return (*t.CachedTexture)[idx]
}

func (t *Texture) getTextureCoords(u, v float64) (uint32, uint32) {

	x := int(u * float64(t.Width))
	y := int(v * float64(t.Height))

    if t.RepeatT {

        flip := t.FlipT && uint(math.Floor(v)) & 1 == 1
        v -= math.Floor(v)
        tmp := int(v * float64(t.Height))

        // does tmp need - 1 not just flip??

        if flip {
            y = t.Height - tmp - 1
        } else {
            y = tmp
        }

    } else {
        y = min(t.Height-1, y)
        y = max(y, 0)
    }

    if t.RepeatS {
        flip := t.FlipS && uint(math.Floor(u)) & 1 == 1 
        u -= math.Floor(u)
        tmp := int(u * float64(t.Width))

        // does tmp need - 1 not just flip??

        if flip {
            x = t.Width - tmp - 1
        } else {
            x = tmp
        }

    } else {
        x = min(t.Width-1, x)
        x = max(x, 0)
    }

    return uint32(x), uint32(y)
}

type BilinearCoords struct {
    X,  Y float64
    X0, Y0 uint32
    X1, Y1 uint32
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
        X0: uint32(x0),
        Y0: uint32(y0),
        X1: uint32(x1),
        Y1: uint32(y1),
    }
}
