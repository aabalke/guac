package main

import (
	gameboy "github.com/aabalke33/guac/emu/gb"
	"github.com/aabalke33/guac/emu/gba"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
    //WIDTH_COUNT = 8
    WIDTH_COUNT = 6
)

type Menu struct {
    SelectedIdx int
    data []GameData
}

func (m *Menu) InputHandler(g *Game, keys []ebiten.Key, buttons []ebiten.GamepadButton) {

    for _, key := range keys {
        switch key {
        case ebiten.KeyUp:
            m.SelectedIdx = max(0, (m.SelectedIdx) - WIDTH_COUNT)
        case ebiten.KeyDown:
            m.SelectedIdx = min(len(m.data) - 1, (m.SelectedIdx) + WIDTH_COUNT)
        case ebiten.KeyRight:
            m.SelectedIdx = min(len(m.data) - 1, (m.SelectedIdx) + 1)
        case ebiten.KeyLeft:
            m.SelectedIdx = max(0, (m.SelectedIdx) - 1)
        case ebiten.KeyEnter:
            m.SelectConsole(g)
            m = nil
        }
    }
    
    for _, button := range buttons {
        switch button {
        case ebiten.GamepadButton2:
            m.SelectConsole(g)
            m = nil
        case ebiten.GamepadButton16:
            m.SelectedIdx = min(len(m.data) - 1, (m.SelectedIdx) + 1)
        case ebiten.GamepadButton18:
            m.SelectedIdx = max(0, (m.SelectedIdx) - 1)
        case ebiten.GamepadButton15:
            m.SelectedIdx = max(0, (m.SelectedIdx) - WIDTH_COUNT)
        case ebiten.GamepadButton17:
            m.SelectedIdx = min(len(m.data) - 1, (m.SelectedIdx) + WIDTH_COUNT)
        }
    }
}

func (m *Menu) SelectConsole(g *Game) {

    rom := m.data[m.SelectedIdx]

    switch rom.Type {
    case GBA:
        g.gba = gba.NewGBA(rom.RomPath)
        g.flags.Type = GBA
    case GB:
        g.gb = gameboy.NewGameBoy(rom.RomPath)
        g.flags.Type = GB
    default:
        panic("Selected Unknown Console")
    }
}

func (m *Menu) DrawMenu(screen *ebiten.Image) {

    sw, _ := screen.Bounds().Dx(), screen.Bounds().Dy()
    elementUnit := float64(sw / WIDTH_COUNT)

    for i := range len(m.data) {
        x := float64(i % WIDTH_COUNT) * elementUnit
        y := float64(i / WIDTH_COUNT) * elementUnit
        m.Image(screen, x, y, elementUnit, i)
    }
}

func (m *Menu) Image(screen *ebiten.Image, x, y, elementUnit float64, i int) {

    img := (m.data)[i].Image

    s := (elementUnit / float64(img.Bounds().Dx()))

    opts := &ebiten.DrawImageOptions{}
    opts.GeoM.Scale(s, s)
    opts.GeoM.Translate(x, y)

    if shadeUnselected := i != m.SelectedIdx; shadeUnselected {
        opts.ColorScale.ScaleAlpha(0.5)
    }

    screen.DrawImage(img, opts)
}
