package ui

import (
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
)

const TICKS_PER_TOAST = 60

type Toast struct {
	enabled  bool
	duration int
	ui       *ebitenui.UI
	res      *Resources
}

func NewToast(res *Resources) *Toast {
	return &Toast{
		res: res,
		ui: &ebitenui.UI{
			Container: widget.NewContainer(
				widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
			),
			PrimaryTheme: NewTheme(res),
		},
	}
}

func (t *Toast) AddMessage(message string) {

	container := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(t.res.sec),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
				Padding:            widget.NewInsetsSimple(16),
			}),
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
			})),

		widget.TextOpts.Text(
			message,
			t.res.fonts.smallFace,
			*t.res.fgClr,
		),
	))

	t.ui.Container.RemoveChildren()
	t.ui.Container.AddChild(container)
	t.duration = TICKS_PER_TOAST
}

func (t *Toast) Update() {
	t.enabled = t.duration > 0
	t.duration--
}
