package ui

import (
	"github.com/ebitenui/ebitenui"
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
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	}

	HEX_KEYS = []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "A", "B", "C", "D", "E", "F",
	}

	ALPHA_KEYS_UPPER = []string{
		"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
		"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
	}

	ALPHA_KEYS_LOWER = []string{
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
		"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
	}

	KEYS_KEY_CONTROLLER = []string{
		"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
		"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
		"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
		",", " ",
	}
)

type Keyboard struct {
	label  *widget.Text
	caller *widget.TextInput
	text   *widget.TextInput
	ui     *ebitenui.UI // this should be identical to main ui
	prev   widget.Containerer
	root   *widget.Container
	main   *widget.Container
	top    *widget.Container
	alpha  *widget.Container
	dec    *widget.Container
	hex    *widget.Container
	Keys   []string
	res    *Resources
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

		widget.ContainerOpts.BackgroundImage(k.res.bg),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Spacing(4, 16),
			widget.GridLayoutOpts.Stretch(
				[]bool{},
				[]bool{},
			),
		)),
	)

	k.top = widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionCenter,
				VerticalPosition:   widget.GridLayoutPositionCenter,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Spacing(4, 4),
			widget.GridLayoutOpts.Stretch(
				[]bool{false, true},
				[]bool{},
			),
		)),
	)

	k.main.AddChild(k.top)
	k.root.AddChild(k.main)

	k.alpha = k.buildBoard(8, KEYS_KEY_CONTROLLER)
	k.dec = k.buildBoard(3, DEC_KEYS)
	k.hex = k.buildBoard(4, HEX_KEYS)

	return k
}

func (k *Keyboard) buildBoard(columns int, keys []string) *widget.Container {

	if columns > 8 {
		panic("keyboard is not setup to handle more than 8 columns")
	}

	board := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionCenter,
				VerticalPosition:   widget.GridLayoutPositionCenter,
			}),
		),

		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Spacing(4, 4),
		)),
	)

	l := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionCenter,
				VerticalPosition:   widget.GridLayoutPositionCenter,
			}),
		),

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
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionCenter,
				VerticalPosition:   widget.GridLayoutPositionCenter,
			}),
		),

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
			k.text.SetText(k.text.GetText() + key)
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
		if s := k.text.GetText(); len(s) != 0 {
			k.text.SetText(s[:len(s)-1])
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
		widget.TextOpts.Padding(&widget.Insets{Right: 16}),
		widget.TextOpts.Text(
			label,
			k.res.fonts.smallFace,
			*k.res.fgClr,
		),
		widget.TextOpts.Position(
			widget.TextPositionStart,
			widget.TextPositionCenter,
		))

	k.text = _newTextBoxInput(v)

	k.top.RemoveChildren()
	k.top.AddChild(k.label, k.text)

	k.main.RemoveChildren()
	k.main.AddChild(k.top)

	switch board {
	case BOARD_ALPHA:
		k.main.AddChild(k.alpha)
	case BOARD_DEC:
		k.main.AddChild(k.dec)
	case BOARD_HEX:
		k.main.AddChild(k.hex)
	case BOARD_KEYBIND:
		k.main.AddChild(k.alpha)
	}

	k.ui.SetFocusedWidget(k.ui.Container.GetFocusers()[1])
}

func (k *Keyboard) Close(save bool) {
	k.ui.Container = k.prev
	k.ui.SetFocusedWidget(k.caller)
	if save {
		k.caller.SetText(k.text.GetText())
	}
}

func _newTextBoxInput(value any) *widget.TextInput {
	input := widget.NewTextInput()
	input.SetText(toString(value))
	return input
}
