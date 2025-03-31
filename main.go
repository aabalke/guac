package main

import (
	"time"
    "fmt"
	"github.com/aabalke33/guac/sdl"
)

const (
    MAX_INSTR = 0
    FPS = 60
)

func main() {

	start := time.Now().UnixMilli()

    s := sdl.SDLStruct{}
    s.Init()
    defer s.Close()

    s.Update()

	end := time.Now().UnixMilli()
	fmt.Printf("\n\nRuntime: %d ms\n\n", end-start)
}
