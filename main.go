package main

import (
	"bytes"
	_ "embed"
	"flag"
	"image"
	"image/png"
	_ "image/png"
	"log"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"os"
	"runtime/pprof"
)

const (
    NONE = iota
    GB
    GBA
)

//go:embed icons/icon.png
var icon []byte

func main() {

    flags := getFlags()

    var f *os.File

    if flags.Profile {

        f, err := os.Create("cpu.prof")
        if err != nil {
            panic(err)
        }

        pprof.StartCPUProfile(f)

        ebiten.SetTPS(400)
    }

    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    ebiten.SetWindowTitle("guac emulator")
    ebiten.SetWindowIcon([]image.Image{loadIcon()})
    ebiten.SetWindowSize(240 * 3, 160 * 3)

    if err := ebiten.RunGame(NewGame(flags)); err != nil && err != exit {
        log.Fatal(err)
    }

    if flags.Profile {
        pprof.StopCPUProfile()
        f.Close()
    }
}

func loadIcon() image.Image {

    img, err := png.Decode(bytes.NewReader(icon))
    if err != nil {
        panic(err)
    }

    return img
}

type Flags struct {
    ConsoleMode bool
    Type int
    RomPath string
    Profile bool
}

func getFlags() Flags {
    romPath := flag.String("r", "", "rom path")
    profile := flag.Bool("p", false, "use profiler")
    flag.Parse()

    f := Flags{
        RomPath: *romPath,
        Profile: *profile,
    }

    switch {
    case *romPath == "":
        f.Type = NONE
        f.ConsoleMode = true
    case strings.HasSuffix(*romPath, ".gb"):
        f.Type = GB
    case strings.HasSuffix(*romPath, ".gbc"):
        f.Type = GB
    case strings.HasSuffix(*romPath, ".gba"):
        f.Type = GBA
    default:
        panic("Flag Parsing Error. Rom Path must end with gba, gbc, gb extension")
    }

    return f
}
