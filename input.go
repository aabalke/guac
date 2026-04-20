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

            g.StdFPS = !g.unlimitedFPS

			if g.unlimitedFPS {
				ebiten.SetTPS(UNLIMITED_FPS)
                println("setting unlimited fps")
			} else {
				ebiten.SetTPS(60)
                println("setting 60fps")
			}

		case slices.Contains(keyConfig.Pause, keyStr):
			g.TogglePause()
		case slices.Contains(keyConfig.Mute, keyStr):
			g.ToggleMute()
			//case slices.Contains([]string{"B"}, keyStr):
			//    isProfiling = true
			//    pprof.StartCPUProfile(f)

        case slices.Contains(keyConfig.Fps15, keyStr):
            ebiten.SetTPS(15)
			println("setting 15fps")
            g.StdFPS = false
        case slices.Contains(keyConfig.Fps30, keyStr):
            ebiten.SetTPS(30)
			println("setting 30fps")
            g.StdFPS = false
        case slices.Contains(keyConfig.Fps60, keyStr):
            ebiten.SetTPS(60)
			println("setting 60fps")
            g.StdFPS = true
        case slices.Contains(keyConfig.Fps120, keyStr):
            ebiten.SetTPS(120)
			println("setting 120fps")
            g.StdFPS = false
        case slices.Contains(keyConfig.Fps180, keyStr):
            ebiten.SetTPS(180)
			println("setting 180fps")
            g.StdFPS = false
        case slices.Contains(keyConfig.Fps240, keyStr):
            ebiten.SetTPS(240)
			println("setting 240fps")
            g.StdFPS = false
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

	return false
}
