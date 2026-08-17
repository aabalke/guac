package gba

import (
	"slices"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/gba/gpio"
	"github.com/hajimehoshi/ebiten/v2"
)

func (gba *GBA) InputHandler(justKeys, keys []ebiten.Key, justButtons, buttons []ebiten.StandardGamepadButton) {
	k := &gba.Keypad.Input
	*k = 0x3FF

	keyConfig := config.Conf.Gba.Keyboard
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
		}
	}

	for _, key := range justKeys {
		if slices.Contains(keyConfig.RotationToggle, key) {
			r := &config.Conf.Gba.Rotation
			*r = (*r + 1) % 4
		}
	}

	buttonConfig := config.Conf.Gba.Controller
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

	for _, button := range justButtons {
		if slices.Contains(buttonConfig.RotationToggle, button) {
			r := &config.Conf.Gba.Rotation
			*r = (*r + 1) % 4
		}
	}

	if gba.Mem.Gpio != nil {
		for _, device := range gba.Mem.Gpio.Devices {
			d, ok := device.(*gpio.Solar)
			if !ok {
				continue
			}
			keyConfig := config.Conf.Gba.Keyboard
			for _, key := range keys {
				switch {
				case slices.Contains(keyConfig.SolarLevel0, key):
					d.SetLevel(0)
				case slices.Contains(keyConfig.SolarLevel1, key):
					d.SetLevel(25)
				case slices.Contains(keyConfig.SolarLevel2, key):
					d.SetLevel(50)
				case slices.Contains(keyConfig.SolarLevel3, key):
					d.SetLevel(75)
				case slices.Contains(keyConfig.SolarLevel4, key):
					d.SetLevel(100)
				}
			}
			buttonConfig := config.Conf.Gba.Controller
			for _, button := range buttons {
				switch {
				case slices.Contains(buttonConfig.SolarLevel0, button):
					d.SetLevel(0)
				case slices.Contains(buttonConfig.SolarLevel1, button):
					d.SetLevel(25)
				case slices.Contains(buttonConfig.SolarLevel2, button):
					d.SetLevel(50)
				case slices.Contains(buttonConfig.SolarLevel3, button):
					d.SetLevel(75)
				case slices.Contains(buttonConfig.SolarLevel4, button):
					d.SetLevel(100)
				}
			}
		}
	}

	gba.Keypad.KeyIRQ()
}
