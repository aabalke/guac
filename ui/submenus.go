package ui

import (
	"github.com/aabalke/guac/config"
	"github.com/ebitenui/ebitenui/widget"
)

const (
	MENU_GENERAL = iota
	MENU_UI
	MENU_GB
	MENU_GBA
	MENU_NDS
	MENU_RETURN
)

type SidebarField struct {
	label string
	f     func(g *Game)
}

var fields = []SidebarField{}

func init() {

	fields = []SidebarField{
		{
			"general",
			func(g *Game) {
				sub := g.ui.content
				NewGeneralMenu(g, sub)
				g.ui.focus.submenu = sub.GetFocusers()
				g.ui.focus.BuildFocus(g.ui.ui)
			}},
		{
			"ui",
			func(g *Game) {
				sub := g.ui.content
				NewUiMenu(g, sub)
				g.ui.focus.submenu = sub.GetFocusers()
				g.ui.focus.BuildFocus(g.ui.ui)
			}},
		{
			"gb",
			func(g *Game) {
				sub := g.ui.content
				NewGbMenu(g, sub)
				g.ui.focus.submenu = sub.GetFocusers()
				g.ui.focus.BuildFocus(g.ui.ui)
			}},
		{
			"gba",
			func(g *Game) {
				sub := g.ui.content
				NewGbaMenu(g, sub)
				g.ui.focus.submenu = sub.GetFocusers()
				g.ui.focus.BuildFocus(g.ui.ui)
			}},
		{
			"nds",
			func(g *Game) {
				sub := g.ui.content
				NewNdsMenu(g, sub)
				g.ui.focus.submenu = sub.GetFocusers()
				g.ui.focus.BuildFocus(g.ui.ui)
			}},
		{
			"return",
			func(g *Game) {
				switch g.ui.PrevPageId {
				case PAGE_HOME:
					NewHome(g)
				case PAGE_PAUSE:
					NewPause(g)
				}
			}},
	}
}

func NewGeneralMenu(g *Game, parent *widget.Container) {

	var (
		res     = g.ui.res
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
			config.Conf.DecodeGeneralController()
			if len(g.gamepadIds) != 0 {
				g.ui.focus.FocusLastSubMenu()
			}

		}),
	)
}

func NewUiMenu(g *Game, parent *widget.Container) {

	var (
		res     = g.ui.res
		tmp     = config.Conf.Ui
		face    = res.fonts.smallFace
		bigFace = res.fonts.face
		clr     = config.Conf.Ui.MenuForegroundColor

		oldId = g.ui.PrevPageId

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
		NewApplyPalettesMenu(&g.ui.focus.horizontalGroup, theme_palettes, clrInputs, face, res.buttonImage),

		NewSaveButton(res.fonts.face, res.buttonImage, func(*widget.ButtonClickedEventArgs) {
			config.Conf.Ui = tmp
			res.Update()
			NewSettings(g, oldId, MENU_UI)

			if len(g.gamepadIds) != 0 {
				g.ui.focus.FocusLastSubMenu()
			}
		}),
	)
}

func NewGbMenu(g *Game, parent *widget.Container) {

	var (
		res     = g.ui.res
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
		NewApplyPalettesMenu(&g.ui.focus.horizontalGroup, dmg_palettes, clrInputs, face, res.buttonImage),

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

			config.DecodeController(&config.Conf.Gb.ControllerConfig)

			if len(g.gamepadIds) != 0 {
				g.ui.focus.FocusLastSubMenu()
			}
		}),
	)
}

func NewGbaMenu(g *Game, parent *widget.Container) {

	var (
		res     = g.ui.res
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
			config.DecodeController(&config.Conf.Gba.ControllerConfig)
			if len(g.gamepadIds) != 0 {
				g.ui.focus.FocusLastSubMenu()
			}
		}),
	)
}

func NewNdsMenu(g *Game, parent *widget.Container) {

	var (
		res     = g.ui.res
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

		NewLabel("layout", face, clr),
		NewRadioInput(
			&g.ui.focus.horizontalGroup,
			&tmp.Screen.OLayout,
			[]string{
				"vertical",
				"horizontal",
				"hybrid",
			}, face, res.buttonImage),

		NewLabel("sizing", face, clr),
		NewRadioInput(
			&g.ui.focus.horizontalGroup,
			&tmp.Screen.OSizing, []string{
				"even",
				"only top",
				"only bottom",
			}, face, res.buttonImage),

		NewLabel("rotation", face, clr),
		NewRadioInput(
			&g.ui.focus.horizontalGroup,
			&tmp.Screen.ORotation, []string{
				"0",
				"90",
				"180",
				"270",
			}, face, res.buttonImage),

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

		NewLabel("scene export", bigFace, clr),
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
			config.DecodeController(&config.Conf.Nds.ControllerConfig)

			if len(g.gamepadIds) != 0 {
				g.ui.focus.FocusLastSubMenu()
			}
		}),
	)
}
