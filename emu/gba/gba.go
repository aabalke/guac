package gba

import (
	"context"
	"sync"
	"time"

	"github.com/aabalke/guac/common/bus"
	"github.com/aabalke/guac/common/profiler"
	"github.com/aabalke/guac/common/stats"
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/gba/apu"
	"github.com/aabalke/guac/emu/gba/cart"
	"github.com/aabalke/guac/emu/gba/cpu"
	"github.com/aabalke/guac/emu/scheduler"
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

	Cpu               *cpu.Cpu
	Scheduler         *scheduler.Scheduler
	Mem               *Memory
	Cartridge         *cart.Cartridge
	PPU               *PPU
	Timers            [4]*Timer
	Dma               *Dma
	Apu               *apu.Apu
	Irq               *Irq
	InstInjectionFunc func(op uint32)
	Keypad            Key

	Pixels []byte
	Image  *ebiten.Image

	vsyncAddr       uint32
	Save, Booted    bool
	IdleOptimize    bool
	CyclesPerSndGen int64
}

func NewGBA(ctx *audio.Context, path string, muted bool) *GBA {
	gba := &GBA{
		Pixels:       make([]byte, SCREEN_WIDTH*SCREEN_HEIGHT*4),
		Image:        ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT),
		Apu:          apu.NewApu(ctx, BUFFER_SIZE),
		Scheduler:    scheduler.NewScheduler(),
		IdleOptimize: config.Conf.Gba.IdleOptimize,
	}

	gba.PPU = &PPU{gba: gba}
	gba.Irq = NewIrq(gba, gba.Scheduler)
	gba.Mem = NewMemory(gba)
	gba.Cpu = cpu.NewCpu(gba.Mem, gba.Cycles, gba.Idle)
	gba.Keypad = Key{Irq: gba.Irq, Input: 0x3FF}
	gba.Irq.CpuIrqLine = &gba.Cpu.IrqLine
	gba.Irq.IME = true

	for i := range 4 {
		gba.Timers[i] = NewTimer(gba, i)
	}

	gba.Dma = NewDma(gba)
	gba.Mem.LoadBios()
	gba.LoadGame(path)
	gba.AddGpios()

	if ctx != nil {
		gba.CyclesPerSndGen = int64(CPU_SPEED / ctx.SampleRate())
		gba.Scheduler.Schedule(EVENT_SND_FRAME_SEQ, 1, 0, gba.ClockFrameSequencerEvent, nil)
		gba.Scheduler.Schedule(EVENT_SND_SAMPLE_GEN, 1, 0, gba.AudioSampleEvent, nil)
	}

	// matches nanoboy
	gba.Mem.IO[6] = 225
	gba.Mem.Dispstat |= DISP_HBL
	gba.Mem.Dispstat |= DISP_VBL
	gba.Scheduler.Schedule(EVENT_END_SCANLINE, 1, CYCLES_HBLANK, gba.ScanlineEndEvent, nil)

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

func (gba *GBA) LoadGame(path string) {
	gba.Cartridge = cart.NewCartridge(path, path+".save")
}

func (gba *GBA) DirectBoot() {
	reg := &gba.Cpu.Reg

	gba.Irq.IME = true

	reg.CPSR.Set(0x1F)
	reg.SPSR[cpu.ModeBank[cpu.MODE_IRQ]].Set(0x10)

	reg.R[cpu.PC] = 0x800_0000
	reg.R[cpu.LR] = 0x800_0000
	reg.LR[cpu.ModeBank[cpu.MODE_SYS]] = 0x800_0000
	reg.LR[cpu.ModeBank[cpu.MODE_IRQ]] = 0x800_0000
	reg.LR[cpu.ModeBank[cpu.MODE_SWI]] = 0x800_0000

	reg.R[cpu.SP] = 0x300_7F00
	reg.SP[cpu.ModeBank[cpu.MODE_SYS]] = 0x300_7F00
	reg.SP[cpu.ModeBank[cpu.MODE_IRQ]] = 0x300_7FA0
	reg.SP[cpu.ModeBank[cpu.MODE_SWI]] = 0x300_7FE0

	gba.Cpu.Op[0] = 0xF000_0000
	gba.Cpu.Op[1] = 0xF000_0000

	gba.Mem.postflg = 1
}

func (gba *GBA) BiosBoot() {
	gba.Cpu.Exception(cpu.VEC_RESET, cpu.MODE_SYS)
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
