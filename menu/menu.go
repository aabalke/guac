package menu

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

const (
    WIDTH_COUNT = 6
    //WIDTH_COUNT = 4
)

type Menu struct {
    SelectedIdx int
    Data []GameData

    menuPlayer *MenuPlayer
}

func NewMenu(context *audio.Context) *Menu {
    m := &Menu{
        Data: LoadGameData(),
    }

    p, err := NewMenuPlayer(context)
    if err != nil {
        panic(err)
    }

    m.menuPlayer = p

    return m
}

func (m *Menu) InputHandler(keys []ebiten.Key, buttons []ebiten.GamepadButton) bool {

    m.menuPlayer.handleChannels()

    for _, key := range keys {
        switch key {
        case ebiten.KeyUp, ebiten.KeyW:
            if m.SelectedIdx - WIDTH_COUNT < 0 {
                m.menuPlayer.update(1)
            } else {
                m.menuPlayer.update(0)
                m.SelectedIdx = max(0, (m.SelectedIdx) - WIDTH_COUNT)
            }
        case ebiten.KeyDown, ebiten.KeyS:
            if m.SelectedIdx + WIDTH_COUNT > len(m.Data) - 1 {
                m.menuPlayer.update(1)
            } else {
                m.menuPlayer.update(0)
                m.SelectedIdx = min(len(m.Data) - 1, (m.SelectedIdx) + WIDTH_COUNT)
            }
        case ebiten.KeyRight, ebiten.KeyD:
            if m.SelectedIdx + 1 > len(m.Data) - 1 {
                m.menuPlayer.update(1)
            } else {
                m.menuPlayer.update(0)
                m.SelectedIdx = min(len(m.Data) - 1, (m.SelectedIdx) + 1)
            }
        case ebiten.KeyLeft, ebiten.KeyA:
            if m.SelectedIdx - 1 < 0 {
                m.menuPlayer.update(1)
            } else {
                m.menuPlayer.update(0)
                m.SelectedIdx = max(0, (m.SelectedIdx) - 1)
            }
        case ebiten.KeyEnter, ebiten.KeyJ:
            m.menuPlayer.update(2)
            return true
        }
    }
    
    for _, button := range buttons {
        switch button {
        case ebiten.GamepadButton2:
            m.menuPlayer.update(2)
            return true
        case ebiten.GamepadButton16:
            if m.SelectedIdx + 1 > len(m.Data) - 1 {
                m.menuPlayer.update(1)
            } else {
                m.menuPlayer.update(0)
                m.SelectedIdx = min(len(m.Data) - 1, (m.SelectedIdx) + 1)
            }
        case ebiten.GamepadButton18:
            if m.SelectedIdx - 1 < 0 {
                m.menuPlayer.update(1)
            } else {
                m.menuPlayer.update(0)
                m.SelectedIdx = max(0, (m.SelectedIdx) - 1)
            }
        case ebiten.GamepadButton15:
            if m.SelectedIdx - WIDTH_COUNT < 0 {
                m.menuPlayer.update(1)
            } else {
                m.menuPlayer.update(0)
                m.SelectedIdx = max(0, (m.SelectedIdx) - WIDTH_COUNT)
            }
        case ebiten.GamepadButton17:
            if m.SelectedIdx + WIDTH_COUNT > len(m.Data) - 1 {
                m.menuPlayer.update(1)
            } else {
                m.menuPlayer.update(0)
                m.SelectedIdx = min(len(m.Data) - 1, (m.SelectedIdx) + WIDTH_COUNT)
            }
        }
    }

    return false
}

func (m *Menu) DrawMenu(screen *ebiten.Image) {

    sw, _ := screen.Bounds().Dx(), screen.Bounds().Dy()
    elementUnit := float64(sw / WIDTH_COUNT)

    row := float64(m.SelectedIdx / WIDTH_COUNT)
    //maxRow := float64((len(m.Data) - 1) / WIDTH_COUNT)

    //var rowOffset float64
    //switch row {
    //case 0: rowOffset = 0
    //case maxRow:

    //    // not sure how to handle currently

    //    if maxRow * elementUnit > float64(screen.Bounds().Dy()) {
    //        rowOffset = (elementUnit * row) - float64(screen.Bounds().Dy())
    //    } else {
    //        rowOffset = elementUnit * row
    //    }


    //default:
    //    rowOffset = elementUnit * row
    //}

    rowOffset := elementUnit * row


    for i := range len(m.Data) {
        x := float64(i % WIDTH_COUNT) * elementUnit
        y := float64(i / WIDTH_COUNT) * elementUnit - rowOffset
        m.Image(screen, x, y, elementUnit, i)
    }
}

func (m *Menu) Image(screen *ebiten.Image, x, y, elementUnit float64, i int) {

    img := (m.Data)[i].Image

    s := (elementUnit / float64(img.Bounds().Dx()))

    opts := &ebiten.DrawImageOptions{}
    opts.GeoM.Scale(s, s)
    opts.GeoM.Translate(x, y)

    if shadeUnselected := i != m.SelectedIdx; shadeUnselected {
        opts.ColorScale.ScaleAlpha(0.5)
    }

    screen.DrawImage(img, opts)
}
