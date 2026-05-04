package ui

import (
	"github.com/aabalke/guac/utils"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/utilities/mobile"
	"github.com/ebitenui/ebitenui/widget"
	"image/color"
)

const (
	MIN_HEIGHT   = 32
	BUTTON_WIDTH = 256
)

const (
	WIDGET_HDR = iota //header
	WIDGET_CBX        //checkbox
	WIDGET_KEY        //keybinding
	WIDGET_DEC        //decimal
	WIDGET_HEX        //hexadecimal
	WIDGET_FLE        //file
	WIDGET_DIR        //directory
	WIDGET_TXT        //text
	WIDGET_RAD        //radio
)

func NewHeader(text string, res *Resources) *widget.Text {
	return widget.NewText(widget.TextOpts.Text(text, res.fonts.face, *res.fgClr))
}

func NewLabel(text string) *widget.Text {
	t := widget.NewText()
	t.Label = text
	return t
}

func NewLinkText(text string) *widget.Text {
	t := widget.NewText(
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

	t.Label = text

	return t
}

func NewSeparator() *widget.Container {
	return widget.NewContainer()
}

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

		widget.TextInputOpts.SubmitHandler(func(args *widget.TextInputChangedEventArgs) {
			ui.keyboard.Open(ui, input, board, label, value)
		}),
		widget.TextInputOpts.ChangedHandler(func(args *widget.TextInputChangedEventArgs) {
			fromString(value, args.InputText)
		}),
	)

	input.SetText(toString(value))

	return input
}

func NewKeybindInput(ui *Ui, label string, value any) *widget.TextInput {
	return NewTextBoxInput(ui, BOARD_KEYBIND, label, value, NoValidation())
}

func NewDecimalInput(ui *Ui, label string, value any, maxValue int) *widget.TextInput {
	return NewTextBoxInput(ui, BOARD_DEC, label, value, NumberValidation(maxValue))
}

func NewHexInput(ui *Ui, label string, value any, maxValue int) *widget.TextInput {
	return NewTextBoxInput(ui, BOARD_HEX, label, value, NumberValidation(maxValue))
}

func NewTextInput(ui *Ui, label string, value any) *widget.TextInput {
	return NewTextBoxInput(ui, BOARD_ALPHA, label, value, NoValidation())
}

func NewSaveButton(text string, f func(args *widget.ButtonClickedEventArgs)) widget.PreferredSizeLocateableWidget {
	b := widget.NewButton(widget.ButtonOpts.ClickedHandler(f))
	b.SetText(text)
	return b
}

func NewColorInput(ui *Ui, label string, v *color.Color, validation func(s string) (bool, *string)) widget.PreferredSizeLocateableWidget {

	colorBox := widget.NewContainer()

	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{true, true}, []bool{true, true}),
			widget.GridLayoutOpts.Spacing(8, 0),
		)),
	)

	var input *widget.TextInput

	input = widget.NewTextInput(
		widget.TextInputOpts.MobileInputMode(mobile.TEXT),
		widget.TextInputOpts.Validation(validation),
		widget.TextInputOpts.SubmitOnEnter(false),
		widget.TextInputOpts.AllowDuplicateSubmit(true),
		widget.TextInputOpts.SubmitHandler(func(*widget.TextInputChangedEventArgs) {
			ui.keyboard.Open(ui, input, BOARD_HEX, label, v)
		}),
		widget.TextInputOpts.ChangedHandler(func(a *widget.TextInputChangedEventArgs) {
			*v = utils.HexToColor(a.InputText)
			colorBox.SetBackgroundImage(image.NewNineSliceColor(*v))
		}),
	)

	colorBox.SetBackgroundImage(image.NewNineSliceColor(*v))

	input.SetText(utils.ColorToHex(*v))

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
					c[0].(*widget.TextInput).SetText("#" + pal[j])
					clr := image.NewNineSliceColor(utils.HexToColor(pal[j]))
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

func dialogInput(v *string, dialogFunc func() string) widget.PreferredSizeLocateableWidget {

	var input *widget.TextInput

	onClick := func(input *widget.TextInput) {
		f := dialogFunc()
		*v = f
		input.SetText(trim(f, MAX_DIALOG_LEN))
		input.CursorMoveStart()
	}

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

func NewFileInput(v *string) widget.PreferredSizeLocateableWidget {
	return dialogInput(v, func() string { return utils.OpenFile("Open", "All Files") })
}

func NewDirectoryInput(v *string, defaultPath string) widget.PreferredSizeLocateableWidget {
	return dialogInput(v, func() string { return utils.OpenDirectory("Choose", defaultPath) })
}

func NewRadioInput(focusRadios *[][]widget.Focuser, v *int, values []string, res *Resources) widget.PreferredSizeLocateableWidget {

	radio := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
		)),
	)

	bs := []widget.RadioGroupElement{}
	focusRadio := []widget.Focuser{}
	for i := range values {
		b := NewRadioButton(v, values[i], i, res)
		radio.AddChild(b)
		bs = append(bs, b)
		focusRadio = append(focusRadio, b)
	}

	*focusRadios = append(*focusRadios, focusRadio)

	widget.NewRadioGroup(
		widget.RadioGroupOpts.Elements(bs...),
		widget.RadioGroupOpts.InitialElement(bs[*v]),
	)

	return radio
}

func NewRadioButton(v *int, label string, value int, res *Resources) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
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

		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   16,
			Right:  16,
			Top:    4,
			Bottom: 4,
		}),

		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			*v = value
		}),
	)
}

func NewRadioStringInput(focusRadios *[][]widget.Focuser, v *string, values []string, res *Resources) widget.PreferredSizeLocateableWidget {

	radio := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
		)),
	)

	bs := []widget.RadioGroupElement{}
	focusRadio := []widget.Focuser{}
	init := 0
	for i, value := range values {
		b := NewRadioStringButton(v, values[i], value, res)
		radio.AddChild(b)
		bs = append(bs, b)
		focusRadio = append(focusRadio, b)

		if value == *v {
			init = i
		}
	}

	*focusRadios = append(*focusRadios, focusRadio)

	widget.NewRadioGroup(
		widget.RadioGroupOpts.Elements(bs...),
		widget.RadioGroupOpts.InitialElement(bs[init]),
	)

	return radio
}

func NewRadioStringButton(v *string, label string, value string, res *Resources) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
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

		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   16,
			Right:  16,
			Top:    4,
			Bottom: 4,
		}),

		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			*v = value
		}),
	)
}
