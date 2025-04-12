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

	GetLayout() Layout
    SetLayout(Layout)

    GetStatus() Status
    SetStatus(Status)
}

type Layout struct {
    X, Y, H, W, Z int32
}

type Status struct {
    Active, Visible, Hovered, Selected bool
}
