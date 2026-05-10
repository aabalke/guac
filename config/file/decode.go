package file

import (
	"image/color"
	"os"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/utils"
	"github.com/hajimehoshi/ebiten/v2"
)

func Decode() {
	c := &Config{config: &config.Conf}
	c.readFile(CONFIG_PATH)
	c.decodeGeneral()
	c.decodeUi()
	c.decodeProfile()
	c.decodeGb()
	c.decodeGba()
	c.decodeNds()
}

func (c *Config) readFile(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}

		f, err := os.Create(path)
		if err != nil {
			panic(err)
		}

		_, err = f.Write(DEF_CONFIG)
		if err != nil {
			panic(err)
		}

		b = DEF_CONFIG
	}

	_, err = toml.Decode(string(b), &c)
	if err != nil {
		panic(err)
	}
}

func (c *Config) decodeGeneral() {
	c.config.General.Muted = c.General.Muted
	c.config.General.TargetFps = c.General.TargetFps
	c.config.General.ShowFps = c.General.ShowFps
	c.config.General.InitFullscreen = c.General.InitFullscreen
	c.config.General.Vsync = c.General.Vsync
	c.config.General.RomPath = c.General.RomPath
	c.config.General.DisableSaves = c.General.DisableSaves
	c.config.General.IntegerScaling = c.General.IntegerScaling
	c.config.General.IntegerScalingRatio = c.General.IntegerScalingRatio

	in := &c.General.Keyboard
	confKey := &c.config.General.Keyboard

	tomls := []*[]string{
		&in.Select,
		&in.Return,
		&in.Mute,
		&in.Pause,
		&in.Left,
		&in.Right,
		&in.Up,
		&in.Down,
		&in.Fullscreen,
		&in.Quit,
	}

	outputsKeys := []*[]ebiten.Key{
		&confKey.Select,
		&confKey.Return,
		&confKey.Mute,
		&confKey.Pause,
		&confKey.Left,
		&confKey.Right,
		&confKey.Up,
		&confKey.Down,
		&confKey.Fullscreen,
		&confKey.Quit,
	}

	for i := range len(tomls) {
		for j := range *tomls[i] {
			if str, ok := utils.StringToKey((*tomls[i])[j]); ok {
				*outputsKeys[i] = append(*outputsKeys[i], str)
			}
		}
	}

	in = &c.General.Controller
	conf := &c.config.General.Controller

	tomls = []*[]string{
		&in.Select,
		&in.Return,
		&in.Mute,
		&in.Pause,
		&in.Left,
		&in.Right,
		&in.Up,
		&in.Down,
		&in.Fullscreen,
		&in.Quit,
	}

	outputs := []*[]ebiten.StandardGamepadButton{
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

	for i := range len(tomls) {
		for j := range len(*tomls[i]) {
			if str, ok := utils.StringToGamepadButton((*tomls[i])[j]); ok {
				*outputs[i] = append(*outputs[i], str)
			}
		}
	}
}

func (c *Config) decodeUi() {
	c.config.Ui.Backdrop = utils.HexToColor(c.Ui.Backdrop)
	c.config.Ui.MenuBackgroundColor = utils.HexToColor(c.Ui.Background)
	c.config.Ui.MenuForegroundColor = utils.HexToColor(c.Ui.Foreground)
	c.config.Ui.MenuSecondaryColor = utils.HexToColor(c.Ui.Secondary)

	switch c.Ui.Language {
	case "spanish", "es": // english
		c.config.Ui.Language = 1
	default:
		c.config.Ui.Language = 0
	}
}

func (c *Config) decodeProfile() {
	c.config.Profile.Enabled = c.Profile.Enabled
	c.config.Profile.FilePath = c.Profile.FilePath
	c.config.Profile.StartTick = c.Profile.StartTick
	c.config.Profile.EndTick = c.Profile.EndTick

	if c.Profile.StartTick >= c.Profile.EndTick {
		panic("profile config invalid, provided Start >= End")
	}
}

func (c *Config) decodeGb() {
	c.decodeKeyboard(&c.Gb.Keyboard, &c.config.Gb.KeyboardConfig)
	c.decodeController(&c.Gb.Controller, &c.config.Gb.ControllerConfig)

	var pals []color.Color

	for i := range c.Gb.Palette {
		pals = append(pals, utils.HexToColor(c.Gb.Palette[i]))
	}

	if invalid := len(pals) != 4; invalid {
		// greyscale palette
		c.config.Gb.Palette = [4]color.Color{
			color.RGBA{0xFF, 0xFF, 0xFF, 0xFF},
			color.RGBA{0xCC, 0xCC, 0xCC, 0xFF},
			color.RGBA{0x77, 0x77, 0x77, 0xFF},
			color.RGBA{0x00, 0x00, 0x00, 0xFF},
		}
	} else {
		c.config.Gb.Palette[0] = pals[0]
		c.config.Gb.Palette[1] = pals[1]
		c.config.Gb.Palette[2] = pals[2]
		c.config.Gb.Palette[3] = pals[3]
	}
}

func (c *Config) decodeGba() {
	c.decodeKeyboard(&c.Gba.Keyboard, &c.config.Gba.KeyboardConfig)
	c.decodeController(&c.Gba.Controller, &c.config.Gba.ControllerConfig)

	c.config.Gba.IdleOptimize = c.Gba.IdleOptimize
	c.config.Gba.SoundClockUpdateCycles = c.Gba.SoundClockUpdateCycles
}

func (c *Config) decodeNds() {
	c.decodeKeyboard(&c.Nds.Keyboard, &c.config.Nds.KeyboardConfig)
	c.decodeController(&c.Nds.Controller, &c.config.Nds.ControllerConfig)

	if utils.IsFile(c.Nds.Bios.Arm7Path) {
		c.config.Nds.Bios.Arm7Path = c.Nds.Bios.Arm7Path
	}

	if utils.IsFile(c.Nds.Bios.Arm9Path) {
		c.config.Nds.Bios.Arm9Path = c.Nds.Bios.Arm9Path
	}

	c.config.Nds.Rtc.AdditionalHours = c.Nds.Rtc.AdditionalHours

	// if utils.IsDirectory(c.Nds.Export.Directory) {
	// need to create directory if empty
	c.config.Nds.Export.Directory = c.Nds.Export.Directory
	//}

	c.config.Nds.Export.ShadowPolys = c.Nds.Export.ShadowPolys

	switch strings.ToLower(c.Nds.Screen.Layout) {
	case "horizontal":
		c.config.Nds.Screen.Layout = 1
	case "hybrid":
		c.config.Nds.Screen.Layout = 2
	default:
		c.config.Nds.Screen.Layout = 0
	}

	switch strings.ToLower(c.Nds.Screen.Sizing) {
	case "only top":
		c.config.Nds.Screen.Sizing = 1
	case "only bottom":
		c.config.Nds.Screen.Sizing = 2
	default:
		c.config.Nds.Screen.Sizing = 0
	}

	switch c.Nds.Screen.Rotation {
	case 90:
		c.config.Nds.Screen.Rotation = 1
	case 180:
		c.config.Nds.Screen.Rotation = 2
	case 270:
		c.config.Nds.Screen.Rotation = 3
	default:
		c.config.Nds.Screen.Rotation = 0
	}

	c.decodeNdsFirmware()
	c.decodeNdsJit()
}

func (c *Config) decodeNdsFirmware() {
	f := &c.Nds.Firmware
	conf := &c.config.Nds.Firmware

	if utils.IsFile(f.FilePath) {
		conf.FilePath = f.FilePath
	}

	clr, ok := config.ColorNameToId[strings.ToLower(f.FavoriteColor)]
	if !ok {
		clr = config.FW_CLR_GRAY
	}

	conf.Color = clr

	switch {
	case len(f.Nickname) >= 10:
		conf.Nickname = f.Nickname[:10]
	case len(f.Nickname) == 0:
		f.Nickname = "guac"
	default:
		conf.Nickname = f.Nickname
	}

	switch {
	case len(f.Message) >= 26:
		conf.Message = f.Message[:26]
	case len(f.Message) == 0:
		f.Message = "Guac by Aaron Balke!"
	default:
		conf.Message = f.Message
	}

	if len(f.Message) >= 26 {
		conf.Message = f.Message[:26]
	} else {
		conf.Message = f.Message
	}

	conf.BirthdayMonth = min(32, f.BirthdayMonth)
	if f.BirthdayMonth == 0 {
		f.BirthdayMonth = 8
	}

	conf.BirthdayDay = min(13, f.BirthdayDay)
	if f.BirthdayDay == 0 {
		f.BirthdayDay = 8
	}
}

func (c *Config) decodeNdsJit() {
	isX86 := runtime.GOARCH == "amd64" || runtime.GOARCH == "386"

	c.config.Nds.Jit.Enabled = c.Nds.Jit.Enabled && isX86

	if !c.config.Nds.Jit.Enabled {
		c.Nds.Jit.BatchInst = 1
	}

	if c.Nds.Jit.LoopCnt != 0 {
		c.config.Nds.Jit.LoopCnt = c.Nds.Jit.LoopCnt
	} else {
		c.config.Nds.Jit.LoopCnt = 255
	}

	if c.Nds.Jit.BlockCnt != 0 {
		c.config.Nds.Jit.BlockCnt = c.Nds.Jit.BlockCnt
	} else {
		c.config.Nds.Jit.BlockCnt = 0x1000
	}

	c.config.Nds.Jit.BatchInstA9 = max(c.Nds.Jit.BatchInst, 2)
	c.config.Nds.Jit.BatchInstA7 = max(c.Nds.Jit.BatchInst/2, 1)
}

func (c *Config) decodeKeyboard(in *EmulatorInput, conf *config.EmulatorKeyboard) {
	tomls := []*[]string{
		&in.A,
		&in.B,
		&in.Select,
		&in.Start,
		&in.Left,
		&in.Right,
		&in.Up,
		&in.Down,
		&in.R,
		&in.L,
		&in.X,
		&in.Y,
		&in.Hinge,
		&in.Debug,
		&in.LayoutToggle,
		&in.SizingToggle,
		&in.RotationToggle,
		&in.ExportScene,
	}

	outputs := []*[]ebiten.Key{
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

	for i := range len(tomls) {
		for j := range *tomls[i] {
			if str, ok := utils.StringToKey((*tomls[i])[j]); ok {
				*outputs[i] = append(*outputs[i], str)
			}
		}
	}
}

func (c *Config) decodeController(in *EmulatorInput, conf *config.EmulatorController) {
	tomls := []*[]string{
		&in.A,
		&in.B,
		&in.Select,
		&in.Start,
		&in.Left,
		&in.Right,
		&in.Up,
		&in.Down,
		&in.R,
		&in.L,
		&in.X,
		&in.Y,
		&in.Hinge,
	}

	outputs := []*[]ebiten.StandardGamepadButton{
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
	}

	for i := range len(tomls) {
		for j := range len(*tomls[i]) {
			if str, ok := utils.StringToGamepadButton((*tomls[i])[j]); ok {
				*outputs[i] = append(*outputs[i], str)
			}
		}
	}
}
