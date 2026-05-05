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
	Sidebar SidebarLocalization `toml:"sidebar"`
	General GeneralLocalization `toml:"general"`
	Ui      UiLocalization      `toml:"ui"`
	Gb      GbLocalization      `toml:"gb"`
	Gba     GbaLocalization     `toml:"gba"`
	Nds     NdsLocalization     `toml:"nds"`
}

type SidebarLocalization struct {
	General string `toml:"general"`
	Ui      string `toml:"ui"`
	Gb      string `toml:"gb"`
	Gba     string `toml:"gba"`
	Nds     string `toml:"nds"`
	Return  string `toml:"return"`
}

type GeneralLocalization struct {
	General              string `toml:"general"`
	Muted                string `toml:"muted"`
	ShowFps              string `toml:"show_fps"`
	InitFullscreen       string `toml:"init_fullscreen"`
	TargetFps            string `toml:"target_fps"`
	Keyboard             string `toml:"keyboard"`
	Controller           string `toml:"controller"`
	Select               string `toml:"select"`
	Return               string `toml:"return"`
	Mute                 string `toml:"mute"`
	Pause                string `toml:"pause"`
	Left                 string `toml:"left"`
	Right                string `toml:"right"`
	Up                   string `toml:"up"`
	Down                 string `toml:"down"`
	Fullscreen           string `toml:"fullscreen"`
	Quit                 string `toml:"quit"`
	KeyboardSelect       string `toml:"keyboard_select"`
	KeyboardReturn       string `toml:"keyboard_return"`
	KeyboardMute         string `toml:"keyboard_mute"`
	KeyboardPause        string `toml:"keyboard_pause"`
	KeyboardLeft         string `toml:"keyboard_left"`
	KeyboardRight        string `toml:"keyboard_right"`
	KeyboardUp           string `toml:"keyboard_up"`
	KeyboardDown         string `toml:"keyboard_down"`
	KeyboardFullscreen   string `toml:"keyboard_fullscreen"`
	KeyboardQuit         string `toml:"keyboard_quit"`
	ControllerSelect     string `toml:"controller_select"`
	ControllerReturn     string `toml:"controller_return"`
	ControllerMute       string `toml:"controller_mute"`
	ControllerPause      string `toml:"controller_pause"`
	ControllerLeft       string `toml:"controller_left"`
	ControllerRight      string `toml:"controller_right"`
	ControllerUp         string `toml:"controller_up"`
	ControllerDown       string `toml:"controller_down"`
	ControllerFullscreen string `toml:"controller_fullscreen"`
	ControllerQuit       string `toml:"controller_quit"`
	Save                 string `toml:"save"`
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
}

type GbLocalization struct {
	DmgPalette       string `toml:"dmg_palette"`
	Lightest         string `toml:"lightest"`
	Light            string `toml:"light"`
	Dark             string `toml:"dark"`
	Darkest          string `toml:"darkest"`
	DmgLightest      string `toml:"dmg_lightest"`
	DmgLight         string `toml:"dmg_light"`
	DmgDark          string `toml:"dmg_dark"`
	DmgDarkest       string `toml:"dmg_darkest"`
	ApplyPalette     string `toml:"apply_palette"`
	Keyboard         string `toml:"keyboard"`
	Controller       string `toml:"controller"`
	A                string `toml:"a"`
	B                string `toml:"b"`
	Select           string `toml:"select"`
	Start            string `toml:"start"`
	Left             string `toml:"left"`
	Right            string `toml:"right"`
	Up               string `toml:"up"`
	Down             string `toml:"down"`
	KeyboardA        string `toml:"keyboard_a"`
	KeyboardB        string `toml:"keyboard_b"`
	KeyboardSelect   string `toml:"keyboard_select"`
	KeyboardStart    string `toml:"keyboard_start"`
	KeyboardLeft     string `toml:"keyboard_left"`
	KeyboardRight    string `toml:"keyboard_right"`
	KeyboardUp       string `toml:"keyboard_up"`
	KeyboardDown     string `toml:"keyboard_down"`
	ControllerA      string `toml:"controller_a"`
	ControllerB      string `toml:"controller_b"`
	ControllerSelect string `toml:"controller_select"`
	ControllerStart  string `toml:"controller_start"`
	ControllerLeft   string `toml:"controller_left"`
	ControllerRight  string `toml:"controller_right"`
	ControllerUp     string `toml:"controller_up"`
	ControllerDown   string `toml:"controller_down"`
	Save             string `toml:"save"`
}

type GbaLocalization struct {
	General          string `toml:"general"`
	OptmizeIdleLoops string `toml:"optimize_idle_loops"`
	SoundClockCycles string `toml:"sound_clock_cycles"`
	Keyboard         string `toml:"keyboard"`
	Controller       string `toml:"controller"`
	A                string `toml:"a"`
	B                string `toml:"b"`
	Select           string `toml:"select"`
	Start            string `toml:"start"`
	Left             string `toml:"left"`
	Right            string `toml:"right"`
	Up               string `toml:"up"`
	Down             string `toml:"down"`
	L                string `toml:"l"`
	R                string `toml:"r"`
	KeyboardA        string `toml:"keyboard_a"`
	KeyboardB        string `toml:"keyboard_b"`
	KeyboardSelect   string `toml:"keyboard_select"`
	KeyboardStart    string `toml:"keyboard_start"`
	KeyboardLeft     string `toml:"keyboard_left"`
	KeyboardRight    string `toml:"keyboard_right"`
	KeyboardUp       string `toml:"keyboard_up"`
	KeyboardDown     string `toml:"keyboard_down"`
	KeyboardL        string `toml:"keyboard_l"`
	KeyboardR        string `toml:"keyboard_r"`
	ControllerA      string `toml:"controller_a"`
	ControllerB      string `toml:"controller_b"`
	ControllerSelect string `toml:"controller_select"`
	ControllerStart  string `toml:"controller_start"`
	ControllerLeft   string `toml:"controller_left"`
	ControllerRight  string `toml:"controller_right"`
	ControllerUp     string `toml:"controller_up"`
	ControllerDown   string `toml:"controller_down"`
	ControllerL      string `toml:"controller_l"`
	ControllerR      string `toml:"controller_r"`
	Save             string `toml:"save"`
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

	KeyboardA              string `toml:"keyboard_a"`
	KeyboardB              string `toml:"keyboard_b"`
	KeyboardSelect         string `toml:"keyboard_select"`
	KeyboardStart          string `toml:"keyboard_start"`
	KeyboardLeft           string `toml:"keyboard_left"`
	KeyboardRight          string `toml:"keyboard_right"`
	KeyboardUp             string `toml:"keyboard_up"`
	KeyboardDown           string `toml:"keyboard_down"`
	KeyboardL              string `toml:"keyboard_l"`
	KeyboardR              string `toml:"keyboard_r"`
	KeyboardX              string `toml:"keyboard_x"`
	KeyboardY              string `toml:"keyboard_y"`
	KeyboardHinge          string `toml:"keyboard_hinge"`
	KeyboardDebug          string `toml:"keyboard_debug"`
	KeyboardLayoutToggle   string `toml:"keyboard_layout_toggle"`
	KeyboardSizingToggle   string `toml:"keyboard_sizing_toggle"`
	KeyboardRotationToggle string `toml:"keyboard_rotation_toggle"`
	KeyboardExportToggle   string `toml:"keyboard_export_toggle"`

	ControllerA      string `toml:"controller_a"`
	ControllerB      string `toml:"controller_b"`
	ControllerSelect string `toml:"controller_select"`
	ControllerStart  string `toml:"controller_start"`
	ControllerLeft   string `toml:"controller_left"`
	ControllerRight  string `toml:"controller_right"`
	ControllerUp     string `toml:"controller_up"`
	ControllerDown   string `toml:"controller_down"`
	ControllerL      string `toml:"controller_l"`
	ControllerR      string `toml:"controller_r"`
	ControllerX      string `toml:"controller_x"`
	ControllerY      string `toml:"controller_y"`
	Save             string `toml:"save"`
}
