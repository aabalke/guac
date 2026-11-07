package ui

import (
	"image/color"

	"github.com/aabalke/guac/utils"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

type BoardType int

const (
	BOARD_ALPHA BoardType = iota
	BOARD_DEC
	BOARD_HEX
	BOARD_KEYBIND
)

var (
	DEC_KEYS = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

	HEX_KEYS = []string{
		"0", "1", "2", "3", "4", "5", "6", "7",
		"8", "9", "A", "B", "C", "D", "E", "F",
	}

	KEYS_KEY_CONTROLLER = []string{
		"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
		"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
		"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", ",", " ",
	}
)

func openKeyboard(ui *Ui, caller *widget.TextInput, board BoardType, label string, v any, validation func(string) (bool, *string)) {
	callerPageId := ui.PageId
	callerRoot := ui.ui.Container
	callerReturnFocuser := ui.returnFocuser

	ui.PageId = PAGE_KEYBOARD

	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(ui.res.bg),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	main := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),

		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(4),
		)),
	)

	top := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Spacing(16, 0),
		)),
	)

	boardContainer := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),

		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(4),
		)),
	)

	labelWidget := widget.NewText(
		widget.TextOpts.Text(
			label,
			ui.res.fonts.smallFace,
			*ui.res.fgClr,
		),
	)

	var inputWidget widget.PreferredSizeLocateableWidget
	if color, ok := v.(*color.Color); ok {
		inputWidget = newKeyboardColorInput(ui.res.fgClr, color, validation)
	} else {
		inputWidget = newKeyboardTextBoxInput(v, 128*3, validation)
	}

	controlButtons, cancel, enter := buildControlButtons(ui.res, inputWidget)

	var keys *widget.Container
	switch board {
	case BOARD_ALPHA:
		keys = buildBoard(ui.res, inputWidget, 8, ui.res.localization.AsSlice("settings.alphabet"))
	case BOARD_DEC:
		keys = buildBoard(ui.res, inputWidget, 3, DEC_KEYS)
	case BOARD_HEX:
		keys = buildBoard(ui.res, inputWidget, 4, HEX_KEYS)
	case BOARD_KEYBIND:
		keys = buildBoard(ui.res, inputWidget, 8, KEYS_KEY_CONTROLLER)
	}

	boardContainer.AddChild(keys, controlButtons)
	top.AddChild(labelWidget, inputWidget)
	main.AddChild(top, boardContainer)
	root.AddChild(main)

	ui.ui.Container = root
	ui.ui.SetFocusedWidget(ui.ui.Container.GetFocusers()[1])

	ui.returnFocuser = cancel

	exit := func(save bool) {
		ui.ui.Container = callerRoot
		ui.PageId = callerPageId
		ui.ui.SetFocusedWidget(caller)
		ui.returnFocuser = callerReturnFocuser
		if save {
			switch input := inputWidget.(type) {
			case *widget.TextInput:
				caller.SetText(input.GetText())
			case *widget.Container:
				text := input.GetFocusers()[0].(*widget.TextInput)
				caller.SetText(text.GetText())
			}
		}
	}

	cancel.ClickedEvent.RemoveAllHandlers()
	cancel.ClickedEvent.AddHandler(func(any) { exit(false) })

	enter.ClickedEvent.RemoveAllHandlers()
	enter.ClickedEvent.AddHandler(func(any) { exit(true) })
}

func buildControlButtons(res *Resources, input widget.PreferredSizeLocateableWidget) (*widget.Container, *widget.Button, *widget.Button) {
	// TODO: NEED TO LOCALIZE

	var f func()
	switch input := input.(type) {
	case *widget.TextInput:
		f = func() {
			if s := input.GetText(); len(s) != 0 {
				input.SetText(s[:len(s)-1])
			}
		}
	case *widget.Container: // color
		f = func() {
			children := input.Children()
			text := children[0].(*widget.TextInput)
			if s := text.GetText(); len(s) != 0 {
				text.SetText(s[:len(s)-1])
			}

			clr := image.NewBorderedNineSliceColor(utils.HexToColor(text.GetText()), *res.fgClr, 2)
			colorBox := children[1].(*widget.Container)
			colorBox.SetBackgroundImage(clr)
		}
	}

	backspace := newKeyboardKey(res, "backspace", 160, f)
	cancel := newKeyboardKey(res, "cancel", 160, func() {})
	enter := newKeyboardKey(res, "enter", 160, func() {})

	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Spacing(4, 4),
			widget.GridLayoutOpts.Stretch(
				[]bool{true},
				[]bool{},
			),
		)),
	)

	c.AddChild(backspace, cancel, enter)

	return c, cancel, enter
}

func buildBoard(res *Resources, input widget.PreferredSizeLocateableWidget, cols int, keys []string) *widget.Container {
	if cols > 8 {
		panic("keyboard is not setup to handle more than 8 columns")
	}

	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(cols),
			widget.GridLayoutOpts.Spacing(4, 4),
			widget.GridLayoutOpts.Stretch(
				[]bool{true, true, true, true, true, true, true, true},
				[]bool{},
			),
		)),
	)

	var f func(key string)
	switch input := input.(type) {
	case *widget.TextInput:
		f = func(key string) {
			input.SetText(input.GetText() + key)
		}
	case *widget.Container:
		f = func(key string) {
			children := input.Children()
			text := children[0].(*widget.TextInput)
			colorBox := children[1].(*widget.Container)
			v := text.GetText() + key
			text.SetText(v)
			clr := image.NewBorderedNineSliceColor(utils.HexToColor(v), *res.fgClr, 2)
			colorBox.SetBackgroundImage(clr)
		}
	}

	for _, key := range keys {
		c.AddChild(newKeyboardKey(res, key, 64, func() { f(key) }))
	}

	return c
}

func newKeyboardKey(res *Resources, label string, minWidth int, f func()) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				MaxWidth:  minWidth,
				MaxHeight: 64,
			}),

			widget.WidgetOpts.MinSize(minWidth, 64),
		),

		widget.ButtonOpts.Image(res.buttonBorderedImage),

		widget.ButtonOpts.Text(
			label,
			res.fonts.smallFace,
			&widget.ButtonTextColor{
				Idle: *res.fgClr,
			},
		),

		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			f()
		}),
	)
}

func newKeyboardTextBoxInput(value any, minSize int, validation func(string) (bool, *string)) *widget.TextInput {
	input := widget.NewTextInput(
		widget.TextInputOpts.Validation(validation),
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(minSize, 0),
		),
	)
	input.SetText(toString(value))
	return input
}

func newKeyboardColorInput(fgClr, value *color.Color, validation func(string) (bool, *string)) widget.PreferredSizeLocateableWidget {
	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{}, []bool{true}),
			widget.GridLayoutOpts.Spacing(4, 0),
		)),
	)

	input := newKeyboardTextBoxInput(value, 128, validation)

	colorBox := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(128, 0)),
		widget.ContainerOpts.BackgroundImage(image.NewBorderedNineSliceColor(*value, *fgClr, 2)),
	)

	c.AddChild(input, colorBox)

	return c
}
