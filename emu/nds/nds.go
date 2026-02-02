package nds

import (
	"fmt"
	"os"
	"sync"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/cpu"
	"github.com/aabalke/guac/emu/cpu/arm7"
	"github.com/aabalke/guac/emu/cpu/arm9"
	"github.com/aabalke/guac/emu/cpu/arm9/cp15"
	"github.com/aabalke/guac/emu/nds/cart"
	"github.com/aabalke/guac/emu/nds/debug"
	"github.com/aabalke/guac/emu/nds/mem"
	"github.com/aabalke/guac/emu/nds/mem/dma"
	"github.com/aabalke/guac/emu/nds/ppu"
	"github.com/aabalke/guac/emu/nds/snd"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/oto"
)

const (
	SCREEN_WIDTH  = 256
	SCREEN_HEIGHT = 192

	// the graphics run zt 33Mhz ( arm7 speed, so arm9 runs twice every cycle)
	NUM_SCANLINES   = SCREEN_HEIGHT + 70 // or 71
	CYCLES_HDRAW    = 1606
	CYCLES_HBLANK   = 526 // or 524 need to verify
	CYCLES_SCANLINE = CYCLES_HDRAW + CYCLES_HBLANK
	CYCLES_VDRAW    = CYCLES_SCANLINE * SCREEN_HEIGHT
	CYCLES_VBLANK   = CYCLES_SCANLINE * 70 // or 71
	CYCLES_FRAME    = CYCLES_VDRAW + CYCLES_VBLANK

	// sound
	CPU_FREQ_HZ   = 33513982
	SND_FREQUENCY = 48000 // sample rate
	SND_SAMPLES   = 1024  // 512 in gba?

	// timer and geo shouldn't be checked every inst
	// these should probably be replaced with a less lazy method
	TIMER_CYCLE_MASK = 0xF
	GEO_CYCLE_MASK   = 0xF

	SINGLE_THREAD = !true // debugging
)

var (
	RASTERIZE_WG = sync.WaitGroup{}
)

type Nds struct {
	mem       mem.Mem
	arm7      *arm7.Cpu
	arm9      *arm9.Cpu
	ppu       *ppu.PPU
	Cartridge cart.Cartridge

	dma7 [4]dma.DMA
	dma9 [4]dma.DMA

	Muted, Paused, Drawn  bool
	ImageTop, ImageBottom *ebiten.Image
	BtmAbs                struct{ T, B, L, R, W, H int }

	AccCycles   uint32
	TimerCycles uint8
	GeoCycles   uint8

	Frame uint64

	isSimd bool
}

func NewNds(path string, audioCtx *oto.Context) *Nds {

	nds := Nds{
		ImageTop:    ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT),
		ImageBottom: ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT),
	}

	irq7 := cpu.Irq{}
	irq9 := cpu.Irq{}

	nds.ppu = ppu.NewPPU(&irq9)

	for i := range 4 {
		nds.mem.Timers[i].Idx = i
		nds.mem.Timers[i].IsArm9 = true
		nds.mem.Timers[i+4].Idx = i
	}

	cp15 := &cp15.Cp15{}
	cp15.Init(&nds.mem)

	nds.arm7 = arm7.NewCpu(config.Conf.Nds.NdsJit.Enabled, &nds.mem, &irq7)
	nds.arm9 = arm9.NewCpu(config.Conf.Nds.NdsJit.Enabled, &nds.mem, &irq9, cp15)

	s := snd.NewSnd(
		audioCtx,
		CPU_FREQ_HZ,
		SND_FREQUENCY,
		SND_SAMPLES,
	)

	nds.mem = mem.NewMemory(
		&nds.arm7.Reg.R[15],
		&nds.arm7.Halted, &nds.arm9.Halted,
		&nds.dma7, &nds.dma9,
		&irq7, &irq9,
		nds.arm7.Jit, nds.arm9.Jit,
		&nds.Cartridge, nds.ppu, s, path+".save")

	s.Mem = &nds.mem

	for i := range 4 {
		nds.dma9[i].Init(i, &nds.mem, &irq9, true)
		nds.dma7[i].Init(i, &nds.mem, &irq7, false)
	}

	nds.LoadGame(path)
	//nds.arm9.Reset()
	nds.DirtyInit()
	debug.Init("./log.csv")

	return &nds
}

func (nds *Nds) Update() {

	if nds.Paused {
		return
	}

	if !nds.ppu.EngineA.Dispcnt.Is3D {
		nds.UpdateFrame()
		return
	}

	if SINGLE_THREAD {
		nds.UpdateFrame()
		nds.ppu.Rasterizer.Render.UpdateRender()
		return
	}

	RASTERIZE_WG.Add(2)

	go func() {
		defer RASTERIZE_WG.Done()
		nds.ppu.Rasterizer.Render.UpdateRender()
	}()

	go func() {
		defer RASTERIZE_WG.Done()
		nds.UpdateFrame()
	}()

	RASTERIZE_WG.Wait()
}

func (nds *Nds) UpdateFrame() {

	for nds.Drawn = false; !nds.Drawn; {

		if config.Conf.Nds.NdsJit.Enabled {

			for c := uint32(0); c < config.Conf.Nds.NdsJit.BatchInstA9; {
				c += nds.StepArm9()
			}

			for c := uint32(0); c < config.Conf.Nds.NdsJit.BatchInstA7; {
				c += nds.StepArm7()
			}

			nds.VideoUpdate(config.Conf.Nds.NdsJit.BatchInstA7)

			for c := uint32(0); c < config.Conf.Nds.NdsJit.BatchInstA7; {
				nds.StepOther()
				c++
			}

		} else {

			if arm := !nds.arm9.Reg.CPSR.T; arm {
				nds.StepArm9()
			} else {
				nds.StepArm9()
				nds.StepArm9()
			}

			if nds.arm7.Reg.CPSR.T || nds.AccCycles&1 == 0 {
				nds.StepArm7()
			}

			nds.VideoUpdate(1)
			nds.StepOther()
		}
	}

	if config.Conf.Nds.NdsJit.Enabled {
		nds.arm7.Jit.DeletePages()
		nds.arm9.Jit.DeletePages()
	}

	nds.mem.Snd.Play(nds.Muted)
	nds.Frame++
}

func (nds *Nds) StepOther() {

	if nds.TimerCycles&TIMER_CYCLE_MASK == 0 {
		nds.UpdateTimers(TIMER_CYCLE_MASK + 1)
	}

	nds.TimerCycles++
}

func (nds *Nds) StepArm9() uint32 {

	nds.arm9.CheckIrq()

	if nds.arm9.Halted {
		return 0xFFFF_FFFF // max to exit step
	}

	r := &nds.arm9.Reg.R

	cycles, ok := nds.arm9.Execute()
	if !ok {
		fmt.Printf("ARM9 Decode Error: PC %08X\n", r[15])
		os.Exit(1)
	}

	if nds.GeoCycles&GEO_CYCLE_MASK == 0 {
		nds.CheckGeoDmas()
	}

	nds.GeoCycles++

	if nds.ppu.Rasterizer.GeoEngine.GxStat.FifoIrq != 0 {
		nds.arm9.Irq.SetIRQ(cpu.IRQ_GEO_CMD_FIFO)
	}

	return uint32(cycles)
}

func (nds *Nds) StepArm7() uint32 {

	nds.arm7.CheckIrq()

	if nds.arm7.Halted {
		return 0xFFFF_FFFF // max to exit step
	}

	r7 := &nds.arm7.Reg.R

	cycles, ok := nds.arm7.Execute()
	if !ok {
		fmt.Printf("ARM7 Decode Error: PC %08X\n", r7[15])
		os.Exit(1)
	}

	return uint32(cycles)
}

func (nds *Nds) ToggleMute() bool {
	nds.Muted = !nds.Muted
	return nds.Muted
}

func (nds *Nds) TogglePause() bool {
	nds.Paused = !nds.Paused
	return nds.Paused
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
	nds.Paused = true

	debug.L.Close()
	nds.arm7.Jit.Close()
	nds.arm9.Jit.Close()
}

func (nds *Nds) LoadGame(path string) {
	nds.Cartridge = cart.NewCartridge(path, path+".save")
}

func (nds *Nds) DirtyInit() {

	nds.mem.DirtyTransfer()

	nds.arm9.Reg.R[12] = nds.Cartridge.Header.Arm9EntryAddr
	nds.arm9.Reg.R[13] = 0x3002F7C
	nds.arm9.Reg.R[14] = nds.Cartridge.Header.Arm9EntryAddr
	nds.arm9.Reg.R[15] = nds.Cartridge.Header.Arm9EntryAddr
	nds.arm9.Reg.CPSR.Set(0x000_001F)

	nds.arm7.Reg.R[12] = nds.Cartridge.Header.Arm7EntryAddr
	//nds.arm7.Reg.R[13] = 0x3002F7C
	nds.arm7.Reg.R[14] = nds.Cartridge.Header.Arm7EntryAddr
	nds.arm7.Reg.R[15] = nds.Cartridge.Header.Arm7EntryAddr
	nds.arm7.Reg.CPSR.Set(0x000_001F)

	nds.arm7.Halted = false
	nds.arm9.Halted = false
}

// RidgeX/ygba BSD3
func (nds *Nds) VideoUpdate(cycles uint32) {

	dispstat := &nds.mem.Dispstat
	vcount := nds.mem.Vcount

	prevFrameCycles := nds.AccCycles
	nds.AccCycles += cycles //% CYCLES_FRAME
	if nds.AccCycles >= CYCLES_FRAME {
		nds.AccCycles -= CYCLES_FRAME
	}
	currFrameCycles := nds.AccCycles

	prevScanlineCycles := prevFrameCycles % CYCLES_SCANLINE
	currScanlineCycles := currFrameCycles % CYCLES_SCANLINE

	inHblank := currScanlineCycles >= CYCLES_HDRAW
	prevInHdraw := prevScanlineCycles < CYCLES_HDRAW
	if enteredHblank := inHblank && prevInHdraw; enteredHblank {

		dispstat.H = true
		if dispstat.A9HIrq {
			nds.arm9.Irq.SetIRQ(1)
		}

		if dispstat.A7HIrq {
			nds.arm7.Irq.SetIRQ(1)
		}

		if vcount < SCREEN_HEIGHT {
			nds.ppu.Graphics(vcount, uint32(nds.Frame))
			nds.CheckDmas(dma.ARM9_DMA_MODE_HBL, true)
		}
	}

	if newScanline := currScanlineCycles < prevScanlineCycles; newScanline {

		nds.mem.Snd.SoundClock(CYCLES_SCANLINE)

		dispstat.H = false

		vcount++
		if vcount >= NUM_SCANLINES {
			vcount = 0
		}

		nds.mem.Vcount = vcount

		capture := &nds.ppu.Capture

		switch vcount {
		case 0:
			if capture.Enabled {
				capture.StartCapture()
			}
			nds.CheckDmas(dma.ARM9_DMA_MODE_STA, true)
			nds.ppu.EngineA.Backgrounds[2].BgAffineReset()
			nds.ppu.EngineA.Backgrounds[3].BgAffineReset()
			nds.ppu.EngineB.Backgrounds[2].BgAffineReset()
			nds.ppu.EngineB.Backgrounds[3].BgAffineReset()

			if nds.ppu.Rasterizer.GeoEngine.Disp3dCnt.RearPlaneBitmapEnabled {
				nds.ppu.Rasterizer.RearPlane.Cache()
			}

		case SCREEN_HEIGHT:
			if capture.ActiveCapture {
				capture.EndCapture()
			}

			dispstat.V = true
			nds.CheckDmas(dma.DMA_MODE_VBL, true)
			nds.CheckDmas(dma.DMA_MODE_VBL, false)

			if nds.ppu.Rasterizer.Buffers.SwapSet {
				nds.ppu.Rasterizer.Buffers.Swap()
			}

		case SCREEN_HEIGHT + 1:
			if dispstat.A9VIrq {
				nds.arm9.Irq.SetIRQ(0)
			}

			if dispstat.A7VIrq {
				nds.arm7.Irq.SetIRQ(0)
			}
		case NUM_SCANLINES - 1:
			dispstat.V = false
		}

		match := dispstat.A9LYC == vcount
		dispstat.A9VC = match
		if dispstat.A9VCIrq && match {
			nds.arm9.Irq.SetIRQ(2)
		}

		match = dispstat.A7LYC == vcount
		dispstat.A7VC = match
		if dispstat.A7VCIrq && match {
			nds.arm7.Irq.SetIRQ(2)
		}
	}

	if currFrameCycles < prevFrameCycles {
		nds.Drawn = true
	}
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

func (nds *Nds) CheckGeoDmas() {

	for i := range 4 {

		if !nds.dma9[i].Enabled {
			continue
		}

		if nds.dma9[i].Mode != dma.ARM9_DMA_MODE_GEO {
			continue
		}

		nds.dma9[i].GxTransfer()
	}
}

func (nds *Nds) UpdateTimers(cycles uint32) {

	overflow, setIrq := false, false

	for i := range uint32(8) {

		if i == 4 {
			overflow, setIrq = false, false
		}

		t := &nds.mem.Timers[i]

		if !t.Enabled {
			continue
		}

		overflow, setIrq = t.Update(overflow, cycles)
		if setIrq {
			if i < 4 {
				nds.arm9.Irq.SetIRQ(3 + i)
			} else {
				nds.arm7.Irq.SetIRQ(i - 1) // 3 - 4 + i (i is 4 - 8) not 0 - 4
			}
		}
	}
}
