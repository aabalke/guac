package gba

import (
	"fmt"
    "os"

	"github.com/aabalke33/guac/emu/gba/utils"
	"github.com/aabalke33/guac/emu/gba/cart"
    "time"
)

const (
	SCREEN_WIDTH  = 240
	SCREEN_HEIGHT = 160
    VCOUNT = 0x6
)

var (
    _ = fmt.Sprintln("")
    _ = os.Args
    CURR_INST = 0
)

var start time.Time
var end time.Time

var (
    DRAWN = false
    video_cycles = uint32(0)
)

const (
    NUM_SCANLINES   = (SCREEN_HEIGHT + 68)

    CYCLES_HDRAW    = 1006
    CYCLES_HBLANK   = 226
    CYCLES_SCANLINE = (CYCLES_HDRAW + CYCLES_HBLANK)     // 1232
    CYCLES_VDRAW    = (CYCLES_SCANLINE * SCREEN_HEIGHT)  // 197120
    CYCLES_VBLANK   = (CYCLES_SCANLINE * 68)             // 83776
    CYCLES_FRAME    = (CYCLES_VDRAW + CYCLES_VBLANK)     // 280896
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
    Objects *[128]Object

    Clock int
    FPS int
    Cycles int
    Scanline int
 
    Timers Timers
    Dma    [4]DMA

    //VCOUNT uint32

    GraphicsDirty bool
    IntrWait uint32
    Save bool
    InterruptStack *InterruptStack
    GBA_LOCK bool
    OpenBusOpcode uint32
    DmaOnRefresh bool

    prevVcountMatch bool

    //Cache *Cache
}

func (gba *GBA) Update(exit *bool, instCount int) int {

    //gba.DrawFrame(exit, instCount)
    //return 0

    if gba.Paused {
        return 0
    }

    //gba.VCOUNT = 0

    gba.Mem.IO[VCOUNT] = 0

    frameCycles := uint32(0)
    for range 160 {
        frameCycles = gba.UpdateScanline(frameCycles)
    }

    gba.Mem.Dispstat.SetVBlank(true)

    gba.checkDmas(DMA_MODE_VBL)

    //frameCycles = gba.UpdateScanline(frameCycles)

    dispstat := uint32(gba.Mem.Dispstat)
    if utils.BitEnabled(dispstat, 3) {
        gba.InterruptStack.setIRQ(0)
    }

    for range 66 {
        frameCycles = gba.UpdateScanline(frameCycles)
    }

    gba.Mem.Dispstat.SetVBlank(false)

    frameCycles = gba.UpdateScanline(frameCycles)

    return instCount
}

func (gba *GBA) UpdateScanline(frameCycles uint32) uint32 {

    vcount := uint32(gba.Mem.IO[VCOUNT])

    dispstat := uint32(gba.Mem.Dispstat)
    lyc := gba.Mem.ReadIODirect(0x05, 1)
    match := lyc == vcount
    gba.Mem.Dispstat.SetVCFlag(match)

    vcounterIRQ := utils.BitEnabled(dispstat, 5)
    if vcounterIRQ && match {
        gba.InterruptStack.setIRQ(2)
    }

    //Although the drawing time is only 960 cycles (240*4), the H-Blank flag is "0" for a total of 1006 cycles. gbatek
    frameCycles += gba.Exec(1006)

    gba.scanlineGraphics(vcount)

    gba.Mem.Dispstat.SetHBlank(true)

    dispstat = uint32(gba.Mem.Dispstat)
    if utils.BitEnabled(dispstat, 4) {
        gba.InterruptStack.setIRQ(1)
    }

    if vblank := utils.BitEnabled(dispstat, 0); !vblank { // this is necessary for dmas do not remove
        gba.checkDmas(DMA_MODE_HBL)
    }

    frameCycles += gba.Exec(1232 - 1006)

    gba.Mem.Dispstat.SetHBlank(false)

    if gba.DmaOnRefresh {
        gba.Dma[1].transferFifo()
        gba.Dma[2].transferFifo()
    }

    gba.Dma[3].transferVideo(vcount)

    gba.Mem.IO[VCOUNT]++

    if gba.Mem.IO[VCOUNT] > 228 {
        panic("VCOUNT SET TO MORE THAN 227")
    }

    return frameCycles
}

func (gba *GBA) Exec(requiredCycles uint32) uint32 {

    r := &gba.Cpu.Reg.R

    accCycles := uint32(0)

    for accCycles < requiredCycles {

        cycles := 4

        if gba.Paused {
            return 0
        }

        if (r[PC] >= 0x400_0000 && r[PC] < 0x800_0000) || r[PC] % 2 != 0 || r[PC] >= 0xE00_0000 {
            r := &gba.Cpu.Reg.R
            panic(fmt.Sprintf("INVALID PC CURR %d PC %08X OPCODE %08X", CURR_INST, r[PC], gba.Mem.Read32(r[PC])))
        }

        if gba.Halted {
            AckIntrWait(gba)
        }

        opcode := gba.Mem.Read32(r[PC])
        gba.OpenBusOpcode = gba.Mem.Read32((r[PC] &^ 0b11) + 8)


        if !gba.Halted {
            cycles = gba.Cpu.Execute(opcode)
        } else {
            cycles = 1
        }

        //if r[PC] == 0x8042FDE && r[0] == 0x1 {
        //    r[0] = 1
        //}
        //if r[PC] == 0x80765E0 && CURR_INST > 5337054 {
        //if r[PC] == 0x8077110 {
        ////if r[PC] == 0x8044CCC {
        //    gba.Debugger.print(CURR_INST)
        //    os.Exit(0)
        //}

        //if CURR_INST >= 7_000_000 {
        //    os.Exit(0)
        //}

        gba.Timers.Update(uint32(cycles))
        //gba.Timers.Update(uint32(0))
        accCycles += gba.InterruptStack.checkIRQ()


        accCycles += uint32(cycles)

        if !gba.Halted {
            CURR_INST++
        }
    }

    return accCycles
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
	debugPixels := make([]byte, 1080*1080*4)

    obj := Object{}
    objs := &[128]Object{}

    for i := range 128 {
        objs[i] = obj
    }

	gba := GBA{
        Clock: 16_780_000,
        FPS: 60, // 59.7374117111
        Pixels: &pixels,
        DebugPixels: &debugPixels,
        Objects: objs,
        //VCOUNT: 126,
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

    //gba.Cache.Build(gba)
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

//func (gba *GBA) DrawFrame(exit *bool, instCount int) int {
//
//    r := &gba.Cpu.Reg.R
//
//    if gba.Paused {
//        return 0
//    }
//
//    gba.Mem.IO[VCOUNT] = 0
//
//    frameCycles := uint32(0)
//    for range 160 {
//        frameCycles = gba.UpdateScanline(frameCycles)
//    }
//
//    gba.Mem.Dispstat.SetVBlank(true)
//
//    gba.checkDmas(DMA_MODE_VBL)
//
//    //frameCycles = gba.UpdateScanline(frameCycles)
//
//    dispstat := uint32(gba.Mem.Dispstat)
//    if utils.BitEnabled(dispstat, 3) {
//        gba.InterruptStack.setIRQ(0)
//    }
//
//    for range 66 {
//        frameCycles = gba.UpdateScanline(frameCycles)
//    }
//
//    gba.Mem.Dispstat.SetVBlank(false)
//
//    frameCycles = gba.UpdateScanline(frameCycles)
//
//    c := uint32(0)
//
//    for c < 280896 {
//
//        cycles := 4
//
//        if gba.Paused {
//            return 0
//        }
//
//        if (r[PC] >= 0x400_0000 && r[PC] < 0x800_0000) || r[PC] % 2 != 0 || r[PC] >= 0xE00_0000 {
//            r := &gba.Cpu.Reg.R
//            panic(fmt.Sprintf("INVALID PC CURR %d PC %08X OPCODE %08X", CURR_INST, r[PC], gba.Mem.Read32(r[PC])))
//        }
//
//        if gba.Halted {
//            AckIntrWait(gba)
//        }
//
//        opcode := gba.Mem.Read32(r[PC])
//        gba.OpenBusOpcode = gba.Mem.Read32((r[PC] &^ 0b11) + 8)
//
//
//        if !gba.Halted {
//            cycles = gba.Cpu.Execute(opcode)
//        } else {
//            cycles = 1
//        }
//
//        //gba.VideoUpdate(uint32(cycles))
//
//        gba.Timers.Update(uint32(cycles))
//        //gba.Timers.Update(uint32(0))
//        c += gba.InterruptStack.checkIRQ()
//
//
//        c += uint32(cycles)
//
//        if !gba.Halted {
//            CURR_INST++
//        }
//
//
//    }
//
//
//    return instCount
//
//
//}
//
//func (gba *GBA) VideoUpdate(cycles uint32) {
//    last_frame_cycles := video_cycles
//    video_cycles := (video_cycles + cycles) % CYCLES_FRAME
//    frame_cycles := video_cycles
//
//    last_line_cycles := last_frame_cycles % CYCLES_SCANLINE;
//    line_cycles := frame_cycles % CYCLES_SCANLINE;
//
//    if (line_cycles >= CYCLES_HDRAW && last_line_cycles < CYCLES_HDRAW) {
//        vcount := uint32(gba.Mem.IO[VCOUNT])
//        if vcount < SCREEN_HEIGHT {
//
//            gba.scanlineGraphics(vcount)
//        }
//
//        gba.Mem.Dispstat.SetHBlank(true)
//        dispstat := uint32(gba.Mem.Dispstat)
//        if utils.BitEnabled(dispstat, 4) {
//            gba.InterruptStack.setIRQ(1)
//        }
//
//        vcount = uint32(gba.Mem.IO[VCOUNT])
//        if vcount < SCREEN_HEIGHT {
//            gba.checkDmas(DMA_MODE_HBL)
//        }
//    }
//
//    if (line_cycles < last_line_cycles) {
//
//        gba.Mem.Dispstat.SetHBlank(false)
//        gba.Mem.IO[VCOUNT] = (gba.Mem.IO[VCOUNT] + 1) % NUM_SCANLINES
//
//        switch gba.Mem.IO[VCOUNT] {
//        case 0:
//        case SCREEN_HEIGHT:
//            gba.Mem.Dispstat.SetVBlank(true)
//            gba.checkDmas(DMA_MODE_VBL)
//        case SCREEN_HEIGHT + 1:
//            dispstat := uint32(gba.Mem.Dispstat)
//            if utils.BitEnabled(dispstat, 3) {
//                gba.InterruptStack.setIRQ(0)
//            }
//        case NUM_SCANLINES - 1:
//            gba.Mem.Dispstat.SetVBlank(false)
//        }
//
//        vcount := uint32(gba.Mem.IO[VCOUNT])
//        dispstat := uint32(gba.Mem.Dispstat)
//        lyc := gba.Mem.ReadIODirect(0x05, 1)
//        match := lyc == vcount
//        gba.Mem.Dispstat.SetVCFlag(match)
//
//        vcounterIRQ := utils.BitEnabled(dispstat, 5)
//        if vcounterIRQ && match {
//            gba.InterruptStack.setIRQ(2)
//        }
//    }
//
//    if (frame_cycles < last_frame_cycles) {
//        DRAWN = true
//    }
//}
