package nds

import (
	"slices"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/input"
	"github.com/hajimehoshi/ebiten/v2"
)

func (nds *Nds) InputHandler(keys []ebiten.Key, buttons []ebiten.StandardGamepadButton, mouse *input.Mouse, frame uint64) {

	var (
		keyCfg    = config.Conf.Nds.KeyboardConfig
		buttonCfg = config.Conf.Nds.ControllerConfig
		k         = &nds.mem.Keypad.KEYINPUT
		k2        = &nds.mem.Keypad.KEYINPUT2
	)

	*k = 0b11_1111_1111
	*k2 |= 0b0100_1011
	*k2 &^= 0b1000_0000

	mouseInput(nds, mouse, k2)

	for _, key := range keys {

		keyStr := key.String()

		switch {
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

	for _, button := range buttons {

		buttonStr := int(button)

		switch {
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

	if inBounds := (
        mouse.X >= abs.L &&
		mouse.X <  abs.R &&
		mouse.Y >= abs.T &&
		mouse.Y <  abs.B); !inBounds || !mouse.DraggedLeft {

		tsc.TouchActive = false

		return
	}

	tsc.TouchActive = true

	s := float32(SCREEN_WIDTH) / float32(abs.W)

	tsc.TouchX = uint16(float32(mouse.X-abs.L) * s)
	tsc.TouchY = uint16(float32(mouse.Y-abs.T) * s)
	*k2 &^= 0b100_0000
}
