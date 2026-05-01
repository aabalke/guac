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
		res = g.ui.res
		tmp = config.Conf.General
		k   = &tmp.KeyboardConfig
		c   = &tmp.ControllerConfig
	)

	createSubMenu(parent,
		NewHeader("general", res),
		NewSeparator(),

		NewLabel("muted"),
		NewCheckbox(&tmp.Muted),

		NewLabel("show fps"),
		NewCheckbox(&tmp.ShowFps),

		NewLabel("initialize fullscreen"),
		NewCheckbox(&tmp.InitFullscreen),

		NewLabel("target fps"),
		NewTextBoxInput(&tmp.TargetFps, NumberValidation(100_000)),

		NewHeader("keyboard", res),
		NewSeparator(),

		NewLabel("select"),
		NewTextBoxInput(&k.Select, NoValidation()),
		NewLabel("return"),
		NewTextBoxInput(&k.Return, NoValidation()),
		NewLabel("mute"),
		NewTextBoxInput(&k.Mute, NoValidation()),
		NewLabel("pause"),
		NewTextBoxInput(&k.Pause, NoValidation()),
		NewLabel("left"),
		NewTextBoxInput(&k.Left, NoValidation()),
		NewLabel("right"),
		NewTextBoxInput(&k.Right, NoValidation()),
		NewLabel("up"),
		NewTextBoxInput(&k.Up, NoValidation()),
		NewLabel("down"),
		NewTextBoxInput(&k.Down, NoValidation()),
		NewLabel("fullscreen"),
		NewTextBoxInput(&k.Fullscreen, NoValidation()),
		NewLabel("quit"),
		NewTextBoxInput(&k.Quit, NoValidation()),

		NewHeader("controller", res),
		NewSeparator(),

		NewLabel("select"),
		NewTextBoxInput(&c.Select, NoValidation()),
		NewLabel("return"),
		NewTextBoxInput(&c.Return, NoValidation()),
		NewLabel("mute"),
		NewTextBoxInput(&c.Mute, NoValidation()),
		NewLabel("pause"),
		NewTextBoxInput(&c.Pause, NoValidation()),
		NewLabel("left"),
		NewTextBoxInput(&c.Left, NoValidation()),
		NewLabel("right"),
		NewTextBoxInput(&c.Right, NoValidation()),
		NewLabel("up"),
		NewTextBoxInput(&c.Up, NoValidation()),
		NewLabel("down"),
		NewTextBoxInput(&c.Down, NoValidation()),
		NewLabel("fullscreen"),
		NewTextBoxInput(&c.Fullscreen, NoValidation()),
		NewLabel("quit"),
		NewTextBoxInput(&c.Quit, NoValidation()),

		NewSaveButton(func(*widget.ButtonClickedEventArgs) {
			config.Conf.General = tmp
			config.Conf.DecodeGeneralController()
			if len(g.gamepadIds) != 0 {
				g.ui.focus.FocusLastSubMenu()
			}

			g.ui.toast.AddMessage("saved")

		}),
	)
}

func NewUiMenu(g *Game, parent *widget.Container) {

	var (
		res = g.ui.res
		tmp = config.Conf.Ui

		oldId = g.ui.PrevPageId

		clrInputs = [4]widget.PreferredSizeLocateableWidget{
			NewColorInput(&tmp.Backdrop, NoValidation()),
			NewColorInput(&tmp.MenuBackgroundColor, NoValidation()),
			NewColorInput(&tmp.MenuForegroundColor, NoValidation()),
			NewColorInput(&tmp.MenuSecondaryColor, NoValidation()),
		}
	)

	createSubMenu(parent,
		NewHeader("ui", res),
		NewSeparator(),

		NewLabel("backdrop"),
		clrInputs[0],

		NewLabel("bg color"),
		clrInputs[1],

		NewLabel("fg color"),
		clrInputs[2],

		NewLabel("accent color"),
		clrInputs[3],

		NewLabel("apply theme"),
		NewApplyPalettesMenu(&g.ui.focus.horizontalGroup, theme_palettes, clrInputs, res),

		NewSaveButton(func(*widget.ButtonClickedEventArgs) {
			config.Conf.Ui = tmp
			res.Update()
			NewSettings(g, oldId, MENU_UI)

			if len(g.gamepadIds) != 0 {
				g.ui.focus.FocusLastSubMenu()
			}
			g.ui.toast.AddMessage("saved")
		}),
	)
}

func NewGbMenu(g *Game, parent *widget.Container) {

	var (
		res = g.ui.res
		tmp = config.Conf.Gb
		key = &tmp.KeyboardConfig
		con = &tmp.ControllerConfig
		pal = &tmp.Palette

		clrInputs = [4]widget.PreferredSizeLocateableWidget{
			NewColorInput(&pal[0], NoValidation()),
			NewColorInput(&pal[1], NoValidation()),
			NewColorInput(&pal[2], NoValidation()),
			NewColorInput(&pal[3], NoValidation()),
		}
	)

	createSubMenu(parent,
		NewHeader("dmg palette", res),
		NewSeparator(),

		NewLabel("lightest"),
		clrInputs[0],
		NewLabel("light"),
		clrInputs[1],
		NewLabel("dark"),
		clrInputs[2],
		NewLabel("darkest"),
		clrInputs[3],

		NewLabel("apply palette"),
		NewApplyPalettesMenu(&g.ui.focus.horizontalGroup, dmg_palettes, clrInputs, res),

		NewHeader("keyboard", res),
		NewSeparator(),

		NewLabel("a"),
		NewTextBoxInput(&key.A, NoValidation()),
		NewLabel("b"),
		NewTextBoxInput(&key.B, NoValidation()),
		NewLabel("select"),
		NewTextBoxInput(&key.Select, NoValidation()),
		NewLabel("start"),
		NewTextBoxInput(&key.Start, NoValidation()),
		NewLabel("left"),
		NewTextBoxInput(&key.Left, NoValidation()),
		NewLabel("right"),
		NewTextBoxInput(&key.Right, NoValidation()),
		NewLabel("up"),
		NewTextBoxInput(&key.Up, NoValidation()),
		NewLabel("down"),
		NewTextBoxInput(&key.Down, NoValidation()),

		NewHeader("controller", res),
		NewSeparator(),

		NewLabel("a"),
		NewTextBoxInput(&con.A, NoValidation()),
		NewLabel("b"),
		NewTextBoxInput(&con.B, NoValidation()),
		NewLabel("select"),
		NewTextBoxInput(&con.Select, NoValidation()),
		NewLabel("start"),
		NewTextBoxInput(&con.Start, NoValidation()),
		NewLabel("left"),
		NewTextBoxInput(&con.Left, NoValidation()),
		NewLabel("right"),
		NewTextBoxInput(&con.Right, NoValidation()),
		NewLabel("up"),
		NewTextBoxInput(&con.Up, NoValidation()),
		NewLabel("down"),
		NewTextBoxInput(&con.Down, NoValidation()),
		NewSaveButton(func(*widget.ButtonClickedEventArgs) {
			config.Conf.Gb = tmp

			config.DecodeController(&config.Conf.Gb.ControllerConfig)

			if len(g.gamepadIds) != 0 {
				g.ui.focus.FocusLastSubMenu()
			}
			g.ui.toast.AddMessage("saved")
		}),
	)
}

func NewGbaMenu(g *Game, parent *widget.Container) {

	var (
		res = g.ui.res
		tmp = config.Conf.Gba
		key = &tmp.KeyboardConfig
		con = &tmp.ControllerConfig
	)

	createSubMenu(parent,

		NewHeader("general", res),
		NewSeparator(),

		NewLabel("optimize idle loops"),
		NewCheckbox(&tmp.IdleOptimize),

		NewLabel("snd clk cycles"),
		NewTextBoxInput(&tmp.SoundClockUpdateCycles, NoValidation()),

		NewHeader("keyboard", res),
		NewSeparator(),

		NewLabel("a"),
		NewTextBoxInput(&key.A, NoValidation()),
		NewLabel("b"),
		NewTextBoxInput(&key.B, NoValidation()),
		NewLabel("select"),
		NewTextBoxInput(&key.Select, NoValidation()),
		NewLabel("start"),
		NewTextBoxInput(&key.Start, NoValidation()),
		NewLabel("left"),
		NewTextBoxInput(&key.Left, NoValidation()),
		NewLabel("right"),
		NewTextBoxInput(&key.Right, NoValidation()),
		NewLabel("up"),
		NewTextBoxInput(&key.Up, NoValidation()),
		NewLabel("down"),
		NewTextBoxInput(&key.Down, NoValidation()),

		NewLabel("l"),
		NewTextBoxInput(&key.L, NoValidation()),
		NewLabel("r"),
		NewTextBoxInput(&key.R, NoValidation()),

		NewHeader("controller", res),
		NewSeparator(),

		NewLabel("a"),
		NewTextBoxInput(&con.A, NoValidation()),
		NewLabel("b"),
		NewTextBoxInput(&con.B, NoValidation()),
		NewLabel("select"),
		NewTextBoxInput(&con.Select, NoValidation()),
		NewLabel("start"),
		NewTextBoxInput(&con.Start, NoValidation()),
		NewLabel("left"),
		NewTextBoxInput(&con.Left, NoValidation()),
		NewLabel("right"),
		NewTextBoxInput(&con.Right, NoValidation()),
		NewLabel("up"),
		NewTextBoxInput(&con.Up, NoValidation()),
		NewLabel("down"),
		NewTextBoxInput(&con.Down, NoValidation()),
		NewLabel("l"),
		NewTextBoxInput(&con.L, NoValidation()),
		NewLabel("r"),
		NewTextBoxInput(&con.R, NoValidation()),
		NewSaveButton(func(*widget.ButtonClickedEventArgs) {
			config.Conf.Gba = tmp
			config.DecodeController(&config.Conf.Gba.ControllerConfig)
			if len(g.gamepadIds) != 0 {
				g.ui.focus.FocusLastSubMenu()
			}
			g.ui.toast.AddMessage("saved")
		}),
	)
}

func NewNdsMenu(g *Game, parent *widget.Container) {

	var (
		res = g.ui.res
		tmp = config.Conf.Nds
		key = &tmp.KeyboardConfig
		con = &tmp.ControllerConfig
	)

	createSubMenu(parent,

		NewHeader("screen", res),
		NewSeparator(),

		NewLabel("layout"),
		NewRadioInput(
			&g.ui.focus.horizontalGroup,
			&tmp.Screen.OLayout, []string{
				"vertical",
				"horizontal",
				"hybrid",
			}, res),

		NewLabel("sizing"),
		NewRadioInput(
			&g.ui.focus.horizontalGroup,
			&tmp.Screen.OSizing, []string{
				"even",
				"only top",
				"only bottom",
			}, res),

		NewLabel("rotation"),
		NewRadioInput(
			&g.ui.focus.horizontalGroup,
			&tmp.Screen.ORotation, []string{
				"0",
				"90",
				"180",
				"270",
			}, res),

		NewHeader("rtc", res),
		NewSeparator(),

		NewLabel("additional hours"),
		NewTextBoxInput(&tmp.Rtc.AdditionalHours, NoValidation()),

		NewHeader("bios", res),
		NewSeparator(),

		NewLabel("arm7 path"),
		NewFileInput(&tmp.Bios.Arm7Path),
		NewLabel("arm9 path"),
		NewFileInput(&tmp.Bios.Arm9Path),

		NewHeader("firmware", res),
		NewSeparator(),

		NewLabel("file path"),
		NewFileInput(&tmp.Firmware.FilePath),
		NewLabel("nickname"),
		NewTextBoxInput(&tmp.Firmware.Nickname, NoValidation()),
		NewLabel("message"),
		NewTextBoxInput(&tmp.Firmware.Message, NoValidation()),
		NewLabel("favorite color"),
		NewTextBoxInput(&tmp.Firmware.FavoriteColor, NoValidation()),

		NewHeader("scene export", res),
		NewSeparator(),

		NewLabel("output directory"),
		NewDirectoryInput(&tmp.Export.Directory, "./export"),
		NewLabel("shadow polygons"),
		NewCheckbox(&tmp.Export.ShadowPolys),

		NewHeader("keyboard", res),
		NewSeparator(),

		NewLabel("a"),
		NewTextBoxInput(&key.A, NoValidation()),
		NewLabel("b"),
		NewTextBoxInput(&key.B, NoValidation()),
		NewLabel("select"),
		NewTextBoxInput(&key.Select, NoValidation()),
		NewLabel("start"),
		NewTextBoxInput(&key.Start, NoValidation()),
		NewLabel("left"),
		NewTextBoxInput(&key.Left, NoValidation()),
		NewLabel("right"),
		NewTextBoxInput(&key.Right, NoValidation()),
		NewLabel("up"),
		NewTextBoxInput(&key.Up, NoValidation()),
		NewLabel("down"),
		NewTextBoxInput(&key.Down, NoValidation()),
		NewLabel("l"),
		NewTextBoxInput(&key.L, NoValidation()),
		NewLabel("r"),
		NewTextBoxInput(&key.R, NoValidation()),

		NewLabel("x"),
		NewTextBoxInput(&key.X, NoValidation()),
		NewLabel("y"),
		NewTextBoxInput(&key.Y, NoValidation()),
		NewLabel("hinge"),
		NewTextBoxInput(&key.Hinge, NoValidation()),
		NewLabel("debug"),
		NewTextBoxInput(&key.Debug, NoValidation()),
		NewLabel("layout toggle"),
		NewTextBoxInput(&key.LayoutToggle, NoValidation()),
		NewLabel("sizing toggle"),
		NewTextBoxInput(&key.SizingToggle, NoValidation()),
		NewLabel("rotation toggle"),
		NewTextBoxInput(&key.RotationToggle, NoValidation()),
		NewLabel("export toggle"),
		NewTextBoxInput(&key.ExportScene, NoValidation()),

		NewHeader("controller", res),
		NewSeparator(),

		NewLabel("a"),
		NewTextBoxInput(&con.A, NoValidation()),
		NewLabel("b"),
		NewTextBoxInput(&con.B, NoValidation()),
		NewLabel("select"),
		NewTextBoxInput(&con.Select, NoValidation()),
		NewLabel("start"),
		NewTextBoxInput(&con.Start, NoValidation()),
		NewLabel("left"),
		NewTextBoxInput(&con.Left, NoValidation()),
		NewLabel("right"),
		NewTextBoxInput(&con.Right, NoValidation()),
		NewLabel("up"),
		NewTextBoxInput(&con.Up, NoValidation()),
		NewLabel("down"),
		NewTextBoxInput(&con.Down, NoValidation()),
		NewLabel("l"),
		NewTextBoxInput(&con.L, NoValidation()),
		NewLabel("r"),
		NewTextBoxInput(&con.R, NoValidation()),
		NewLabel("x"),
		NewTextBoxInput(&con.X, NoValidation()),
		NewLabel("y"),
		NewTextBoxInput(&con.Y, NoValidation()),

		NewSaveButton(func(*widget.ButtonClickedEventArgs) {
			config.Conf.Nds = tmp
			config.DecodeController(&config.Conf.Nds.ControllerConfig)

			if len(g.gamepadIds) != 0 {
				g.ui.focus.FocusLastSubMenu()
			}
			g.ui.toast.AddMessage("saved")
		}),
	)
}
