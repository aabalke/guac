package ui

import (
	"github.com/aabalke/guac/utils"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

// TODO: LOCALIZE
var extensions = []string{
	"gb", "gbc", "gba", "nds", "zip",
}

func NewHome(g *Game) {
	l := g.ui.res.localization

	b1 := NewCenteredButton(l["main.open"], func(*widget.ButtonClickedEventArgs) {
		if len(g.ui.gamepadIds) != 0 {
			NewFileMenu(g, g.ui.PageId)
		} else {

			file := utils.OpenFile(
				l["main.dialog_title"],
				l["main.dialog_desc"],
				extensions...,
			)

			g.InitConsole(file)
		}
	})

	b2 := NewCenteredButton(l["main.settings"], func(*widget.ButtonClickedEventArgs) {
		NewSettings(g, g.ui.PageId, MENU_GENERAL)
	})

	b3 := NewCenteredButton(l["main.quit"], func(*widget.ButtonClickedEventArgs) {
		g.quit = true
	})

	g.ui.returnFocuser = b3

	c := NewCenteredPage(g.ui.res.bg, b1, b2, b3)
	root := newUi(g)
	root.AddChild(c)
	g.ui.PageId = PAGE_HOME
}

func NewPause(g *Game) {
	l := g.ui.res.localization

	b1 := NewCenteredButton(l["pause.resume"], func(*widget.ButtonClickedEventArgs) {
		g.TogglePause()
	})

	b2 := NewCenteredButton(l["pause.settings"], func(*widget.ButtonClickedEventArgs) {
		NewSettings(g, g.ui.PageId, MENU_GENERAL)
	})

	b3 := NewCenteredButton(l["pause.main"], func(*widget.ButtonClickedEventArgs) {
		NewHome(g)

		if g.nds != nil || g.gba != nil || g.gb != nil {
			g.emuClose()
		}

		g.nds, g.gba, g.gb, g.paused = nil, nil, nil, false
	})

	g.ui.returnFocuser = b3

	c := NewCenteredPage(g.ui.res.bg, b1, b2, b3)
	root := newUi(g)
	root.AddChild(c)
	g.ui.PageId = PAGE_PAUSE
}

func NewCenteredPage(bg *image.NineSlice, buttons ...*widget.Button) *widget.Container {
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

	return c
}

func NewCenteredButton(text string, f func(*widget.ButtonClickedEventArgs)) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.TextLabel(text),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
			widget.WidgetOpts.MinSize(BUTTON_WIDTH, 0),
		),

		widget.ButtonOpts.ClickedHandler(f),
	)
}

func newUi(g *Game) widget.Containerer {
	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(g.ui.res.bg),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	g.ui.ui = &ebitenui.UI{
		Container:    root,
		PrimaryTheme: NewTheme(g.ui.res),
	}

	return root
}
