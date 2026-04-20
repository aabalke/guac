package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"image"
	"image/png"
	_ "image/png"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2"

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

func main() {

	//go debugMemoryErrors()

	config.Conf.Decode()

	flags := getFlags()

    if flags.PrintFps {
        go printTPS()
    }

	switch {
	case flags.FPS != 60:
		ebiten.SetTPS(flags.FPS)

	case flags.Profile:

		fi, err := os.Create("cpu.prof")
		if err != nil {
			panic(err)
		}

		f = fi

		ebiten.SetTPS(UNLIMITED_FPS)

	case flags.Unlimited:
		ebiten.SetTPS(UNLIMITED_FPS)
	}

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("guac emulator")
	ebiten.SetWindowIcon([]image.Image{loadIcon()})
	//ebiten.SetWindowPosition(100, 100)
	ebiten.SetWindowSize(256*4, 192*4)
	//ebiten.SetWindowSize(1280, 720)

	ebiten.SetVsyncEnabled(!config.Conf.VsyncDisabled)

	ebiten.SetCursorMode(ebiten.CursorModeHidden)
	if config.Conf.Fullscreen {
		ebiten.SetFullscreen(true)
	}

	opts := &ebiten.RunGameOptions{
		//SingleThread: true,
		//InitUnfocused: true,
	}

	g := NewGame(flags)

	if flags.Profile || flags.Unlimited {
		g.unlimitedFPS = true
	}

	err := ebiten.RunGameWithOptions(g, opts)
	if err != nil && err != exit {
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
	FPS         int
	Muted       bool
    PrintFps    bool
}

func getFlags() Flags {
	romPath := flag.String("r", "", "rom path")
	profile := flag.Bool("p", false, "use profiler")
	unlimited := flag.Bool("u", false, "unlimited tps")
	fps := flag.Int("fps", 60, "fps limit")
	muted := flag.Bool("m", false, "is muted")
	printfps := flag.Bool("print", false, "print fps")

	flag.Parse()

	f := Flags{
		RomPath:   *romPath,
		Profile:   *profile,
		Unlimited: *unlimited,
		FPS:       *fps,
		Muted:     *muted,
        PrintFps : *printfps,
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

// make sure no memory leaks
func debugMemoryErrors() {

	const maxMemoryAlloc = (1 << 30)

	var m runtime.MemStats

	for {
		runtime.ReadMemStats(&m)

		if m.Alloc > maxMemoryAlloc {
			panic("Memory Limit Exceeded")
		}

		time.Sleep(time.Second)
	}
}

func printTPS() {

	t := time.NewTicker(time.Second * 2)

	for range t.C {
		fmt.Printf("TPS % 4.2f\n", ebiten.ActualTPS())
	}
}
