package ui

import (
	"image/color"
	"strconv"
	"time"

	"github.com/aabalke/guac/utils"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/utilities/mobile"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	MIN_HEIGHT   = 32
	BUTTON_WIDTH = 256
)

func NewHeader(text string, res *Resources) *widget.Text {
	return widget.NewText(
		widget.TextOpts.Text(text, res.fonts.face, *res.fgClr),
		widget.TextOpts.Padding(&widget.Insets{24, 24, 24, 0}),
	)
}

func NewLabel(text string) *widget.Text {
	return widget.NewText(
		widget.TextOpts.TextLabel(text),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	)
}

func NewLinkText(text string) *widget.Text {
	return widget.NewText(
		widget.TextOpts.TextLabel(text),
		widget.TextOpts.LinkClickedHandler(func(args *widget.LinkEventArgs) {
			utils.OpenLink(args.Id)
		}),
		widget.TextOpts.ProcessBBCode(true),

		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
				StretchHorizontal:  true,
				StretchVertical:    true,
			}),
		),
	)
}

func NewSeparator() *widget.Container { return widget.NewContainer() }

func NewCheckbox(value *bool) widget.PreferredSizeLocateableWidget {
	state := widget.WidgetUnchecked
	if *value {
		state = widget.WidgetChecked
	}

	return widget.NewCheckbox(
		widget.CheckboxOpts.StateChangedHandler(func(args *widget.CheckboxChangedEventArgs) {
			*value = args.State == widget.WidgetChecked
		}),

		widget.CheckboxOpts.InitialState(state),
	)
}

func NewTextBoxInput(ui *Ui, board int, label string, value any, validation func(s string) (bool, *string)) *widget.TextInput {
	var input *widget.TextInput

	input = widget.NewTextInput(
		widget.TextInputOpts.MobileInputMode(mobile.TEXT),
		widget.TextInputOpts.Validation(validation),

		widget.TextInputOpts.SubmitOnEnter(false),
		widget.TextInputOpts.AllowDuplicateSubmit(true),

		widget.TextInputOpts.SubmitHandler(func(*widget.TextInputChangedEventArgs) {
			ui.keyboard.Open(ui, input, board, label, value, validation)
		}),
		widget.TextInputOpts.ChangedHandler(func(args *widget.TextInputChangedEventArgs) {
			fromString(value, args.InputText)
		}),

		widget.TextInputOpts.FocusHandler(func(args *widget.TextInputFocusEventArgs) {
			if args.Focused {
				ui.ActiveInputs++
			} else {
				ui.ActiveInputs--
			}
		}),
	)

	input.SetText(toString(value))

	return input
}

func NewKeybindInput(ui *Ui, v *[]ebiten.Key) widget.PreferredSizeLocateableWidget {
	return _newBindingInput(ui, v, nil)
}

func NewGamepadInput(ui *Ui, v *[]ebiten.StandardGamepadButton) widget.PreferredSizeLocateableWidget {
	return _newBindingInput(ui, nil, v)
}

func _newBindingInput(ui *Ui, keys *[]ebiten.Key, buttons *[]ebiten.StandardGamepadButton) widget.PreferredSizeLocateableWidget {
	placeholder := ""
	buttonContainerMin := widget.WidgetOpts.MinSize(64, 1)

	buttonText := func(label string) widget.ButtonOpt {
		return widget.ButtonOpts.Text(
			label,
			ui.res.fonts.smallFace,
			&widget.ButtonTextColor{
				Idle: *ui.res.fgClr,
			},
		)
	}

	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{true}),
			widget.GridLayoutOpts.Spacing(4, 0),
		)),
	)

	sideContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(buttonContainerMin),
	)

	buttonContainer := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
				StretchVertical:    true,
				StretchHorizontal:  true,
			}),
		),

		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{true, true}, []bool{true}),
			widget.GridLayoutOpts.Spacing(4, 0),
		)),
	)

	count := 5

	input := NewBindingInput(
		BindingInputOpts.Placeholder(placeholder),

		BindingInputOpts.ChangedHandler(func(args *BindingInputChangedEventArgs) {
			if args.BindingInput.takingInputs {
				if keys != nil {
					*keys = utils.AppendKeyUnique(*keys, args.Keys)
					args.BindingInput.SetText(toString(keys))
				} else {
					*buttons = utils.AppendButtonUnique(*buttons, args.Buttons)
					args.BindingInput.SetText(toString(buttons))
				}
				args.BindingInput.takingInputs = false
				count = 0
			}
		}),
	)

	if keys != nil {
		input.SetText(toString(keys))
	} else {
		input.SetText(toString(buttons))
	}

	clock := widget.NewText(

		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
				StretchVertical:    true,
				StretchHorizontal:  true,
			}),
		),

		widget.TextOpts.Text(
			strconv.Itoa(count),
			ui.res.fonts.smallFace,
			*ui.res.fgClr,
		),
		widget.TextOpts.WidgetOpts(buttonContainerMin),
		widget.TextOpts.Padding(&paddingSidesInset),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
	)

	var appendButton, cancelButton *widget.Button

	clockCountDown := func() {
		for range time.Tick(time.Second) {
			if count <= 0 {
				input.takingInputs = false
				ui.ActiveInputs--
				count = 5
				clock.Label = strconv.Itoa(count)
				ui.ui.SetFocusedWidget(appendButton)
				input.SetFocus(false)
				clock.GetWidget().SetVisibility(widget.Visibility_Hide_Blocking)
				buttonContainer.GetWidget().SetVisibility(widget.Visibility_Show)
				break
			}

			count--
			clock.Label = strconv.Itoa(count)
		}
	}

	appendButton = widget.NewButton(
		buttonText("+"),
		widget.ButtonOpts.TextPadding(&paddingSidesInset),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			input.takingInputs = true
			ui.ActiveInputs++
			ui.focus.DeFocus()
			input.SetFocus(true)
			clock.GetWidget().SetVisibility(widget.Visibility_Show)
			buttonContainer.GetWidget().SetVisibility(widget.Visibility_Hide_Blocking)
			go clockCountDown()
		}),
	)

	cancelButton = widget.NewButton(
		buttonText("x"),
		widget.ButtonOpts.TextPadding(&paddingSidesInset),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			input.SetText("")
			if keys != nil {
				*keys = []ebiten.Key{}
			} else {
				*buttons = []ebiten.StandardGamepadButton{}
			}
			input.takingInputs = false
			ui.ActiveInputs--
		}),
	)

	buttonContainer.AddChild(appendButton, cancelButton)
	sideContainer.AddChild(buttonContainer, clock)
	container.AddChild(input, sideContainer)

	clock.GetWidget().SetVisibility(widget.Visibility_Hide_Blocking)

	if buttons != nil {
		input.SetGamepadBinding(&ui.gamepadIds)
	}

	return container
}

func NewDecimalInput(ui *Ui, label string, value any, maxValue int) *widget.TextInput {
	return NewTextBoxInput(ui, BOARD_DEC, label, value, NumberValidation(maxValue))
}

func NewHexInput(ui *Ui, label string, value any, maxValue int) *widget.TextInput {
	return NewTextBoxInput(ui, BOARD_HEX, label, value, NumberValidation(maxValue))
}

func NewSaveButton(text string, f func(*widget.ButtonClickedEventArgs)) widget.PreferredSizeLocateableWidget {
	return widget.NewButton(
		widget.ButtonOpts.TextLabel(text),
		widget.ButtonOpts.ClickedHandler(f),
	)
}

func NewColorInput(ui *Ui, label string, v *color.Color, validation func(s string) (bool, *string)) widget.PreferredSizeLocateableWidget {
	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{true, true}, []bool{true, true}),
			widget.GridLayoutOpts.Spacing(8, 0),
		)),
	)

	colorBox := widget.NewContainer()
	var input *widget.TextInput

	input = widget.NewTextInput(
		widget.TextInputOpts.MobileInputMode(mobile.TEXT),
		widget.TextInputOpts.Validation(validation),
		widget.TextInputOpts.SubmitOnEnter(false),
		widget.TextInputOpts.AllowDuplicateSubmit(true),
		widget.TextInputOpts.SubmitHandler(func(*widget.TextInputChangedEventArgs) {
			ui.keyboard.Open(ui, input, BOARD_HEX, label, v, validation)
		}),
		widget.TextInputOpts.ChangedHandler(func(a *widget.TextInputChangedEventArgs) {
			*v = utils.HexToColor(a.InputText)
			colorBox.SetBackgroundImage(image.NewBorderedNineSliceColor(*v, *ui.res.fgClr, 2))
		}),
	)

	input.SetText(utils.ColorToHex(*v))

	colorBox.SetBackgroundImage(image.NewBorderedNineSliceColor(*v, *ui.res.fgClr, 2))

	container.AddChild(input, colorBox)

	return container
}

func NewApplyPalettesMenu(focusGroups *[][]widget.Focuser, pals map[string][4]string, clrInputs [4]widget.PreferredSizeLocateableWidget, res *Resources) widget.PreferredSizeLocateableWidget {
	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
		)),
	)

	focusRadio := []widget.Focuser{}

	for label, pal := range pals {
		b := widget.NewButton(
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
					MaxWidth: BUTTON_WIDTH,
					Stretch:  true,
				}),
			),

			widget.ButtonOpts.Text(
				label,
				res.fonts.smallFace,
				&widget.ButtonTextColor{
					Idle: *res.fgClr,
				},
			),

			widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
				for j := range 4 {
					c := clrInputs[j].(*widget.Container).Children()
					c[0].(*widget.TextInput).SetText(pal[j])
					clr := image.NewBorderedNineSliceColor(utils.HexToColor(pal[j]), *res.fgClr, 2)
					c[1].(*widget.Container).SetBackgroundImage(clr)
				}
			}),
		)

		focusRadio = append(focusRadio, b)
		c.AddChild(b)
	}

	*focusGroups = append(*focusGroups, focusRadio)

	return c
}

func NewFileInput(v *string) widget.PreferredSizeLocateableWidget {
	return dialogInput(v, func() string { return utils.OpenFile("Open", "All Files") })
}

func NewDirectoryInput(v *string, defaultPath string) widget.PreferredSizeLocateableWidget {
	return dialogInput(v, func() string { return utils.OpenDirectory("Choose", defaultPath) })
}

func dialogInput(v *string, dialogFunc func() string) widget.PreferredSizeLocateableWidget {
	onClick := func(input *widget.TextInput) {
		f := dialogFunc()
		*v = f
		input.SetText(trim(f, MAX_DIALOG_LEN))
		input.CursorMoveStart()
	}

	var input *widget.TextInput
	input = widget.NewTextInput(

		widget.TextInputOpts.CaretWidth(0),

		widget.TextInputOpts.MobileInputMode(mobile.TEXT),

		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.MouseButtonClickedHandler(func(*widget.WidgetMouseButtonClickedEventArgs) {
				onClick(input)
			}),
		),

		// used for button input
		widget.TextInputOpts.SubmitHandler(func(*widget.TextInputChangedEventArgs) {
			onClick(input)
		}),
	)

	input.SetText(trim(*v, MAX_DIALOG_LEN))

	return input
}

func NewRadioInput(focusRadios *[][]widget.Focuser, ptr *int, values []string, res *Resources) widget.PreferredSizeLocateableWidget {
	radio := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
		)),
	)

	bs := []widget.RadioGroupElement{}
	focusRadio := []widget.Focuser{}
	for i := range values {
		b := NewRadioButton(values[i], ptr, i, res)
		radio.AddChild(b)
		bs = append(bs, b)
		focusRadio = append(focusRadio, b)
	}

	*focusRadios = append(*focusRadios, focusRadio)

	widget.NewRadioGroup(
		widget.RadioGroupOpts.Elements(bs...),
		widget.RadioGroupOpts.InitialElement(bs[*ptr]),
	)

	return radio
}

func NewRadioButton(label string, ptr *int, value int, res *Resources) *widget.Button {
	return widget.NewButton(

		widget.ButtonOpts.Text(
			label,
			res.fonts.smallFace,
			&widget.ButtonTextColor{
				Idle: *res.fgClr,
			},
		),

		widget.ButtonOpts.TextPadding(&radioInset),

		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			*ptr = value
		}),
	)
}
