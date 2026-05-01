package ui

import (
	"github.com/aabalke/guac/utils"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

func NewHome(g *Game) {

	g.ui.focus.ClearFocus()

	b1 := NewCenteredButton("open a rom", func() {
		file := utils.OpenFile(
			"Open",
			"Roms (*.gb, *.gbc, *.gba, *.nds)",
			"gb", "gbc", "gba", "nds",
		)

		g.InitConsole(file)
	})

	b2 := NewCenteredButton("settings", func() {
		NewSettings(g, g.ui.PageId, MENU_GENERAL)
	})

	b3 := NewCenteredButton("quit", func() {
		g.quit = true
	})

	root := NewCenteredPage(g.ui.res.bg, b1, b2, b3)

	g.ui.PageId = PAGE_HOME
	g.ui.ui = &ebitenui.UI{
		Container:    root,
		PrimaryTheme: NewTheme(g.ui.res),
	}
	g.ui.focus.other = g.ui.ui.Container.GetFocusers()
	g.ui.focus.BuildFocus(g.ui.ui)
}

func NewPause(g *Game) {

	g.ui.focus.ClearFocus()

	b1 := NewCenteredButton("resume", func() {
		g.TogglePause()
	})

	b2 := NewCenteredButton("settings", func() {
		NewSettings(g, g.ui.PageId, MENU_GENERAL)
	})

	b3 := NewCenteredButton("main menu", func() {
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
	g.ui.ui = &ebitenui.UI{
		Container:    root,
		PrimaryTheme: NewTheme(g.ui.res),
	}
	g.ui.focus.other = g.ui.ui.Container.GetFocusers()
	g.ui.focus.BuildFocus(g.ui.ui)
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
			widget.RowLayoutOpts.Spacing(5),
		)),
	)

	for _, b := range buttons {
		c.AddChild(b)
	}

	root.AddChild(c)

	return root
}

func NewCenteredButton(text string, f func()) *widget.Button {

	b := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				MaxWidth: BUTTON_WIDTH,
				Stretch:  true,
			}),
		),

		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			f()
		}),
	)

	b.SetText(text)

	return b
}
