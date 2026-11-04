package ui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/config/file"
	"github.com/aabalke/guac/utils"
	"github.com/ebitenui/ebitenui/widget"
)

func NewFileMenu(g *Game, oldId PageId) {
	g.ui.PrevPageId = oldId

	menu := &ScrollableMenu{}
	g.ui.scrollableMenu = menu

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
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true, true}, []bool{false, true}),
		)),
	)

	menu.content = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true}),
		)),
	)

	menu.root = NewScrollableContainer(g.ui)

	p := getAbsPath(config.Conf.General.RomsPath)
	entries, err := os.ReadDir(p)
	if err != nil {
		panic(err)
	}

	elements := []widget.PreferredSizeLocateableWidget{}
	for _, v := range entries {
		if v.IsDir() {
			elements = append(elements, NewFileEntryButton(g.ui, "> "+v.Name(),
				func(bcea *widget.ButtonClickedEventArgs) {
					p := getAbsPath(config.Conf.General.RomsPath)
					config.Conf.General.RomsPath = filepath.Join(p, v.Name())
					NewFileMenu(g, g.ui.PrevPageId)
				}))
			continue
		}
	}

	exts := []string{}
	for _, ext := range extensions {
		exts = append(exts, "."+ext)
	}

	for _, v := range entries {

		if v.IsDir() {
			continue
		}

		for _, ext := range exts {
			if strings.HasSuffix(v.Name(), ext) {
				elements = append(elements, NewFileEntryButton(g.ui, v.Name(),
					func(bcea *widget.ButtonClickedEventArgs) {
						g.InitConsole(filepath.Join(config.Conf.General.RomsPath, v.Name()))
					}))
				break
			}
		}
	}

	menu.content.RemoveChildren()
	menu.content.AddChild(elements...)
	c.AddChild(NewNavBar(g), menu.root)
	root.AddChild(c)

	g.ui.PageId = PAGE_FILE_SELECT
	newUi(g, root)

	g.ui.ui.SetFocusedWidget(menu.content.GetFocusers()[0])
}

func getAbsPath(p string) string {
	p, err := filepath.Abs(p)
	if err != nil {
		p, _ = filepath.Abs("./")
	}

	return p
}

func NewNavBar(g *Game) *widget.Container {
	ui := g.ui
	l := g.ui.res.localization.FileMenu
	navbar := newOneCol()

	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(24)),
			widget.GridLayoutOpts.Spacing(32, 16),
			widget.GridLayoutOpts.Columns(3),
			widget.GridLayoutOpts.Stretch(
				[]bool{true, false, false},
				[]bool{},
			),
		)),
	)

	c.AddChild(dialogInput(&config.Conf.General.RomsPath, func() string {
		s := utils.OpenDirectory(l.Choose, config.Conf.General.RomsPath, config.Conf.General.RomsPath)
		config.Conf.General.RomsPath = s
		NewFileMenu(g, g.ui.PrevPageId)
		return s
	}))

	c.AddChild(NewNavBarButton(ui, l.Up,
		func(args *widget.ButtonClickedEventArgs) {
			p := getAbsPath(config.Conf.General.RomsPath)
			p = filepath.Dir(strings.Trim(p, "/"))

			config.Conf.General.RomsPath = p
			NewFileMenu(g, g.ui.PrevPageId)
		}),
	)
	c.AddChild(NewNavBarButton(ui, l.Return,
		func(args *widget.ButtonClickedEventArgs) {
			switch g.ui.PrevPageId {
			case PAGE_HOME:
				NewHome(g)
			case PAGE_PAUSE:
				NewPause(g)
			}

			file.Encode()
		}),
	)
	navbar.AddChild(NewHeader(l.OpenRom, ui.res), c)

	return navbar
}

func NewFileEntryButton(ui *Ui, text string, f func(*widget.ButtonClickedEventArgs)) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.TextPosition(widget.TextPositionStart, widget.TextPositionCenter),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),

			widget.WidgetOpts.MinSize(BUTTON_WIDTH, 0),
		),

		widget.ButtonOpts.ClickedHandler(f),
		widget.ButtonOpts.Text(
			text,
			ui.res.fonts.smallFace,
			&widget.ButtonTextColor{
				Idle: *ui.res.fgClr,
			},
		),
	)
}

func NewNavBarButton(ui *Ui, label string, f func(args *widget.ButtonClickedEventArgs)) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.TextPosition(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.ButtonOpts.ClickedHandler(f),
		widget.ButtonOpts.Text(
			label,
			ui.res.fonts.smallFace,
			&widget.ButtonTextColor{
				Idle: *ui.res.fgClr,
			},
		),
	)
}
