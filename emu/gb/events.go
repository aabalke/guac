package gb

import (
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/scheduler"
)

const (
	EVENT_VBK scheduler.Event = iota
	EVENT_HBK
	EVENT_DRW
	EVENT_END_SCANLINE
	EVENT_SND_SAMPLE_GEN
	EVENT_SND_FRAME_SEQ
	EVENT_APU_TONE1
	EVENT_APU_TONE2
	EVENT_APU_WAVE
	EVENT_APU_NOISE
)

var APU_EVENTS = [4]scheduler.Event{EVENT_APU_TONE1, EVENT_APU_TONE2, EVENT_APU_WAVE, EVENT_APU_NOISE}

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
		gb.bgPriority = [height][width]bool{}
		*ly = 0
		gb.WindowLY = 0
	}

	gb.Stat.Match = *ly == gb.MemoryBus.IO[LYC]
	if gb.Stat.Match && gb.Stat.IrqLyc {
		gb.SetIrq(IRQ_LCD)
	}

	gb.Scheduler.Schedule(EVENT_END_SCANLINE, 1, CYCLES_END_SCANLINE-late, gb.eventScanlineEnd, nil)

	switch {
	case *ly < height:
		gb.Scheduler.Schedule(EVENT_DRW, 1, CYCLES_DRAW-late, gb.eventDraw, nil)
		gb.Scheduler.Schedule(EVENT_HBK, 1, CYCLES_HBLANK-late, gb.eventHblank, nil)

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
			gb.Image.WritePixels(gb.Pixels)
		}
	}
}

func (gb *GameBoy) ClockApuChannel(late int64, arg any) {
	// clock apu channels based on internal div in gb
	// fifo does not need to be clocked since clocked externally by timers

	idx := arg.(uint8)

	switch idx {
	case 0:
		gb.Apu.ToneChannel1.Clock()
	case 1:
		gb.Apu.ToneChannel2.Clock()
	case 2:
		gb.Apu.WaveChannel.Clock()
	case 3:
		gb.Apu.NoiseChannel.Clock()
	}

	gb.ScheduleApuChannel(late, idx)
}

func (gb *GameBoy) ScheduleApuChannel(late int64, idx uint8) {
	period := int64(0)

	switch idx {
	case 0, 1:
		ch := &gb.Apu.ToneChannel1
		if idx == 1 {
			ch = &gb.Apu.ToneChannel2
		}

		if !ch.ChannelEnabled {
			return
		}
		period = int64(2048 - ch.Shadow)

	case 2:
		ch := &gb.Apu.WaveChannel
		if !ch.ChannelEnabled {
			return
		}
		period = int64(2048-ch.Shadow) << 1

	case 3:
		ch := &gb.Apu.NoiseChannel
		if !ch.ChannelEnabled {
			return
		}

		period = 8

		if ch.Divider > 0 {
			period = int64(ch.Divider) << 4
		}

		period <<= int64(ch.Shift)
	}

	// this will keep same pitch
	period = (period * gb.CurrFps) / FPS
	gb.Scheduler.Schedule(APU_EVENTS[idx], 1, period-late, gb.ClockApuChannel, idx)
}

func (gb *GameBoy) ClockFrameSequencerEvent(late int64, arg any) {
	// I believe this is based on div, and will need to be reset based on div falling edge
	// see polling version, but confirm
	gb.Apu.ClockFrameSequencer()
	gb.Scheduler.Schedule(EVENT_SND_FRAME_SEQ, 1, CYCLES_FRAME_SEQ-late, gb.ClockFrameSequencerEvent, nil)
}

func (gb *GameBoy) AudioSampleEvent(late int64, arg any) {
	gb.Apu.SoundClock()
	gb.Scheduler.Schedule(EVENT_SND_SAMPLE_GEN, 1, gb.CyclesPerSndGen-late, gb.AudioSampleEvent, nil)
}
