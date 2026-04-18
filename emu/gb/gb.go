package gameboy

import (
	"log"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/gb/apu"
	"github.com/aabalke/guac/emu/gb/cartridge"
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
	Palette [][]uint8
	Pixels  []byte

	Color     bool
	bgPalette ColorPalette
	spPalette ColorPalette

	UnpackedMonoPals [3][4]uint32

	Cartridge *cartridge.Cartridge
	Cpu       *Cpu
	MemoryBus MemoryBus
	FPS       int

	Stat Stat
	Lcdc Lcdc

	// cycles are tcycles, 1/4 mcycles
	frameCycles        int
	Cycles             int
	Clock              int
	DoubleSpeedFlag    uint8
	PrepareSpeedToggle bool
	Timer              Timer

	Joypad uint8

	Image      *ebiten.Image
	Screen     [width][height]uint32
	spMinx     [width]int32
	bgPriority [width][height]bool
	pixelDrawn [width]bool

	Paused bool
	Muted  bool

	Apu *apu.Apu
}

type Timer struct {
	DotCounter int // should this be seperate?

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

func NewGameBoy(path string, ctx *oto.Context) *GameBoy {

	img := ebiten.NewImage(width, height)

	gb := &GameBoy{
		Image:     img,
		Cpu:       NewCpu(),
		FPS:       60,
		Clock:     4194304, // t cycle count
		Joypad:    0xFF,
		Cartridge: cartridge.NewCartridge(path, path+".save"),
		Palette:   config.Conf.Gb.Palette,
	}

	gb.Lcdc.gb = gb
	gb.MemoryBus.Hdma.gb = gb

	gb.bgPalette.Init()
	gb.spPalette.Init()
	gb.Pixels = make([]byte, width*height*4)

	const (
		SND_FREQUENCY = 48000 // sample rate
		SND_SAMPLES   = 512
	)
	gb.Apu = apu.NewApu(ctx, gb.Clock, SND_FREQUENCY, SND_SAMPLES)

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

	L = NewLogger("./loggy", gb)

	return gb
}

func (gb *GameBoy) GetSize() (int32, int32) {
	return height, width
}

func (gb *GameBoy) GetPixels() []byte {
	return gb.Pixels
}

func (gb *GameBoy) Update() {
	if gb.Paused {
		return
	}

	targetCycles := gb.Clock / gb.FPS << gb.DoubleSpeedFlag
	for gb.frameCycles < targetCycles {
		if gb.Cpu.Halted {
			gb.Tick(4)
		} else {
			gb.Execute()
		}

		gb.Tick(gb.UpdateInterrupt())

        targetCycles = gb.Clock / gb.FPS << gb.DoubleSpeedFlag
	}

	gb.frameCycles -= targetCycles

	gb.Apu.Play(gb.Muted)
}

func (gb *GameBoy) Tick(tCycles int) {

	if tCycles == 0 {
		return
	}

	tCycles >>= int(gb.DoubleSpeedFlag)

	gb.frameCycles += tCycles
	gb.Cycles = tCycles

	if gb.Lcdc.Enabled {
		gb.UpdateGraphics(tCycles)
	}

	gb.UpdateTimers(tCycles) // frame sequencer is here since div apu is controlled by div

    gb.MemoryBus.Oam.Tick(gb, tCycles)

	gb.Apu.WaveChannel.ClockWave(uint32(tCycles), uint32(gb.frameCycles))
	gb.Apu.SoundClock(uint32(tCycles), uint32(gb.DoubleSpeedFlag))
}

func (gb *GameBoy) ToggleMute() bool {
	gb.Muted = !gb.Muted
	return gb.Muted
}

func (gb *GameBoy) TogglePause() bool {
	gb.Paused = !gb.Paused
	return gb.Paused
}

func (gb *GameBoy) SetIrq(bit uint8) {
	gb.Cpu.IF |= bit
}

func (gb *GameBoy) UpdateInterrupt() int {

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
		gb.Cpu.IF &^= (1 << i)

        // stack push
        gb.Cpu.SP--
        gb.Write(gb.Cpu.SP, uint8(gb.Cpu.PC>>8))
        gb.Cpu.SP--
        gb.Write(gb.Cpu.SP, uint8(gb.Cpu.PC))

		gb.Cpu.PC = IRQ_SRC[i]
		gb.Cpu.isBranching = true

		return 20
	}

	return 0
}

var IRQ_SRC = [...]uint16{0x40, 0x48, 0x50, 0x58, 0x60}

func (gb *GameBoy) UpdateTimers(cycles int) {

	t := &gb.Timer

	cycles <<= gb.DoubleSpeedFlag

	prev := t.Div
	t.Div += uint16(cycles)

	mask := uint16(1 << 12)
	mask <<= gb.DoubleSpeedFlag
	if prev&mask != 0 && t.Div&mask == 0 {
		gb.Apu.ClockFrameSequencer()
	}

	if !t.Enabled {
		return
	}

	if t.BCycle && cycles >= 4 {
		t.BCycle = false
	}

	if t.PendingOverflow && cycles >= 4 {
		t.TIMA = t.TMA
		gb.SetIrq(IRQ_TMR)
		t.PendingOverflow = false
		t.BCycle = true
	}

	// instead of keeping separate counter, use falling edge to inc
	// requires figuring out count of edges first

	period := fallingEdgeBits[t.FreqBits] << 1
	edgeCnt := (t.Div / period) - (prev / period)
	for range edgeCnt {

		if t.BCycle {
			t.BCycle = false
		}

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

var fallingEdgeBits = [...]uint16{1 << 9, 1 << 3, 1 << 5, 1 << 7}

func (gb *GameBoy) toggleDoubleSpeed() {

	if !gb.PrepareSpeedToggle {
		return
	}

	gb.PrepareSpeedToggle = false
	if gb.DoubleSpeedFlag != 0 {
		gb.DoubleSpeedFlag = 0
	} else {
		gb.DoubleSpeedFlag = 1
	}

	gb.Cpu.Halted = false

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
