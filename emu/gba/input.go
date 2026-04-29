package gba

import (
	"slices"

	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2"
)

func (gba *GBA) InputHandler(keys []ebiten.Key, buttons []ebiten.StandardGamepadButton) {

	keyConfig := config.Conf.Gba.KeyboardConfig
	buttonConfig := config.Conf.Gba.ControllerConfig

	k := &gba.Keypad.KEYINPUT

	*k = 0b11_1111_1111

	for _, key := range keys {

		keyStr := key.String()

		switch {
		case slices.Contains(keyConfig.A, keyStr):
			*k &^= 0b1
		case slices.Contains(keyConfig.B, keyStr):
			*k &^= 0b10
		case slices.Contains(keyConfig.Select, keyStr):
			*k &^= 0b100
		case slices.Contains(keyConfig.Start, keyStr):
			*k &^= 0b1000
		case slices.Contains(keyConfig.Right, keyStr):
			*k &^= 0b10000
		case slices.Contains(keyConfig.Left, keyStr):
			*k &^= 0b100000
		case slices.Contains(keyConfig.Up, keyStr):
			*k &^= 0b1000000
		case slices.Contains(keyConfig.Down, keyStr):
			*k &^= 0b10000000
		case slices.Contains(keyConfig.R, keyStr):
			*k &^= 0b100000000
		case slices.Contains(keyConfig.L, keyStr):
			*k &^= 0b1000000000
		}
	}

	for _, button := range buttons {

		switch {
		case slices.Contains(buttonConfig.A, button):
			*k &^= 0b1
		case slices.Contains(buttonConfig.B, button):
			*k &^= 0b10
		case slices.Contains(buttonConfig.Select, button):
			*k &^= 0b100
		case slices.Contains(buttonConfig.Start, button):
			*k &^= 0b1000
		case slices.Contains(buttonConfig.Right, button):
			*k &^= 0b10000
		case slices.Contains(buttonConfig.Left, button):
			*k &^= 0b100000
		case slices.Contains(buttonConfig.Up, button):
			*k &^= 0b1000000
		case slices.Contains(buttonConfig.Down, button):
			*k &^= 0b10000000
		case slices.Contains(buttonConfig.R, button):
			*k &^= 0b100000000
		case slices.Contains(buttonConfig.L, button):
			*k &^= 0b1000000000
		}
	}

	if gba.Keypad.keyIRQ() {
		gba.Irq.SetIRQ(12)
	}
}
