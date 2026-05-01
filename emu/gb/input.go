package gb

import (
	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2"
	"slices"
)

func (gb *GameBoy) InputHandler(keys []ebiten.Key, buttons []ebiten.StandardGamepadButton) {

	var (
		keyConfig    = config.Conf.Gb.KeyboardConfig
		buttonConfig = config.Conf.Gb.ControllerConfig
		k            = &gb.Joypad
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
		}
	}

	for _, button := range buttons {
		switch {
		case slices.Contains(buttonConfig.A, button):
			*k &^= 1 << 4
		case slices.Contains(buttonConfig.B, button):
			*k &^= 1 << 5
		case slices.Contains(buttonConfig.Select, button):
			*k &^= 1 << 6
		case slices.Contains(buttonConfig.Start, button):
			*k &^= 1 << 7
		case slices.Contains(buttonConfig.Right, button):
			*k &^= 1 << 0
		case slices.Contains(buttonConfig.Left, button):
			*k &^= 1 << 1
		case slices.Contains(buttonConfig.Up, button):
			*k &^= 1 << 2
		case slices.Contains(buttonConfig.Down, button):
			*k &^= 1 << 3
		}
	}

	if *k != 0xFF {
		gb.SetIrq(IRQ_JPD)
	}
}

func (gb *GameBoy) getJoypad() uint8 {
	joyp := gb.MemoryBus.JoypadReg
	if dpad := (joyp>>4)&1 == 0; dpad {
		return (joyp & 0x30) | (gb.Joypad & 0xF) | 0xC0
	} else if ssba := (joyp>>5)&1 == 0; ssba {
		return (joyp & 0x30) | (gb.Joypad >> 4) | 0xC0
	} else {
		return joyp | 0xCF // all released
	}
}
