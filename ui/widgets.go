package ui

import (
	"image/color"
	"strconv"
	"strings"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/utils"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/utilities/mobile"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	MIN_HEIGHT   = 32
	BUTTON_WIDTH = 256
)

func NewLabel(text string, face *text.Face, clr color.Color) *widget.Text {
	return widget.NewText(
		widget.TextOpts.Text(text, face, clr),
		widget.TextOpts.WidgetOpts(),
	)
}

func NewLinkText(text string, face *text.Face, clr color.Color) *widget.Text {
	return widget.NewText(
		widget.TextOpts.LinkClickedHandler(func(args *widget.LinkEventArgs) {
			utils.OpenLink(args.Id)
		}),
		widget.TextOpts.ProcessBBCode(true),

		widget.TextOpts.Text(text, face, clr),
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

func NewSeparator() *widget.Container {
	return widget.NewContainer()
}

func NewCheckbox(value *bool, img *widget.CheckboxImage) widget.PreferredSizeLocateableWidget {

	state := widget.WidgetUnchecked
	if *value {
		state = widget.WidgetChecked
	}

	return widget.NewCheckbox(
		widget.CheckboxOpts.Image(img),

		widget.CheckboxOpts.StateChangedHandler(func(args *widget.CheckboxChangedEventArgs) {
			*value = args.State == widget.WidgetChecked
		}),

		widget.CheckboxOpts.InitialState(state),
	)
}

func NewTextBoxInput(value any, face *text.Face, validation func(s string) (bool, *string)) widget.PreferredSizeLocateableWidget {

	clr := config.Conf.Ui.MenuForegroundColor
	input := widget.NewTextInput(

		widget.TextInputOpts.MobileInputMode(mobile.TEXT),

		widget.TextInputOpts.Image(&widget.TextInputImage{
			Idle: image.NewBorderedNineSliceColor(color.Transparent, clr, 2),
		}),

		widget.TextInputOpts.Face(face),

		widget.TextInputOpts.Color(&widget.TextInputColor{
			Idle:  clr,
			Caret: clr,
		}),

		widget.TextInputOpts.Padding(&paddingInset),

		widget.TextInputOpts.Validation(validation),

		widget.TextInputOpts.ChangedHandler(func(args *widget.TextInputChangedEventArgs) {
			// when "changed"

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

		combined := ""

		for i, s := range *v {
			if i > 0 && i < len(*v) {
				combined += ", "
			}
			combined += s
		}

		input.SetText(combined)

	case *[]int:

		combined := ""

		for i, s := range *v {
			if i > 0 && i < len(*v) {
				combined += ", "
			}
			combined += strconv.Itoa(s)
		}

		input.SetText(combined)

	case *[]ebiten.StandardGamepadButton:

		combined := ""
		for i, s := range *v {
			if i > 0 && i < len(*v) {
				combined += ", "
			}
			combined += utils.GamepadButtonToString(s)
		}

		input.SetText(combined)

	default:
		panic("not supported text box input")
	}

	return input
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

func NewSaveButton(face *text.Face, img *widget.ButtonImage, f func(args *widget.ButtonClickedEventArgs)) widget.PreferredSizeLocateableWidget {
	clr := config.Conf.Ui.MenuForegroundColor

	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),

		widget.ButtonOpts.Image(img),

		widget.ButtonOpts.Text(
			"save",
			face,
			&widget.ButtonTextColor{
				Idle: clr,
			},
		),

		widget.ButtonOpts.TextPadding(&buttonInset),

		widget.ButtonOpts.ClickedHandler(f),
	)
}

func NewColorInput(v *color.Color, face *text.Face, validation func(s string) (bool, *string)) widget.PreferredSizeLocateableWidget {

	clr := config.Conf.Ui.MenuForegroundColor

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

		widget.TextInputOpts.Image(&widget.TextInputImage{
			Idle: image.NewBorderedNineSliceColor(color.Transparent, clr, 2),
		}),

		widget.TextInputOpts.Face(face),

		widget.TextInputOpts.Color(&widget.TextInputColor{
			Idle:  clr,
			Caret: clr,
		}),

		widget.TextInputOpts.Padding(&paddingInset),

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

func NewApplyPalettesMenu(focusGroups *[][]widget.Focuser, pals map[string][4]string, clrInputs [4]widget.PreferredSizeLocateableWidget, face *text.Face, img *widget.ButtonImage) widget.PreferredSizeLocateableWidget {
	clr := config.Conf.Ui.MenuForegroundColor

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

			widget.ButtonOpts.Image(img),

			widget.ButtonOpts.Text(
				label,
				face,
				&widget.ButtonTextColor{
					Idle: clr,
				},
			),

			widget.ButtonOpts.TextPadding(&buttonInset),

			widget.ButtonOpts.TextPosition(
				widget.TextPositionCenter,
				widget.TextPositionCenter,
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

func NewFileInput(v *string, face *text.Face) widget.PreferredSizeLocateableWidget {

	clr := config.Conf.Ui.MenuForegroundColor

	var input *widget.TextInput

	onClick := func() {
		f := utils.OpenFile("Open", "All Files")

		*v = f

		if max := 24; len(f) >= max {
			f = "..." + f[len(f)-max-3:]
		}

		input.SetText(f)
		input.CursorMoveStart()
	}

	input = widget.NewTextInput(

		widget.TextInputOpts.MobileInputMode(mobile.TEXT),

		widget.TextInputOpts.Image(&widget.TextInputImage{
			Idle: image.NewBorderedNineSliceColor(color.Transparent, clr, 2),
		}),

		widget.TextInputOpts.Face(face),

		widget.TextInputOpts.Color(&widget.TextInputColor{
			Idle:  clr,
			Caret: clr,
		}),

		widget.TextInputOpts.Padding(&paddingInset),

		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.MouseButtonClickedHandler(func(args *widget.WidgetMouseButtonClickedEventArgs) {
				onClick()
			}),
		),

		// used for button input
		widget.TextInputOpts.SubmitHandler(func(args *widget.TextInputChangedEventArgs) {
			onClick()
		}),

		widget.TextInputOpts.CaretWidth(0),
	)

	input.SetText(*v)

	return input
}

func NewDirectoryInput(v *string, face *text.Face, defaultPath string) widget.PreferredSizeLocateableWidget {

	clr := config.Conf.Ui.MenuForegroundColor

	var input *widget.TextInput

	onClick := func(args *widget.WidgetMouseButtonClickedEventArgs) {
		f := utils.OpenDirectory("Choose", defaultPath)

		*v = f

		if max := 12; len(f) >= max {
			f = "..." + f[len(f)-max-3:]
		}

		input.SetText(f)
		input.CursorMoveStart()
	}

	input = widget.NewTextInput(

		widget.TextInputOpts.MobileInputMode(mobile.TEXT),

		widget.TextInputOpts.Image(&widget.TextInputImage{
			Idle: image.NewBorderedNineSliceColor(color.Transparent, clr, 2),
		}),

		widget.TextInputOpts.Face(face),

		widget.TextInputOpts.Color(&widget.TextInputColor{
			Idle:  clr,
			Caret: clr,
		}),

		widget.TextInputOpts.Padding(&paddingInset),

		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.MouseButtonClickedHandler(onClick),
		),

		widget.TextInputOpts.CaretWidth(0),
	)

	input.SetText(*v)

	return input
}

func NewRadioInput(focusRadios *[][]widget.Focuser, v *int, values []string, face *text.Face, img *widget.ButtonImage) widget.PreferredSizeLocateableWidget {

	radio := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
		)),
	)

	bs := []widget.RadioGroupElement{}

	focusRadio := []widget.Focuser{}

	for i := range values {
		b := NewRadioButton(v, values[i], i, face, img)
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

func NewRadioButton(v *int, label string, value int, face *text.Face, img *widget.ButtonImage) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				MaxWidth: BUTTON_WIDTH,
				Stretch:  true,
			}),
		),

		widget.ButtonOpts.Image(img),

		widget.ButtonOpts.Text(
			label,
			face,
			&widget.ButtonTextColor{
				Idle: config.Conf.Ui.MenuForegroundColor,
			},
		),

		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   16,
			Right:  16,
			Top:    4,
			Bottom: 4,
		}),

		widget.ButtonOpts.TextPosition(
			widget.TextPositionStart,
			widget.TextPositionCenter,
		),

		widget.ButtonOpts.ClickedHandler(
			func(*widget.ButtonClickedEventArgs) {
				*v = value
			},
		),
	)
}
