package sdl


import (
    "fmt"
	"math"
	"github.com/aabalke33/guac/emu/gba"
	"github.com/veandco/go-sdl2/sdl"
)

type DebugMenu struct {
	Renderer *sdl.Renderer
	texture  *sdl.Texture
	pixels   chan []byte
	parent   Component
	children []*Component
	tH, tW   int32
	Layout   Layout
	ratio    float64
	Status   Status
	GBA       *gba.GBA
    color sdl.Color
}

func NewDebugMenu(parent Component, ratio float64, layout Layout, gba *gba.GBA, color sdl.Color) *DebugMenu {

	r := parent.GetRenderer()

	pixels := make(chan []byte, 1)

	tH, tW := gba.GetSize()

	texture, _ := r.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, tW, tH)

	s := Status{
		Active:   true,
		Visible:  true,
		Hovered:  false,
		Selected: false,
	}

	b := DebugMenu{
        color:     color,
		Renderer: r,
		GBA:       gba,
		parent:   parent,
		texture:  texture,
		pixels:   pixels,
		ratio:    ratio,
		Layout:   layout,
		tH:       tH,
		tW:       tW,
		Status:   s,
	}


	b.Resize()

	return &b
}

func (b *DebugMenu) Update(event sdl.Event) bool {

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
			(*b.GBA).ToggleMute()
		}
	}

    if pause {
        (*b.GBA).TogglePause()
        //switch c := b.parent.(type) {
        //case *Scene: InitPauseMenu(b.Renderer, c, b.GBA)
        //default: panic("Parent of Gameboy Emulator Frame is not Scene")
        //}
    }

	(*b.GBA).InputHandler(event)

	ChildFuncUpdate(b, func(child *Component) bool {
		return (*child).Update(event)
	})

	return false
}

func (b *DebugMenu) View() {
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

func (b *DebugMenu) Add(c Component) {
	b.children = append(b.children, &c)
}

func (b *DebugMenu) Resize() {
	h := GetI32(b.Layout.H)
	SetI32(&b.Layout.W, math.Floor(float64(h)*b.ratio))
}

func (b *DebugMenu) GetChildren() []*Component {
	return b.children
}

func (b *DebugMenu) GetParent() *Component {
	return &b.parent
}

func (b *DebugMenu) GetLayout() *Layout {
	return &b.Layout
}

func (b *DebugMenu) GetStatus() Status {
	return b.Status
}

func (b *DebugMenu) SetChildren(c []*Component) {
	b.children = c
}

func (b *DebugMenu) SetStatus(s Status) {
	b.Status = s
}

func (b *DebugMenu) SetLayout(l Layout) {
	b.Layout = l
}
func (b *DebugMenu) GetRenderer() *sdl.Renderer {
	return b.Renderer
}

func InitDebugMenu(scene *Scene) {
    debugl := NewLayout(&scene.H, 0, 0, 0, 1)
    dMenu := NewDebugMenu(scene, 1, debugl, Gba, C_Grey)

    r := &Gba.Cpu.Reg.R
    cpsr := &Gba.Cpu.Reg.CPSR
    spsr := &Gba.Cpu.Reg.SPSR[0]

    t := uint32(0)

    dMenu.Add(NewDebugFrame(dMenu, 1, NewLayout(600, 600,0,0, 7), Gba))
    dMenu.Add(NewTextPointer(dMenu, NewLayout(0,0,0,0,8), "Registers", &t, 18, C_White, C_White, "relativeParent"))
    for i := range 12 {
        dMenu.Add(NewTextPointer(dMenu, NewLayout(0,0,0,25+(i * 25),8), fmt.Sprintf("%02d", i+1), &r[i], 18, C_White, C_White, "relativeParent"))
    }

    dMenu.Add(NewTextPointer(dMenu, NewLayout(0,0,0,25+(12 * 25),8), fmt.Sprint("SP"), &r[13], 18, C_White, C_White, "relativeParent"))
    dMenu.Add(NewTextPointer(dMenu, NewLayout(0,0,0,25+(13 * 25),8), fmt.Sprint("LR"), &r[14], 18, C_White, C_White, "relativeParent"))
    dMenu.Add(NewTextPointer(dMenu, NewLayout(0,0,0,25+(14 * 25),8), fmt.Sprint("PC"), &r[15], 18, C_White, C_White, "relativeParent"))

    dMenu.Add(NewCondPointer(dMenu, NewLayout(0,0,0,25+(15 * 25),8), fmt.Sprint("CPSR"), cpsr, 18, C_White, C_White, "relativeParent"))
    dMenu.Add(NewCondPointer(dMenu, NewLayout(0,0,0,25+(16 * 25),8), fmt.Sprint("SPSR"), spsr, 18, C_White, C_White, "relativeParent"))


    InitVRAM(Gba, dMenu)

    scene.Add(dMenu)
}

func InitVRAM(gba *gba.GBA, dMenu *DebugMenu) {

    x, y := int32(0), int32(460)

    base := int32(0x600_0000)

    count := int32(0)
    for i := base; i < base + 0x500; i += 32 {

        addr := fmt.Sprintf("%08X", i)

        v := uint32(i)

        dMenu.Add(NewAddrPointer(dMenu, NewLayout(0,0,x,y+(count * 20),8), addr, v, 12, C_White, C_White, "relativeParent"))

        count++
    }


}
