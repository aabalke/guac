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
	Stats           *stats.Stats
	Scheduler       *scheduler.Scheduler
	mem             *mem.Mem
	arm7            *arm7.Cpu
	arm9            *arm9.Cpu
	ppu             *ppu.PPU
	Cartridge       *cart.Cartridge
	Screen          *Screen
	dma7            [4]dma.DMA
	dma9            [4]dma.DMA
	CyclesPerSndGen int64
	Muted           bool
}

func NewNds(ctx *audio.Context, path string, muted bool) *Nds {
	nds := Nds{
		Scheduler: scheduler.NewScheduler(),
	}

	nds.mem = &mem.Mem{}

	nds.Screen = NewScreen()

	irq7 := cpu.Irq{}
	irq9 := cpu.Irq{}

	nds.ppu = ppu.NewPPU(&irq9)

	for i := range 8 {
		nds.mem.Timers[i] = timer.NewTimer(nds.Scheduler, TimerEvents, nds.OnTimerOverflow, i)
	}

	cp15 := &cp15.Cp15{}
	cp15.Init(nds.mem)

	nds.arm7 = arm7.NewCpu(config.Conf.Nds.Jit.Enabled, &nds.mem.Bus7, &irq7)
	nds.arm9 = arm9.NewCpu(config.Conf.Nds.Jit.Enabled, &nds.mem.Bus9, &irq9, cp15)

	s := snd.NewSnd(ctx, BUFFER_SIZE)

	nds.mem.InitMemory(
		&nds.arm7.Reg.R[15],
		&nds.arm7.Halted, &nds.arm9.Halted,
		&nds.dma7, &nds.dma9,
		&irq7, &irq9,
		nds.arm7.Jit, nds.arm9.Jit,
		nds.Cartridge, nds.ppu, s,
	)

	s.Mem = nds.mem

	for i := range 4 {
		nds.dma9[i].Init(i, nds.mem, &irq9, true)
		nds.dma7[i].Init(i, nds.mem, &irq7, false)
	}

	nds.Cartridge = cart.NewCartridge(
		path, nds.mem.Arm7Bios,
		&irq7, &irq9,
		&nds.dma7, &nds.dma9,
	)

	nds.mem.Cartridge = nds.Cartridge

	nds.DirectBoot()

	if config.Conf.General.Logger {
		debug.Init("./log.csv")
	}
	nds.ToggleMute(muted)

	if ctx != nil {
		nds.CyclesPerSndGen = int64(CPU_FREQ_HZ / ctx.SampleRate())
		nds.Scheduler.Schedule(EVENT_SND_SAMPLE_GEN, 1, 0, nds.AudioSampleEvent, nil)
	}

	nds.Scheduler.Schedule(EVENT_END_SCANLINE, 1, CYCLES_HBLANK, nds.ScanlineEndEvent, nil)

	return &nds
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
	nds.UpdateFrame()
	if nds.ppu.EngineA.Dispcnt.Is3D {
		nds.ppu.Rasterizer.Render.UpdateRender()
	}

	t, b := nds.GetScreens()
	nds.Screen.Mu.Lock()
	nds.Screen.Top.WritePixels(*t)
	nds.Screen.Bottom.WritePixels(*b)
	nds.Screen.Mu.Unlock()
}

func (nds *Nds) UpdateFrame() {
	var arm7 bool

	nextFrame := nds.Scheduler.CurrentCycle + CYCLES_FRAME
	for nds.Scheduler.CurrentCycle < nextFrame {

		nds.arm9.CheckIrq()

		if !nds.arm9.Halted {

			if _, ok := nds.arm9.Execute(); !ok {
				panic(fmt.Sprintf("ARM9 Decode Error: PC %08X\n", nds.arm9.Reg.R[15]))
			}

			for i := range 4 {
				if d := &nds.dma9[i]; d.Enabled && d.Mode == dma.ARM9_DMA_MODE_GEO {
					d.GxTransfer()
				}
			}

			if nds.ppu.Rasterizer.GeoEngine.GxStat.FifoIrq != 0 {
				nds.arm9.Irq.SetIRQ(cpu.IRQ_GEO_CMD_FIFO)
			}
		}

		if arm7 {
			nds.arm7.CheckIrq()

			if !nds.arm7.Halted {
				if _, ok := nds.arm7.Execute(); !ok {
					panic(fmt.Sprintf("ARM7 Decode Error: PC %08X\n", nds.arm7.Reg.R[15]))
				}
			}
		}

		nds.Tick(1)
		arm7 = !arm7
	}
}

func (nds *Nds) Tick(cycles int64) {
	nds.Scheduler.Add(cycles)
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
	nds.arm7.Jit.Close()
	nds.arm9.Jit.Close()
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
	if t.Idx < 4 {
		if t.Irq {
			nds.arm9.Irq.SetIRQ(3 + uint32(t.Idx))
		}

		if t.Idx != 3 {
			if next := nds.mem.Timers[t.Idx+1]; next.Enabled && next.Cascade {
				next.Counter++
				if next.Counter >= 0x10000 {
					next.OverflowHandle(late)
				}
			}
		}

		return
	}

	if t.Irq {
		nds.arm7.Irq.SetIRQ(3 + uint32(t.Idx-4))
	}

	if t.Idx != 7 {
		if next := nds.mem.Timers[t.Idx+1]; next.Enabled && next.Cascade {
			next.Counter++
			if next.Counter >= 0x10000 {
				next.OverflowHandle(late)
			}
		}
	}
}
