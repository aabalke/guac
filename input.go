package main

import (
	"slices"

	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2"
)

func (g *Game) inputHandler(justKeys []ebiten.Key, justButtons []ebiten.StandardGamepadButton) bool {

	keyConfig := config.Conf.KeyboardConfig
	buttonConfig := config.Conf.ControllerConfig

	for _, key := range justKeys {

		keyStr := key.String()

		switch {
		case slices.Contains(keyConfig.Fullscreen, keyStr):
			ebiten.SetFullscreen(!ebiten.IsFullscreen())
		case slices.Contains(keyConfig.Quit, keyStr):
			return true
		case slices.Contains(keyConfig.Unlimited, keyStr):

			g.unlimitedFPS = !g.unlimitedFPS

			if g.unlimitedFPS {
				ebiten.SetTPS(UNLIMITED_FPS)
			} else {
				ebiten.SetTPS(60)
			}

		case slices.Contains(keyConfig.Pause, keyStr):
			g.TogglePause()
		case slices.Contains(keyConfig.Mute, keyStr):
			g.ToggleMute()
			//case slices.Contains([]string{"B"}, keyStr):
			//    isProfiling = true
			//    pprof.StartCPUProfile(f)

		case slices.Contains([]string{"Numpad4"}, keyStr):
			ebiten.SetTPS(15)
			println("setting 15fps")
		case slices.Contains([]string{"Numpad5"}, keyStr):
			ebiten.SetTPS(30)
			println("setting 30fps")
		case slices.Contains([]string{"Numpad1"}, keyStr):
			ebiten.SetTPS(60)
			println("setting 60fps")
		case slices.Contains([]string{"Numpad2"}, keyStr):
			ebiten.SetTPS(120)
			println("setting 120fps")
		case slices.Contains([]string{"Numpad3"}, keyStr):
			ebiten.SetTPS(180)
			println("setting 180fps")
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
	}

	return false
}
