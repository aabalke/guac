package ui

import (
	"github.com/aabalke/guac/common/bus"
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

func newRow(stretch []bool) *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(24)),
			widget.GridLayoutOpts.Spacing(32, 16),
			widget.GridLayoutOpts.Columns(len(stretch)),
			widget.GridLayoutOpts.Stretch(
				stretch,
				[]bool{},
			))))
}

func newOneCol() *widget.Container { return newRow([]bool{true}) }
func newTwoCol() *widget.Container { return newRow([]bool{false, true}) }

func NewGeneralMenu(g *Game) *widget.Container {
	tmp := config.Conf.General
	l := g.ui.res.localization

	menu := newMenu()
	general := newTwoCol()
	menu.AddChild(NewHeader(l["settings.general"], g.ui.res), general)
	general.AddChild(
		NewLabel(l["settings.muted"]), NewCheckbox(&tmp.Muted),
		NewLabel(l["settings.show_fps"]), NewCheckbox(&tmp.ShowFps),
		NewLabel(l["settings.init_fullscreen"]), NewCheckbox(&tmp.InitFullscreen),
		NewLabel(l["settings.target_fps"]), NewDecimalInput(g.ui, l["settings.target_fps"], &tmp.TargetFps, 1_000_000),
		NewLabel(l["settings.vsync_enabled"]), NewCheckbox(&tmp.Vsync),
		NewLabel(l["settings.disable_saves"]), NewCheckbox(&tmp.DisableSaves),
		NewLabel(l["settings.integer_scaling"]), NewCheckbox(&tmp.IntegerScaling),
		NewLabel(l["settings.integer_scaling_ratio"]), NewDecimalInput(g.ui, l["settings.integer_scaling_ratio"], &tmp.IntegerScalingRatio, 10),
		NewSeparator(), NewLinkText(l["settings.integer_scaling_ratio"]),
		NewLabel(l["settings.sample_rate"]), NewDecimalInput(g.ui, l["settings.sample_rate"], &tmp.SampleRate, 192000),
		NewSeparator(), NewLinkText(l["settings.sample_rate_desc"]),
		NewLabel(l["settings.roms_path"]), NewDirectoryInput(&tmp.RomsPath, tmp.RomsPath, "./"),
	)

	inputLabels := []string{
		l["inputs.select"],
		l["inputs.return"],
		l["inputs.mute"],
		l["inputs.pause"],
		l["inputs.left"],
		l["inputs.right"],
		l["inputs.up"],
		l["inputs.down"],
		l["inputs.fullscreen"],
		l["inputs.quit"],
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
	menu.AddChild(NewHeader(l["settings.keyboard"], g.ui.res), keys)
	menu.AddChild(NewHeader(l["settings.controller"], g.ui.res), gamepad)

	for i := range inputLabels {
		keys.AddChild(NewLabel(inputLabels[i]), NewKeybindInput(g.ui, keyPtrs[i]))
		gamepad.AddChild(NewLabel(inputLabels[i]), NewGamepadInput(g.ui, butPtrs[i]))
	}

	menu.AddChild(NewSaveButton(l["settings.save"], func(*widget.ButtonClickedEventArgs) {
		config.Conf.General = tmp

		g.Bus.Publish(bus.Event{
			Type: bus.SET_FPS,
		})

		ebiten.SetVsyncEnabled(config.Conf.General.Vsync)

		g.ui.scrollableMenu.content.RemoveChildren()
		g.ui.scrollableMenu.content.AddChild(NewGeneralMenu(g))

		file.Encode()

		if len(g.ui.gamepadIds) != 0 {
			FocusLast(g.ui)
		}

		g.ui.toast.AddMessage(l["toast.saved"])
	}))

	return menu
}

func NewUiMenu(g *Game) *widget.Container {
	var (
		res   = g.ui.res
		tmp   = config.Conf.Ui
		oldId = g.ui.PrevPageId

		l = g.ui.res.localization

		clrInputs = [4]widget.PreferredSizeLocateableWidget{
			NewColorInput(g.ui, l["settings.backdrop"], &tmp.Backdrop, HexValidation(0xFFFFFF)),
			NewColorInput(g.ui, l["settings.bg_color"], &tmp.MenuBackgroundColor, HexValidation(0xFFFFFF)),
			NewColorInput(g.ui, l["settings.fg_color"], &tmp.MenuForegroundColor, HexValidation(0xFFFFFF)),
			NewColorInput(g.ui, l["settings.accent_color"], &tmp.MenuSecondaryColor, HexValidation(0xFFFFFF)),
		}
	)

	menu := newMenu()

	ui := newTwoCol()
	menu.AddChild(NewHeader(l["settings.ui"], res), ui)

	ui.AddChild(
		NewLabel(l["settings.language"]), NewRadioInput(&tmp.Language, l.AsSlice("settings.languages"), res),
		NewLabel(l["settings.backdrop"]), clrInputs[0],
		NewLabel(l["settings.bg_color"]), clrInputs[1],
		NewLabel(l["settings.fg_color"]), clrInputs[2],
		NewLabel(l["settings.accent_color"]), clrInputs[3],
		NewLabel(l["settings.apply_theme"]),
		NewApplyPalettesMenu(theme_palettes, clrInputs, res),
	)

	menu.AddChild(NewSaveButton(l["settings.save"], func(*widget.ButtonClickedEventArgs) {
		config.Conf.Ui = tmp
		g.ui.res.localization = NewLocalization(LangOptions(config.Conf.Ui.Language))

		res.Update()
		g.ui.keyboard = NewKeyboard(g.ui.res, l.AsSlice("settings.alphabet"))

		NewSettings(g, oldId, MENU_UI)

		file.Encode()

		if len(g.ui.gamepadIds) != 0 {
			FocusLast(g.ui)
		}

		g.ui.toast.AddMessage(l["toast.saved"])
	}))

	return menu
}

func NewGbMenu(g *Game) *widget.Container {
	tmp := config.Conf.Gb
	l := g.ui.res.localization
	pal := &tmp.Palette
	clrInputs := [4]widget.PreferredSizeLocateableWidget{
		NewColorInput(g.ui, l["settings.dmg_lightest"], &pal[0], HexValidation(0xFFFFFF)),
		NewColorInput(g.ui, l["settings.dmg_light"], &pal[1], HexValidation(0xFFFFFF)),
		NewColorInput(g.ui, l["settings.dmg_dark"], &pal[2], HexValidation(0xFFFFFF)),
		NewColorInput(g.ui, l["settings.dmg_darkest"], &pal[3], HexValidation(0xFFFFFF)),
	}

	menu := newMenu()

	general := newTwoCol()
	menu.AddChild(NewHeader(l["settings.general"], g.ui.res), general)
	general.AddChild(NewLabel(l["settings.system"]), NewRadioInput(&tmp.System, l.AsSlice("settings.systems"), g.ui.res))

	graphics := newTwoCol()
	menu.AddChild(graphics)
	graphics.AddChild(NewLabel(l["settings.rotation"]), NewRadioInput(&tmp.Rotation, l.AsSlice("settings.rotations"), g.ui.res))

	palette := newTwoCol()
	menu.AddChild(NewHeader(l["settings.dmg_palette"], g.ui.res), palette)
	palette.AddChild(
		NewLabel(l["settings.dmg_lightest"]), clrInputs[0],
		NewLabel(l["settings.dmg_light"]), clrInputs[1],
		NewLabel(l["settings.dmg_dark"]), clrInputs[2],
		NewLabel(l["settings.dmg_darkest"]), clrInputs[3],
		NewLabel(l["settings.apply_palette"]),
		NewApplyPalettesMenu(dmg_palettes, clrInputs, g.ui.res),
	)

	cc := newTwoCol()
	strength := int(tmp.ColorCorrection.Strength * 100)
	ghosting := tmp.ColorCorrection.ScreenGhosting
	menu.AddChild(NewHeader(l["settings.color_correction"], g.ui.res), cc)
	cc.AddChild(NewLabel(l["settings.color_correction_type"]), NewRadioInput(&tmp.ColorCorrection.Type, l.AsSlice("settings.color_correction_types"), g.ui.res))
	cc.AddChild(NewLabel(l["settings.color_correction_strength"]), NewDecimalInput(g.ui, l["color_correction_strength"], &strength, 100))
	cc.AddChild(NewLabel(l["settings.color_correction_screen_ghosting"]), NewCheckbox(&ghosting))

	bios := newTwoCol()
	menu.AddChild(NewHeader(l["settings.bios"], g.ui.res), bios)
	bios.AddChild(
		NewLabel(l["settings.dmg_path"]), NewFileInput(&tmp.Bios.DmgPath),
		NewLabel(l["settings.gbc_path"]), NewFileInput(&tmp.Bios.GbcPath),
		NewLabel(l["settings.direct_boot"]), NewCheckbox(&tmp.Bios.Direct),
	)

	inputLabels := []string{
		l["inputs.a"],
		l["inputs.b"],
		l["inputs.select"],
		l["inputs.start"],
		l["inputs.left"],
		l["inputs.right"],
		l["inputs.up"],
		l["inputs.down"],
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
	menu.AddChild(NewHeader(l["settings.keyboard"], g.ui.res), keys)
	menu.AddChild(NewHeader(l["settings.controller"], g.ui.res), gamepad)

	for i := range inputLabels {
		keys.AddChild(NewLabel(inputLabels[i]), NewKeybindInput(g.ui, keyPtrs[i]))
		gamepad.AddChild(NewLabel(inputLabels[i]), NewGamepadInput(g.ui, butPtrs[i]))
	}

	menu.AddChild(NewSaveButton(l["settings.save"], func(*widget.ButtonClickedEventArgs) {
		config.Conf.Gb = tmp
		config.Conf.Gb.ColorCorrection.Strength = float32(strength) / 100

		config.Conf.Gb.ColorCorrection.ScreenGhosting = ghosting
		if g.gb != nil {
			g.gb.UpdateGhosting()
		}

		g.ui.scrollableMenu.content.RemoveChildren()
		g.ui.scrollableMenu.content.AddChild(NewGbMenu(g))

		file.Encode()

		if len(g.ui.gamepadIds) != 0 {
			FocusLast(g.ui)
		}

		g.ui.toast.AddMessage(l["toast.saved"])
	}))

	return menu
}

func NewGbaMenu(g *Game) *widget.Container {
	tmp := config.Conf.Gba
	l := g.ui.res.localization

	menu := newMenu()

	general := newTwoCol()
	menu.AddChild(NewHeader(l["settings.general"], g.ui.res), general)

	general.AddChild(
		NewLabel(l["settings.optimize_idle_loops"]), NewCheckbox(&tmp.IdleOptimize),
		NewLabel(l["settings.rotation"]), NewRadioInput(&tmp.Rotation, l.AsSlice("settings.rotations"), g.ui.res),
	)

	hardware := newTwoCol()
	menu.AddChild(NewHeader(l["settings.hardware"], g.ui.res), hardware)

	hardware.AddChild(
		NewLabel(l["settings.backup_type"]), NewRadioInput(&tmp.Hardware.BackupType, l.AsSlice("settings.backup_types"), g.ui.res),
		NewLabel(l["settings.force_rtc"]), NewCheckbox(&tmp.Hardware.ForceRtc),
		NewLabel(l["settings.force_solar_sensor"]), NewCheckbox(&tmp.Hardware.ForceSolarSensor),
		NewLabel(l["solar_sensor_level"]), NewDecimalInput(g.ui, l["solar_sensor_level"], &tmp.Hardware.SolarSensorLevel, 100),
	)

	cc := newTwoCol()
	strength := int(tmp.ColorCorrection.Strength * 100)
	ghosting := tmp.ColorCorrection.ScreenGhosting
	menu.AddChild(NewHeader(l["settings.color_correction"], g.ui.res), cc)
	cc.AddChild(NewLabel(l["settings.color_correction_type"]), NewRadioInput(&tmp.ColorCorrection.Type, l.AsSlice("settings.color_correction_types"), g.ui.res))
	cc.AddChild(NewLabel(l["settings.color_correction_strength"]), NewDecimalInput(g.ui, l["color_correction_strength"], &strength, 100))
	cc.AddChild(NewLabel(l["settings.color_correction_screen_ghosting"]), NewCheckbox(&ghosting))

	bios := newTwoCol()
	menu.AddChild(NewHeader(l["settings.bios"], g.ui.res), bios)

	bios.AddChild(
		NewLabel(l["settings.bios_path"]), NewFileInput(&tmp.Bios.Path),
		NewLabel(l["settings.direct_boot"]), NewCheckbox(&tmp.Bios.Direct),
	)

	inputLabels := []string{
		l["inputs.a"],
		l["inputs.b"],
		l["inputs.select"],
		l["inputs.start"],
		l["inputs.left"],
		l["inputs.right"],
		l["inputs.up"],
		l["inputs.down"],
		l["inputs.l"],
		l["inputs.r"],
		l["settings.solar_minimum"],
		l["settings.solar_1"],
		l["settings.solar_2"],
		l["settings.solar_3"],
		l["settings.solar_maximum"],
		l["settings.rotation_toggle"],
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
	menu.AddChild(NewHeader(l["settings.keyboard"], g.ui.res), keys)
	menu.AddChild(NewHeader(l["settings.controller"], g.ui.res), gamepad)

	for i := range inputLabels {
		keys.AddChild(NewLabel(inputLabels[i]), NewKeybindInput(g.ui, keyPtrs[i]))
		gamepad.AddChild(NewLabel(inputLabels[i]), NewGamepadInput(g.ui, butPtrs[i]))
	}

	menu.AddChild(NewSaveButton(l["settings.save"], func(*widget.ButtonClickedEventArgs) {
		if tmp.Bios.Path == "" {
			tmp.Bios.Direct = true
		}

		config.Conf.Gba = tmp
		config.Conf.Gba.ColorCorrection.Strength = float32(strength) / 100
		config.Conf.Gba.ColorCorrection.ScreenGhosting = ghosting
		if g.gba != nil {
			g.gba.UpdateGhosting()

			g.gba.IdleOptimize = tmp.IdleOptimize

			if g.gba.Mem.Gpio != nil {
				for _, device := range g.gba.Mem.Gpio.Devices {
					if d, ok := device.(*gpio.Solar); ok {
						d.SetLevel(uint8(tmp.Hardware.SolarSensorLevel))
					}
				}
			}
		}

		g.ui.scrollableMenu.content.RemoveChildren()
		g.ui.scrollableMenu.content.AddChild(NewGbaMenu(g))

		file.Encode()

		if len(g.ui.gamepadIds) != 0 {
			FocusLast(g.ui)
		}

		g.ui.toast.AddMessage(l["toast.saved"])
	}))

	return menu
}

func NewNdsMenu(g *Game) *widget.Container {
	tmp := config.Conf.Nds
	l := g.ui.res.localization

	menu := newMenu()

	screen := newTwoCol()
	menu.AddChild(NewHeader(l["settings.screen"], g.ui.res), screen)

	screen.AddChild(
		NewLabel(l["settings.layout"]), NewRadioInput(&tmp.Screen.Layout, l.AsSlice("settings.layouts"), g.ui.res),
		NewLabel(l["settings.sizing"]), NewRadioInput(&tmp.Screen.Sizing, l.AsSlice("settings.sizings"), g.ui.res),
		NewLabel(l["settings.rotation"]), NewRadioInput(&tmp.Screen.Rotation, l.AsSlice("settings.rotations"), g.ui.res),
	)

	rtc := newTwoCol()
	menu.AddChild(NewHeader(l["settings.rtc"], g.ui.res), rtc)

	rtc.AddChild(
		NewLabel(l["settings.additional_hours"]), NewDecimalInput(g.ui, l["settings.additional_hours"], &tmp.Rtc.AdditionalHours, 24),
	)

	cc := newTwoCol()
	strength := int(tmp.ColorCorrection.Strength * 100)
	ghosting := tmp.ColorCorrection.ScreenGhosting
	menu.AddChild(NewHeader(l["settings.color_correction"], g.ui.res), cc)
	cc.AddChild(NewLabel(l["settings.color_correction_type"]), NewRadioInput(&tmp.ColorCorrection.Type, l.AsSlice("settings.color_correction_types"), g.ui.res))
	cc.AddChild(NewLabel(l["settings.color_correction_strength"]), NewDecimalInput(g.ui, l["color_correction_strength"], &strength, 100))
	cc.AddChild(NewLabel(l["settings.color_correction_screen_ghosting"]), NewCheckbox(&ghosting))

	bios := newTwoCol()
	menu.AddChild(NewHeader(l["settings.bios"], g.ui.res), bios)

	bios.AddChild(
		NewLabel(l["settings.arm7_path"]), NewFileInput(&tmp.Bios.Arm7Path),
		NewLabel(l["settings.arm9_path"]), NewFileInput(&tmp.Bios.Arm9Path),
	)

	firmware := newTwoCol()
	menu.AddChild(NewHeader(l["settings.firmware"], g.ui.res), firmware)
	favColor := config.ColorNames[tmp.Firmware.Color]

	firmware.AddChild(
		NewLabel(l["settings.file_path"]), NewFileInput(&tmp.Firmware.FilePath),
		NewLabel(l["settings.nickname"]), NewTextBoxInput(g.ui, BOARD_ALPHA, l["settings.nickname"], &tmp.Firmware.Nickname, StringValidation(10)),
		NewLabel(l["settings.message"]), NewTextBoxInput(g.ui, BOARD_ALPHA, l["settings.message"], &tmp.Firmware.Message, StringValidation(26)),
		NewLabel(l["settings.favorite_color"]), NewTextBoxInput(g.ui, BOARD_ALPHA, l["settings.favorite_color"], &favColor, ColorNdsValidation()),
	)

	export := newTwoCol()
	menu.AddChild(NewHeader(l["settings.scene_export"], g.ui.res), export)

	export.AddChild(
		NewLabel(l["settings.output_directory"]), NewDirectoryInput(&tmp.Export.Directory, tmp.Export.Directory, "./export"),
		NewLabel(l["settings.shadow_polygons"]), NewCheckbox(&tmp.Export.ShadowPolys),
	)

	inputLabels := []string{
		l["inputs.a"],
		l["inputs.b"],
		l["inputs.select"],
		l["inputs.start"],
		l["inputs.left"],
		l["inputs.right"],
		l["inputs.up"],
		l["inputs.down"],
		l["inputs.l"],
		l["inputs.r"],
		l["inputs.x"],
		l["inputs.y"],
		l["settings.hinge"],
		l["settings.layout_toggle"],
		l["settings.sizing_toggle"],
		l["settings.rotation_toggle"],
		l["settings.export_toggle"],
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
	menu.AddChild(NewHeader(l["settings.keyboard"], g.ui.res), keys)
	menu.AddChild(NewHeader(l["settings.controller"], g.ui.res), gamepad)

	for i := range inputLabels {
		keys.AddChild(NewLabel(inputLabels[i]), NewKeybindInput(g.ui, keyPtrs[i]))
		gamepad.AddChild(NewLabel(inputLabels[i]), NewGamepadInput(g.ui, butPtrs[i]))
	}

	menu.AddChild(NewSaveButton(l["settings.save"], func(*widget.ButtonClickedEventArgs) {
		config.Conf.Nds = tmp
		config.Conf.Nds.Firmware.Color = config.ColorNameToId[favColor]
		config.Conf.Nds.ColorCorrection.Strength = float32(strength) / 100
		config.Conf.Nds.ColorCorrection.ScreenGhosting = ghosting
		if g.nds != nil {
			g.nds.Screen.UpdateGhosting()
		}

		g.ui.scrollableMenu.content.RemoveChildren()
		g.ui.scrollableMenu.content.AddChild(NewNdsMenu(g))

		file.Encode()

		if len(g.ui.gamepadIds) != 0 {
			FocusLast(g.ui)
		}

		g.ui.toast.AddMessage(l["toast.saved"])
	}))

	return menu
}

func NewAboutMenu(g *Game) *widget.Container {
	l := g.ui.res.localization

	menu := newMenu()

	about := newOneCol()
	menu.AddChild(NewHeader(l["settings.about"], g.ui.res))
	menu.AddChild(about)

	about.AddChild(NewLinkText(mainLink))
	about.AddChild(NewLinkText(l["settings.version"]))
	about.AddChild(NewLinkText(l["settings.copyright"]))
	about.AddChild(NewLinkText(l["settings.thank_yous"]))

	return menu
}
