package ui

import (
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/config/file"
	"github.com/aabalke/guac/emu/gba/gpio"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

func newMenu() *widget.Container {
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

func newOneCol() *widget.Container {
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

func newTwoCol() *widget.Container {
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

	menu := newMenu()

	general := newTwoCol()
	menu.AddChild(NewHeader(l.General, g.ui.res), general)

	general.AddChild(
		NewLabel(l.Muted), NewCheckbox(&tmp.Muted),
		NewLabel(l.ShowFps), NewCheckbox(&tmp.ShowFps),
		NewLabel(l.InitFullscreen), NewCheckbox(&tmp.InitFullscreen),
		NewLabel(l.TargetFps), NewDecimalInput(g.ui, l.TargetFps, &tmp.TargetFps, 1_000_000),
		NewLabel(l.VsyncEnabled), NewCheckbox(&tmp.Vsync),
		NewLabel(l.DisableSaves), NewCheckbox(&tmp.DisableSaves),
		NewLabel(l.IntegerScaling), NewCheckbox(&tmp.IntegerScaling),
		NewLabel(l.IntegerScalingRatio), NewDecimalInput(g.ui, l.IntegerScalingRatio, &tmp.IntegerScalingRatio, 10),
		NewSeparator(), NewLinkText(l.IntegerScalingDesc),
		NewLabel(l.SampleRate), NewDecimalInput(g.ui, l.SampleRate, &tmp.SampleRate, 192000),
		NewSeparator(), NewLinkText(l.SampleRateDesc),
	)

	inputLabels := []string{
		l.Select,
		l.Return,
		l.Mute,
		l.Pause,
		l.Left,
		l.Right,
		l.Up,
		l.Down,
		l.Fullscreen,
		l.Quit,
	}

	keyPtrs := []*[]ebiten.Key{
		&tmp.Keyboard.Select,
		&tmp.Keyboard.Return,
		&tmp.Keyboard.Mute,
		&tmp.Keyboard.Pause,
		&tmp.Keyboard.Left,
		&tmp.Keyboard.Right,
		&tmp.Keyboard.Up,
		&tmp.Keyboard.Down,
		&tmp.Keyboard.Fullscreen,
		&tmp.Keyboard.Quit,
	}

	butPtrs := []*[]ebiten.StandardGamepadButton{
		&tmp.Controller.Select,
		&tmp.Controller.Return,
		&tmp.Controller.Mute,
		&tmp.Controller.Pause,
		&tmp.Controller.Left,
		&tmp.Controller.Right,
		&tmp.Controller.Up,
		&tmp.Controller.Down,
		&tmp.Controller.Fullscreen,
		&tmp.Controller.Quit,
	}

	keys, gamepad := newTwoCol(), newTwoCol()
	menu.AddChild(NewHeader(l.Keyboard, g.ui.res), keys)
	menu.AddChild(NewHeader(l.Controller, g.ui.res), gamepad)

	for i := range inputLabels {
		keys.AddChild(NewLabel(inputLabels[i]), NewKeybindInput(g.ui, keyPtrs[i]))
		gamepad.AddChild(NewLabel(inputLabels[i]), NewGamepadInput(g.ui, butPtrs[i]))
	}

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

	g.ui.focus.BuildMenuFocus(menu, keys, gamepad)

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

	menu := newMenu()

	ui := newTwoCol()
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

			file.Encode()

			if len(g.ui.gamepadIds) != 0 {
				g.ui.focus.FocusLastSubMenu()
			}

			g.ui.toast.AddMessage(g.ui.res.localization.Toast.Saved)
		}),
	)

	g.ui.focus.BuildMenuFocus(menu, nil, nil)

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

	menu := newMenu()

	general := newTwoCol()
	menu.AddChild(NewHeader(l.General, g.ui.res), general)
	general.AddChild(
		NewLabel(l.System), NewRadioInput(&g.ui.focus.horizontalGroup, &tmp.System, l.Systems, g.ui.res),
	)

	palette := newTwoCol()
	menu.AddChild(NewHeader(l.DmgPalette, g.ui.res), palette)
	palette.AddChild(
		NewLabel(l.Lightest), clrInputs[0],
		NewLabel(l.Light), clrInputs[1],
		NewLabel(l.Dark), clrInputs[2],
		NewLabel(l.Darkest), clrInputs[3],
		NewLabel(l.ApplyPalette),
		NewApplyPalettesMenu(&g.ui.focus.horizontalGroup, dmg_palettes, clrInputs, g.ui.res),
	)

	bios := newTwoCol()
	menu.AddChild(NewHeader(l.Bios, g.ui.res), bios)
	bios.AddChild(
		NewLabel(l.DmgPath), NewFileInput(&tmp.Bios.DmgPath),
		NewLabel(l.GbcPath), NewFileInput(&tmp.Bios.GbcPath),
		NewLabel(l.DirectBoot), NewCheckbox(&tmp.Bios.Direct),
	)

	inputLabels := []string{
		l.A,
		l.B,
		l.Select,
		l.Start,
		l.Left,
		l.Right,
		l.Up,
		l.Down,
	}

	keyPtrs := []*[]ebiten.Key{
		&tmp.Keyboard.A,
		&tmp.Keyboard.B,
		&tmp.Keyboard.Select,
		&tmp.Keyboard.Start,
		&tmp.Keyboard.Left,
		&tmp.Keyboard.Right,
		&tmp.Keyboard.Up,
		&tmp.Keyboard.Down,
	}

	butPtrs := []*[]ebiten.StandardGamepadButton{
		&tmp.Controller.A,
		&tmp.Controller.B,
		&tmp.Controller.Select,
		&tmp.Controller.Start,
		&tmp.Controller.Left,
		&tmp.Controller.Right,
		&tmp.Controller.Up,
		&tmp.Controller.Down,
	}

	keys, gamepad := newTwoCol(), newTwoCol()
	menu.AddChild(NewHeader(l.Keyboard, g.ui.res), keys)
	menu.AddChild(NewHeader(l.Controller, g.ui.res), gamepad)

	for i := range inputLabels {
		keys.AddChild(NewLabel(inputLabels[i]), NewKeybindInput(g.ui, keyPtrs[i]))
		gamepad.AddChild(NewLabel(inputLabels[i]), NewGamepadInput(g.ui, butPtrs[i]))
	}

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

	g.ui.focus.BuildMenuFocus(menu, keys, gamepad)

	return menu
}

func NewGbaMenu(g *Game) *widget.Container {
	tmp := config.Conf.Gba
	l := g.ui.res.localization.Settings.Gba

	menu := newMenu()

	general := newTwoCol()
	menu.AddChild(NewHeader(l.General, g.ui.res), general)

	general.AddChild(
		NewLabel(l.OptmizeIdleLoops), NewCheckbox(&tmp.IdleOptimize),
		NewLabel(l.Rotation), NewRadioInput(&g.ui.focus.horizontalGroup, &tmp.Rotation, l.Rotations, g.ui.res),
	)

	hardware := newTwoCol()
	menu.AddChild(NewHeader(l.Hardware, g.ui.res), hardware)

	hardware.AddChild(
		NewLabel(l.BackupType), NewRadioInput(&g.ui.focus.horizontalGroup, &tmp.Hardware.BackupType, l.BackupTypes, g.ui.res),
		NewLabel(l.ForceRtc), NewCheckbox(&tmp.Hardware.ForceRtc),
		NewLabel(l.ForceSolarSensor), NewCheckbox(&tmp.Hardware.ForceSolarSensor),
		NewLabel(l.SolarSensorLevel), NewDecimalInput(g.ui, l.SolarSensorLevel, &tmp.Hardware.SolarSensorLevel, 100),
	)

	bios := newTwoCol()
	menu.AddChild(NewHeader(l.Bios, g.ui.res), bios)

	bios.AddChild(
		NewLabel(l.BiosPath), NewFileInput(&tmp.Bios.Path),
		NewLabel(l.DirectBoot), NewCheckbox(&tmp.Bios.Direct),
	)

	inputLabels := []string{
		l.A,
		l.B,
		l.Select,
		l.Start,
		l.Left,
		l.Right,
		l.Up,
		l.Down,
		l.L,
		l.R,
		l.SolarMin,
		l.Solar1,
		l.Solar2,
		l.Solar3,
		l.SolarMax,
		l.RotationToggle,
	}

	keyPtrs := []*[]ebiten.Key{
		&tmp.Keyboard.A,
		&tmp.Keyboard.B,
		&tmp.Keyboard.Select,
		&tmp.Keyboard.Start,
		&tmp.Keyboard.Left,
		&tmp.Keyboard.Right,
		&tmp.Keyboard.Up,
		&tmp.Keyboard.Down,
		&tmp.Keyboard.L,
		&tmp.Keyboard.R,
		&tmp.Keyboard.SolarLevel0,
		&tmp.Keyboard.SolarLevel1,
		&tmp.Keyboard.SolarLevel2,
		&tmp.Keyboard.SolarLevel3,
		&tmp.Keyboard.SolarLevel4,
		&tmp.Keyboard.RotationToggle,
	}

	butPtrs := []*[]ebiten.StandardGamepadButton{
		&tmp.Controller.A,
		&tmp.Controller.B,
		&tmp.Controller.Select,
		&tmp.Controller.Start,
		&tmp.Controller.Left,
		&tmp.Controller.Right,
		&tmp.Controller.Up,
		&tmp.Controller.Down,
		&tmp.Controller.L,
		&tmp.Controller.R,
		&tmp.Controller.SolarLevel0,
		&tmp.Controller.SolarLevel1,
		&tmp.Controller.SolarLevel2,
		&tmp.Controller.SolarLevel3,
		&tmp.Controller.SolarLevel4,
		&tmp.Controller.RotationToggle,
	}

	keys, gamepad := newTwoCol(), newTwoCol()
	menu.AddChild(NewHeader(l.Keyboard, g.ui.res), keys)
	menu.AddChild(NewHeader(l.Controller, g.ui.res), gamepad)

	for i := range inputLabels {
		keys.AddChild(NewLabel(inputLabels[i]), NewKeybindInput(g.ui, keyPtrs[i]))
		gamepad.AddChild(NewLabel(inputLabels[i]), NewGamepadInput(g.ui, butPtrs[i]))
	}

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

	g.ui.focus.BuildMenuFocus(menu, keys, gamepad)

	return menu
}

func NewNdsMenu(g *Game) *widget.Container {
	tmp := config.Conf.Nds
	l := g.ui.res.localization.Settings.Nds

	menu := newMenu()

	screen := newTwoCol()
	menu.AddChild(NewHeader(l.Screen, g.ui.res), screen)

	screen.AddChild(
		NewLabel(l.Layout), NewRadioInput(&g.ui.focus.horizontalGroup, &tmp.Screen.Layout, l.Layouts, g.ui.res),
		NewLabel(l.Sizing), NewRadioInput(&g.ui.focus.horizontalGroup, &tmp.Screen.Sizing, l.Sizings, g.ui.res),
		NewLabel(l.Rotation), NewRadioInput(&g.ui.focus.horizontalGroup, &tmp.Screen.Rotation, l.Rotations, g.ui.res),
	)

	rtc := newTwoCol()
	menu.AddChild(NewHeader(l.Rtc, g.ui.res), rtc)

	rtc.AddChild(
		NewLabel(l.AdditionalHours), NewDecimalInput(g.ui, l.AdditionalHours, &tmp.Rtc.AdditionalHours, 24),
	)

	bios := newTwoCol()
	menu.AddChild(NewHeader(l.Bios, g.ui.res), bios)

	bios.AddChild(
		NewLabel(l.Arm7Path), NewFileInput(&tmp.Bios.Arm7Path),
		NewLabel(l.Arm9Path), NewFileInput(&tmp.Bios.Arm9Path),
	)

	firmware := newTwoCol()
	menu.AddChild(NewHeader(l.Firmware, g.ui.res), firmware)
	favColor := config.ColorNames[tmp.Firmware.Color]

	firmware.AddChild(
		NewLabel(l.FilePath), NewFileInput(&tmp.Firmware.FilePath),
		NewLabel(l.Nickname), NewTextBoxInput(g.ui, BOARD_ALPHA, l.Nickname, &tmp.Firmware.Nickname, StringValidation(10)),
		NewLabel(l.Message), NewTextBoxInput(g.ui, BOARD_ALPHA, l.Message, &tmp.Firmware.Message, StringValidation(26)),
		NewLabel(l.FavoriteColor), NewTextBoxInput(g.ui, BOARD_ALPHA, l.FavoriteColor, &favColor, ColorNdsValidation()),
	)

	export := newTwoCol()
	menu.AddChild(NewHeader(l.SceneExport, g.ui.res), export)

	export.AddChild(
		NewLabel(l.OutputDirectory), NewDirectoryInput(&tmp.Export.Directory, "./export"),
		NewLabel(l.ShadowPolygons), NewCheckbox(&tmp.Export.ShadowPolys),
	)

	inputLabels := []string{
		l.A,
		l.B,
		l.Select,
		l.Start,
		l.Left,
		l.Right,
		l.Up,
		l.Down,
		l.L,
		l.R,
		l.X,
		l.Y,
		l.Hinge,
		l.LayoutToggle,
		l.SizingToggle,
		l.RotationToggle,
		l.ExportToggle,
	}

	keyPtrs := []*[]ebiten.Key{
		&tmp.Keyboard.A,
		&tmp.Keyboard.B,
		&tmp.Keyboard.Select,
		&tmp.Keyboard.Start,
		&tmp.Keyboard.Left,
		&tmp.Keyboard.Right,
		&tmp.Keyboard.Up,
		&tmp.Keyboard.Down,
		&tmp.Keyboard.L,
		&tmp.Keyboard.R,
		&tmp.Keyboard.X,
		&tmp.Keyboard.Y,
		&tmp.Keyboard.Hinge,
		&tmp.Keyboard.LayoutToggle,
		&tmp.Keyboard.SizingToggle,
		&tmp.Keyboard.RotationToggle,
		&tmp.Keyboard.ExportScene,
	}

	butPtrs := []*[]ebiten.StandardGamepadButton{
		&tmp.Controller.A,
		&tmp.Controller.B,
		&tmp.Controller.Select,
		&tmp.Controller.Start,
		&tmp.Controller.Left,
		&tmp.Controller.Right,
		&tmp.Controller.Up,
		&tmp.Controller.Down,
		&tmp.Controller.L,
		&tmp.Controller.R,
		&tmp.Controller.X,
		&tmp.Controller.Y,
		&tmp.Controller.Hinge,
		&tmp.Controller.LayoutToggle,
		&tmp.Controller.SizingToggle,
		&tmp.Controller.RotationToggle,
		&tmp.Controller.ExportScene,
	}

	keys, gamepad := newTwoCol(), newTwoCol()
	menu.AddChild(NewHeader(l.Keyboard, g.ui.res), keys)
	menu.AddChild(NewHeader(l.Controller, g.ui.res), gamepad)

	for i := range inputLabels {
		keys.AddChild(NewLabel(inputLabels[i]), NewKeybindInput(g.ui, keyPtrs[i]))
		gamepad.AddChild(NewLabel(inputLabels[i]), NewGamepadInput(g.ui, butPtrs[i]))
	}

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

	g.ui.focus.BuildMenuFocus(menu, keys, gamepad)
	return menu
}

func NewAboutMenu(g *Game) *widget.Container {
	l := g.ui.res.localization.Settings.About

	menu := newMenu()

	about := newOneCol()
	menu.AddChild(NewHeader(l.About, g.ui.res))
	menu.AddChild(about)

	about.AddChild(NewLinkText(mainLink))
	about.AddChild(NewLinkText(l.Version))
	about.AddChild(NewLinkText(l.Copyright))
	about.AddChild(NewLinkText(l.ThankYous))

	g.ui.focus.BuildMenuFocus(menu, nil, nil)

	return menu
}
