package ui

import (
	img "image"
	"image/color"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/utilities/mobile"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type BindingInputParams struct {
	Image   *BindingInputImage
	Color   *BindingInputColor
	Padding *widget.Insets
	Face    *text.Face
}

type BindingInput struct {
	definedParams  BindingInputParams
	computedParams BindingInputParams

	takingInputs bool
	dirty        bool
	inputText    string

	isButtonInput bool
	gamepadIds    *map[ebiten.GamepadID]struct{}
	inputKeys     []ebiten.Key
	inputButtons  []ebiten.StandardGamepadButton
	// inputMouse   []ebiten.MouseButton

	placeholderText string
	mobileInputMode mobile.InputMode

	ChangedEvent *event.Event

	widgetOpts []widget.WidgetOpt
	init       *widget.MultiOnce
	widget     *widget.Widget
	text       *widget.Text
	renderBuf  *image.MaskedRenderBuffer
	mask       *image.NineSlice
	state      bindingInputState

	tabOrder int
	focused  bool
	focusMap map[widget.FocusDirection]widget.Focuser
}

type BindingInputOpt func(t *BindingInput)

type BindingInputOptions struct{}

type BindingInputImage struct {
	Idle  *image.NineSlice
	Focus *image.NineSlice
}

type BindingInputColor struct {
	Idle  color.Color
	Focus color.Color
}

type BindingInputChangedEventArgs struct {
	BindingInput *BindingInput
	InputText    string
	Keys         []ebiten.Key
	Buttons      []ebiten.StandardGamepadButton
	OffsetX      int
	OffsetY      int
}

type BindingInputChangedHandlerFunc func(args *BindingInputChangedEventArgs)

type bindingInputState func() (bindingInputState, bool)

var BindingInputOpts BindingInputOptions

func NewBindingInput(opts ...BindingInputOpt) *BindingInput {
	t := &BindingInput{
		ChangedEvent:    &event.Event{},
		init:            &widget.MultiOnce{},
		renderBuf:       image.NewMaskedRenderBuffer(),
		mobileInputMode: mobile.TEXT,
		focusMap:        make(map[widget.FocusDirection]widget.Focuser),
	}
	t.state = t.idleState()
	t.init.Append(t.createWidget)

	for _, o := range opts {
		o(t)
	}

	return t
}

func (t *BindingInput) SetGamepadBinding(gamepadIds *map[ebiten.GamepadID]struct{}) {
	t.isButtonInput = true
	t.gamepadIds = gamepadIds
}

func (t *BindingInput) createWidget() {
	t.widget = widget.NewWidget(append([]widget.WidgetOpt{
		widget.WidgetOpts.TrackHover(true),
	}, t.widgetOpts...)...)
	t.widget.SetFocusable(t)
	t.mask = image.NewNineSliceColor(color.NRGBA{255, 0, 255, 255})
}

func (t *BindingInput) Validate() {
	t.init.Do()
	t.populateComputedParams()

	if t.computedParams.Face == nil {
		panic("BindingInput: Font Face is required.")
	}
	if t.computedParams.Color == nil {
		panic("BindingInput: Color is required.")
	}
	if t.computedParams.Color.Idle == nil {
		panic("BindingInput: Color.Idle is required.")
	}

	t.text = widget.NewText(widget.TextOpts.Text("", t.computedParams.Face, color.White))
	t.text.Validate()
}

func (t *BindingInput) populateComputedParams() {
	params := BindingInputParams{Color: &BindingInputColor{}, Image: &BindingInputImage{}}

	if theme := t.GetWidget().GetTheme(); theme != nil {
		if theme.TextInputTheme != nil {
			if theme.TextInputTheme.Color != nil {
				params.Color.Idle = theme.TextInputTheme.Color.Idle
				params.Color.Focus = theme.TextInputTheme.Color.Hover
			}
			if theme.TextInputTheme.Face != nil {
				params.Face = theme.TextInputTheme.Face
			} else {
				params.Face = theme.DefaultFace
			}
			if theme.TextInputTheme.Image != nil {
				params.Image.Idle = theme.TextInputTheme.Image.Idle
				params.Image.Focus = theme.TextInputTheme.Image.Hover
			}
			params.Padding = theme.TextInputTheme.Padding
		}
	}

	if t.definedParams.Color != nil {
		params.Color.Idle = t.definedParams.Color.Idle
		params.Color.Focus = t.definedParams.Color.Focus
	}
	if t.definedParams.Face != nil {
		params.Face = t.definedParams.Face
	}
	if t.definedParams.Image != nil {
		params.Image.Idle = t.definedParams.Image.Idle
		params.Image.Focus = t.definedParams.Image.Focus
	}
	if t.definedParams.Padding != nil {
		params.Padding = t.definedParams.Padding
	}

	// Set Default values
	if params.Image == nil {
		params.Image = &BindingInputImage{}
	}
	if params.Padding == nil {
		params.Padding = &widget.Insets{}
	}

	t.computedParams = params
}

func (o BindingInputOptions) WidgetOpts(opts ...widget.WidgetOpt) BindingInputOpt {
	return func(t *BindingInput) {
		t.widgetOpts = append(t.widgetOpts, opts...)
	}
}

func (o BindingInputOptions) ChangedHandler(f BindingInputChangedHandlerFunc) BindingInputOpt {
	return func(t *BindingInput) {
		t.ChangedEvent.AddHandler(func(args any) {
			if arg, ok := args.(*BindingInputChangedEventArgs); ok {
				f(arg)
			}
		})
	}
}

func (o BindingInputOptions) Image(i *BindingInputImage) BindingInputOpt {
	return func(t *BindingInput) {
		t.definedParams.Image = i
	}
}

func (o BindingInputOptions) Color(c *BindingInputColor) BindingInputOpt {
	return func(t *BindingInput) {
		t.definedParams.Color = c
	}
}

func (o BindingInputOptions) Padding(i *widget.Insets) BindingInputOpt {
	return func(t *BindingInput) {
		t.definedParams.Padding = i
	}
}

func (o BindingInputOptions) Face(f *text.Face) BindingInputOpt {
	return func(t *BindingInput) {
		t.definedParams.Face = f
	}
}

func (o BindingInputOptions) Placeholder(s string) BindingInputOpt {
	return func(t *BindingInput) {
		t.placeholderText = s
	}
}

func (o BindingInputOptions) TabOrder(to int) BindingInputOpt {
	return func(t *BindingInput) {
		t.tabOrder = to
	}
}

// https://css-tricks.com/everything-you-ever-wanted-to-know-about-inputmode
func (o BindingInputOptions) MobileInputMode(mobileInputMode mobile.InputMode) BindingInputOpt {
	return func(t *BindingInput) {
		t.mobileInputMode = mobileInputMode
	}
}

func (t *BindingInput) GetWidget() *widget.Widget {
	t.init.Do()
	return t.widget
}

func (t *BindingInput) SetLocation(rect img.Rectangle) {
	t.init.Do()
	t.widget.Rect = rect
}

func (t *BindingInput) PreferredSize() (int, int) {
	t.init.Do()

	h := t.computedParams.Padding.Top + t.computedParams.Padding.Bottom
	if t.widget != nil && h < t.widget.MinHeight {
		h = t.widget.MinHeight
	}

	w := 50
	if t.widget != nil && w < t.widget.MinWidth {
		w = t.widget.MinWidth
	}

	h = max(1, h)

	return w, h
}

func (t *BindingInput) Update(updObj *widget.UpdateObject) {
	t.init.Do()

	for {
		newState, rerun := t.state()
		if newState != nil {
			t.state = newState
		}
		if !rerun {
			break
		}
	}

	if t.dirty {
		t.ChangedEvent.Fire(&BindingInputChangedEventArgs{
			InputText:    t.inputText,
			Keys:         t.inputKeys,
			Buttons:      t.inputButtons,
			BindingInput: t,
		})
	}

	t.dirty = false

	t.widget.Update(updObj)
	if t.text != nil {
		t.text.Update(updObj)
	}
}

func (t *BindingInput) idleState() bindingInputState {
	return func() (bindingInputState, bool) {
		if !t.focused {
			return t.idleState(), false
		}

		if t.isButtonInput {

			buttons := []ebiten.StandardGamepadButton{}

			for id := range *t.gamepadIds {
				if !ebiten.IsStandardGamepadLayoutAvailable(id) {
					continue
				}

				buttons = inpututil.AppendJustPressedStandardGamepadButtons(id, buttons)
			}

			if len(buttons) != 0 {
				t.init.Do()
				if t.takingInputs {
					t.inputButtons = buttons
					t.inputText = toString(&buttons)
					t.dirty = true
				}
			}

		} else {

			keys := inpututil.AppendJustPressedKeys([]ebiten.Key{})

			if len(keys) != 0 {
				t.init.Do()
				if t.takingInputs {
					t.inputKeys = keys
					t.inputText = toString(&keys)
					t.dirty = true
				}
			}
		}

		return t.idleState(), false
	}
}

func (t *BindingInput) Render(screen *ebiten.Image) {
	t.init.Do()
	t.widget.Render(screen)
	t.renderImage(screen)
	t.renderText(screen)
}

func (t *BindingInput) renderImage(screen *ebiten.Image) {
	if t.computedParams.Image == nil || t.computedParams.Image.Idle == nil {
		return
	}

	i := t.computedParams.Image.Idle
	if t.focused && t.computedParams.Image.Focus != nil {
		i = t.computedParams.Image.Focus
	}

	rect := t.widget.Rect
	i.Draw(screen, rect.Dx(), rect.Dy(), func(opts *ebiten.DrawImageOptions) {
		opts.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
	})
}

func (t *BindingInput) renderText(screen *ebiten.Image) {
	t.renderBuf.DrawTextInput(
		screen,
		func(buf *ebiten.Image) {
			t.drawText(buf)
		},
		func(buf *ebiten.Image) {
			rect := t.widget.Rect
			t.mask.Draw(buf, rect.Dx()-t.computedParams.Padding.Left-t.computedParams.Padding.Right, rect.Dy()-t.computedParams.Padding.Top-t.computedParams.Padding.Bottom,
				func(opts *ebiten.DrawImageOptions) {
					opts.GeoM.Translate(float64(t.computedParams.Padding.Left), float64(t.computedParams.Padding.Top))
					opts.Blend = ebiten.BlendCopy
				})
		},
		t.widget.Rect,
	)
}

func (t *BindingInput) drawText(screen *ebiten.Image) {
	rect := t.widget.Rect
	rect.Max.X -= rect.Min.X
	rect.Min.X = 0
	rect.Max.Y -= rect.Min.Y
	rect.Min.Y = 0

	tr := rect
	tr = tr.Add(img.Point{t.computedParams.Padding.Left, t.computedParams.Padding.Top})

	t.text.SetLocation(tr)
	if len([]rune(t.inputText)) > 0 {
		t.text.Label = t.inputText
	} else {
		t.text.Label = t.placeholderText
	}

	t.text.SetColor(t.computedParams.Color.Idle)
	t.text.Render(screen)
}

func (t *BindingInput) GetText() string {
	return t.inputText
}

func (t *BindingInput) SetText(text string) {
	t.init.Do()
	t.inputText = text
}

// Focus this is used by ebitenui to focus the element - however it cannot be focused
// directly - only through "append" button selection
func (t *BindingInput) Focus(focused bool) {}

func (t *BindingInput) FocusExplicit(focused bool) {
	t.init.Do()
	t.GetWidget().FireFocusEvent(t, focused, img.Point{-1, -1})
	t.focused = focused
}

func (t *BindingInput) IsFocused() bool {
	return t.focused
}

func (t *BindingInput) FocusClearAll() {
	t.focusMap = map[widget.FocusDirection]widget.Focuser{}
}

func (t *BindingInput) TabOrder() int {
	return t.tabOrder
}

func (t *BindingInput) GetFocus(direction widget.FocusDirection) widget.Focuser {
	return t.focusMap[direction]
}

func (t *BindingInput) AddFocus(direction widget.FocusDirection, focus widget.Focuser) {
	t.focusMap[direction] = focus
}
