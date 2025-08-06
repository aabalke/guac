package gba

import (
	"slices"

	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2"
)

func (gba *GBA) InputHandler(keys []ebiten.Key, buttons []ebiten.StandardGamepadButton) {

	keyConfig := config.Conf.Gba.KeyboardConfig
	buttonConfig := config.Conf.Gba.ControllerConfig

	jp := &gba.Keypad.KEYINPUT

	*jp = 0b11_1111_1111

	for _, key := range keys {

		keyStr := key.String()

		switch {
		case slices.Contains(keyConfig.A, keyStr):
			*jp &^= 0b1
		case slices.Contains(keyConfig.B, keyStr):
			*jp &^= 0b10
		case slices.Contains(keyConfig.Select, keyStr):
			*jp &^= 0b100
		case slices.Contains(keyConfig.Start, keyStr):
			*jp &^= 0b1000
		case slices.Contains(keyConfig.Right, keyStr):
			*jp &^= 0b10000
		case slices.Contains(keyConfig.Left, keyStr):
			*jp &^= 0b100000
		case slices.Contains(keyConfig.Up, keyStr):
			*jp &^= 0b1000000
		case slices.Contains(keyConfig.Down, keyStr):
			*jp &^= 0b10000000
		case slices.Contains(keyConfig.R, keyStr):
			*jp &^= 0b100000000
		case slices.Contains(keyConfig.L, keyStr):
			*jp &^= 0b1000000000
		}
	}

	for _, button := range buttons {

		buttonStr := int(button)

		switch {
		case slices.Contains(buttonConfig.A, buttonStr):
			*jp &^= 0b1
		case slices.Contains(buttonConfig.B, buttonStr):
			*jp &^= 0b10
		case slices.Contains(buttonConfig.Select, buttonStr):
			*jp &^= 0b100
		case slices.Contains(buttonConfig.Start, buttonStr):
			*jp &^= 0b1000
		case slices.Contains(buttonConfig.Right, buttonStr):
			*jp &^= 0b10000
		case slices.Contains(buttonConfig.Left, buttonStr):
			*jp &^= 0b100000
		case slices.Contains(buttonConfig.Up, buttonStr):
			*jp &^= 0b1000000
		case slices.Contains(buttonConfig.Down, buttonStr):
			*jp &^= 0b10000000
		case slices.Contains(buttonConfig.R, buttonStr):
			*jp &^= 0b100000000
		case slices.Contains(buttonConfig.L, buttonStr):
			*jp &^= 0b1000000000
		}
	}

	if gba.Keypad.keyIRQ() {
		gba.Irq.setIRQ(12)
	}
}
