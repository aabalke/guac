package ui

import (
	"image/color"
	"math"

	"github.com/aabalke/guac/config"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	MENU_GENERAL = "general"
	MENU_UI      = "ui"
	MENU_GB      = "gb"
	MENU_GBA     = "gba"
	MENU_NDS     = "nds"
	MENU_RETURN  = "return"
)

var opt = widget.ContainerOpts

func NewSettings(g *Game, oldId PageId, initMenu string) *UiPage {

	res := g.res

	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.bg),
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

	sub := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(24)),
			widget.GridLayoutOpts.Spacing(32, 16),
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, true}, []bool{}),
		)),
	)

	scr, resetFunc := NewScrollableContainer(res, sub)
	sidebar := NewSidebar(g, initMenu, resetFunc, sub, oldId)

	c.AddChild(sidebar, scr)
	root.AddChild(c)

	ui := &UiPage{
		Id: PAGE_SETTINGS,
		ui: &ebitenui.UI{
			Container: root,
		}}

	//ui.SetDebugMode(true)

	return ui
}

func NewScrollableContainer(res *UiResources, content *widget.Container) (*widget.Container, func()) {

	clr := image.NewNineSliceColor(config.Conf.Ui.MenuForegroundColor)

	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Spacing(2, 0),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{true}),
		)),
	)

	var slider *widget.Slider

	scrollable := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(content),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Mask: image.NewNineSliceColor(color.Black),
			Idle: transparentNine,
		}),
	)

	resetFunc := func() {
		scrollable.ScrollTop = 0
		slider.Current = 0
	}

	pageSizeFunc := func() int {

		scrollableHeight := scrollable.ViewRect().Dy()
		contentHeight := content.GetWidget().Rect.Dy()

		if scrollableHeight >= contentHeight {
			resetFunc()
			return 0
		}

		return int(math.Round(
			float64(scrollableHeight) /
				float64(contentHeight) *
				1000))
	}

	slider = widget.NewSlider(
		widget.SliderOpts.InitialCurrent(0),
		widget.SliderOpts.Orientation(widget.DirectionVertical),
		widget.SliderOpts.MinMax(0, 1000),
		widget.SliderOpts.PageSizeFunc(pageSizeFunc),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			scrollable.ScrollTop = float64(args.Slider.Current) / 1000
		}),
		widget.SliderOpts.Images(
			&widget.SliderTrackImage{
				Idle:  transparentNine,
				Hover: transparentNine,
			},
			&widget.ButtonImage{
				Idle:    res.sec,
				Hover:   clr,
				Pressed: clr,
			},
		),

		widget.SliderOpts.WidgetOpts(
			widget.WidgetOpts.OnUpdate(func(widget.HasWidget) {
				scrollableHeight := scrollable.ViewRect().Dy()
				contentHeight := content.GetWidget().Rect.Dy()

				if scrollableHeight >= contentHeight {
					slider.GetWidget().SetVisibility(widget.Visibility_Hide)
					return
				}
				slider.GetWidget().SetVisibility(widget.Visibility_Show)
			}),
		),
	)

	scrollable.GetWidget().ScrolledEvent.AddHandler(func(args any) {

		if scrollable.ViewRect().Dy() >= content.GetWidget().Rect.Dy() {
			resetFunc()
			return
		}

		if a, ok := args.(*widget.WidgetScrolledEventArgs); ok {
			slider.Current -= int(math.Round(a.Y * float64(pageSizeFunc())))
		}
	})

	root.AddChild(scrollable, slider)

	return root, resetFunc
}

func NewSidebar(g *Game, initMenu string, resetFunc func(), sub *widget.Container, oldId PageId) *widget.Container {

	var (
		res    = g.res
		face   = res.fonts.face
		img    = res.buttonImage
		fields = map[string]func(){
			"general": func() {
				NewGeneralMenu(sub, res)
			},
			"ui": func() {
				NewUiMenu(g, oldId, sub, res)
			},
			"gb": func() {
				NewGbMenu(sub, res)
			},
			"gba": func() {
				NewGbaMenu(sub, res)
			},
			"nds": func() {
				NewNdsMenu(sub, res)
			},
			"return": func() {
				switch oldId {
				case PAGE_HOME:
					g.ui = NewHome(g)
				case PAGE_PAUSE:
					g.ui = NewPause(g)
				}
			},
		}
		order = []string{
			MENU_GENERAL,
			MENU_UI,
			MENU_GB,
			MENU_GBA,
			MENU_NDS,
			MENU_RETURN,
		}
	)

	sidebar := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(24)),
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		)),
	)

	bs := []widget.RadioGroupElement{}
	init := 0
	for i, label := range order {
		b := NewSidebarButton(label, fields[label], resetFunc, face, img)
		sidebar.AddChild(b)
		bs = append(bs, b)

		if label == initMenu {
			init = i
		}
	}

	widget.NewRadioGroup(
		widget.RadioGroupOpts.Elements(bs...),
		widget.RadioGroupOpts.InitialElement(bs[init]),
	)

	// initialize first menu
	fields[initMenu]()

	return sidebar
}

func NewSidebarButton(label string, f, resetFunc func(), face *text.Face, img *widget.ButtonImage) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				MaxWidth: BUTTON_WIDTH,
				Stretch:  true,
			}),
		),

		widget.ButtonOpts.Image(img),

		widget.ButtonOpts.Text(
			label,
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

		widget.ButtonOpts.TextPosition(
			widget.TextPositionStart,
			widget.TextPositionCenter,
		),

		widget.ButtonOpts.ClickedHandler(
			func(*widget.ButtonClickedEventArgs) {
				f()
				resetFunc()
			},
		),
	)
}

func createSubMenu(parent *widget.Container, children ...widget.PreferredSizeLocateableWidget) {
	parent.RemoveChildren()
	for _, child := range children {
		parent.AddChild(child)
	}
}

func NewGeneralMenu(parent *widget.Container, res *UiResources) {

	var (
		tmp     = config.Conf.General
		k       = &tmp.KeyboardConfig
		c       = &tmp.ControllerConfig
		face    = res.fonts.smallFace
		bigFace = res.fonts.face

		clr = config.Conf.Ui.MenuForegroundColor
	)

	createSubMenu(parent,
		NewLabel("general", bigFace, clr),
		NewSeparator(),

		NewLabel("muted", face, clr),
		NewCheckbox(&tmp.Muted, res.checkbox),

		NewLabel("show fps", face, clr),
		NewCheckbox(&tmp.ShowFps, res.checkbox),

		NewLabel("initialize fullscreen", face, clr),
		NewCheckbox(&tmp.InitFullscreen, res.checkbox),

		NewLabel("target fps", face, clr),
		NewTextBoxInput(&tmp.TargetFps, face, NumberValidation(100_000)),

		NewLabel("keyboard", bigFace, clr),
		NewSeparator(),

		NewLabel("select", face, clr),
		NewTextBoxInput(&k.Select, face, NoValidation()),
		NewLabel("mute", face, clr),
		NewTextBoxInput(&k.Mute, face, NoValidation()),
		NewLabel("pause", face, clr),
		NewTextBoxInput(&k.Pause, face, NoValidation()),
		NewLabel("left", face, clr),
		NewTextBoxInput(&k.Left, face, NoValidation()),
		NewLabel("right", face, clr),
		NewTextBoxInput(&k.Right, face, NoValidation()),
		NewLabel("up", face, clr),
		NewTextBoxInput(&k.Up, face, NoValidation()),
		NewLabel("down", face, clr),
		NewTextBoxInput(&k.Down, face, NoValidation()),
		NewLabel("fullscreen", face, clr),
		NewTextBoxInput(&k.Fullscreen, face, NoValidation()),
		NewLabel("quit", face, clr),
		NewTextBoxInput(&k.Quit, face, NoValidation()),

		NewLabel("controller", bigFace, clr),
		NewSeparator(),

		NewLabel("select", face, clr),
		NewTextBoxInput(&c.Select, face, NoValidation()),
		NewLabel("mute", face, clr),
		NewTextBoxInput(&c.Mute, face, NoValidation()),
		NewLabel("pause", face, clr),
		NewTextBoxInput(&c.Pause, face, NoValidation()),
		NewLabel("left", face, clr),
		NewTextBoxInput(&c.Left, face, NoValidation()),
		NewLabel("right", face, clr),
		NewTextBoxInput(&c.Right, face, NoValidation()),
		NewLabel("up", face, clr),
		NewTextBoxInput(&c.Up, face, NoValidation()),
		NewLabel("down", face, clr),
		NewTextBoxInput(&c.Down, face, NoValidation()),
		NewLabel("fullscreen", face, clr),
		NewTextBoxInput(&c.Fullscreen, face, NoValidation()),
		NewLabel("quit", face, clr),
		NewTextBoxInput(&c.Quit, face, NoValidation()),

		NewSaveButton(res.fonts.face, res.buttonImage, func(*widget.ButtonClickedEventArgs) {
			config.Conf.General = tmp
		}),
	)
}

func NewUiMenu(g *Game, oldId PageId, parent *widget.Container, res *UiResources) {

	var (
		tmp     = config.Conf.Ui
		face    = res.fonts.smallFace
		bigFace = res.fonts.face
		clr     = config.Conf.Ui.MenuForegroundColor

		clrInputs = [4]widget.PreferredSizeLocateableWidget{
			NewColorInput(&tmp.Backdrop, face, NoValidation()),
			NewColorInput(&tmp.MenuBackgroundColor, face, NoValidation()),
			NewColorInput(&tmp.MenuForegroundColor, face, NoValidation()),
			NewColorInput(&tmp.MenuSecondaryColor, face, NoValidation()),
		}
	)

	createSubMenu(parent,
		NewLabel("ui", bigFace, clr),
		NewSeparator(),

		NewLabel("backdrop", face, clr),
		clrInputs[0],

		NewLabel("bg color", face, clr),
		clrInputs[1],

		NewLabel("fg color", face, clr),
		clrInputs[2],

		NewLabel("accent color", face, clr),
		clrInputs[3],

		NewLabel("apply theme", face, clr),
		NewApplyPalettesMenu(theme_palettes, clrInputs, face, res.buttonImage),

		NewSaveButton(res.fonts.face, res.buttonImage, func(*widget.ButtonClickedEventArgs) {
			config.Conf.Ui = tmp
			res.Update()

			g.ui = NewSettings(g, oldId, "ui")
		}),
	)
}

func NewGbMenu(parent *widget.Container, res *UiResources) {

	var (
		tmp     = config.Conf.Gb
		key     = &tmp.KeyboardConfig
		con     = &tmp.ControllerConfig
		pal     = &tmp.Palette
		face    = res.fonts.smallFace
		bigFace = res.fonts.face

		clrInputs = [4]widget.PreferredSizeLocateableWidget{
			NewColorInput(&pal[0], face, NoValidation()),
			NewColorInput(&pal[1], face, NoValidation()),
			NewColorInput(&pal[2], face, NoValidation()),
			NewColorInput(&pal[3], face, NoValidation()),
		}

		clr = config.Conf.Ui.MenuForegroundColor
	)

	createSubMenu(parent,
		NewLabel("dmg palette", bigFace, clr),
		NewSeparator(),

		NewLabel("lightest", face, clr),
		clrInputs[0],
		NewLabel("light", face, clr),
		clrInputs[1],
		NewLabel("dark", face, clr),
		clrInputs[2],
		NewLabel("darkest", face, clr),
		clrInputs[3],

		NewLabel("apply palette", face, clr),
		NewApplyPalettesMenu(dmg_palettes, clrInputs, face, res.buttonImage),

		NewLabel("keyboard", bigFace, clr),
		NewSeparator(),

		NewLabel("a", face, clr),
		NewTextBoxInput(&key.A, face, NoValidation()),
		NewLabel("b", face, clr),
		NewTextBoxInput(&key.B, face, NoValidation()),
		NewLabel("select", face, clr),
		NewTextBoxInput(&key.Select, face, NoValidation()),
		NewLabel("start", face, clr),
		NewTextBoxInput(&key.Start, face, NoValidation()),
		NewLabel("left", face, clr),
		NewTextBoxInput(&key.Left, face, NoValidation()),
		NewLabel("right", face, clr),
		NewTextBoxInput(&key.Right, face, NoValidation()),
		NewLabel("up", face, clr),
		NewTextBoxInput(&key.Up, face, NoValidation()),
		NewLabel("down", face, clr),
		NewTextBoxInput(&key.Down, face, NoValidation()),

		NewLabel("controller", bigFace, clr),
		NewSeparator(),

		NewLabel("a", face, clr),
		NewTextBoxInput(&con.A, face, NoValidation()),
		NewLabel("b", face, clr),
		NewTextBoxInput(&con.B, face, NoValidation()),
		NewLabel("select", face, clr),
		NewTextBoxInput(&con.Select, face, NoValidation()),
		NewLabel("start", face, clr),
		NewTextBoxInput(&con.Start, face, NoValidation()),
		NewLabel("left", face, clr),
		NewTextBoxInput(&con.Left, face, NoValidation()),
		NewLabel("right", face, clr),
		NewTextBoxInput(&con.Right, face, NoValidation()),
		NewLabel("up", face, clr),
		NewTextBoxInput(&con.Up, face, NoValidation()),
		NewLabel("down", face, clr),
		NewTextBoxInput(&con.Down, face, NoValidation()),
		NewSaveButton(res.fonts.face, res.buttonImage, func(*widget.ButtonClickedEventArgs) {
			config.Conf.Gb = tmp
		}),
	)
}

func NewGbaMenu(parent *widget.Container, res *UiResources) {

	var (
		tmp     = config.Conf.Gba
		key     = &tmp.KeyboardConfig
		con     = &tmp.ControllerConfig
		face    = res.fonts.smallFace
		bigFace = res.fonts.face
		clr     = config.Conf.Ui.MenuForegroundColor
	)

	createSubMenu(parent,

		NewLabel("general", bigFace, clr),
		NewSeparator(),

		NewLabel("optimize idle loops", face, clr),
		NewCheckbox(&tmp.IdleOptimize, res.checkbox),

		NewLabel("snd clk cycles", face, clr),
		NewTextBoxInput(&tmp.SoundClockUpdateCycles, face, NoValidation()),

		NewLabel("keyboard", bigFace, clr),
		NewSeparator(),

		NewLabel("a", face, clr),
		NewTextBoxInput(&key.A, face, NoValidation()),
		NewLabel("b", face, clr),
		NewTextBoxInput(&key.B, face, NoValidation()),
		NewLabel("select", face, clr),
		NewTextBoxInput(&key.Select, face, NoValidation()),
		NewLabel("start", face, clr),
		NewTextBoxInput(&key.Start, face, NoValidation()),
		NewLabel("left", face, clr),
		NewTextBoxInput(&key.Left, face, NoValidation()),
		NewLabel("right", face, clr),
		NewTextBoxInput(&key.Right, face, NoValidation()),
		NewLabel("up", face, clr),
		NewTextBoxInput(&key.Up, face, NoValidation()),
		NewLabel("down", face, clr),
		NewTextBoxInput(&key.Down, face, NoValidation()),

		NewLabel("l", face, clr),
		NewTextBoxInput(&key.L, face, NoValidation()),
		NewLabel("r", face, clr),
		NewTextBoxInput(&key.R, face, NoValidation()),

		NewLabel("controller", bigFace, clr),
		NewSeparator(),

		NewLabel("a", face, clr),
		NewTextBoxInput(&con.A, face, NoValidation()),
		NewLabel("b", face, clr),
		NewTextBoxInput(&con.B, face, NoValidation()),
		NewLabel("select", face, clr),
		NewTextBoxInput(&con.Select, face, NoValidation()),
		NewLabel("start", face, clr),
		NewTextBoxInput(&con.Start, face, NoValidation()),
		NewLabel("left", face, clr),
		NewTextBoxInput(&con.Left, face, NoValidation()),
		NewLabel("right", face, clr),
		NewTextBoxInput(&con.Right, face, NoValidation()),
		NewLabel("up", face, clr),
		NewTextBoxInput(&con.Up, face, NoValidation()),
		NewLabel("down", face, clr),
		NewTextBoxInput(&con.Down, face, NoValidation()),
		NewLabel("l", face, clr),
		NewTextBoxInput(&con.L, face, NoValidation()),
		NewLabel("r", face, clr),
		NewTextBoxInput(&con.R, face, NoValidation()),
		NewSaveButton(res.fonts.face, res.buttonImage, func(*widget.ButtonClickedEventArgs) {
			config.Conf.Gba = tmp
		}),
	)
}

func NewNdsMenu(parent *widget.Container, res *UiResources) {

	var (
		tmp     = config.Conf.Nds
		key     = &tmp.KeyboardConfig
		con     = &tmp.ControllerConfig
		face    = res.fonts.smallFace
		bigFace = res.fonts.face
		clr     = config.Conf.Ui.MenuForegroundColor
	)

	createSubMenu(parent,

		NewLabel("screen", bigFace, clr),
		NewSeparator(),

		//NewLabel("rotation", bigFace, clr),
		//NewRadioInput(tmp.Screen.ORotation, face, img),

		NewLabel("rtc", bigFace, clr),
		NewSeparator(),

		NewLabel("additional hours", face, clr),
		NewTextBoxInput(&tmp.Rtc.AdditionalHours, face, NoValidation()),

		NewLabel("bios", bigFace, clr),
		NewSeparator(),

		NewLabel("arm7 path", face, clr),
		NewFileInput(&tmp.Bios.Arm7Path, face),
		NewLabel("arm9 path", face, clr),
		NewFileInput(&tmp.Bios.Arm9Path, face),

		NewLabel("firmware", bigFace, clr),
		NewSeparator(),

		NewLabel("file path", face, clr),
		NewFileInput(&tmp.Firmware.FilePath, face),
		NewLabel("nickname", face, clr),
		NewTextBoxInput(&tmp.Firmware.Nickname, face, NoValidation()),
		NewLabel("message", face, clr),
		NewTextBoxInput(&tmp.Firmware.Message, face, NoValidation()),
		NewLabel("favorite color", face, clr),
		NewTextBoxInput(&tmp.Firmware.FavoriteColor, face, NoValidation()),

		NewLabel("3d scene exporter", bigFace, clr),
		NewSeparator(),

		NewLabel("output directory", face, clr),
		NewDirectoryInput(&tmp.Export.Directory, face, "./export"),
		NewLabel("shadow polygons", face, clr),
		NewCheckbox(&tmp.Export.ShadowPolys, res.checkbox),

		NewLabel("keyboard", bigFace, clr),
		NewSeparator(),

		NewLabel("a", face, clr),
		NewTextBoxInput(&key.A, face, NoValidation()),
		NewLabel("b", face, clr),
		NewTextBoxInput(&key.B, face, NoValidation()),
		NewLabel("select", face, clr),
		NewTextBoxInput(&key.Select, face, NoValidation()),
		NewLabel("start", face, clr),
		NewTextBoxInput(&key.Start, face, NoValidation()),
		NewLabel("left", face, clr),
		NewTextBoxInput(&key.Left, face, NoValidation()),
		NewLabel("right", face, clr),
		NewTextBoxInput(&key.Right, face, NoValidation()),
		NewLabel("up", face, clr),
		NewTextBoxInput(&key.Up, face, NoValidation()),
		NewLabel("down", face, clr),
		NewTextBoxInput(&key.Down, face, NoValidation()),
		NewLabel("l", face, clr),
		NewTextBoxInput(&key.L, face, NoValidation()),
		NewLabel("r", face, clr),
		NewTextBoxInput(&key.R, face, NoValidation()),

		NewLabel("x", face, clr),
		NewTextBoxInput(&key.X, face, NoValidation()),
		NewLabel("y", face, clr),
		NewTextBoxInput(&key.Y, face, NoValidation()),
		NewLabel("hinge", face, clr),
		NewTextBoxInput(&key.Hinge, face, NoValidation()),
		NewLabel("debug", face, clr),
		NewTextBoxInput(&key.Debug, face, NoValidation()),
		NewLabel("layout toggle", face, clr),
		NewTextBoxInput(&key.LayoutToggle, face, NoValidation()),
		NewLabel("sizing toggle", face, clr),
		NewTextBoxInput(&key.SizingToggle, face, NoValidation()),
		NewLabel("rotation toggle", face, clr),
		NewTextBoxInput(&key.RotationToggle, face, NoValidation()),
		NewLabel("export toggle", face, clr),
		NewTextBoxInput(&key.ExportScene, face, NoValidation()),

		NewLabel("controller", bigFace, clr),
		NewSeparator(),

		NewLabel("a", face, clr),
		NewTextBoxInput(&con.A, face, NoValidation()),
		NewLabel("b", face, clr),
		NewTextBoxInput(&con.B, face, NoValidation()),
		NewLabel("select", face, clr),
		NewTextBoxInput(&con.Select, face, NoValidation()),
		NewLabel("start", face, clr),
		NewTextBoxInput(&con.Start, face, NoValidation()),
		NewLabel("left", face, clr),
		NewTextBoxInput(&con.Left, face, NoValidation()),
		NewLabel("right", face, clr),
		NewTextBoxInput(&con.Right, face, NoValidation()),
		NewLabel("up", face, clr),
		NewTextBoxInput(&con.Up, face, NoValidation()),
		NewLabel("down", face, clr),
		NewTextBoxInput(&con.Down, face, NoValidation()),
		NewLabel("l", face, clr),
		NewTextBoxInput(&con.L, face, NoValidation()),
		NewLabel("r", face, clr),
		NewTextBoxInput(&con.R, face, NoValidation()),
		NewLabel("x", face, clr),
		NewTextBoxInput(&con.X, face, NoValidation()),
		NewLabel("y", face, clr),
		NewTextBoxInput(&con.Y, face, NoValidation()),

		NewSaveButton(res.fonts.face, res.buttonImage, func(*widget.ButtonClickedEventArgs) {
			config.Conf.Nds = tmp
		}),
	)
}
