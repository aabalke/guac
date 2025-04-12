package sdl

import (
	"github.com/aabalke33/guac/emu"
	"github.com/veandco/go-sdl2/sdl"
)

type Container struct {
    Renderer      *sdl.Renderer
	parent        *Component
	children      []*Component
	W, H, X, Y, Z int32
	Status        Status
	Emulator      *emu.Emulator
	color         sdl.Color
}

func NewContainer(renderer *sdl.Renderer, parent Component, layout Layout, emulator emu.Emulator, color sdl.Color) *Container {

	s := Status{
		Active:   true,
		Visible:  true,
		Hovered:  false,
		Selected: false,
	}

	b := Container{
        Renderer: renderer,
		color:    color,
		Emulator: &emulator,
		parent:   &parent,
		X:        layout.X,
		Y:        layout.Y,
		W:        layout.W,
		H:        layout.H,
		Z:        layout.Z,
		Status:   s,
	}

	b.Resize()

	return &b
}

func (b *Container) Update(event sdl.Event) bool {

	//if !b.Active {
	//	return
	//}

	ChildFuncUpdate(b, func(child *Component) bool {
		return (*child).Update(event)
	})

    return false
}

func (b *Container) View() {
	//if !b.Active {
	//	return
	//}

	b.X, b.Y, b.W, b.H = positionCenter(b, b.parent)

	distributeEvenlyVertical(b)

	if b.color.A != 0 {
		b.Renderer.SetDrawColor(b.color.R, b.color.G, b.color.B, b.color.A)
		rect := sdl.Rect{X: b.X, Y: b.Y, W: b.W, H: b.H}
		b.Renderer.FillRect(&rect)
	}

	ChildFunc(b, func(child *Component) {
		(*child).View()
	})
}

func (b *Container) Add(c Component) {
	b.children = append(b.children, &c)
}

func (b *Container) Resize() {
	ChildFunc(b, func(child *Component) {
		(*child).Resize()
	})
}

func (b *Container) GetChildren() []*Component {
	return b.children
}

func (b *Container) GetParent() *Component {
	return b.parent
}

func (b *Container) GetLayout() Layout {
	return Layout{X: b.X, Y: b.Y, H: b.H, W: b.W, Z: b.Z}
}

func (b *Container) GetStatus() Status {
	return b.Status
}

func (b *Container) SetChildren(c []*Component) {
	b.children = c
}

func (b *Container) SetStatus(s Status) {
	b.Status = s
}

func (b *Container) SetLayout(l Layout) {
	b.W = l.W
	b.H = l.H
	b.X = l.X
	b.Y = l.Y
	b.Z = l.Z
}
