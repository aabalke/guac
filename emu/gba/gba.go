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
    MAX_COUNT = 73
    //MAX_COUNT = 73 // test against bios.asm
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

    Clock int
    FPS int
}

func NewGBA() *GBA {

	pixels := make([]byte, SCREEN_WIDTH*SCREEN_HEIGHT*4)

	gba := GBA{
        Clock: 16780000,
        FPS: 60,
        Pixels: &pixels,
    }

    gba.Mem = NewMemory(&gba)
    gba.Cpu = NewCpu(&gba)
    gba.Debugger = &Debugger{&gba}

    //gba.LoadBios("./emu/gba/bios.gba")
    //gba.LoadBios("./emu/gba/bios_custom.gba")

    // custom bios
    //gba.Mem.Write32(0x00, 0xE129F000)
    //gba.Mem.Write32(0x04, 0xE59FD00C)
    //gba.Mem.Write32(0x08, 0xE3A0F000)
    //gba.Mem.Write32(0x0C, 0xE1A0E00F)
    //gba.Mem.Write32(0x10, 0xEA00008D)
    //gba.Mem.Write32(0x14, 0xE3A00000)
    //gba.Mem.Write32(0x18, 0xE3A01002)
    //gba.Mem.Write32(0x1C, 0xE12FFF1E)

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

func (gba *GBA) Update(exit *bool, instCount int) int {

    if gba.Paused {
        return 0
    }

    updateCycles := 0
    for range MAX_COUNT + 1 {
    //for updateCycles < (gba.Clock / gba.FPS) {

        cycles := 4

        opcode := gba.Mem.Read32(gba.Cpu.Reg.R[15])

        gba.Cpu.Execute(opcode)

        if CURR_INST == MAX_COUNT {
            gba.Paused = true
            gba.Debugger.print(CURR_INST)
            gba.Debugger.saveBg2()
        }

        updateCycles += cycles
        CURR_INST++
        instCount++
	}

    gba.updateDisplay()

    return instCount
}

func (gba *GBA) toggleThumb() {

    reg := &gba.Cpu.Reg

    newFlag := reg.R[PC] & 1 > 0

    reg.CPSR.SetFlag(FLAG_T, newFlag)

    if newFlag {
        reg.R[PC] &^= 1
        // pipe
        return
    }

    reg.R[PC] &^= 3
    // pipe
}
