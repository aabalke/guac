package gba

import (
	"github.com/aabalke/guac/emu/scheduler"
)

type RegisteredEvents struct {
	HblankVDraw  scheduler.EventIdx
	HblankVBlank scheduler.EventIdx
	ScanlineEnd  scheduler.EventIdx
	ApuTone1     scheduler.EventIdx
	ApuTone2     scheduler.EventIdx
	ApuWave      scheduler.EventIdx
	ApuNoise     scheduler.EventIdx
	FrameSeq     scheduler.EventIdx
	AudioSample  scheduler.EventIdx
}

func (gba *GBA) registerEvents() {
	gba.RegisteredEvents = RegisteredEvents{
		HblankVDraw:  gba.Scheduler.Register(gba.HblankVDrawEvent, 1),
		HblankVBlank: gba.Scheduler.Register(gba.HblankVBlankEvent, 1),
		ScanlineEnd:  gba.Scheduler.Register(gba.ScanlineEndEvent, 1),
		ApuTone1:     gba.Scheduler.Register(gba.ClockTone1, 1),
		ApuTone2:     gba.Scheduler.Register(gba.ClockTone2, 1),
		ApuWave:      gba.Scheduler.Register(gba.ClockWave, 1),
		ApuNoise:     gba.Scheduler.Register(gba.ClockNoise, 1),
		FrameSeq:     gba.Scheduler.Register(gba.ClockFrameSequencerEvent, 1),
		AudioSample:  gba.Scheduler.Register(gba.AudioSampleEvent, 1),
	}
}

func (gba *GBA) ClockTone1(late int64, arg any) {
	ch := &gba.Apu.ToneChannel1
	ch.Clock()

	if !ch.ChannelEnabled {
		return
	}
	period := int64(2048 - ch.Shadow)
	period <<= 2
	gba.Scheduler.Schedule(gba.RegisteredEvents.ApuTone1, period-late, nil)
}

func (gba *GBA) ClockTone2(late int64, arg any) {
	ch := &gba.Apu.ToneChannel2
	ch.Clock()

	if !ch.ChannelEnabled {
		return
	}
	period := int64(2048 - ch.Shadow)
	period <<= 2
	gba.Scheduler.Schedule(gba.RegisteredEvents.ApuTone2, period-late, nil)
}

func (gba *GBA) ClockWave(late int64, arg any) {
	ch := &gba.Apu.WaveChannel

	ch.Clock()
	if !ch.ChannelEnabled {
		return
	}
	period := int64(2048-ch.Shadow) << 1
	period <<= 2
	gba.Scheduler.Schedule(gba.RegisteredEvents.ApuWave, period-late, nil)
}

func (gba *GBA) ClockNoise(late int64, arg any) {
	ch := &gba.Apu.NoiseChannel
	ch.Clock()
	if !ch.ChannelEnabled {
		return
	}

	period := int64(8)

	if ch.Divider > 0 {
		period = int64(ch.Divider) << 4
	}

	period <<= int64(ch.Shift)
	period <<= 2
	gba.Scheduler.Schedule(gba.RegisteredEvents.ApuNoise, period-late, nil)
}

func (gba *GBA) ClockFrameSequencerEvent(late int64, arg any) {
	// I believe this is based on div, and will need to be reset based on div falling edge
	// see polling version, but confirm
	gba.Apu.ClockFrameSequencer()
	gba.Scheduler.Schedule(gba.RegisteredEvents.FrameSeq, CYCLES_PER_FRAME_SEQ-late, nil)
}

func (gba *GBA) AudioSampleEvent(late int64, arg any) {
	gba.Apu.SoundClock()
	gba.Scheduler.Schedule(gba.RegisteredEvents.AudioSample, gba.CyclesPerSndGen-late, nil)
}

func (gba *GBA) HblankVDrawEvent(late int64, arg any) {
	gba.Mem.Dispstat |= DISP_HBL
	if (gba.Mem.Dispstat>>4)&1 != 0 {
		gba.Irq.SetIRQ(1)
	}

	vcount := gba.Mem.IO[6]
	gba.Dma.videoDma(vcount, late)

	updateBackgrounds(gba, &gba.PPU.Dispcnt)
	gba.PPU.bgPriorities = gba.getBgPriority(uint32(vcount), gba.PPU.Dispcnt.Mode, &gba.PPU.Backgrounds)
	gba.PPU.objPriorities = gba.getObjPriority(uint32(vcount), &gba.PPU.Objects)
	gba.scanlineGraphics(uint32(vcount))
	gba.PPU.Backgrounds[2].BgAffineUpdate()
	gba.PPU.Backgrounds[3].BgAffineUpdate()
	gba.Dma.raise(DMA_MODE_HBL, late)
}

func (gba *GBA) HblankVBlankEvent(late int64, arg any) {
	gba.Mem.Dispstat |= DISP_HBL
	if (gba.Mem.Dispstat>>4)&1 != 0 {
		gba.Irq.SetIRQ(1)
	}

	vcount := gba.Mem.IO[6]
	gba.Dma.videoDma(vcount, late)
}

func (gba *GBA) ScanlineEndEvent(late int64, arg any) {
	vcount := &gba.Mem.IO[6]

	gba.Mem.Dispstat &^= DISP_HBL

	*vcount++

	switch *vcount {
	case SCREEN_HEIGHT:
		gba.Mem.Dispstat |= DISP_VBL
		gba.Dma.raise(DMA_MODE_VBL, late)
		if (gba.Mem.Dispstat>>3)&1 != 0 {
			gba.Irq.SetIRQ(0)
		}

	case 227:
		gba.Mem.Dispstat &^= DISP_VBL
	case 228:
		*vcount = 0

		gba.Mu.Lock()
		if gba.ghostOpts != nil && gba.Stats.Frame()&1 != 0 {
			gba.Ghost.WritePixels(gba.Pixels)
		} else {
			gba.Image.WritePixels(gba.Pixels)
		}
		gba.Mu.Unlock()

		gba.PPU.Backgrounds[2].BgAffineReset()
		gba.PPU.Backgrounds[3].BgAffineReset()
	}

	gba.Mem.Dispstat &^= DISP_VCF

	if match := gba.Mem.Dispstat.GetLYC() == *vcount; match {
		gba.Mem.Dispstat |= DISP_VCF

		if vcounterIRQ := (gba.Mem.Dispstat>>5)&1 != 0; vcounterIRQ {
			gba.Irq.SetIRQ(2)
		}
	}

	gba.Scheduler.Schedule(gba.RegisteredEvents.ScanlineEnd, CYCLES_SCANLINE-late, nil)
	if *vcount < SCREEN_HEIGHT {
		gba.Scheduler.Schedule(gba.RegisteredEvents.HblankVDraw, CYCLES_HDRAW-late, nil)
	} else {
		gba.Scheduler.Schedule(gba.RegisteredEvents.HblankVBlank, CYCLES_HDRAW-late, nil)
	}
}
