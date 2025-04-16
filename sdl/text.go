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
	initX, initY int32
	surW, surH    int32
	Status        Status
    Color, ColorAlt sdl.Color
    positionMethod string
}

func NewText(renderer *sdl.Renderer, parent Component, layout Layout, text string, fontSize int, color, colorAlt sdl.Color, positionMethod string) *Text {

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
		//X:        parent.GetLayout().X,
		//Y:        parent.GetLayout().Y,
		X:        layout.X,
		Y:        layout.Y,
		Z:        layout.Z,
		initX:        layout.X,
		initY:        layout.Y,
		Status:   s,
        Color: color,
        ColorAlt: colorAlt,
        positionMethod: positionMethod,
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


    switch b.positionMethod {
    case "centerParent":
        b.X, b.Y, b.W, b.H = positionCenter(b, b.parent)
    //case "evenlyVertical":
    //    b.X, b.Y, b.W, b.H = positionCenter(b, b.parent)
    //    distributeEvenlyVertical(b)
    case "relativeParent":
        l := Layout{X: b.initX, Y: b.initY, H: b.H, W: b.H, Z: b.Z}
        b.X, b.Y, b.W, b.H, b.Z = positionRelative(l, (*b.parent).GetLayout())
    case "":
    default: panic("Unknown position Method Text")
    }



	rect := sdl.Rect{X: b.X, Y: b.Y, W: surface.W, H: surface.H}

	surface.Free()

	b.renderer.Copy(texture, nil, &rect)

    texture.Destroy()

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
