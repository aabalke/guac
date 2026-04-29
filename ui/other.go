package ui

import (
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/utils"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

func NewHome(g *Game) {

	g.ui.focus.ClearFocus()

	res := g.ui.res
	face := res.fonts.face
	img := res.buttonImage

	b1 := NewCenteredButton("open a rom", face, img, func() {
		file := utils.OpenFile(
			"Open",
			"Roms (*.gb, *.gbc, *.gba, *.nds)",
			"gb", "gbc", "gba", "nds",
		)

		g.InitConsole(file)
	})

	b2 := NewCenteredButton("settings", face, img, func() {
		NewSettings(g, g.ui.PageId, MENU_GENERAL)
	})

	b3 := NewCenteredButton("quit", face, img, func() {
		g.quit = true
	})

	root := NewCenteredPage(res.bg, b1, b2, b3)
	g.ui.PageId = PAGE_HOME
	g.ui.ui = &ebitenui.UI{Container: root}
	g.ui.focus.other = g.ui.ui.Container.GetFocusers()
	g.ui.focus.BuildFocus(g.ui.ui)
}

func NewPause(g *Game) {
	g.ui.focus.ClearFocus()

	res := g.ui.res
	face := res.fonts.face
	img := res.buttonImage

	b1 := NewCenteredButton("resume", face, img, func() {
		g.TogglePause()
	})

	b2 := NewCenteredButton("settings", face, img, func() {
		NewSettings(g, g.ui.PageId, MENU_GENERAL)
	})

	b3 := NewCenteredButton("main menu", face, img, func() {
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

	root := NewCenteredPage(res.bg, b1, b2, b3)
	g.ui.PageId = PAGE_PAUSE
	g.ui.ui = &ebitenui.UI{Container: root}
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

func NewCenteredButton(text string, face *text.Face, img *widget.ButtonImage, f func()) *widget.Button {

	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				MaxWidth: BUTTON_WIDTH,
				Stretch:  true,
			}),
		),

		widget.ButtonOpts.Image(img),

		widget.ButtonOpts.Text(text, face, &widget.ButtonTextColor{
			Idle: config.Conf.Ui.MenuForegroundColor,
		}),

		widget.ButtonOpts.TextPadding(&buttonInset),

		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			f()
		}),
	)
}
