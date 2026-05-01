package ui

import (
	"image/color"
	"strconv"
	"strings"

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
	)
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

func NewTextBoxInput(value any, validation func(s string) (bool, *string)) widget.PreferredSizeLocateableWidget {

	input := widget.NewTextInput(
		widget.TextInputOpts.MobileInputMode(mobile.TEXT),
		widget.TextInputOpts.Validation(validation),
		widget.TextInputOpts.ChangedHandler(func(args *widget.TextInputChangedEventArgs) {
			switch v := value.(type) {
			case *int:
				*v, _ = strconv.Atoi(args.InputText)
			case *string:
				*v = args.InputText
			case *[]string:
				*v = strings.Split(strings.ReplaceAll(args.InputText, " ", ""), ",")
			case *[]int:
				a := strings.Split(strings.ReplaceAll(args.InputText, " ", ""), ",")
				nums := []int{}

				for _, num := range a {
					n, _ := strconv.Atoi(num)
					nums = append(nums, n)
				}

				*v = nums

			case *[]ebiten.StandardGamepadButton:
				strs := strings.Split(strings.ReplaceAll(args.InputText, " ", ""), ",")

				*v = []ebiten.StandardGamepadButton{}
				for i := range strs {
					*v = append(*v, utils.StringToGamepadButton(strs[i]))
				}

			default:
				panic("not supported text box input")
			}
		}),
	)

	switch v := value.(type) {
	case *int:
		input.SetText(strconv.Itoa(*v))
	case *string:
		input.SetText(*v)
	case *[]string:
		input.SetText(join(*v, ", ", func(s string) string { return s }))

	case *[]int:
		input.SetText(join(*v, ", ", strconv.Itoa))

	case *[]ebiten.StandardGamepadButton:
		input.SetText(join(*v, ", ", utils.GamepadButtonToString))

	default:
		panic("not supported text box input")
	}

	return input
}

func join[T any](vals []T, sep string, f func(T) string) string {
	out := make([]string, len(vals))
	for i, v := range vals {
		out[i] = f(v)
	}
	return strings.Join(out, sep)
}

func NumberValidation(maxValue int) func(string) (bool, *string) {
	return func(s string) (bool, *string) {

		var b strings.Builder
		for _, r := range s {
			if r >= '0' && r <= '9' {
				b.WriteRune(r)
			}
		}

		digits := b.String()

		v, _ := strconv.Atoi(digits)

		if v >= maxValue {
			digits = strconv.Itoa(maxValue)
			return false, &digits
		}

		return false, &digits
	}
}

func NoValidation() func(string) (bool, *string) {
	return func(s string) (bool, *string) {
		return true, &s
	}
}

func NewSaveButton(f func(args *widget.ButtonClickedEventArgs)) widget.PreferredSizeLocateableWidget {

	b := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),

		widget.ButtonOpts.ClickedHandler(f),
	)

	b.SetText("save")

	return b
}

func NewColorInput(v *color.Color, validation func(s string) (bool, *string)) widget.PreferredSizeLocateableWidget {

	colorBox := widget.NewContainer()

	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{true, true}, []bool{true, true}),
			widget.GridLayoutOpts.Spacing(10, 0),
		)),
	)

	input := widget.NewTextInput(
		widget.TextInputOpts.MobileInputMode(mobile.TEXT),
		widget.TextInputOpts.Validation(validation),
		widget.TextInputOpts.ChangedHandler(func(args *widget.TextInputChangedEventArgs) {
			*v = utils.HexToColor(args.InputText)
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

const MAX_DIALOG_LEN = 24

func trim(s string, max int) string {
	r := []rune(s)

	if len(r) <= max {
		return s
	}

	return "..." + string(r[len(r)-(max-len([]rune("..."))):])
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
