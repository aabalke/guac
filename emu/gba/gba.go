package gba

import (
	"fmt"
    "os"

	"github.com/aabalke33/guac/emu/gba/utils"
	"github.com/aabalke33/guac/emu/gba/cart"
)

const (
	SCREEN_WIDTH  = 240
	SCREEN_HEIGHT = 160
)

var (
    _ = fmt.Sprintln("")
    _ = os.Args
    CURR_INST = 0
    //MAX_COUNT = 100_000

    IN_EXCEPTION = false
    EXCEPTION_COUNT = 0
)

type GBA struct {
    GamePath string
    Debugger *Debugger
    Cartridge *cart.Cartridge
    Cpu *Cpu
    Mem *Memory
	Screen [SCREEN_WIDTH][SCREEN_HEIGHT]uint32
	Pixels *[]byte
	DebugPixels *[]byte

	Paused bool
	Muted  bool 
    Joypad uint16
    Halted bool
    ExitHalt bool

    Apu *APU

    Clock int
    FPS int
    Cycles int
    Scanline int
 
    Timers Timers
    Dma    [4]DMA

    VCOUNT uint32

    GraphicsDirty bool

    IntrWait uint32

    Save bool

    InterruptStack *InterruptStack

    GBA_LOCK bool

    OpenBusOpcode uint32

    DmaOnRefresh bool
}

var SAVED_CYCLES = uint32(0)

func (gba *GBA) Update(exit *bool, instCount int) int {

    if gba.Paused {
        return 0
    }

    gba.VCOUNT = 0

    frameCycles := uint32(0)
    for range 160 {
        frameCycles = gba.UpdateScanline(frameCycles)
    }

    dispstat := uint32(gba.Mem.Dispstat)
    if utils.BitEnabled(dispstat, 3) {
        gba.triggerIRQ(0, "VBLANK")
    }

    gba.Mem.Dispstat.SetVBlank(true)

    gba.checkDmas(DMA_MODE_VBL)

    for range 68 {
        frameCycles = gba.UpdateScanline(frameCycles)
    }

    gba.Mem.Dispstat.SetVBlank(false)

    //frameCycles = gba.UpdateScanline(frameCycles) // 227

    SAVED_CYCLES = frameCycles - (1232 * 228)

    //gba.checkDmas(DMA_MODE_REF)

    //gba.graphics()

    return instCount
}

func (gba *GBA) UpdateScanline(frameCycles uint32) uint32 {

    gba.scanlineGraphics(gba.VCOUNT)

    dispstat := gba.Mem.Read16(0x400_0004)
    lyc := gba.Mem.Read8(0x400_0005)
    vcounterIRQ := utils.BitEnabled(dispstat, 5)

    if vcounterIRQ && gba.VCOUNT == lyc {
        gba.triggerIRQ(2, "VCOUNTER")
    }

    if gba.VCOUNT == lyc {
        gba.Mem.Dispstat.SetVCFlag(true)
    } else {
        gba.Mem.Dispstat.SetVCFlag(false)
    }

    //Although the drawing time is only 960 cycles (240*4), the H-Blank flag is "0" for a total of 1006 cycles. gbatek
    frameCycles = gba.Exec(1006, frameCycles)

    dispstat = uint32(gba.Mem.Dispstat)
    if utils.BitEnabled(dispstat, 4) && gba.VCOUNT < 160 {
        gba.triggerIRQ(1, "HBLANK")
    }

    gba.Mem.Dispstat.SetHBlank(true)

    if vblank := utils.BitEnabled(dispstat, 0); !vblank { // this is necessary for dmas do not remove
        gba.checkDmas(DMA_MODE_HBL)
    }

    frameCycles = gba.Exec(1232 - 1006, frameCycles)

    gba.Mem.Dispstat.SetHBlank(false)
    //gba.Mem.IO[0x202] &^= 0b10

    //if gba.DmaOnRefresh {
        gba.Dma[1].transferFifo()
        gba.Dma[2].transferFifo()
    //}
    gba.Dma[3].transferVideo(gba.VCOUNT)

    gba.VCOUNT++

    return frameCycles
}

var AAAA bool

func (gba *GBA) Exec(requiredCycles, frameCycles uint32) uint32 {

    r := &gba.Cpu.Reg.R

    accCycles := frameCycles

    for accCycles < requiredCycles {

        cycles := 4

        if gba.Paused {
            return 0
        }

        opcode := gba.Mem.Read32(r[PC])

        gba.OpenBusOpcode = gba.Mem.Read32(r[PC] + 8)

        if (r[PC] >= 0x400_0000 && r[PC] < 0x800_0000) || r[PC] % 2 != 0 {

            r := &gba.Cpu.Reg.R

            panic(fmt.Sprintf("INVALID PC CURR %d PC %08X OPCODE %08X", CURR_INST, r[PC], gba.Mem.Read32(r[PC])))
        }

        if gba.Halted {
            AckIntrWait(gba)
        }

        if gba.Halted && gba.ExitHalt {
            gba.Halted = false
        } 

        if !gba.Halted {
            cycles = gba.Cpu.Execute(opcode)
        }

        //if r[PC] == 0x800_0230 {
        //    //fmt.Printf("PC %08X R2 %08X CURR %d\n", r[PC], r[2], CURR_INST)
        //    gba.Debugger.print(CURR_INST)
        //    gba.Paused = true
        //    os.Exit(0)
        //}

        if !gba.Halted {
            CURR_INST++
        }

        //gba.Timers.Increment(uint32(cycles))
        gba.Timers.Update(uint32(cycles))

        accCycles += uint32(cycles)
    }

    return accCycles - requiredCycles
}

var FLAG bool

var PREV_VALUE uint32

func (gba *GBA) checkDmas(mode uint32) {
    for i := range gba.Dma {
        if gba.Dma[i].checkMode(mode) {
            gba.Dma[i].transfer()
        }
    }
}

func NewGBA() *GBA {

	pixels := make([]byte, SCREEN_WIDTH*SCREEN_HEIGHT*4)
	debugPixels := make([]byte, 1080*1080*4)

	gba := GBA{
        Clock: 16_780_000,
        FPS: 60, // 59.7374117111
        Pixels: &pixels,
        DebugPixels: &debugPixels,
    }

    gba.InterruptStack = &InterruptStack{
        Gba: &gba,
        Interrupts: []Interrupt{},
        Skip: false,
        Print: false,
    }

    gba.Debugger = &Debugger{Gba: &gba, Version: 1}

    gba.Mem = NewMemory(&gba)
    gba.Cpu = NewCpu(&gba)

    gba.Cpu.Reg.CPSR.SetFlag(FLAG_I, false)

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

	gba.Apu = &APU{
		SampleRate: 44100,
		Enabled:    true,
        gba: &gba,
	}

    gba.Apu.Init()

    //gba.Logger = NewLogger(".log.txt", &gba)

    gba.LoadBios("./emu/gba/res/bios_magia.gba")

    //gba.SoftReset()

	return &gba
}

func (gba *GBA) GetSize() (int32, int32) {
	return SCREEN_HEIGHT, SCREEN_WIDTH
}

func (gba *GBA) GetPixels() []byte {
	return *gba.Pixels
}

func (gba *GBA) GetDebugPixels() []byte {
	return *gba.DebugPixels
}

func (gba *GBA) ToggleMute() bool {
	gba.Muted = !gba.Muted
	return gba.Muted
}

func (gba *GBA) TogglePause() bool {
	gba.Paused = !gba.Paused
	return gba.Paused
}

func (gba *GBA) ToggleSaveState() {
    path := gba.GamePath + ".gob"
    SaveState(gba, path)
}

func (gba *GBA) Close() {
	gba.Muted = true
	gba.Paused = true
}

func (gba *GBA) LoadGame(path string, useState bool) {
    gba.GamePath = path
    gba.Cartridge = cart.NewCartridge(path, path + ".save")

    if useState {
        path := gba.GamePath + ".gob"
        LoadState(gba, path)
        gba.Paused = true
    }
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

func (gba *GBA) triggerIRQ(irq uint32, cause string) {

    mem := gba.Mem

    mem.BIOS_MODE = BIOS_IRQ

    iack := mem.Read16(0x400_0202)

    iack |= (1 << irq)
    mem.IO[0x202] = uint8(iack)
    mem.IO[0x203] = uint8(iack>>8)
	gba.Halted = false

    //fmt.Printf("IRQ EXCEPTION CHECK AT TRIGGER\n")
	gba.checkIRQ(cause)
}

func (gba *GBA) checkIRQ(cause string) {

    interruptEnabled := !gba.Cpu.Reg.CPSR.GetFlag(FLAG_I)
    ime := utils.BitEnabled(gba.Mem.Read16(0x400_0208), 0)
    interrupts := (gba.Mem.Read16(0x400_0200) & gba.Mem.Read16(0x400_0202)) != 0
    if interruptEnabled && ime && interrupts {
        //if AAAA {
        //    fmt.Printf("INTERRUPT INTERRUPT %s\n", cause)
        //}
        gba.InterruptStack.Execute(cause)
    }
}
