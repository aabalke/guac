package sdl

import (
	"image"
	//"math"
	"os"
	"unsafe"

	_ "image/jpeg"
	_ "image/png"

	"github.com/veandco/go-sdl2/sdl"
)

type Image struct {
	Renderer   *sdl.Renderer
	texture    *sdl.Texture
	pixels     []byte
	parent     *Component
	children   []*Component
	W, H, X, Y, Z, tH, tW int32
	initW, initH, initX, initY, initZ int32
	Status     Status
    positionMethod string
}

func NewImage(Renderer *sdl.Renderer, parent Component, layout Layout, filepath string, positionMethod string) *Image {

    pixels, tW, tH := GetImage(filepath)

    texture, _ := Renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, tW, tH)
    texture.Update(nil, unsafe.Pointer(&(pixels)[0]), int(tW*4))

	s := Status{
		Active:   true,
		Visible:  true,
		Hovered:  false,
		Selected: false,
	}

	b := Image{
		Renderer: Renderer,
		parent:   &parent,
		texture:  texture,
		pixels:   pixels,
		X:        layout.X,
		Y:        layout.Y,
		W:        layout.W,
		H:        layout.H,
		Z:        layout.Z,
		initX:        layout.X,
		initY:        layout.Y,
		initW:        layout.W,
		initH:        layout.H,
		initZ:        layout.Z,
		tH:       tH,
		tW:       tW,
		Status:   s,
        positionMethod: positionMethod,
	}

	b.Resize()

	return &b
}

func (b *Image) Close() {
    b.texture.Destroy()
}

func GetImage(filepath string) (pixels []byte, width, height int32) {

    f, err := os.Open(filepath)
    if err != nil {
        panic(err)
    }

    defer f.Close()

    img, _, err := image.Decode(f)
    if err != nil {
        panic(err)
    }

    width =  int32(img.Bounds().Max.X)
    height = int32(img.Bounds().Max.Y)

    for y := range int(height) {
        for x := range int(height) {
            r, g, b, a := img.At(x, y).RGBA()

            r1 := uint8(float64(r) / 0x10000 * 0x100)
            g1 := uint8(float64(g) / 0x10000 * 0x100)
            b1 := uint8(float64(b) / 0x10000 * 0x100)
            a1 := uint8(float64(a) / 0x10000 * 0x100)

            pixels = append(pixels, r1)
            pixels = append(pixels, g1)
            pixels = append(pixels, b1)
            pixels = append(pixels, a1)
        }
    }

    return pixels, width, height
}



func (b *Image) Update(event sdl.Event) bool {

	//if !b.Active {
	//	return
	//}

	ChildFuncUpdate(b, func(child *Component) bool {
		return (*child).Update(event)
	})

    return false
}

func (b *Image) View() {
	//if !b.Active {
	//	return
	//}
    switch b.positionMethod {
    case "centerParent":
        b.X, b.Y, b.W, b.H = positionCenter(b, b.parent)
    case "evenlyVertical":
        b.X, b.Y, b.W, b.H = positionCenter(b, b.parent)
        distributeEvenlyVertical(b)
    case "relativeParent":
        l := Layout{X: b.initX, Y: b.initY, H: b.initH, W: b.initW, Z: b.Z}
        b.X, b.Y, b.W, b.H, b.Z = positionRelative(l, (*b.parent).GetLayout())
    case "":
    default: panic("position method unknown")
    }


	//b.Renderer.Clear()
	//win, _ := b.Renderer.GetWindow()
	//w, h := win.GetSize()

	//b.X = int32(math.Floor(float64(w)/2 - float64(b.W)/2))
	//b.Y = int32(math.Floor(float64(h)/2 - float64(b.H)/2))

	rect := sdl.Rect{X: b.X, Y: b.Y, W: b.W, H: b.H}
	b.Renderer.Copy(b.texture, nil, &rect)

	ChildFunc(b, func(child *Component) {
		(*child).View()
	})
}

func (b *Image) Add(c Component) {
	b.children = append(b.children, &c)
}

func (b *Image) Resize() {
	ChildFunc(b, func(child *Component) {
		(*child).Resize()
	})
}

func (b *Image) GetChildren() []*Component {
	return b.children
}

func (b *Image) GetParent() *Component {
	return b.parent
}

func (b *Image) GetLayout() Layout {
	return Layout{X: b.X, Y: b.Y, H: b.H, W: b.W, Z: b.Z}
}

func (b *Image) GetStatus() Status {
	return b.Status
}

func (b *Image) SetChildren(c []*Component) {
	b.children = c
}

func (b *Image) SetStatus(s Status) {
	b.Status = s
}

func (b *Image) SetLayout(l Layout) {
	b.W = l.W
	b.H = l.H
	b.X = l.X
	b.Y = l.Y
	b.Z = l.Z
}
