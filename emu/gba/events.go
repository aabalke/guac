package gba

import "github.com/aabalke/guac/emu/scheduler"

const (
	EVENT_VBK scheduler.Event = iota
	EVENT_HBK
	EVENT_DRW
	EVENT_END_FRAME
	EVENT_END_SCANLINE
	EVENT_SND_SAMPLE_GEN
	EVENT_TIMER_RELOAD
	EVENT_TIMER_OVERFLOW0
	EVENT_TIMER_OVERFLOW1
	EVENT_TIMER_OVERFLOW2
	EVENT_TIMER_OVERFLOW3
	EVENT_TIMER_CONTROL
	EVENT_DMA0
	EVENT_DMA1
	EVENT_DMA2
	EVENT_DMA3
	EVENT_IRQ_SET
	EVENT_SIO
	EVENT_SND_FRAME_SEQ
	EVENT_APU_TONE1
	EVENT_APU_TONE2
	EVENT_APU_WAVE
	EVENT_APU_NOISE
)

var (
	APU_EVENTS      = [4]scheduler.Event{EVENT_APU_TONE1, EVENT_APU_TONE2, EVENT_APU_WAVE, EVENT_APU_NOISE}
	DMA_EVENTS      = [4]scheduler.Event{EVENT_DMA0, EVENT_DMA1, EVENT_DMA2, EVENT_DMA3}
	OVERFLOW_EVENTS = [4]scheduler.Event{EVENT_TIMER_OVERFLOW0, EVENT_TIMER_OVERFLOW1, EVENT_TIMER_OVERFLOW2, EVENT_TIMER_OVERFLOW3}
)

func (gba *GBA) ClockApuChannel(late int64, arg any) {
	// clock apu channels based on internal div in gb
	// fifo does not need to be clocked since clocked externally by timers

	idx := arg.(uint8)

	switch idx {
	case 0:
		gba.Apu.ToneChannel1.Clock()
	case 1:
		gba.Apu.ToneChannel2.Clock()
	case 2:
		gba.Apu.WaveChannel.Clock()
	case 3:
		gba.Apu.NoiseChannel.Clock()
	}

	gba.ScheduleApuChannel(late, idx)
}

func (gba *GBA) ScheduleApuChannel(late int64, idx uint8) {
	period := int64(0)

	switch idx {
	case 0, 1:
		ch := &gba.Apu.ToneChannel1
		if idx == 1 {
			ch = &gba.Apu.ToneChannel2
		}

		if !ch.ChannelEnabled {
			return
		}
		period = int64(2048 - ch.Shadow)

	case 2:
		ch := &gba.Apu.WaveChannel
		if !ch.ChannelEnabled {
			return
		}
		period = int64(2048-ch.Shadow) << 1

	case 3:
		ch := &gba.Apu.NoiseChannel
		if !ch.ChannelEnabled {
			return
		}

		period = 8

		if ch.Divider > 0 {
			period = int64(ch.Divider) << 4
		}

		period <<= int64(ch.Shift)
	}

	// period * 4 since gba is 4x speed
	period <<= 2

	// this will keep same pitch
	// period = (period * gba.CurrFps) / FPS
	gba.Scheduler.Schedule(APU_EVENTS[idx], 1, period-late, gba.ClockApuChannel, idx)
}

func (gba *GBA) ClockFrameSequencerEvent(late int64, arg any) {
	gba.Apu.ClockFrameSequencer()
	gba.Scheduler.Schedule(EVENT_SND_FRAME_SEQ, 1, CYCLES_PER_FRAME_SEQ-late, gba.ClockFrameSequencerEvent, nil)
}

func (gba *GBA) AudioSampleEvent(late int64, arg any) {
	gba.Apu.SoundClock()
	gba.Scheduler.Schedule(EVENT_SND_SAMPLE_GEN, 1, gba.CyclesPerSndGen-late, gba.AudioSampleEvent, nil)
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

		gba.Frame++

		gba.Mu.Lock()
		gba.Image.WritePixels(gba.Pixels)
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

	gba.Scheduler.Schedule(EVENT_END_SCANLINE, 1, CYCLES_SCANLINE-late, gba.ScanlineEndEvent, nil)

	if *vcount < SCREEN_HEIGHT {
		gba.Scheduler.Schedule(EVENT_HBK, 1, CYCLES_HDRAW-late, gba.HblankVDrawEvent, nil)
	} else {
		gba.Scheduler.Schedule(EVENT_HBK, 1, CYCLES_HDRAW-late, gba.HblankVBlankEvent, nil)
	}
}
