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

	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2"
	//_ "github.com/silbinarywolf/preferdiscretegpu" //no profiler change

	"os"
	"runtime/pprof"
)

const (
	NONE = iota
	GB
	GBA
	NDS
)

//go:embed icons/icon.png
var icon []byte

var f *os.File

var isProfiling bool
var profileFrames uint32

func main() {

	config.Conf.Decode()

	flags := getFlags()

	if flags.Profile {

        fi, err := os.Create("cpu.prof")
		if err != nil {
			panic(err)
		}

        f = fi

		pprof.StartCPUProfile(f)

		ebiten.SetTPS(2000)

	} else if flags.Unlimited {
		ebiten.SetTPS(2000)
	}

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("guac emulator")
	ebiten.SetWindowIcon([]image.Image{loadIcon()})
	//ebiten.SetWindowPosition(100, 100)
	ebiten.SetWindowSize(256*4, 192*4)
	//ebiten.SetWindowSize(1280, 720)

    ebiten.SetCursorMode(ebiten.CursorModeHidden)
	if config.Conf.Fullscreen {
		ebiten.SetFullscreen(true)
	}

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
	Type        int
	RomPath     string
	Profile     bool
	Unlimited   bool
}

func getFlags() Flags {
	romPath := flag.String("r", "", "rom path")
	profile := flag.Bool("p", false, "use profiler")
	unlimited := flag.Bool("u", false, "unlimited tps")

	flag.Parse()

	f := Flags{
		RomPath:   *romPath,
		Profile:   *profile,
		Unlimited: *unlimited,
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
	case strings.HasSuffix(*romPath, ".nds"):
		f.Type = NDS
	default:
		panic("Flag Parsing Error. Rom Path must end with gba, gbc, gb extension")
	}

	return f
}
