package gba

import (
	"fmt"
	"os"

	"time"

	"github.com/aabalke33/guac/emu/gba/apu"
	"github.com/aabalke33/guac/emu/gba/cart"
	"github.com/aabalke33/guac/emu/gba/utils"
	"github.com/hajimehoshi/ebiten/v2"
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


var (
    DRAWN = false

    start time.Time
    end time.Time
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
    Debugger Debugger
    Cartridge cart.Cartridge
    Cpu Cpu
    Mem Memory
	Pixels []byte
    Image *ebiten.Image
	Paused, Muted bool
    Halted bool
    ExitHalt bool
    Cycles int
    Scanline int
    Timers Timers
    Dma    [4]DMA
    IntrWait uint32
    Save bool
    Irq Irq
    GBA_LOCK bool
    OpenBusOpcode uint32
    DmaOnRefresh bool
    AccCycles uint32
    Keypad Keypad

    PPU PPU

    Cache Cache
}

func (gba *GBA) SoftReset() {
    gba.exception(VEC_SWI, MODE_SWI)
}

func (gba *GBA) Update() {

    r := &gba.Cpu.Reg.R

    gba.AccCycles = 0

    if gba.Paused {
        return
    }

    //st := time.Now()
    DRAWN = false
    for {

        cycles := 4

        if (r[PC] >= 0x400_0000 && r[PC] < 0x800_0000) || r[PC] % 2 != 0 || r[PC] >= 0xE00_0000 {
            panic(fmt.Sprintf("INVALID PC CURR %d PC %08X OPCODE %08X", CURR_INST, r[PC], gba.Mem.Read32(r[PC])))
        }

        opcode := gba.Mem.Read32(r[PC])
        gba.OpenBusOpcode = gba.Mem.Read32((r[PC] &^ 0b11) + 8)

        if !gba.Halted {
            cycles = gba.Cpu.Execute(opcode)
        }

        gba.Tick(uint32(cycles))

        // irq has to be at end (count up tests)
        gba.Irq.checkIRQ()

        if !gba.Halted {
            CURR_INST++
        }

        if DRAWN {
            break
        }
    }

    //cpuDurations[frame % 100] = time.Since(st).Milliseconds()
    //getProfilerTimes(frame)

    apu.ApuInstance.SoundBufferWrap()
    apu.ApuInstance.Play(gba.Muted)

    return
}

func (gba *GBA) Tick(cycles uint32) {
    gba.VideoUpdate(uint32(cycles))
    apu.ApuInstance.SoundClock(uint32(cycles))
    gba.Timers.Update(uint32(cycles))
}

func (gba *GBA) checkDmas(mode uint32) {
    for i := range gba.Dma {
        if gba.Dma[i].checkMode(mode) {
            gba.Dma[i].transfer(false)
        }
    }
}

func NewGBA(path string) *GBA {

    img := ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT)

	gba := GBA{
        Pixels: make([]byte, SCREEN_WIDTH*SCREEN_HEIGHT*4),
        Image: img,
        Keypad: Keypad{ KEYINPUT: 0x3FF },
    }

    gba.PPU.gba = &gba
    gba.Irq.Gba = &gba
    gba.Debugger = Debugger{Gba: &gba, Version: 1}

    gba.Mem = NewMemory(&gba)
    gba.Cpu.Gba = &gba

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

    gba.LoadBios("./emu/gba/res/bios_magia.gba")

    apu.InitApuInstance()
    apu.InitAudio()

    aveAverages = make([]int64, 0)

    gba.SoftReset()
    gba.LoadGame(path)

    //gbaConsole.Cache.BuildCache(gbaConsole)

	return &gba
}

func (gba *GBA) GetSize() (int32, int32) {
	return SCREEN_HEIGHT, SCREEN_WIDTH
}

func (gba *GBA) GetPixels() []byte {
	return gba.Pixels
}

func (gba *GBA) GetDebugPixels() []byte {
	return []byte{}
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
    apu.ApuInstance.Close()
}

func (gba *GBA) LoadGame(path string) {
    gba.GamePath = path
    gba.Cartridge = cart.NewCartridge(path, path + ".save")
}

func (gba *GBA) toggleThumb() {

    reg := &gba.Cpu.Reg

    newFlag := reg.R[PC] & 1 > 0

    reg.CPSR.SetFlag(FLAG_T, newFlag)

    if newFlag {
        reg.R[PC] &^= 1
        return
    }

    reg.R[PC] &^= 3
}

func (gba *GBA) VideoUpdate(cycles uint32) {

    dispstat := &gba.Mem.Dispstat

    prevFrameCycles := gba.AccCycles
    gba.AccCycles = (gba.AccCycles + cycles) % CYCLES_FRAME
    currFrameCycles := gba.AccCycles

    prevScanlineCycles := prevFrameCycles % CYCLES_SCANLINE
    currScanlineCycles := currFrameCycles % CYCLES_SCANLINE

    inHblank := currScanlineCycles >= CYCLES_HDRAW
    prevInHdraw := prevScanlineCycles < CYCLES_HDRAW
    if enteredHblank := inHblank && prevInHdraw; enteredHblank {

        dispstat.SetHBlank(true)
        if utils.BitEnabled(uint32(*dispstat), 4) {
            gba.Irq.setIRQ(1)
        }

        if vcount := uint32(gba.Mem.IO[VCOUNT]); vcount < SCREEN_HEIGHT {
            go gba.scanlineGraphics(vcount)
            //gba.scanlineGraphics(vcount)
            gba.checkDmas(DMA_MODE_HBL)
        }
    }

    if newScanline := currScanlineCycles < prevScanlineCycles; newScanline {

        dispstat.SetHBlank(false)

        vcount := &gba.Mem.IO[VCOUNT]
        *vcount = (*vcount + 1) % NUM_SCANLINES

        switch *vcount {
        case SCREEN_HEIGHT:
            dispstat.SetVBlank(true)
            gba.checkDmas(DMA_MODE_VBL)
        case SCREEN_HEIGHT + 1:
            if utils.BitEnabled(uint32(*dispstat), 3) {
                gba.Irq.setIRQ(0)
            }
        case NUM_SCANLINES - 1:
            dispstat.SetVBlank(false)
        }

        match := gba.Mem.IO[0x5] == *vcount
        dispstat.SetVCFlag(match)

        if vcounterIRQ := utils.BitEnabled(uint32(*dispstat), 5); vcounterIRQ && match {
            gba.Irq.setIRQ(2)
        }
    }

    if currFrameCycles < prevFrameCycles {
        DRAWN = true
    }
}
