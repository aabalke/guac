package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	FW_CLR_GRAY uint8 = iota
	FW_CLR_BROWN
	FW_CLR_RED
	FW_CLR_PINK
	FW_CLR_ORANGE
	FW_CLR_YELLOW
	FW_CLR_LIME_GREEN
	FW_CLR_GREEN
	FW_CLR_DARK_GREEN
	FW_CLR_SEA_GREEN
	FW_CLR_TURQUOISE
	FW_CLR_BLUE
	FW_CLR_DARK_BLUE
	FW_CLR_DARK_PURPLE
	FW_CLR_VIOLET
	FW_CLR_MAGENTA
)

var clr = map[string]uint8{
	"Gray":        FW_CLR_GRAY,
	"Brown":       FW_CLR_BROWN,
	"Red":         FW_CLR_RED,
	"Pink":        FW_CLR_PINK,
	"Orange":      FW_CLR_ORANGE,
	"Yellow":      FW_CLR_YELLOW,
	"Lime Green":  FW_CLR_LIME_GREEN,
	"Green":       FW_CLR_GREEN,
	"Dark Green":  FW_CLR_DARK_GREEN,
	"Sea Green":   FW_CLR_SEA_GREEN,
	"Turquoise":   FW_CLR_TURQUOISE,
	"Blue":        FW_CLR_BLUE,
	"Dark Blue":   FW_CLR_DARK_BLUE,
	"Dark Purple": FW_CLR_DARK_PURPLE,
	"Violet":      FW_CLR_VIOLET,
	"Magenta":     FW_CLR_MAGENTA,
}

type NdsConfig struct {
	Jit              Jit                      `toml:"jit"`
	Screen           NdsScreen                `toml:"screen"`
	KeyboardConfig   EmulatorKeyboardConfig   `toml:"keyboard"`
	ControllerConfig EmulatorControllerConfig `toml:"controller"`
	Firmware         NdsFirmware              `toml:"firmware"`
	Rtc              NdsRtc                   `toml:"rtc"`
	Export           NdsExport                `toml:"export"`
	Threads          int                      `toml:"threads"`
	DisableSaves     bool                     `toml:"disable_saves"`
	FrameSkip        uint32                   `toml:"frame_skip"`
	DynamicFrameSkip bool                     `toml:"dynamic_frame_skip"`
	Bios             NdsBios                  `toml:"bios"`
}

func (c *Config) decodeNds() {
	DecodeController(&c.Nds.ControllerConfig)
	c.decodeNdsScreen()
	c.decodeNdsBios()
	c.decodeNdsFirmware()
	c.decodeNdsExport()
}

type NdsBios struct {
	Arm7Path string `toml:"arm7_path"`
	Arm9Path string `toml:"arm9_path"`
}

func (c *Config) decodeNdsBios() {

	if !isFile(c.Nds.Bios.Arm7Path) {
		fmt.Printf("Nds Arm7 Bios not provided, using built-in\n")
		c.Nds.Bios.Arm7Path = ""
	} else {
		fmt.Printf("Nds Arm7 Bios provided.\n")
	}

	if !isFile(c.Nds.Bios.Arm9Path) {
		fmt.Printf("Nds Arm9 Bios not provided, using built-in\n")
		c.Nds.Bios.Arm9Path = ""
	} else {
		fmt.Printf("Nds Arm9 Bios provided.\n")
	}
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return !info.IsDir()
}

type NdsFirmware struct {
	FilePath      string `toml:"file_path"`
	Nickname      string `toml:"nickname"`
	Message       string `toml:"message"`
	FavoriteColor string `toml:"favorite_color"`
	BirthdayMonth uint8  `toml:"birthday_month"`
	BirthdayDay   uint8  `toml:"birthday_month"`
	Color         uint8
}

func (c *Config) decodeNdsFirmware() {

	f := &c.Nds.Firmware

	if !isFile(f.FilePath) {
		fmt.Printf("Nds Firmware not provided.\n")
	} else {
		fmt.Printf("Nds Firmware provided.\n")
	}

	clr, ok := clr[f.FavoriteColor]
	if !ok {
		clr = 0
	}

	f.Color = clr

	if len(f.Nickname) >= 10 {
		panic("Nds Firmware config setting Nickname is too long. Must be < 10 characters")
	}

	if len(f.Message) >= 26 {
		panic("Nds Firmware config setting Message is too long. Must be < 26 characters")
	}

	if f.BirthdayDay >= 32 {
		panic("Nds Firmware config setting BirthdayDay is too long. Must be < 26 characters")
	}

	if f.BirthdayMonth >= 13 {
		panic("Nds Firmware config setting BirthdayMonth is too long. Must be < 26 characters")
	}

	// 8/8/2025 is default

	if f.BirthdayDay == 0 {
		f.BirthdayDay = 8
	}

	if f.BirthdayMonth == 0 {
		f.BirthdayMonth = 8
	}

	if f.Nickname == "" {
		f.Nickname = "guac"
	}

	if f.Message == "" {
		f.Message = "Guac emulator by Aaron Balke!"
	}
}

type NdsScreen struct {
	ConfigLayout   string `toml:"layout"`
	ConfigSizing   string `toml:"sizing"`
	ConfigRotation int    `toml:"rotation"`

	OLayout   int
	OSizing   int
	ORotation int
}

func (c *Config) decodeNdsScreen() {

	switch c.Nds.Screen.ConfigLayout {
	case "horizontal":
		c.Nds.Screen.OLayout = 1
	case "hybrid":
		c.Nds.Screen.OLayout = 2
	default:
		c.Nds.Screen.OLayout = 0
	}

	switch c.Nds.Screen.ConfigSizing {
	case "only top":
		c.Nds.Screen.OSizing = 1
	case "only bottom":
		c.Nds.Screen.OSizing = 2
	default:
		c.Nds.Screen.OSizing = 0
	}

	switch c.Nds.Screen.ConfigRotation {
	case 90:
		c.Nds.Screen.ORotation = 1
	case 180:
		c.Nds.Screen.ORotation = 2
	case 270:
		c.Nds.Screen.ORotation = 3
	default:
		c.Nds.Screen.ORotation = 0
	}
}

type NdsRtc struct {
	AdditionalHours int `toml:"additional_hours"`
}

const (
	FORMAT_OBJ = iota
	FORMAT_GLTF
)

type NdsExport struct {
	Directory   string `toml:"directory"`
	FileType    string `toml:"file_type"`
	ShadowPolys bool   `toml:"shadow_polygons"`
	Format      int
}

func (c *Config) decodeNdsExport() {
	switch {
	case strings.Contains(c.Nds.Export.FileType, "glft"):
		c.Nds.Export.Format = FORMAT_GLTF
	default:
		c.Nds.Export.Format = FORMAT_OBJ
	}
}
