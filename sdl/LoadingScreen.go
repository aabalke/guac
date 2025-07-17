package sdl

import (
	"math"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

type LoadingScreen struct {
	Renderer    *sdl.Renderer
	parent      Component
	children    []*Component
	Layout      Layout
	ratio       float64
	Status      Status
	color       sdl.Color
	SelectedIdx int
}

func NewLoadingScreen(parent Component, layout Layout, color sdl.Color, duration time.Duration) *LoadingScreen {

	ratio := 1.0

	s := Status{
		Active:   true,
		Visible:  true,
		Hovered:  false,
		Selected: false,
	}

	b := LoadingScreen{
		Renderer: parent.GetRenderer(),
		color:    color,
		parent:   parent,
		ratio:    ratio,
		Layout:   layout,
		Status:   s,
	}

	timer := time.NewTimer(duration)

	go func() {
		<-timer.C
		b.Status.Active = false
	}()

	b.Resize()

	return &b
}

func (b *LoadingScreen) Update(event sdl.Event) bool {

	if !b.Status.Active {
		return false
	}

	switch e := event.(type) {
	case *sdl.KeyboardEvent:

		if e.State != sdl.RELEASED {
			break
		}

		switch e.Keysym.Sym {
		case sdl.K_RETURN:
			return true
		}
	}

	ChildFuncUpdate(b, func(child *Component) bool {
		return (*child).Update(event)
	})

	return false
}

func (b *LoadingScreen) View() {
	if !b.Status.Active || !b.Status.Visible {
		return
	}

	win, _ := b.Renderer.GetWindow()
	winW, winH := win.GetSize()

	SetI32(&b.Layout.X, math.Floor(float64(winW)/2-float64(GetI32(b.Layout.W))/2))
	SetI32(&b.Layout.Y, math.Floor(float64(winH)/2-float64(GetI32(b.Layout.H))/2))

	x := GetI32(b.Layout.X)
	y := GetI32(b.Layout.Y)
	w := GetI32(b.Layout.W)
	h := GetI32(b.Layout.H)
	rect := sdl.Rect{X: x, Y: y, W: w, H: h}

	b.Renderer.SetDrawColor(b.color.R, b.color.G, b.color.B, b.color.A)
	b.Renderer.FillRect(&rect)

	ChildFunc(b, func(child *Component) {
		(*child).View()
	})
}

func (b *LoadingScreen) Add(c Component) {
	b.children = append(b.children, &c)
}

func (b *LoadingScreen) Resize() {
	ChildFunc(b, func(child *Component) {
		(*child).Resize()
	})
}

func (b *LoadingScreen) GetChildren() []*Component {
	return b.children
}

func (b *LoadingScreen) GetParent() *Component {
	return &b.parent
}

func (b *LoadingScreen) GetLayout() *Layout {
	return &b.Layout
}

func (b *LoadingScreen) GetStatus() Status {
	return b.Status
}

func (b *LoadingScreen) SetChildren(c []*Component) {
	b.children = c
}

func (b *LoadingScreen) SetStatus(s Status) {
	b.Status = s
}

func (b *LoadingScreen) SetLayout(l Layout) {
	b.Layout = l
}

func (b *LoadingScreen) GetRenderer() *sdl.Renderer {
	return b.Renderer
}

func InitLoadingScreen(renderer *sdl.Renderer, scene *Scene, duration time.Duration) {

	z := 25

	c := C_White

	l := NewLayout(&scene.H, &scene.W, 0, 0, z)
	loadingScreen := NewLoadingScreen(scene, l, C_Green, duration)

	l = NewLayout(100, 600, 0, 0, z+1)
	container := NewContainer(loadingScreen, l, C_Transparent, "evenlyVertical")
	container.Add(NewText(container, NewLayout(0, 0, 0, 0, z+2), "guac emulator", 48, c, c, ""))

	l = NewLayout(50, 600, 0, 0, z+1)
	container2 := NewContainer(container, l, C_Transparent, "evenlyVertical")
	container2.Add(NewText(container2, NewLayout(0, 0, 0, 0, z+2), "alpha 1.0.0", 24, c, c, ""))
	container2.Add(NewText(container2, NewLayout(0, 0, 0, 0, z+2), "developed by aaron balke", 24, c, c, ""))

	container.Add(container2)
	loadingScreen.Add(container)
	scene.Add(loadingScreen)
}
