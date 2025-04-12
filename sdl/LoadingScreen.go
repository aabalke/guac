package sdl

import (
	"math"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

type LoadingScreen struct {
    Renderer *sdl.Renderer
	parent      Component
	children    []*Component
	X, Y, Z     int32
	W, H        *int32
	ratio       float64
	Status      Status
	color       sdl.Color
	SelectedIdx int
}

func NewLoadingScreen(renderer *sdl.Renderer, parent Component, h, w *int32, x, y, z int32, color sdl.Color, duration time.Duration) *LoadingScreen {

	ratio := 1.0

	s := Status{
		Active:   true,
		Visible:  true,
		Hovered:  false,
		Selected: false,
	}

	b := LoadingScreen{
        Renderer: renderer,
		color:  color,
		parent: parent,
		ratio:  ratio,
		X:      x,
		Y:      y,
		W:      w,
		H:      h,
		Z:      z,
		Status: s,
	}

    timer := time.NewTimer(duration)

    go func() {
        <- timer.C
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
			b.HandleSelected()
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
	w, h := win.GetSize()

	b.X = int32(math.Floor(float64(w)/2 - float64(*b.W)/2))
	b.Y = int32(math.Floor(float64(h)/2 - float64(*b.H)/2))

	b.Renderer.SetDrawColor(b.color.R, b.color.G, b.color.B, b.color.A)
	rect := sdl.Rect{X: b.X, Y: b.Y, W: *b.W, H: *b.H}
	b.Renderer.FillRect(&rect)

	ChildFunc(b, func(child *Component) {
		(*child).View()
	})
}

func (b *LoadingScreen) Add(c Component) {
	b.children = append(b.children, &c)
}

func (b *LoadingScreen) Resize() {
	//b.W = int32(math.Floor(float64(*b.H) * b.ratio))

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

func (b *LoadingScreen) GetLayout() Layout {
	return Layout{X: b.X, Y: b.Y, H: *b.H, W: *b.W, Z: b.Z}
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
	//b.W = l.W
	//b.H = l.H
	b.X = l.X
	b.Y = l.Y
	b.Z = l.Z
}

func (b *LoadingScreen) InitOptions() {

    return

	container := *(b.GetChildren()[0])
	options := container.GetChildren()

	oStatus := (*options[0]).GetStatus()
	nStatus := Status{
		Active:   oStatus.Active,
		Visible:  oStatus.Visible,
		Hovered:  oStatus.Hovered,
		Selected: true,
	}
	(*options[0]).SetStatus(nStatus)
}

func (b *LoadingScreen) UpdateSelected(reverse bool) {
    return

	originalIdx := b.SelectedIdx

	if reverse {
		b.SelectedIdx--
	} else {
		b.SelectedIdx++
	}

	if b.SelectedIdx < 0 {
		b.SelectedIdx = 0
	}

	container := *(b.GetChildren()[0])
	options := container.GetChildren()

	if b.SelectedIdx >= len(options) {
		b.SelectedIdx = len(options) - 1
	}

	oStatus := (*options[originalIdx]).GetStatus()
	nStatus := Status{
		Active:   oStatus.Active,
		Visible:  oStatus.Visible,
		Hovered:  oStatus.Hovered,
		Selected: false,
	}
	(*options[originalIdx]).SetStatus(nStatus)

	oStatus = (*options[b.SelectedIdx]).GetStatus()
	nStatus = Status{
		Active:   oStatus.Active,
		Visible:  oStatus.Visible,
		Hovered:  oStatus.Hovered,
		Selected: true,
	}
	(*options[b.SelectedIdx]).SetStatus(nStatus)
}

func (b *LoadingScreen) HandleSelected() {
    b.Status.Active = false
}
