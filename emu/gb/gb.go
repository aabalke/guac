package gb

import (
	"image/color"
	"log"
	"unsafe"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/gb/apu"
	"github.com/aabalke/guac/emu/gb/cartridge"
	"github.com/aabalke/guac/utils"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/oto"
)

const (
	width  = 160
	height = 144

	IRQ_VBL = 1 << 0
	IRQ_LCD = 1 << 1
	IRQ_TMR = 1 << 2
	IRQ_SER = 1 << 3
	IRQ_JPD = 1 << 4
)

type GameBoy struct {
	// Palette [][]uint8
	Palette *[4]color.Color

	DrawOptions ebiten.DrawImageOptions

	Color     bool
	bgPalette ColorPalette
	spPalette ColorPalette

	UnpackedMonoPals [3][4]uint32

	Scheduler *Scheduler

	Cartridge *cartridge.Cartridge
	Cpu       *Cpu
	MemoryBus MemoryBus

	Stat Stat
	Lcdc Lcdc

	WindowLY uint8 // windows internal line counter

	// cycles are tcycles, 1/4 mcycles
	Clock              int
	DoubleSpeedFlag    uint8
	PrepareSpeedToggle bool
	Timer              Timer

	Joypad uint8

	Image      *ebiten.Image
	Pixels     []byte
	Screen     [height][width]uint32
	spMinx     [width]int32
	bgPriority [height][width]bool
	pixelDrawn [width]bool

	Paused bool
	Muted  bool

	Apu *apu.Apu
}

type Timer struct {
	Div      uint16
	TIMA     uint8
	TMA      uint8
	Enabled  bool
	FreqBits uint8

	// for 8 cycles after overflow there is odd behavior
	// Pending Overflow 0-4 cycles after, BCycle 4-8 after
	PendingOverflow bool
	BCycle          bool
}

const (
	CPU_SPEED          = 4194304
	SND_FREQ           = 48000 // native sample rate
	CYCLES_PER_SND_GEN = CPU_SPEED / SND_FREQ
	SND_SAMPLES        = 512
)

func NewGameBoy(path string, ctx *oto.Context) *GameBoy {
	img := ebiten.NewImage(width, height)

	gb := &GameBoy{
		Image:     img,
		Cpu:       NewCpu(),
		Clock:     CPU_SPEED, // t cycle count
		Joypad:    0xFF,
		Cartridge: cartridge.NewCartridge(path, path+".save"),
		Palette:   &config.Conf.Gb.Palette,
		Scheduler: NewScheduler(),
		Apu:       apu.NewApu(ctx, CPU_SPEED, SND_FREQ, SND_SAMPLES),
	}

	// ebiten engine requires a slice, Screen is easier to edit as an array of arrays
	// instead of building an intermediate rep, pixels will just point to Screen
	gb.Pixels = unsafe.Slice((*byte)(unsafe.Pointer(&gb.Screen[0])), height*width*4)

	gb.Lcdc.gb = gb
	gb.MemoryBus.Hdma.gb = gb
	gb.bgPalette.Init()
	gb.spPalette.Init()

	if gb.Cartridge.ColorMode {
		gb.Color = true
	}

	if gb.Color {
		gb.Cpu.a = 0x11
		log.Printf("Color mode: GBC")
	} else {
		log.Printf("Color mode: DMG")
	}

	initMemory(gb)

	if config.Conf.General.Logger {
		L = NewLogger("./loggy", gb)
	}

	gb.Scheduler.schedule(EVENT_SND_SAMPLE_GEN, 0)
	gb.Scheduler.schedule(EVENT_SND_WAVE_CLOCK, 0)
	gb.Scheduler.schedule(EVENT_SND_FRAME_SEQ, 0)

	return gb
}

func (gb *GameBoy) UpdateFromConfig() {
	v := gb.MemoryBus.IO[0x47]
	gb.UnpackedMonoPals[0][0] = utils.ColorToUint32(gb.Palette[(v>>0)&3])
	gb.UnpackedMonoPals[0][1] = utils.ColorToUint32(gb.Palette[(v>>2)&3])
	gb.UnpackedMonoPals[0][2] = utils.ColorToUint32(gb.Palette[(v>>4)&3])
	gb.UnpackedMonoPals[0][3] = utils.ColorToUint32(gb.Palette[(v>>6)&3])

	v = gb.MemoryBus.IO[0x48]
	gb.UnpackedMonoPals[1][1] = utils.ColorToUint32(gb.Palette[(v>>2)&3])
	gb.UnpackedMonoPals[1][2] = utils.ColorToUint32(gb.Palette[(v>>4)&3])
	gb.UnpackedMonoPals[1][3] = utils.ColorToUint32(gb.Palette[(v>>6)&3])

	v = gb.MemoryBus.IO[0x49]
	gb.UnpackedMonoPals[2][1] = utils.ColorToUint32(gb.Palette[(v>>2)&3])
	gb.UnpackedMonoPals[2][2] = utils.ColorToUint32(gb.Palette[(v>>4)&3])
	gb.UnpackedMonoPals[2][3] = utils.ColorToUint32(gb.Palette[(v>>6)&3])
}

func (gb *GameBoy) GetSize() (int32, int32) {
	return height, width
}

func (gb *GameBoy) GetPixels() []byte {
	return gb.Pixels
}

const (
	CYCLES_PER_FRAME        = 70224
	CYCLES_PER_END_SCANLINE = CYCLES_PER_FRAME / 154
	CYCLES_PER_VBLANK       = CYCLES_PER_FRAME / 154 * 144
	CYCLES_PER_DRAW         = 80
	CYCLES_PER_HBLANK       = 80 + 172
)

func (gb *GameBoy) Update(stdFps bool) {
	if gb.Paused {
		return
	}

	gb.Scheduler.schedule(EVENT_END_FRAME, CYCLES_PER_FRAME)
	gb.Scheduler.schedule(EVENT_END_SCANLINE, CYCLES_PER_END_SCANLINE)
	gb.Scheduler.schedule(EVENT_VBK, CYCLES_PER_VBLANK)
	gb.Scheduler.schedule(EVENT_DRW, CYCLES_PER_DRAW)
	gb.Scheduler.schedule(EVENT_HBK, CYCLES_PER_HBLANK)

	for {
		nextEvent := gb.Scheduler.popNext()

		for gb.Scheduler.CurrentCycle < nextEvent.InitCycle {
			gb.Tick(gb.UpdateInterrupt())
			if gb.Cpu.Halted {
				gb.Tick(4)
			} else {
				gb.Execute()
			}
		}

		if done := gb.handleEvent(nextEvent, stdFps); done {
			return
		}
	}
}

func (gb *GameBoy) handleEvent(event ScheduledEvent, stdFps bool) bool {
	overshoot := gb.Scheduler.CurrentCycle - event.InitCycle
	gb.Scheduler.CurrentCycle = event.InitCycle

	switch event.Event {
	case EVENT_SND_SAMPLE_GEN:
		gb.Apu.SoundClock()
		gb.Scheduler.schedule(EVENT_SND_SAMPLE_GEN, CYCLES_PER_SND_GEN)

	case EVENT_VBK:
		if gb.Lcdc.Enabled {
			gb.Stat.Mode = PPU_VBLANK
			if gb.Stat.IrqVBlank {
				gb.SetIrq(IRQ_LCD)
			}

			gb.SetIrq(IRQ_VBL)
			gb.Image.WritePixels(gb.Pixels)
		}

	case EVENT_DRW:
		if gb.Lcdc.Enabled {
			gb.Stat.Mode = PPU_DRAW
		}

	case EVENT_HBK:

		if gb.Lcdc.Enabled {
			gb.drawScanline(int32(gb.MemoryBus.IO[LY]))
			gb.Stat.Mode = PPU_HBLANK
			if gb.Stat.IrqHBlank {
				gb.SetIrq(IRQ_LCD)
			}

			if gb.Color && gb.MemoryBus.Hdma.Enabled && !gb.Cpu.Halted {
				gb.MemoryBus.Hdma.Transfer(1)
			}
		}

	case EVENT_END_SCANLINE:

		if gb.Lcdc.Enabled {

			gb.MemoryBus.IO[LY]++

			gb.Stat.Match = gb.MemoryBus.IO[LY] == gb.MemoryBus.IO[LYC]
			if gb.Stat.Match && gb.Stat.IrqLyc {
				gb.SetIrq(IRQ_LCD)
			}

			if event.InitCycle+CYCLES_PER_END_SCANLINE != CYCLES_PER_FRAME {
				gb.Scheduler.scheduleAt(EVENT_END_SCANLINE, event.InitCycle+CYCLES_PER_END_SCANLINE)
			}

			if gb.MemoryBus.IO[LY] < height {
				gb.Scheduler.scheduleAt(EVENT_DRW, event.InitCycle+CYCLES_PER_DRAW)
				gb.Scheduler.scheduleAt(EVENT_HBK, event.InitCycle+CYCLES_PER_HBLANK)

				gb.Stat.Mode = PPU_OAM
				if gb.Stat.IrqOam {
					gb.SetIrq(IRQ_LCD)
				}
			}
		}

	case EVENT_END_FRAME:
		if gb.Lcdc.Enabled {
			gb.bgPriority = [height][width]bool{}
			gb.MemoryBus.IO[LY] = 0
			gb.WindowLY = 0
			gb.Stat.Match = gb.MemoryBus.IO[LY] == gb.MemoryBus.IO[LYC]
			if gb.Stat.Match && gb.Stat.IrqLyc {
				gb.SetIrq(IRQ_LCD)
			}
			gb.Stat.Mode = PPU_OAM
			if gb.Stat.IrqOam {
				gb.SetIrq(IRQ_LCD)
			}
		}

		gb.Apu.Play(gb.Muted, stdFps)
		gb.Scheduler.endFrame()
		//gb.Scheduler.CurrentCycle += overshoot
		return true

	case EVENT_SND_WAVE_CLOCK:
		ch := &gb.Apu.WaveChannel
		if !ch.ChannelEnabled {
			break
		}
		ch.WavePosition = (ch.WavePosition + 1) & 0x1F
		ch.LastReadCycle = uint32(event.InitCycle)
		if ch.WavePosition&1 == 0 {
			ch.Sample = ch.SampleByte >> 4
		} else {
			ch.ActivePeriod = ch.Period
			b := ch.Ram[ch.WavePosition>>1]
			ch.SampleByte = b
			ch.Sample = ch.SampleByte & 0xF
		}
		gb.scheduleWaveClock(event.InitCycle)

	case EVENT_SND_FRAME_SEQ:
		// I believe this is based on div, and will need to be reset based on div falling edge
		// see polling version, but confirm
		gb.Apu.ClockFrameSequencer()
		gb.Scheduler.scheduleAt(EVENT_SND_FRAME_SEQ, event.InitCycle+8192)

	default:
		panic("unsetup event")
	}

	gb.Scheduler.CurrentCycle += overshoot

	return false
}

//go:inline
func (gb *GameBoy) Tick(tCycles int64) {
	gb.Scheduler.CurrentCycle += tCycles >> gb.DoubleSpeedFlag

	if gb.Timer.Enabled {
		gb.UpdateTimers(tCycles)
	} else {
		gb.Timer.Div += uint16(tCycles)
	}

	//gb.MemoryBus.Oam.Tick(gb, tCycles)
}

func (gb *GameBoy) ToggleMute() bool {
	gb.Muted = !gb.Muted
	return gb.Muted
}

func (gb *GameBoy) TogglePause() bool {
	gb.Paused = !gb.Paused
	return gb.Paused
}

var IRQ_SRC = [...]uint16{0x40, 0x48, 0x50, 0x58, 0x60}

func (gb *GameBoy) SetIrq(bit uint8) {
	gb.Cpu.IF |= bit
}

func (gb *GameBoy) UpdateInterrupt() int64 {
	if gb.Cpu.PendingInterrupt {
		gb.Cpu.IME = true
		gb.Cpu.PendingInterrupt = false
		return 0
	}

	if !gb.Cpu.IME && !gb.Cpu.Halted {
		return 0
	}

	handling := gb.Cpu.IF & gb.Cpu.IE & 0x1F
	if noIRQ := handling == 0; noIRQ {
		return 0
	}

	if !gb.Cpu.IME && gb.Cpu.Halted {
		gb.Cpu.Halted = false
		return 20
	}

	for i := range 5 {

		if (handling>>i)&1 == 0 {
			continue
		}

		gb.Cpu.IME = false
		gb.Cpu.Halted = false

		// see mooneye/acceptance/interrupt/ie_push for stack handling
		gb.Cpu.SP--

		if gb.Cpu.SP != 0xFFFF {
			gb.Cpu.IF &^= (1 << i)
			gb.Write(gb.Cpu.SP, uint8(gb.Cpu.PC>>8))
		} else {
			gb.Cpu.PC = 0x0
			gb.Cpu.isBranching = true
			continue
		}

		gb.Cpu.SP--
		gb.Write(gb.Cpu.SP, uint8(gb.Cpu.PC))

		gb.Cpu.PC = IRQ_SRC[i]
		gb.Cpu.isBranching = true

		return 20
	}

	return 0
}

var fallingEdgeBits = [...]uint16{1 << 9, 1 << 3, 1 << 5, 1 << 7}

func (gb *GameBoy) UpdateTimers(cycles int64) {
	// have to handle edgecnt with div overflow (prev will be 0xFFC, div will be 0)
	// see oracle of ages and polemon gold for behavior
	var (
		t       = &gb.Timer
		period  = uint32(fallingEdgeBits[t.FreqBits] << 1)
		prev    = uint32(t.Div)
		next    = prev + uint32(cycles)
		edgeCnt = (next / period) - (prev / period)
	)

	t.Div = uint16(next)

	for range edgeCnt {

		t.BCycle = false

		if t.PendingOverflow {
			t.TIMA = t.TMA
			gb.SetIrq(IRQ_TMR)
			t.PendingOverflow = false
			t.BCycle = true
		}

		if overflow := t.TIMA == 0xFF; overflow {
			t.TIMA = 0
			t.PendingOverflow = true
			continue
		}

		t.TIMA++
	}
}

func (gb *GameBoy) toggleDoubleSpeed() {
	if !gb.PrepareSpeedToggle {
		return
	}

	gb.PrepareSpeedToggle = false
	gb.Cpu.Halted = false
	gb.DoubleSpeedFlag = (^gb.DoubleSpeedFlag) & 1
	gb.MemoryBus.IO[0x4D] = gb.DoubleSpeedFlag << 7
}

func (gb *GameBoy) Close() {
	gb.Muted = true
	gb.Paused = true
	gb.Apu.Close()

	if L != nil {
		L.Close()
	}
}

func (gb *GameBoy) Draw(screen *ebiten.Image) {
	var (
		sw = float64(screen.Bounds().Dx())
		sh = float64(screen.Bounds().Dy())
	)

	gb.DrawOptions.GeoM.Reset()

	scale := utils.ScaleImage(sw, sh, width, height)

	offsetX := (sw - (width * scale)) / 2
	offsetY := (sh - (height * scale)) / 2

	gb.DrawOptions.GeoM.Scale(scale, scale)
	gb.DrawOptions.GeoM.Translate(offsetX, offsetY)
	screen.DrawImage(gb.Image, &gb.DrawOptions)
}

func (gb *GameBoy) scheduleWaveClock(cycle int64) {
	if !gb.Apu.WaveChannel.ChannelEnabled {
		return
	}
	period := int64(2048-gb.Apu.WaveChannel.ActivePeriod) << 1
	gb.Scheduler.scheduleAt(EVENT_SND_WAVE_CLOCK, cycle+period)
}
