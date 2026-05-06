package ui

import (
	"math"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
)

func NewSettings(g *Game, oldId PageId, initMenu int) {

	g.ui.focus.ClearFocus()
	g.ui.PrevPageId = oldId

	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(g.ui.res.bg),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	c := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
				StretchVertical:    true,
			}),
			widget.WidgetOpts.MinSize(1024, 64),
		),

		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch(
				[]bool{false, true},
				[]bool{true, true},
			),
		)),
	)

	g.ui.content = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(24)),
			widget.GridLayoutOpts.Spacing(32, 16),
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, true}, []bool{}),
		)),
	)

	scr := NewScrollableContainer(g.ui)
	NewSidebar(g, initMenu)

	c.AddChild(g.ui.sidebar, scr)
	root.AddChild(c)

	g.ui.PageId = PAGE_SETTINGS
	g.ui.ui = &ebitenui.UI{
		Container:    root,
		PrimaryTheme: NewTheme(g.ui.res),
	}

	//ui.SetDebugMode(true)

	g.ui.focus.sidebar = g.ui.sidebar.GetFocusers()
	g.ui.focus.BuildFocus(g.ui.ui)
	g.ui.focus.FocusSidebar(0)
}

func NewScrollableContainer(ui *Ui) *widget.Container {

	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Spacing(2, 0),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{true}),
		)),
	)

	ui.scrollable = widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(ui.content),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.Image(&scrollContainerImage),
	)

	pageSizeFunc := func() int {

		scrollableHeight := ui.scrollable.ViewRect().Dy()
		contentHeight := ui.content.GetWidget().Rect.Dy()
		if scrollableHeight >= contentHeight {
			ui.scrollable.ScrollTop = 0
			ui.slider.Current = 0
			return 0
		}

		return int(math.Round(
			float64(scrollableHeight) /
				float64(contentHeight) *
				1000))
	}

	ui.slider = widget.NewSlider(
		widget.SliderOpts.InitialCurrent(0),
		widget.SliderOpts.Orientation(widget.DirectionVertical),
		widget.SliderOpts.MinMax(0, 1000),
		widget.SliderOpts.PageSizeFunc(pageSizeFunc),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {

			if args.Dragging {
				ui.focus.DeFocus()
			}

			ui.scrollable.ScrollTop = float64(args.Slider.Current) / 1000
		}),

		widget.SliderOpts.WidgetOpts(
			widget.WidgetOpts.OnUpdate(func(args widget.HasWidget) {

				scrollableHeight := ui.scrollable.ViewRect().Dy()
				contentHeight := ui.content.GetWidget().Rect.Dy()

				if scrollableHeight >= contentHeight {
					ui.slider.GetWidget().SetVisibility(widget.Visibility_Hide)
					return
				}
				ui.slider.GetWidget().SetVisibility(widget.Visibility_Show)
			}),
		),
	)

	ui.scrollable.GetWidget().ScrolledEvent.AddHandler(func(args any) {

		ui.focus.DeFocus()

		scrollableHeight := ui.scrollable.ViewRect().Dy()
		contentHeight := ui.content.GetWidget().Rect.Dy()
		if scrollableHeight >= contentHeight {
			ui.scrollable.ScrollTop = 0
			ui.slider.Current = 0
			return
		}

		if a, ok := args.(*widget.WidgetScrolledEventArgs); ok {
			ui.slider.Current -= int(math.Round(a.Y * float64(pageSizeFunc())))
		}
	})

	root.AddChild(ui.scrollable, ui.slider)

	return root
}

func NewSidebar(g *Game, initMenu int) {

	g.ui.sidebar = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(24)),
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		)),
	)

	radios := []widget.RadioGroupElement{}
	var initButton *widget.Button
	sidebarFields := NewSidebarFields(g.ui.res)
	for i, field := range sidebarFields {
		b := NewSidebarButton(g, field)
		g.ui.sidebar.AddChild(b)
		radios = append(radios, b)

		if i == initMenu {
			initButton = b
		}
	}

	widget.NewRadioGroup(
		widget.RadioGroupOpts.Elements(radios...),
		widget.RadioGroupOpts.InitialElement(initButton),
	)

	// initialize first menu
	sidebarFields[initMenu].f(g)
	g.ui.focus.submenu = g.ui.content.GetFocusers()
	g.ui.focus.BuildFocus(g.ui.ui)
}

func NewSidebarButton(g *Game, sf SidebarField) *widget.Button {

	b := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),

		widget.ButtonOpts.TextPosition(
			widget.TextPositionStart,
			widget.TextPositionCenter,
		),

		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			sf.f(g)
			g.ui.scrollable.ScrollTop = 0
			g.ui.slider.Current = 0
		}),
	)

	b.SetText(sf.label)

	return b
}
