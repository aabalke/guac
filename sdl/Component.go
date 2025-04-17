package sdl

import (
	"github.com/veandco/go-sdl2/sdl"
)

type Component interface {
	Update(event sdl.Event) bool
	View()
	Resize()

	Add(component Component)
	GetChildren() []*Component
	SetChildren([]*Component)
	GetParent() *Component

	GetLayout() *Layout
	SetLayout(Layout)

	GetStatus() Status
	SetStatus(Status)

	GetRenderer() *sdl.Renderer
}

type Layout struct {
	X, Y, H, W, Z I32
}

func NewLayout(h, w, x, y, z I32) Layout {
	l := Layout{}
	SetI32(&l.H, h)
	SetI32(&l.W, w)
	SetI32(&l.X, x)
	SetI32(&l.Y, y)
	SetI32(&l.Z, z)
	return l
}

type Status struct {
	Active, Visible, Hovered, Selected bool
}
