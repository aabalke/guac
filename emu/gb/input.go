package gb

import (
	"slices"

	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2"
)

func (gb *GameBoy) InputHandler(keys []ebiten.Key, buttons []ebiten.StandardGamepadButton) {
	var (
		keyConfig    = config.Conf.Gb.KeyboardConfig
		buttonConfig = config.Conf.Gb.ControllerConfig
		k            = &gb.Joypad
	)

	*k = 0xFF

	for _, key := range keys {
		switch {
		case slices.Contains(keyConfig.A, key):
			*k &^= 1 << 4
		case slices.Contains(keyConfig.B, key):
			*k &^= 1 << 5
		case slices.Contains(keyConfig.Select, key):
			*k &^= 1 << 6
		case slices.Contains(keyConfig.Start, key):
			*k &^= 1 << 7
		case slices.Contains(keyConfig.Right, key):
			*k &^= 1 << 0
		case slices.Contains(keyConfig.Left, key):
			*k &^= 1 << 1
		case slices.Contains(keyConfig.Up, key):
			*k &^= 1 << 2
		case slices.Contains(keyConfig.Down, key):
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
