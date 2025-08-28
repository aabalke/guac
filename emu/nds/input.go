package nds

import (
	"slices"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/input"
	"github.com/hajimehoshi/ebiten/v2"
)

func (nds *Nds) InputHandler(keys []ebiten.Key, buttons []ebiten.StandardGamepadButton, mouse *input.Mouse) {

	keyConfig := config.Conf.Nds.KeyboardConfig
	buttonConfig := config.Conf.Nds.ControllerConfig

	k := &nds.mem.Keypad.KEYINPUT
	k2 := &nds.mem.Keypad.KEYINPUT2

	*k = 0b11_1111_1111
	*k2 |=  0b0100_0011
	*k2 &^= 0b1000_0000

    mouseInput(nds, mouse, k2)

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
		case slices.Contains(keyConfig.X, keyStr):
			*k2 &^= 0b1
		case slices.Contains(keyConfig.Y, keyStr):
			*k2 &^= 0b10
		case slices.Contains(keyConfig.Y, keyStr):
			*k2 &^= 0b10
		case slices.Contains(keyConfig.Hinge, keyStr):
			*k2 |= 0b1000_0000
		}
	}

	for _, button := range buttons {

		buttonStr := int(button)

		switch {
		case slices.Contains(buttonConfig.A, buttonStr):
			*k &^= 0b1
		case slices.Contains(buttonConfig.B, buttonStr):
			*k &^= 0b10
		case slices.Contains(buttonConfig.Select, buttonStr):
			*k &^= 0b100
		case slices.Contains(buttonConfig.Start, buttonStr):
			*k &^= 0b1000
		case slices.Contains(buttonConfig.Right, buttonStr):
			*k &^= 0b10000
		case slices.Contains(buttonConfig.Left, buttonStr):
			*k &^= 0b100000
		case slices.Contains(buttonConfig.Up, buttonStr):
			*k &^= 0b1000000
		case slices.Contains(buttonConfig.Down, buttonStr):
			*k &^= 0b10000000
		case slices.Contains(buttonConfig.R, buttonStr):
			*k &^= 0b100000000
		case slices.Contains(buttonConfig.L, buttonStr):
			*k &^= 0b1000000000
		case slices.Contains(buttonConfig.X, buttonStr):
			*k2 &^= 0b1
		case slices.Contains(buttonConfig.Y, buttonStr):
			*k2 &^= 0b10
		case slices.Contains(buttonConfig.Hinge, buttonStr):
			*k2 |= 0b1000_0000
		}
	}

	if nds.mem.Keypad.KeyIRQ() {
		nds.arm9.Irq.SetIRQ(12)
		nds.arm7.Irq.SetIRQ(12)
	}
}

func mouseInput(nds *Nds, mouse *input.Mouse, k2 *uint16) {

    tsc := &nds.mem.Spi.Tsc

    if inBounds := (
        mouse.X >= nds.BtmAbs.L && mouse.X < nds.BtmAbs.R &&
        mouse.Y >= nds.BtmAbs.T && mouse.Y < nds.BtmAbs.B);
        !inBounds || !mouse.DraggedLeft {
            tsc.TouchX = 0x000
            tsc.TouchY = 0xFFF
            return
    }

    s := float32(SCREEN_WIDTH) / float32(nds.BtmAbs.W)

    tsc.TouchX = uint16(float32(mouse.X - nds.BtmAbs.L) * s)
    tsc.TouchY = uint16(float32(mouse.Y - nds.BtmAbs.T) * s)
    *k2 &^= 0b100_0000
}
