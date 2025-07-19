package main

import (
	//"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	//"github.com/hajimehoshi/ebiten/v2/text"
)

type Pause struct {
    overlay *ebiten.Image
}


func (p *Pause) InputHandler(g *Game, keys []ebiten.Key, buttons []ebiten.GamepadButton) {

    //for _, key := range keys {
    //    switch key {
    //    case ebiten.KeyUp:
    //        p.SelectedIdx = p.x(0, (p.SelectedIdx) - WIDTH_COUNT)
    //    case ebiten.KeyDown:
    //        p.SelectedIdx = p.n(len(p.data) - 1, (p.SelectedIdx) + WIDTH_COUNT)
    //    case ebiten.KeyEnter:
    //        p.SelectConsole(g)
    //        p = nil
    //    }
    //}
}

func (p *Pause) DrawPause(screen *ebiten.Image) {
    //screen.Fill(color.RGBA{R: 255, G: 0, B: 0, A: 128})

    screen.DrawImage(p.overlay, nil)

    opts := &ebiten.DrawImageOptions{}
    opts.GeoM.Scale(float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy()))
    opts.ColorScale.ScaleAlpha(0.75)

    screen.DrawImage(p.overlay, opts)

    //text.Draw(screen, "Paused", g.fontFace, 100, 100, color.White)

}
