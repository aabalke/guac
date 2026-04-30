package ui

import (
	"fmt"
	"slices"

	"github.com/aabalke/guac/config"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) GetGamepadButtons() (justButtons, buttons []ebiten.StandardGamepadButton) {

	g.gamepadIdBuf = inpututil.AppendJustConnectedGamepadIDs(g.gamepadIdBuf[:0])

	for _, id := range g.gamepadIdBuf {
		fmt.Printf("Gamepad Connected id %d\n", id)
		g.gamepadIds[id] = struct{}{}

		g.ui.focus.FocusSidebar(0)
	}

	for id := range g.gamepadIds {
		if inpututil.IsGamepadJustDisconnected(id) {
			fmt.Printf("Gamepad Disconnected id %d\n", id)
			delete(g.gamepadIds, id)

			if len(g.gamepadIds) == 0 {
				g.ui.focus.DeFocus()
			}
		}
	}

	for id := range g.gamepadIds {
		if !ebiten.IsStandardGamepadLayoutAvailable(id) {
			continue
		}

		justButtons = inpututil.AppendJustPressedStandardGamepadButtons(id, justButtons)
		buttons = inpututil.AppendPressedStandardGamepadButtons(id, buttons)
	}

	justButtons = uniqueButtons(&justButtons)
	buttons = uniqueButtons(&buttons)

	return justButtons, buttons
}

func uniqueButtons(in *[]ebiten.StandardGamepadButton) []ebiten.StandardGamepadButton {
	set := make(map[ebiten.StandardGamepadButton]struct{}, len(*in))

	for _, b := range *in {
		set[b] = struct{}{}
	}

	out := make([]ebiten.StandardGamepadButton, 0, len(set))
	for b := range set {
		out = append(out, b)
	}

	return out
}

func (g *Game) GetInput() (justKeys, keys []ebiten.Key, justButtons, buttons []ebiten.StandardGamepadButton) {

	g.mouse.Update()

	justKeys = inpututil.AppendJustPressedKeys(justKeys)
	keys = inpututil.AppendPressedKeys(keys)
	justButtons, buttons = g.GetGamepadButtons()

	if !ebiten.IsFocused() {
		return
	}

	keyConfig := config.Conf.General.KeyboardConfig
	buttonConfig := config.Conf.General.ControllerConfig

	for _, key := range justKeys {
		switch keyStr := key.String(); {
		case slices.Contains(keyConfig.Fullscreen, keyStr):
			ebiten.SetFullscreen(!ebiten.IsFullscreen())
		case slices.Contains(keyConfig.Quit, keyStr):
			g.quit = true
		case slices.Contains(keyConfig.Pause, keyStr):
			g.TogglePause()
		case slices.Contains(keyConfig.Mute, keyStr):
			g.ToggleMute()
		}
	}

	if g.ui != nil {
		g.ButtonInput(justButtons, buttons)
	}

	for _, button := range justButtons {
		switch {
		case slices.Contains(buttonConfig.Fullscreen, button):
			ebiten.SetFullscreen(!ebiten.IsFullscreen())
		case slices.Contains(buttonConfig.Quit, button):
			g.quit = true
		case slices.Contains(buttonConfig.Pause, button):
			g.TogglePause()
		case slices.Contains(buttonConfig.Mute, button):
			g.ToggleMute()
		}
	}

	return justKeys, keys, justButtons, buttons
}

func (g *Game) ButtonInput(justButtons, buttons []ebiten.StandardGamepadButton) {

	buttonConfig := config.Conf.General.ControllerConfig

	switch g.ui.PageId {
	case PAGE_HOME, PAGE_PAUSE:
		for _, button := range justButtons {
			switch {
			case slices.Contains(buttonConfig.Up, button):
				g.ui.ui.ChangeFocus(widget.FOCUS_NORTH)

			case slices.Contains(buttonConfig.Down, button):
				g.ui.ui.ChangeFocus(widget.FOCUS_SOUTH)

			case slices.Contains(buttonConfig.Select, button):
				if b, ok := g.ui.ui.GetFocusedWidget().(*widget.Button); ok {
					b.Click()
				}
			}
		}

	case PAGE_SETTINGS:

		// this set is for scrolling through options quickly with the controller
		// kicks in after x number of ticks
		for gp := range g.gamepadIds {
			for _, button := range buttons {
				if inpututil.StandardGamepadButtonPressDuration(gp, button) > 20 {
					switch {
					case slices.Contains(buttonConfig.Up, button):
						g.ui.ui.ChangeFocus(widget.FOCUS_NORTH)

					case slices.Contains(buttonConfig.Down, button):
						g.ui.ui.ChangeFocus(widget.FOCUS_SOUTH)
					}

				}
			}
		}

		for _, button := range justButtons {

			switch {
			case slices.Contains(buttonConfig.Up, button):
				g.ui.ui.ChangeFocus(widget.FOCUS_NORTH)

			case slices.Contains(buttonConfig.Down, button):
				g.ui.ui.ChangeFocus(widget.FOCUS_SOUTH)

			case slices.Contains(buttonConfig.Right, button):
				g.ui.ui.ChangeFocus(widget.FOCUS_EAST)

			case slices.Contains(buttonConfig.Left, button):
				g.ui.ui.ChangeFocus(widget.FOCUS_WEST)

			//case slices.Contains(buttonConfig., button):
			//switch g.ui.PrevPageId {
			//case PAGE_HOME:
			//	NewHome(g)
			//case PAGE_PAUSE:
			//	NewPause(g)
			//}

			case slices.Contains(buttonConfig.Select, button):
				switch w := g.ui.ui.GetFocusedWidget().(type) {
				case *widget.Button:
					w.Click()
				case *widget.Checkbox:
					w.Click()
				case *widget.TextInput:
					w.Submit()
				}
			}
		}

	}
}
