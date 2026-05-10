package ui

import (
	"fmt"
	"slices"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/utils"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) GetGamepadButtons() (justButtons, buttons []ebiten.StandardGamepadButton) {
	g.gamepadIdBuf = inpututil.AppendJustConnectedGamepadIDs(g.gamepadIdBuf[:0])

	for _, id := range g.gamepadIdBuf {
		fmt.Printf("Gamepad Connected id %d\n", id)
		g.ui.toast.AddMessage(g.ui.res.localization.Toast.ControllerConnected)
		g.gamepadIds[id] = struct{}{}

		g.ui.focus.FocusSidebar(0)
	}

	for id := range g.gamepadIds {
		if inpututil.IsGamepadJustDisconnected(id) {
			fmt.Printf("Gamepad Disconnected id %d\n", id)
			g.ui.toast.AddMessage(g.ui.res.localization.Toast.ControllerDisconnected)
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

	utils.MakeUnique(&justButtons)
	utils.MakeUnique(&buttons)

	return justButtons, buttons
}

func (g *Game) GetInput() (justKeys, keys []ebiten.Key, justButtons, buttons []ebiten.StandardGamepadButton) {
	g.mouse.Update()

	justKeys = inpututil.AppendJustPressedKeys(justKeys)
	keys = inpututil.AppendPressedKeys(keys)
	justButtons, buttons = g.GetGamepadButtons()

	if !ebiten.IsFocused() {
		return
	}

	keyConfig := config.Conf.General.Keyboard
	buttonConfig := config.Conf.General.Controller

	for _, key := range justKeys {
		switch {
		case slices.Contains(keyConfig.Fullscreen, key):
			ebiten.SetFullscreen(!ebiten.IsFullscreen())
		case slices.Contains(keyConfig.Quit, key):
			g.quit = true
		case slices.Contains(keyConfig.Pause, key):
			g.TogglePause()
		case slices.Contains(keyConfig.Mute, key):
			g.ToggleMute()
		}
	}

	if g.ui.ui != nil {
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
	buttonConfig := config.Conf.General.Controller

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
			case slices.Contains(buttonConfig.Return, button):
				g.ui.focus.FocusLast()
			}
		}

	case PAGE_SETTINGS, PAGE_KEYBOARD:

		g.getFastFocus(buttons)

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

			case slices.Contains(buttonConfig.Return, button):
				if g.ui.PageId == PAGE_SETTINGS {
					g.ui.focus.FocusLastSidebar()
				} else {
					g.ui.ui.SetFocusedWidget(g.ui.keyboard.cancelButtons[g.ui.keyboard.currBoard])
				}

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

var (
	inputAcc float64
	inputHz  = 1.0 / 60
)

func (g *Game) getFastFocus(buttons []ebiten.StandardGamepadButton) {
	// this set is for scrolling through options quickly with the controller

	buttonConfig := config.Conf.General.Controller
	dt := 1.0 / float64(ebiten.ActualTPS())
	inputAcc += dt

	if inputAcc < inputHz {
		return
	}

	inputAcc -= inputHz

	for gp := range g.gamepadIds {
		for _, button := range buttons {
			if inpututil.StandardGamepadButtonPressDuration(gp, button) < int(60) {
				continue
			}
			switch {
			case slices.Contains(buttonConfig.Up, button):
				g.ui.ui.ChangeFocus(widget.FOCUS_NORTH)

			case slices.Contains(buttonConfig.Down, button):
				g.ui.ui.ChangeFocus(widget.FOCUS_SOUTH)

			case slices.Contains(buttonConfig.Right, button):
				g.ui.ui.ChangeFocus(widget.FOCUS_EAST)

			case slices.Contains(buttonConfig.Left, button):
				g.ui.ui.ChangeFocus(widget.FOCUS_WEST)
			}
		}
	}
}
