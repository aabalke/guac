package gameboy

import (
	"slices"

	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2"
)

func (gb *GameBoy) InputHandler(keys []ebiten.Key, buttons []ebiten.StandardGamepadButton) {

	var (
		keyConfig    = config.Conf.Gb.KeyboardConfig
		buttonConfig = config.Conf.Gba.ControllerConfig
		k            = &gb.Joypad
        noKeys = true
        noButs = true
	)

	*k = 0xFF

	for _, key := range keys {
		switch keyStr := key.String(); {
		case slices.Contains(keyConfig.A, keyStr):
			*k &^= 1 << 4
		case slices.Contains(keyConfig.B, keyStr):
			*k &^= 1 << 5
		case slices.Contains(keyConfig.Select, keyStr):
			*k &^= 1 << 6
		case slices.Contains(keyConfig.Start, keyStr):
			*k &^= 1 << 7
		case slices.Contains(keyConfig.Right, keyStr):
			*k &^= 1 << 0
		case slices.Contains(keyConfig.Left, keyStr):
			*k &^= 1 << 1
		case slices.Contains(keyConfig.Up, keyStr):
			*k &^= 1 << 2
		case slices.Contains(keyConfig.Down, keyStr):
			*k &^= 1 << 3
		default:
            noKeys = true
		}
	}

	for _, button := range buttons {
		switch buttonStr := int(button); {
		case slices.Contains(buttonConfig.A, buttonStr):
			*k &^= 1 << 4
		case slices.Contains(buttonConfig.B, buttonStr):
			*k &^= 1 << 5
		case slices.Contains(buttonConfig.Select, buttonStr):
			*k &^= 1 << 6
		case slices.Contains(buttonConfig.Start, buttonStr):
			*k &^= 1 << 7
		case slices.Contains(buttonConfig.Right, buttonStr):
			*k &^= 1 << 0
		case slices.Contains(buttonConfig.Left, buttonStr):
			*k &^= 1 << 1
		case slices.Contains(buttonConfig.Up, buttonStr):
			*k &^= 1 << 2
		case slices.Contains(buttonConfig.Down, buttonStr):
			*k &^= 1 << 3
		default:
            noButs = true
		}
	}

    if noKeys && noButs {
        return
    }

    gb.SetIrq(IRQ_JPD)
}
