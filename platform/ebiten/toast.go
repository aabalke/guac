package ui

import (
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

type Toast struct {
	active int
	ui     *ebitenui.UI
	res    *Resources
	bus    *TimedEventBus
}

func NewToast(ui *Ui) *Toast {
	return &Toast{
		res: ui.res,
		ui: &ebitenui.UI{
			Container: widget.NewContainer(
				widget.ContainerOpts.Layout(NewStaticLayout()),
			),
			PrimaryTheme: NewTheme(ui.res),
		},
		bus: ui.TimedEvents,
	}
}

func (t *Toast) Message(message string) {
	e := &TimedEvent{}

	const DURATION = 120
	const TRANS_DURATION = 40
	const POS_DT = 1000

	e.TicksLeft = DURATION

	padding := 16
	data := &StaticLayoutData{}
	_, bot := ebiten.WindowSize()
	data.Position.X = padding - int(POS_DT)
	data.Position.Y = bot - (padding + buttonInset.Bottom + 36)

	container := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(t.res.sec),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(data),
		),

		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(&buttonInset),
		)),
	)

	container.AddChild(widget.NewText(
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),

		widget.TextOpts.Text(
			message,
			t.res.fonts.smallFace,
			*t.res.fgClr,
		),
	))

	t.ui.Container.AddChild(container)

	t.active++

	e.OnTick = func() {
		if e.TicksLeft > DURATION-TRANS_DURATION {
			transLast := DURATION - TRANS_DURATION
			mul := (float32(e.TicksLeft-transLast) / TRANS_DURATION)
			mul = CalcBezierBlend(mul)
			data.Position.X = padding - int(POS_DT*mul)
			t.ui.Container.RequestRelayout()
		}

		if e.TicksLeft < TRANS_DURATION {
			transLast := TRANS_DURATION
			mul := (float32(e.TicksLeft-transLast) / TRANS_DURATION)
			mul = CalcBezierBlend(mul)
			data.Position.X = padding - int(POS_DT*mul)
			t.ui.Container.RequestRelayout()
		}

		_, bot := ebiten.WindowSize()
		data.Position.Y = bot - (padding + buttonInset.Bottom + 36)
	}

	e.OnDone = func() {
		t.ui.Container.RemoveChild(container)
		t.active--
	}

	t.bus.AddEvent(e)
}

// https://stackoverflow.com/questions/13462001/ease-in-and-ease-out-animation-formula
func CalcBezierBlend(t float32) float32 {
	return t * t * (3.0 - 2.0*t)
}
