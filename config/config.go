package config

import (
	"fmt"
	"image/color"
	"log"
	"os"

	_ "embed"

	"github.com/BurntSushi/toml"
	"github.com/aabalke/guac/utils"
	"github.com/hajimehoshi/ebiten/v2"
	sys "golang.org/x/sys/cpu"
)

//go:embed default.toml
var defaultConfig []byte

const CONFIG_PATH = "./config.toml"

var Conf Config

//type Profiler struct {
//	Enabled    bool
//	StartFrame int
//	EndFrame   int
//}
//

type Config struct {
	Ui              Ui          `toml:"ui"`
	General         General     `toml:"general"`
	CancelAudioInit bool        `toml:"cancel_audio_init"`
	Mouse           MouseConfig `toml:"mouse"`
	Jit             Jit         `toml:"jit"`
	Gb              GbConfig    `toml:"gb"`
	Gba             GbaConfig   `toml:"gba"`
	Nds             NdsConfig   `toml:"nds"`
}

type Ui struct {
	//GamesPerRow      int  `toml:"games_per_row"`
	TomlBackdrop        int `toml:"backdrop_color"`
	TomlBackground      int `toml:"menu_background_color"`
	TomlForeground      int `toml:"menu_foreground_color"`
	TomlSecondary       int `toml:"menu_secondary_color"`
	Backdrop            color.Color
	MenuBackgroundColor color.Color
	MenuForegroundColor color.Color
	MenuSecondaryColor  color.Color
	//MenuFontFace        text.Face

	Language string `toml:"language"`
}

type General struct {
	Muted          bool `toml:"muted"`
	TargetFps      int  `toml:"target_fps"`
	ShowFps        bool `toml:"show_fps"`
	VsyncDisabled  bool `toml:"vsync_disabled"`
	InitFullscreen bool `toml:"fullscreen"`

	KeyboardConfig   KeyboardConfig   `toml:"keyboard"`
	ControllerConfig ControllerConfig `toml:"controller"`
}

type MouseConfig struct {
	Fill            bool    `toml:"fill"`
	Stroke          bool    `toml:"stroke"`
	UnSelectedAlpha float32 `toml:"unselected_alpha"`
	CursorSize      int     `toml:"cursor_diameter"`
	StrokeSize      int     `toml:"stroke_width"`
	TomlFillColor   int     `toml:"fill_color"`
	TomlStrokeColor int     `toml:"stroke_color"`
	FillColor       []uint8
	StrokeColor     []uint8
}

type GbConfig struct {
	TomlPalette      []int                    `toml:"dmg_palette"`
	KeyboardConfig   EmulatorKeyboardConfig   `toml:"keyboard"`
	ControllerConfig EmulatorControllerConfig `toml:"controller"`
	ConsoleType      string                   `toml:"type"`

	//Palette  [][]uint8
	Palette  [4]color.Color
	ForceDMG bool
	ForceGBC bool
}

type GbaConfig struct {
	KeyboardConfig   EmulatorKeyboardConfig   `toml:"keyboard"`
	ControllerConfig EmulatorControllerConfig `toml:"controller"`

	SkipHle                bool `toml:"skip_hle"`
	Threads                int  `toml:"threads"`
	IdleOptimize           bool `toml:"idle_optimize"`
	SoundClockUpdateCycles int  `toml:"sound_clock_update_cycles"`
	DisableSaves           bool `toml:"disable_saves"`
}

type KeyboardConfig struct {
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

type ControllerConfig struct {
	Select     []ebiten.StandardGamepadButton
	Return     []ebiten.StandardGamepadButton
	Mute       []ebiten.StandardGamepadButton
	Pause      []ebiten.StandardGamepadButton
	Left       []ebiten.StandardGamepadButton
	Right      []ebiten.StandardGamepadButton
	Up         []ebiten.StandardGamepadButton
	Down       []ebiten.StandardGamepadButton
	Fullscreen []ebiten.StandardGamepadButton
	Quit       []ebiten.StandardGamepadButton

	TomlSelect     []string `toml:"select"`
	TomlReturn     []string `toml:"return"`
	TomlMute       []string `toml:"mute"`
	TomlPause      []string `toml:"pause"`
	TomlLeft       []string `toml:"left"`
	TomlRight      []string `toml:"right"`
	TomlUp         []string `toml:"up"`
	TomlDown       []string `toml:"down"`
	TomlFullscreen []string `toml:"fullscreen"`
	TomlQuit       []string `toml:"quit"`
}

type EmulatorKeyboardConfig struct {
	A      []string `toml:"a"`
	B      []string `toml:"b"`
	Select []string `toml:"select"`
	Start  []string `toml:"start"`
	Left   []string `toml:"left"`
	Right  []string `toml:"right"`
	Up     []string `toml:"up"`
	Down   []string `toml:"down"`
	R      []string `toml:"r"`
	L      []string `toml:"l"`
	X      []string `toml:"x"`
	Y      []string `toml:"y"`
	Hinge  []string `toml:"hinge"`
	Debug  []string `toml:"Debug"`

	LayoutToggle   []string `toml:"layout_toggle"`
	SizingToggle   []string `toml:"sizing_toggle"`
	RotationToggle []string `toml:"rotation_toggle"`
	ExportScene    []string `toml:"export_scene"`
}

type EmulatorControllerConfig struct {
	TomlA      []string `toml:"a"`
	TomlB      []string `toml:"b"`
	TomlSelect []string `toml:"select"`
	TomlStart  []string `toml:"start"`
	TomlLeft   []string `toml:"left"`
	TomlRight  []string `toml:"right"`
	TomlUp     []string `toml:"up"`
	TomlDown   []string `toml:"down"`
	TomlR      []string `toml:"r"`
	TomlL      []string `toml:"l"`
	TomlX      []string `toml:"x"`
	TomlY      []string `toml:"y"`
	TomlHinge  []string `toml:"hinge"`

	A      []ebiten.StandardGamepadButton
	B      []ebiten.StandardGamepadButton
	Select []ebiten.StandardGamepadButton
	Start  []ebiten.StandardGamepadButton
	Left   []ebiten.StandardGamepadButton
	Right  []ebiten.StandardGamepadButton
	Up     []ebiten.StandardGamepadButton
	Down   []ebiten.StandardGamepadButton
	R      []ebiten.StandardGamepadButton
	L      []ebiten.StandardGamepadButton
	X      []ebiten.StandardGamepadButton
	Y      []ebiten.StandardGamepadButton
	Hinge  []ebiten.StandardGamepadButton
}

func (c *Config) Decode() {

	b, err := os.ReadFile(CONFIG_PATH)
	if err != nil {
		if os.IsNotExist(err) {

			f, err2 := os.Create(CONFIG_PATH)
			if err2 != nil {
				panic(err2)
			}

			_, err2 = f.Write(defaultConfig)
			if err2 != nil {
				panic(err2)
			}

			b = defaultConfig

		} else {
			panic(err)
		}
	}

	_, err = toml.Decode(string(b), &c)
	if err != nil {
		panic(err)
	}

	c.Ui.Backdrop = color.RGBA{
		R: uint8(c.Ui.TomlBackdrop >> 16),
		G: uint8(c.Ui.TomlBackdrop >> 8),
		B: uint8(c.Ui.TomlBackdrop),
		A: 0xFF,
	}

	c.Ui.MenuBackgroundColor = color.RGBA{
		R: uint8(c.Ui.TomlBackground >> 16),
		G: uint8(c.Ui.TomlBackground >> 8),
		B: uint8(c.Ui.TomlBackground),
		A: 0xFF,
	}
	c.Ui.MenuForegroundColor = color.RGBA{
		R: uint8(c.Ui.TomlForeground >> 16),
		G: uint8(c.Ui.TomlForeground >> 8),
		B: uint8(c.Ui.TomlForeground),
		A: 0xFF,
	}
	c.Ui.MenuSecondaryColor = color.RGBA{
		R: uint8(c.Ui.TomlSecondary >> 16),
		G: uint8(c.Ui.TomlSecondary >> 8),
		B: uint8(c.Ui.TomlSecondary),
		A: 0xFF,
	}

	//if c.Ui.GamesPerRow == 0 {
	//	errMessageStart := "Invalid Config:"
	//	errMessageEnd := "Using 6 games per row in menu."
	//	log.Printf("%s %s %s\n", errMessageStart, "GamesPerRow == 0.", errMessageEnd)
	//	c.Ui.GamesPerRow = 6
	//}

	c.DecodeGeneralController()

	c.decodeJit()

	c.decodeGb()

	DecodeController(&c.Gba.ControllerConfig)

	c.decodeNds()

	c.decodeMouse()
}

func (c *Config) DecodeGeneralController() {

	conf := &c.General.ControllerConfig

	configs := []*[]string{
		&conf.TomlSelect,
		&conf.TomlReturn,
		&conf.TomlMute,
		&conf.TomlPause,
		&conf.TomlLeft,
		&conf.TomlRight,
		&conf.TomlUp,
		&conf.TomlDown,
		&conf.TomlFullscreen,
		&conf.TomlQuit,
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

	for i := range len(configs) {
		for j := range len(*configs[i]) {
			*outputs[i] = append(*outputs[i], utils.StringToGamepadButton((*configs[i])[j]))
		}
	}
}

func DecodeController(conf *EmulatorControllerConfig) {

	configs := []*[]string{
		&conf.TomlA,
		&conf.TomlB,
		&conf.TomlSelect,
		&conf.TomlStart,
		&conf.TomlLeft,
		&conf.TomlRight,
		&conf.TomlUp,
		&conf.TomlDown,
		&conf.TomlR,
		&conf.TomlL,
		&conf.TomlX,
		&conf.TomlY,
		&conf.TomlHinge,
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

	if len(outputs) != len(configs) {
		panic("decode controller has different lens")
	}

	for i := range len(configs) {
		for j := range len(*configs[i]) {
			*outputs[i] = append(*outputs[i], utils.StringToGamepadButton((*configs[i])[j]))
		}
	}
}

func (c *Config) decodeGb() {

	DecodeController(&c.Gb.ControllerConfig)

	pal := c.Gb.TomlPalette

	invalid := false

	switch c.Gb.ConsoleType {
	case "dmg":
		c.Gb.ForceDMG = true
	case "gbc":
		c.Gb.ForceGBC = true
	}

	errMessageStart := "Invalid Config:"
	errMessageEnd := "Using default palette."

	switch len(pal) {
	case 0:
		log.Printf("%s %s %s\n", errMessageStart, "gb palette not provided.", errMessageEnd)
		invalid = true
	case 4:

		for i := range 4 {
			if pal[i] < 0 || pal[i] > 0xFFFFFF {
				s := fmt.Sprintf("gb palette value idx %d has invalid 8 bit value.", i)

				log.Printf("%s %s %s\n", errMessageStart, s, errMessageEnd)
				invalid = true
			}
		}
	default:
		log.Printf("%s %s %s\n", errMessageStart, "gb palette len != 4.", errMessageEnd)
		invalid = true

	}

	if invalid {
		// greyscale palette
		c.Gb.Palette = [4]color.Color{
			color.RGBA{0xFF, 0xFF, 0xFF, 0xFF},
			color.RGBA{0xCC, 0xCC, 0xCC, 0xFF},
			color.RGBA{0x77, 0x77, 0x77, 0xFF},
			color.RGBA{0x00, 0x00, 0x00, 0xFF},
		}
	} else {

		c.Gb.Palette = [4]color.Color{
			color.RGBA{uint8(pal[0] >> 16), uint8(pal[0] >> 8), uint8(pal[0]), 0xFF},
			color.RGBA{uint8(pal[1] >> 16), uint8(pal[1] >> 8), uint8(pal[1]), 0xFF},
			color.RGBA{uint8(pal[2] >> 16), uint8(pal[2] >> 8), uint8(pal[2]), 0xFF},
			color.RGBA{uint8(pal[3] >> 16), uint8(pal[3] >> 8), uint8(pal[3]), 0xFF},
		}
	}
}

func (c *Config) decodeMouse() {

	pal := c.Mouse.TomlFillColor

	invalid := false

	errMessageStart := "Invalid Mouse Config:"
	errMessageEnd := "Using default fill color."

	if pal < 0 || pal > 0xFFFFFF {
		s := fmt.Sprintf("mouse fill palette value has invalid 8 bit value.")

		log.Printf("%s %s %s\n", errMessageStart, s, errMessageEnd)
		invalid = true
	}

	if invalid {
		// greyscale palette
		c.Mouse.FillColor = []uint8{
			0xFF, 0xFF, 0xFF,
		}
	} else {
		c.Mouse.FillColor = []uint8{
			uint8(pal >> 16), uint8(pal >> 8), uint8(pal),
		}
	}

	pal = c.Mouse.TomlStrokeColor

	invalid = false

	errMessageStart = "Invalid Mouse Config:"
	errMessageEnd = "Using default stroke color."

	if pal < 0 || pal > 0xFFFFFF {
		s := fmt.Sprintf("mouse stroke palette value has invalid 8 bit value.")

		log.Printf("%s %s %s\n", errMessageStart, s, errMessageEnd)
		invalid = true
	}

	if invalid {
		// greyscale palette
		c.Mouse.StrokeColor = []uint8{
			0xFF, 0xFF, 0xFF,
		}
	} else {
		c.Mouse.StrokeColor = []uint8{
			uint8(pal >> 16), uint8(pal >> 8), uint8(pal),
		}
	}
}

type Jit struct {
	Enabled   bool   `toml:"enabled"`
	BatchInst uint32 `toml:"batch_inst"`

	LoopCnt   uint32 `toml:"loop_cnt"`
	BlockCnt  uint32 `toml:"block_cnt"`
	PageShift uint32 `toml:"page_shift"`

	BatchInstA9 uint32
	BatchInstA7 uint32
}

func (c *Config) decodeJit() {

	if Conf.Jit.Enabled && !sys.X86.HasSSE2 {

		errMessageStart := "Invalid Config:"
		errMessageEnd := "Disabling Jit Compiler."
		log.Printf("%s %s %s\n", errMessageStart, "native machine not x86 instruction set.", errMessageEnd)
		//fmt.Printf("Warning)
		Conf.Jit.Enabled = false
	}

	if !Conf.Jit.Enabled {
		Conf.Jit.BatchInst = 1

	}

	Conf.Jit.BatchInstA9 = max(Conf.Jit.BatchInst, 2)
	Conf.Jit.BatchInstA7 = max(Conf.Jit.BatchInst/2, 1)

	//if Conf.Nds.NdsJit.PageCount == 0 {
	//	errMessageStart := "Invalid Config:"
	//	errMessageEnd := "Setting Jit Compiler. Page count to 1024."
	//	log.Printf("%s %s\n", errMessageStart, errMessageEnd)
	//	//fmt.Printf("Warning)
	//	Conf.Nds.NdsJit.PageCount = 0x1024_0000
	//}
}
