package gb

import (
	"image/color"
	"time"
	"unsafe"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/gb/apu"
	"github.com/aabalke/guac/emu/gb/cartridge"
	"github.com/aabalke/guac/emu/scheduler"
	"github.com/aabalke/guac/utils"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

const (
	width  = 160
	height = 144

	IRQ_VBL = 1 << 0
	IRQ_LCD = 1 << 1
	IRQ_TMR = 1 << 2
	IRQ_SER = 1 << 3
	IRQ_JPD = 1 << 4

	CPU_SPEED           = 4194304
	BUFFER_SIZE         = 20 * time.Millisecond
	CYCLES_FRAME        = 70224
	CYCLES_END_SCANLINE = CYCLES_FRAME / 154
	CYCLES_VBLANK       = CYCLES_FRAME / 154 * 144
	CYCLES_DRAW         = 80
	CYCLES_HBLANK       = 80 + 172
	CYCLES_FRAME_SEQ    = 8192

	FPS = 60
)

type GameBoy struct {
	// Palette [][]uint8
	Palette *[4]color.Color

	DrawOptions ebiten.DrawImageOptions

	Color     bool
	bgPalette ColorPalette
	spPalette ColorPalette

	UnpackedMonoPals [3][4]uint32

	Scheduler *scheduler.Scheduler

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

	InstInjectionFunc func(gb *GameBoy, op uint8)

	CyclesPerSndGen int64
	CurrFps         int64
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

func NewGameBoy(ctx *audio.Context, path string) *GameBoy {
	img := ebiten.NewImage(width, height)

	gb := &GameBoy{
		Image:     img,
		Cpu:       NewCpu(),
		Clock:     CPU_SPEED, // t cycle count
		Joypad:    0xFF,
		Cartridge: cartridge.NewCartridge(path, path+".save"),
		Palette:   &config.Conf.Gb.Palette,
		Scheduler: scheduler.NewScheduler(),
		Apu:       apu.NewApu(ctx, BUFFER_SIZE),
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

	gb.MemoryBus.WRAMBank = 1
	gb.MemoryBus.Hdma.Dst = 0xFFFF
	gb.MemoryBus.Hdma.Src = 0xFFFF
	gb.InitSaveLoop()

	if config.Conf.Gb.Bios.Direct {
		gb.DirectBoot()
	} else {
		gb.BiosBoot()
	}

	if config.Conf.General.Logger {
		L = NewLogger("./loggy", gb)
	}

	if ctx != nil {
		gb.CyclesPerSndGen = int64(CPU_SPEED / ctx.SampleRate())
		gb.Scheduler.Schedule(EVENT_SND_FRAME_SEQ, 1, 0, gb.ClockFrameSequencerEvent, nil)
		gb.Scheduler.Schedule(EVENT_SND_SAMPLE_GEN, 1, 0, gb.AudioSampleEvent, nil)
	}

	return gb
}

func (gb *GameBoy) BiosBoot() {
	if gb.Color {
		if p := config.Conf.Gb.Bios.GbcPath; p != "" {
			buf, _, _ := utils.ReadFile(p)
			gb.MemoryBus.Boot = &buf
			return
		}
	} else {
		if p := config.Conf.Gb.Bios.DmgPath; p != "" {
			buf, _, _ := utils.ReadFile(p)
			gb.MemoryBus.Boot = &buf
			return
		}
	}

	gb.DirectBoot()
}

func (gb *GameBoy) DirectBoot() {
	gb.MemoryBus.bootflg = 1
	gb.Cpu.PC = 0x100
	gb.Cpu.SP = 0xFFFE
	gb.Cpu.a = 0x01
	gb.Cpu.b = 0x00
	gb.Cpu.c = 0x13
	gb.Cpu.d = 0x00
	gb.Cpu.e = 0xD8
	gb.Cpu.h = 0x01
	gb.Cpu.l = 0x4D
	gb.Cpu.f = Flags{
		Z: true,
		S: false,
		H: true,
		C: true,
	}

	if gb.Color {
		gb.Cpu.a = 0x11
	}

	// memory
	gb.Write(0xFF04, 0x1E) // not sure on this one
	gb.Write(0xFF05, 0x00)
	gb.Write(0xFF06, 0x00)
	gb.Write(0xFF07, 0x00)
	gb.Cpu.IF = 0xE1
	gb.Write(0xFF10, 0x80)
	gb.Write(0xFF11, 0xBF)
	gb.Write(0xFF12, 0xF3)
	gb.Write(0xFF14, 0xBF)
	gb.Write(0xFF16, 0x3F)
	gb.Write(0xFF17, 0x00)
	gb.Write(0xFF19, 0xBF)
	gb.Write(0xFF1A, 0x7F)
	gb.Write(0xFF1B, 0xFF)
	gb.Write(0xFF1C, 0x9F)
	gb.Write(0xFF1E, 0xBF)
	gb.Write(0xFF20, 0xFF)
	gb.Write(0xFF21, 0x00)
	gb.Write(0xFF22, 0x00)
	gb.Write(0xFF23, 0xBF)
	gb.Write(0xFF24, 0x77)
	gb.Write(0xFF25, 0xF3)

	gb.Write(0xFF26, 0xF1)

	gb.Write(0xFF40, 0x91)
	gb.Write(0xFF41, 0x81)
	gb.Write(0xFF42, 0x00)
	gb.Write(0xFF43, 0x00)
	//gb.Write(0xFF44, 0x90)
	gb.Write(0xFF45, 0x00)
	gb.Write(0xFF47, 0xFC)
	gb.Write(0xFF48, 0xFF)
	gb.Write(0xFF49, 0xFF)
	gb.Write(0xFF4A, 0x00)
	gb.Write(0xFF4B, 0x00)
	gb.Write(0xFFFF, 0x00)
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

func (gb *GameBoy) Update(fps int64) {
	if gb.Paused {
		return
	}

	gb.CurrFps = fps

	if gb.Apu.Ctx != nil {
		gb.CyclesPerSndGen = (int64(CPU_SPEED/gb.Apu.Ctx.SampleRate()) * fps) / 60
	}

	nextFrame := gb.Scheduler.CurrentCycle + CYCLES_FRAME
	for gb.Scheduler.CurrentCycle < nextFrame {
		if gb.Cpu.Halted {
			gb.Tick(4)
		} else {
			gb.Execute()
		}
		// pending irq should be checked after handling irq, see mooneye/ei_sequence
		gb.Tick(gb.UpdateInterrupt())
		if gb.Cpu.PendingInterrupt {
			gb.Cpu.IME = true
			gb.Cpu.PendingInterrupt = false
		}
	}
}

//go:inline
func (gb *GameBoy) Tick(tCycles int64) {
	gb.Scheduler.Add(tCycles >> gb.DoubleSpeedFlag)

	if gb.Timer.Enabled {
		gb.UpdateTimers(tCycles)
	} else {
		gb.Timer.Div += uint16(tCycles)
	}

	if gb.MemoryBus.Oam.Pending || gb.MemoryBus.Oam.IsActive {
		gb.MemoryBus.Oam.Tick(gb, tCycles)
	}
}

func (gb *GameBoy) ToggleMute() bool {
	gb.Muted = !gb.Muted
	gb.Apu.ToggleMute(gb.Muted)
	return gb.Muted
}

func (gb *GameBoy) TogglePause() bool {
	gb.Paused = !gb.Paused
	gb.Apu.TogglePause(gb.Paused)
	return gb.Paused
}

var IRQ_SRC = [...]uint16{0x40, 0x48, 0x50, 0x58, 0x60}

func (gb *GameBoy) SetIrq(bit uint8) {
	gb.Cpu.IF |= bit
}

func (gb *GameBoy) UpdateInterrupt() int64 {
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

	t.BCycle = false

	if t.PendingOverflow {
		t.TIMA = t.TMA
		gb.SetIrq(IRQ_TMR)
		t.PendingOverflow = false
		t.BCycle = true
	}

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
	if gb.PrepareSpeedToggle {
		gb.PrepareSpeedToggle = false
		gb.Cpu.Halted = false
		gb.DoubleSpeedFlag = (^gb.DoubleSpeedFlag) & 1
		gb.MemoryBus.IO[0x4D] = gb.DoubleSpeedFlag << 7
	}
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
		sw      = float64(screen.Bounds().Dx())
		sh      = float64(screen.Bounds().Dy())
		scale   = utils.ScaleImage(sw, sh, width, height)
		offsetX = (sw - (width * scale)) / 2
		offsetY = (sh - (height * scale)) / 2
	)

	gb.DrawOptions.GeoM.Reset()
	gb.DrawOptions.GeoM.Scale(scale, scale)
	gb.DrawOptions.GeoM.Translate(offsetX, offsetY)
	screen.DrawImage(gb.Image, &gb.DrawOptions)
}
