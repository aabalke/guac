package ui

import (
	"github.com/aabalke/guac/utils"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

func NewHome(g *Game) {
	g.ui.focus.ClearFocus()

	l := g.ui.res.localization.Main

	b1 := NewCenteredButton(l.Open, func(*widget.ButtonClickedEventArgs) {
		file := utils.OpenFile(
			l.DialogTitle,
			l.DialogDesc,
			"gb", "gbc", "gba", "nds",
		)

		g.InitConsole(file)
	})

	b2 := NewCenteredButton(l.Settings, func(*widget.ButtonClickedEventArgs) {
		NewSettings(g, g.ui.PageId, MENU_GENERAL)
	})

	b3 := NewCenteredButton(l.Quit, func(*widget.ButtonClickedEventArgs) {
		g.quit = true
	})

	root := NewCenteredPage(g.ui.res.bg, b1, b2, b3)
	g.ui.PageId = PAGE_HOME
	newUi(g, root)
	buildOtherFocus(g.ui.ui.Container.GetFocusers())
}

func NewPause(g *Game) {
	g.ui.focus.ClearFocus()

	l := g.ui.res.localization.Pause

	b1 := NewCenteredButton(l.Resume, func(*widget.ButtonClickedEventArgs) {
		g.TogglePause()
	})

	b2 := NewCenteredButton(l.Settings, func(*widget.ButtonClickedEventArgs) {
		NewSettings(g, g.ui.PageId, MENU_GENERAL)
	})

	b3 := NewCenteredButton(l.Main, func(*widget.ButtonClickedEventArgs) {
		NewHome(g)

		if g.nds != nil {
			g.nds.Close()
		}
		if g.gba != nil {
			g.gba.Close()
		}
		if g.gb != nil {
			g.gb.Close()
		}

		g.nds = nil
		g.gba = nil
		g.gb = nil
		g.paused = false
	})

	root := NewCenteredPage(g.ui.res.bg, b1, b2, b3)
	g.ui.PageId = PAGE_PAUSE
	newUi(g, root)
	buildOtherFocus(g.ui.ui.Container.GetFocusers())
}

func NewCenteredPage(bg *image.NineSlice, buttons ...*widget.Button) *widget.Container {
	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(bg),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	c := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),

		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(50)),
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(4),
		)),
	)

	for _, b := range buttons {
		c.AddChild(b)
	}

	root.AddChild(c)

	return root
}

func NewCenteredButton(text string, f func(*widget.ButtonClickedEventArgs)) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.TextLabel(text),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),

			widget.WidgetOpts.MinSize(BUTTON_WIDTH, 0),
		),

		widget.ButtonOpts.ClickedHandler(f),
	)
}

func newUi(g *Game, root *widget.Container) {
	g.ui.ui = &ebitenui.UI{
		Container:    root,
		PrimaryTheme: NewTheme(g.ui.res),
	}
	g.ui.focus.ui = g.ui.ui
}
