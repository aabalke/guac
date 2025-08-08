package main

import (
	_ "embed"
	"errors"
	"log"
	"slices"

	"github.com/aabalke/guac/config"
	gameboy "github.com/aabalke/guac/emu/gb"
	"github.com/aabalke/guac/emu/gba"
	"github.com/aabalke/guac/menu"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/oto"
)

var (
	exit = errors.New("Exit")
)

type Game struct {
	flags Flags
	gba   *gba.GBA
	gb    *gameboy.GameBoy
	menu  *menu.Menu
	pause *Pause
	frame uint64

	paused        bool
	pauseEndFrame uint64

	gamepad          ebiten.GamepadID
	gamepadConnected bool

	menuCtx *audio.Context
	emuCtx  *oto.Context
}

func NewGame(flags Flags) *Game {

	g := &Game{
		flags:  flags,
		emuCtx: NewAudioContext(),
	}

	if !config.Conf.CancelAudioInit {
		g.menuCtx = audio.NewContext(SND_FREQUENCY)
	}

	switch g.flags.Type {
	case NONE:

		g.menu = menu.NewMenu(g.menuCtx)
		g.pause = NewPause()

	case GBA:
		g.gba = gba.NewGBA(flags.RomPath, g.emuCtx)
	case GB:
		g.gb = gameboy.NewGameBoy(flags.RomPath, g.emuCtx)
	}

	return g
}

func (g *Game) GetGamepadButtons() ([]ebiten.StandardGamepadButton, []ebiten.StandardGamepadButton) {

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

	justButtons := inpututil.AppendJustPressedStandardGamepadButtons(g.gamepad, []ebiten.StandardGamepadButton{})
	buttons := inpututil.AppendPressedStandardGamepadButtons(g.gamepad, []ebiten.StandardGamepadButton{})

	return justButtons, buttons
}

func (g *Game) Update() error {

	//if g.flags.Profile && g.frame >= 1000 {
	if g.flags.Profile && g.frame >= 5000 {
		return exit
	}

	g.frame++

	justKeys := inpututil.AppendJustPressedKeys([]ebiten.Key{})
	keys := inpututil.AppendPressedKeys([]ebiten.Key{})
	justButtons, buttons := g.GetGamepadButtons()

	keyConfig := config.Conf.KeyboardConfig
	buttonConfig := config.Conf.ControllerConfig

	for _, key := range justKeys {

		keyStr := key.String()

		switch {
		case slices.Contains(keyConfig.Fullscreen, keyStr):
			ebiten.SetFullscreen(!ebiten.IsFullscreen())
		case slices.Contains(keyConfig.Quit, keyStr):
			return exit
		case slices.Contains(keyConfig.Pause, keyStr):
			g.TogglePause()
		case slices.Contains(keyConfig.Mute, keyStr):
			g.ToggleMute()
		}
	}

	for _, button := range justButtons {

		buttonStr := int(button)

		switch {
		case slices.Contains(buttonConfig.Pause, buttonStr):
			g.TogglePause()

		case slices.Contains(buttonConfig.Mute, buttonStr):
			g.ToggleMute()
		}
	}

	if g.paused {
		g.pause.InputHandler(g, justKeys, justButtons)
		return nil
	}

	if g.frame-g.pauseEndFrame < 10 {
		// pressing select on pause can sometimes input into emulator,
		// this gives time from the pause and emulator starting again
		return nil
	}

	switch g.flags.Type {
	case NONE:
		selected := g.menu.InputHandler(justKeys, justButtons, g.frame)
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
	case GBA:
		g.gba.TogglePause()
	case GB:
		g.gb.TogglePause()
	}
}

func (g *Game) ToggleMute() {
	if !(g.flags.Type == NONE) && g.flags.ConsoleMode {
		g.pause.muted = !g.pause.muted
	}

	switch g.flags.Type {
	case GBA:
		g.gba.ToggleMute()
	case GB:
		g.gb.ToggleMute()
	}
}

func (g *Game) Draw(screen *ebiten.Image) {

	screen.Fill(config.Conf.Backdrop)

	switch g.flags.Type {
	case NONE:
		g.menu.DrawMenu(screen, g.frame)
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
