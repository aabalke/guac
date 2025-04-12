package sdl

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type Text struct {
	renderer      *sdl.Renderer
	font          *ttf.Font
	text          string
	parent        *Component
	children      []*Component
	W, H, X, Y, Z int32
	surW, surH    int32
	Status        Status
    Color, ColorAlt sdl.Color
}

func NewText(renderer *sdl.Renderer, parent Component, z int32, text string, fontSize int, color, colorAlt sdl.Color) *Text {

	font, err := ttf.OpenFont("./museo.otf", fontSize)
	if err != nil {
		panic(err)
	}

	s := Status{
		Active:   true,
		Visible:  true,
		Hovered:  false,
		Selected: false,
	}

	b := Text{
		font:     font,
		text:     text,
		renderer: renderer,
		parent:   &parent,
		X:        parent.GetLayout().X,
		Y:        parent.GetLayout().Y,
		Z:        z,
		Status:   s,
        Color: color,
        ColorAlt: colorAlt,
	}

	b.Resize()

	return &b
}

func (b *Text) Update(event sdl.Event) bool {

	ChildFuncUpdate(b, func(child *Component) bool{
		return (*child).Update(event)
	})

    return false
}

func (b *Text) View() {

	if !b.Status.Visible {
		return
	}

    c := b.Color
	if b.Status.Selected {
        c = b.ColorAlt
	}

	//surface, err := b.font.RenderUTF8Solid(b.text, c)
	surface, err := b.font.RenderUTF8Blended(b.text, c)
	if err != nil {
		panic(err)
	}

	texture, err := b.renderer.CreateTextureFromSurface(surface)
	if err != nil {
		panic(err)
	}

	b.W, b.H = surface.W, surface.H

	rect := sdl.Rect{X: b.X, Y: b.Y, W: surface.W, H: surface.H}

	surface.Free()

	b.renderer.Copy(texture, nil, &rect)

	ChildFunc(b, func(child *Component) {
		(*child).View()
	})
}

func (b *Text) Add(c Component) {
	b.children = append(b.children, &c)
}

func (b *Text) Resize() {

	//if !b.Active {
	//	return
	//}

	ChildFunc(b, func(child *Component) {
		(*child).Resize()
	})
}

func (b *Text) GetChildren() []*Component {
	return b.children
}

func (b *Text) GetParent() *Component {
	return b.parent
}

func (b *Text) GetLayout() Layout {
	return Layout{X: b.X, Y: b.Y, H: b.H, W: b.W, Z: b.Z}
}

func (b *Text) GetStatus() Status {
	return b.Status
}

func (b *Text) SetChildren(c []*Component) {
	b.children = c
}

func (b *Text) SetStatus(s Status) {
	b.Status = s
}

func (b *Text) SetLayout(l Layout) {
	b.W = l.W
	b.H = l.H
	b.X = l.X
	b.Y = l.Y
	b.Z = l.Z
}

func (b *Text) UpdateText(s string) {
	b.text = s
	b.View()
}
