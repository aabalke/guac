package main

import (
	_ "embed"
	"errors"
	"image/color"
	"log"

	gameboy "github.com/aabalke33/guac/emu/gb"
	"github.com/aabalke33/guac/emu/gba"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/oto"

	"github.com/aabalke33/guac/menu"
)

var (
    darkGrey = color.RGBA{R: 10, G: 10, B: 10, A: 255}
    exit = errors.New("Exit")
)

type Game struct {
    flags Flags
    gba *gba.GBA
    gb *gameboy.GameBoy
    menu *menu.Menu
    pause *Pause
    frame uint64

    paused bool

    gamepad ebiten.GamepadID
    gamepadConnected bool

    menuCtx *audio.Context

    emuCtx *oto.Context
}

func NewGame(flags Flags) *Game {

    g := &Game{
        flags: flags,
        menuCtx: audio.NewContext(SND_FREQUENCY),
        emuCtx: NewAudioContext(),
    }

    switch g.flags.Type {
    case NONE:

        //ebiten.SetFullscreen(true)

        g.menu = menu.NewMenu(g.menuCtx)
        g.pause = NewPause()

    case GBA:
        g.gba = gba.NewGBA(flags.RomPath, g.emuCtx)
    case GB:
        g.gb = gameboy.NewGameBoy(flags.RomPath, g.emuCtx)
    }

    return g
}

func (g *Game) GetGamepadButtons() ([]ebiten.GamepadButton, []ebiten.GamepadButton) {

    gamepads := inpututil.AppendJustConnectedGamepadIDs([]ebiten.GamepadID{})

    if len(gamepads) > 0 && !g.gamepadConnected {
        log.Printf("Gamepad has been connected\n")
        g.gamepad = gamepads[0]
        g.gamepadConnected = true
    }

    if inpututil.IsGamepadJustDisconnected(g.gamepad) && g.gamepadConnected {
        log.Printf("Gamepad has been disconnected\n")
        g.gamepadConnected = false
    }

    justButtons := inpututil.AppendJustPressedGamepadButtons(g.gamepad, []ebiten.GamepadButton{})
    buttons := inpututil.AppendPressedGamepadButtons(g.gamepad, []ebiten.GamepadButton{})

    return justButtons, buttons
}


func (g *Game) Update() error {

    if g.flags.Profile && g.frame >= 1000 {
        return exit
    }

    justKeys := inpututil.AppendJustPressedKeys([]ebiten.Key{})
    keys := inpututil.AppendPressedKeys([]ebiten.Key{})
    justButtons, buttons := g.GetGamepadButtons()

    for _, key := range justKeys {
        switch key {
        case ebiten.KeyF11:
            ebiten.SetFullscreen(!ebiten.IsFullscreen())
        case ebiten.KeyQ:
            return exit
        case ebiten.KeyP:
            g.TogglePause()
        case ebiten.KeyM:
            g.ToggleMute()
        }
    }

    for _, button := range justButtons {
        switch button {
        case ebiten.GamepadButton9:
            g.TogglePause()

        case ebiten.GamepadButton8:
            g.ToggleMute()
        }
    }

    if g.paused {
        g.pause.InputHandler(g, justKeys, justButtons)
        return nil
    } else {

        switch g.flags.Type {
        case NONE:
            selected := g.menu.InputHandler(justKeys, justButtons)
            if selected {
                g.SelectConsole()
                g.menu = nil
            }
        case GBA:
            g.gba.InputHandler(keys, buttons)
            g.gba.Update()
            g.gba.Image.WritePixels(g.gba.Pixels)
        case GB:
            g.gb.InputHandler(keys, buttons)
            g.gb.Update()
            g.gb.Image.WritePixels(*g.gb.Pixels)
        }
    }

    g.frame++

    return nil
}

func (g *Game) SelectConsole() {

    m := g.menu

    rom := m.Data[m.SelectedIdx]

    switch rom.Type {
    case GBA:
        g.gba = gba.NewGBA(rom.RomPath, g.emuCtx)
        g.flags.Type = GBA
    case GB:
        g.gb = gameboy.NewGameBoy(rom.RomPath, g.emuCtx)
        g.flags.Type = GB
    default:
        panic("Selected Unknown Console")
    }

    m.Data = menu.ReorderGameData(&m.Data, m.SelectedIdx)
    menu.WriteGameData(&m.Data)
}

func (g *Game) TogglePause() {

    if !(g.flags.Type == NONE) && g.flags.ConsoleMode {
        g.paused = !g.paused
    }

    switch g.flags.Type {
        case GBA: g.gba.TogglePause()
        case GB:  g.gb.TogglePause()
    }
}

func (g *Game) ToggleMute() {
    switch g.flags.Type {
        case GBA: g.gba.ToggleMute()
        case GB:  g.gb.ToggleMute()
    }
}

func (g *Game) Draw(screen *ebiten.Image) {

    screen.Fill(darkGrey)

    switch g.flags.Type {
    case NONE:
        g.menu.DrawMenu(screen)
        return
    case GBA:
        ImageFillScreen(screen, g.gba.Image)
    case GB:
        ImageFillScreen(screen, g.gb.Image)
    }

    if g.paused {
        g.pause.DrawPause(screen)
    }
}

func ImageFillScreen(screen *ebiten.Image, image *ebiten.Image) {

    sw, sh := float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy())
    iw, ih := float64(image.Bounds().Dx()), float64(image.Bounds().Dy())

    scaleX := sw / iw
    scaleY := sh / ih
    scale := min(scaleX, scaleY)

    offsetX := (sw - (iw * scale)) / 2
    offsetY := (sh - (ih * scale)) / 2

    op := &ebiten.DrawImageOptions{}
    op.GeoM.Scale(scale, scale)
    op.GeoM.Translate(offsetX, offsetY)
    screen.DrawImage(image, op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
    return outsideWidth, outsideHeight
}
