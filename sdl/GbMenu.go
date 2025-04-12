package sdl

import (
	"math"

	gameboy "github.com/aabalke33/guac/emu/gb"
	"github.com/veandco/go-sdl2/sdl"
)

type GbMenu struct {
    Renderer *sdl.Renderer
	parent      Component
	children    []*Component
	X, Y, Z     int32
	W, H        *int32
	ratio       float64
	Status      Status
	Gb          *gameboy.GameBoy
	color       sdl.Color
	SelectedIdx int
}

func NewGbMenu(renderer *sdl.Renderer, parent Component, h, w *int32, x, y, z int32, gb *gameboy.GameBoy, color sdl.Color) *GbMenu {

	ratio := 1.0

	s := Status{
		Active:   true,
		Visible:  true,
		Hovered:  false,
		Selected: false,
	}

	b := GbMenu{
        Renderer: renderer,
		color:  color,
		Gb:     gb,
		parent: parent,
		ratio:  ratio,
		X:      x,
		Y:      y,
		W:      w,
		H:      h,
		Z:      z,
		Status: s,
	}

	b.Resize()

	return &b
}

func (b *GbMenu) Update(event sdl.Event) bool {

	if !b.Status.Active {
		return false
	}

	switch e := event.(type) {
	case *sdl.KeyboardEvent:

		if e.State != sdl.RELEASED {
			break
		}

		switch e.Keysym.Sym {
		case sdl.K_DOWN:
			b.UpdateSelected(false)
		case sdl.K_UP:
			b.UpdateSelected(true)
		case sdl.K_RETURN:
			b.HandleSelected()
        case sdl.K_p:
            return true
		}
	}

	ChildFuncUpdate(b, func(child *Component) bool {
		return (*child).Update(event)
	})

    return false
}

func (b *GbMenu) View() {
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

func (b *GbMenu) Add(c Component) {
	b.children = append(b.children, &c)
}

func (b *GbMenu) Resize() {
	//b.W = int32(math.Floor(float64(*b.H) * b.ratio))

	ChildFunc(b, func(child *Component) {
		(*child).Resize()
	})
}

func (b *GbMenu) GetChildren() []*Component {
	return b.children
}

func (b *GbMenu) GetParent() *Component {
	return &b.parent
}

func (b *GbMenu) GetLayout() Layout {
	return Layout{X: b.X, Y: b.Y, H: *b.H, W: *b.W, Z: b.Z}
}

func (b *GbMenu) GetStatus() Status {
	return b.Status
}

func (b *GbMenu) SetChildren(c []*Component) {
	b.children = c
}

func (b *GbMenu) SetStatus(s Status) {
	b.Status = s
}

func (b *GbMenu) SetLayout(l Layout) {
	//b.W = l.W
	//b.H = l.H
	b.X = l.X
	b.Y = l.Y
	b.Z = l.Z
}

func (b *GbMenu) InitOptions() {

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

func (b *GbMenu) UpdateSelected(reverse bool) {

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

func (b *GbMenu) HandleSelected() {

	switch b.SelectedIdx {
	case 0:
		b.Gb.TogglePause()
		b.Status.Active = false
	case 1:

		textComponent := (*(*b.children[0]).GetChildren()[1])
		switch c := textComponent.(type) {
		case *Text:
			if muted := b.Gb.ToggleMute(); muted {
				c.UpdateText("unmute")
				break
			}

			c.UpdateText("mute")
		}

	case 2:
		b.parent.SetStatus(Status{Active: false})
	default:
		panic("Gb Menu Container has no children")
	}
}
