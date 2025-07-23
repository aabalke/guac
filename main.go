package main

import (
	"flag"
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

func main() {


    flags := getFlags()

    var f *os.File

    if flags.Profile {

        f, err := os.Create("cpu.prof")
        if err != nil {
            panic(err)
        }

        pprof.StartCPUProfile(f)
        //ebiten.SetTPS(360)
    }

    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    ebiten.SetWindowTitle("Guac Emulator")

    if err := ebiten.RunGame(NewGame(flags)); err != nil && err != exit {
        log.Fatal(err)
    }

    if flags.Profile {
        pprof.StopCPUProfile()
        f.Close()
    }
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
