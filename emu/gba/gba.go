package gba

import (
	"fmt"
	"os"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/apu"
	"github.com/aabalke/guac/emu/cpu"
	arm7 "github.com/aabalke/guac/emu/cpu/arm7"
	"github.com/aabalke/guac/emu/gba/cart"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/oto"
)

var _ = fmt.Sprintf
var _ = os.Exit

const (
	PC            = 15
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

type GBA struct {
	Debugger  Debugger
	Cartridge cart.Cartridge
	Cpu       *arm7.Cpu
	Mem       Memory
	PPU       PPU
	Timers    [4]Timer
	Dma       [4]DMA
	Irq       cpu.Irq
	Apu       *apu.Apu

	Paused, Muted, Save, Drawn bool
	OpenBusOpcode              uint32
	AccCycles                  uint32
	Keypad                     Keypad

	SoundCycles     uint32
	SoundCyclesMask uint32

	vsyncAddr uint32

	Pixels []byte
	Image  *ebiten.Image

	Frame uint64
}

func (gba *GBA) Update() {

	gba.AccCycles = 0

	if gba.Paused {
		return
	}

	gba.Drawn = false

	for !gba.Drawn {

		cycles := 4

		if !gba.Cpu.Halted {

			thumb := gba.Cpu.Reg.CPSR.T

			insts, ok := gba.Cpu.Execute()
			if !ok {
				panic("BAD")
			}

			// do not care about cycle accuracy right now
			if thumb {
				cycles = insts << 1
			} else {
				cycles = insts << 2
			}

		}

		gba.Tick(uint32(cycles))

		if gba.vsyncAddr != 0 && gba.Cpu.Reg.R[15] == gba.vsyncAddr {
			vblRaised := gba.Irq.IdleIrq&1 == 1
			vblHandled := gba.Irq.IF&1 != 1
			if !(vblRaised && vblHandled) {
				gba.Cpu.Halted = true
			}

			gba.Irq.IdleIrq = gba.Irq.IF
		}

		// irq has to be at end (count up tests)
		gba.Cpu.CheckIrq()

		if !gba.Cpu.Halted {
			CURR_INST++
		}
	}

	gba.Apu.Play(gba.Muted)
	gba.Frame++
}

func (gba *GBA) Tick(cycles uint32) {

	gba.SoundCycles += cycles

	if gba.SoundCycles >= gba.SoundCyclesMask {
		gba.Apu.SoundClock(gba.SoundCycles, false)
		gba.SoundCycles &= (gba.SoundCyclesMask - 1)
	}

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
		Pixels:          make([]byte, SCREEN_WIDTH*SCREEN_HEIGHT*4),
		Image:           ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT),
		Keypad:          Keypad{KEYINPUT: 0x3FF},
		Apu:             apu.NewApu(ctx, CPU_FREQ_HZ, SND_FREQUENCY, SND_SAMPLES),
		SoundCyclesMask: max(0x80, uint32(config.Conf.Gba.SoundClockUpdateCycles)),
	}

	gba.Debugger = Debugger{Gba: &gba, Version: 1}

	gba.Irq = cpu.Irq{}
	gba.Mem = NewMemory(&gba)
	gba.Cpu = arm7.NewCpu(&gba.Mem, &gba.Irq)

	gba.PPU.gba = &gba

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
	gba.Cpu.Exception(arm7.VEC_SWI, arm7.MODE_SWI)
	//gba.startupNoBios()
	gba.LoadGame(path)
	gba.SetIdleAddr()
	//InitTrig()

	startScanline := uint32(0)

	gba.Mem.BIOS_MODE = arm7.BIOS_STARTUP
	gba.Mem.IO[0x6] = uint8(startScanline)
	gba.AccCycles = CYCLES_SCANLINE*startScanline + 859

	gba.Cpu.Reg.CPSR.I = false

	return &gba
}

func (gba *GBA) startupNoBios() {
	//
	//    c := gba.Cpu
	//
	//    BANK_ID := arm7gba.BANK_ID
	//
	//	c.Irq.IME = true
	//
	//	c.Reg.R[PC] = 0x0800_0000
	//	c.Reg.CPSR = 0x0000_001F
	//	c.Reg.SPSR[BANK_ID[MODE_IRQ]] = 0x0000_0010
	//	c.Reg.R[0] = 0x0000_0CA5
	//
	//	c.Reg.R[LR] = 0x0800_0000
	//	c.Reg.LR[BANK_ID[MODE_SYS]] = 0x0800_0000
	//	c.Reg.LR[BANK_ID[MODE_USR]] = 0x0800_0000
	//	c.Reg.LR[BANK_ID[MODE_IRQ]] = 0x0800_0000
	//	c.Reg.LR[BANK_ID[MODE_SWI]] = 0x0800_0000
	//
	//	c.Reg.R[SP] = 0x0300_7F00
	//	c.Reg.SP[BANK_ID[MODE_SYS]] = 0x0300_7F00
	//	c.Reg.SP[BANK_ID[MODE_USR]] = 0x0300_7F00
	//	c.Reg.SP[BANK_ID[MODE_IRQ]] = 0x0300_7FA0
	//	c.Reg.SP[BANK_ID[MODE_SWI]] = 0x0300_7FE0
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

// RidgeX/ygba BSD3
func (gba *GBA) VideoUpdate(cycles uint32) {

	dispstat := &gba.Mem.Dispstat
	vcount := gba.Mem.IO[0x6]

	prevFrameCycles := gba.AccCycles
	gba.AccCycles += cycles //% CYCLES_FRAME
	if gba.AccCycles >= CYCLES_FRAME {
		gba.AccCycles -= CYCLES_FRAME
	}
	currFrameCycles := gba.AccCycles

	prevScanlineCycles := prevFrameCycles % CYCLES_SCANLINE
	currScanlineCycles := currFrameCycles % CYCLES_SCANLINE

	inHblank := currScanlineCycles >= CYCLES_HDRAW
	prevInHdraw := prevScanlineCycles < CYCLES_HDRAW
	if enteredHblank := inHblank && prevInHdraw; enteredHblank {

		dispstat.SetHBlank(true)
		if (*dispstat>>4)&1 != 0 {
			gba.Irq.SetIRQ(1)
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
		//gba.Apu.SoundClock(1232, false)

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
			// bios/bios.gba needs irq set on screen_height, iridion 3d needs screen_height + 1
			// I believe this is cycle related
		case SCREEN_HEIGHT + 1:
			if (*dispstat>>3)&1 != 0 {
				gba.Irq.SetIRQ(0)
			}
		case NUM_SCANLINES - 1:
			dispstat.SetVBlank(false)
		}

		match := dispstat.GetLYC() == vcount
		dispstat.SetVCFlag(match)

		if vcounterIRQ := (*dispstat>>5)&1 != 0; vcounterIRQ && match {
			gba.Irq.SetIRQ(2)
		}
	}

	if currFrameCycles < prevFrameCycles {
		gba.Drawn = true
	}
}
