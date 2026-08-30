package gba

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/aabalke/guac/common/bus"
	"github.com/aabalke/guac/common/profiler"
	"github.com/aabalke/guac/common/stats"
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/cpu/arm7"
	"github.com/aabalke/guac/emu/gba/apu"
	"github.com/aabalke/guac/emu/gba/cart"
	"github.com/aabalke/guac/emu/gba/irq"
	"github.com/aabalke/guac/emu/gba/timer"
	"github.com/aabalke/guac/emu/scheduler"
	"github.com/aabalke/guac/platform/ebiten/shader"
	"github.com/aabalke/guac/utils"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

const (
	FPS = float64(59.727500569606)
	// FPS             = 60
	SCREEN_WIDTH    = 240
	SCREEN_HEIGHT   = 160
	MAX_SCANLINE    = 228
	BUFFER_SIZE     = 40 * time.Millisecond // low power machines need at least 40ms, may need to make controllable
	CPU_SPEED       = 16777216
	CYCLES_HDRAW    = 1006
	CYCLES_HBLANK   = 226
	CYCLES_SCANLINE = CYCLES_HDRAW + CYCLES_HBLANK
	CYCLES_VDRAW    = CYCLES_SCANLINE * SCREEN_HEIGHT
	CYCLES_VBLANK   = CYCLES_SCANLINE * (MAX_SCANLINE - SCREEN_HEIGHT)
	CYCLES_FRAME    = CYCLES_VDRAW + CYCLES_VBLANK

	// dmg is 8192, needs to be 4x to match 512hz clocking
	CYCLES_PER_FRAME_SEQ = 8192 * 4
)

// this is to ensure nothing stupid has occured in development
func init() {
	if CPU_SPEED != 16777216 {
		panic("gba: cpu speed is wrong")
	}

	if CYCLES_FRAME != 280896 {
		panic("gba: cycles per frame is wrong")
	}

	if CYCLES_SCANLINE != 1232 {
		panic("gba: cycles per scanline is wrong")
	}
}

type GBA struct {
	Mu    sync.Mutex
	Stats *stats.Stats

	Cpu               *arm7.Cpu
	Scheduler         *scheduler.Scheduler
	Mem               *Memory
	Cartridge         *cart.Cartridge
	PPU               *PPU
	Timers            [4]*timer.Timer
	Dma               *Dma
	Apu               *apu.Apu
	Irq               *irq.Irq
	InstInjectionFunc func(op uint32)
	Keypad            Key

	Pixels       []byte
	Image, Ghost *ebiten.Image

	vsyncAddr       uint32
	Save, Booted    bool
	IdleOptimize    bool
	CyclesPerSndGen int64

	RegisteredEvents RegisteredEvents

	imageOpts, ghostOpts  *ebiten.DrawImageOptions
	ColorCorrectionShader *shader.ColorCorrectionShader
}

func (gba *GBA) UpdateGhosting() {
	if config.Conf.Gba.ColorCorrection.ScreenGhosting {
		gba.ghostOpts = &ebiten.DrawImageOptions{}
		gba.ghostOpts.ColorScale.ScaleAlpha(0.5)
	} else {
		gba.ghostOpts = nil
	}
}

func NewGBA(ctx *audio.Context, path string, muted bool) *GBA {
	gba := &GBA{
		Pixels:       make([]byte, SCREEN_WIDTH*SCREEN_HEIGHT*4),
		Image:        ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT),
		Ghost:        ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT),
		Apu:          apu.NewApu(ctx, BUFFER_SIZE),
		Scheduler:    scheduler.NewScheduler(),
		IdleOptimize: config.Conf.Gba.IdleOptimize,
		ColorCorrectionShader: shader.NewColorCorrectionShader(
			SCREEN_WIDTH,
			SCREEN_HEIGHT,
			&config.Conf.Gba.ColorCorrection.Type,
			&config.Conf.Gba.ColorCorrection.Strength),
	}

	gba.imageOpts = &ebiten.DrawImageOptions{}
	gba.UpdateGhosting()

	gba.PPU = &PPU{gba: gba}
	gba.Mem = NewMemory(gba)
	gba.Cpu = arm7.NewCpu(gba.Mem, gba.Cycles, gba.Idle)
	gba.Irq = irq.NewIrq(gba.Scheduler, &gba.Cpu.IrqLine)
	gba.Keypad = Key{Irq: gba.Irq, Input: 0x3FF}
	gba.Mem.Sio = NewSio(gba.Irq, gba.Scheduler)

	gba.registerEvents()

	for i := range 4 {
		gba.Timers[i] = timer.NewTimer(gba.Scheduler, gba.OnTimerOverflow, i)
		if i > 0 {
			gba.Timers[i-1].Next = gba.Timers[i]
		}
	}

	gba.Dma = NewDma(gba)
	gba.Mem.LoadBios()
	gba.Cartridge = cart.NewCartridge(path)
	gba.AddGpios()

	if ctx != nil {
		gba.CyclesPerSndGen = int64(CPU_SPEED / ctx.SampleRate())
		gba.Scheduler.Schedule(gba.RegisteredEvents.FrameSeq, 0, nil)
		gba.Scheduler.Schedule(gba.RegisteredEvents.AudioSample, 0, nil)
	}

	// matches nanoboy
	gba.Mem.IO[6] = 225
	gba.Mem.Dispstat |= DISP_HBL
	gba.Mem.Dispstat |= DISP_VBL
	gba.Scheduler.Schedule(gba.RegisteredEvents.ScanlineEnd, CYCLES_HBLANK, nil)

	gba.SetIdleAddr()
	gba.Apu.SoundBias = 0x0200

	if config.Conf.Gba.Bios.Direct {
		gba.DirectBoot()
	} else {
		gba.BiosBoot()
	}

	gba.Booted = true
	gba.Apu.ToggleMute(muted)

	return gba
}

func (gba *GBA) Run(ctx context.Context, eventBus *bus.EventBus) {
	var (
		inputCh, unSubInputCh   = eventBus.Subscribe(bus.INPUT, 64)
		muteCh, unSubMuteCh     = eventBus.Subscribe(bus.MUTE, 1)
		pauseCh, unSubPauseCh   = eventBus.Subscribe(bus.PAUSE, 1)
		setFpsCh, unSubSetFpsCh = eventBus.Subscribe(bus.SET_FPS, 1)
	)

	gba.Stats = stats.NewStats()
	go gba.Stats.RunSampler(ctx)

	defer unSubInputCh()
	defer unSubMuteCh()
	defer unSubPauseCh()
	defer unSubSetFpsCh()

	if gba.Apu != nil {
		defer gba.Apu.Close()
	}

	if gba.Apu.Ctx != nil {
		gba.CyclesPerSndGen = int64(((float64(CPU_SPEED) / float64(gba.Apu.Ctx.SampleRate())) * float64(config.Conf.General.TargetFps)) / FPS)
	}

	paused := false

	for {
		if config.Conf.Profile.Enabled {
			profiler.Profile(gba.Stats.Frame())
		}

		for drained := false; !drained; {
			select {
			case <-ctx.Done():
				return
			case e := <-inputCh:
				gba.InputHandler(
					e.Data.(bus.InputData).JustKeys,
					e.Data.(bus.InputData).Keys,
					e.Data.(bus.InputData).JustButtons,
					e.Data.(bus.InputData).Buttons,
				)
			case muted := <-muteCh:
				gba.Apu.ToggleMute(muted.Data.(bool))
			case pause := <-pauseCh:
				paused = pause.Data.(bool)
				gba.Apu.TogglePause(paused)
			case <-setFpsCh:
				if gba.Apu.Ctx != nil {
					gba.CyclesPerSndGen = int64(((float64(CPU_SPEED) / float64(gba.Apu.Ctx.SampleRate())) * float64(config.Conf.General.TargetFps)) / FPS)
				}

			default:
				drained = true
			}
		}

		if !paused {
			gba.Update()
			gba.Stats.TickFrame()
		}
	}
}

func (gba *GBA) Update() {
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

	if gba.Mem.Timings.Active {
		gba.Mem.Timings.Step(cycles)
	}

	if gba.IdleOptimize {
		gba.CheckIdleLoopOptimization()
	}
}

func (gba *GBA) CheckIdleLoopOptimization() {
	if gba.Cpu.Reg.R[15] == gba.vsyncAddr {
		prevFlag := gba.Irq.IdleIrq&1 != 0
		ifFlag := gba.Irq.IF&1 != 0
		if !prevFlag || ifFlag {
			gba.Cpu.Halted = true
		}
		gba.Irq.IdleIrq = gba.Irq.IF
	}
}

func (gba *GBA) DirectBoot() {
	reg := &gba.Cpu.Reg

	gba.Irq.IME = true

	reg.CPSR.Set(0x1F)
	reg.SPSR[arm7.ModeBank[arm7.MODE_IRQ]].Set(0x10)

	reg.R[arm7.PC] = 0x800_0000
	reg.R[arm7.LR] = 0x800_0000
	reg.LR[arm7.ModeBank[arm7.MODE_SYS]] = 0x800_0000
	reg.LR[arm7.ModeBank[arm7.MODE_IRQ]] = 0x800_0000
	reg.LR[arm7.ModeBank[arm7.MODE_SWI]] = 0x800_0000

	reg.R[arm7.SP] = 0x300_7F00
	reg.SP[arm7.ModeBank[arm7.MODE_SYS]] = 0x300_7F00
	reg.SP[arm7.ModeBank[arm7.MODE_IRQ]] = 0x300_7FA0
	reg.SP[arm7.ModeBank[arm7.MODE_SWI]] = 0x300_7FE0

	gba.Cpu.Op[0] = 0xF000_0000
	gba.Cpu.Op[1] = 0xF000_0000

	gba.Mem.postflg = 1
}

func (gba *GBA) BiosBoot() {
	gba.Cpu.Exception(arm7.VEC_RESET, arm7.MODE_SYS)
}

func (gba *GBA) Frame() uint64 {
	if gba.Stats != nil {
		return gba.Stats.Frame()
	}

	return 0
}

func (gba *GBA) FPS() float64 {
	if gba.Stats != nil {
		return gba.Stats.FPS()
	}

	return 0
}

func (gba *GBA) OnTimerOverflow(t *timer.Timer, late int64) {
	if t.Irq {
		gba.Irq.SetIRQ(3 + uint32(t.Idx))
	}

	if t.Idx < 2 && gba.Apu.Enabled {

		if aTick := (gba.Apu.SoundCntH>>10)&1 == uint16(t.Idx); aTick {

			fifo := &gba.Apu.FifoA
			fifo.Load()

			if refill := fifo.Count <= 3; refill {
				ch := gba.Dma.Chs[1]
				if ch.Enabled && ch.Mode == DMA_MODE_SPE {
					gba.Scheduler.Schedule(ch.startEvent, 2-late, nil)
				}
			}
		}

		if bTick := (gba.Apu.SoundCntH>>14)&1 == uint16(t.Idx); bTick {

			fifo := &gba.Apu.FifoB
			fifo.Load()

			if refill := fifo.Count <= 3; refill {
				ch := gba.Dma.Chs[2]
				if ch.Enabled && ch.Mode == DMA_MODE_SPE {
					gba.Scheduler.Schedule(ch.startEvent, 2-late, nil)
				}
			}
		}
	}

	if next := t.Next; next != nil && next.Enabled && next.Cascade {
		next.Counter++
		if next.Counter >= 0x10000 {
			next.OverflowHandle(late)
		}
	}
}

func (gba *GBA) Draw(dst *ebiten.Image) {
	src := gba.Image

	if gba.ghostOpts != nil {
		src.DrawImage(gba.Ghost, gba.ghostOpts)
	}

	if *gba.ColorCorrectionShader.Type != config.CLR_CORR_NONE {
		src = gba.ColorCorrectionShader.Draw(src)
	}

	var (
		canvasW  = float64(src.Bounds().Dx())
		canvasH  = float64(src.Bounds().Dy())
		rotation = config.Conf.Gba.Rotation
		radians  = float64(rotation) * math.Pi / 2
		screenW  = float64(dst.Bounds().Dx())
		screenH  = float64(dst.Bounds().Dy())
		fitW     = canvasW
		fitH     = canvasH
	)

	if rot := rotation == config.ROT_90 || rotation == config.ROT_270; rot {
		fitW, fitH = canvasH, canvasW
	}

	scale := utils.ScaleImage(screenW, screenH, fitW, fitH)

	gba.imageOpts.GeoM.Reset()
	gba.imageOpts.GeoM.Translate(-canvasW/2, -canvasH/2)
	gba.imageOpts.GeoM.Rotate(radians)
	gba.imageOpts.GeoM.Scale(scale, scale)
	gba.imageOpts.GeoM.Translate(screenW/2, screenH/2)

	gba.Mu.Lock()
	dst.DrawImage(src, gba.imageOpts)
	gba.Mu.Unlock()
}
