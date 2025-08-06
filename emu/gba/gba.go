package gba

import (
	"fmt"
	"os"

	"github.com/aabalke/guac/emu/apu"
	"github.com/aabalke/guac/emu/gba/cart"
	"github.com/aabalke/guac/emu/gba/utils"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/oto"
)

const (
	SCREEN_WIDTH  = 240
	SCREEN_HEIGHT = 160

	NUM_SCANLINES   = SCREEN_HEIGHT + 68
	CYCLES_HDRAW    = 1006
	CYCLES_HBLANK   = 226
	CYCLES_SCANLINE = CYCLES_HDRAW + CYCLES_HBLANK
	CYCLES_VDRAW    = CYCLES_SCANLINE * SCREEN_HEIGHT
	CYCLES_VBLANK   = CYCLES_SCANLINE * 68
	CYCLES_FRAME    = CYCLES_VDRAW + CYCLES_VBLANK
)

var CURR_INST = uint64(0)

var (
    _ = os.Args
    _ = fmt.Sprintf("")
)

type GBA struct {
	Debugger  Debugger
	Cartridge cart.Cartridge
	Cpu       Cpu
	Mem       Memory
	PPU       PPU
	Timers    [4]Timer
	Dma       [4]DMA
	Irq       Irq
	Apu       *apu.Apu

	Paused, Muted, Save, Halted, Drawn bool
	OpenBusOpcode                      uint32
	AccCycles                          uint32
	Keypad                             Keypad

    vsyncAddr uint32

	Pixels []byte
	Image  *ebiten.Image

	Frame uint64
}

func (gba *GBA) SoftReset() {
	gba.exception(VEC_SWI, MODE_SWI)
}

func (gba *GBA) Update() {

	gba.AccCycles = 0

	if gba.Paused {
		return
	}

	r := &gba.Cpu.Reg.R
    _ = r[PC]
	gba.Drawn = false

	for !gba.Drawn {


		cycles := 4

		if !gba.Halted {
			cycles = gba.Cpu.Execute()
		}

		gba.Tick(uint32(cycles))

        //if gba.vsyncAddr != 0 && r[PC] == gba.vsyncAddr {
        //    vblRaised := gba.Irq.IdleIrq & 1 == 1
        //    vblHandled := gba.Irq.IF & 1 != 1
        //    if (!(vblRaised && vblHandled)) {
        //        gba.Halted = true
        //    }

        //    gba.Irq.IdleIrq = gba.Irq.IF
        //}

		// irq has to be at end (count up tests)
		gba.Irq.checkIRQ()

		if !gba.Halted {
			CURR_INST++
		}

        ////if r[PC] == 0x800_01FC {
        //    gba.Debugger.print(int(CURR_INST))
        //    os.Exit(0)
        //}
	}

	gba.Apu.Play(gba.Muted)

	gba.Frame++

	return
}

func (gba *GBA) Tick(cycles uint32) {
	gba.VideoUpdate(uint32(cycles))
	gba.UpdateTimers(uint32(cycles))
}

func NewGBA(path string, ctx *oto.Context) *GBA {

	const (
		CPU_FREQ_HZ   = 16777216
		SND_FREQUENCY = 48000 // sample rate
		SND_SAMPLES   = 512
	)

	gba := GBA{
		Pixels: make([]byte, SCREEN_WIDTH*SCREEN_HEIGHT*4),
		Image:  ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT),
		Keypad: Keypad{KEYINPUT: 0x3FF},
		Apu:    apu.NewApu(ctx, CPU_FREQ_HZ, SND_FREQUENCY, SND_SAMPLES),
	}

	gba.PPU.gba = &gba
	gba.Irq.Gba = &gba
	gba.Debugger = Debugger{Gba: &gba, Version: 1}

	gba.Mem = NewMemory(&gba)
	gba.Cpu.Gba = &gba

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

	gba.LoadBios()
	gba.SoftReset()
	gba.LoadGame(path)
    gba.SetIdleAddr()
    InitTrig()

	gba.Mem.BIOS_MODE = BIOS_STARTUP
    gba.Mem.IO[0x6] = 126
    gba.AccCycles = CYCLES_SCANLINE * 126 + 859

	return &gba
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
	gba.Apu.Close()
}

func (gba *GBA) LoadGame(path string) {
	gba.Cartridge = cart.NewCartridge(path, path+".save")
}

func (gba *GBA) toggleThumb() {

	reg := &gba.Cpu.Reg

	newFlag := reg.R[PC]&1 > 0

	reg.CPSR.SetThumb(newFlag, &gba.Cpu)

	if newFlag {
		reg.R[PC] &^= 1
		return
	}

	reg.R[PC] &^= 3
}

// RidgeX/ygba BSD3
func (gba *GBA) VideoUpdate(cycles uint32) {

	dispstat := &gba.Mem.Dispstat
	vcount := gba.Mem.IO[0x6]

	prevFrameCycles := gba.AccCycles
	gba.AccCycles += cycles //% CYCLES_FRAME
    if gba.AccCycles >= CYCLES_FRAME {
        gba.AccCycles -=CYCLES_FRAME
    }
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

		if vcount < SCREEN_HEIGHT {
            updateBackgrounds(gba, &gba.PPU.Dispcnt)
            gba.PPU.bgPriorities = gba.getBgPriority(uint32(vcount), gba.PPU.Dispcnt.Mode, &gba.PPU.Backgrounds)
            gba.PPU.objPriorities = gba.getObjPriority(uint32(vcount), &gba.PPU.Objects)
            gba.scanlineGraphics(uint32(vcount))
			gba.PPU.Backgrounds[2].BgAffineUpdate()
			gba.PPU.Backgrounds[3].BgAffineUpdate()
			gba.checkDmas(DMA_MODE_HBL)
		}
	}

	if newScanline := currScanlineCycles < prevScanlineCycles; newScanline {

        // this 1232 cycle count is estimate, should replace with actual
        gba.Apu.SoundClock(1232, false)

		dispstat.SetHBlank(false)

        vcount++
        if vcount == NUM_SCANLINES {
            vcount = 0
        }

		gba.Mem.IO[0x6] = vcount

		switch vcount {
		case 0:
			gba.PPU.Backgrounds[2].BgAffineReset()
			gba.PPU.Backgrounds[3].BgAffineReset()
		case SCREEN_HEIGHT:
			dispstat.SetVBlank(true)
			gba.checkDmas(DMA_MODE_VBL)
		//case SCREEN_HEIGHT + 1:
			if utils.BitEnabled(uint32(*dispstat), 3) {
				gba.Irq.setIRQ(0)
			}
		case NUM_SCANLINES - 1:
			dispstat.SetVBlank(false)
		}

		match := dispstat.GetLYC() == vcount
		dispstat.SetVCFlag(match)

		if vcounterIRQ := utils.BitEnabled(uint32(*dispstat), 5); vcounterIRQ && match {
			gba.Irq.setIRQ(2)
		}
	}

	if currFrameCycles < prevFrameCycles {
		gba.Drawn = true
	}
}
