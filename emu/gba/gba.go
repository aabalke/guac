package gba

import (
	cart "github.com/aabalke33/guac/emu/gba/cart"
)

const (
	SCREEN_WIDTH  = 240
	SCREEN_HEIGHT = 160
)

var (
    CURR_INST = 0
    //MAX_COUNT = 380
    MAX_COUNT = 380
)

type GBA struct {
    Debugger *Debugger
    Cartridge *cart.Cartridge
    Cpu *Cpu
    Mem *Memory
	Screen [SCREEN_WIDTH][SCREEN_HEIGHT]uint32
	Pixels *[]byte

	Paused bool
	Muted  bool
}

func NewGBA() *GBA {

	gba := GBA{}
	pixels := make([]byte, SCREEN_WIDTH*SCREEN_HEIGHT*4)
	gba.Pixels = &pixels

    gba.Mem = NewMemory(&gba)
    gba.Cpu = NewCpu(&gba)
    gba.Debugger = &Debugger{&gba}

	return &gba
}

func (gba *GBA) GetSize() (int32, int32) {
	return SCREEN_HEIGHT, SCREEN_WIDTH
}

func (gba *GBA) GetPixels() []byte {
	return *gba.Pixels
}

func (gba *GBA) ToggleMute() bool {
	gba.Muted = !gba.Muted
	return gba.Muted
}

func (gba *GBA) TogglePause() bool {
	gba.Paused = !gba.Paused
	return gba.Paused
}

func (gba *GBA) Close() {
	gba.Muted = true
	gba.Paused = true
}

func (gba *GBA) LoadGame(path string) {

	gba.Cartridge = cart.NewCartridge(path, "")
}

func (gba *GBA) Update() bool {

    for range MAX_COUNT + 1 {
        if CURR_INST == MAX_COUNT {
            gba.Debugger.print(CURR_INST)
            return true
        }

        opcode := gba.Mem.Read32(gba.Cpu.Reg.R[15])

        gba.Cpu.Execute(opcode)

        CURR_INST++
	}

    return true
}
