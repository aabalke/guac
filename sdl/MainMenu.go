package sdl

import (
	"fmt"
	"math"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

type MainMenu struct {
	Renderer    *sdl.Renderer
	parent      Component
	children    []*Component
	Layout      Layout
	ratio       float64
	Status      Status
	color       sdl.Color
	SelectedIdx int
	GameDatas   []GameData
}

func NewMainMenu(parent Component, layout Layout, color sdl.Color, duration time.Duration, gameDatas []GameData) *MainMenu {

	ratio := 1.0

	s := Status{
		Active:   true,
		Visible:  true,
		Hovered:  false,
		Selected: false,
	}

	b := MainMenu{
		Renderer:  parent.GetRenderer(),
		color:     color,
		parent:    parent,
		ratio:     ratio,
		Layout:    layout,
		Status:    s,
		GameDatas: gameDatas,
	}

	//timer := time.NewTimer(duration)

	//go func() {
	//	<-timer.C
	//	//b.Status.Active = false

	//}()

	b.Resize()

	return &b
}

func (b *MainMenu) Update(event sdl.Event) bool {

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

func (b *MainMenu) View() {
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

func (b *MainMenu) Add(c Component) {
	b.children = append(b.children, &c)
}

func (b *MainMenu) Resize() {
	//b.W = int32(math.Floor(float64(*b.H) * b.ratio))

	if len(b.GetChildren()) == 0 {
		return
	}

	container := *(b.GetChildren()[0])

	containerY := (GetI32(b.GetLayout().H) / 2) - (200 / 2) - int32(250*b.SelectedIdx)
	SetI32(&container.GetLayout().Y, containerY)

	ChildFunc(b, func(child *Component) {
		(*child).Resize()
	})
}

func (b *MainMenu) GetChildren() []*Component {
	return b.children
}

func (b *MainMenu) GetParent() *Component {
	return &b.parent
}

func (b *MainMenu) GetLayout() *Layout {
	return &b.Layout
}

func (b *MainMenu) GetStatus() Status {
	return b.Status
}

func (b *MainMenu) SetChildren(c []*Component) {
	b.children = c
}

func (b *MainMenu) SetStatus(s Status) {
	b.Status = s
}

func (b *MainMenu) SetLayout(l Layout) {
	b.Layout = l
}

func (b *MainMenu) InitOptions() {

	container := *(b.GetChildren()[0])

	containerY := (GetI32(b.GetLayout().H) / 2) - (200 / 2) - int32(250*b.SelectedIdx)
	SetI32(&container.GetLayout().Y, containerY)

	{
		gameData := container.GetChildren()
		children := (*gameData[0]).GetChildren()
		filterComponent := (children[len(children)-1])

		oStatus := (*filterComponent).GetStatus()
		nStatus := Status{
			Active:   oStatus.Active,
			Visible:  oStatus.Visible,
			Hovered:  oStatus.Hovered,
			Selected: true,
		}

		(*filterComponent).SetStatus(nStatus)
	}
}

func (b *MainMenu) UpdateSelected(reverse bool) {

	container := *(b.GetChildren()[0])
	gameData := container.GetChildren()

	originalIdx := b.SelectedIdx
	var newIdx int

	switch {
	case reverse && originalIdx == 0:
		newIdx = 0
	case !reverse && originalIdx == len(gameData)-1:
		newIdx = originalIdx
	case reverse:
		newIdx = originalIdx - 1
	case !reverse:
		newIdx = originalIdx + 1
	}

	{ // move container vertical
		l := container.GetLayout()

		switch {
		case newIdx > originalIdx:
			SetI32(&l.Y, GetI32(l.Y)-250)
		case newIdx < originalIdx:
			SetI32(&l.Y, GetI32(l.Y)+250)
		}
	}

	{ // change selection color
		children := (*gameData[originalIdx]).GetChildren()
		filterComponent := (children[len(children)-1])

		oStatus := (*filterComponent).GetStatus()
		nStatus := Status{
			Active:   oStatus.Active,
			Visible:  oStatus.Visible,
			Hovered:  oStatus.Hovered,
			Selected: false,
		}
		(*filterComponent).SetStatus(nStatus)

		children = (*gameData[newIdx]).GetChildren()
		filterComponent = (children[len(children)-1])

		oStatus = (*filterComponent).GetStatus()
		nStatus = Status{
			Active:   oStatus.Active,
			Visible:  oStatus.Visible,
			Hovered:  oStatus.Hovered,
			Selected: true,
		}
		(*filterComponent).SetStatus(nStatus)
	}

	b.SelectedIdx = newIdx
}

func (b *MainMenu) HandleSelected() {

	switch scene := b.parent.(type) {
	case *Scene:
		path := b.GameDatas[b.SelectedIdx].RomPath

		Gb.LoadGame(path)

		l := NewLayout(&scene.H, 0, 0, 0, 1)
		(b.parent).Add(NewGbFrame(scene, 160.0/144, l, Gb))
		Gb.Paused = false

		b.Status.Active = false

        b.GameDatas = ReorderGameData(&b.GameDatas, b.SelectedIdx)

        WriteGameData(&b.GameDatas)

		return
	}

	panic("Parent of Main Menu is not Scene")
}

func InitMainMenu(scene *Scene, duration time.Duration) {

    gameDatas := LoadGameData()

	l := NewLayout(&scene.H, &scene.W, 0, 0, 2)
	menu := NewMainMenu(scene, l, C_White, duration, *gameDatas)

	l = NewLayout(len(*gameDatas)*250, 800, 0, 0, 3)
	container := NewContainer(menu, l, C_Transparent, "centerHorizontal")

	for i, v := range *gameDatas {
		InitGameData(container, duration, v, i, 10, i == menu.SelectedIdx)
	}

	menu.Add(container)
	menu.InitOptions()
	scene.Add(menu)
}

func InitGameData(parent Component, duration time.Duration, data GameData, i, z int, selected bool) {

	y := int32(i * 250)

	H := 200
	W := 800

	l := NewLayout(H, W, 0, y, z)
	container := NewContainer(parent, l, C_Brown, "relativeParent")
	l = NewLayout(H, 200, 0, 0, z+1)
	container.Add(NewImage(container, l, data.ArtPath, "relativeParent"))

	{
		l = NewLayout(H, 600, 200, 0, z+2)
		nested := NewContainer(container, l, C_Brown, "relativeParent")

		l = NewLayout(0, 0, 25, 50, z+3)
		nested.Add(NewText(nested, l, data.Name, 48, C_White, C_White, "relativeParent"))

		t := fmt.Sprintf("%s, %d", data.Console, data.Year)
		l = NewLayout(0, 0, 25, 100, z+3)
		nested.Add(NewText(nested, l, t, 24, C_White, C_White, "relativeParent"))

		container.Add(nested)
	}

	l = NewLayout(H, W, 0, 0, z+10)
	container.Add(NewSelectableContainer(container, l, C_Brown50, C_Transparent, "relativeParent"))

	parent.Add(container)
}

func (b *MainMenu) GetRenderer() *sdl.Renderer {
	return b.Renderer
}

