package main

import (
	"flag"
	"time"
    "fmt"
	"github.com/aabalke33/guac/sdl"
)

const (
    MAX_INSTR = 0
    FPS = 60
)

func main() {

	romPath := flag.String("r", "", "rom path")
	debug := flag.Bool("debug", false, "debug")
	flag.Parse()

	start := time.Now().UnixMilli()

    s := sdl.SDLStruct{}
    s.Init(*debug)
    defer s.Close(*debug)

    s.Update(*debug, *romPath)

	end := time.Now().UnixMilli()
	fmt.Printf("\n\nRuntime: %d ms\n\n", end-start)
}
