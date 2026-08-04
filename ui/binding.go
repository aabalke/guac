package ui

import (
	img "image"
	"image/color"
	"math"
	"time"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/utilities/mobile"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type TextInputParams struct {
	Image                *widget.TextInputImage
	Color                *widget.TextInputColor
	Padding              *widget.Insets
	Face                 *text.Face
	RepeatDelay          *time.Duration
	RepeatInterval       *time.Duration
	Secure               *bool
	ClearOnSubmit        *bool
	IgnoreEmptySubmit    *bool
	AllowDuplicateSubmit *bool
	SubmitOnEnter        *bool
}

type BindingInput struct {
	definedParams  TextInputParams
	computedParams TextInputParams

	ChangedEvent *event.Event
	SubmitEvent  *event.Event

	takingInputs bool
	dirty        bool
	inputText    string
	inputKeys    []ebiten.Key

	validationFunc  TextInputValidationFunc
	placeholderText string
	mobileInputMode mobile.InputMode

	PressedEvent       *event.Event
	ReleasedEvent      *event.Event
	ClickedEvent       *event.Event
	CursorEnteredEvent *event.Event
	CursorMovedEvent   *event.Event
	CursorExitedEvent  *event.Event
	StateChangedEvent  *event.Event

	widgetOpts     []widget.WidgetOpt
	init           *widget.MultiOnce
	widget         *widget.Widget
	text           *widget.Text
	renderBuf      *image.MaskedRenderBuffer
	mask           *image.NineSlice
	cursorPosition int
	state          textInputState
	scrollOffset   int

	hovering bool
	pressing bool

	tabOrder int
	focused  bool
	focusMap map[widget.FocusDirection]widget.Focuser
}

type TextInputOpt func(t *BindingInput)

type TextInputOptions struct{}

type TextInputImage struct {
	Idle     *image.NineSlice
	Disabled *image.NineSlice
	// Highlight defaults to image.NewNineSliceColor(color.NRGBA{6, 67, 161, 100}).
	Highlight *image.NineSlice

	Hover           *image.NineSlice
	Pressed         *image.NineSlice
	PressedHover    *image.NineSlice
	PressedDisabled *image.NineSlice
}

type TextInputColor struct {
	Idle     color.Color
	Disabled color.Color
	Hover    color.Color
	Pressed  color.Color
}

type TextInputPressedEventArgs struct {
	TextInput *BindingInput
	OffsetX   int
	OffsetY   int
}

type TextInputReleasedEventArgs struct {
	TextInput *BindingInput
	Inside    bool
	OffsetX   int
	OffsetY   int
}

type TextInputClickedEventArgs struct {
	TextInput *BindingInput
	OffsetX   int
	OffsetY   int
}

type TextInputHoverEventArgs struct {
	TextInput *BindingInput
	Entered   bool
	OffsetX   int
	OffsetY   int
	DiffX     int
	DiffY     int
}

type BindingInputChangedEventArgs struct {
	TextInput *BindingInput
	InputText string
	Keys      []ebiten.Key
	OffsetX   int
	OffsetY   int
}

type TextInputPressedHandlerFunc func(args *TextInputPressedEventArgs)

type TextInputReleasedHandlerFunc func(args *TextInputReleasedEventArgs)

type TextInputClickedHandlerFunc func(args *TextInputClickedEventArgs)

type TextInputCursorHoverHandlerFunc func(args *TextInputHoverEventArgs)

type TextInputChangedHandlerFunc func(args *BindingInputChangedEventArgs)

type TextInputValidationFunc func(newInputText string) (bool, *string)

type textInputState func() (textInputState, bool)

var BindingInputOpts TextInputOptions

func NewBindingInput(opts ...TextInputOpt) *BindingInput {
	t := &BindingInput{
		PressedEvent:       &event.Event{},
		ReleasedEvent:      &event.Event{},
		ClickedEvent:       &event.Event{},
		CursorEnteredEvent: &event.Event{},
		CursorMovedEvent:   &event.Event{},
		CursorExitedEvent:  &event.Event{},
		StateChangedEvent:  &event.Event{},

		ChangedEvent: &event.Event{},
		SubmitEvent:  &event.Event{},

		init:      &widget.MultiOnce{},
		renderBuf: image.NewMaskedRenderBuffer(),

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

func (t *BindingInput) Validate() {
	t.init.Do()
	t.populateComputedParams()

	if t.computedParams.Face == nil {
		panic("TextInput: Font Face is required.")
	}
	if t.computedParams.Color == nil {
		panic("TextInput: Color is required.")
	}
	if t.computedParams.Color.Idle == nil {
		panic("TextInput: Color.Idle is required.")
	}

	t.initWidget()
}

func (t *BindingInput) populateComputedParams() {
	TRUE := true
	FALSE := false
	params := TextInputParams{Color: &widget.TextInputColor{}, Image: &widget.TextInputImage{}}

	theme := t.GetWidget().GetTheme()

	// Set theme values
	if theme != nil {
		if theme.TextInputTheme != nil {
			params.AllowDuplicateSubmit = theme.TextInputTheme.AllowDuplicateSubmit
			params.ClearOnSubmit = theme.TextInputTheme.ClearOnSubmit
			if theme.TextInputTheme.Color != nil {
				params.Color.Idle = theme.TextInputTheme.Color.Idle
				params.Color.Disabled = theme.TextInputTheme.Color.Disabled
				params.Color.Caret = theme.TextInputTheme.Color.Caret
				params.Color.DisabledCaret = theme.TextInputTheme.Color.DisabledCaret
				params.Color.Hover = theme.TextInputTheme.Color.Hover
				params.Color.Pressed = theme.TextInputTheme.Color.Pressed
			}
			if theme.TextInputTheme.Face != nil {
				params.Face = theme.TextInputTheme.Face
			} else {
				params.Face = theme.DefaultFace
			}
			params.IgnoreEmptySubmit = theme.TextInputTheme.IgnoreEmptySubmit
			if theme.TextInputTheme.Image != nil {
				params.Image.Idle = theme.TextInputTheme.Image.Idle
				params.Image.Disabled = theme.TextInputTheme.Image.Disabled
				params.Image.Highlight = theme.TextInputTheme.Image.Highlight
				params.Image.Hover = theme.TextInputTheme.Image.Hover
				params.Image.Pressed = theme.TextInputTheme.Image.Pressed
				params.Image.PressedDisabled = theme.TextInputTheme.Image.PressedDisabled
				params.Image.PressedHover = theme.TextInputTheme.Image.PressedHover
			}
			params.Padding = theme.TextInputTheme.Padding
			params.RepeatDelay = theme.TextInputTheme.RepeatDelay
			params.RepeatInterval = theme.TextInputTheme.RepeatInterval
			params.Secure = theme.TextInputTheme.Secure
			params.SubmitOnEnter = theme.TextInputTheme.SubmitOnEnter
		}
	}

	// Set Defined values
	if t.definedParams.AllowDuplicateSubmit != nil {
		params.AllowDuplicateSubmit = t.definedParams.AllowDuplicateSubmit
	}
	if t.definedParams.ClearOnSubmit != nil {
		params.ClearOnSubmit = t.definedParams.ClearOnSubmit
	}
	if t.definedParams.Color != nil {
		params.Color.Idle = t.definedParams.Color.Idle
		params.Color.Disabled = t.definedParams.Color.Disabled
		params.Color.Caret = t.definedParams.Color.Caret
		params.Color.DisabledCaret = t.definedParams.Color.DisabledCaret
		params.Color.Hover = t.definedParams.Color.Hover
		params.Color.Pressed = t.definedParams.Color.Pressed
	}
	if t.definedParams.Face != nil {
		params.Face = t.definedParams.Face
	}
	if t.definedParams.IgnoreEmptySubmit != nil {
		params.IgnoreEmptySubmit = t.definedParams.IgnoreEmptySubmit
	}
	if t.definedParams.Image != nil {
		params.Image.Idle = t.definedParams.Image.Idle
		params.Image.Disabled = t.definedParams.Image.Disabled
		params.Image.Highlight = t.definedParams.Image.Highlight
		params.Image.Pressed = t.definedParams.Image.Pressed
		params.Image.PressedDisabled = t.definedParams.Image.PressedDisabled
		params.Image.PressedHover = t.definedParams.Image.PressedHover
		params.Image.Hover = t.definedParams.Image.Hover
	}
	if t.definedParams.Padding != nil {
		params.Padding = t.definedParams.Padding
	}
	if t.definedParams.RepeatDelay != nil {
		params.RepeatDelay = t.definedParams.RepeatDelay
	}
	if t.definedParams.RepeatInterval != nil {
		params.RepeatDelay = t.definedParams.RepeatDelay
	}
	if t.definedParams.Secure != nil {
		params.Secure = t.definedParams.Secure
	}
	if t.definedParams.SubmitOnEnter != nil {
		params.SubmitOnEnter = t.definedParams.SubmitOnEnter
	}

	// Set Default values
	if params.Image == nil {
		params.Image = &widget.TextInputImage{}
	}
	if params.Image.Highlight == nil {
		params.Image.Highlight = image.NewNineSliceColor(color.NRGBA{6, 67, 161, 100})
	}
	if params.RepeatDelay == nil {
		delay := 300 * time.Millisecond
		params.RepeatDelay = &delay
	}
	if params.RepeatInterval == nil {
		interval := 35 * time.Millisecond
		params.RepeatInterval = &interval
	}
	if params.Padding == nil {
		params.Padding = &widget.Insets{}
	}
	if params.Secure == nil {
		params.Secure = &FALSE
	}
	if params.ClearOnSubmit == nil {
		params.ClearOnSubmit = &FALSE
	}
	if params.IgnoreEmptySubmit == nil {
		params.IgnoreEmptySubmit = &FALSE
	}
	if params.AllowDuplicateSubmit == nil {
		params.AllowDuplicateSubmit = &FALSE
	}
	if params.SubmitOnEnter == nil {
		params.SubmitOnEnter = &TRUE
	}
	if params.Color != nil && params.Color.Caret == nil {
		params.Color.Caret = params.Color.Idle
	}

	t.computedParams = params
}

func (o TextInputOptions) WidgetOpts(opts ...widget.WidgetOpt) TextInputOpt {
	return func(t *BindingInput) {
		t.widgetOpts = append(t.widgetOpts, opts...)
	}
}

func (o TextInputOptions) ChangedHandler(f TextInputChangedHandlerFunc) TextInputOpt {
	return func(t *BindingInput) {
		t.ChangedEvent.AddHandler(func(args any) {
			if arg, ok := args.(*BindingInputChangedEventArgs); ok {
				f(arg)
			}
		})
	}
}

func (o TextInputOptions) SubmitHandler(f TextInputChangedHandlerFunc) TextInputOpt {
	return func(t *BindingInput) {
		t.SubmitEvent.AddHandler(func(args any) {
			if arg, ok := args.(*BindingInputChangedEventArgs); ok {
				f(arg)
			}
		})
	}
}

func (o TextInputOptions) ClearOnSubmit(clearOnSubmit bool) TextInputOpt {
	return func(t *BindingInput) {
		t.definedParams.ClearOnSubmit = &clearOnSubmit
	}
}

func (o TextInputOptions) IgnoreEmptySubmit(ignoreEmptySubmit bool) TextInputOpt {
	return func(t *BindingInput) {
		t.definedParams.IgnoreEmptySubmit = &ignoreEmptySubmit
	}
}

func (o TextInputOptions) AllowDuplicateSubmit(allowDuplicateSubmit bool) TextInputOpt {
	return func(t *BindingInput) {
		t.definedParams.AllowDuplicateSubmit = &allowDuplicateSubmit
	}
}

func (o TextInputOptions) Image(i *widget.TextInputImage) TextInputOpt {
	return func(t *BindingInput) {
		t.definedParams.Image = i
	}
}

func (o TextInputOptions) Color(c *widget.TextInputColor) TextInputOpt {
	return func(t *BindingInput) {
		t.definedParams.Color = c
	}
}

func (o TextInputOptions) Padding(i *widget.Insets) TextInputOpt {
	return func(t *BindingInput) {
		t.definedParams.Padding = i
	}
}

func (o TextInputOptions) Face(f *text.Face) TextInputOpt {
	return func(t *BindingInput) {
		t.definedParams.Face = f
	}
}

func (o TextInputOptions) RepeatInterval(i time.Duration) TextInputOpt {
	return func(t *BindingInput) {
		t.definedParams.RepeatInterval = &i
	}
}

func (o TextInputOptions) Validation(f TextInputValidationFunc) TextInputOpt {
	return func(t *BindingInput) {
		t.validationFunc = f
	}
}

func (o TextInputOptions) Placeholder(s string) TextInputOpt {
	return func(t *BindingInput) {
		t.placeholderText = s
	}
}

func (o TextInputOptions) Secure(b bool) TextInputOpt {
	return func(t *BindingInput) {
		t.definedParams.Secure = &b
	}
}

func (o TextInputOptions) PressedHandler(f TextInputPressedHandlerFunc) TextInputOpt {
	return func(b *BindingInput) {
		b.PressedEvent.AddHandler(func(args any) {
			if arg, ok := args.(*TextInputPressedEventArgs); ok {
				f(arg)
			}
		})
	}
}

func (o TextInputOptions) ReleasedHandler(f TextInputReleasedHandlerFunc) TextInputOpt {
	return func(b *BindingInput) {
		b.ReleasedEvent.AddHandler(func(args any) {
			if arg, ok := args.(*TextInputReleasedEventArgs); ok {
				f(arg)
			}
		})
	}
}

func (o TextInputOptions) ClickedHandler(f TextInputClickedHandlerFunc) TextInputOpt {
	return func(b *BindingInput) {
		b.ClickedEvent.AddHandler(func(args any) {
			if arg, ok := args.(*TextInputClickedEventArgs); ok {
				f(arg)
			}
		})
	}
}

func (o TextInputOptions) CursorEnteredHandler(f TextInputCursorHoverHandlerFunc) TextInputOpt {
	return func(b *BindingInput) {
		b.CursorEnteredEvent.AddHandler(func(args any) {
			if arg, ok := args.(*TextInputHoverEventArgs); ok {
				f(arg)
			}
		})
	}
}

func (o TextInputOptions) CursorMovedHandler(f TextInputCursorHoverHandlerFunc) TextInputOpt {
	return func(b *BindingInput) {
		b.CursorMovedEvent.AddHandler(func(args any) {
			if arg, ok := args.(*TextInputHoverEventArgs); ok {
				f(arg)
			}
		})
	}
}

func (o TextInputOptions) CursorExitedHandler(f TextInputCursorHoverHandlerFunc) TextInputOpt {
	return func(b *BindingInput) {
		b.CursorExitedEvent.AddHandler(func(args any) {
			if arg, ok := args.(*TextInputHoverEventArgs); ok {
				f(arg)
			}
		})
	}
}

func (o TextInputOptions) StateChangedHandler(f TextInputChangedHandlerFunc) TextInputOpt {
	return func(b *BindingInput) {
		b.StateChangedEvent.AddHandler(func(args any) {
			if arg, ok := args.(*BindingInputChangedEventArgs); ok {
				f(arg)
			}
		})
	}
}

func (o TextInputOptions) TabOrder(to int) TextInputOpt {
	return func(t *BindingInput) {
		t.tabOrder = to
	}
}

// SubmitOnEnter ... Sets if the input will submit when pressing enter or not.
func (o TextInputOptions) SubmitOnEnter(submitOnEnter bool) TextInputOpt {
	return func(t *BindingInput) {
		t.definedParams.SubmitOnEnter = &submitOnEnter
	}
}

// MobileInputMode ... Sets the keyboard type to use when viewed on a mobile browser.
//
// https://css-tricks.com/everything-you-ever-wanted-to-know-about-inputmode
func (o TextInputOptions) MobileInputMode(mobileInputMode mobile.InputMode) TextInputOpt {
	return func(t *BindingInput) {
		t.mobileInputMode = mobileInputMode
	}
}

/*********** End of Configuration *****************/

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

	w := 50

	if t.widget != nil && h < t.widget.MinHeight {
		h = t.widget.MinHeight
	}
	if t.widget != nil && w < t.widget.MinWidth {
		w = t.widget.MinWidth
	}

	h = max(1, h)

	return w, h
}

func (t *BindingInput) Render(screen *ebiten.Image) {
	t.init.Do()
	t.widget.Render(screen)
	t.renderImage(screen)
	t.renderText(screen)
}

func (t *BindingInput) Update(updObj *widget.UpdateObject) {
	t.init.Do()
	t.text.GetWidget().Disabled = t.widget.Disabled

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
		t.dirty = false
		t.ChangedEvent.Fire(&BindingInputChangedEventArgs{
			Keys:      t.inputKeys,
			TextInput: t,
			InputText: t.inputText,
		})
	}

	t.widget.Update(updObj)
	if t.text != nil {
		t.text.Update(updObj)
	}
}

func (t *BindingInput) idleState() textInputState {
	return func() (textInputState, bool) {
		if !t.focused {
			return t.idleState(), false
		}

		keys := inpututil.AppendJustPressedKeys([]ebiten.Key{})

		if len(keys) == 0 {
			return t.idleState(), false
		}

		if !t.widget.Disabled {
			t.Insert(keys)
		}

		return t.idleState(), false
	}
}

func (t *BindingInput) Insert(keys []ebiten.Key) {
	t.init.Do()

	if !t.takingInputs {
		return
	}

	t.inputKeys = keys
	t.inputText = toString(&keys)
	t.dirty = true
	t.cursorPosition = len(t.inputText) - 1
}

func (t *BindingInput) renderImage(screen *ebiten.Image) {
	if t.computedParams.Image == nil || t.computedParams.Image.Idle == nil {
		return
	}

	//pressed := (t.pressing && t.hovering)
	i := t.computedParams.Image.Idle

	if t.focused && t.computedParams.Image.Hover != nil {
		i = t.computedParams.Image.Hover
	}

	//switch {
	//	case t.widget.Disabled && t.computedParams.Image.Disabled != nil:
	//		if pressed && t.computedParams.Image.PressedDisabled != nil {
	//			i = t.computedParams.Image.PressedDisabled
	//		} else {
	//			i = t.computedParams.Image.Disabled
	//		}
	//case (t.focused || t.hovering) && t.computedParams.Image.Hover != nil:
	//		if pressed && t.computedParams.Image.PressedHover != nil {
	//			i = t.computedParams.Image.PressedHover
	//		} else {
	//			i = t.computedParams.Image.Hover
	//		}
	//	case pressed:
	//		if t.computedParams.Image.Pressed != nil {
	//			i = t.computedParams.Image.Pressed
	//		}
	//	}

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

	// rebuild origin at 0,0, so we can just apply translate to image instead of using
	// mask buffer image of screen size, only initial input size
	rect.Max.X -= rect.Min.X
	rect.Min.X = 0
	rect.Max.Y -= rect.Min.Y
	rect.Min.Y = 0

	tr := rect
	tr = tr.Add(img.Point{t.computedParams.Padding.Left, t.computedParams.Padding.Top})

	inputStr := t.inputText

	cx := 0
	if t.focused {
		sub := string([]rune(inputStr)[:t.cursorPosition])
		cx = fontAdvance(sub, t.computedParams.Face)

		dx := tr.Min.X + t.scrollOffset + cx + t.computedParams.Padding.Right - rect.Max.X
		if dx > 0 {
			t.scrollOffset -= dx
		}

		dx = tr.Min.X + t.scrollOffset + cx - t.computedParams.Padding.Left - rect.Min.X
		if dx < 0 {
			t.scrollOffset -= dx
		}
		// if t.dragStartIndex != -1 {
		//	dragString := string([]rune(inputStr)[:t.dragStartIndex])
		//	dragXStart := fontAdvance(dragString, t.computedParams.Face)

		//	dragStartDraw := min(dragXStart, cx)
		//	dragEndDraw := max(dragXStart, cx)

		//	// Change the Dx and the tr.Min.X based on selection
		//	t.computedParams.Image.Highlight.Draw(screen, dragEndDraw-dragStartDraw, tr.Dy(),
		//		func(opts *ebiten.DrawImageOptions) {
		//			opts.GeoM.Translate(float64(tr.Min.X+dragStartDraw+t.scrollOffset), float64(tr.Min.Y))
		//		})
		//}

	}
	tr = tr.Add(img.Point{t.scrollOffset, 0})

	t.text.SetLocation(tr)
	if len([]rune(t.inputText)) > 0 {
		t.text.Label = inputStr
	} else {
		t.text.Label = t.placeholderText
	}
	if (t.widget.Disabled || len([]rune(t.inputText)) == 0) && t.computedParams.Color.Disabled != nil {
		t.text.SetColor(t.computedParams.Color.Disabled)
	} else {
		switch {
		case t.pressing && t.computedParams.Color.Pressed != nil:
			t.text.SetColor(t.computedParams.Color.Pressed)
		case t.hovering && t.computedParams.Color.Hover != nil:
			t.text.SetColor(t.computedParams.Color.Hover)
		default:
			t.text.SetColor(t.computedParams.Color.Idle)
		}
	}
	t.text.Render(screen)
}

func (t *BindingInput) GetText() string {
	return t.inputText
}

func (t *BindingInput) SetText(text string) {
	t.init.Do()
	t.inputText = text
}

/** Focuser Interface - Start **/

// Focus this is used by ebitenui to focus the element - however it cannot be focused
// directly - only through "append" button selection
func (t *BindingInput) Focus(focused bool) {
	//t.init.Do()
	//t.GetWidget().FireFocusEvent(t, focused, img.Point{-1, -1})
	//t.focused = focused
}

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

func (t *BindingInput) GetFocuses() map[widget.FocusDirection]widget.Focuser {
	return t.focusMap
}

//func (t *BindingInput) RemoveFocus() {
//
//}

/** Focuser Interface - End **/

func (t *BindingInput) createWidget() {
	t.widget = widget.NewWidget(append([]widget.WidgetOpt{
		widget.WidgetOpts.TrackHover(true),

		widget.WidgetOpts.CursorEnterHandler(func(args *widget.WidgetCursorEnterEventArgs) {
			if !t.widget.Disabled {
				t.hovering = true
			}
			if t.hovering {
				t.CursorEnteredEvent.Fire(&TextInputHoverEventArgs{
					TextInput: t,
					Entered:   true,
					OffsetX:   args.OffsetX,
					OffsetY:   args.OffsetY,
					DiffX:     0,
					DiffY:     0,
				})
			}
		}),

		widget.WidgetOpts.CursorMoveHandler(func(args *widget.WidgetCursorMoveEventArgs) {
			t.CursorMovedEvent.Fire(&TextInputHoverEventArgs{
				TextInput: t,
				Entered:   false,
				OffsetX:   args.OffsetX,
				OffsetY:   args.OffsetY,
				DiffX:     args.DiffX,
				DiffY:     args.DiffY,
			})
		}),

		widget.WidgetOpts.CursorExitHandler(func(args *widget.WidgetCursorExitEventArgs) {
			if t.hovering {
				t.hovering = false
				t.CursorExitedEvent.Fire(&TextInputHoverEventArgs{
					TextInput: t,
					Entered:   false,
					OffsetX:   args.OffsetX,
					OffsetY:   args.OffsetY,
					DiffX:     0,
					DiffY:     0,
				})
			}
		}),

		widget.WidgetOpts.MouseButtonPressedHandler(func(args *widget.WidgetMouseButtonPressedEventArgs) {
			if !t.widget.Disabled && args.Button == ebiten.MouseButtonLeft {
				t.pressing = true
				t.PressedEvent.Fire(&TextInputPressedEventArgs{
					TextInput: t,
					OffsetX:   args.OffsetX,
					OffsetY:   args.OffsetY,
				})
			}
		}),

		widget.WidgetOpts.MouseButtonReleasedHandler(func(args *widget.WidgetMouseButtonReleasedEventArgs) {
			if t.pressing && !t.widget.Disabled && args.Button == ebiten.MouseButtonLeft {
				t.ReleasedEvent.Fire(&TextInputReleasedEventArgs{
					TextInput: t,
					Inside:    args.Inside,
					OffsetX:   args.OffsetX,
					OffsetY:   args.OffsetY,
				})
			}

			t.pressing = false
		}),

		widget.WidgetOpts.MouseButtonClickedHandler(func(args *widget.WidgetMouseButtonClickedEventArgs) {
			if !t.widget.Disabled && args.Button == ebiten.MouseButtonLeft {
				t.ClickedEvent.Fire(&TextInputClickedEventArgs{
					TextInput: t,
					OffsetX:   args.OffsetX,
					OffsetY:   args.OffsetY,
				})
			}
		}),
	}, t.widgetOpts...)...)
	t.widget.SetFocusable(t)
	t.mask = image.NewNineSliceColor(color.NRGBA{255, 0, 255, 255})
}

func (t *BindingInput) initWidget() {
	t.text = widget.NewText(widget.TextOpts.Text("", t.computedParams.Face, color.White))
	t.text.Validate()
}

func fontAdvance(s string, f *text.Face) int {
	a := text.Advance(s, *f)
	return int(math.Round(a))
}
