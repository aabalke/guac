package file

import (
	_ "embed"
	"image/color"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/utils"
	"github.com/hajimehoshi/ebiten/v2"
	sys "golang.org/x/sys/cpu"
)

//go:embed default.toml
var DEF_CONFIG []byte

const CONFIG_PATH = "./config.toml"

type Config struct {
	config  *config.Config
	General General `toml:"general"`
	Ui      Ui      `toml:"ui"`
	Gb      Gb      `toml:"gb"`
	Gba     Gba     `toml:"gba"`
	Nds     Nds     `toml:"nds"`
}

type General struct {
	Muted          bool         `toml:"muted"`
	TargetFps      int          `toml:"target_fps"`
	ShowFps        bool         `toml:"show_fps"`
	InitFullscreen bool         `toml:"fullscreen"`
	Keyboard       GeneralInput `toml:"keyboard"`
	Controller     GeneralInput `toml:"controller"`
}

type GeneralInput struct {
	Select     []string `toml:"select"`
	Return     []string `toml:"return"`
	Mute       []string `toml:"mute"`
	Pause      []string `toml:"pause"`
	Left       []string `toml:"left"`
	Right      []string `toml:"right"`
	Up         []string `toml:"up"`
	Down       []string `toml:"down"`
	Fullscreen []string `toml:"fullscreen"`
	Quit       []string `toml:"quit"`
}

type Ui struct {
	Backdrop   int    `toml:"backdrop_color"`
	Background int    `toml:"menu_background_color"`
	Foreground int    `toml:"menu_foreground_color"`
	Secondary  int    `toml:"menu_secondary_color"`
	Language   string `toml:"language"`
}

type Gb struct {
	Palette    []int         `toml:"dmg_palette"`
	Keyboard   EmulatorInput `toml:"keyboard"`
	Controller EmulatorInput `toml:"controller"`
}

type Gba struct {
	IdleOptimize           bool          `toml:"idle_optimize"`
	SoundClockUpdateCycles int           `toml:"sound_clock_update_cycles"`
	Keyboard               EmulatorInput `toml:"keyboard"`
	Controller             EmulatorInput `toml:"controller"`
}

type Nds struct {
	Keyboard   EmulatorInput `toml:"keyboard"`
	Controller EmulatorInput `toml:"controller"`
	Bios       NdsBios       `toml:"bios"`
	Rtc        NdsRtc        `toml:"rtc"`
	Export     NdsExport     `toml:"export"`
	Screen     NdsScreen     `toml:"screen"`
	Firmware   NdsFirmware   `toml:"firmware"`
	Jit        NdsJit        `toml:"jit"`
}

type NdsBios struct {
	Arm7Path string `toml:"arm7_path"`
	Arm9Path string `toml:"arm9_path"`
}

type NdsRtc struct {
	AdditionalHours int `toml:"additional_hours"`
}

type NdsExport struct {
	Directory   string `toml:"directory"`
	ShadowPolys bool   `toml:"shadow_polygons"`
}

type NdsScreen struct {
	Layout   string `toml:"layout"`
	Sizing   string `toml:"sizing"`
	Rotation int    `toml:"rotation"`
}

type NdsFirmware struct {
	FilePath      string `toml:"file_path"`
	Nickname      string `toml:"nickname"`
	Message       string `toml:"message"`
	FavoriteColor string `toml:"favorite_color"`
	BirthdayMonth uint8  `toml:"birthday_month"`
	BirthdayDay   uint8  `toml:"birthday_day"`
}

type NdsJit struct {
	Enabled   bool   `toml:"enabled"`
	BatchInst uint32 `toml:"batch_inst"`
	LoopCnt   uint32 `toml:"loop_cnt"`
	BlockCnt  uint32 `toml:"block_cnt"`
}

type EmulatorInput struct {
	A              []string `toml:"a"`
	B              []string `toml:"b"`
	Select         []string `toml:"select"`
	Start          []string `toml:"start"`
	Left           []string `toml:"left"`
	Right          []string `toml:"right"`
	Up             []string `toml:"up"`
	Down           []string `toml:"down"`
	R              []string `toml:"r"`
	L              []string `toml:"l"`
	X              []string `toml:"x"`
	Y              []string `toml:"y"`
	Hinge          []string `toml:"hinge"`
	Debug          []string `toml:"Debug"`
	LayoutToggle   []string `toml:"layout_toggle"`
	SizingToggle   []string `toml:"sizing_toggle"`
	RotationToggle []string `toml:"rotation_toggle"`
	ExportScene    []string `toml:"export_scene"`
}

func Decode() {
	c := &Config{config: &config.Conf}
	c.open()
	c.decodeGeneral()
	c.decodeUi()
	c.decodeGb()
	c.decodeGba()
	c.decodeNds()
}

func (c *Config) open() {

	b, err := os.ReadFile(CONFIG_PATH)
	if err != nil {
		if os.IsNotExist(err) {

			f, err2 := os.Create(CONFIG_PATH)
			if err2 != nil {
				panic(err2)
			}

			_, err2 = f.Write(DEF_CONFIG)
			if err2 != nil {
				panic(err2)
			}

			b = DEF_CONFIG

		} else {
			panic(err)
		}
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

	outputsKeys := []*[]string{
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
		*outputsKeys[i] = *tomls[i]
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
			*outputs[i] = append(*outputs[i], utils.StringToGamepadButton((*tomls[i])[j]))
		}
	}
}

func (c *Config) decodeUi() {

	c.config.Ui.Backdrop = color.RGBA{
		R: uint8(c.Ui.Backdrop >> 16),
		G: uint8(c.Ui.Backdrop >> 8),
		B: uint8(c.Ui.Backdrop),
		A: 0xFF,
	}

	c.config.Ui.MenuBackgroundColor = color.RGBA{
		R: uint8(c.Ui.Background >> 16),
		G: uint8(c.Ui.Background >> 8),
		B: uint8(c.Ui.Background),
		A: 0xFF,
	}
	c.config.Ui.MenuForegroundColor = color.RGBA{
		R: uint8(c.Ui.Foreground >> 16),
		G: uint8(c.Ui.Foreground >> 8),
		B: uint8(c.Ui.Foreground),
		A: 0xFF,
	}
	c.config.Ui.MenuSecondaryColor = color.RGBA{
		R: uint8(c.Ui.Secondary >> 16),
		G: uint8(c.Ui.Secondary >> 8),
		B: uint8(c.Ui.Secondary),
		A: 0xFF,
	}

	switch c.Ui.Language {
	case "spanish", "es": // english
		c.config.Ui.Language = 1
	default:
		c.config.Ui.Language = 0
	}
}

func (c *Config) decodeGb() {

	c.decodeKeyboard(&c.Gb.Keyboard, &c.config.Gb.KeyboardConfig)
	c.decodeController(&c.Gb.Controller, &c.config.Gb.ControllerConfig)

	pal := c.Gb.Palette

	invalid := len(pal) != 4

	for _, v := range c.Gb.Palette {
		if v < 0 || v > 0xFFFFFF {
			invalid = false
		}
	}

	if invalid {
		// greyscale palette
		c.config.Gb.Palette = [4]color.Color{
			color.RGBA{0xFF, 0xFF, 0xFF, 0xFF},
			color.RGBA{0xCC, 0xCC, 0xCC, 0xFF},
			color.RGBA{0x77, 0x77, 0x77, 0xFF},
			color.RGBA{0x00, 0x00, 0x00, 0xFF},
		}
	} else {

		c.config.Gb.Palette = [4]color.Color{
			color.RGBA{uint8(pal[0] >> 16), uint8(pal[0] >> 8), uint8(pal[0]), 0xFF},
			color.RGBA{uint8(pal[1] >> 16), uint8(pal[1] >> 8), uint8(pal[1]), 0xFF},
			color.RGBA{uint8(pal[2] >> 16), uint8(pal[2] >> 8), uint8(pal[2]), 0xFF},
			color.RGBA{uint8(pal[3] >> 16), uint8(pal[3] >> 8), uint8(pal[3]), 0xFF},
		}
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

	if utils.IsDirectory(c.Nds.Export.Directory) {
		// need to create directory if empty
		c.config.Nds.Export.Directory = c.Nds.Export.Directory
	}

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

	c.config.Nds.Jit.Enabled = c.Nds.Jit.Enabled && sys.X86.HasSSE2

	if c.config.Nds.Jit.Enabled {
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

	outputs := []*[]string{
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
		*outputs[i] = *tomls[i]
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
			*outputs[i] = append(*outputs[i], utils.StringToGamepadButton((*tomls[i])[j]))
		}
	}
}
