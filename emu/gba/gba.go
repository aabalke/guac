package gba

import (
	"github.com/aabalke/guac/emu/cpu"
	"github.com/aabalke/guac/emu/cpu/arm7"
	"github.com/aabalke/guac/emu/gba/apu"
	"github.com/aabalke/guac/emu/gba/cart"
	"github.com/aabalke/guac/utils"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/oto"
)

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

	CPU_SPEED          = 16777216
	SND_FREQ           = 48000 // native sample rate
	CYCLES_PER_SND_GEN = CPU_SPEED / SND_FREQ
	SND_SAMPLES        = 512
)

var CURR_INST = uint64(0)

type GBA struct {
	Cartridge *cart.Cartridge
	Cpu       *arm7.Cpu
	Mem       *Memory
	PPU       *PPU
	Timers    [4]Timer
	Dma       [4]DMA
	Irq       cpu.Irq
	Apu       *apu.Apu
	Scheduler *Scheduler

	Paused, Muted, Save, Drawn bool
	OpenBusOpcode              uint32
	AccCycles                  uint32
	Keypad                     Keypad

	vsyncAddr uint32

	Pixels      []byte
	Image       *ebiten.Image
	DrawOptions ebiten.DrawImageOptions

	Frame uint64
}

func (gba *GBA) Update(stdFps bool) {
	gba.AccCycles = 0

	if gba.Paused {
		return
	}

	gba.Scheduler.schedule(EVENT_END_FRAME, CYCLES_FRAME)
	gba.Scheduler.schedule(EVENT_END_SCANLINE, CYCLES_SCANLINE)
	//gba.Scheduler.schedule(EVENT_VBK, CYCLES_PER_VBLANK)
	gba.Scheduler.schedule(EVENT_HBK, CYCLES_HDRAW)

	for {

		nextEvent := gba.Scheduler.popNext()

		for gba.Scheduler.CurrentCycle < nextEvent.InitCycle {
			if gba.Cpu.Halted {
				gba.Tick(4)
			} else {
				thumb := gba.Cpu.Reg.CPSR.T

				insts, ok := gba.Cpu.Execute()
				if !ok {
					panic("BAD")
				}

				// do not care about cycle accuracy right now
				if thumb {
					gba.Tick(uint32(insts << 1))
				} else {
					gba.Tick(uint32(insts << 2))
				}
			}

			if gba.vsyncAddr != 0 && gba.Cpu.Reg.R[15] == gba.vsyncAddr {
				vblRaised := gba.Irq.IdleIrq&1 == 1
				vblHandled := gba.Irq.IF&1 != 1
				if !(vblRaised && vblHandled) {
					gba.Cpu.Halted = true
				}

				gba.Irq.IdleIrq = gba.Irq.IF
			}

			gba.Cpu.CheckIrq()
		}

		if !gba.Cpu.Halted {
			CURR_INST++
		}

		if done := gba.handleEvent(nextEvent, stdFps); done {
			return
		}

	}
}

func (gba *GBA) handleEvent(event ScheduledEvent, stdFps bool) bool {
	overshoot := gba.Scheduler.CurrentCycle - event.InitCycle
	gba.Scheduler.CurrentCycle = event.InitCycle

	switch event.Event {
	case EVENT_SND_SAMPLE_GEN:
		gba.Apu.SoundClock()
		gba.Scheduler.schedule(EVENT_SND_SAMPLE_GEN, CYCLES_PER_SND_GEN)

	case EVENT_HBK:
		dispstat := &gba.Mem.Dispstat
		dispstat.SetHBlank(true)
		if (*dispstat>>4)&1 != 0 {
			gba.Irq.SetIRQ(1)
		}

		if vcount := gba.Mem.IO[0x6]; vcount < SCREEN_HEIGHT {
			updateBackgrounds(gba, &gba.PPU.Dispcnt)
			gba.PPU.bgPriorities = gba.getBgPriority(uint32(vcount), gba.PPU.Dispcnt.Mode, &gba.PPU.Backgrounds)
			gba.PPU.objPriorities = gba.getObjPriority(uint32(vcount), &gba.PPU.Objects)
			gba.scanlineGraphics(uint32(vcount))
			gba.PPU.Backgrounds[2].BgAffineUpdate()
			gba.PPU.Backgrounds[3].BgAffineUpdate()
			gba.checkDmas(DMA_MODE_HBL)
		}

	case EVENT_END_SCANLINE:

		dispstat := &gba.Mem.Dispstat
		vcount := &gba.Mem.IO[0x6]

		dispstat.SetHBlank(false)

		*vcount++

		switch *vcount {
		case SCREEN_HEIGHT:
			dispstat.SetVBlank(true)
			gba.checkDmas(DMA_MODE_VBL)
		// bios/bios.gba needs irq set on screen_height, iridion 3d needs screen_height + 1
		// I believe this is cycle related
		case SCREEN_HEIGHT + 1:
			if (*dispstat>>3)&1 != 0 {
				gba.Irq.SetIRQ(0)
			}
		}

		match := dispstat.GetLYC() == *vcount
		dispstat.SetVCFlag(match)

		if vcounterIRQ := (*dispstat>>5)&1 != 0; vcounterIRQ && match {
			gba.Irq.SetIRQ(2)
		}

		if event.InitCycle+CYCLES_SCANLINE != CYCLES_FRAME {
			gba.Scheduler.scheduleAt(EVENT_END_SCANLINE, event.InitCycle+CYCLES_SCANLINE)
			gba.Scheduler.scheduleAt(EVENT_HBK, event.InitCycle+CYCLES_HDRAW)
		}

	case EVENT_END_FRAME:

		dispstat := &gba.Mem.Dispstat
		vcount := &gba.Mem.IO[0x6]

		gba.Apu.Play(gba.Muted, stdFps)
		gba.Frame++
		gba.Image.WritePixels(gba.Pixels)
		gba.Scheduler.endFrame()
		//gba.Scheduler.CurrentCycle += overshoot
		//
		*vcount = 0
		dispstat.SetVBlank(false)

		match := dispstat.GetLYC() == *vcount
		dispstat.SetVCFlag(match)

		if vcounterIRQ := (*dispstat>>5)&1 != 0; vcounterIRQ && match {
			gba.Irq.SetIRQ(2)
		}
		gba.PPU.Backgrounds[2].BgAffineReset()
		gba.PPU.Backgrounds[3].BgAffineReset()

		return true
	}

	gba.Scheduler.CurrentCycle += overshoot
	return false
}

func (gba *GBA) Tick(cycles uint32) {
	gba.Scheduler.CurrentCycle += int64(cycles)
	gba.UpdateTimers(cycles)
}

func NewGBA(path string, ctx *oto.Context) *GBA {
	const (
		SND_SAMPLES = 512
	)

	gba := GBA{
		Pixels:    make([]byte, SCREEN_WIDTH*SCREEN_HEIGHT*4),
		Image:     ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT),
		Keypad:    Keypad{KEYINPUT: 0x3FF},
		Apu:       apu.NewApu(ctx, CPU_SPEED, SND_FREQ, SND_SAMPLES),
		PPU:       &PPU{},
		Scheduler: NewScheduler(),
	}

	gba.PPU.gba = &gba

	gba.Irq = cpu.Irq{}
	gba.Mem = NewMemory(&gba)
	//gba.Cpu = arm7.NewCpu(config.Conf.Jit.Enabled, &gba.Mem, &gba.Irq)
	gba.Cpu = arm7.NewCpu(false, gba.Mem, &gba.Irq)

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

	//gba.Mem.BIOS_MODE = arm7.BIOS_STARTUP
	gba.Mem.IO[0x6] = uint8(startScanline)
	gba.AccCycles = CYCLES_SCANLINE*startScanline + 859

	gba.Cpu.Reg.CPSR.I = false

	gba.Scheduler.schedule(EVENT_SND_SAMPLE_GEN, 0)

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

func (gb *GBA) Draw(screen *ebiten.Image) {
	var (
		sw = float64(screen.Bounds().Dx())
		sh = float64(screen.Bounds().Dy())
	)

	gb.DrawOptions.GeoM.Reset()

	scale := utils.ScaleImage(sw, sh, SCREEN_WIDTH, SCREEN_HEIGHT)

	offsetX := (sw - (SCREEN_WIDTH * scale)) / 2
	offsetY := (sh - (SCREEN_HEIGHT * scale)) / 2

	gb.DrawOptions.GeoM.Scale(scale, scale)
	gb.DrawOptions.GeoM.Translate(offsetX, offsetY)
	screen.DrawImage(gb.Image, &gb.DrawOptions)
}
