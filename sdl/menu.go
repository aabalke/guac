package sdl

import (
	comp "github.com/aabalke33/go-sdl2-components/Components"
	"github.com/veandco/go-sdl2/sdl"
)

type Menu struct {
	parent        *comp.Component
	children      []*comp.Component
	x, y, w, h, z int32
	color         sdl.Color
	Active        bool
    Hidden        bool
}

func NewMenu(parent comp.Component, x, y, w, h, z int32, color sdl.Color) *Menu {
	return &Menu{
		parent:   &parent,
		children: []*comp.Component{},
		x:        x,
		y:        y,
		w:        w,
		h:        h,
		z:        z,
		color:    color,
		Active:   true,
	}
}

func (b *Menu) Update(dt float64, event sdl.Event) {

	switch e := event.(type) {
	case *sdl.MouseButtonEvent:
		if e.Type == sdl.MOUSEBUTTONDOWN {
			b.Active = !b.Active
		}

	}
    
    comp.ChildFunc(b, func(child *comp.Component) {
        (*child).Update(1/comp.FPS, event)
    })
}

func (b *Menu) View(renderer *sdl.Renderer) {
	if !b.Active {
		return
	}

	renderer.SetDrawColor(b.color.R, b.color.G, b.color.B, b.color.A)
	rect := sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h}
	renderer.FillRect(&rect)

	comp.ChildFunc(b, func(child *comp.Component) {
		(*child).View(renderer)
	})
}

func (b *Menu) Add(c comp.Component) {
	b.children = append(b.children, &c)
}

func (b *Menu) GetZ() int32 {
	return b.z
}

func (b *Menu) Resize() {
	return
}

func (b *Menu) IsActive() bool {
	return b.Active
}

func (b *Menu) GetChildren() []*comp.Component {
	return b.children
}

func (b *Menu) GetParent() *comp.Component {
	return b.parent
}

func (b *Menu) GetSize() (int32, int32) {
	return b.h, b.w
}

func (b *Menu) SetChildren(c []*comp.Component) {
    b.children = c
}
