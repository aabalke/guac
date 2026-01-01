package nds

import (
	"fmt"
	"os"
	"sync"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/nds/cart"
	"github.com/aabalke/guac/emu/nds/cpu"
	"github.com/aabalke/guac/emu/nds/cpu/arm7"
	"github.com/aabalke/guac/emu/nds/cpu/arm9"
	"github.com/aabalke/guac/emu/nds/cpu/arm9/cp15"
	"github.com/aabalke/guac/emu/nds/debug"
	"github.com/aabalke/guac/emu/nds/mem"
	"github.com/aabalke/guac/emu/nds/mem/dma"
	"github.com/aabalke/guac/emu/nds/ppu"
	"github.com/aabalke/guac/emu/nds/snd"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/oto"
)

var (
	_ = os.Args
	_ = fmt.Sprintf("")
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

	// timer
	TIMER_CYCLE_MASK = 0b111
)

type Nds struct {
	mem       mem.Mem
	arm7      *arm7.Cpu
	arm9      *arm9.Cpu
	ppu       *ppu.PPU
	Cartridge cart.Cartridge

	Muted, Paused, Drawn  bool
	ImageTop, ImageBottom *ebiten.Image
	BtmAbs                struct{ T, B, L, R, W, H int }

	AccCycles   uint32
	TimerCycles uint8

	Frame uint64
}

func NewNds(path string, audioCtx *oto.Context) *Nds {

	nds := Nds{
		ImageTop:    ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT),
		ImageBottom: ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT),
	}

	irq7 := cpu.Irq{}
	irq9 := cpu.Irq{IsArm9: true}

	nds.ppu = ppu.NewPPU(&irq9)

	for i := range 8 {
		if i < 4 {
			nds.mem.Timers[i].IsArm9 = true
		}

		nds.mem.Timers[i].Idx = i % 4
	}

	cp15 := &cp15.Cp15{}
	cp15.Init(&nds.mem)

	nds.arm7 = arm7.NewCpu(&nds.mem, &irq7)
	nds.arm9 = arm9.NewCpu(&nds.mem, &irq9, cp15)

	s := snd.NewSnd(
		audioCtx,
		CPU_FREQ_HZ,
		SND_FREQUENCY,
		SND_SAMPLES,
	)

	nds.mem = mem.NewMemory(
		&nds.arm7.Reg.R[15],
		&nds.arm7.Halted, &nds.arm9.Halted,
		&nds.arm7.Dma, &nds.arm9.Dma,
		&irq7, &irq9,
		//nds.arm7.Jit, nds.arm9.Jit,
		nil, nds.arm9.Jit,
		&nds.Cartridge, nds.ppu, s, path+".save")

	s.Mem = &nds.mem

	nds.arm9.Dma[0].Init(0, &nds.mem, &irq9, true)
	nds.arm9.Dma[1].Init(1, &nds.mem, &irq9, true)
	nds.arm9.Dma[2].Init(2, &nds.mem, &irq9, true)
	nds.arm9.Dma[3].Init(3, &nds.mem, &irq9, true)

	nds.arm7.Dma[0].Init(0, &nds.mem, &irq7, false)
	nds.arm7.Dma[1].Init(1, &nds.mem, &irq7, false)
	nds.arm7.Dma[2].Init(2, &nds.mem, &irq7, false)
	nds.arm7.Dma[3].Init(3, &nds.mem, &irq7, false)

	nds.LoadGame(path)
	//nds.arm9.Reset()

	nds.DirtyInit()

	debug.Init("./log.csv")

	return &nds
}

var wg2 = sync.WaitGroup{}

const singleThread = !true

func (nds *Nds) Update() {

	if nds.Paused {
		return
	}

	if singleThread {
		nds.UpdateFrame()
		if nds.ppu.EngineA.Dispcnt.Is3D {
			nds.ppu.Rasterizer.Render.UpdateRender()
		}
		return
	}

	wg2.Add(2)

	go func() {
		defer wg2.Done()
		if nds.ppu.EngineA.Dispcnt.Is3D {
			nds.ppu.Rasterizer.Render.UpdateRender()
		}
	}()

	go func() {
		defer wg2.Done()
		nds.UpdateFrame()
	}()

	wg2.Wait()
}

func (nds *Nds) UpdateFrame() {

	for nds.Drawn = false; !nds.Drawn; {

		//nds.checkBadPc()

		if config.Conf.Nds.NdsJit.Enabled {

			for c := int(config.Conf.Nds.NdsJit.BatchInstA9); c >= 0; {
				c -= int(nds.StepArm9())
			}

			for c := int(config.Conf.Nds.NdsJit.BatchInstA7); c >= 0; {
				c -= int(nds.StepArm7())
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

		//debug.CURR_INST++
	}

	nds.mem.Snd.Play(nds.Muted)
	nds.Frame++
}

func (nds *Nds) StepOther() {

	if nds.TimerCycles&TIMER_CYCLE_MASK == 0 {
		nds.UpdateTimers(TIMER_CYCLE_MASK + 1)
	}

	nds.TimerCycles += 1
}

func (nds *Nds) StepArm9() uint32 {

	nds.arm9.CheckIrq()

	if nds.arm9.Halted {
		return 0xFFFF_FFFF // max to exit step
	}

	r := &nds.arm9.Reg.R

	//Log(nds, 0, 10_000, true)
	cycles, ok := nds.arm9.Execute()
	if !ok {
		fmt.Printf("ARM9 Decode Error: PC %08X CURR %d\n", r[15], debug.CURR_INST)
		os.Exit(0)
	}

	//nds.CheckGeoDmas()
	if dmaC&D == 0 {
		nds.CheckGeoDmas()
	}

	dmaC++

	if nds.ppu.Rasterizer.GeoEngine.GxStat.FifoIrq != 0 {
		nds.arm9.Irq.SetIRQ(cpu.IRQ_GEO_CMD_FIFO)
	}

	return uint32(cycles)
}

var dmaC = uint32(0)

const D = 0b1111

func (nds *Nds) StepArm7() uint32 {

	nds.arm7.CheckIrq()

	if nds.arm7.Halted {
		return 0xFFFF_FFFF // max to exit step
	}

	r7 := &nds.arm7.Reg.R

	//uhh.UpdatePcs(*r7, nds.mem.Read32(r7[15], false), uint32(nds.arm7.Reg.CPSR))
	cycles, ok := nds.arm7.Execute()
	if !ok {
		//uhh.PrintPcs()
		fmt.Printf("ARM7 Decode Error: PC %08X CURR %d\n", r7[15], debug.CURR_INST)
		os.Exit(0)
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
}

func (nds *Nds) LoadGame(path string) {
	nds.Cartridge = cart.NewCartridge(path, path+".save")
}

//temp
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

		dispstat.SetHBlank(true)
		if (dispstat.A9>>4)&1 != 0 {
			nds.arm9.Irq.SetIRQ(1)
		}

		if (dispstat.A7>>4)&1 != 0 {
			nds.arm7.Irq.SetIRQ(1)
		}

		if vcount < SCREEN_HEIGHT {

			a := &nds.ppu.EngineA
			b := &nds.ppu.EngineB
			updateBackgrounds(a)
			updateBackgrounds(b)
			a.BgPriorities = nds.getBgPriority(vcount, a.Dispcnt.Mode, &a.Backgrounds, &a.Windows)
			b.BgPriorities = nds.getBgPriority(vcount, b.Dispcnt.Mode, &b.Backgrounds, &b.Windows)

			a.ObjPriorities = nds.getObjPriority(uint32(vcount), &a.Objects, &a.Windows)
			b.ObjPriorities = nds.getObjPriority(uint32(vcount), &b.Objects, &b.Windows)

			nds.graphics(uint32(vcount))
			a.Backgrounds[2].BgAffineUpdate()
			a.Backgrounds[3].BgAffineUpdate()
			b.Backgrounds[2].BgAffineUpdate()
			b.Backgrounds[3].BgAffineUpdate()
			nds.CheckDmas(dma.ARM9_DMA_MODE_HBL, true)
		}
	}

	if newScanline := currScanlineCycles < prevScanlineCycles; newScanline {

		nds.mem.Snd.SoundClock(CYCLES_SCANLINE)

		dispstat.SetHBlank(false)

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
			dispstat.SetVBlank(true)
			nds.CheckDmas(dma.DMA_MODE_VBL, true)
			nds.CheckDmas(dma.DMA_MODE_VBL, false)

			if nds.ppu.Rasterizer.Buffers.SwapSet {
				nds.ppu.Rasterizer.Buffers.Swap()
			}

		case SCREEN_HEIGHT + 1:
			if (dispstat.A9>>3)&1 != 0 {
				nds.arm9.Irq.SetIRQ(0)
			}

			if (dispstat.A7>>3)&1 != 0 {
				nds.arm7.Irq.SetIRQ(0)
			}
		case NUM_SCANLINES - 1:
			dispstat.SetVBlank(false)
		}

		match := dispstat.GetLYC(true) == vcount
		dispstat.SetVCFlag(match, true)
		if vcntIrq := (dispstat.A9>>5)&1 != 0; vcntIrq && match {
			nds.arm9.Irq.SetIRQ(2)
		}

		match = dispstat.GetLYC(false) == vcount
		dispstat.SetVCFlag(match, false)
		if vcntIrq := (dispstat.A7>>5)&1 != 0; vcntIrq && match {
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
			if ok := nds.arm9.Dma[i].CheckMode(mode); ok {
				nds.arm9.Dma[i].Transfer()
			}
		}
		return
	}

	for i := range 4 {
		if ok := nds.arm7.Dma[i].CheckMode(mode); ok {
			nds.arm7.Dma[i].Transfer()
		}
	}
}

func (nds *Nds) CheckGeoDmas() {

	for i := range 4 {

		if !nds.arm9.Dma[i].Enabled {
			continue
		}

		if nds.arm9.Dma[i].Mode != dma.ARM9_DMA_MODE_GEO {
			continue
		}

		// never true
		//if overHalf := nds.ppu.Rasterizer.GeoEngine.GxStat.FifoEntries >= 128; overHalf {
		//    continue
		//}

		nds.arm9.Dma[i].GxTransfer()
		//nds.arm9.Dma[i].Transfer()
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
		//overflow, setIrq = t.UpdateSingle(overflow)
		if setIrq {
			if i < 4 {
				nds.arm9.Irq.SetIRQ(3 + i)
			} else {
				nds.arm7.Irq.SetIRQ(i - 1) // 3 - 4 + i (i is 4 - 8) not 0 - 4
			}
		}
	}
}
