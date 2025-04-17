package sdl

import (
	"math"

	gameboy "github.com/aabalke33/guac/emu/gb"
	"github.com/veandco/go-sdl2/sdl"
)

type GbMenu struct {
	Renderer    *sdl.Renderer
	parent      Component
	children    []*Component
	Layout      Layout
	ratio       float64
	Status      Status
	Gb          *gameboy.GameBoy
	color       sdl.Color
	SelectedIdx int
}

func NewGbMenu(parent Component, layout Layout, gb *gameboy.GameBoy, color sdl.Color) *GbMenu {

	ratio := 1.0

	s := Status{
		Active:   true,
		Visible:  true,
		Hovered:  false,
		Selected: false,
	}

	b := GbMenu{
		Renderer: parent.GetRenderer(),
		color:    color,
		Gb:       gb,
		parent:   parent,
		ratio:    ratio,
		Layout:   layout,
		Status:   s,
	}

	b.Resize()

	return &b
}

func (b *GbMenu) Update(event sdl.Event) bool {

	if !b.Status.Active {
		return false
	}

	childrenDirty := false

	switch e := event.(type) {
	case *sdl.KeyboardEvent:

		if e.State != sdl.RELEASED {
			break
		}

		switch e.Keysym.Sym {
		case sdl.K_DOWN:
			b.UpdateSelected(false)
			childrenDirty = true
		case sdl.K_UP:
			b.UpdateSelected(true)
			childrenDirty = true
		case sdl.K_RETURN:
			b.HandleSelected()
			childrenDirty = true
		case sdl.K_p:
			return true
		}
	}

	if childrenDirty {

		var updateText func(child *Component)

		updateText = func(child *Component) {
			switch t := (*child).(type) {
			case *Text:
				t.Dirty = true
				return
			}

			for _, c := range (*child).GetChildren() {
				updateText(c)
			}
		}

		ChildFunc(b, func(child *Component) {
			updateText(child)
		})
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
	winW, winH := win.GetSize()

	SetI32(&b.Layout.X, math.Floor(float64(winW)/2-float64(GetI32(b.Layout.W))/2))
	SetI32(&b.Layout.Y, math.Floor(float64(winH)/2-float64(GetI32(b.Layout.H))/2))

	x := GetI32(b.Layout.X)
	y := GetI32(b.Layout.Y)
	w := GetI32(b.Layout.W)
	h := GetI32(b.Layout.H)

	b.Renderer.SetDrawColor(b.color.R, b.color.G, b.color.B, b.color.A)
	rect := sdl.Rect{X: x, Y: y, W: w, H: h}
	b.Renderer.FillRect(&rect)

	ChildFunc(b, func(child *Component) {
		(*child).View()
	})
}

func (b *GbMenu) Add(c Component) {
	b.children = append(b.children, &c)
}

func (b *GbMenu) Resize() {
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

func (b *GbMenu) GetLayout() *Layout {
	return &b.Layout
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
	b.Layout = l
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

        switch p := b.parent.(type) {
        case *Scene:
            Gb.Close()

            for _, child := range p.GetChildren() {
                switch g := (*child).(type) {
                case *GbFrame:
                    g.SetStatus(Status{Active: false})
                }
            }

            b.SetStatus(Status{Active: false})
            Gb = gameboy.NewGameBoy()
            InitMainMenu(p, 0)
            return
        }

        panic("Parent of Gb Menu is not a Scene")

	default:
		panic("Gb Menu Container has no children")
	}
}
func (b *GbMenu) GetRenderer() *sdl.Renderer {
	return b.Renderer
}

func InitPauseMenu(renderer *sdl.Renderer, scene *Scene, gb *gameboy.GameBoy) {

	c := sdl.Color{R: 228, G: 199, B: 153, A: 255}
	c2 := sdl.Color{R: 255, G: 255, B: 255, A: 255}

	l := NewLayout(&scene.H, &scene.W, 0, 0, 2)
	pause := NewGbMenu(scene, l, gb, C_Brown)

	l = NewLayout(400, 200, 100, 100, 3)
	container := NewContainer(pause, l, C_Transparent, "evenlyVertical")

	text := "mute"
	if gb.Muted {
		text = "unmute"
	}

	container.Add(NewText(container, NewLayout(0, 0, 0, 0, 5), "resume", 48, c, c2, ""))
	container.Add(NewText(container, NewLayout(0, 0, 0, 0, 5), text, 48, c, c2, ""))
	container.Add(NewText(container, NewLayout(0, 0, 0, 0, 5), "exit", 48, c, c2, ""))
	pause.Add(container)
	//pause.Add(NewText(s.Renderer, container, 5, "always save your game in the emulator before exiting", 16))
	scene.Add(pause)

	pause.InitOptions()
}
