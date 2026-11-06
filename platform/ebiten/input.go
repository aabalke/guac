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
	g.ui.gamepadIdBuf = inpututil.AppendJustConnectedGamepadIDs(g.ui.gamepadIdBuf[:0])

	for _, id := range g.ui.gamepadIdBuf {
		fmt.Printf("Gamepad Connected id %d\n", id)
		g.ui.toast.Message(g.ui.res.localization["toast.controller_connected"])
		g.ui.gamepadIds[id] = struct{}{}
		FocusFirst(g.ui)
	}

	for id := range g.ui.gamepadIds {
		if inpututil.IsGamepadJustDisconnected(id) {
			fmt.Printf("Gamepad Disconnected id %d\n", id)
			g.ui.toast.Message(g.ui.res.localization["toast.controller_disconnected"])
			delete(g.ui.gamepadIds, id)

			if len(g.ui.gamepadIds) == 0 {
				DeFocus(g.ui)
			}
		}
	}

	for id := range g.ui.gamepadIds {
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

func (g *Game) HandleGlobalInputs(justKeys []ebiten.Key, justButtons, buttons []ebiten.StandardGamepadButton) {
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
}

func (g *Game) ButtonInput(justButtons, buttons []ebiten.StandardGamepadButton) {
	buttonConfig := config.Conf.General.Controller

	for _, button := range justButtons {
		switch {
		case slices.Contains(buttonConfig.Up, button):
			g.dirFocus(button, widget.FOCUS_NORTH)

		case slices.Contains(buttonConfig.Down, button):
			g.dirFocus(button, widget.FOCUS_SOUTH)

		case slices.Contains(buttonConfig.Right, button):
			g.dirFocus(button, widget.FOCUS_EAST)

		case slices.Contains(buttonConfig.Left, button):
			g.dirFocus(button, widget.FOCUS_WEST)

		case slices.Contains(buttonConfig.Return, button):
			g.ui.ui.SetFocusedWidget(g.ui.returnFocuser)

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

func (g *Game) dirFocus(button ebiten.StandardGamepadButton, dir widget.FocusDirection) {
	g.ui.ui.ChangeFocus(dir)

	const DELAY = 20

	e := &TimedEvent{}
	switch dir {
	case widget.FOCUS_NORTH:
		e.eventType = EVENT_FAST_FOCUS_NORTH
	case widget.FOCUS_SOUTH:
		e.eventType = EVENT_FAST_FOCUS_SOUTH
	case widget.FOCUS_EAST:
		e.eventType = EVENT_FAST_FOCUS_EAST
	case widget.FOCUS_WEST:
		e.eventType = EVENT_FAST_FOCUS_WEST
	}

	g.ui.TimedEvents.CancelEventType(e.eventType)

	e.TicksLeft = DELAY
	e.OnTick = func() {
		if e.TicksLeft > 1 {
			return
		}

		var active bool

		for gp := range g.ui.gamepadIds {
			if inpututil.StandardGamepadButtonPressDuration(gp, button) != 0 {
				active = true
				break
			}
		}

		if active {
			g.ui.ui.ChangeFocus(dir)
			e.TicksLeft++
		}
	}

	g.ui.TimedEvents.AddEvent(e)
}
