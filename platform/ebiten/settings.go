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

type SidebarButtonData struct {
	label   string
	onClick func(*widget.ButtonClickedEventArgs)
}

func NewSettings(g *Game, oldId PageId, initMenu int) {
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

	scrollRoot, scrollContainer := NewScrollableContainer(g.ui)
	buttonDatas := NewSidebarButtonDatas(g, scrollContainer)
	sidebar := NewSidebar(g, buttonDatas, initMenu)

	c.AddChild(sidebar, scrollRoot)
	root.AddChild(c)

	g.ui.PrevPageId = oldId
	g.ui.PageId = PAGE_SETTINGS

	g.ui.returnFocuser = sidebar.GetFocusers()[len(sidebar.GetFocusers())-1]

	newUi(g, root)

	FocusFirst(g.ui)
}

func NewSidebarButtonDatas(g *Game, scrollContainer *widget.ScrollContainer) []SidebarButtonData {
	l := g.ui.res.localization

	labels := []string{
		l["settings.general"],
		l["settings.ui"],
		l["settings.gb"],
		l["settings.gba"],
		l["settings.nds"],
		l["settings.about"],
		l["settings.return"],
	}

	builderFuncs := []func(g *Game) *widget.Container{
		NewGeneralMenu,
		NewUiMenu,
		NewGbMenu,
		NewGbaMenu,
		NewNdsMenu,
		NewAboutMenu,
		// return needs special handling
	}

	datas := make([]SidebarButtonData, len(labels))

	for i, f := range builderFuncs {
		datas[i] = SidebarButtonData{
			label: labels[i],
			onClick: func(*widget.ButtonClickedEventArgs) {
				g.ui.scrollableMenu.content.RemoveChildren()
				g.ui.scrollableMenu.content.AddChild(f(g))
				scrollContainer.ScrollTop = 0
				g.ui.scrollableMenu.slider.Current = 0
			},
		}
	}

	returnIdx := len(labels) - 1

	datas[returnIdx].label = labels[returnIdx]

	switch g.ui.PrevPageId {
	case PAGE_HOME:
		datas[returnIdx].onClick = func(*widget.ButtonClickedEventArgs) { NewHome(g) }
	case PAGE_PAUSE:
		datas[returnIdx].onClick = func(*widget.ButtonClickedEventArgs) { NewPause(g) }
	}

	return datas
}

func NewSidebar(g *Game, buttons []SidebarButtonData, initMenu int) *widget.Container {
	sidebar := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(24)),
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		)),
	)

	var (
		radios     []widget.RadioGroupElement
		initButton *widget.Button
	)

	for i, button := range buttons {
		b := NewSidebarButton(button.label, button.onClick)
		sidebar.AddChild(b)
		radios = append(radios, b)

		if i == initMenu {
			initButton = b
		}
	}

	widget.NewRadioGroup(
		widget.RadioGroupOpts.Elements(radios...),
		widget.RadioGroupOpts.InitialElement(initButton),
	)

	buttons[initMenu].onClick(nil)

	return sidebar
}

func NewSidebarButton(label string, f func(*widget.ButtonClickedEventArgs)) *widget.Button {
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

		widget.ButtonOpts.ClickedHandler(f),
	)
}

func NewScrollableContainer(ui *Ui) (*widget.Container, *widget.ScrollContainer) {
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

	scrollable := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(menu.content),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.Image(&scrollContainerImage),
	)

	pageSizeFunc := func() int {
		scrollableHeight := scrollable.ViewRect().Dy()
		contentHeight := menu.content.GetWidget().Rect.Dy()
		if scrollableHeight >= contentHeight {
			scrollable.ScrollTop = 0
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

			scrollable.ScrollTop = float64(args.Slider.Current) / 1000
		}),

		widget.SliderOpts.WidgetOpts(
			widget.WidgetOpts.OnUpdate(func(widget.HasWidget) {
				scrollableHeight := scrollable.ViewRect().Dy()
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

	scrollable.GetWidget().ScrolledEvent.AddHandler(func(args any) {
		DeFocus(ui)

		scrollableHeight := scrollable.ViewRect().Dy()
		contentHeight := menu.content.GetWidget().Rect.Dy()
		if scrollableHeight >= contentHeight {
			scrollable.ScrollTop = 0
			menu.slider.Current = 0
			return
		}

		if a, ok := args.(*widget.WidgetScrolledEventArgs); ok {
			menu.slider.Current -= int(math.Round(a.Y * float64(pageSizeFunc())))
		}
	})

	menu.root.AddChild(scrollable, menu.slider)
	return menu.root, scrollable
}
