package config

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

var Conf Config

type FirmwareColor = int

const (
	FW_CLR_GRAY FirmwareColor = iota
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

var ColorNames = []string{
	"gray",
	"brown",
	"red",
	"pink",
	"orange",
	"yellow",
	"lime green",
	"green",
	"dark green",
	"sea green",
	"turquoise",
	"blue",
	"dark blue",
	"dark purple",
	"violet",
	"magenta",
}

var ColorNameToId = map[string]FirmwareColor{
	ColorNames[0x0]: FW_CLR_GRAY,
	ColorNames[0x1]: FW_CLR_BROWN,
	ColorNames[0x2]: FW_CLR_RED,
	ColorNames[0x3]: FW_CLR_PINK,
	ColorNames[0x4]: FW_CLR_ORANGE,
	ColorNames[0x5]: FW_CLR_YELLOW,
	ColorNames[0x6]: FW_CLR_LIME_GREEN,
	ColorNames[0x7]: FW_CLR_GREEN,
	ColorNames[0x8]: FW_CLR_DARK_GREEN,
	ColorNames[0x9]: FW_CLR_SEA_GREEN,
	ColorNames[0xA]: FW_CLR_TURQUOISE,
	ColorNames[0xB]: FW_CLR_BLUE,
	ColorNames[0xC]: FW_CLR_DARK_BLUE,
	ColorNames[0xD]: FW_CLR_DARK_PURPLE,
	ColorNames[0xE]: FW_CLR_VIOLET,
	ColorNames[0xF]: FW_CLR_MAGENTA,
}

type Config struct {
	General General
	Ui      Ui
	Gb      Gb
	Gba     Gba
	Nds     NdsConfig
}

type General struct {
	Muted          bool
	TargetFps      int
	ShowFps        bool
	InitFullscreen bool
	Keyboard       GeneralKeyboard
	Controller     GeneralController
}

type GeneralKeyboard struct {
	Select     []string
	Return     []string
	Mute       []string
	Pause      []string
	Left       []string
	Right      []string
	Up         []string
	Down       []string
	Fullscreen []string
	Quit       []string
}

type GeneralController struct {
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
}

type Ui struct {
	Backdrop            color.Color
	MenuBackgroundColor color.Color
	MenuForegroundColor color.Color
	MenuSecondaryColor  color.Color
	Language            int
}

type Gb struct {
	Palette          [4]color.Color
	KeyboardConfig   EmulatorKeyboard
	ControllerConfig EmulatorController
}

type Gba struct {
	IdleOptimize           bool
	SoundClockUpdateCycles int
	KeyboardConfig         EmulatorKeyboard
	ControllerConfig       EmulatorController
}

type NdsConfig struct {
	Screen           NdsScreen
	Firmware         NdsFirmware
	Rtc              NdsRtc
	Export           NdsExport
	Bios             NdsBios
	Jit              NdsJit
	KeyboardConfig   EmulatorKeyboard
	ControllerConfig EmulatorController
}

type NdsBios struct {
	Arm7Path string
	Arm9Path string
}

type NdsRtc struct {
	AdditionalHours int
}

type NdsExport struct {
	Directory   string
	ShadowPolys bool
}

type NdsScreen struct {
	Layout   int
	Sizing   int
	Rotation int
}

type NdsFirmware struct {
	FilePath      string
	Nickname      string
	Message       string
	BirthdayMonth uint8
	BirthdayDay   uint8
	Color         FirmwareColor
}

type NdsJit struct {
	Enabled     bool
	LoopCnt     uint32
	BlockCnt    uint32
	BatchInstA9 uint32
	BatchInstA7 uint32
}

type EmulatorKeyboard struct {
	A              []string
	B              []string
	Select         []string
	Start          []string
	Left           []string
	Right          []string
	Up             []string
	Down           []string
	R              []string
	L              []string
	X              []string
	Y              []string
	Hinge          []string
	Debug          []string
	LayoutToggle   []string
	SizingToggle   []string
	RotationToggle []string
	ExportScene    []string
}

type EmulatorController struct {
	A              []ebiten.StandardGamepadButton
	B              []ebiten.StandardGamepadButton
	Select         []ebiten.StandardGamepadButton
	Start          []ebiten.StandardGamepadButton
	Left           []ebiten.StandardGamepadButton
	Right          []ebiten.StandardGamepadButton
	Up             []ebiten.StandardGamepadButton
	Down           []ebiten.StandardGamepadButton
	R              []ebiten.StandardGamepadButton
	L              []ebiten.StandardGamepadButton
	X              []ebiten.StandardGamepadButton
	Y              []ebiten.StandardGamepadButton
	Hinge          []ebiten.StandardGamepadButton
	Debug          []ebiten.StandardGamepadButton
	LayoutToggle   []ebiten.StandardGamepadButton
	SizingToggle   []ebiten.StandardGamepadButton
	RotationToggle []ebiten.StandardGamepadButton
	ExportScene    []ebiten.StandardGamepadButton
}
