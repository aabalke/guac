package sdl

import (
	"math"
	"unsafe"

	gameboy "github.com/aabalke33/guac/emu/gb"
	"github.com/veandco/go-sdl2/sdl"
)

type GbFrame struct {
	Renderer   *sdl.Renderer
	texture    *sdl.Texture
	pixels     chan []byte
	parent     Component
	children   []*Component
	W, X, Y, Z int32
	tH, tW     int32
	H          *int32
	ratio      float64
	Status     Status
	Gb         *gameboy.GameBoy
}

func NewGbFrame(Renderer *sdl.Renderer, parent Component, ratio float64, h *int32, x, y, z int32, gb *gameboy.GameBoy) *GbFrame {

	pixels := make(chan []byte, 1)

	tH, tW := gb.GetSize()

	texture, _ := Renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, tW, tH)

	s := Status{
		Active:   true,
		Visible:  true,
		Hovered:  false,
		Selected: false,
	}

	b := GbFrame{
		Renderer: Renderer,
		Gb:       gb,
		parent:   parent,
		texture:  texture,
		pixels:   pixels,
		ratio:    ratio,
		X:        x,
		Y:        y,
		H:        h,
		Z:        z,
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

	switch e := event.(type) {
	case *sdl.KeyboardEvent:
		if e.State != sdl.RELEASED {
			break
		}

		switch e.Keysym.Sym {
		case sdl.K_p:
			(*b.Gb).TogglePause()

			switch c := b.parent.(type) {
			case *Scene:
				InitPauseMenu(b.Renderer, c, b.Gb)
			default:
				panic("Parent of Gameboy Emulator Frame is not Scene")
			}

		case sdl.K_m:
			(*b.Gb).ToggleMute()
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
	w, h := win.GetSize()

	b.X = int32(math.Floor(float64(w)/2 - float64(b.W)/2))
	b.Y = int32(math.Floor(float64(h)/2 - float64(*b.H)/2))

	rect := sdl.Rect{X: b.X, Y: b.Y, W: b.W, H: *b.H}
	b.Renderer.Copy(b.texture, nil, &rect)

	ChildFunc(b, func(child *Component) {
		(*child).View()
	})
}

func (b *GbFrame) Add(c Component) {
	b.children = append(b.children, &c)
}

func (b *GbFrame) Resize() {
	b.W = int32(math.Floor(float64(*b.H) * b.ratio))
}

func (b *GbFrame) GetChildren() []*Component {
	return b.children
}

func (b *GbFrame) GetParent() *Component {
	return &b.parent
}

func (b *GbFrame) GetLayout() Layout {
	return Layout{X: b.X, Y: b.Y, H: *b.H, W: b.W, Z: b.Z}
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
	b.W = l.W
	//b.H = l.H
	b.X = l.X
	b.Y = l.Y
	b.Z = l.Z
}
