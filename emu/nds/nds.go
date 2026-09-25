package nds

import (
	"context"
	"time"

	"github.com/aabalke/guac/common/bus"
	"github.com/aabalke/guac/common/profiler"
	"github.com/aabalke/guac/common/stats"
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/cpu/arm7"
	"github.com/aabalke/guac/emu/cpu/arm9"
	"github.com/aabalke/guac/emu/gba/timer"
	"github.com/aabalke/guac/emu/nds/cart"
	"github.com/aabalke/guac/emu/nds/irq"
	"github.com/aabalke/guac/emu/nds/mem"
	"github.com/aabalke/guac/emu/nds/mem/dma"
	"github.com/aabalke/guac/emu/nds/ppu"
	"github.com/aabalke/guac/emu/nds/snd"
	"github.com/aabalke/guac/emu/scheduler"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

const (
	SCREEN_WIDTH  = 256
	SCREEN_HEIGHT = 192

	FPS = float64(59.8261)

	// the graphics run zt 33Mhz ( arm7 speed, so arm9 runs twice every cycle)
	NUM_SCANLINES   = SCREEN_HEIGHT + 70
	CYCLES_HDRAW    = 1606
	CYCLES_HBLANK   = 524
	CYCLES_SCANLINE = CYCLES_HDRAW + CYCLES_HBLANK
	CYCLES_VDRAW    = CYCLES_SCANLINE * SCREEN_HEIGHT
	CYCLES_VBLANK   = CYCLES_SCANLINE * 71 // or 71
	CYCLES_FRAME    = CYCLES_VDRAW + CYCLES_VBLANK

	// sound
	CPU_FREQ_HZ = 33554432
	BUFFER_SIZE = 40 * time.Millisecond // low power machines need at least 40ms, may need to make controllable
)

func init() {
	if CYCLES_SCANLINE != 2130 {
		panic("scanline cycles must be 2130")
	}

	if CYCLES_FRAME != 560190 {
		panic("frame cycles must be 560190")
	}
}

type Nds struct {
	Stats            *stats.Stats
	Scheduler        *scheduler.Scheduler
	mem              *mem.Mem
	arm7             *arm7.Cpu
	arm9             *arm9.Cpu
	irq7, irq9       *irq.Irq
	ppu              *ppu.PPU
	Cartridge        *cart.Cartridge
	Screen           *Screen
	dma7             *dma.Dma
	dma9             *dma.Dma
	RegisteredEvents RegisteredEvents
	CyclesPerSndGen  int64
	Muted            bool

	Timings7, Timings9 *Timings
}

func NewNds(ctx *audio.Context, path string, muted bool) *Nds {
	pageShift := 16
	jitConfig := arm7.JitConfig{
		AddressSpace:   0x1_0000_0000,
		PageShift:      uint32(pageShift),
		PageMask:       (1 << pageShift) - 1,
		NativePagesize: 0x10000,
		MinInstCnt:     8,
		MaxInstCnt:     64,
		BlockCnt:       4096,
		LoopThreshold:  255,
		Enabled:        false,
	}

	nds := &Nds{
		Scheduler: scheduler.NewScheduler(),
		mem:       &mem.Mem{},
		Screen:    NewScreen(),
		Timings7:  NewTimings(),
		Timings9:  NewTimings(),
	}

	nds.arm7 = arm7.NewCpu(&nds.mem.Bus7, jitConfig, nds.Cycles7, nds.Idle7)
	nds.arm9 = arm9.NewCpu(&nds.mem.Bus9, nds.Idle9, nds.Tick9, nds.Cycles9, nds.SetCyclesPerInst)
	nds.irq7 = irq.NewIrq(nds.Scheduler, &nds.arm7.IrqLine)
	nds.irq9 = irq.NewIrq(nds.Scheduler, &nds.arm9.IrqLine)
	nds.ppu = ppu.NewPPU(nds.irq9)

	nds.registerEvents()

	for i := range 4 {
		nds.mem.Timers7[i] = timer.NewTimer(nds.Scheduler, nds.OnTimerOverflow, i)
		nds.mem.Timers9[i] = timer.NewTimer(nds.Scheduler, nds.OnTimerOverflow, i)
		nds.mem.Timers9[i].IsArm9 = true
		if i > 0 {
			nds.mem.Timers7[i-1].Next = nds.mem.Timers7[i]
			nds.mem.Timers9[i-1].Next = nds.mem.Timers9[i]
		}

	}

	nds.dma7 = dma.NewDma(&nds.mem.Bus7, nds.Scheduler, nds.irq7, nds.Tick7, nds.CyclesDma7)
	nds.dma9 = dma.NewDma(&nds.mem.Bus9, nds.Scheduler, nds.irq9, nds.Tick9, nds.CyclesDma9)

	nds.dma9.IsArm9 = true

	nds.Cartridge = cart.NewCartridge(
		path, nds.mem.Arm7Bios,
		nds.irq7, nds.irq9,
		nds.dma7, nds.dma9,
	)

	nds.mem.InitMemory(
		&nds.arm7.Reg.R[15],
		&nds.arm7.Halted,
		nds.dma7, nds.dma9,
		nds.irq7, nds.irq9,
		nds.Cartridge, nds.ppu, snd.NewSnd(ctx, &nds.mem.Bus7, BUFFER_SIZE),
	)

	nds.DirectBoot()

	//if config.Conf.General.Logger {
	//	debug.Init("./log.csv")
	//}

	nds.ToggleMute(muted)

	if ctx != nil {
		nds.CyclesPerSndGen = int64(CPU_FREQ_HZ / ctx.SampleRate())
		nds.Scheduler.Schedule(nds.RegisteredEvents.AudioSample, 0, nil)
	}

	nds.Scheduler.Schedule(nds.RegisteredEvents.ScanlineEnd, CYCLES_HBLANK, nil)

	return nds
}

func (nds *Nds) Run(ctx context.Context, eventBus *bus.EventBus) {
	var (
		inputCh, unSubInputCh   = eventBus.Subscribe(bus.INPUT, 64)
		muteCh, unSubMuteCh     = eventBus.Subscribe(bus.MUTE, 1)
		pauseCh, unSubPauseCh   = eventBus.Subscribe(bus.PAUSE, 1)
		setFpsCh, unSubSetFpsCh = eventBus.Subscribe(bus.SET_FPS, 1)
	)

	nds.Stats = stats.NewStats()
	go nds.Stats.RunSampler(ctx)

	defer unSubInputCh()
	defer unSubMuteCh()
	defer unSubPauseCh()
	defer unSubSetFpsCh()
	defer nds.Close()

	if nds.mem.Snd.Ctx != nil {
		nds.CyclesPerSndGen = int64(((float64(CPU_FREQ_HZ) / float64(nds.mem.Snd.Ctx.SampleRate())) * float64(config.Conf.General.TargetFps)) / FPS)
	}

	paused := false

	for {
		if config.Conf.Profile.Enabled {
			profiler.Profile(nds.Stats.Frame())
		}

		for drained := false; !drained; {
			select {
			case <-ctx.Done():
				return
			case e := <-inputCh:
				nds.InputHandler(
					e.Data.(bus.InputData).JustKeys,
					e.Data.(bus.InputData).Keys,
					e.Data.(bus.InputData).JustButtons,
					e.Data.(bus.InputData).Buttons,
				)
			case muted := <-muteCh:
				nds.mem.Snd.ToggleMute(muted.Data.(bool))
			case pause := <-pauseCh:
				paused = pause.Data.(bool)
				nds.mem.Snd.TogglePause(paused)
			case <-setFpsCh:
				if nds.mem.Snd.Ctx != nil {
					nds.CyclesPerSndGen = int64(((float64(CPU_FREQ_HZ) / float64(nds.mem.Snd.Ctx.SampleRate())) * float64(config.Conf.General.TargetFps)) / FPS)
				}

			default:
				drained = true
			}
		}

		if !paused {
			nds.Update()
			nds.Stats.TickFrame()
		}
	}
}

func (nds *Nds) Update() {
	nextFrame := nds.Scheduler.CurrentCycle + CYCLES_FRAME
	for nds.Scheduler.CurrentCycle < nextFrame {
		if nds.arm9.Halted {

			for nds.Scheduler.CurrentCycle < nextFrame && !nds.irq9.IrqAvailable {
				// cant use get remaining - believe since arm7 uses same scheduler itd skip new arm7 events
				// will need to fix when arm7 scheduler situation is handled
				nds.Tick9(1)
			}

			if nds.irq9.IrqAvailable {
				nds.Tick9(1)
				nds.arm9.Halted = false
			}

		} else {

			nds.arm9.Step()

			nds.CheckGeoDmas()

			if nds.ppu.Rasterizer.GeoEngine.GxStat.FifoIrq != 0 {
				nds.irq9.SetIRQ(irq.IRQ_GEO_CMD_FIFO)
			}
		}
	}
}

func (nds *Nds) CheckGeoDmas() {
	for i := range 4 {
		if ch := &nds.dma9.Chs[i]; ch.Enabled && ch.Mode == dma.ARM9_DMA_MODE_GEO {
			nds.dma9.Chs[i].GxTransfer(0, 0)
		}
	}
}

func (nds *Nds) ToggleMute(muted bool) bool {
	nds.Muted = muted
	nds.mem.Snd.ToggleMute(nds.Muted)
	return nds.Muted
}

func (nds *Nds) GetScreens() (t, b *[]byte) {
	pa := &nds.ppu.EngineA.Pixels
	pb := &nds.ppu.EngineB.Pixels

	if nds.ppu.TopA {
		return pa, pb
	}

	return pb, pa
}

func (nds *Nds) Close() {
	nds.Muted = true
	nds.mem.Snd.Close()
	//if debug.L != nil {
	//	debug.L.Close()
	//}
}

func (nds *Nds) DirectBoot() {
	nds.mem.DirectBootMemory()

	nds.arm9.Reg.R[12] = nds.Cartridge.Header.Arm9EntryAddr
	nds.arm9.Reg.R[13] = 0x3002F7C
	nds.arm9.Reg.R[14] = nds.Cartridge.Header.Arm9EntryAddr
	nds.arm9.Reg.R[15] = nds.Cartridge.Header.Arm9EntryAddr
	nds.arm9.Reg.CPSR.Set(0x1F)

	nds.arm7.Reg.R[12] = nds.Cartridge.Header.Arm7EntryAddr
	//nds.arm7.Reg.R[13] = 0x3002F7C
	nds.arm7.Reg.R[14] = nds.Cartridge.Header.Arm7EntryAddr
	nds.arm7.Reg.R[15] = nds.Cartridge.Header.Arm7EntryAddr
	nds.arm7.Reg.CPSR.Set(0x1F)

	nds.arm7.ReloadPipe()
	nds.arm9.ReloadPipe()

	nds.arm7.Reload = false
	nds.arm7.Timestamp = 0

	nds.arm9.Reload = false
	nds.arm9.Timestamp = 0
	nds.arm9.InstCycles = 0
	nds.arm9.DataCycles = 0
	nds.arm9.IdleCycles = 0
}

func (nds *Nds) Frame() uint64 {
	if nds.Stats != nil {
		return nds.Stats.Frame()
	}

	return 0
}

func (nds *Nds) FPS() float64 {
	if nds.Stats != nil {
		return nds.Stats.FPS()
	}

	return 0
}

func (nds *Nds) OnTimerOverflow(t *timer.Timer, late int64) {
	if t.Irq {
		if t.IsArm9 {
			nds.irq9.SetIRQ(3 + uint32(t.Idx))
		} else {
			nds.irq7.SetIRQ(3 + uint32(t.Idx))
		}
	}

	if next := t.Next; next != nil && next.Enabled && next.Cascade {
		next.Counter++
		if next.Counter >= 0x10000 {
			next.OverflowHandle(late)
		}
	}
}
