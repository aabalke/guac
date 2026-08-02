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

const (
	WIDGET_HDR = iota // header
	WIDGET_CBX        // checkbox
	WIDGET_KEY        // keybinding
	WIDGET_DEC        // decimal
	WIDGET_HEX        // hexadecimal
	WIDGET_FLE        // file
	WIDGET_DIR        // directory
	WIDGET_TXT        // text
	WIDGET_LNK        // link
	WIDGET_RAD        // radio
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

		widget.TextInputOpts.SubmitHandler(func(*widget.TextInputChangedEventArgs) {
			ui.keyboard.Open(ui, input, board, label, value, validation)
		}),
		widget.TextInputOpts.ChangedHandler(func(args *widget.TextInputChangedEventArgs) {
			fromString(value, args.InputText)
		}),
	)

	input.SetText(toString(value))

	return input
}

func NewKeybindInput(ui *Ui, v *[]ebiten.Key) widget.PreferredSizeLocateableWidget {
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
			widget.GridLayoutOpts.Columns(3),
			widget.GridLayoutOpts.Stretch([]bool{true, false, false}, []bool{true}),
			widget.GridLayoutOpts.Spacing(4, 0),
		)),
	)

	buttonContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, false}, []bool{true}),
			widget.GridLayoutOpts.Spacing(4, 0),
		)),

		widget.ContainerOpts.WidgetOpts(buttonContainerMin),
	)

	count := 5

	input := NewBindingInput(
		BindingInputOpts.Placeholder("Click to add binding"),

		BindingInputOpts.ChangedHandler(func(args *BindingInputChangedEventArgs) {
			if args.TextInput.takingInputs {
				*v = utils.AppendKeyUnique(*v, args.Keys)
				args.TextInput.SetText(toString(v))
				args.TextInput.takingInputs = false
				count = 0
			}
		}),
	)

	input.SetText(toString(v))

	clock := widget.NewButton(
		buttonText(strconv.Itoa(count)),
		widget.ButtonOpts.WidgetOpts(buttonContainerMin),
		widget.ButtonOpts.Image(&transparentButtonImage),
		widget.ButtonOpts.TextPadding(&paddingSidesInset),
	)

	clockCountDown := func() {
		for range time.Tick(time.Second) {
			if count <= 0 {
				container.RemoveChildren()
				container.AddChild(input, buttonContainer)
				input.takingInputs = false
				count = 5
				clock.SetText(strconv.Itoa(count))
				break
			}

			count--
			clock.SetText(strconv.Itoa(count))

		}
	}

	buttonContainer.AddChild(widget.NewButton(
		buttonText("+"),
		widget.ButtonOpts.TextPadding(&paddingSidesInset),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			input.takingInputs = true
			ui.ui.SetFocusedWidget(input)
			container.RemoveChildren()
			container.AddChild(input, clock)
			go clockCountDown()
		}),
	))

	buttonContainer.AddChild(widget.NewButton(
		buttonText("x"),
		widget.ButtonOpts.TextPadding(&paddingSidesInset),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			input.SetText("")
			*v = []ebiten.Key{}
			input.takingInputs = false
		}),
	))

	container.AddChild(input, buttonContainer)

	return container
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
