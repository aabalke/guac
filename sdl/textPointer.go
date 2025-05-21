package sdl

import (
    "fmt"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type TextPointer struct {
	Renderer        *sdl.Renderer
	Texture         *sdl.Texture
	Dirty           bool
	font            *ttf.Font
	text            *uint32
    prefix          string
	currText            string
	parent          *Component
	children        []*Component
	Layout          Layout
	InitLayout      Layout
	surW, surH      int32
	Status          Status
	Color, ColorAlt sdl.Color
	positionMethod  string
}

func NewTextPointer(parent Component, layout Layout, prefix string, text *uint32, fontSize int, color, colorAlt sdl.Color, positionMethod string) *TextPointer {

	font, err := ttf.OpenFont("./mono.ttf", fontSize)
	if err != nil {
		panic(err)
	}

	s := Status{
		Active:   true,
		Visible:  true,
		Hovered:  false,
		Selected: false,
	}

	b := TextPointer{
		font:           font,
        prefix:         prefix,
		text:           text,
		Renderer:       parent.GetRenderer(),
		parent:         &parent,
		Layout:         layout,
		InitLayout:     layout,
		Status:         s,
		Color:          color,
		ColorAlt:       colorAlt,
		positionMethod: positionMethod,
	}

	b.Resize()

	return &b
}

func (b *TextPointer) Update(event sdl.Event) bool {

	ChildFuncUpdate(b, func(child *Component) bool {
		return (*child).Update(event)
	})

	return false
}

func (b *TextPointer) View() {

	if !b.Status.Visible {
		return
	}

    if fmt.Sprintf("%02s %08X", b.prefix, *b.text) != b.currText {

        b.Dirty = true
        b.currText = fmt.Sprintf("%02s %08X", b.prefix, *b.text)
    }

	if b.Dirty {
		b.Dirty = false
		b.RenderText()
	}

	var x, y, w, h int32
	switch b.positionMethod {
	case "centerHorizontal":
		x, y, w, h = positionHorizontal(b, b.parent)
	case "centerParent":
		x, y, w, h = positionCenter(b, b.parent)
	case "evenlyVertical":
		x, y, w, h = positionCenter(b, b.parent)
		distributeEvenlyVertical(b)
	case "relativeParent":
		pLayout := *(*b.parent).GetLayout()
		x, y, w, h, _ = positionRelative(b.InitLayout, pLayout)
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

	rect := sdl.Rect{X: GetI32(b.Layout.X), Y: GetI32(b.Layout.Y), W: b.surW, H: b.surH}

	b.Renderer.Copy(b.Texture, nil, &rect)

	ChildFunc(b, func(child *Component) {
		(*child).View()
	})
}

func (b *TextPointer) Add(c Component) {
	b.children = append(b.children, &c)
}

func (b *TextPointer) Resize() {

	b.Dirty = true

	ChildFunc(b, func(child *Component) {
		(*child).Resize()
	})
}

func (b *TextPointer) GetChildren() []*Component {
	return b.children
}

func (b *TextPointer) GetParent() *Component {
	return b.parent
}

func (b *TextPointer) GetLayout() *Layout {
	return &b.Layout
}

func (b *TextPointer) GetStatus() Status {
	return b.Status
}

func (b *TextPointer) SetChildren(c []*Component) {
	b.children = c
}

func (b *TextPointer) SetStatus(s Status) {
	b.Status = s
}

func (b *TextPointer) SetLayout(l Layout) {
	b.Layout = l
}

func (b *TextPointer) UpdateText(s *uint32) {
	b.text = s
	b.Dirty = true
	b.View()
}
func (b *TextPointer) GetRenderer() *sdl.Renderer {
	return b.Renderer
}

func (b *TextPointer) RenderText() {

	// cleanup previous
	b.Texture.Destroy()

	c := b.Color
	if b.Status.Selected {
		c = b.ColorAlt
	}

	// only change surface blend on change??
	//surface, err := b.font.RenderUTF8Solid(b.text, c)
	surface, err := b.font.RenderUTF8Blended(b.currText, c)
	if err != nil {
		panic(err)
	}

	b.Texture, err = b.Renderer.CreateTextureFromSurface(surface)
	if err != nil {
		panic(err)
	}

	if surface.W == 0 || surface.H == 0 {
		panic("Surface is 0")
	}

	b.surH = surface.H
	b.surW = surface.W

	SetI32(&b.Layout.W, surface.W)
	SetI32(&b.Layout.H, surface.H)

	surface.Free()
}
