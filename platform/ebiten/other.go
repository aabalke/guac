package ui

import (
	"github.com/aabalke/guac/utils"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

var extensions = []string{
	"gb", "gbc", "gba", "nds", "zip",
}

func NewHome(g *Game) {
	l := g.ui.res.localization.Main

	b1 := NewCenteredButton(l.Open, func(*widget.ButtonClickedEventArgs) {
		if len(g.ui.gamepadIds) != 0 {
			NewFileMenu(g, g.ui.PageId)
		} else {

			file := utils.OpenFile(
				l.DialogTitle,
				l.DialogDesc,
				extensions...,
			)

			g.InitConsole(file)
		}
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
}

func NewPause(g *Game) {
	l := g.ui.res.localization.Pause

	b1 := NewCenteredButton(l.Resume, func(*widget.ButtonClickedEventArgs) {
		g.TogglePause()
	})

	b2 := NewCenteredButton(l.Settings, func(*widget.ButtonClickedEventArgs) {
		NewSettings(g, g.ui.PageId, MENU_GENERAL)
	})

	b3 := NewCenteredButton(l.Main, func(*widget.ButtonClickedEventArgs) {
		NewHome(g)

		if g.nds != nil || g.gba != nil || g.gb != nil {
			g.emuClose()
		}

		g.nds = nil
		g.gba = nil
		g.gb = nil
		g.paused = false
	})

	root := NewCenteredPage(g.ui.res.bg, b1, b2, b3)
	g.ui.PageId = PAGE_PAUSE
	newUi(g, root)
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
}
