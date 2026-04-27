package ui

import (
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/utils"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

func NewHome(g *Game) *UiPage {

	face := g.res.fonts.face
	img := g.res.buttonImage

	b1 := NewCenteredButton("open a rom", face, img, func(*widget.ButtonClickedEventArgs) {
		file := utils.OpenFile("Open", "Roms (*.gb, *.gbc, *.gba, *.nds)", "gb", "gbc", "gba", "nds")
		g.InitConsole(file)
	})

	b2 := NewCenteredButton("settings", face, img, func(*widget.ButtonClickedEventArgs) {
		g.ui = NewSettings(g, g.ui.Id, "general")
	})

	b3 := NewCenteredButton("quit", face, img, func(*widget.ButtonClickedEventArgs) {
		g.quit = true
	})

	return NewCenteredPage(PAGE_HOME, g.res.bg, b1, b2, b3)
}

func NewPause(g *Game) *UiPage {

	face := g.res.fonts.face
	img := g.res.buttonImage

	b1 := NewCenteredButton("resume", face, img, func(*widget.ButtonClickedEventArgs) {
		g.TogglePause()
	})

	b2 := NewCenteredButton("settings", face, img, func(*widget.ButtonClickedEventArgs) {
		g.ui = NewSettings(g, g.ui.Id, "general")
	})

	b3 := NewCenteredButton("main menu", face, img, func(*widget.ButtonClickedEventArgs) {
		g.ui = NewHome(g)

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

	return NewCenteredPage(PAGE_PAUSE, g.res.bg, b1, b2, b3)
}

func NewCenteredPage(id PageId, bg *image.NineSlice, buttons ...*widget.Button) *UiPage {

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

	return &UiPage{
		Id: id,
		ui: &ebitenui.UI{
			Container: root,
		}}
}

func NewCenteredButton(text string, face *text.Face, img *widget.ButtonImage, f func(*widget.ButtonClickedEventArgs)) *widget.Button {

	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				MaxWidth: 256,
				Stretch:  true,
			}),
		),

		widget.ButtonOpts.Image(img),

		widget.ButtonOpts.Text(
			text,
			face,
			&widget.ButtonTextColor{
				Idle: config.Conf.Ui.MenuForegroundColor,
			},
		),

		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   32,
			Right:  32,
			Top:    4,
			Bottom: 4,
		}),

		widget.ButtonOpts.ClickedHandler(f),
	)
}
