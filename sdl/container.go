package sdl

import (
	"github.com/veandco/go-sdl2/sdl"
)

type Container struct {
    Renderer      *sdl.Renderer
	parent        *Component
	children      []*Component
	//W, H, X, Y, Z int32
	//initW, initH, initX, initY, initZ int32

    Layout Layout
    InitLayout Layout


	Status        Status
	color         sdl.Color
    positionMethod string
}

func NewContainer(renderer *sdl.Renderer, parent Component, layout Layout, color sdl.Color, positionMethod string) *Container {

	s := Status{
		Active:   true,
		Visible:  true,
		Hovered:  false,
		Selected: false,
	}

	b := Container{
        Renderer: renderer,
		color:    color,
		parent:   &parent,
        Layout: layout,
        InitLayout: layout,
		//X:        layout.X,
		//Y:        layout.Y,
		//W:        layout.W,
		//H:        layout.H,
		//Z:        layout.Z,
		//initX:        layout.X,
		//initY:        layout.Y,
		//initW:        layout.W,
		//initH:        layout.H,
		//initZ:        layout.Z,


		Status:   s,
        positionMethod: positionMethod,
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

    var x, y, w, h, z int32

    switch b.positionMethod {
    case "centerParent":
        x, y, w, h = positionCenter(b, b.parent)
    case "evenlyVertical":
        x, y, w, h = positionCenter(b, b.parent)
        distributeEvenlyVertical(b)
    case "relativeParent":
        //l := Layout{X: b.initX, Y: b.initY, H: b.initH, W: b.initW, Z: b.Z}
        x, y, w, h, z = positionRelative(b.InitLayout, (*b.parent).GetLayout())
        SetI32(&b.Layout.Z, z)
    case "":
    default: panic("position method unknown")
    }

    SetI32(&b.Layout.X, x)
    SetI32(&b.Layout.Y, y)
    SetI32(&b.Layout.W, w)
    SetI32(&b.Layout.H, h)

	if b.color.A != 0 {
		b.Renderer.SetDrawColor(b.color.R, b.color.G, b.color.B, b.color.A)
		rect := sdl.Rect{X: GetI32(b.Layout.X), Y: GetI32(b.Layout.Y), W: GetI32(b.Layout.W), H: GetI32(b.Layout.H)}
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
	return b.Layout
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
    b.Layout = l
}
