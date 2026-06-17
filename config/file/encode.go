package file

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/utils"
	"github.com/hajimehoshi/ebiten/v2"
)

func Encode() {
	f, err := os.Create(CONFIG_PATH)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	c := &Config{
		config: &config.Conf,
	}

	f.Write([]byte("# this file was generated from saving in the menu system\n"))
	f.Write([]byte("# delete this file to return to defaults\n"))
	f.Write([]byte("# please check https://guacemulator.com for documentation\n"))

	c.encodeGeneral()
	c.encodeUi()
	c.encodeProfile()
	c.encodeGb()
	c.encodeGba()
	c.encodeNds()

	t := toml.NewEncoder(f)
	t.Indent = ""

	if err := t.Encode(c); err != nil {
		panic(err)
	}
}

func (c *Config) encodeGeneral() {
	c.General = General{
		Muted:               c.config.General.Muted,
		TargetFps:           c.config.General.TargetFps,
		ShowFps:             c.config.General.ShowFps,
		InitFullscreen:      c.config.General.InitFullscreen,
		Vsync:               c.config.General.Vsync,
		IntegerScaling:      c.config.General.IntegerScaling,
		IntegerScalingRatio: c.config.General.IntegerScalingRatio,
		// rompath
		DisableSaves: c.config.General.DisableSaves,
	}

	file := &c.General.Keyboard
	conf := &c.config.General.Keyboard

	fileKeys := []*[]string{
		&file.Select,
		&file.Return,
		&file.Mute,
		&file.Pause,
		&file.Left,
		&file.Right,
		&file.Up,
		&file.Down,
		&file.Fullscreen,
		&file.Quit,
	}

	confKeys := []*[]ebiten.Key{
		&conf.Select,
		&conf.Return,
		&conf.Mute,
		&conf.Pause,
		&conf.Left,
		&conf.Right,
		&conf.Up,
		&conf.Down,
		&conf.Fullscreen,
		&conf.Quit,
	}

	for i := range confKeys {
		for j := range *confKeys[i] {
			str := utils.KeyToString((*confKeys[i])[j])
			*fileKeys[i] = append(*fileKeys[i], str)
		}
	}

	file = &c.General.Controller
	confB := &c.config.General.Controller

	fileButtons := []*[]string{
		&file.Select,
		&file.Return,
		&file.Mute,
		&file.Pause,
		&file.Left,
		&file.Right,
		&file.Up,
		&file.Down,
		&file.Fullscreen,
		&file.Quit,
	}

	confButtons := []*[]ebiten.StandardGamepadButton{
		&confB.Select,
		&confB.Return,
		&confB.Mute,
		&confB.Pause,
		&confB.Left,
		&confB.Right,
		&confB.Up,
		&confB.Down,
		&confB.Fullscreen,
		&confB.Quit,
	}

	for i := range confButtons {
		for j := range *confButtons[i] {
			str := utils.GamepadButtonToString((*confButtons[i])[j])
			*fileButtons[i] = append(*fileButtons[i], str)
		}
	}
}

func (c *Config) encodeUi() {
	c.Ui.Backdrop = "0x" + utils.ColorToHex(c.config.Ui.Backdrop)
	c.Ui.Background = "0x" + utils.ColorToHex(c.config.Ui.MenuBackgroundColor)
	c.Ui.Foreground = "0x" + utils.ColorToHex(c.config.Ui.MenuForegroundColor)
	c.Ui.Secondary = "0x" + utils.ColorToHex(c.config.Ui.MenuSecondaryColor)

	switch c.config.Ui.Language {
	case 0:
		c.Ui.Language = "en"
	case 1:
		c.Ui.Language = "es"
	}
}

func (c *Config) encodeProfile() {
	c.Profile.Enabled = c.config.Profile.Enabled
	c.Profile.FilePath = c.config.Profile.FilePath
	c.Profile.StartTick = c.config.Profile.StartTick
	c.Profile.EndTick = c.config.Profile.EndTick
}

func (c *Config) encodeGb() {
	c.encodeKeyboard(&c.Gb.Keyboard, &c.config.Gb.KeyboardConfig)
	c.encodeController(&c.Gb.Controller, &c.config.Gb.ControllerConfig)

	c.Gb.Palette = []string{
		"0x" + utils.ColorToHex(c.config.Gb.Palette[0]),
		"0x" + utils.ColorToHex(c.config.Gb.Palette[1]),
		"0x" + utils.ColorToHex(c.config.Gb.Palette[2]),
		"0x" + utils.ColorToHex(c.config.Gb.Palette[3]),
	}
}

func (c *Config) encodeGba() {
	c.encodeKeyboard(&c.Gba.Keyboard, &c.config.Gba.KeyboardConfig)
	c.encodeController(&c.Gba.Controller, &c.config.Gba.ControllerConfig)
	c.Gba.IdleOptimize = c.config.Gba.IdleOptimize
	c.Gba.SoundClockUpdateCycles = c.config.Gba.SoundClockUpdateCycles
}

func (c *Config) encodeNds() {
	c.encodeKeyboard(&c.Nds.Keyboard, &c.config.Nds.KeyboardConfig)
	c.encodeController(&c.Nds.Controller, &c.config.Nds.ControllerConfig)

	if utils.IsFile(c.Nds.Bios.Arm7Path) {
		c.Nds.Bios.Arm7Path = c.config.Nds.Bios.Arm7Path
	}

	if utils.IsFile(c.Nds.Bios.Arm9Path) {
		c.Nds.Bios.Arm9Path = c.config.Nds.Bios.Arm9Path
	}

	c.Nds.Rtc.AdditionalHours = c.config.Nds.Rtc.AdditionalHours

	// if utils.IsDirectory(c.Nds.Export.Directory) {
	c.Nds.Export.Directory = c.config.Nds.Export.Directory
	//}

	c.Nds.Export.ShadowPolys = c.config.Nds.Export.ShadowPolys

	switch c.config.Nds.Screen.Layout {
	case 1:
		c.Nds.Screen.Layout = "horizontal"
	case 2:
		c.Nds.Screen.Layout = "hybrid"
	default:
		c.Nds.Screen.Layout = "vertical"
	}

	switch c.config.Nds.Screen.Sizing {
	case 1:
		c.Nds.Screen.Sizing = "only top"
	case 2:
		c.Nds.Screen.Sizing = "only bottom"
	default:
		c.Nds.Screen.Sizing = "even"
	}

	switch c.config.Nds.Screen.Rotation {
	case 1:
		c.Nds.Screen.Rotation = 90
	case 2:
		c.Nds.Screen.Rotation = 180
	case 3:
		c.Nds.Screen.Rotation = 270
	default:
		c.Nds.Screen.Rotation = 0
	}

	c.encodeNdsFirmware()
	c.encodeNdsJit()
}

func (c *Config) encodeNdsFirmware() {
	f := &c.Nds.Firmware
	conf := &c.config.Nds.Firmware

	if utils.IsFile(conf.FilePath) {
		f.FilePath = conf.FilePath
	}

	f.FavoriteColor = config.ColorNames[conf.Color]

	f.Nickname = conf.Nickname
	f.Message = conf.Message

	//f.BirthdayDay = conf.BirthdayDay
	//f.BirthdayMonth = conf.BirthdayMonth
}

func (c *Config) encodeNdsJit() {
	c.Nds.Jit.BatchInst = c.config.Nds.Jit.BatchInstA9
	c.Nds.Jit.Enabled = c.config.Nds.Jit.Enabled
	c.Nds.Jit.LoopCnt = c.config.Nds.Jit.LoopCnt
	c.Nds.Jit.BlockCnt = c.config.Nds.Jit.BlockCnt
}

func (c *Config) encodeKeyboard(file *EmulatorInput, conf *config.EmulatorKeyboard) {
	files := []*[]string{
		&file.A,
		&file.B,
		&file.Select,
		&file.Start,
		&file.Left,
		&file.Right,
		&file.Up,
		&file.Down,
		&file.R,
		&file.L,
		&file.X,
		&file.Y,
		&file.Hinge,
		&file.Debug,
		&file.LayoutToggle,
		&file.SizingToggle,
		&file.RotationToggle,
		&file.ExportScene,
	}

	confs := []*[]ebiten.Key{
		&conf.A,
		&conf.B,
		&conf.Select,
		&conf.Start,
		&conf.Left,
		&conf.Right,
		&conf.Up,
		&conf.Down,
		&conf.R,
		&conf.L,
		&conf.X,
		&conf.Y,
		&conf.Hinge,
		&conf.Debug,
		&conf.LayoutToggle,
		&conf.SizingToggle,
		&conf.RotationToggle,
		&conf.ExportScene,
	}

	for i := range len(confs) {
		for j := range *confs[i] {
			str := utils.KeyToString((*confs[i])[j])
			*files[i] = append(*files[i], str)
		}
	}
}

func (c *Config) encodeController(file *EmulatorInput, conf *config.EmulatorController) {
	files := []*[]string{
		&file.A,
		&file.B,
		&file.Select,
		&file.Start,
		&file.Left,
		&file.Right,
		&file.Up,
		&file.Down,
		&file.R,
		&file.L,
		&file.X,
		&file.Y,
		&file.Hinge,
		&file.Debug,
		&file.LayoutToggle,
		&file.SizingToggle,
		&file.RotationToggle,
		&file.ExportScene,
	}

	confs := []*[]ebiten.StandardGamepadButton{
		&conf.A,
		&conf.B,
		&conf.Select,
		&conf.Start,
		&conf.Left,
		&conf.Right,
		&conf.Up,
		&conf.Down,
		&conf.R,
		&conf.L,
		&conf.X,
		&conf.Y,
		&conf.Hinge,
		&conf.Debug,
		&conf.LayoutToggle,
		&conf.SizingToggle,
		&conf.RotationToggle,
		&conf.ExportScene,
	}

	for i := range len(confs) {
		for j := range *confs[i] {
			str := utils.GamepadButtonToString((*confs[i])[j])
			*files[i] = append(*files[i], str)
		}
	}
}
