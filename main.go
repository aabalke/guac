package main

import (
	"flag"
	"github.com/aabalke33/guac/sdl"
)

func main() {

	romPath := flag.String("r", "", "rom path")
    profile := flag.Bool("p", false, "use profiler")
	flag.Parse()

    s := sdl.SDLStruct{}
    s.Init()
    defer s.Close()

    s.Run(*romPath, *profile)
}
