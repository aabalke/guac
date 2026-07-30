package file

import (
	_ "embed"

	"github.com/aabalke/guac/config"
)

//go:embed default.toml
var DEF_CONFIG []byte

const CONFIG_PATH = "./config.toml"

type Config struct {
	config  *config.Config
	General General `toml:"general"`
	Ui      Ui      `toml:"ui"`
	Profile Profile `toml:"profile"`
	Gb      Gb      `toml:"gb"`
	Gba     Gba     `toml:"gba"`
	Nds     Nds     `toml:"nds"`
}

type General struct {
	Muted               bool         `toml:"muted"`
	TargetFps           int          `toml:"target_fps"`
	ShowFps             bool         `toml:"show_fps"`
	InitFullscreen      bool         `toml:"fullscreen"`
	Vsync               bool         `toml:"vsync_enabled"`
	RomPath             string       `toml:"rom_path"`
	IntegerScaling      bool         `toml:"integer_scaling"`
	IntegerScalingRatio int          `toml:"integer_scaling_ratio"`
	SampleRate          int          `toml:"sample_rate"`
	DisableSaves        bool         `toml:"disable_saves"`
	Keyboard            GeneralInput `toml:"keyboard"`
	Controller          GeneralInput `toml:"controller"`
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
	Backdrop   string `toml:"backdrop_color"`
	Background string `toml:"menu_background_color"`
	Foreground string `toml:"menu_foreground_color"`
	Secondary  string `toml:"menu_secondary_color"`
	Language   string `toml:"language"`
}

type Profile struct {
	Enabled   bool   `toml:"enabled"`
	FilePath  string `toml:"file_path"`
	StartTick int64  `toml:"start_tick"`
	EndTick   int64  `toml:"end_tick"`
}

type Gb struct {
	Bios       GbBios        `toml:"bios"`
	Palette    []string      `toml:"dmg_palette"`
	Keyboard   EmulatorInput `toml:"keyboard"`
	Controller EmulatorInput `toml:"controller"`
}

type GbBios struct {
	DmgPath string `toml:"dmg_bios_path"`
	GbcPath string `toml:"gbc_bios_path"`
	Direct  bool   `toml:"direct_boot"`
}

type Gba struct {
	IdleOptimize bool          `toml:"idle_optimize"`
	Rotation     int           `toml:"rotation"`
	Hardware     GbaHardware   `toml:"hardware"`
	Bios         GbaBios       `toml:"bios"`
	Keyboard     EmulatorInput `toml:"keyboard"`
	Controller   EmulatorInput `toml:"controller"`
}

type GbaHardware struct {
	BackupType       string `toml:"backup_type"`
	ForceRtc         bool   `toml:"force_rtc"`
	ForceSolarSensor bool   `toml:"force_solar_sensor"`
	SolarSensorLevel int    `toml:"solar_sensor_level"`
}

type GbaBios struct {
	Path   string `toml:"bios_path"`
	Direct bool   `toml:"direct_boot"`
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

	SolarLevel0 []string `toml:"solar_min"`
	SolarLevel1 []string `toml:"solar_25"`
	SolarLevel2 []string `toml:"solar_50"`
	SolarLevel3 []string `toml:"solar_75"`
	SolarLevel4 []string `toml:"solar_max"`
}
