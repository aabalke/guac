package config

import (
	"fmt"
	"image/color"
	"log"
	"os"

	"github.com/BurntSushi/toml"
    _"embed"
)

//go:embed default.toml
var defaultConfig []byte

const CONFIG_PATH = "./config.toml"

var Conf Config

type Config struct {
	Fullscreen     bool `toml:"fullscreen"`
	TomlBackdrop   int  `toml:"backdrop_color"`
	GamesPerRow int `toml:"games_per_row"`
	Backdrop       color.Color
	Gb             GbConfig       `toml:"gb"`
	Gba            GbaConfig      `toml:"gba"`
	KeyboardConfig KeyboardConfig `toml:"keyboard"`
	ControllerConfig ControllerConfig `toml:"controller"`
}

type GbConfig struct {
	TomlPalette    []int `toml:"dmg_palette"`
	Palette        [][]uint8
	KeyboardConfig EmulatorKeyboardConfig `toml:"keyboard"`
	ControllerConfig EmulatorControllerConfig `toml:"controller"`
}

type GbaConfig struct {
	KeyboardConfig EmulatorKeyboardConfig `toml:"keyboard"`
	ControllerConfig EmulatorControllerConfig `toml:"controller"`
}

type KeyboardConfig struct {
	Select     []string `toml:"select"`
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
	Select     []int `toml:"select"`
	Mute       []int `toml:"mute"`
	Pause      []int `toml:"pause"`
	Left       []int `toml:"left"`
	Right      []int `toml:"right"`
	Up         []int `toml:"up"`
	Down       []int `toml:"down"`
	Fullscreen []int `toml:"fullscreen"`
	Quit       []int `toml:"quit"`
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
}

type EmulatorControllerConfig struct {
	A      []int `toml:"a"`
	B      []int `toml:"b"`
	Select []int `toml:"select"`
	Start  []int `toml:"start"`
	Left   []int `toml:"left"`
	Right  []int `toml:"right"`
	Up     []int `toml:"up"`
	Down   []int `toml:"down"`
	R      []int `toml:"r"`
	L      []int `toml:"l"`
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

	c.Backdrop = color.RGBA{
		R: uint8(c.TomlBackdrop >> 16),
		G: uint8(c.TomlBackdrop >> 8),
		B: uint8(c.TomlBackdrop),
		A: 0xFF,
	}

    if c.GamesPerRow == 0 {
        errMessageStart := "Invalid Config:"
        errMessageEnd := "Using 6 games per row in menu."
		log.Printf("%s %s %s\n", errMessageStart, "GamesPerRow == 0.", errMessageEnd)
        c.GamesPerRow = 6
    }

	c.decodeGb()
}

func (c *Config) decodeGb() {

	pal := c.Gb.TomlPalette

	invalid := false

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
		c.Gb.Palette = [][]uint8{
			{0xFF, 0xFF, 0xFF},
			{0xCC, 0xCC, 0xCC},
			{0x77, 0x77, 0x77},
			{0x00, 0x00, 0x00},
		}
	} else {
		c.Gb.Palette = [][]uint8{
			{uint8(pal[0] >> 16), uint8(pal[0] >> 8), uint8(pal[0])},
			{uint8(pal[1] >> 16), uint8(pal[1] >> 8), uint8(pal[1])},
			{uint8(pal[2] >> 16), uint8(pal[2] >> 8), uint8(pal[2])},
			{uint8(pal[3] >> 16), uint8(pal[3] >> 8), uint8(pal[3])},
		}
	}
}
