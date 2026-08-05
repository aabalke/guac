package ui

import (
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/config/file"
	"github.com/aabalke/guac/emu/gba/gpio"
	"github.com/ebitenui/ebitenui/widget"
)

const (
	MENU_GENERAL = iota
	MENU_UI
	MENU_GB
	MENU_GBA
	MENU_NDS
	MENU_ABOUT
	MENU_RETURN
)

type SidebarField struct {
	label string
	f     func(g *Game)
}

func NewSidebarFields(res *Resources) []SidebarField {
	l := res.localization.Settings.Sidebar

	return []SidebarField{
		{l.General, func(g *Game) {
			g.ui.content.RemoveChildren()
			g.ui.content.AddChild(NewGeneralMenu(g))
		}},
		{l.Ui, func(g *Game) {
			g.ui.content.RemoveChildren()
			g.ui.content.AddChild(NewUiMenu(g))
		}},
		{l.Gb, func(g *Game) {
			g.ui.content.RemoveChildren()
			g.ui.content.AddChild(NewGbMenu(g))
		}},
		{l.Gba, func(g *Game) {
			g.ui.content.RemoveChildren()
			g.ui.content.AddChild(NewGbaMenu(g))
		}},
		{l.Nds, func(g *Game) {
			g.ui.content.RemoveChildren()
			g.ui.content.AddChild(NewNdsMenu(g))
		}},
		{l.About, func(g *Game) {
			g.ui.content.RemoveChildren()
			g.ui.content.AddChild(NewAboutMenu(g))
		}},
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

func NewMenu() *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch(
				[]bool{true},
				[]bool{true},
			),
		)),
	)
}

func NewOneCol() *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(24)),
			widget.GridLayoutOpts.Spacing(32, 16),
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch(
				[]bool{true},
				[]bool{},
			),
		)),
	)
}

func NewTwoCol() *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(24)),
			widget.GridLayoutOpts.Spacing(32, 16),
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch(
				[]bool{false, true},
				[]bool{},
			),
		)),
	)
}

func NewGeneralMenu(g *Game) *widget.Container {
	tmp := config.Conf.General
	l := g.ui.res.localization.Settings.General

	menu := NewMenu()

	general := NewTwoCol()
	menu.AddChild(NewHeader(l.General, g.ui.res), general)

	general.AddChild(
		NewLabel(l.Muted), NewCheckbox(&tmp.Muted),
		NewLabel(l.ShowFps), NewCheckbox(&tmp.ShowFps),
		NewLabel(l.InitFullscreen), NewCheckbox(&tmp.InitFullscreen),
		NewLabel(l.TargetFps), NewDecimalInput(g.ui, l.TargetFps, &tmp.TargetFps, 1_000_000),
		NewLabel(l.VsyncEnabled), NewCheckbox(&tmp.Vsync),
		NewLabel(l.DisableSaves), NewCheckbox(&tmp.DisableSaves),
		NewLabel(l.IntegerScaling), NewCheckbox(&tmp.IntegerScaling),
		NewSeparator(), NewLinkText(l.IntegerScalingDesc),
		NewLabel(l.IntegerScalingRatio), NewDecimalInput(g.ui, l.IntegerScalingRatio, &tmp.IntegerScalingRatio, 10),
		NewSeparator(), NewLinkText(l.SampleRateDesc),
		NewLabel(l.SampleRate), NewDecimalInput(g.ui, l.SampleRate, &tmp.SampleRate, 192000),
	)

	keys := NewTwoCol()
	menu.AddChild(NewHeader(l.Keyboard, g.ui.res), keys)
	k := &tmp.Keyboard
	keys.AddChild(
		NewLabel(l.Select), NewKeybindInput(g.ui, &k.Select),
		NewLabel(l.Return), NewKeybindInput(g.ui, &k.Return),
		NewLabel(l.Mute), NewKeybindInput(g.ui, &k.Mute),
		NewLabel(l.Pause), NewKeybindInput(g.ui, &k.Pause),
		NewLabel(l.Left), NewKeybindInput(g.ui, &k.Left),
		NewLabel(l.Right), NewKeybindInput(g.ui, &k.Right),
		NewLabel(l.Up), NewKeybindInput(g.ui, &k.Up),
		NewLabel(l.Down), NewKeybindInput(g.ui, &k.Down),
		NewLabel(l.Fullscreen), NewKeybindInput(g.ui, &k.Fullscreen),
		NewLabel(l.Quit), NewKeybindInput(g.ui, &k.Quit),
	)

	gamepad := NewTwoCol()
	menu.AddChild(NewHeader(l.Controller, g.ui.res), gamepad)
	c := &tmp.Controller
	gamepad.AddChild(
		NewLabel(l.Select), NewGamepadInput(g.ui, &c.Select),
		NewLabel(l.Return), NewGamepadInput(g.ui, &c.Return),
		NewLabel(l.Mute), NewGamepadInput(g.ui, &c.Mute),
		NewLabel(l.Pause), NewGamepadInput(g.ui, &c.Pause),
		NewLabel(l.Left), NewGamepadInput(g.ui, &c.Left),
		NewLabel(l.Right), NewGamepadInput(g.ui, &c.Right),
		NewLabel(l.Up), NewGamepadInput(g.ui, &c.Up),
		NewLabel(l.Down), NewGamepadInput(g.ui, &c.Down),
		NewLabel(l.Fullscreen), NewGamepadInput(g.ui, &c.Fullscreen),
		NewLabel(l.Quit), NewGamepadInput(g.ui, &c.Quit),
	)

	menu.AddChild(NewSaveButton(l.Save, func(*widget.ButtonClickedEventArgs) {
		config.Conf.General = tmp

		g.ui.content.RemoveChildren()
		g.ui.content.AddChild(NewGeneralMenu(g))

		file.Encode()

		if len(g.ui.gamepadIds) != 0 {
			g.ui.focus.FocusLastSubMenu()
		}

		g.ui.toast.AddMessage(g.ui.res.localization.Toast.Saved)
	}))

	g.ui.focus.BuildMenuFocus(g.ui.ui, menu, keys, gamepad)

	return menu
}

func NewUiMenu(g *Game) *widget.Container {
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

	menu := NewMenu()

	ui := NewTwoCol()
	menu.AddChild(NewHeader(l.Ui, res), ui)

	ui.AddChild(
		NewLabel(l.Language), NewRadioInput(&g.ui.focus.horizontalGroup, &tmp.Language, l.Languages, res),
		NewLabel(l.Backdrop), clrInputs[0],
		NewLabel(l.BgColor), clrInputs[1],
		NewLabel(l.FgColor), clrInputs[2],
		NewLabel(l.AccentColor), clrInputs[3],
		NewLabel(l.ApplyTheme),
		NewApplyPalettesMenu(&g.ui.focus.horizontalGroup, theme_palettes, clrInputs, res),
	)

	menu.AddChild(
		NewSaveButton(l.Save, func(*widget.ButtonClickedEventArgs) {
			config.Conf.Ui = tmp
			g.ui.res.localization = NewLocalization(LangOptions(config.Conf.Ui.Language))

			res.Update()
			g.ui.keyboard = NewKeyboard(g.ui.res, g.ui.res.localization.Settings.Ui.Alphabet)

			NewSettings(g, oldId, MENU_UI)
			// should this be somewhere else?

			file.Encode()

			if len(g.ui.gamepadIds) != 0 {
				g.ui.focus.FocusLastSubMenu()
			}

			g.ui.toast.AddMessage(g.ui.res.localization.Toast.Saved)
		}),
	)

	g.ui.focus.BuildMenuFocus(g.ui.ui, menu, nil, nil)

	return menu
}

func NewGbMenu(g *Game) *widget.Container {
	tmp := config.Conf.Gb
	l := g.ui.res.localization.Settings.Gb
	pal := &tmp.Palette
	clrInputs := [4]widget.PreferredSizeLocateableWidget{
		NewColorInput(g.ui, l.DmgLightest, &pal[0], HexValidation(0xFFFFFF)),
		NewColorInput(g.ui, l.DmgLight, &pal[1], HexValidation(0xFFFFFF)),
		NewColorInput(g.ui, l.DmgDark, &pal[2], HexValidation(0xFFFFFF)),
		NewColorInput(g.ui, l.DmgDarkest, &pal[3], HexValidation(0xFFFFFF)),
	}

	menu := NewMenu()

	general := NewTwoCol()
	menu.AddChild(NewHeader(l.General, g.ui.res), general)
	general.AddChild(
		NewLabel(l.System), NewRadioInput(&g.ui.focus.horizontalGroup, &tmp.System, l.Systems, g.ui.res),
	)

	palette := NewTwoCol()
	menu.AddChild(NewHeader(l.DmgPalette, g.ui.res), palette)
	palette.AddChild(
		NewLabel(l.Lightest), clrInputs[0],
		NewLabel(l.Light), clrInputs[1],
		NewLabel(l.Dark), clrInputs[2],
		NewLabel(l.Darkest), clrInputs[3],
		NewLabel(l.ApplyPalette),
		NewApplyPalettesMenu(&g.ui.focus.horizontalGroup, dmg_palettes, clrInputs, g.ui.res),
	)

	bios := NewTwoCol()
	menu.AddChild(NewHeader(l.Bios, g.ui.res), bios)
	bios.AddChild(
		NewLabel(l.DmgPath), NewFileInput(&tmp.Bios.DmgPath),
		NewLabel(l.GbcPath), NewFileInput(&tmp.Bios.GbcPath),
		NewLabel(l.DirectBoot), NewCheckbox(&tmp.Bios.Direct),
	)

	keys := NewTwoCol()
	menu.AddChild(NewHeader(l.Keyboard, g.ui.res), keys)
	k := &tmp.Keyboard
	keys.AddChild(
		NewLabel(l.A), NewKeybindInput(g.ui, &k.A),
		NewLabel(l.B), NewKeybindInput(g.ui, &k.B),
		NewLabel(l.Select), NewKeybindInput(g.ui, &k.Select),
		NewLabel(l.Start), NewKeybindInput(g.ui, &k.Start),
		NewLabel(l.Left), NewKeybindInput(g.ui, &k.Left),
		NewLabel(l.Right), NewKeybindInput(g.ui, &k.Right),
		NewLabel(l.Up), NewKeybindInput(g.ui, &k.Up),
		NewLabel(l.Down), NewKeybindInput(g.ui, &k.Down),
	)

	gamepad := NewTwoCol()
	menu.AddChild(NewHeader(l.Controller, g.ui.res), gamepad)
	c := &tmp.Controller
	gamepad.AddChild(
		NewLabel(l.A), NewGamepadInput(g.ui, &c.A),
		NewLabel(l.B), NewGamepadInput(g.ui, &c.B),
		NewLabel(l.Select), NewGamepadInput(g.ui, &c.Select),
		NewLabel(l.Start), NewGamepadInput(g.ui, &c.Start),
		NewLabel(l.Left), NewGamepadInput(g.ui, &c.Left),
		NewLabel(l.Right), NewGamepadInput(g.ui, &c.Right),
		NewLabel(l.Up), NewGamepadInput(g.ui, &c.Up),
		NewLabel(l.Down), NewGamepadInput(g.ui, &c.Down),
	)

	menu.AddChild(NewSaveButton(l.Save, func(*widget.ButtonClickedEventArgs) {
		config.Conf.Gb = tmp

		g.ui.content.RemoveChildren()
		g.ui.content.AddChild(NewGbMenu(g))

		file.Encode()

		if len(g.ui.gamepadIds) != 0 {
			g.ui.focus.FocusLastSubMenu()
		}
		g.ui.toast.AddMessage(g.ui.res.localization.Toast.Saved)
	}))

	g.ui.focus.BuildMenuFocus(g.ui.ui, menu, keys, gamepad)

	return menu
}

func NewGbaMenu(g *Game) *widget.Container {
	tmp := config.Conf.Gba
	l := g.ui.res.localization.Settings.Gba

	menu := NewMenu()

	general := NewTwoCol()
	menu.AddChild(NewHeader(l.General, g.ui.res), general)

	general.AddChild(
		NewLabel(l.OptmizeIdleLoops), NewCheckbox(&tmp.IdleOptimize),
		NewLabel(l.Rotation), NewRadioInput(&g.ui.focus.horizontalGroup, &tmp.Rotation, l.Rotations, g.ui.res),
	)

	hardware := NewTwoCol()
	menu.AddChild(NewHeader(l.Hardware, g.ui.res), hardware)

	hardware.AddChild(
		NewLabel(l.BackupType), NewRadioInput(&g.ui.focus.horizontalGroup, &tmp.Hardware.BackupType, l.BackupTypes, g.ui.res),
		NewLabel(l.ForceRtc), NewCheckbox(&tmp.Hardware.ForceRtc),
		NewLabel(l.ForceSolarSensor), NewCheckbox(&tmp.Hardware.ForceSolarSensor),
		NewLabel(l.SolarSensorLevel), NewDecimalInput(g.ui, l.SolarSensorLevel, &tmp.Hardware.SolarSensorLevel, 100),
	)

	bios := NewTwoCol()
	menu.AddChild(NewHeader(l.Bios, g.ui.res), bios)

	bios.AddChild(
		NewLabel(l.BiosPath), NewFileInput(&tmp.Bios.Path),
		NewLabel(l.DirectBoot), NewCheckbox(&tmp.Bios.Direct),
	)

	keys := NewTwoCol()
	menu.AddChild(NewHeader(l.Keyboard, g.ui.res), keys)
	k := &tmp.Keyboard
	keys.AddChild(
		NewLabel(l.A), NewKeybindInput(g.ui, &k.A),
		NewLabel(l.B), NewKeybindInput(g.ui, &k.B),
		NewLabel(l.Select), NewKeybindInput(g.ui, &k.Select),
		NewLabel(l.Start), NewKeybindInput(g.ui, &k.Start),
		NewLabel(l.Left), NewKeybindInput(g.ui, &k.Left),
		NewLabel(l.Right), NewKeybindInput(g.ui, &k.Right),
		NewLabel(l.Up), NewKeybindInput(g.ui, &k.Up),
		NewLabel(l.Down), NewKeybindInput(g.ui, &k.Down),
		NewLabel(l.L), NewKeybindInput(g.ui, &k.L),
		NewLabel(l.R), NewKeybindInput(g.ui, &k.R),
		NewLabel(l.SolarMin), NewKeybindInput(g.ui, &k.SolarLevel0),
		NewLabel(l.Solar1), NewKeybindInput(g.ui, &k.SolarLevel1),
		NewLabel(l.Solar2), NewKeybindInput(g.ui, &k.SolarLevel2),
		NewLabel(l.Solar3), NewKeybindInput(g.ui, &k.SolarLevel3),
		NewLabel(l.SolarMax), NewKeybindInput(g.ui, &k.SolarLevel4),
		NewLabel(l.RotationToggle), NewKeybindInput(g.ui, &k.RotationToggle),
	)

	gamepad := NewTwoCol()
	menu.AddChild(NewHeader(l.Controller, g.ui.res), gamepad)
	c := &tmp.Controller
	gamepad.AddChild(
		NewLabel(l.A), NewGamepadInput(g.ui, &c.A),
		NewLabel(l.B), NewGamepadInput(g.ui, &c.B),
		NewLabel(l.Select), NewGamepadInput(g.ui, &c.Select),
		NewLabel(l.Start), NewGamepadInput(g.ui, &c.Start),
		NewLabel(l.Left), NewGamepadInput(g.ui, &c.Left),
		NewLabel(l.Right), NewGamepadInput(g.ui, &c.Right),
		NewLabel(l.Up), NewGamepadInput(g.ui, &c.Up),
		NewLabel(l.Down), NewGamepadInput(g.ui, &c.Down),
		NewLabel(l.L), NewGamepadInput(g.ui, &c.L),
		NewLabel(l.R), NewGamepadInput(g.ui, &c.R),
		NewLabel(l.SolarMin), NewGamepadInput(g.ui, &c.SolarLevel0),
		NewLabel(l.Solar1), NewGamepadInput(g.ui, &c.SolarLevel1),
		NewLabel(l.Solar2), NewGamepadInput(g.ui, &c.SolarLevel2),
		NewLabel(l.Solar3), NewGamepadInput(g.ui, &c.SolarLevel3),
		NewLabel(l.SolarMax), NewGamepadInput(g.ui, &c.SolarLevel4),
		NewLabel(l.RotationToggle), NewGamepadInput(g.ui, &c.RotationToggle),
	)

	menu.AddChild(NewSaveButton(l.Save, func(*widget.ButtonClickedEventArgs) {
		if tmp.Bios.Path == "" {
			tmp.Bios.Direct = true
		}

		config.Conf.Gba = tmp

		if g.gba != nil {
			g.gba.IdleOptimize = tmp.IdleOptimize

			if g.gba.Mem.Gpio != nil {
				for _, device := range g.gba.Mem.Gpio.Devices {
					if d, ok := device.(*gpio.Solar); ok {
						d.SetLevel(uint8(tmp.Hardware.SolarSensorLevel))
					}
				}
			}
		}

		g.ui.content.RemoveChildren()
		g.ui.content.AddChild(NewGbaMenu(g))

		file.Encode()

		if len(g.ui.gamepadIds) != 0 {
			g.ui.focus.FocusLastSubMenu()
		}
		g.ui.toast.AddMessage(g.ui.res.localization.Toast.Saved)
	}))

	g.ui.focus.BuildMenuFocus(g.ui.ui, menu, keys, gamepad)

	return menu
}

func NewNdsMenu(g *Game) *widget.Container {
	tmp := config.Conf.Nds
	l := g.ui.res.localization.Settings.Nds

	menu := NewMenu()

	screen := NewTwoCol()
	menu.AddChild(NewHeader(l.Screen, g.ui.res), screen)

	screen.AddChild(
		NewLabel(l.Layout), NewRadioInput(&g.ui.focus.horizontalGroup, &tmp.Screen.Layout, l.Layouts, g.ui.res),
		NewLabel(l.Sizing), NewRadioInput(&g.ui.focus.horizontalGroup, &tmp.Screen.Sizing, l.Sizings, g.ui.res),
		NewLabel(l.Rotation), NewRadioInput(&g.ui.focus.horizontalGroup, &tmp.Screen.Rotation, l.Rotations, g.ui.res),
	)

	rtc := NewTwoCol()
	menu.AddChild(NewHeader(l.Rtc, g.ui.res), rtc)

	rtc.AddChild(
		NewLabel(l.AdditionalHours), NewDecimalInput(g.ui, l.AdditionalHours, &tmp.Rtc.AdditionalHours, 24),
	)

	bios := NewTwoCol()
	menu.AddChild(NewHeader(l.Bios, g.ui.res), bios)

	bios.AddChild(
		NewLabel(l.Arm7Path), NewFileInput(&tmp.Bios.Arm7Path),
		NewLabel(l.Arm9Path), NewFileInput(&tmp.Bios.Arm9Path),
	)

	firmware := NewTwoCol()
	menu.AddChild(NewHeader(l.Firmware, g.ui.res), firmware)
	favColor := config.ColorNames[tmp.Firmware.Color]

	firmware.AddChild(
		NewLabel(l.FilePath), NewFileInput(&tmp.Firmware.FilePath),
		NewLabel(l.Nickname), NewTextBoxInput(g.ui, BOARD_ALPHA, l.Nickname, &tmp.Firmware.Nickname, StringValidation(10)),
		NewLabel(l.Message), NewTextBoxInput(g.ui, BOARD_ALPHA, l.Message, &tmp.Firmware.Message, StringValidation(26)),
		NewLabel(l.FavoriteColor), NewTextBoxInput(g.ui, BOARD_ALPHA, l.FavoriteColor, &favColor, ColorNdsValidation()),
	)

	export := NewTwoCol()
	menu.AddChild(NewHeader(l.SceneExport, g.ui.res), export)

	export.AddChild(
		NewLabel(l.OutputDirectory), NewDirectoryInput(&tmp.Export.Directory, "./export"),
		NewLabel(l.ShadowPolygons), NewCheckbox(&tmp.Export.ShadowPolys),
	)

	keys := NewTwoCol()
	menu.AddChild(NewHeader(l.Keyboard, g.ui.res), keys)
	k := &tmp.Keyboard
	keys.AddChild(
		NewLabel(l.A), NewKeybindInput(g.ui, &k.A),
		NewLabel(l.B), NewKeybindInput(g.ui, &k.B),
		NewLabel(l.Select), NewKeybindInput(g.ui, &k.Select),
		NewLabel(l.Start), NewKeybindInput(g.ui, &k.Start),
		NewLabel(l.Left), NewKeybindInput(g.ui, &k.Left),
		NewLabel(l.Right), NewKeybindInput(g.ui, &k.Right),
		NewLabel(l.Up), NewKeybindInput(g.ui, &k.Up),
		NewLabel(l.Down), NewKeybindInput(g.ui, &k.Down),
		NewLabel(l.L), NewKeybindInput(g.ui, &k.L),
		NewLabel(l.R), NewKeybindInput(g.ui, &k.R),
		NewLabel(l.X), NewKeybindInput(g.ui, &k.X),
		NewLabel(l.Y), NewKeybindInput(g.ui, &k.Y),
		NewLabel(l.Hinge), NewKeybindInput(g.ui, &k.Hinge),
		NewLabel(l.LayoutToggle), NewKeybindInput(g.ui, &k.LayoutToggle),
		NewLabel(l.SizingToggle), NewKeybindInput(g.ui, &k.SizingToggle),
		NewLabel(l.RotationToggle), NewKeybindInput(g.ui, &k.RotationToggle),
		NewLabel(l.ExportToggle), NewKeybindInput(g.ui, &k.ExportScene),
	)

	gamepad := NewTwoCol()
	menu.AddChild(NewHeader(l.Controller, g.ui.res), gamepad)
	c := &tmp.Controller
	gamepad.AddChild(
		NewLabel(l.A), NewGamepadInput(g.ui, &c.A),
		NewLabel(l.B), NewGamepadInput(g.ui, &c.B),
		NewLabel(l.Select), NewGamepadInput(g.ui, &c.Select),
		NewLabel(l.Start), NewGamepadInput(g.ui, &c.Start),
		NewLabel(l.Left), NewGamepadInput(g.ui, &c.Left),
		NewLabel(l.Right), NewGamepadInput(g.ui, &c.Right),
		NewLabel(l.Up), NewGamepadInput(g.ui, &c.Up),
		NewLabel(l.Down), NewGamepadInput(g.ui, &c.Down),
		NewLabel(l.L), NewGamepadInput(g.ui, &c.L),
		NewLabel(l.R), NewGamepadInput(g.ui, &c.R),
		NewLabel(l.X), NewGamepadInput(g.ui, &c.X),
		NewLabel(l.Y), NewGamepadInput(g.ui, &c.Y),
		NewLabel(l.Hinge), NewGamepadInput(g.ui, &c.Hinge),
		NewLabel(l.LayoutToggle), NewGamepadInput(g.ui, &c.LayoutToggle),
		NewLabel(l.SizingToggle), NewGamepadInput(g.ui, &c.SizingToggle),
		NewLabel(l.RotationToggle), NewGamepadInput(g.ui, &c.RotationToggle),
		NewLabel(l.ExportToggle), NewGamepadInput(g.ui, &c.ExportScene),
	)

	menu.AddChild(NewSaveButton(l.Save, func(*widget.ButtonClickedEventArgs) {
		config.Conf.Nds = tmp
		config.Conf.Nds.Firmware.Color = config.ColorNameToId[favColor]

		g.ui.content.RemoveChildren()
		g.ui.content.AddChild(NewNdsMenu(g))

		file.Encode()

		if len(g.ui.gamepadIds) != 0 {
			g.ui.focus.FocusLastSubMenu()
		}
		g.ui.toast.AddMessage(g.ui.res.localization.Toast.Saved)
	}))

	g.ui.focus.BuildMenuFocus(g.ui.ui, menu, keys, gamepad)
	return menu
}

func NewAboutMenu(g *Game) *widget.Container {
	l := g.ui.res.localization.Settings.About

	menu := NewMenu()

	about := NewOneCol()
	menu.AddChild(NewHeader(l.About, g.ui.res))
	menu.AddChild(about)

	about.AddChild(NewLinkText(mainLink))
	about.AddChild(NewLinkText(l.Version))
	about.AddChild(NewLinkText(l.Copyright))
	about.AddChild(NewLinkText(l.ThankYous))

	g.ui.focus.BuildMenuFocus(g.ui.ui, menu, nil, nil)

	return menu
}
