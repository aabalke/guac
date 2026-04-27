package ui

import (
	"log"
	"slices"

	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) GetGamepadButtons() (justButtons, buttons []ebiten.StandardGamepadButton) {

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

	justButtons = inpututil.AppendJustPressedStandardGamepadButtons(g.gamepad, justButtons)
	buttons = inpututil.AppendPressedStandardGamepadButtons(g.gamepad, buttons)

	return justButtons, buttons
}

func (g *Game) GetInput() (justKeys, keys []ebiten.Key, justButtons, buttons []ebiten.StandardGamepadButton) {

	g.mouse.Update()

	if !ebiten.IsFocused() {
		return
	}

	justKeys = inpututil.AppendJustPressedKeys(justKeys)
	keys = inpututil.AppendPressedKeys(keys)
	justButtons, buttons = g.GetGamepadButtons()

	keyConfig := config.Conf.General.KeyboardConfig
	buttonConfig := config.Conf.General.ControllerConfig

	for _, key := range justKeys {
		switch keyStr := key.String(); {
		case slices.Contains(keyConfig.Fullscreen, keyStr):
			ebiten.SetFullscreen(!ebiten.IsFullscreen())
		case slices.Contains(keyConfig.Quit, keyStr):
			g.quit = true
		case slices.Contains(keyConfig.Pause, keyStr):
			g.TogglePause()
		case slices.Contains(keyConfig.Mute, keyStr):
			g.ToggleMute()
		}
	}

	for _, button := range justButtons {
		switch buttonStr := int(button); {
		case slices.Contains(buttonConfig.Pause, buttonStr):
			g.TogglePause()
		case slices.Contains(buttonConfig.Mute, buttonStr):
			g.ToggleMute()
		}
	}

	return justKeys, keys, justButtons, buttons
}
