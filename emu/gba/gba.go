package gba

import (
	"fmt"

	cart "github.com/aabalke33/guac/emu/gba/cart"
	"github.com/aabalke33/guac/emu/gba/utils"
)

const (
	SCREEN_WIDTH  = 240
	SCREEN_HEIGHT = 160
)

var (
    _ = fmt.Sprintln("")
    CURR_INST = 0
    //MAX_COUNT = 21500
    //MAX_COUNT = 21853 DMA Test WRAM addr
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
    Joypad uint16
    Halted bool
    ExitHalt bool

    Clock int
    FPS int
    Cycles int
    Scanline int
 
    Timers Timers
    Dma    [4]DMA

    Gt *GraphicsTiming
}

func (gba *GBA) Update(exit *bool, instCount int) int {

    gt := gba.Gt
    r := &gba.Cpu.Reg.R

    if gba.Paused {
        return 0
    }

    gt.reset()

    //for range MAX_COUNT + 1 {
    for gba.Gt.RefreshCycles < (gba.Clock / gba.FPS) {

        cycles := 4

        opcode := gba.Mem.Read32(r[PC])

        if gba.Halted && gba.ExitHalt {
            gba.Halted = false
        } 

        if !gba.Halted {
            cycles = gba.Cpu.Execute(opcode)
        }

        //if CURR_INST == MAX_COUNT {
        //    gba.Paused = true
        //    gba.Debugger.print(CURR_INST)
        //}

        gba.updateIRQ()

        CURR_INST++

        gba.Timers.Increment(uint32(cycles)* 4)
        gt.update(cycles)
	}

    gba.checkDmas(DMA_MODE_REF)

    gba.graphics()

    return instCount
}

func (gba *GBA) checkDmas(mode uint32) {
    for i := range gba.Dma {
        if gba.Dma[i].checkMode(mode) {
            gba.Dma[i].transfer()
        }
    }
}


func NewGBA() *GBA {

	pixels := make([]byte, SCREEN_WIDTH*SCREEN_HEIGHT*4)

	gba := GBA{
        Clock: 16_780_000,
        FPS: 60, // 59.7374117111
        Pixels: &pixels,
    }

    gba.Mem = NewMemory(&gba)
    gba.Cpu = NewCpu(&gba)
    gba.Debugger = &Debugger{&gba}
    gba.Gt = &GraphicsTiming{
        Gba: &gba,
    }

    gba.Timers[0].Gba = &gba
    gba.Timers[1].Gba = &gba
    gba.Timers[2].Gba = &gba
    gba.Timers[3].Gba = &gba

    gba.Timers[0].Idx = 0
    gba.Timers[1].Idx = 1
    gba.Timers[2].Idx = 2
    gba.Timers[3].Idx = 3

    gba.Dma[0].Gba = &gba
    gba.Dma[1].Gba = &gba
    gba.Dma[2].Gba = &gba
    gba.Dma[3].Gba = &gba

    gba.Dma[0].Idx = 0
    gba.Dma[1].Idx = 1
    gba.Dma[2].Idx = 2
    gba.Dma[3].Idx = 3

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

func (gba *GBA) updateIRQ() {

    //for i := range 160 {
    //    // scanline
    //}

    const DISPSTAT = 0x0400_0004
    if dispstat := gba.Mem.Read8(DISPSTAT); utils.BitEnabled(dispstat, 3) {
        const IRQ_VBLANK = 0x00
        gba.triggerIRQ(IRQ_VBLANK)
    }
}

func (gba *GBA) triggerIRQ(irq uint32) {

    const IF = 0x202

    gba.Mem.BIOS_MODE = BIOS_IRQ

    iack := uint16(gba.Mem.Read8(IF))
	iack = iack | (1 << irq)
	gba.Mem.IO[IF] = uint8(iack)
    gba.Mem.IO[IF+1] = uint8(iack>>8)
	gba.Halted = false

	gba.checkIRQ()
}

func (gba *GBA) checkIRQ() {

    const IE = 0x200
    const IF = 0x202
    const IME = 0x208

    cond1 := !gba.Cpu.Reg.CPSR.GetFlag(FLAG_I)
    cond2 := gba.Mem.IO[IME] & 1 > 0
    cond3 := (uint16(gba.Mem.Read8(IE)) & uint16(gba.Mem.Read8(IF))) > 0
    if cond1 && cond2 && cond3 {
        panic("EXECPTION IN CHECK IRQ")
        //g.exception(irqVec, IRQ)
    }
}
