package nds

import (
	"slices"

	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2"
)

func (nds *Nds) InputHandler(justKeys, keys []ebiten.Key, justButtons, buttons []ebiten.StandardGamepadButton) {
	k := &nds.mem.Key.Input
	k2 := &nds.mem.Key.Input2
	*k = 0x3FF
	*k2 = 0x4B

	nds.mouseInput()

	keyCfg := config.Conf.Nds.Keyboard
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
			*nds.Screen.Layout = (*nds.Screen.Layout + 1) % 3
		case slices.Contains(keyCfg.SizingToggle, key):
			*nds.Screen.Sizing = (*nds.Screen.Sizing + 1) % 3
		case slices.Contains(keyCfg.RotationToggle, key):
			*nds.Screen.Rotation = (*nds.Screen.Rotation + 1) % 4
		case slices.Contains(keyCfg.ExportScene, key):
			nds.ppu.Rasterizer.Export.Export()
		}
	}

	buttonCfg := config.Conf.Nds.Controller
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
			*nds.Screen.Layout = (*nds.Screen.Layout + 1) % 3
		case slices.Contains(buttonCfg.SizingToggle, button):
			*nds.Screen.Sizing = (*nds.Screen.Sizing + 1) % 3
		case slices.Contains(buttonCfg.RotationToggle, button):
			*nds.Screen.Rotation = (*nds.Screen.Rotation + 1) % 4
		case slices.Contains(buttonCfg.ExportScene, button):
			nds.ppu.Rasterizer.Export.Export()
		}
	}

	nds.mem.Key.KeyIRQ()
}

func (nds *Nds) mouseInput() {
	if dragged := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft); !dragged {
		return
	}

	inv := nds.Screen.TouchGeoM
	if !inv.IsInvertible() {
		return
	}

	inv.Invert()
	x, y := ebiten.CursorPosition()
	tx, ty := inv.Apply(float64(x), float64(y))

	if tx < 0 || tx >= SCREEN_WIDTH || ty < 0 || ty >= SCREEN_HEIGHT {
		return
	}

	nds.mem.Spi.Tsc.TouchX = uint16(tx)
	nds.mem.Spi.Tsc.TouchY = uint16(ty)
	nds.mem.Key.Input2 &^= 0x40
}
