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

	scrollableRoot := NewScrollableContainer(g.ui)

	g.ui.sidebar = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(24)),
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		)),
	)

	l := g.ui.res.localization.Settings.Sidebar

	fields := []SidebarField{
		{l.General, func(g *Game) {
			g.ui.scrollableMenu.content.RemoveChildren()
			g.ui.scrollableMenu.content.AddChild(NewGeneralMenu(g))
		}},
		{l.Ui, func(g *Game) {
			g.ui.scrollableMenu.content.RemoveChildren()
			g.ui.scrollableMenu.content.AddChild(NewUiMenu(g))
		}},
		{l.Gb, func(g *Game) {
			g.ui.scrollableMenu.content.RemoveChildren()
			g.ui.scrollableMenu.content.AddChild(NewGbMenu(g))
		}},
		{l.Gba, func(g *Game) {
			g.ui.scrollableMenu.content.RemoveChildren()
			g.ui.scrollableMenu.content.AddChild(NewGbaMenu(g))
		}},
		{l.Nds, func(g *Game) {
			g.ui.scrollableMenu.content.RemoveChildren()
			g.ui.scrollableMenu.content.AddChild(NewNdsMenu(g))
		}},
		{l.About, func(g *Game) {
			g.ui.scrollableMenu.content.RemoveChildren()
			g.ui.scrollableMenu.content.AddChild(NewAboutMenu(g))
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
	fields[initMenu].f(g)

	c.AddChild(g.ui.sidebar, scrollableRoot)
	root.AddChild(c)

	g.ui.PageId = PAGE_SETTINGS

	newUi(g, root)

	FocusFirst(g.ui)
}

func NewScrollableContainer(ui *Ui) *widget.Container {
	ui.scrollableMenu = &ScrollableMenu{}
	menu := ui.scrollableMenu

	menu.content = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true}),
		)),
	)

	menu.root = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Spacing(2, 0),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{true}),
		)),
	)

	menu.scrollable = widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(menu.content),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.Image(&scrollContainerImage),
	)

	pageSizeFunc := func() int {
		scrollableHeight := menu.scrollable.ViewRect().Dy()
		contentHeight := menu.content.GetWidget().Rect.Dy()
		if scrollableHeight >= contentHeight {
			menu.scrollable.ScrollTop = 0
			menu.slider.Current = 0
			return 0
		}

		return int(math.Round(
			float64(scrollableHeight) /
				float64(contentHeight) *
				1000,
		))
	}

	menu.slider = widget.NewSlider(
		widget.SliderOpts.InitialCurrent(0),
		widget.SliderOpts.Orientation(widget.DirectionVertical),
		widget.SliderOpts.MinMax(0, 1000),
		widget.SliderOpts.PageSizeFunc(pageSizeFunc),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			if args.Dragging {
				DeFocus(ui)
			}

			menu.scrollable.ScrollTop = float64(args.Slider.Current) / 1000
		}),

		widget.SliderOpts.WidgetOpts(
			widget.WidgetOpts.OnUpdate(func(widget.HasWidget) {
				scrollableHeight := menu.scrollable.ViewRect().Dy()
				contentHeight := menu.content.GetWidget().Rect.Dy()

				var v widget.Visibility
				if scrollableHeight >= contentHeight {
					v = widget.Visibility_Hide
				} else {
					v = widget.Visibility_Show
				}

				menu.slider.GetWidget().SetVisibility(v)
			}),
		),
	)

	menu.scrollable.GetWidget().ScrolledEvent.AddHandler(func(args any) {
		DeFocus(ui)

		scrollableHeight := menu.scrollable.ViewRect().Dy()
		contentHeight := menu.content.GetWidget().Rect.Dy()
		if scrollableHeight >= contentHeight {
			menu.scrollable.ScrollTop = 0
			menu.slider.Current = 0
			return
		}

		if a, ok := args.(*widget.WidgetScrolledEventArgs); ok {
			menu.slider.Current -= int(math.Round(a.Y * float64(pageSizeFunc())))
		}
	})

	menu.root.AddChild(menu.scrollable, menu.slider)
	return menu.root
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
			g.ui.scrollableMenu.scrollable.ScrollTop = 0
			g.ui.scrollableMenu.slider.Current = 0
		}),
	)
}
