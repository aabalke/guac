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
	Renderer       *sdl.Renderer
	texture        *sdl.Texture
	pixels         []byte
	parent         *Component
	children       []*Component
	Layout         Layout
	InitLayout     Layout
	tH, tW         int32
	Status         Status
	positionMethod string
}

func NewImage(parent Component, layout Layout, filepath string, positionMethod string) *Image {

	Renderer := parent.GetRenderer()

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
		Renderer:       Renderer,
		parent:         &parent,
		texture:        texture,
		pixels:         pixels,
		Layout:         layout,
		InitLayout:     layout,
		tH:             tH,
		tW:             tW,
		Status:         s,
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

	width = int32(img.Bounds().Max.X)
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

	var x, y, w, h, z int32
	switch b.positionMethod {
	case "centerHorizontal":
		x, y, w, h = positionHorizontal(b, b.parent)
	case "centerParent":
		x, y, w, h = positionCenter(b, b.parent)
	case "evenlyVertical":
		x, y, w, h = positionCenter(b, b.parent)
		distributeEvenlyVertical(b)
	case "relativeParent":
		//l := Layout{X: b.initX, Y: b.initY, H: b.initH, W: b.initW, Z: b.Z}
		x, y, w, h, z = positionRelative(b.InitLayout, *(*b.parent).GetLayout())
		SetI32(&b.Layout.Z, z)
	case "":
		x = GetI32(b.Layout.X)
		y = GetI32(b.Layout.Y)
		w = GetI32(b.Layout.W)
		h = GetI32(b.Layout.H)
	default:
		panic("position method unknown")
	}

	SetI32(&b.Layout.X, x)
	SetI32(&b.Layout.Y, y)
	SetI32(&b.Layout.W, w)
	SetI32(&b.Layout.H, h)

	rect := sdl.Rect{X: GetI32(b.Layout.X), Y: GetI32(b.Layout.Y), W: GetI32(b.Layout.W), H: GetI32(b.Layout.H)}
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

func (b *Image) GetLayout() *Layout {
	return &b.Layout
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
	b.Layout = l
}
func (b *Image) GetRenderer() *sdl.Renderer {
	return b.Renderer
}
