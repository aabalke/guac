package nds

import (
	"github.com/aabalke/guac/emu/gba/timer"
	"github.com/aabalke/guac/emu/nds/mem/dma"
	"github.com/aabalke/guac/emu/scheduler"
)

const (
	EVENT_VBK scheduler.Event = iota
	EVENT_HBK
	EVENT_END_SCANLINE
	EVENT_SND_SAMPLE_GEN
	EVENT_TIMER_RELOAD
	EVENT_TIMER_CONTROL
	EVENT_TIMER_OVERFLOW0
	EVENT_TIMER_OVERFLOW1
	EVENT_TIMER_OVERFLOW2
	EVENT_TIMER_OVERFLOW3
	EVENT_TIMER_OVERFLOW4
	EVENT_TIMER_OVERFLOW5
	EVENT_TIMER_OVERFLOW6
	EVENT_TIMER_OVERFLOW7

	EVENT_DMA0
	EVENT_DMA1
	EVENT_DMA2
	EVENT_DMA3
	EVENT_IRQ_SET
	EVENT_SIO
	EVENT_SND_FRAME_SEQ
)

var TimerEvents = timer.Events{
	Reload:  EVENT_TIMER_RELOAD,
	Control: EVENT_TIMER_CONTROL,
	Overflow: []scheduler.Event{
		EVENT_TIMER_OVERFLOW0,
		EVENT_TIMER_OVERFLOW1,
		EVENT_TIMER_OVERFLOW2,
		EVENT_TIMER_OVERFLOW3,
		EVENT_TIMER_OVERFLOW4,
		EVENT_TIMER_OVERFLOW5,
		EVENT_TIMER_OVERFLOW6,
		EVENT_TIMER_OVERFLOW7,
	},
}

var DMA_EVENTS = [4]scheduler.Event{EVENT_DMA0, EVENT_DMA1, EVENT_DMA2, EVENT_DMA3}

func (nds *Nds) AudioSampleEvent(late int64, arg any) {
	nds.mem.Snd.SoundClock()
	nds.Scheduler.Schedule(EVENT_SND_SAMPLE_GEN, 1, nds.CyclesPerSndGen-late, nds.AudioSampleEvent, nil)
}

func (nds *Nds) HblankEvent(late int64, arg any) {
	dispstat := &nds.mem.Dispstat

	dispstat.H = true
	if dispstat.A9HIrq {
		nds.arm9.Irq.SetIRQ(1)
	}

	if dispstat.A7HIrq {
		nds.arm7.Irq.SetIRQ(1)
	}

	if vcount := nds.mem.Vcount; vcount < SCREEN_HEIGHT {
		nds.ppu.Graphics(vcount)
		nds.CheckDmas(dma.ARM9_DMA_MODE_HBL, true)
	}
}

func (nds *Nds) ScanlineEndEvent(late int64, arg any) {
	dispstat := &nds.mem.Dispstat
	vcount := &nds.mem.Vcount

	dispstat.H = false

	*vcount++

	switch *vcount {
	case SCREEN_HEIGHT:
		if capture := &nds.ppu.Capture; capture.ActiveCapture {
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

	match := dispstat.A9LYC == *vcount
	dispstat.A9VC = match
	if dispstat.A9VCIrq && match {
		nds.arm9.Irq.SetIRQ(2)
	}

	match = dispstat.A7LYC == *vcount
	dispstat.A7VC = match
	if dispstat.A7VCIrq && match {
		nds.arm7.Irq.SetIRQ(2)
	}

	nds.Scheduler.Schedule(EVENT_END_SCANLINE, 1, CYCLES_SCANLINE-late, nds.ScanlineEndEvent, nil)
	nds.Scheduler.Schedule(EVENT_HBK, 1, CYCLES_HDRAW-late, nds.HblankEvent, nil)
}
