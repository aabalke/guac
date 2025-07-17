package sdl

import (
	"math"

	"github.com/aabalke33/guac/emu/gba"
	"github.com/aabalke33/guac/oto"
	"github.com/veandco/go-sdl2/sdl"
)

type GbaMenu struct {
	Renderer    *sdl.Renderer
	parent      Component
	children    []*Component
	Layout      Layout
	ratio       float64
	Status      Status
	Gba          *gba.GBA
	color       sdl.Color
	SelectedIdx int
}

func NewGbaMenu(parent Component, layout Layout, gba *gba.GBA, color sdl.Color) *GbaMenu {

	ratio := 1.0

	s := Status{
		Active:   true,
		Visible:  true,
		Hovered:  false,
		Selected: false,
	}

	b := GbaMenu{
		Renderer: parent.GetRenderer(),
		color:    color,
		Gba:      gba,
		parent:   parent,
		ratio:    ratio,
		Layout:   layout,
		Status:   s,
	}

	b.Resize()

	return &b
}

func (b *GbaMenu) Update(event sdl.Event) bool {

	if !b.Status.Active {
		return false
	}

	childrenDirty := false

	switch e := event.(type) {
    case *sdl.ControllerButtonEvent:

		if e.State != sdl.RELEASED {
			break
		}

        switch key := e.Button; key {
        case sdl.CONTROLLER_BUTTON_DPAD_DOWN:
			b.UpdateSelected(false)
			childrenDirty = true
        case sdl.CONTROLLER_BUTTON_DPAD_UP:
			b.UpdateSelected(true)
			childrenDirty = true
        case sdl.CONTROLLER_BUTTON_A:
			b.HandleSelected()
			childrenDirty = true
        }

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

func (b *GbaMenu) View() {
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

func (b *GbaMenu) Add(c Component) {
	b.children = append(b.children, &c)
}

func (b *GbaMenu) Resize() {
	ChildFunc(b, func(child *Component) {
		(*child).Resize()
	})
}

func (b *GbaMenu) GetChildren() []*Component {
	return b.children
}

func (b *GbaMenu) GetParent() *Component {
	return &b.parent
}

func (b *GbaMenu) GetLayout() *Layout {
	return &b.Layout
}

func (b *GbaMenu) GetStatus() Status {
	return b.Status
}

func (b *GbaMenu) SetChildren(c []*Component) {
	b.children = c
}

func (b *GbaMenu) SetStatus(s Status) {
	b.Status = s
}

func (b *GbaMenu) SetLayout(l Layout) {
	b.Layout = l
}

func (b *GbaMenu) InitOptions() {

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

func (b *GbaMenu) UpdateSelected(reverse bool) {

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

func (b *GbaMenu) HandleSelected() {

	switch b.SelectedIdx {
	case 0:
		b.Gba.TogglePause()
		b.Status.Active = false
	case 1:

		textComponent := (*(*b.children[0]).GetChildren()[1])
		switch c := textComponent.(type) {
		case *Text:
			if muted := b.Gba.ToggleMute(); muted {
				c.UpdateText("unmute")
				break
			}

			c.UpdateText("mute")
		}

	case 2:

        p := b.parent.(*Scene)

        gbaConsole.Close()

        for _, child := range p.GetChildren() {
            switch g := (*child).(type) {
            case *GbaFrame:
                g.SetStatus(Status{Active: false})
            }
        }

        if oto.OtoPlayer == nil {
            println("GOOD NIL PLAYER")
        }

        b.SetStatus(Status{Active: false})
        gbaConsole = gba.NewGBA()
        InitMainMenu(p, 0)
        return

	default:
		panic("Gb Menu Container has no children")
	}
}
func (b *GbaMenu) GetRenderer() *sdl.Renderer {
	return b.Renderer
}
