package nds

import (
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/nds/mem/dma"
	"github.com/aabalke/guac/emu/scheduler"
)

type RegisteredEvents struct {
	Hblank      scheduler.EventIdx
	ScanlineEnd scheduler.EventIdx
	AudioSample scheduler.EventIdx
}

func (nds *Nds) registerEvents() {
	nds.RegisteredEvents = RegisteredEvents{
		Hblank:      nds.Scheduler.Register(nds.HblankEvent, 1),
		ScanlineEnd: nds.Scheduler.Register(nds.ScanlineEndEvent, 1),
		AudioSample: nds.Scheduler.Register(nds.AudioSampleEvent, 1),
	}
}

func (nds *Nds) AudioSampleEvent(late int64, arg any) {
	nds.mem.Snd.SoundClock()
	nds.Scheduler.Schedule(nds.RegisteredEvents.AudioSample, nds.CyclesPerSndGen-late, nil)
}

func (nds *Nds) HblankEvent(late int64, arg any) {
	d7 := &nds.mem.Dispstat7
	d9 := &nds.mem.Dispstat9

	d7.H = true
	d9.H = true
	if d7.HIrq {
		nds.irq7.SetIRQ(1)
	}
	if d9.HIrq {
		nds.irq9.SetIRQ(1)
	}

	if vcount := nds.mem.Vcount; vcount < SCREEN_HEIGHT {
		nds.ppu.Graphics(vcount)
		nds.CheckDmas(dma.ARM9_DMA_MODE_HBL, true)
	}
}

func (nds *Nds) ScanlineEndEvent(late int64, arg any) {
	d7 := &nds.mem.Dispstat7
	d9 := &nds.mem.Dispstat9
	vcount := &nds.mem.Vcount

	d7.H = false
	d9.H = false

	*vcount++

	switch *vcount {
	case SCREEN_HEIGHT:
		if capture := &nds.ppu.Capture; capture.ActiveCapture {
			capture.EndCapture()
		}

		d7.V = true
		d9.V = true
		nds.CheckDmas(dma.DMA_MODE_VBL, true)
		nds.CheckDmas(dma.DMA_MODE_VBL, false)

		if nds.ppu.Rasterizer.Buffers.SwapSet {
			nds.ppu.Rasterizer.Buffers.Swap()
		}

		if !config.Conf.General.Headless {
			if nds.ppu.EngineA.Dispcnt.Is3D {
				nds.ppu.Rasterizer.Render.UpdateRender()
			}

			t, b := nds.GetScreens()
			nds.Screen.Mu.Lock()
			if nds.Screen.ghostOpts != nil && nds.Stats.Frame()&1 != 0 {
				nds.Screen.TopGhost.WritePixels(*t)
				nds.Screen.BottomGhost.WritePixels(*b)
			} else {
				nds.Screen.Top.WritePixels(*t)
				nds.Screen.Bottom.WritePixels(*b)
			}
			nds.Screen.Mu.Unlock()
		}

	case SCREEN_HEIGHT + 1:
		if d7.VIrq {
			nds.irq7.SetIRQ(0)
		}
		if d9.VIrq {
			nds.irq9.SetIRQ(0)
		}

	case NUM_SCANLINES - 1:
		d7.V = false
		d9.V = false
	case NUM_SCANLINES:
		*vcount = 0

		if capture := &nds.ppu.Capture; capture.Enabled {
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
	}

	match := d7.LYC == *vcount
	d7.VC = match
	if d7.VCIrq && match {
		nds.irq7.SetIRQ(2)
	}

	match = d9.LYC == *vcount
	d9.VC = match
	if d9.VCIrq && match {
		nds.irq9.SetIRQ(2)
	}

	nds.Scheduler.Schedule(nds.RegisteredEvents.ScanlineEnd, CYCLES_SCANLINE-late, nil)
	nds.Scheduler.Schedule(nds.RegisteredEvents.Hblank, CYCLES_HDRAW-late, nil)
}
