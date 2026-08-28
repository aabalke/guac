package ui

import (
	"io"

	"github.com/BurntSushi/toml"
)

type LangOptions int

const (
	ENGLISH LangOptions = iota
	SPANISH
)

var langAbbreviations = map[LangOptions]string{
	ENGLISH: "en",
	SPANISH: "es",
}

func NewLocalization(lang LangOptions) *Localization {
	abb, ok := langAbbreviations[lang]
	if !ok {
		abb = "en"
	}

	path := "assets/local/" + abb + ".toml"

	f, err := embeddedAssets.Open(path)
	if err != nil {
		panic("could not get localization file")
	}

	l := &Localization{}

	data, err := io.ReadAll(f)
	if err != nil {
		panic("could not read localization file")
	}

	_, err = toml.Decode(string(data), l)
	if err != nil {
		panic("could not decode localization file")
	}

	return l
}

type Localization struct {
	Main     MainLocalization     `toml:"main"`
	Pause    PauseLocalization    `toml:"pause"`
	Settings SettingsLocalization `toml:"settings"`
	Toast    ToastLocalization    `toml:"toast"`
}

type ToastLocalization struct {
	Saved                  string `toml:"saved"`
	Muted                  string `toml:"muted"`
	Unmuted                string `toml:"unmuted"`
	ControllerConnected    string `toml:"controller_connected"`
	ControllerDisconnected string `toml:"controller_disconnected"`
}

type MainLocalization struct {
	Open        string `toml:"open"`
	Settings    string `toml:"settings"`
	Quit        string `toml:"quit"`
	DialogTitle string `toml:"dialog_title"`
	DialogDesc  string `toml:"dialog_desc"`
}

type PauseLocalization struct {
	Resume   string `toml:"resume"`
	Settings string `toml:"settings"`
	Main     string `toml:"main"`
}

type SettingsLocalization struct {
	Sidebar         SidebarLocalization `toml:"sidebar"`
	General         GeneralLocalization `toml:"general"`
	Ui              UiLocalization      `toml:"ui"`
	Gb              GbLocalization      `toml:"gb"`
	Gba             GbaLocalization     `toml:"gba"`
	Nds             NdsLocalization     `toml:"nds"`
	About           AboutLocalization   `toml:"about"`
	ColorCorrection ColorCorrection     `toml:"color_correction"`
}

type SidebarLocalization struct {
	General string `toml:"general"`
	Ui      string `toml:"ui"`
	Gb      string `toml:"gb"`
	Gba     string `toml:"gba"`
	Nds     string `toml:"nds"`
	About   string `toml:"about"`
	Return  string `toml:"return"`
}

type AboutLocalization struct {
	About string `toml:"about"`

	Version   string `toml:"version"`
	Copyright string `toml:"copyright"`
	ThankYous string `toml:"thank_yous"`
}

type GeneralLocalization struct {
	General             string `toml:"general"`
	Muted               string `toml:"muted"`
	ShowFps             string `toml:"show_fps"`
	InitFullscreen      string `toml:"init_fullscreen"`
	TargetFps           string `toml:"target_fps"`
	VsyncEnabled        string `toml:"vsync_enabled"`
	DisableSaves        string `toml:"disable_saves"`
	IntegerScalingDesc  string `toml:"integer_scaling_desc"`
	IntegerScaling      string `toml:"integer_scaling"`
	IntegerScalingRatio string `toml:"integer_scaling_ratio"`
	SampleRateDesc      string `toml:"sample_rate_desc"`
	SampleRate          string `toml:"sample_rate"`
	Keyboard            string `toml:"keyboard"`
	Controller          string `toml:"controller"`
	Select              string `toml:"select"`
	Return              string `toml:"return"`
	Mute                string `toml:"mute"`
	Pause               string `toml:"pause"`
	Left                string `toml:"left"`
	Right               string `toml:"right"`
	Up                  string `toml:"up"`
	Down                string `toml:"down"`
	Fullscreen          string `toml:"fullscreen"`
	Quit                string `toml:"quit"`
	Save                string `toml:"save"`
}

type UiLocalization struct {
	Ui        string   `toml:"ui"`
	Language  string   `toml:"language"`
	Languages []string `toml:"languages"`

	Backdrop    string `toml:"backdrop"`
	BgColor     string `toml:"bg_color"`
	FgColor     string `toml:"fg_color"`
	AccentColor string `toml:"accent_color"`
	ApplyTheme  string `toml:"apply_theme"`

	UiBackdrop    string `toml:"ui_backdrop"`
	UiBgColor     string `toml:"ui_bg_color"`
	UiFgColor     string `toml:"ui_fg_color"`
	UiAccentColor string `toml:"ui_accent_color"`
	Save          string `toml:"save"`

	Alphabet []string `toml:"alphabet"`
}

type GbLocalization struct {
	General      string   `toml:"general"`
	System       string   `toml:"system"`
	Systems      []string `toml:"systems"`
	DmgPalette   string   `toml:"dmg_palette"`
	Lightest     string   `toml:"lightest"`
	Light        string   `toml:"light"`
	Dark         string   `toml:"dark"`
	Darkest      string   `toml:"darkest"`
	DmgLightest  string   `toml:"dmg_lightest"`
	DmgLight     string   `toml:"dmg_light"`
	DmgDark      string   `toml:"dmg_dark"`
	DmgDarkest   string   `toml:"dmg_darkest"`
	ApplyPalette string   `toml:"apply_palette"`
	Bios         string   `toml:"bios"`
	DmgPath      string   `toml:"dmg_bios_path"`
	GbcPath      string   `toml:"gbc_bios_path"`
	DirectBoot   string   `toml:"direct_boot"`
	Keyboard     string   `toml:"keyboard"`
	Controller   string   `toml:"controller"`
	A            string   `toml:"a"`
	B            string   `toml:"b"`
	Select       string   `toml:"select"`
	Start        string   `toml:"start"`
	Left         string   `toml:"left"`
	Right        string   `toml:"right"`
	Up           string   `toml:"up"`
	Down         string   `toml:"down"`
	Save         string   `toml:"save"`
}

type GbaLocalization struct {
	General          string   `toml:"general"`
	OptmizeIdleLoops string   `toml:"optimize_idle_loops"`
	Rotation         string   `toml:"rotation"`
	Rotations        []string `toml:"rotations"`

	Hardware         string   `toml:"hardware"`
	BackupType       string   `toml:"backup_type"`
	BackupTypes      []string `toml:"backup_types"`
	ForceRtc         string   `toml:"force_rtc"`
	ForceSolarSensor string   `toml:"force_solar_sensor"`
	SolarSensorLevel string   `toml:"solar_sensor_level"`

	Bios       string `toml:"bios"`
	BiosPath   string `toml:"bios_path"`
	DirectBoot string `toml:"direct_boot"`
	Keyboard   string `toml:"keyboard"`
	Controller string `toml:"controller"`

	A              string `toml:"a"`
	B              string `toml:"b"`
	Select         string `toml:"select"`
	Start          string `toml:"start"`
	Left           string `toml:"left"`
	Right          string `toml:"right"`
	Up             string `toml:"up"`
	Down           string `toml:"down"`
	L              string `toml:"l"`
	R              string `toml:"r"`
	SolarMin       string `toml:"solar_minimum"`
	Solar1         string `toml:"solar_1"`
	Solar2         string `toml:"solar_2"`
	Solar3         string `toml:"solar_3"`
	SolarMax       string `toml:"solar_maximum"`
	RotationToggle string `toml:"rotation_toggle"`

	Save string `toml:"save"`
}

type NdsLocalization struct {
	Screen          string   `toml:"screen"`
	Layout          string   `toml:"layout"`
	Sizing          string   `toml:"sizing"`
	Rotation        string   `toml:"rotation"`
	Layouts         []string `toml:"layouts"`
	Sizings         []string `toml:"sizings"`
	Rotations       []string `toml:"rotations"`
	Rtc             string   `toml:"rtc"`
	AdditionalHours string   `toml:"additional_hours"`
	Bios            string   `toml:"bios"`
	Arm7Path        string   `toml:"arm7_path"`
	Arm9Path        string   `toml:"arm9_path"`
	Firmware        string   `toml:"firmware"`
	FilePath        string   `toml:"file_path"`
	Nickname        string   `toml:"nickname"`
	Message         string   `toml:"message"`
	FavoriteColor   string   `toml:"favorite_color"`
	SceneExport     string   `toml:"scene_export"`
	OutputDirectory string   `toml:"output_directory"`
	ShadowPolygons  string   `toml:"shadow_polygons"`

	Keyboard       string `toml:"keyboard"`
	Controller     string `toml:"controller"`
	A              string `toml:"a"`
	B              string `toml:"b"`
	Select         string `toml:"select"`
	Start          string `toml:"start"`
	Left           string `toml:"left"`
	Right          string `toml:"right"`
	Up             string `toml:"up"`
	Down           string `toml:"down"`
	L              string `toml:"l"`
	R              string `toml:"r"`
	X              string `toml:"x"`
	Y              string `toml:"y"`
	Hinge          string `toml:"hinge"`
	Debug          string `toml:"debug"`
	LayoutToggle   string `toml:"layout_toggle"`
	SizingToggle   string `toml:"sizing_toggle"`
	RotationToggle string `toml:"rotation_toggle"`
	ExportToggle   string `toml:"export_toggle"`

	Save string `toml:"save"`
}

type ColorCorrection struct {
	ColorCorrection string   `toml:"color_correction"`
	Type            string   `toml:"type"`
	Types           []string `toml:"types"`
	Strength        string   `toml:"strength"`
}
