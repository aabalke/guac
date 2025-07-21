package main

import (
	"bytes"
	_ "embed"
	"image/color"

	"github.com/aabalke33/guac/menu"
	"github.com/hajimehoshi/ebiten/v2"

	"image"
	_ "image/png"
)

const (
    ICON_COUNT = 3

    IDX_PAUSE = 0
    IDX_VOLUME = 1
    IDX_EXIT = 2
)

//go:embed icons/resume.png
var IconResume []byte

//go:embed icons/volume_on.png
var IconVolumeOn []byte

//go:embed icons/volume_off.png
var IconVolumeOff []byte

//go:embed icons/close.png
var IconClose []byte

type Pause struct {
    overlay *ebiten.Image
    SelectedIdx int

    Icons []*ebiten.Image

    muted bool
}

func NewPause() *Pause {

    p := &Pause{
        overlay: ebiten.NewImage(1, 1),
    }

    p.overlay.Fill(color.Black)

    p.LoadIcons()

    return p
}

func (p *Pause) LoadIcons() {

    for _, f := range [][]byte{
        IconResume, IconVolumeOn, IconVolumeOff, IconClose,
    } {

        img, _, err := image.Decode(bytes.NewReader(f))
        if err != nil {
            panic(err)
        }

        icon := ebiten.NewImageFromImage(img)
        p.Icons = append(p.Icons, icon)
    }
}

func (p *Pause) InputHandler(g *Game, keys []ebiten.Key, buttons []ebiten.GamepadButton) {

    for _, key := range keys {
        switch key {
        case ebiten.KeyRight, ebiten.KeyD:
            p.SelectedIdx = min(ICON_COUNT - 1, (p.SelectedIdx) + 1)
        case ebiten.KeyLeft, ebiten.KeyA:
            p.SelectedIdx = max(0, (p.SelectedIdx) - 1)
        case ebiten.KeyEnter, ebiten.KeyJ:
            p.handleSelection(g)
        }
    }

    for _, button := range buttons {
        switch button {
        case ebiten.GamepadButton2:
            p.handleSelection(g)
        case ebiten.GamepadButton16:
            p.SelectedIdx = min(ICON_COUNT - 1, (p.SelectedIdx) + 1)
        case ebiten.GamepadButton18:
            p.SelectedIdx = max(0, (p.SelectedIdx) - 1)
        }
    }
}

func (p *Pause) handleSelection(g *Game) {

    switch p.SelectedIdx {
    case IDX_PAUSE:
        g.TogglePause()
    case IDX_VOLUME:
        g.ToggleMute()
        p.muted = !p.muted
    case IDX_EXIT:

        g.flags.Type = NONE

        g.gba.Close()

        g.gb = nil
        g.gba = nil

        g.menu = menu.NewMenu(g.menuCtx)

        g.paused = false
        g.pause = NewPause()
    }
}

func (p *Pause) DrawPause(screen *ebiten.Image) {

    opts := &ebiten.DrawImageOptions{}
    opts.GeoM.Scale(float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy()))
    opts.ColorScale.ScaleAlpha(0.75)
    screen.DrawImage(p.overlay, opts)

    p.drawIcons(screen)
}

func (p *Pause) drawIcons(screen *ebiten.Image) {

    var iconsW = float64(screen.Bounds().Dx()) * 0.75
    var iconsH = float64(screen.Bounds().Dy()) * 0.25
    var iconsX = float64(screen.Bounds().Dx() / 2) - (iconsW / 2)
    var iconsY = float64(screen.Bounds().Dy() / 2) - (iconsH / 2)

    p.drawIcon(screen, p.Icons[0], 0, iconsX, iconsY, iconsH)

    if p.muted {
        p.drawIcon(screen, p.Icons[2], 1, iconsX + (iconsW / 2) - (iconsH / 2), iconsY, iconsH)
    } else {
        p.drawIcon(screen, p.Icons[1], 1, iconsX + (iconsW / 2) - (iconsH / 2), iconsY, iconsH)
    }
    p.drawIcon(screen, p.Icons[3], 2, iconsX + iconsW - iconsH, iconsY, iconsH)
}

func (p *Pause) drawIcon(screen *ebiten.Image, icon *ebiten.Image, i int, x, y, size float64) {

    s := (size / float64(icon.Bounds().Dx()))

    opts := &ebiten.DrawImageOptions{}
    opts.GeoM.Scale(s, s)
    opts.GeoM.Translate(x, y)

    if shadeUnselected := i != p.SelectedIdx; shadeUnselected {
        opts.ColorScale.ScaleAlpha(0.5)
    }

    screen.DrawImage(icon, opts)
}
