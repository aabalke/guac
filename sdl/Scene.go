package sdl

import (
	"github.com/veandco/go-sdl2/sdl"
)

type Scene struct {
	Renderer *sdl.Renderer
	children []*Component
	parent   *Component
	Status   Status
	W, H     int32
	MAX_Z    int32
	color    sdl.Color
}

func NewScene(renderer *sdl.Renderer, w, h, MAX_Z int32, color sdl.Color) *Scene {

	s := Status{
		Active:   true,
		Visible:  true,
		Hovered:  false,
		Selected: false,
	}

	return &Scene{Renderer: renderer, W: w, H: h, MAX_Z: MAX_Z, Status: s, color: color}
}

func (s *Scene) Add(c Component) {
	s.children = append(s.children, &c)
}

func (s *Scene) Update(_ sdl.Event) bool {

	if !s.Status.Active {
		return false
	}

	resized := false
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		case *sdl.QuitEvent:
			s.Status.Active = false
			return false

		case *sdl.WindowEvent:
			//if e.GetType() == sdl.WINDOWEVENT_RESIZED ||
			//    e.GetType() == sdl.WINDOWEVENT_SIZE_CHANGED {
			//        resized = true
			//}

			resized = true

		case *sdl.KeyboardEvent:
			switch e.Keysym.Sym {
			case sdl.K_F11:
                if e.State == sdl.RELEASED {
                    s.toggleFullscreen()
                    resized = true
                    continue
                }
			case sdl.K_q:
                if e.State == sdl.RELEASED {
                    s.Status.Active = false
                }
			}
		}

		ChildFuncUpdate(s, func(child *Component) bool {
			return (*child).Update(event)
		})
	}

	if resized {
		s.Resize()
	}

    return false
}

func (s *Scene) View() {

	if !s.Status.Active {
		return
	}

    s.Renderer.Clear()

	s.Renderer.SetDrawColor(s.color.R, s.color.G, s.color.B, s.color.A)
	rect := sdl.Rect{X: 0, Y: 0, W: s.W, H: s.H}
	s.Renderer.FillRect(&rect)

	ChildFunc(s, func(child *Component) {
		(*child).View()
	})
}

func (s *Scene) toggleFullscreen() {

	win, _ := s.Renderer.GetWindow()
	isFullScreen := win.GetFlags()&sdl.WINDOW_FULLSCREEN_DESKTOP == sdl.WINDOW_FULLSCREEN_DESKTOP
	m, _ := sdl.GetCurrentDisplayMode(0)

	if !isFullScreen {
		win.SetFullscreen(sdl.WINDOW_FULLSCREEN_DESKTOP)
		win.SetSize(m.W, m.H)
		return
	}

	win.SetFullscreen(0)
	win.SetSize(m.W/2, m.H/2)
	win.SetPosition(m.W*1/4, m.H*1/4)
}

func (s *Scene) Resize() {

	if !s.Status.Active {
		return
	}

	win, _ := s.Renderer.GetWindow()
	w, h := win.GetSize()
	s.W, s.H = w, h

	ChildFunc(s, func(child *Component) {
		(*child).Resize()
	})
}

func (s *Scene) GetChildren() []*Component {
	return s.children
}

func (s *Scene) GetParent() *Component {
	return s.parent
}

func (b *Scene) GetLayout() Layout {
	return Layout{X: 0, Y: 0, H: b.H, W: b.W, Z: 0}
}

func (b *Scene) GetStatus() Status {
	return b.Status
}

func (b *Scene) SetChildren(c []*Component) {
	b.children = c
}

func (b *Scene) DeleteInactive() {

	var rec func(Component)

	rec = func(parent Component) {
		ChildFunc(parent, func(child *Component) {
			if status := (*child).GetStatus(); !status.Active {
				RemoveChild(child)
				return
			}

			rec(*child)
		})
	}

	rec(b)
}

func (b *Scene) SetStatus(s Status) {
	b.Status = s
}

func (b *Scene) SetLayout(l Layout) {
}
