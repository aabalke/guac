package config

import (
	"log"

	sys "golang.org/x/sys/cpu"
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

var clr = map[string]uint8 {
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

type NdsFirmware struct {
    Nickname string `toml:"nickname"`
    Message string `toml:"message"`
    FavoriteColor string `toml:"favorite_color"`
    BirthdayMonth uint8 `toml:"birthday_month"`
    BirthdayDay   uint8 `toml:"birthday_month"`
    Color         uint8
}

func (c *Config) decodeNdsFirmware() {

    f := &c.Nds.NdsFirmware

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

type NdsJit struct {
    Enabled bool `toml:"enabled"`
    PageSize uint32 `toml:"page_size"`
    PageCount uint64 `toml:"page_count"`
    BatchInst uint32 `toml:"batch_inst"`
}

func (c *Config) decodeNdsJit() {

    if Conf.Nds.NdsJit.Enabled && !sys.X86.HasSSE2 {

        errMessageStart := "Invalid Config:"
        errMessageEnd := "Disabling Jit Compiler."
		log.Printf("%s %s %s\n", errMessageStart, "native machine not x86 instruction set.", errMessageEnd)
        //fmt.Printf("Warning)
        Conf.Nds.NdsJit.Enabled = false
    }

    if !Conf.Nds.NdsJit.Enabled {
        Conf.Nds.NdsJit.BatchInst = 1
    }

    if Conf.Nds.NdsJit.PageCount == 0 {
        errMessageStart := "Invalid Config:"
        errMessageEnd := "Setting Jit Compiler. Page count to 1024."
		log.Printf("%s %s\n", errMessageStart, errMessageEnd)
        //fmt.Printf("Warning)
        Conf.Nds.NdsJit.PageCount = 0x1024_0000
    }
}
