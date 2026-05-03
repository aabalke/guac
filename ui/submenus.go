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

type Field struct {
	widgettype int
	label      string
	sublabel   string
	ptr        any
	other      any
}

type SidebarField struct {
	label string
	f     func(g *Game)
}

var sidebarFields []SidebarField

func init() {
	sidebarFields = []SidebarField{
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

func buildSubMenu(g *Game, parent *widget.Container, fields []Field) {
	for _, field := range fields {
		switch field.widgettype {
		case WIDGET_HDR:
			parent.AddChild(NewHeader(field.label, g.ui.res), NewSeparator())
		case WIDGET_CBX:
			parent.AddChild(NewLabel(field.label), NewCheckbox(field.ptr.(*bool)))
		case WIDGET_KEY:
			parent.AddChild(NewLabel(field.label), NewKeybindInput(g.ui, field.sublabel, field.ptr))
		case WIDGET_DEC:
			parent.AddChild(NewLabel(field.label), NewDecimalInput(g.ui, field.sublabel, field.ptr, field.other.(int)))
		case WIDGET_HEX:
			parent.AddChild(NewLabel(field.label), NewHexInput(g.ui, field.sublabel, field.ptr, field.other.(int)))
		case WIDGET_FLE:
			parent.AddChild(NewLabel(field.label), NewFileInput(field.ptr.(*string)))
		case WIDGET_DIR:
			parent.AddChild(NewLabel(field.label), NewDirectoryInput(field.ptr.(*string), field.other.(string)))
		case WIDGET_TXT:
			parent.AddChild(NewLabel(field.label), NewTextBoxInput(g.ui, BOARD_ALPHA, field.sublabel, field.ptr, NoValidation()))
		case WIDGET_RAD:
			parent.AddChild(NewLabel(field.label), NewRadioInput(
				&g.ui.focus.horizontalGroup,
				field.ptr.(*int),
				field.other.([]string),
				g.ui.res,
			))
		}
	}
}

func NewGeneralMenu(g *Game, parent *widget.Container) {

	var (
		tmp = config.Conf.General
		k   = &tmp.KeyboardConfig
		c   = &tmp.ControllerConfig
	)

	fields := []Field{
		{WIDGET_HDR, "general", "", nil, nil},
		{WIDGET_CBX, "muted", "", &tmp.Muted, nil},
		{WIDGET_CBX, "show fps", "", &tmp.ShowFps, nil},
		{WIDGET_CBX, "init fullscreen", "", &tmp.InitFullscreen, nil},
		{WIDGET_DEC, "target fps", "target fps", &tmp.TargetFps, 1000000},

		{WIDGET_HDR, "keyboard", "", nil, nil},
		{WIDGET_KEY, "select", "keyboard select", &k.Select, nil},
		{WIDGET_KEY, "return", "keyboard return", &k.Return, nil},
		{WIDGET_KEY, "mute", "keyboard mute", &k.Mute, nil},
		{WIDGET_KEY, "pause", "keyboard pause", &k.Pause, nil},
		{WIDGET_KEY, "left", "keyboard left", &k.Left, nil},
		{WIDGET_KEY, "right", "keyboard right", &k.Right, nil},
		{WIDGET_KEY, "up", "keyboard up", &k.Up, nil},
		{WIDGET_KEY, "down", "keyboard down", &k.Down, nil},
		{WIDGET_KEY, "fullscreen", "keyboard fullscreen", &k.Fullscreen, nil},
		{WIDGET_KEY, "quit", "keyboard quit", &k.Quit, nil},

		{WIDGET_HDR, "controller", "", nil, nil},
		{WIDGET_KEY, "select", "controller select", &c.Select, nil},
		{WIDGET_KEY, "return", "controller return", &c.Return, nil},
		{WIDGET_KEY, "mute", "controller mute", &c.Mute, nil},
		{WIDGET_KEY, "pause", "controller pause", &c.Pause, nil},
		{WIDGET_KEY, "left", "controller left", &c.Left, nil},
		{WIDGET_KEY, "right", "controller right", &c.Right, nil},
		{WIDGET_KEY, "up", "controller up", &c.Up, nil},
		{WIDGET_KEY, "down", "controller down", &c.Down, nil},
		{WIDGET_KEY, "fullscreen", "controller fullscreen", &c.Fullscreen, nil},
		{WIDGET_KEY, "quit", "controller quit", &c.Quit, nil},
	}

	parent.RemoveChildren()
	buildSubMenu(g, parent, fields)

	parent.AddChild(NewSaveButton(func(*widget.ButtonClickedEventArgs) {
		config.Conf.General = tmp

		if len(g.gamepadIds) != 0 {
			g.ui.focus.FocusLastSubMenu()
		}

		g.ui.toast.AddMessage("saved")
	}))
}

func NewUiMenu(g *Game, parent *widget.Container) {

	var (
		res   = g.ui.res
		tmp   = config.Conf.Ui
		oldId = g.ui.PrevPageId

		clrInputs = [4]widget.PreferredSizeLocateableWidget{
			NewColorInput(g.ui, "ui backdrop", &tmp.Backdrop, HexValidation(0xFFFFFF)),
			NewColorInput(g.ui, "ui bg color", &tmp.MenuBackgroundColor, HexValidation(0xFFFFFF)),
			NewColorInput(g.ui, "ui fg color", &tmp.MenuForegroundColor, HexValidation(0xFFFFFF)),
			NewColorInput(g.ui, "ui accent color", &tmp.MenuSecondaryColor, HexValidation(0xFFFFFF)),
		}
	)
	parent.RemoveChildren()

	parent.AddChild(
		NewHeader("ui", res), NewSeparator(),
		NewLabel("backdrop"), clrInputs[0],
		NewLabel("bg color"), clrInputs[1],
		NewLabel("fg color"), clrInputs[2],
		NewLabel("accent color"), clrInputs[3],
		NewLabel("apply theme"),
		NewApplyPalettesMenu(&g.ui.focus.horizontalGroup, theme_palettes, clrInputs, res),
		NewSaveButton(func(*widget.ButtonClickedEventArgs) {
			config.Conf.Ui = tmp
			res.Update()
			NewSettings(g, oldId, MENU_UI)
			// should this be somewhere else?
			g.ui.keyboard = NewKeyboard(g.ui.res)

			if len(g.gamepadIds) != 0 {
				g.ui.focus.FocusLastSubMenu()
			}

			g.ui.toast.AddMessage("saved")
		}))
}

func NewGbMenu(g *Game, parent *widget.Container) {

	var (
		tmp = config.Conf.Gb
		k   = &tmp.KeyboardConfig
		c   = &tmp.ControllerConfig
		pal = &tmp.Palette

		clrInputs = [4]widget.PreferredSizeLocateableWidget{
			NewColorInput(g.ui, "dmg lightest", &pal[0], HexValidation(0xFFFFFF)),
			NewColorInput(g.ui, "dmg light", &pal[1], HexValidation(0xFFFFFF)),
			NewColorInput(g.ui, "dmg dark", &pal[2], HexValidation(0xFFFFFF)),
			NewColorInput(g.ui, "dmg darkest", &pal[3], HexValidation(0xFFFFFF)),
		}
	)

	fields := []Field{
		{WIDGET_HDR, "keyboard", "", nil, nil},
		{WIDGET_KEY, "a", "gb keyboard a", &k.A, nil},
		{WIDGET_KEY, "b", "gb keyboard b", &k.B, nil},
		{WIDGET_KEY, "select", "gb keyboard select", &k.Select, nil},
		{WIDGET_KEY, "start", "gb keyboard start", &k.Start, nil},
		{WIDGET_KEY, "left", "gb keyboard left", &k.Left, nil},
		{WIDGET_KEY, "right", "gb keyboard right", &k.Right, nil},
		{WIDGET_KEY, "up", "gb keyboard up", &k.Up, nil},
		{WIDGET_KEY, "down", "gb keyboard down", &k.Down, nil},

		{WIDGET_HDR, "controller", "", nil, nil},
		{WIDGET_KEY, "a", "gb controller a", &c.A, nil},
		{WIDGET_KEY, "b", "gb controller b", &c.B, nil},
		{WIDGET_KEY, "select", "gb controller select", &c.Select, nil},
		{WIDGET_KEY, "start", "gb controller start", &c.Start, nil},
		{WIDGET_KEY, "left", "gb controller left", &c.Left, nil},
		{WIDGET_KEY, "right", "gb controller right", &c.Right, nil},
		{WIDGET_KEY, "up", "gb controller up", &c.Up, nil},
		{WIDGET_KEY, "down", "gb controller down", &c.Down, nil},
	}

	parent.RemoveChildren()

	parent.AddChild(
		NewHeader("dmg palette", g.ui.res), NewSeparator(),
		NewLabel("lightest"), clrInputs[0],
		NewLabel("light"), clrInputs[1],
		NewLabel("dark"), clrInputs[2],
		NewLabel("darkest"), clrInputs[3],
		NewLabel("apply palette"),
		NewApplyPalettesMenu(&g.ui.focus.horizontalGroup, dmg_palettes, clrInputs, g.ui.res),
	)

	buildSubMenu(g, parent, fields)

	parent.AddChild(NewSaveButton(func(*widget.ButtonClickedEventArgs) {
		config.Conf.Gb = tmp
		if len(g.gamepadIds) != 0 {
			g.ui.focus.FocusLastSubMenu()
		}
		g.ui.toast.AddMessage("saved")
	}))
}

func NewGbaMenu(g *Game, parent *widget.Container) {

	var (
		tmp = config.Conf.Gba
		k   = &tmp.KeyboardConfig
		c   = &tmp.ControllerConfig
	)

	fields := []Field{
		{WIDGET_HDR, "general", "", nil, nil},
		{WIDGET_CBX, "optimize idle loops", "", &tmp.IdleOptimize, nil},
		{WIDGET_HEX, "snd clk cycles", "snd clk cycles", &tmp.SoundClockUpdateCycles, 1000},

		{WIDGET_HDR, "keyboard", "", nil, nil},
		{WIDGET_KEY, "a", "gba keyboard a", &k.A, nil},
		{WIDGET_KEY, "b", "gba keyboard b", &k.B, nil},
		{WIDGET_KEY, "select", "gba keyboard select", &k.Select, nil},
		{WIDGET_KEY, "start", "gba keyboard start", &k.Start, nil},
		{WIDGET_KEY, "left", "gba keyboard left", &k.Left, nil},
		{WIDGET_KEY, "right", "gba keyboard right", &k.Right, nil},
		{WIDGET_KEY, "up", "gba keyboard up", &k.Up, nil},
		{WIDGET_KEY, "down", "gba keyboard down", &k.Down, nil},
		{WIDGET_KEY, "l", "gba keyboard l", &k.L, nil},
		{WIDGET_KEY, "r", "gba keyboard r", &k.R, nil},

		{WIDGET_HDR, "controller", "", nil, nil},
		{WIDGET_KEY, "a", "gba controller a", &c.A, nil},
		{WIDGET_KEY, "b", "gba controller b", &c.B, nil},
		{WIDGET_KEY, "select", "gba controller select", &c.Select, nil},
		{WIDGET_KEY, "start", "gba controller start", &c.Start, nil},
		{WIDGET_KEY, "left", "gba controller left", &c.Left, nil},
		{WIDGET_KEY, "right", "gba controller right", &c.Right, nil},
		{WIDGET_KEY, "up", "gba controller up", &c.Up, nil},
		{WIDGET_KEY, "down", "gba controller down", &c.Down, nil},
		{WIDGET_KEY, "l", "gba controller l", &c.L, nil},
		{WIDGET_KEY, "r", "gba controller r", &c.R, nil},
	}

	parent.RemoveChildren()
	buildSubMenu(g, parent, fields)

	parent.AddChild(NewSaveButton(func(*widget.ButtonClickedEventArgs) {
		config.Conf.Gba = tmp
		if len(g.gamepadIds) != 0 {
			g.ui.focus.FocusLastSubMenu()
		}
		g.ui.toast.AddMessage("saved")
	}))
}

func NewNdsMenu(g *Game, parent *widget.Container) {

	var (
		tmp = config.Conf.Nds
		k   = &tmp.KeyboardConfig
		c   = &tmp.ControllerConfig
	)

	fields := []Field{
		{WIDGET_HDR, "screen", "", nil, nil},
		{WIDGET_RAD, "layout", "", &tmp.Screen.OLayout, []string{"vertical", "horizontal", "hybrid"}},
		{WIDGET_RAD, "sizing", "", &tmp.Screen.OSizing, []string{"even", "only top", "only bottom"}},
		{WIDGET_RAD, "rotation", "", &tmp.Screen.ORotation, []string{"0", "90", "180", "270"}},

		{WIDGET_HDR, "rtc", "", nil, nil},
		{WIDGET_DEC, "additional hours", "additional hours", &tmp.Rtc.AdditionalHours, 24},

		{WIDGET_HDR, "bios", "", nil, nil},
		{WIDGET_FLE, "arm7 path", "", &tmp.Bios.Arm7Path, nil},
		{WIDGET_FLE, "arm9 path", "", &tmp.Bios.Arm9Path, nil},

		{WIDGET_HDR, "firmware", "", nil, nil},
		{WIDGET_FLE, "file path", "", &tmp.Firmware.FilePath, nil},
		{WIDGET_TXT, "nickname", "nickname", &tmp.Firmware.Nickname, nil},
		{WIDGET_TXT, "message", "message", &tmp.Firmware.Message, nil},
		{WIDGET_TXT, "favorite color", "favorite color", &tmp.Firmware.FavoriteColor, nil},

		{WIDGET_HDR, "scene export", "", nil, nil},
		{WIDGET_DIR, "output directory", "", &tmp.Export.Directory, "./export"},
		{WIDGET_CBX, "shadow polygons", "", &tmp.Export.ShadowPolys, nil},

		{WIDGET_HDR, "keyboard", "", nil, nil},
		{WIDGET_KEY, "a", "nds keyboard a", &k.A, nil},
		{WIDGET_KEY, "b", "nds keyboard b", &k.B, nil},
		{WIDGET_KEY, "select", "nds keyboard select", &k.Select, nil},
		{WIDGET_KEY, "start", "nds keyboard start", &k.Start, nil},
		{WIDGET_KEY, "left", "nds keyboard left", &k.Left, nil},
		{WIDGET_KEY, "right", "nds keyboard right", &k.Right, nil},
		{WIDGET_KEY, "up", "nds keyboard up", &k.Up, nil},
		{WIDGET_KEY, "down", "nds keyboard down", &k.Down, nil},
		{WIDGET_KEY, "l", "nds keyboard l", &k.L, nil},
		{WIDGET_KEY, "r", "nds keyboard r", &k.R, nil},
		{WIDGET_KEY, "x", "nds keyboard x", &k.X, nil},
		{WIDGET_KEY, "y", "nds keyboard y", &k.Y, nil},
		{WIDGET_KEY, "hinge", "nds keyboard hinge", &k.Hinge, nil},
		{WIDGET_KEY, "debug", "nds keyboard debug", &k.Debug, nil},
		{WIDGET_KEY, "layout toggle", "nds keyboard layout toggle", &k.LayoutToggle, nil},
		{WIDGET_KEY, "sizing toggle", "nds keyboard sizing toggle", &k.SizingToggle, nil},
		{WIDGET_KEY, "rotation toggle", "nds keyboard rotation toggle", &k.RotationToggle, nil},
		{WIDGET_KEY, "export toggle", "nds keyboard export toggle", &k.ExportScene, nil},

		{WIDGET_HDR, "controller", "", nil, nil},
		{WIDGET_KEY, "a", "nds controller a", &c.A, nil},
		{WIDGET_KEY, "b", "nds controller b", &c.B, nil},
		{WIDGET_KEY, "select", "nds controller select", &c.Select, nil},
		{WIDGET_KEY, "start", "nds controller start", &c.Start, nil},
		{WIDGET_KEY, "left", "nds controller left", &c.Left, nil},
		{WIDGET_KEY, "right", "nds controller right", &c.Right, nil},
		{WIDGET_KEY, "up", "nds controller up", &c.Up, nil},
		{WIDGET_KEY, "down", "nds controller down", &c.Down, nil},
		{WIDGET_KEY, "l", "nds controller l", &c.L, nil},
		{WIDGET_KEY, "r", "nds controller r", &c.R, nil},
		{WIDGET_KEY, "x", "nds controller x", &c.X, nil},
		{WIDGET_KEY, "y", "nds controller y", &c.Y, nil},
	}

	parent.RemoveChildren()
	buildSubMenu(g, parent, fields)

	parent.AddChild(NewSaveButton(func(*widget.ButtonClickedEventArgs) {
		config.Conf.Nds = tmp
		if len(g.gamepadIds) != 0 {
			g.ui.focus.FocusLastSubMenu()
		}
		g.ui.toast.AddMessage("saved")
	}))
}
