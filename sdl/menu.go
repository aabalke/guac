package sdl

//
//import (
//	comp "github.com/aabalke33/go-sdl2-components/Components"
//	"github.com/veandco/go-sdl2/sdl"
//)
//
//type Menu struct {
//	parent        *comp.Component
//	children      []*comp.Component
//	X, Y, W, H, Z int32
//	color         sdl.Color
//	Status        comp.Status
//	Hidden        bool
//}
//
//func NewMenu(parent comp.Component, x, y, w, h, z int32, color sdl.Color) *Menu {
//
//	s := comp.Status{
//		Active:   true,
//		Visible:  true,
//		Hovered:  false,
//		Selected: false,
//	}
//
//    m := &Menu{
//		parent:   &parent,
//		children: []*comp.Component{},
//		X:        x,
//		Y:        y,
//		W:        w,
//		H:        h,
//		Z:        z,
//		color:    color,
//		Status:   s,
//	}
//
//    return m
//}
//
//func (b *Menu) Update(dt float64, event sdl.Event) {
//
//	switch e := event.(type) {
//	case *sdl.MouseButtonEvent:
//		if e.Type == sdl.MOUSEBUTTONDOWN {
//			b.Status.Active = !b.Status.Active
//		}
//	}
//
//	comp.ChildFunc(b, func(child *comp.Component) {
//		(*child).Update(1/comp.FPS, event)
//	})
//}
//
//func (b *Menu) View(renderer *sdl.Renderer) {
//	//if !b.Active {
//	//	return
//	//}
//
//	renderer.SetDrawColor(b.color.R, b.color.G, b.color.B, b.color.A)
//	rect := sdl.Rect{X: b.X, Y: b.Y, W: b.W, H: b.H}
//	renderer.FillRect(&rect)
//
//	comp.ChildFunc(b, func(child *comp.Component) {
//		(*child).View(renderer)
//	})
//}
//
//func (b *Menu) Add(c comp.Component) {
//	b.children = append(b.children, &c)
//}
//
//func (b *Menu) Resize() {
//	return
//}
//
//func (b *Menu) GetChildren() []*comp.Component {
//	return b.children
//}
//
//func (b *Menu) GetParent() *comp.Component {
//	return b.parent
//}
//
//func (b *Menu) GetLayout() comp.Layout {
//	return comp.Layout{X: b.X, Y: b.Y, H: b.H, W: b.W, Z: b.Z}
//}
//
//func (b *Menu) GetStatus() comp.Status {
//	return b.Status
//}
//
//func (b *Menu) SetChildren(c []*comp.Component) {
//	b.children = c
//}
//
//func (b *Menu) SetStatus(s comp.Status) {
//    b.Status = s
//}
//
//func (b *Menu) SetLayout(l comp.Layout) {
//    b.W = l.W
//    b.H = l.H
//    b.X = l.X
//    b.Y = l.Y
//    b.Z = l.Z
//}
//
//func (b *Menu) InitOptions() {
//
//    container := *(b.children[0])
//    options := container.GetChildren()
//
//    oStatus := (*options[0]).GetStatus()
//    nStatus := comp.Status{
//        Active: oStatus.Active,
//        Visible: oStatus.Visible,
//        Hovered: oStatus.Hovered,
//        Selected: true,
//    }
//    (*options[0]).SetStatus(nStatus)
//}
