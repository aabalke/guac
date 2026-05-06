package ui

import (
	"image/color"

	"github.com/aabalke/guac/utils"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

const (
	BOARD_ALPHA = iota
	BOARD_DEC
	BOARD_HEX
	BOARD_KEYBIND
)

var (
	DEC_KEYS = []string{
		"0", "1", "2", "3", "4",
		"5", "6", "7", "8", "9",
	}

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

type Keyboard struct {
	label  *widget.Text
	caller *widget.TextInput
	widget widget.PreferredSizeLocateableWidget
	ui     *ebitenui.UI // this should be identical to main ui
	prev   widget.Containerer
	root   *widget.Container
	main   *widget.Container
	top    *widget.Container
	boards [4]*widget.Container

	cancelButton widget.Focuser

	Keys []string
	res  *Resources
}

func NewKeyboard(res *Resources) *Keyboard {

	k := &Keyboard{
		res: res,
	}

	k.root = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(k.res.bg),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	k.main = widget.NewContainer(
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

	k.top = widget.NewContainer(
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

	k.main.AddChild(k.top)
	k.root.AddChild(k.main)

	k.boards[BOARD_ALPHA] = k.buildBoard(8, KEYS_KEY_CONTROLLER) // will need to update based on lang
	k.boards[BOARD_DEC] = k.buildBoard(3, DEC_KEYS)
	k.boards[BOARD_HEX] = k.buildBoard(4, HEX_KEYS)
	k.boards[BOARD_KEYBIND] = k.buildBoard(8, KEYS_KEY_CONTROLLER)

	return k
}

func (k *Keyboard) buildBoard(columns int, keys []string) *widget.Container {

	if columns > 8 {
		panic("keyboard is not setup to handle more than 8 columns")
	}

	board := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),

		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(4),
		)),
	)

	l := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(columns),
			widget.GridLayoutOpts.Spacing(4, 4),
			widget.GridLayoutOpts.Stretch(
				[]bool{true, true, true, true, true, true, true, true},
				[]bool{},
			),
		)),
	)

	r := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Spacing(4, 4),
			widget.GridLayoutOpts.Stretch(
				[]bool{true},
				[]bool{},
			),
		)),
	)

	for _, key := range keys {
		l.AddChild(k.buildKey(key, 64, func() {

			switch input := k.widget.(type) {
			case *widget.TextInput:
				input.SetText(input.GetText() + key)
			case *widget.Container: // color
				children := input.Children()
				text := children[0].(*widget.TextInput)
				v := text.GetText() + key
				text.SetText(v)
				colorBox := children[1].(*widget.Container)
				clr := image.NewBorderedNineSliceColor(utils.HexToColor(v), *k.res.fgClr, 2)
				colorBox.SetBackgroundImage(clr)
			}
		}))
	}

	keyFocusers := l.GetFocusers()
	keyCnt := len(keyFocusers)
	for i := range keyFocusers {
		if north := i - columns; north >= 0 {
			keyFocusers[i].AddFocus(widget.FOCUS_NORTH, keyFocusers[north])
		}
		if south := i + columns; south < keyCnt {
			keyFocusers[i].AddFocus(widget.FOCUS_SOUTH, keyFocusers[south])
		}
		if west := i - 1; west%columns != columns-1 && west >= 0 {
			keyFocusers[i].AddFocus(widget.FOCUS_WEST, keyFocusers[west])
		}
		if east := i + 1; east%columns != 0 && east < keyCnt {
			keyFocusers[i].AddFocus(widget.FOCUS_EAST, keyFocusers[east])
		}
	}

	backspace := k.buildKey("backspace", 160, func() {
		switch input := k.widget.(type) {
		case *widget.TextInput:
			if s := input.GetText(); len(s) != 0 {
				input.SetText(s[:len(s)-1])
			}

		case *widget.Container: // color
			children := input.Children()
			text := children[0].(*widget.TextInput)
			if s := text.GetText(); len(s) != 0 {
				text.SetText(s[:len(s)-1])
			}

			clr := image.NewBorderedNineSliceColor(utils.HexToColor(text.GetText()), *k.res.fgClr, 2)
			colorBox := children[1].(*widget.Container)
			colorBox.SetBackgroundImage(clr)
		}
	})
	cancel := k.buildKey("cancel", 160, func() { k.Close(false) })
	enter := k.buildKey("enter", 160, func() { k.Close(true) })

	keyFocusers[columns-1].AddFocus(widget.FOCUS_EAST, backspace)
	backspace.AddFocus(widget.FOCUS_WEST, keyFocusers[columns-1])
	backspace.AddFocus(widget.FOCUS_SOUTH, cancel)

	keyFocusers[(columns*2)-1].AddFocus(widget.FOCUS_EAST, cancel)
	cancel.AddFocus(widget.FOCUS_WEST, keyFocusers[(columns*2)-1])
	cancel.AddFocus(widget.FOCUS_SOUTH, enter)
	cancel.AddFocus(widget.FOCUS_NORTH, backspace)

	keyFocusers[(columns*3)-1].AddFocus(widget.FOCUS_EAST, enter)
	enter.AddFocus(widget.FOCUS_WEST, keyFocusers[(columns*3)-1])
	enter.AddFocus(widget.FOCUS_NORTH, cancel)

	k.cancelButton = cancel

	r.AddChild(backspace, cancel, enter)
	board.AddChild(l, r)

	return board
}

func (k *Keyboard) buildKey(key string, minWidth int, f func()) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				MaxWidth:  minWidth,
				MaxHeight: 64,
			}),

			widget.WidgetOpts.MinSize(minWidth, 64),
		),

		widget.ButtonOpts.Image(k.res.buttonBorderedImage),

		widget.ButtonOpts.Text(
			key,
			k.res.fonts.smallFace,
			&widget.ButtonTextColor{
				Idle: *k.res.fgClr,
			},
		),

		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			f()
		}),
	)
}

func (k *Keyboard) Open(ui *Ui, caller *widget.TextInput, board int, label string, v any) {

	ui.PageId = PAGE_KEYBOARD

	k.ui = ui.ui
	k.prev = k.ui.Container
	k.ui.Container = k.root

	k.caller = caller

	k.label = widget.NewText(
		widget.TextOpts.Text(
			label,
			k.res.fonts.smallFace,
			*k.res.fgClr,
		),
	)

	if color, ok := v.(*color.Color); ok {
		k.widget = _newColorInput(ui.res.fgClr, color)
	} else {
		k.widget = _newTextBoxInput(v)
	}

	k.top.RemoveChildren()
	k.top.AddChild(k.label, k.widget)

	k.main.RemoveChildren()
	k.main.AddChild(k.top, k.boards[board])

	k.ui.SetFocusedWidget(k.ui.Container.GetFocusers()[1])
}

func (k *Keyboard) Close(save bool) {
	k.ui.Container = k.prev
	k.ui.SetFocusedWidget(k.caller)
	if save {
		switch input := k.widget.(type) {
		case *widget.TextInput:
			k.caller.SetText(input.GetText())
		case *widget.Container:
			text := input.GetFocusers()[0].(*widget.TextInput)
			k.caller.SetText(text.GetText())
		}
	}
}

func _newTextBoxInput(value any) *widget.TextInput {
	input := widget.NewTextInput(
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(128, 0),
		),
	)
	input.SetText(toString(value))
	return input
}

func _newColorInput(fgClr, value *color.Color) widget.PreferredSizeLocateableWidget {

	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{}, []bool{true}),
			widget.GridLayoutOpts.Spacing(4, 0),
		)),
	)

	input := _newTextBoxInput(value)
	clr := image.NewBorderedNineSliceColor(*value, *fgClr, 2)

	colorBox := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(128, 0),
		),

		widget.ContainerOpts.BackgroundImage(clr),
	)

	container.AddChild(input, colorBox)

	return container
}
