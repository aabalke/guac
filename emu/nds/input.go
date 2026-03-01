package nds

import (
	"fmt"
	"slices"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/input"
	"github.com/hajimehoshi/ebiten/v2"
)

var _ = fmt.Sprint

func (nds *Nds) InputHandler(justKeys, keys []ebiten.Key, buttons []ebiten.StandardGamepadButton, mouse *input.Mouse, frame uint64) {

	var (
		keyCfg    = config.Conf.Nds.KeyboardConfig
		buttonCfg = config.Conf.Nds.ControllerConfig
		k         = &nds.mem.Keypad.KEYINPUT
		k2        = &nds.mem.Keypad.KEYINPUT2
	)

	*k = 0x3FF
	*k2 |= 0b0100_1011
	*k2 &^= 0b1000_0000

	mouseInput(nds, mouse, k2)

	for _, key := range keys {
		switch keyStr := key.String(); {
		case slices.Contains(keyCfg.A, keyStr):
			*k &^= 1 << 0
		case slices.Contains(keyCfg.B, keyStr):
			*k &^= 1 << 1
		case slices.Contains(keyCfg.Select, keyStr):
			*k &^= 1 << 2
		case slices.Contains(keyCfg.Start, keyStr):
			*k &^= 1 << 3
		case slices.Contains(keyCfg.Right, keyStr):
			*k &^= 1 << 4
		case slices.Contains(keyCfg.Left, keyStr):
			*k &^= 1 << 5
		case slices.Contains(keyCfg.Up, keyStr):
			*k &^= 1 << 6
		case slices.Contains(keyCfg.Down, keyStr):
			*k &^= 1 << 7
		case slices.Contains(keyCfg.R, keyStr):
			*k &^= 1 << 8
		case slices.Contains(keyCfg.L, keyStr):
			*k &^= 1 << 9
		case slices.Contains(keyCfg.X, keyStr):
			*k2 &^= 1 << 0
		case slices.Contains(keyCfg.Y, keyStr):
			*k2 &^= 1 << 1
		}
	}

	for _, key := range justKeys {
		switch keyStr := key.String(); {
		case slices.Contains(keyCfg.LayoutToggle, keyStr):
			nds.Screen.inputHandler(SCREEN_LAYOUT)
		case slices.Contains(keyCfg.SizingToggle, keyStr):
			nds.Screen.inputHandler(SCREEN_SIZING)
		case slices.Contains(keyCfg.RotationToggle, keyStr):
			nds.Screen.inputHandler(SCREEN_ROTATION)
		case slices.Contains(keyCfg.ExportScene, keyStr):
            nds.ppu.Rasterizer.Export.Export()
		}
	}

	for _, button := range buttons {
		switch buttonStr := int(button); {
		case slices.Contains(buttonCfg.A, buttonStr):
			*k &^= 1 << 0
		case slices.Contains(buttonCfg.B, buttonStr):
			*k &^= 1 << 1
		case slices.Contains(buttonCfg.Select, buttonStr):
			*k &^= 1 << 2
		case slices.Contains(buttonCfg.Start, buttonStr):
			*k &^= 1 << 3
		case slices.Contains(buttonCfg.Right, buttonStr):
			*k &^= 1 << 4
		case slices.Contains(buttonCfg.Left, buttonStr):
			*k &^= 1 << 5
		case slices.Contains(buttonCfg.Up, buttonStr):
			*k &^= 1 << 6
		case slices.Contains(buttonCfg.Down, buttonStr):
			*k &^= 1 << 7
		case slices.Contains(buttonCfg.R, buttonStr):
			*k &^= 1 << 8
		case slices.Contains(buttonCfg.L, buttonStr):
			*k &^= 1 << 9
		case slices.Contains(buttonCfg.X, buttonStr):
			*k2 &^= 1 << 0
		case slices.Contains(buttonCfg.Y, buttonStr):
			*k2 &^= 1 << 1
		}
	}

	if nds.mem.Keypad.KeyIRQ() {
		nds.arm9.Irq.SetIRQ(12)
		nds.arm7.Irq.SetIRQ(12)
	}
}

func mouseInput(nds *Nds, mouse *input.Mouse, k2 *uint16) {

	abs := nds.Screen.BtmAbs
	tsc := &nds.mem.Spi.Tsc

	if !mouse.DraggedLeft {
		tsc.TouchActive = false
		return
	}

	// effectively rot, translate of real mouse coords to rotated bottom screen coords

	switch nds.Screen.Rotation {
	case ROT_0:

		if inBounds := (mouse.X >= abs.L &&
			mouse.X < abs.R &&
			mouse.Y >= abs.T &&
			mouse.Y < abs.B); !inBounds {
			tsc.TouchActive = false
			return
		}

		s := float32(SCREEN_WIDTH) / float32(abs.W)
		tsc.TouchX = uint16(float32(mouse.X-abs.L)*s) - 1
		tsc.TouchY = uint16(float32(mouse.Y-abs.T)*s) - 1

	case ROT_90:

		if inBounds := (mouse.X >= abs.B &&
			mouse.X < abs.T &&
			mouse.Y >= abs.L &&
			mouse.Y < abs.R); !inBounds {
			tsc.TouchActive = false
			return
		}

		s := float32(SCREEN_WIDTH) / float32(abs.H)
		tsc.TouchX = uint16(float32(mouse.Y-abs.L)*s) - 1
		tsc.TouchY = uint16(float32(abs.T-mouse.X)*s) - 1

	case ROT_180:

		if inBounds := (mouse.X >= abs.R &&
			mouse.X < abs.L &&
			mouse.Y >= abs.B &&
			mouse.Y < abs.T); !inBounds {
			tsc.TouchActive = false
			return
		}

		s := float32(SCREEN_WIDTH) / float32(abs.W)
		tsc.TouchX = uint16(float32(abs.L-mouse.X)*s) - 1
		tsc.TouchY = uint16(float32(abs.T-mouse.Y)*s) - 1

	case ROT_270:

		if inBounds := (mouse.X >= abs.T &&
			mouse.X < abs.B &&
			mouse.Y >= abs.R &&
			mouse.Y < abs.L); !inBounds {
			tsc.TouchActive = false
			return
		}

		s := float32(SCREEN_WIDTH) / float32(abs.H)
		tsc.TouchX = uint16(float32(abs.L-mouse.Y)*s) - 1
		tsc.TouchY = uint16(float32(mouse.X-abs.T)*s) - 1
	}

	tsc.TouchActive = true
	*k2 &^= 0b100_0000
}
