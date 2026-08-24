package nds

import (
	"context"
	"fmt"
	"time"

	"github.com/aabalke/guac/common/bus"
	"github.com/aabalke/guac/common/profiler"
	"github.com/aabalke/guac/common/stats"
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/cpu"
	"github.com/aabalke/guac/emu/cpu/arm7"
	"github.com/aabalke/guac/emu/cpu/arm9"
	"github.com/aabalke/guac/emu/cpu/arm9/cp15"
	"github.com/aabalke/guac/emu/gba/timer"
	"github.com/aabalke/guac/emu/nds/cart"
	"github.com/aabalke/guac/emu/nds/debug"
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
	CYCLES_HBLANK   = 526 // or 524 need to verify
	CYCLES_SCANLINE = CYCLES_HDRAW + CYCLES_HBLANK
	CYCLES_VDRAW    = CYCLES_SCANLINE * SCREEN_HEIGHT
	CYCLES_VBLANK   = CYCLES_SCANLINE * 70 // or 71
	CYCLES_FRAME    = CYCLES_VDRAW + CYCLES_VBLANK

	// sound
	CPU_FREQ_HZ = 33554432
	BUFFER_SIZE = 40 * time.Millisecond // low power machines need at least 40ms, may need to make controllable
)

type Nds struct {
	Stats            *stats.Stats
	Scheduler        *scheduler.Scheduler
	mem              *mem.Mem
	arm7             *arm7.Cpu
	arm9             *arm9.Cpu
	irq7             *cpu.Irq
	irq9             *irq.Irq
	ppu              *ppu.PPU
	Cartridge        *cart.Cartridge
	Screen           *Screen
	dma7             [4]dma.DMA
	dma9             [4]dma.DMA
	RegisteredEvents RegisteredEvents
	CyclesPerSndGen  int64
	Muted            bool

	Arm7Cycles int64
}

func NewNds(ctx *audio.Context, path string, muted bool) *Nds {
	nds := &Nds{
		Scheduler: scheduler.NewScheduler(),
		mem:       &mem.Mem{},
		Screen:    NewScreen(),
	}

	nds.arm9 = arm9.NewCpu(&nds.mem.Bus9, cp15.NewCp15(&nds.mem.Tcm))

	nds.irq7 = &cpu.Irq{}
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

		nds.dma7[i].Init(i, &nds.mem.Bus7, nds.Scheduler, nds.irq7, false)
		nds.dma9[i].Init(i, &nds.mem.Bus9, nds.Scheduler, nds.irq9, true)
	}

	nds.arm7 = arm7.NewCpu(&nds.mem.Bus7, nds.irq7)

	nds.Cartridge = cart.NewCartridge(
		path, nds.mem.Arm7Bios,
		nds.irq7, nds.irq9,
		&nds.dma7, &nds.dma9,
	)

	nds.mem.InitMemory(
		&nds.arm7.Reg.R[15],
		&nds.arm7.Halted,
		&nds.dma7, &nds.dma9,
		nds.irq7, nds.irq9,
		nds.Cartridge, nds.ppu, snd.NewSnd(ctx, &nds.mem.Bus7, BUFFER_SIZE),
	)

	nds.DirectBoot()

	if config.Conf.General.Logger {
		debug.Init("./log.csv")
	}

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
				nds.Tick(1)
			}

			if nds.irq9.IrqAvailable {
				nds.Tick(1)
				nds.arm9.Halted = false
			}

		} else {

			if _, ok := nds.arm9.Execute(); !ok {
				panic(fmt.Sprintf("ARM9 Decode Error: PC %08X\n", nds.arm9.Reg.R[15]))
			}

			if nds.ppu.Rasterizer.GeoEngine.GxStat.FifoIrq != 0 {
				nds.irq9.SetIRQ(cpu.IRQ_GEO_CMD_FIFO)
			}

			nds.Tick(1)
		}
	}
}

func (nds *Nds) Tick(cycles int64) {
	nds.Scheduler.Add(cycles)

	for nds.Arm7Cycles < nds.Scheduler.Now()>>1 {
		nds.arm7.CheckIrq()

		if !nds.arm7.Halted {
			if _, ok := nds.arm7.Execute(); !ok {
				panic(fmt.Sprintf("ARM7 Decode Error: PC %08X\n", nds.arm7.Reg.R[15]))
			}
		}

		nds.Arm7Cycles++

		// maybe get all new arm7 events with scheduler cycles < current arm9 and catch up every time
		// would need to calc scheduler cycle based on arm7 time not arm9
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
	if debug.L != nil {
		debug.L.Close()
	}
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

	nds.arm7.Halted = false
	nds.arm9.Halted = false
}

func (nds *Nds) CheckDmas(mode uint32, arm9 bool) {
	if arm9 {
		for i := range 4 {
			if ok := nds.dma9[i].CheckMode(mode); ok {
				nds.dma9[i].Transfer()
			}
		}
		return
	}

	for i := range 4 {
		if ok := nds.dma7[i].CheckMode(mode); ok {
			nds.dma7[i].Transfer()
		}
	}
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
			nds.arm7.Irq.SetIRQ(3 + uint32(t.Idx))
		}
	}

	if next := t.Next; next != nil && next.Enabled && next.Cascade {
		next.Counter++
		if next.Counter >= 0x10000 {
			next.OverflowHandle(late)
		}
	}
}
