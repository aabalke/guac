package gameboy

import (
	"slices"

	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2"
)

func (gb *GameBoy) InputHandler(keys []ebiten.Key, buttons []ebiten.StandardGamepadButton) {

	keyConfig := config.Conf.Gb.KeyboardConfig
	buttonConfig := config.Conf.Gba.ControllerConfig

	tempJoypad := &gb.Joypad

	reqInterrupt := false

	*tempJoypad = 0b1111_1111

	for _, key := range keys {

		keyStr := key.String()
		switch {
		case slices.Contains(keyConfig.A, keyStr):
			*tempJoypad &^= 0b10000
			reqInterrupt = true
		case slices.Contains(keyConfig.B, keyStr):
			*tempJoypad &^= 0b100000
			reqInterrupt = true
		case slices.Contains(keyConfig.Select, keyStr):
			*tempJoypad &^= 0b1000000
			reqInterrupt = true
		case slices.Contains(keyConfig.Start, keyStr):
			*tempJoypad &^= 0b10000000
			reqInterrupt = true
		case slices.Contains(keyConfig.Right, keyStr):
			*tempJoypad &^= 0b1
			reqInterrupt = true
		case slices.Contains(keyConfig.Left, keyStr):
			*tempJoypad &^= 0b10
			reqInterrupt = true
		case slices.Contains(keyConfig.Up, keyStr):
			*tempJoypad &^= 0b100
			reqInterrupt = true
		case slices.Contains(keyConfig.Down, keyStr):
			*tempJoypad &^= 0b1000
			reqInterrupt = true
		}
	}

	for _, button := range buttons {

		buttonStr := int(button)

		switch {
		case slices.Contains(buttonConfig.A, buttonStr):
			*tempJoypad &^= 0b10000
			reqInterrupt = true
		case slices.Contains(buttonConfig.B, buttonStr):
			*tempJoypad &^= 0b100000
			reqInterrupt = true
		case slices.Contains(buttonConfig.Select, buttonStr):
			*tempJoypad &^= 0b1000000
			reqInterrupt = true
		case slices.Contains(buttonConfig.Start, buttonStr):
			*tempJoypad &^= 0b10000000
			reqInterrupt = true
		case slices.Contains(buttonConfig.Right, buttonStr):
			*tempJoypad &^= 0b1
			reqInterrupt = true
		case slices.Contains(buttonConfig.Left, buttonStr):
			*tempJoypad &^= 0b10
			reqInterrupt = true
		case slices.Contains(buttonConfig.Up, buttonStr):
			*tempJoypad &^= 0b100
			reqInterrupt = true
		case slices.Contains(buttonConfig.Down, buttonStr):
			*tempJoypad &^= 0b1000
			reqInterrupt = true
		}

	}

	if reqInterrupt {
		const JOYPAD uint8 = 0b10000
		gb.RequestInterrupt(JOYPAD)
	}
}
