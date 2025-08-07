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
)

//go:embed icons/icon.png
var icon []byte

func main() {

	config.Conf.Decode()

	flags := getFlags()

	var f *os.File

	if flags.Profile {

		f, err := os.Create("cpu.prof")
		if err != nil {
			panic(err)
		}

		pprof.StartCPUProfile(f)

		//ebiten.SetTPS(480)
		ebiten.SetTPS(2000)
	}
	//ebiten.SetTPS(2000)

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("guac emulator")
	ebiten.SetWindowIcon([]image.Image{loadIcon()})
	//ebiten.SetWindowPosition(100, 100)
	ebiten.SetWindowSize(240*4, 160*4)
	//ebiten.SetWindowSize(1280, 720)
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
