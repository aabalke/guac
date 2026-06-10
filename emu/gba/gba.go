package gba

import (
	"github.com/aabalke/guac/emu/cpu"
	"github.com/aabalke/guac/emu/cpu/arm"
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

type GBA struct {
	Cartridge *cart.Cartridge
	Cpu       *arm.Cpu
	Mem       *Memory
	PPU       *PPU
	Timers    [4]*Timer
	Dma       [4]DMA
	Irq       cpu.Irq
	Apu       *apu.Apu
	Scheduler *Scheduler

	Paused, Muted, Save, Drawn bool
	OpenBusOpcode              uint32
	Keypad                     Keypad

	Pixels      []byte
	Image       *ebiten.Image
	DrawOptions ebiten.DrawImageOptions

	Frame uint64
}

func (gba *GBA) Update(stdFps bool) {
	if gba.Paused {
		return
	}

	for {

		event := gba.Scheduler.peekNext()

		for gba.Scheduler.CurrentCycle < event.InitCycle {

			if gba.Cpu.Halted {

				if gba.Irq.IE&gba.Irq.IF == 0 {
					gba.Tick(1)
					continue
				}

				gba.Cpu.Halted = false
			}

			gba.Tick(gba.Cpu.Step())
		}

		for {
			next := gba.Scheduler.peekNext()
			if next == nil || next.InitCycle > gba.Scheduler.CurrentCycle {
				break
			}
			e := gba.Scheduler.popNext()
			overshoot := gba.Scheduler.CurrentCycle - e.InitCycle
			if done := e.Func(overshoot, e.Args); done {
				return
			}
		}
	}
}

func (gba *GBA) Tick(cycles int) {
	gba.Scheduler.CurrentCycle += int64(cycles)
}

func NewGBA(path string, ctx *oto.Context) *GBA {
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
	gba.Cpu = arm.NewCpu(false, gba.Mem, &gba.Irq)

	gba.Timers[0] = NewTimer(&gba, 0)
	gba.Timers[1] = NewTimer(&gba, 1)
	gba.Timers[2] = NewTimer(&gba, 2)
	gba.Timers[3] = NewTimer(&gba, 3)

	gba.Dma[0].Gba = &gba
	gba.Dma[1].Gba = &gba
	gba.Dma[2].Gba = &gba
	gba.Dma[3].Gba = &gba

	gba.Dma[0].Idx = 0
	gba.Dma[1].Idx = 1
	gba.Dma[2].Idx = 2
	gba.Dma[3].Idx = 3

	gba.LoadBios()
	gba.Cpu.Exception(arm.VEC_SWI, arm.MODE_SWI)
	//gba.startupNoBios()
	gba.LoadGame(path)

	//gba.Mem.BIOS_MODE = arm7.BIOS_STARTUP
	gba.Mem.IO[6] = 0

	gba.Cpu.Reg.CPSR.I = false

	gba.Scheduler.schedule(EVENT_SND_SAMPLE_GEN, 0, gba.AudioSampleEvent, nil)
	gba.Scheduler.schedule(EVENT_END_FRAME, 0, gba.FrameEndEvent, nil)
	gba.Scheduler.schedule(EVENT_END_SCANLINE, 0, gba.ScanlineEndEvent, nil)

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

func (gba *GBA) Draw(screen *ebiten.Image) {
	var (
		sw      = float64(screen.Bounds().Dx())
		sh      = float64(screen.Bounds().Dy())
		scale   = utils.ScaleImage(sw, sh, SCREEN_WIDTH, SCREEN_HEIGHT)
		offsetX = (sw - (SCREEN_WIDTH * scale)) / 2
		offsetY = (sh - (SCREEN_HEIGHT * scale)) / 2
	)

	gba.DrawOptions.GeoM.Reset()
	gba.DrawOptions.GeoM.Scale(scale, scale)
	gba.DrawOptions.GeoM.Translate(offsetX, offsetY)
	screen.DrawImage(gba.Image, &gba.DrawOptions)
}

func (gba *GBA) startupNoBios() {
	c := gba.Cpu

	BANK_ID := arm.BANK_ID

	c.Irq.IME = true

	c.Reg.R[PC] = 0x0800_0000
	c.Reg.CPSR.Set(0x0000_001F)
	c.Reg.SPSR[BANK_ID[arm.MODE_IRQ]].Set(0x0000_0010)
	c.Reg.R[0] = 0x0000_0CA5

	c.Reg.R[arm.LR] = 0x0800_0000
	c.Reg.LR[BANK_ID[arm.MODE_SYS]] = 0x0800_0000
	c.Reg.LR[BANK_ID[arm.MODE_USR]] = 0x0800_0000
	c.Reg.LR[BANK_ID[arm.MODE_IRQ]] = 0x0800_0000
	c.Reg.LR[BANK_ID[arm.MODE_SWI]] = 0x0800_0000

	c.Reg.R[arm.SP] = 0x0300_7F00
	c.Reg.SP[BANK_ID[arm.MODE_SYS]] = 0x0300_7F00
	c.Reg.SP[BANK_ID[arm.MODE_USR]] = 0x0300_7F00
	c.Reg.SP[BANK_ID[arm.MODE_IRQ]] = 0x0300_7FA0
	c.Reg.SP[BANK_ID[arm.MODE_SWI]] = 0x0300_7FE0
}
