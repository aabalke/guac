package emu

import (
	"github.com/veandco/go-sdl2/sdl"
)

type Emulator interface {
    GetPixels() []byte
    InputHandler(e sdl.Event)
    Update(*bool, int) int
    GetSize() (int32, int32)
    TogglePause() bool
    ToggleMute() bool
}
