package ui

import (
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/config/file"
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

func NewSidebarFields(res *Resources) []SidebarField {
	l := res.localization.Settings.Sidebar

	return []SidebarField{
		{l.General, func(g *Game) { NewGeneralMenu(g, g.ui.content) }},
		{l.Ui, func(g *Game) { NewUiMenu(g, g.ui.content) }},
		{l.Gb, func(g *Game) { NewGbMenu(g, g.ui.content) }},
		{l.Gba, func(g *Game) { NewGbaMenu(g, g.ui.content) }},
		{l.Nds, func(g *Game) { NewNdsMenu(g, g.ui.content) }},
		{
			l.Return, func(g *Game) {
				switch g.ui.PrevPageId {
				case PAGE_HOME:
					NewHome(g)
				case PAGE_PAUSE:
					NewPause(g)
				}
			},
		},
	}
}

func buildSubMenu(g *Game, parent *widget.Container, fields []Field) {
	for _, field := range fields {
		switch field.widgettype {
		case WIDGET_HDR:
			parent.AddChild(NewHeader(field.label, g.ui.res),
				NewSeparator())
		case WIDGET_CBX:
			parent.AddChild(NewLabel(field.label),
				NewCheckbox(field.ptr.(*bool)))
		case WIDGET_KEY:
			parent.AddChild(NewLabel(field.label),
				NewKeybindInput(g.ui, field.sublabel, field.ptr, field.other.(func(string) (bool, *string))))
		case WIDGET_DEC:
			parent.AddChild(NewLabel(field.label),
				NewDecimalInput(g.ui, field.sublabel, field.ptr, field.other.(int)))
		case WIDGET_HEX:
			parent.AddChild(NewLabel(field.label),
				NewHexInput(g.ui, field.sublabel, field.ptr, field.other.(int)))
		case WIDGET_FLE:
			parent.AddChild(NewLabel(field.label),
				NewFileInput(field.ptr.(*string)))
		case WIDGET_DIR:
			parent.AddChild(NewLabel(field.label),
				NewDirectoryInput(field.ptr.(*string), field.other.(string)))
		case WIDGET_TXT:
			parent.AddChild(NewLabel(field.label),
				NewTextBoxInput(g.ui, BOARD_ALPHA, field.sublabel, field.ptr, field.other.(func(s string) (bool, *string))))
		case WIDGET_LNK:
			parent.AddChild(NewSeparator(),
				NewLinkText(field.other.(string)))
		case WIDGET_RAD:
			parent.AddChild(
				NewLabel(field.label),
				NewRadioInput(
					&g.ui.focus.horizontalGroup,
					field.ptr.(*int),
					field.other.([]string),
					g.ui.res,
				),
			)
		}
	}
}

func NewGeneralMenu(g *Game, parent *widget.Container) {
	var (
		tmp = config.Conf.General
		k   = &tmp.Keyboard
		c   = &tmp.Controller

		l = g.ui.res.localization.Settings.General
	)

	fields := []Field{
		{WIDGET_HDR, l.General, "", nil, nil},
		{WIDGET_CBX, l.Muted, "", &tmp.Muted, nil},
		{WIDGET_CBX, l.ShowFps, "", &tmp.ShowFps, nil},
		{WIDGET_CBX, l.InitFullscreen, "", &tmp.InitFullscreen, nil},
		{WIDGET_DEC, l.TargetFps, l.TargetFps, &tmp.TargetFps, 1_000_000},
		{WIDGET_CBX, l.VsyncEnabled, "", &tmp.Vsync, nil},
		{WIDGET_CBX, l.DisableSaves, "", &tmp.DisableSaves, nil},
		{WIDGET_CBX, l.IntegerScaling, "", &tmp.IntegerScaling, nil},
		{WIDGET_LNK, "", "", nil, "a ratio of zero is dynamic"},
		{WIDGET_DEC, l.IntegerScalingRatio, "", &tmp.IntegerScalingRatio, 10},

		{WIDGET_HDR, l.Keyboard, "", nil, nil},
		{WIDGET_LNK, "", "", nil, keybindsLink},
		{WIDGET_KEY, l.Select, l.KeyboardSelect, &k.Select, KeyValidation()},
		{WIDGET_KEY, l.Return, l.KeyboardReturn, &k.Return, KeyValidation()},
		{WIDGET_KEY, l.Mute, l.KeyboardMute, &k.Mute, KeyValidation()},
		{WIDGET_KEY, l.Pause, l.KeyboardPause, &k.Pause, KeyValidation()},
		{WIDGET_KEY, l.Left, l.KeyboardLeft, &k.Left, KeyValidation()},
		{WIDGET_KEY, l.Right, l.KeyboardRight, &k.Right, KeyValidation()},
		{WIDGET_KEY, l.Up, l.KeyboardUp, &k.Up, KeyValidation()},
		{WIDGET_KEY, l.Down, l.KeyboardDown, &k.Down, KeyValidation()},
		{WIDGET_KEY, l.Fullscreen, l.KeyboardFullscreen, &k.Fullscreen, KeyValidation()},
		{WIDGET_KEY, l.Quit, l.KeyboardQuit, &k.Quit, KeyValidation()},

		{WIDGET_HDR, l.Controller, "", nil, nil},
		{WIDGET_LNK, "", "", nil, controllerLink},
		{WIDGET_KEY, l.Select, l.ControllerSelect, &c.Select, ControllerValidation()},
		{WIDGET_KEY, l.Return, l.ControllerReturn, &c.Return, ControllerValidation()},
		{WIDGET_KEY, l.Mute, l.ControllerMute, &c.Mute, ControllerValidation()},
		{WIDGET_KEY, l.Pause, l.ControllerPause, &c.Pause, ControllerValidation()},
		{WIDGET_KEY, l.Left, l.ControllerLeft, &c.Left, ControllerValidation()},
		{WIDGET_KEY, l.Right, l.ControllerRight, &c.Right, ControllerValidation()},
		{WIDGET_KEY, l.Up, l.ControllerUp, &c.Up, ControllerValidation()},
		{WIDGET_KEY, l.Down, l.ControllerDown, &c.Down, ControllerValidation()},
		{WIDGET_KEY, l.Fullscreen, l.ControllerFullscreen, &c.Fullscreen, ControllerValidation()},
		{WIDGET_KEY, l.Quit, l.ControllerQuit, &c.Quit, ControllerValidation()},
	}

	parent.RemoveChildren()
	buildSubMenu(g, parent, fields)

	parent.AddChild(NewSaveButton(l.Save, func(*widget.ButtonClickedEventArgs) {
		config.Conf.General = tmp

		parent.RemoveChildren()
		NewGeneralMenu(g, parent)

		file.Encode()

		if len(g.gamepadIds) != 0 {
			g.ui.focus.FocusLastSubMenu()
		}

		g.ui.toast.AddMessage(g.ui.res.localization.Toast.Saved)
	}))

	g.ui.focus.submenu = parent.GetFocusers()
	g.ui.focus.BuildFocus(g.ui.ui)
}

func NewUiMenu(g *Game, parent *widget.Container) {
	var (
		res   = g.ui.res
		tmp   = config.Conf.Ui
		oldId = g.ui.PrevPageId

		l = g.ui.res.localization.Settings.Ui

		clrInputs = [4]widget.PreferredSizeLocateableWidget{
			NewColorInput(g.ui, l.UiBackdrop, &tmp.Backdrop, HexValidation(0xFFFFFF)),
			NewColorInput(g.ui, l.UiBgColor, &tmp.MenuBackgroundColor, HexValidation(0xFFFFFF)),
			NewColorInput(g.ui, l.UiFgColor, &tmp.MenuForegroundColor, HexValidation(0xFFFFFF)),
			NewColorInput(g.ui, l.UiAccentColor, &tmp.MenuSecondaryColor, HexValidation(0xFFFFFF)),
		}
	)
	parent.RemoveChildren()

	parent.AddChild(
		NewHeader(l.Ui, res), NewSeparator(),
		NewLabel(l.Language), NewRadioInput(&g.ui.focus.horizontalGroup, &tmp.Language, l.Languages, res),
		NewLabel(l.Backdrop), clrInputs[0],
		NewLabel(l.BgColor), clrInputs[1],
		NewLabel(l.FgColor), clrInputs[2],
		NewLabel(l.AccentColor), clrInputs[3],
		NewLabel(l.ApplyTheme),
		NewApplyPalettesMenu(&g.ui.focus.horizontalGroup, theme_palettes, clrInputs, res),
		NewSaveButton(l.Save, func(*widget.ButtonClickedEventArgs) {
			config.Conf.Ui = tmp
			g.ui.res.localization = NewLocalization(LangOptions(config.Conf.Ui.Language))

			res.Update()
			g.ui.keyboard = NewKeyboard(g.ui.res, g.ui.res.localization.Settings.Ui.Alphabet)

			NewSettings(g, oldId, MENU_UI)
			// should this be somewhere else?

			file.Encode()

			if len(g.gamepadIds) != 0 {
				g.ui.focus.FocusLastSubMenu()
			}

			g.ui.toast.AddMessage(g.ui.res.localization.Toast.Saved)
		}),
	)
	g.ui.focus.submenu = parent.GetFocusers()
	g.ui.focus.BuildFocus(g.ui.ui)
}

func NewGbMenu(g *Game, parent *widget.Container) {
	var (
		tmp = config.Conf.Gb
		k   = &tmp.KeyboardConfig
		c   = &tmp.ControllerConfig
		pal = &tmp.Palette

		l = g.ui.res.localization.Settings.Gb

		clrInputs = [4]widget.PreferredSizeLocateableWidget{
			NewColorInput(g.ui, l.DmgLightest, &pal[0], HexValidation(0xFFFFFF)),
			NewColorInput(g.ui, l.DmgLight, &pal[1], HexValidation(0xFFFFFF)),
			NewColorInput(g.ui, l.DmgDark, &pal[2], HexValidation(0xFFFFFF)),
			NewColorInput(g.ui, l.DmgDarkest, &pal[3], HexValidation(0xFFFFFF)),
		}
	)

	fields := []Field{
		{WIDGET_HDR, l.Keyboard, "", nil, nil},
		{WIDGET_LNK, "", "", nil, keybindsLink},
		{WIDGET_KEY, l.A, l.KeyboardA, &k.A, KeyValidation()},
		{WIDGET_KEY, l.B, l.KeyboardB, &k.B, KeyValidation()},
		{WIDGET_KEY, l.Select, l.KeyboardSelect, &k.Select, KeyValidation()},
		{WIDGET_KEY, l.Start, l.KeyboardStart, &k.Start, KeyValidation()},
		{WIDGET_KEY, l.Left, l.KeyboardLeft, &k.Left, KeyValidation()},
		{WIDGET_KEY, l.Right, l.KeyboardRight, &k.Right, KeyValidation()},
		{WIDGET_KEY, l.Up, l.KeyboardUp, &k.Up, KeyValidation()},
		{WIDGET_KEY, l.Down, l.KeyboardDown, &k.Down, KeyValidation()},

		{WIDGET_HDR, l.Controller, "", nil, nil},
		{WIDGET_LNK, "", "", nil, controllerLink},
		{WIDGET_KEY, l.A, l.ControllerA, &c.A, ControllerValidation()},
		{WIDGET_KEY, l.B, l.ControllerB, &c.B, ControllerValidation()},
		{WIDGET_KEY, l.Select, l.ControllerSelect, &c.Select, ControllerValidation()},
		{WIDGET_KEY, l.Start, l.ControllerStart, &c.Start, ControllerValidation()},
		{WIDGET_KEY, l.Left, l.ControllerLeft, &c.Left, ControllerValidation()},
		{WIDGET_KEY, l.Right, l.ControllerRight, &c.Right, ControllerValidation()},
		{WIDGET_KEY, l.Up, l.ControllerUp, &c.Up, ControllerValidation()},
		{WIDGET_KEY, l.Down, l.ControllerDown, &c.Down, ControllerValidation()},
	}

	parent.RemoveChildren()

	parent.AddChild(
		NewHeader(l.DmgPalette, g.ui.res), NewSeparator(),
		NewLabel(l.Lightest), clrInputs[0],
		NewLabel(l.Light), clrInputs[1],
		NewLabel(l.Dark), clrInputs[2],
		NewLabel(l.Darkest), clrInputs[3],
		NewLabel(l.ApplyPalette),
		NewApplyPalettesMenu(&g.ui.focus.horizontalGroup, dmg_palettes, clrInputs, g.ui.res),
	)

	buildSubMenu(g, parent, fields)

	parent.AddChild(NewSaveButton(l.Save, func(*widget.ButtonClickedEventArgs) {
		config.Conf.Gb = tmp

		parent.RemoveChildren()
		NewGbMenu(g, parent)
		file.Encode()

		if len(g.gamepadIds) != 0 {
			g.ui.focus.FocusLastSubMenu()
		}
		g.ui.toast.AddMessage(g.ui.res.localization.Toast.Saved)
	}))
	g.ui.focus.submenu = parent.GetFocusers()
	g.ui.focus.BuildFocus(g.ui.ui)
}

func NewGbaMenu(g *Game, parent *widget.Container) {
	var (
		tmp = config.Conf.Gba
		k   = &tmp.KeyboardConfig
		c   = &tmp.ControllerConfig

		l = g.ui.res.localization.Settings.Gba
	)

	fields := []Field{
		{WIDGET_HDR, l.General, "", nil, nil},
		{WIDGET_CBX, l.OptmizeIdleLoops, "", &tmp.IdleOptimize, nil},
		{WIDGET_HEX, l.SoundClockCycles, l.SoundClockCycles, &tmp.SoundClockUpdateCycles, 1000},

		{WIDGET_HDR, l.Keyboard, "", nil, nil},
		{WIDGET_LNK, "", "", nil, keybindsLink},
		{WIDGET_KEY, l.A, l.KeyboardA, &k.A, KeyValidation()},
		{WIDGET_KEY, l.B, l.KeyboardB, &k.B, KeyValidation()},
		{WIDGET_KEY, l.Select, l.KeyboardSelect, &k.Select, KeyValidation()},
		{WIDGET_KEY, l.Start, l.KeyboardStart, &k.Start, KeyValidation()},
		{WIDGET_KEY, l.Left, l.KeyboardLeft, &k.Left, KeyValidation()},
		{WIDGET_KEY, l.Right, l.KeyboardRight, &k.Right, KeyValidation()},
		{WIDGET_KEY, l.Up, l.KeyboardUp, &k.Up, KeyValidation()},
		{WIDGET_KEY, l.Down, l.KeyboardDown, &k.Down, KeyValidation()},
		{WIDGET_KEY, l.L, l.KeyboardL, &k.L, KeyValidation()},
		{WIDGET_KEY, l.R, l.KeyboardR, &k.R, KeyValidation()},

		{WIDGET_HDR, l.Controller, "", nil, nil},
		{WIDGET_LNK, "", "", nil, controllerLink},
		{WIDGET_KEY, l.A, l.ControllerA, &c.A, ControllerValidation()},
		{WIDGET_KEY, l.B, l.ControllerB, &c.B, ControllerValidation()},
		{WIDGET_KEY, l.Select, l.ControllerSelect, &c.Select, ControllerValidation()},
		{WIDGET_KEY, l.Start, l.ControllerStart, &c.Start, ControllerValidation()},
		{WIDGET_KEY, l.Left, l.ControllerLeft, &c.Left, ControllerValidation()},
		{WIDGET_KEY, l.Right, l.ControllerRight, &c.Right, ControllerValidation()},
		{WIDGET_KEY, l.Up, l.ControllerUp, &c.Up, ControllerValidation()},
		{WIDGET_KEY, l.Down, l.ControllerDown, &c.Down, ControllerValidation()},
		{WIDGET_KEY, l.L, l.ControllerL, &k.L, ControllerValidation()},
		{WIDGET_KEY, l.R, l.ControllerR, &k.R, ControllerValidation()},
	}

	parent.RemoveChildren()
	buildSubMenu(g, parent, fields)

	parent.AddChild(NewSaveButton(l.Save, func(*widget.ButtonClickedEventArgs) {
		config.Conf.Gba = tmp

		parent.RemoveChildren()
		NewGbaMenu(g, parent)
		file.Encode()

		if len(g.gamepadIds) != 0 {
			g.ui.focus.FocusLastSubMenu()
		}
		g.ui.toast.AddMessage(g.ui.res.localization.Toast.Saved)
	}))
	g.ui.focus.submenu = parent.GetFocusers()
	g.ui.focus.BuildFocus(g.ui.ui)
}

func NewNdsMenu(g *Game, parent *widget.Container) {
	var (
		tmp = config.Conf.Nds
		k   = &tmp.KeyboardConfig
		c   = &tmp.ControllerConfig

		l = g.ui.res.localization.Settings.Nds

		favColor = config.ColorNames[tmp.Firmware.Color]
	)

	fields := []Field{
		{WIDGET_HDR, l.Screen, "", nil, nil},
		{WIDGET_RAD, l.Layout, "", &tmp.Screen.Layout, l.Layouts},
		{WIDGET_RAD, l.Sizing, "", &tmp.Screen.Sizing, l.Sizings},
		{WIDGET_RAD, l.Rotation, "", &tmp.Screen.Rotation, l.Rotations},

		{WIDGET_HDR, l.Rtc, "", nil, nil},
		{WIDGET_DEC, l.AdditionalHours, l.AdditionalHours, &tmp.Rtc.AdditionalHours, 24},

		{WIDGET_HDR, l.Bios, "", nil, nil},
		{WIDGET_FLE, l.Arm7Path, "", &tmp.Bios.Arm7Path, nil},
		{WIDGET_FLE, l.Arm9Path, "", &tmp.Bios.Arm9Path, nil},

		{WIDGET_HDR, l.Firmware, "", nil, nil},
		{WIDGET_FLE, l.FilePath, "", &tmp.Firmware.FilePath, nil},
		{WIDGET_TXT, l.Nickname, l.Nickname, &tmp.Firmware.Nickname, StringValidation(10)},
		{WIDGET_TXT, l.Message, l.Message, &tmp.Firmware.Message, StringValidation(26)},
		{WIDGET_TXT, l.FavoriteColor, l.FavoriteColor, &favColor, ColorNdsValidation()},

		{WIDGET_HDR, l.SceneExport, "", nil, nil},
		{WIDGET_DIR, l.OutputDirectory, "", &tmp.Export.Directory, "./export"},
		{WIDGET_CBX, l.ShadowPolygons, "", &tmp.Export.ShadowPolys, nil},

		{WIDGET_HDR, l.Keyboard, "", nil, nil},
		{WIDGET_LNK, "", "", nil, keybindsLink},
		{WIDGET_KEY, l.A, l.KeyboardA, &k.A, KeyValidation()},
		{WIDGET_KEY, l.B, l.KeyboardB, &k.B, KeyValidation()},
		{WIDGET_KEY, l.Select, l.KeyboardSelect, &k.Select, KeyValidation()},
		{WIDGET_KEY, l.Start, l.KeyboardStart, &k.Start, KeyValidation()},
		{WIDGET_KEY, l.Left, l.KeyboardLeft, &k.Left, KeyValidation()},
		{WIDGET_KEY, l.Right, l.KeyboardRight, &k.Right, KeyValidation()},
		{WIDGET_KEY, l.Up, l.KeyboardUp, &k.Up, KeyValidation()},
		{WIDGET_KEY, l.Down, l.KeyboardDown, &k.Down, KeyValidation()},
		{WIDGET_KEY, l.L, l.KeyboardL, &k.L, KeyValidation()},
		{WIDGET_KEY, l.R, l.KeyboardR, &k.R, KeyValidation()},
		{WIDGET_KEY, l.X, l.KeyboardX, &k.X, KeyValidation()},
		{WIDGET_KEY, l.Y, l.KeyboardY, &k.Y, KeyValidation()},
		{WIDGET_KEY, l.Hinge, l.KeyboardHinge, &k.Hinge, KeyValidation()},
		{WIDGET_KEY, l.Debug, l.KeyboardDebug, &k.Debug, KeyValidation()},
		{WIDGET_KEY, l.LayoutToggle, l.KeyboardLayoutToggle, &k.LayoutToggle, KeyValidation()},
		{WIDGET_KEY, l.SizingToggle, l.KeyboardSizingToggle, &k.SizingToggle, KeyValidation()},
		{WIDGET_KEY, l.RotationToggle, l.KeyboardRotationToggle, &k.RotationToggle, KeyValidation()},
		{WIDGET_KEY, l.ExportToggle, l.KeyboardExportToggle, &k.ExportScene, KeyValidation()},

		{WIDGET_HDR, l.Controller, "", nil, nil},
		{WIDGET_LNK, "", "", nil, controllerLink},
		{WIDGET_KEY, l.A, l.ControllerA, &c.A, ControllerValidation()},
		{WIDGET_KEY, l.B, l.ControllerB, &c.B, ControllerValidation()},
		{WIDGET_KEY, l.Select, l.ControllerSelect, &c.Select, ControllerValidation()},
		{WIDGET_KEY, l.Start, l.ControllerStart, &c.Start, ControllerValidation()},
		{WIDGET_KEY, l.Left, l.ControllerLeft, &c.Left, ControllerValidation()},
		{WIDGET_KEY, l.Right, l.ControllerRight, &c.Right, ControllerValidation()},
		{WIDGET_KEY, l.Up, l.ControllerUp, &c.Up, ControllerValidation()},
		{WIDGET_KEY, l.Down, l.ControllerDown, &c.Down, ControllerValidation()},
		{WIDGET_KEY, l.L, l.ControllerL, &k.L, ControllerValidation()},
		{WIDGET_KEY, l.R, l.ControllerR, &k.R, ControllerValidation()},
		{WIDGET_KEY, l.X, l.ControllerX, &k.X, ControllerValidation()},
		{WIDGET_KEY, l.Y, l.ControllerY, &k.Y, ControllerValidation()},
	}

	parent.RemoveChildren()
	buildSubMenu(g, parent, fields)

	parent.AddChild(NewSaveButton(l.Save, func(*widget.ButtonClickedEventArgs) {
		config.Conf.Nds = tmp
		config.Conf.Nds.Firmware.Color = config.ColorNameToId[favColor]

		parent.RemoveChildren()
		NewNdsMenu(g, parent)

		file.Encode()

		if len(g.gamepadIds) != 0 {
			g.ui.focus.FocusLastSubMenu()
		}
		g.ui.toast.AddMessage(g.ui.res.localization.Toast.Saved)
	}))

	g.ui.focus.submenu = parent.GetFocusers()
	g.ui.focus.BuildFocus(g.ui.ui)
}
