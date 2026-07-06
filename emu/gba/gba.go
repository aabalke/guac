package gba

import (
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/gba/apu"
	"github.com/aabalke/guac/emu/gba/cart"
	"github.com/aabalke/guac/utils"
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

	CPU_SPEED          = 16777216
	SND_FREQ           = 48000 // native sample rate
	CYCLES_PER_SND_GEN = CPU_SPEED / SND_FREQ
	SND_SAMPLES        = 512
)

type GBA struct {
	Cpu               *Cpu
	Scheduler         *Scheduler
	Mem               *Memory
	Cartridge         *cart.Cartridge
	PPU               *PPU
	Timers            [4]*Timer
	Dma               [4]*Dma
	Apu               *apu.Apu
	Keypad            Key
	Irq               *Irq
	InstInjectionFunc func(op uint32)

	Frame uint64

	Pixels      []byte
	Image       *ebiten.Image
	DrawOptions ebiten.DrawImageOptions

	Paused, Muted, Save, Drawn bool

	Booted bool
}

func NewGBA(path string, ctx *oto.Context) *GBA {
	gba := &GBA{
		Pixels:    make([]byte, SCREEN_WIDTH*SCREEN_HEIGHT*4),
		Image:     ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT),
		Apu:       apu.NewApu(ctx, CPU_SPEED, SND_FREQ, SND_SAMPLES),
		Scheduler: NewScheduler(),
	}

	gba.PPU = &PPU{gba: gba}
	gba.Irq = NewIrq(gba.Scheduler)
	gba.Mem = NewMemory(gba)
	gba.Cpu = NewCpu(gba)
	gba.Keypad = Key{Irq: gba.Irq, Input: 0x3FF}
	gba.Irq.CpuIrqLine = &gba.Cpu.IrqLine
	gba.Irq.IME = true

	for i := range 4 {
		gba.Timers[i] = NewTimer(gba, i)
		gba.Dma[i] = NewDma(gba, i)
	}

	gba.Mem.LoadBios()
	gba.LoadGame(path)

	gba.Scheduler.schedule(EVENT_SND_SAMPLE_GEN, 1, 0, gba.AudioSampleEvent, nil)
	gba.Scheduler.schedule(EVENT_END_FRAME, 1, 0, gba.FrameEndEvent, nil)
	gba.Scheduler.schedule(EVENT_END_SCANLINE, 1, 0, gba.ScanlineEndEvent, nil)

	// matches nanoboy
	// gba.Mem.IO[6] = 225
	// gba.Mem.Dispstat.SetVBlank(true)
	// gba.Mem.Dispstat.SetHBlank(true)
	// gba.Scheduler.schedule(EVENT_END_SCANLINE, 1, CYCLES_HBLANK, gba.ScanlineEndEvent, nil)
	// gba.Scheduler.schedule(EVENT_END_FRAME, 1, CYCLES_FRAME-(CYCLES_SCANLINE*225), gba.FrameEndEvent, nil)

	if config.Conf.Gba.Bios.Direct {
		gba.DirectBoot()
	} else {
		gba.BiosBoot()
	}

	gba.Booted = true

	return gba
}

func (gba *GBA) Update(stdFps bool) {
	if gba.Paused {
		return
	}

	nextFrame := gba.Scheduler.CurrentCycle + CYCLES_FRAME
	for gba.Scheduler.CurrentCycle < nextFrame {
		if gba.Cpu.Halted {
			for gba.Scheduler.CurrentCycle < nextFrame && !gba.Irq.IrqAvailable {
				gba.CheckDmas()
				if gba.Irq.IrqAvailable {
					continue
				}

				gba.Tick(gba.Scheduler.GetRemaining())
			}

			if gba.Irq.IrqAvailable {
				gba.Tick(1)
				gba.Cpu.Halted = false
			}

		} else {

			if gba.InstInjectionFunc != nil {
				gba.InstInjectionFunc(gba.Cpu.Op[0])
			}

			gba.Cpu.Step()
		}
	}
}

func (gba *GBA) Tick(cycles int64) {
	gba.Scheduler.Add(cycles)
	if gba.Mem.Prefetch.Active {
		gba.Mem.Prefetch.Step(cycles)
	}
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

func (gba *GBA) DirectBoot() {
	reg := &gba.Cpu.Reg
	BANK_ID := BANK_ID

	gba.Irq.IME = true

	reg.CPSR.Set(0x1F)
	reg.SPSR[BANK_ID[MODE_IRQ]].Set(0x10)
	//reg.R[0] = 0x0CA5

	reg.R[PC] = 0x800_0000
	reg.R[LR] = 0x800_0000
	reg.LR[BANK_ID[MODE_SYS]] = 0x800_0000
	reg.LR[BANK_ID[MODE_IRQ]] = 0x800_0000
	reg.LR[BANK_ID[MODE_SWI]] = 0x800_0000

	reg.R[SP] = 0x300_7F00
	reg.SP[BANK_ID[MODE_SYS]] = 0x300_7F00
	reg.SP[BANK_ID[MODE_IRQ]] = 0x300_7FA0
	reg.SP[BANK_ID[MODE_SWI]] = 0x300_7FE0

	gba.Cpu.Op[0] = 0xF000_0000
	gba.Cpu.Op[1] = 0xF000_0000
	gba.Cpu.NonSeq = true

	gba.Mem.postflg = 1
}

func (gba *GBA) BiosBoot() {
	gba.Cpu.Exception(VEC_RESET, MODE_SYS)
}
