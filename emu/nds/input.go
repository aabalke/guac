package nds

import (
	"slices"

	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2"
)

func (nds *Nds) InputHandler(justKeys, keys []ebiten.Key, justButtons, buttons []ebiten.StandardGamepadButton) {
	var (
		keyCfg    = config.Conf.Nds.Keyboard
		buttonCfg = config.Conf.Nds.Controller
		k         = &nds.mem.Keypad.KEYINPUT
		k2        = &nds.mem.Keypad.KEYINPUT2
	)

	*k = 0x3FF
	*k2 |= 0b0100_1011
	*k2 &^= 0b1000_0000

	mouseInput(nds, k2)

	for _, key := range keys {
		switch {
		case slices.Contains(keyCfg.A, key):
			*k &^= 1 << 0
		case slices.Contains(keyCfg.B, key):
			*k &^= 1 << 1
		case slices.Contains(keyCfg.Select, key):
			*k &^= 1 << 2
		case slices.Contains(keyCfg.Start, key):
			*k &^= 1 << 3
		case slices.Contains(keyCfg.Right, key):
			*k &^= 1 << 4
		case slices.Contains(keyCfg.Left, key):
			*k &^= 1 << 5
		case slices.Contains(keyCfg.Up, key):
			*k &^= 1 << 6
		case slices.Contains(keyCfg.Down, key):
			*k &^= 1 << 7
		case slices.Contains(keyCfg.R, key):
			*k &^= 1 << 8
		case slices.Contains(keyCfg.L, key):
			*k &^= 1 << 9
		case slices.Contains(keyCfg.X, key):
			*k2 &^= 1 << 0
		case slices.Contains(keyCfg.Y, key):
			*k2 &^= 1 << 1
		case slices.Contains(keyCfg.Hinge, key):
			*k2 &^= 1 << 7
		}
	}

	for _, key := range justKeys {
		switch {
		case slices.Contains(keyCfg.LayoutToggle, key):
			nds.Screen.inputHandler(SCREEN_LAYOUT)
		case slices.Contains(keyCfg.SizingToggle, key):
			nds.Screen.inputHandler(SCREEN_SIZING)
		case slices.Contains(keyCfg.RotationToggle, key):
			nds.Screen.inputHandler(SCREEN_ROTATION)
		case slices.Contains(keyCfg.ExportScene, key):
			nds.ppu.Rasterizer.Export.Export()
		}
	}

	for _, button := range buttons {
		switch {
		case slices.Contains(buttonCfg.A, button):
			*k &^= 1 << 0
		case slices.Contains(buttonCfg.B, button):
			*k &^= 1 << 1
		case slices.Contains(buttonCfg.Select, button):
			*k &^= 1 << 2
		case slices.Contains(buttonCfg.Start, button):
			*k &^= 1 << 3
		case slices.Contains(buttonCfg.Right, button):
			*k &^= 1 << 4
		case slices.Contains(buttonCfg.Left, button):
			*k &^= 1 << 5
		case slices.Contains(buttonCfg.Up, button):
			*k &^= 1 << 6
		case slices.Contains(buttonCfg.Down, button):
			*k &^= 1 << 7
		case slices.Contains(buttonCfg.R, button):
			*k &^= 1 << 8
		case slices.Contains(buttonCfg.L, button):
			*k &^= 1 << 9
		case slices.Contains(buttonCfg.X, button):
			*k2 &^= 1 << 0
		case slices.Contains(buttonCfg.Y, button):
			*k2 &^= 1 << 1
		case slices.Contains(buttonCfg.Hinge, button):
			*k2 &^= 1 << 7
		}
	}

	for _, button := range justButtons {
		switch {
		case slices.Contains(buttonCfg.LayoutToggle, button):
			nds.Screen.inputHandler(SCREEN_LAYOUT)
		case slices.Contains(buttonCfg.SizingToggle, button):
			nds.Screen.inputHandler(SCREEN_SIZING)
		case slices.Contains(buttonCfg.RotationToggle, button):
			nds.Screen.inputHandler(SCREEN_ROTATION)
		case slices.Contains(buttonCfg.ExportScene, button):
			nds.ppu.Rasterizer.Export.Export()
		}
	}

	if nds.mem.Keypad.KeyIRQ() {
		nds.arm9.Irq.SetIRQ(12)
		nds.arm7.Irq.SetIRQ(12)
	}
}

func mouseInput(nds *Nds, k2 *uint16) {
	dragged := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	x, y := ebiten.CursorPosition()

	abs := nds.Screen.BtmAbs
	tsc := &nds.mem.Spi.Tsc

	if !dragged {
		tsc.TouchActive = false
		return
	}

	// effectively rot, translate of real coords to rotated bottom screen coords

	switch *nds.Screen.Rotation {
	case ROT_0:

		if inBounds := (x >= abs.L &&
			x < abs.R &&
			y >= abs.T &&
			y < abs.B); !inBounds {
			tsc.TouchActive = false
			return
		}

		s := float32(SCREEN_WIDTH) / float32(abs.W)
		tsc.TouchX = uint16(float32(x-abs.L)*s) - 1
		tsc.TouchY = uint16(float32(y-abs.T)*s) - 1

	case ROT_90:

		if inBounds := (x >= abs.B &&
			x < abs.T &&
			y >= abs.L &&
			y < abs.R); !inBounds {
			tsc.TouchActive = false
			return
		}

		s := float32(SCREEN_WIDTH) / float32(abs.H)
		tsc.TouchX = uint16(float32(y-abs.L)*s) - 1
		tsc.TouchY = uint16(float32(abs.T-x)*s) - 1

	case ROT_180:

		if inBounds := (x >= abs.R &&
			x < abs.L &&
			y >= abs.B &&
			y < abs.T); !inBounds {
			tsc.TouchActive = false
			return
		}

		s := float32(SCREEN_WIDTH) / float32(abs.W)
		tsc.TouchX = uint16(float32(abs.L-x)*s) - 1
		tsc.TouchY = uint16(float32(abs.T-y)*s) - 1

	case ROT_270:

		if inBounds := (x >= abs.T &&
			x < abs.B &&
			y >= abs.R &&
			y < abs.L); !inBounds {
			tsc.TouchActive = false
			return
		}

		s := float32(SCREEN_WIDTH) / float32(abs.H)
		tsc.TouchX = uint16(float32(abs.L-y)*s) - 1
		tsc.TouchY = uint16(float32(x-abs.T)*s) - 1
	}

	tsc.TouchActive = true
	*k2 &^= 0b100_0000
}
