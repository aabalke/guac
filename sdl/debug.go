package sdl

import (
	"math"
	"unsafe"

	comp "github.com/aabalke33/go-sdl2-components/Components"
	//"github.com/aabalke33/guac/emu"
	gameboy "github.com/aabalke33/guac/emu/gb"
	"github.com/veandco/go-sdl2/sdl"
)

type DebugFrame struct {
	renderer   *sdl.Renderer
	texture    *sdl.Texture
	pixels     chan []byte
	parent     comp.Component
	children   []*comp.Component
	w, x, y, z int32
	tH, tW     int32
	h          *int32
	ratio      float64
	Active     bool
    Gameboy   *gameboy.GameBoy
}

func NewDebugFrame(renderer *sdl.Renderer, parent comp.Component, ratio float64, h *int32, x, y, z int32, gb *gameboy.GameBoy) *DebugFrame {

    pixels := make(chan []byte, 1)

    var tW int32 = 8
    var tH int32 = 8
	texture, _ := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, tW, tH)


	b := DebugFrame{
		renderer: renderer,
        Gameboy:  gb,
		parent:   parent,
		texture:  texture,
		pixels:   pixels,
		ratio:    ratio,
		x:        x,
		y:        y,
		h:        h,
		z:        z,
		tH:       tH,
		tW:       tW,
		Active:   true,
	}

    go b.UpdatePixels()

	b.Resize()

	return &b
}

func (b *DebugFrame) UpdatePixels() {

    for {
        select {
        case b.pixels <- b.Gameboy.GetBgTiles():
        }
    }
}

func (b *DebugFrame) Update(dt float64, event sdl.Event) {

	if !b.Active {
		return
	}

    //(*b.Emulator).InputHandler(event)

    comp.ChildFunc(b, func(child *comp.Component) {
        (*child).Update(1/comp.FPS, event)
    })
}

func (b *DebugFrame) View(renderer *sdl.Renderer) {
	if !b.Active {
		return
	}

    select {
    case pixels := <-b.pixels:
        b.texture.Update(nil, unsafe.Pointer(&pixels[0]), int(b.tW*4))
    }

	b.renderer.Clear()
	win, _ := b.renderer.GetWindow()
	w, h := win.GetSize()

	b.x = int32(math.Floor(float64(w)/2 - float64(b.w)/2))
	b.y = int32(math.Floor(float64(h)/2 - float64(*b.h)/2))

	rect := sdl.Rect{X: b.x, Y: b.y, W: b.w, H: *b.h}
	b.renderer.Copy(b.texture, nil, &rect)

	comp.ChildFunc(b, func(child *comp.Component) {
		(*child).View(renderer)
	})
}

func (b *DebugFrame) Add(c comp.Component) {
	b.children = append(b.children, &c)
}

func (b *DebugFrame) GetZ() int32 {
	return b.z
}

func (b *DebugFrame) Resize() {
	b.w = int32(math.Floor(float64(*b.h) * b.ratio))
}

func (b *DebugFrame) IsActive() bool {
	return b.Active
}

func (b *DebugFrame) GetChildren() []*comp.Component {
	return b.children
}

func (b *DebugFrame) GetParent() *comp.Component {
	return &b.parent
}

func (b *DebugFrame) GetSize() (int32, int32) {
	return *b.h, b.w
}

func (b *DebugFrame) SetChildren(c []*comp.Component) {
    b.children = c
}
