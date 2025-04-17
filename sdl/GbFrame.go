package sdl

import (
	"math"
	"unsafe"

	gameboy "github.com/aabalke33/guac/emu/gb"
	"github.com/veandco/go-sdl2/sdl"
)

type GbFrame struct {
	Renderer *sdl.Renderer
	texture  *sdl.Texture
	pixels   chan []byte
	parent   Component
	children []*Component
	tH, tW   int32
	Layout   Layout
	ratio    float64
	Status   Status
	Gb       *gameboy.GameBoy
}

func NewGbFrame(parent Component, ratio float64, layout Layout, gb *gameboy.GameBoy) *GbFrame {

	r := parent.GetRenderer()

	pixels := make(chan []byte, 1)

	tH, tW := gb.GetSize()

	texture, _ := r.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, tW, tH)

	s := Status{
		Active:   true,
		Visible:  true,
		Hovered:  false,
		Selected: false,
	}

	b := GbFrame{
		Renderer: r,
		Gb:       gb,
		parent:   parent,
		texture:  texture,
		pixels:   pixels,
		ratio:    ratio,
		Layout:   layout,
		tH:       tH,
		tW:       tW,
		Status:   s,
	}

	go b.UpdatePixels()

	b.Resize()

	return &b
}

func (b *GbFrame) UpdatePixels() {

	for {
		select {
		case b.pixels <- (*b.Gb).GetPixels():
		}
	}
}

func (b *GbFrame) Update(event sdl.Event) bool {

	//if !b.Active {
	//	return
	//}

    pause := false

	switch e := event.(type) {
    case *sdl.ControllerButtonEvent:

		if e.State != sdl.RELEASED {
			break
		}

        switch key := e.Button; key {
        case sdl.CONTROLLER_BUTTON_GUIDE:
            pause = true
        }

	case *sdl.KeyboardEvent:
		if e.State != sdl.RELEASED {
			break
		}

		switch e.Keysym.Sym {
		case sdl.K_p:
            pause = true
		case sdl.K_m:
			(*b.Gb).ToggleMute()
		}
	}

    if pause {
        (*b.Gb).TogglePause()
        switch c := b.parent.(type) {
        case *Scene: InitPauseMenu(b.Renderer, c, b.Gb)
        default: panic("Parent of Gameboy Emulator Frame is not Scene")
        }
    }

	(*b.Gb).InputHandler(event)

	ChildFuncUpdate(b, func(child *Component) bool {
		return (*child).Update(event)
	})

	return false
}

func (b *GbFrame) View() {
	//if !b.Active {
	//	return
	//}

	select {
	case pixels := <-b.pixels:
		b.texture.Update(nil, unsafe.Pointer(&pixels[0]), int(b.tW*4))
	}

	b.Renderer.Clear()
	win, _ := b.Renderer.GetWindow()
	winW, winH := win.GetSize()

	SetI32(&b.Layout.X, math.Floor(float64(winW)/2-float64(GetI32(b.Layout.W))/2))
	SetI32(&b.Layout.Y, math.Floor(float64(winH)/2-float64(GetI32(b.Layout.H))/2))

	x := GetI32(b.Layout.X)
	y := GetI32(b.Layout.Y)
	w := GetI32(b.Layout.W)
	h := GetI32(b.Layout.H)
	rect := sdl.Rect{X: x, Y: y, W: w, H: h}

	b.Renderer.Copy(b.texture, nil, &rect)

	ChildFunc(b, func(child *Component) {
		(*child).View()
	})
}

func (b *GbFrame) Add(c Component) {
	b.children = append(b.children, &c)
}

func (b *GbFrame) Resize() {
	h := GetI32(b.Layout.H)
	SetI32(&b.Layout.W, math.Floor(float64(h)*b.ratio))
}

func (b *GbFrame) GetChildren() []*Component {
	return b.children
}

func (b *GbFrame) GetParent() *Component {
	return &b.parent
}

func (b *GbFrame) GetLayout() *Layout {
	return &b.Layout
}

func (b *GbFrame) GetStatus() Status {
	return b.Status
}

func (b *GbFrame) SetChildren(c []*Component) {
	b.children = c
}

func (b *GbFrame) SetStatus(s Status) {
	b.Status = s
}

func (b *GbFrame) SetLayout(l Layout) {
	b.Layout = l
}
func (b *GbFrame) GetRenderer() *sdl.Renderer {
	return b.Renderer
}
