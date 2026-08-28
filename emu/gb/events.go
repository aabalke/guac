package gb

import (
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/scheduler"
)

type RegisteredEvents struct {
	Draw        scheduler.EventIdx
	Hblank      scheduler.EventIdx
	Scanline    scheduler.EventIdx
	ApuTone1    scheduler.EventIdx
	ApuTone2    scheduler.EventIdx
	ApuWave     scheduler.EventIdx
	ApuNoise    scheduler.EventIdx
	FrameSeq    scheduler.EventIdx
	AudioSample scheduler.EventIdx
}

func (gb *GameBoy) registerEvents() {
	gb.RegisteredEvents = RegisteredEvents{
		Draw:        gb.Scheduler.Register(gb.eventDraw, 1),
		Hblank:      gb.Scheduler.Register(gb.eventHblank, 1),
		Scanline:    gb.Scheduler.Register(gb.eventScanlineEnd, 1),
		ApuTone1:    gb.Scheduler.Register(gb.ClockTone1, 1),
		ApuTone2:    gb.Scheduler.Register(gb.ClockTone2, 1),
		ApuWave:     gb.Scheduler.Register(gb.ClockWave, 1),
		ApuNoise:    gb.Scheduler.Register(gb.ClockNoise, 1),
		FrameSeq:    gb.Scheduler.Register(gb.ClockFrameSequencerEvent, 1),
		AudioSample: gb.Scheduler.Register(gb.AudioSampleEvent, 1),
	}
}

func (gb *GameBoy) eventDraw(late int64, arg any) {
	if !gb.Lcdc.Enabled {
		return
	}

	gb.Stat.Mode = PPU_DRAW

	// if gb.Lcdc.WindowEnabled &&
	//	gb.MemoryBus.IO[WY] <= gb.MemoryBus.IO[LY] &&
	//	gb.MemoryBus.IO[WX] < 167 {

	//	gb.Scheduler.penalize(EVENT_HBK, 6+(int64(gb.MemoryBus.IO[WX])-7)&7)
	//}
}

func (gb *GameBoy) eventHblank(late int64, arg any) {
	if !gb.Lcdc.Enabled {
		return
	}

	gb.drawScanline(int32(gb.MemoryBus.IO[LY]))
	gb.Stat.Mode = PPU_HBLANK
	if gb.Stat.IrqHBlank {
		gb.SetIrq(IRQ_LCD)
	}

	if gb.Color && gb.MemoryBus.Hdma.Enabled && !gb.Cpu.Halted {
		gb.MemoryBus.Hdma.Transfer(1)
	}
}

func (gb *GameBoy) eventScanlineEnd(late int64, arg any) {
	if !gb.Lcdc.Enabled {
		return
	}

	ly := &gb.MemoryBus.IO[LY]

	*ly++

	if *ly == 154 {
		gb.bgPriority = [SCREEN_HEIGHT][SCREEN_WIDTH]bool{}
		*ly = 0
		gb.WindowLY = 0
	}

	gb.Stat.Match = *ly == gb.MemoryBus.IO[LYC]
	if gb.Stat.Match && gb.Stat.IrqLyc {
		gb.SetIrq(IRQ_LCD)
	}

	gb.Scheduler.Schedule(gb.RegisteredEvents.Scanline, CYCLES_END_SCANLINE-late, nil)

	switch {
	case *ly < SCREEN_HEIGHT:
		gb.Scheduler.Schedule(gb.RegisteredEvents.Draw, CYCLES_DRAW-late, nil)
		gb.Scheduler.Schedule(gb.RegisteredEvents.Hblank, CYCLES_HBLANK-late, nil)

		gb.Stat.Mode = PPU_OAM
		if gb.Stat.IrqOam {
			gb.SetIrq(IRQ_LCD)
		}
	case *ly == 144:
		gb.Stat.Mode = PPU_VBLANK
		if gb.Stat.IrqVBlank {
			gb.SetIrq(IRQ_LCD)
		}

		gb.SetIrq(IRQ_VBL)

		if !config.Conf.General.Headless {
			gb.Mu.Lock()
			gb.Image.WritePixels(gb.Pixels)
			gb.Mu.Unlock()
		}
	}
}

func (gb *GameBoy) ClockTone1(late int64, arg any) {
	ch := &gb.Apu.ToneChannel1
	ch.Clock()

	if !ch.ChannelEnabled {
		return
	}
	period := int64(2048 - ch.Shadow)
	gb.Scheduler.Schedule(gb.RegisteredEvents.ApuTone1, period-late, nil)
}

func (gb *GameBoy) ClockTone2(late int64, arg any) {
	ch := &gb.Apu.ToneChannel2
	ch.Clock()

	if !ch.ChannelEnabled {
		return
	}
	period := int64(2048 - ch.Shadow)
	gb.Scheduler.Schedule(gb.RegisteredEvents.ApuTone2, period-late, nil)
}

func (gb *GameBoy) ClockWave(late int64, arg any) {
	ch := &gb.Apu.WaveChannel

	ch.Clock()
	if !ch.ChannelEnabled {
		return
	}
	period := int64(2048-ch.Shadow) << 1
	gb.Scheduler.Schedule(gb.RegisteredEvents.ApuWave, period-late, nil)
}

func (gb *GameBoy) ClockNoise(late int64, arg any) {
	ch := &gb.Apu.NoiseChannel
	ch.Clock()
	if !ch.ChannelEnabled {
		return
	}

	period := int64(8)

	if ch.Divider > 0 {
		period = int64(ch.Divider) << 4
	}

	period <<= int64(ch.Shift)
	gb.Scheduler.Schedule(gb.RegisteredEvents.ApuNoise, period-late, nil)
}

func (gb *GameBoy) ClockFrameSequencerEvent(late int64, arg any) {
	// I believe this is based on div, and will need to be reset based on div falling edge
	// see polling version, but confirm
	gb.Apu.ClockFrameSequencer()
	gb.Scheduler.Schedule(gb.RegisteredEvents.FrameSeq, CYCLES_FRAME_SEQ-late, nil)
}

func (gb *GameBoy) AudioSampleEvent(late int64, arg any) {
	gb.Apu.SoundClock()
	gb.Scheduler.Schedule(gb.RegisteredEvents.AudioSample, gb.CyclesPerSndGen-late, nil)
}
