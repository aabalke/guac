package ui

import (
	"math"

	"github.com/ebitenui/ebitenui/widget"
)

const (
	MENU_GENERAL = iota
	MENU_UI
	MENU_GB
	MENU_GBA
	MENU_NDS
	MENU_ABOUT
	MENU_RETURN
)

type SidebarField struct {
	label string
	f     func(g *Game)
}

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
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true}),
		)),
	)

	scr := NewScrollableContainer(g.ui)

	// sidebar start

	g.ui.sidebar = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(24)),
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		)),
	)

	l := g.ui.res.localization.Settings.Sidebar

	fields := []SidebarField{
		{l.General, func(g *Game) {
			g.ui.content.RemoveChildren()
			g.ui.content.AddChild(NewGeneralMenu(g))
		}},
		{l.Ui, func(g *Game) {
			g.ui.content.RemoveChildren()
			g.ui.content.AddChild(NewUiMenu(g))
		}},
		{l.Gb, func(g *Game) {
			g.ui.content.RemoveChildren()
			g.ui.content.AddChild(NewGbMenu(g))
		}},
		{l.Gba, func(g *Game) {
			g.ui.content.RemoveChildren()
			g.ui.content.AddChild(NewGbaMenu(g))
		}},
		{l.Nds, func(g *Game) {
			g.ui.content.RemoveChildren()
			g.ui.content.AddChild(NewNdsMenu(g))
		}},
		{l.About, func(g *Game) {
			g.ui.content.RemoveChildren()
			g.ui.content.AddChild(NewAboutMenu(g))
		}},
		{
			l.Return, func(g *Game) {
				switch g.ui.PrevPageId {
				case PAGE_HOME:
					NewHome(g)
				case PAGE_PAUSE:
					NewPause(g)
				}
			},
		},
	}

	var (
		radios     []widget.RadioGroupElement
		initButton *widget.Button
	)
	for i, field := range fields {
		b := NewSidebarButton(g, field.label, field.f)
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

	// init first menu
	g.ui.focus.sidebar = g.ui.sidebar.GetFocusers()
	fields[initMenu].f(g)

	// sidebar end

	c.AddChild(g.ui.sidebar, scr)
	root.AddChild(c)

	g.ui.PageId = PAGE_SETTINGS

	newUi(g, root)

	g.ui.focus.ui = g.ui.ui

	g.ui.focus.FocusSidebar(0)
}

func NewScrollableContainer(ui *Ui) *widget.Container {
	scr := widget.NewContainer(
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
				1000,
		))
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
			widget.WidgetOpts.OnUpdate(func(widget.HasWidget) {
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

	scr.AddChild(ui.scrollable, ui.slider)
	return scr
}

func NewSidebarButton(g *Game, label string, f func(g *Game)) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.TextLabel(label),
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
			f(g)
			g.ui.scrollable.ScrollTop = 0
			g.ui.slider.Current = 0
		}),
	)
}
