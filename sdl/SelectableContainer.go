package sdl

import (
	"github.com/veandco/go-sdl2/sdl"
)

type SelectableContainer struct {
	Renderer        *sdl.Renderer
	parent          *Component
	children        []*Component
	Layout          Layout
	InitLayout      Layout
	Status          Status
	Color, ColorAlt sdl.Color
	positionMethod  string
}

func NewSelectableContainer(parent Component, layout Layout, color, colorAlt sdl.Color, positionMethod string) *SelectableContainer {

	s := Status{
		Active:   true,
		Visible:  true,
		Hovered:  false,
		Selected: false,
	}

	b := SelectableContainer{
		Renderer:       parent.GetRenderer(),
		Color:          color,
		ColorAlt:       colorAlt,
		parent:         &parent,
		Layout:         layout,
		InitLayout:     layout,
		Status:         s,
		positionMethod: positionMethod,
	}

	b.Resize()

	return &b
}

func (b *SelectableContainer) Update(event sdl.Event) bool {

	//if !b.Active {
	//	return
	//}

	ChildFuncUpdate(b, func(child *Component) bool {
		return (*child).Update(event)
	})

	return false
}

func (b *SelectableContainer) View() {

	if !b.Status.Visible {
		return
	}

	c := b.Color
	if b.Status.Selected {
		c = b.ColorAlt
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
		//l := Layout{X: b.initX, Y: b.initY, H: b.initH, W: b.initW, Z: b.Z}
		x, y, w, h, _ = positionRelative(b.InitLayout, *(*b.parent).GetLayout())
		//SetI32(&b.Layout.Z, z)
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

	b.Renderer.SetDrawColor(c.R, c.G, c.B, c.A)
	rect := sdl.Rect{X: GetI32(b.Layout.X), Y: GetI32(b.Layout.Y), W: GetI32(b.Layout.W), H: GetI32(b.Layout.H)}
	b.Renderer.FillRect(&rect)

	ChildFunc(b, func(child *Component) {
		(*child).View()
	})
}

func (b *SelectableContainer) Add(c Component) {
	b.children = append(b.children, &c)
}

func (b *SelectableContainer) Resize() {
	ChildFunc(b, func(child *Component) {
		(*child).Resize()
	})
}

func (b *SelectableContainer) GetChildren() []*Component {
	return b.children
}

func (b *SelectableContainer) GetParent() *Component {
	return b.parent
}

func (b *SelectableContainer) GetLayout() *Layout {
	return &b.Layout
}

func (b *SelectableContainer) GetStatus() Status {
	return b.Status
}

func (b *SelectableContainer) SetChildren(c []*Component) {
	b.children = c
}

func (b *SelectableContainer) SetStatus(s Status) {
	b.Status = s
}

func (b *SelectableContainer) SetLayout(l Layout) {
	b.Layout = l
}

func (b *SelectableContainer) GetRenderer() *sdl.Renderer {
	return b.Renderer
}
