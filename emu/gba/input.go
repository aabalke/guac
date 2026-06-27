package gba

import (
	"slices"

	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2"
)

func (gba *GBA) InputHandler(keys []ebiten.Key, buttons []ebiten.StandardGamepadButton) {
	k := &gba.Keypad.Input
	*k = 0x3FF

	keyConfig := config.Conf.Gba.KeyboardConfig
	for _, key := range keys {
		switch {
		case slices.Contains(keyConfig.A, key):
			*k &^= 1 << 0
		case slices.Contains(keyConfig.B, key):
			*k &^= 1 << 1
		case slices.Contains(keyConfig.Select, key):
			*k &^= 1 << 2
		case slices.Contains(keyConfig.Start, key):
			*k &^= 1 << 3
		case slices.Contains(keyConfig.Right, key):
			*k &^= 1 << 4
		case slices.Contains(keyConfig.Left, key):
			*k &^= 1 << 5
		case slices.Contains(keyConfig.Up, key):
			*k &^= 1 << 6
		case slices.Contains(keyConfig.Down, key):
			*k &^= 1 << 7
		case slices.Contains(keyConfig.R, key):
			*k &^= 1 << 8
		case slices.Contains(keyConfig.L, key):
			*k &^= 1 << 9

		case key == ebiten.KeyH:
			B[0] = true
		}
	}

	buttonConfig := config.Conf.Gba.ControllerConfig
	for _, button := range buttons {
		switch {
		case slices.Contains(buttonConfig.A, button):
			*k &^= 1 << 0
		case slices.Contains(buttonConfig.B, button):
			*k &^= 1 << 1
		case slices.Contains(buttonConfig.Select, button):
			*k &^= 1 << 2
		case slices.Contains(buttonConfig.Start, button):
			*k &^= 1 << 3
		case slices.Contains(buttonConfig.Right, button):
			*k &^= 1 << 4
		case slices.Contains(buttonConfig.Left, button):
			*k &^= 1 << 5
		case slices.Contains(buttonConfig.Up, button):
			*k &^= 1 << 6
		case slices.Contains(buttonConfig.Down, button):
			*k &^= 1 << 7
		case slices.Contains(buttonConfig.R, button):
			*k &^= 1 << 8
		case slices.Contains(buttonConfig.L, button):
			*k &^= 1 << 9
		}
	}

	gba.Keypad.keyIRQ()
}
