package gameboy

import (
	"fmt"
	"log"
	"os"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/apu"
	"github.com/aabalke/guac/emu/gb/cartridge"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/oto"
)

const (
	DIV  = 0xFF04
	TIMA = 0xFF05
	TMA  = 0xFF06
	TAC  = 0xFF07

	tileWidth = 8
	width     = 160
	height    = 144

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

	Cartridge cartridge.Cartridge
	Cpu       *Cpu
	MemoryBus MemoryBus
	FPS       int

    // cycles are tcycles, 1/4 mcycles
    frameCycles int
	Cycles int
	Clock              int
	DoubleSpeed        bool
	PrepareSpeedToggle bool
	Timer              Timer

	Joypad uint8

	Image      *ebiten.Image
	Screen     [width][height]uint32
	bgPriority [width][height]bool
	spMinx     [width]int32
	pixelDrawn [width]bool

	Paused bool
	Muted  bool

	Apu *apu.Apu
}

type Timer struct {
	DivReg           int
	Counter          int
	ScanlineCounter  int
	InterruptPending bool
}

func NewGameBoy(path string, ctx *oto.Context) *GameBoy {

	img := ebiten.NewImage(width, height)

	gb := &GameBoy{
		Image:  img,
		Cpu:    NewCpu(),
		FPS:    60,
		Clock:  4194304,
		Joypad: 0xFF,
		Cartridge: cartridge.Cartridge{
			Data:    make([]uint8, 0),
			RomPath: path,
			SavPath: path + ".save",
		},
		Palette: config.Conf.Gb.Palette,
	}

	gb.bgPalette.Init()
	gb.spPalette.Init()
	gb.Pixels = make([]byte, width*height*4)

	const (
		SND_FREQUENCY = 48000 // sample rate
		SND_SAMPLES   = 512
	)
	gb.Apu = apu.NewApu(ctx, gb.Clock, SND_FREQUENCY, SND_SAMPLES)

	gb.LoadGame(path)

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
    multiplier := 1
    if gb.DoubleSpeed {
        multiplier = 2
    }
    targetCycles := gb.Clock / gb.FPS * multiplier
    for gb.frameCycles < targetCycles {
        if gb.Cpu.Halted {
            gb.Tick(4) // halted still burns cycles
        } else {
            gb.Execute() // ticking happens inside here now
        }

        gb.Tick(gb.UpdateInterrupt())
    }
    gb.frameCycles -= targetCycles // carry over, don't reset to 0

    gb.Apu.Play(gb.Muted)
}

func (gb *GameBoy) Tick(tCycles int) {

    if tCycles == 0 {
        return
    }
    if gb.DoubleSpeed {
        tCycles /= 2
    }
    gb.frameCycles += tCycles
    gb.Cycles = tCycles

    gb.UpdateGraphics()
    gb.UpdateTimers()

    gb.Apu.SoundClock(uint32(tCycles), gb.DoubleSpeed)
}

func (gb *GameBoy) ToggleMute() bool {
	gb.Muted = !gb.Muted
	return gb.Muted
}

func (gb *GameBoy) TogglePause() bool {
	gb.Paused = !gb.Paused
	return gb.Paused
}

func (gb *GameBoy) LoadGame(filepath string) {

	buffer, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	for i := range len(buffer) {
		gb.Cartridge.Data = append(gb.Cartridge.Data, uint8(buffer[i]))
	}

	gb.Cartridge.ParseHeader()
	gb.loadCartridge()
	initMemory(gb)
}

func (gb *GameBoy) loadCartridge() {

	log.Printf("Title: %s\n", gb.Cartridge.Title)

    switch {
    case config.Conf.Gb.ForceGBC:
        gb.Cartridge.ColorMode = true
    case config.Conf.Gb.ForceDMG:
        gb.Cartridge.ColorMode = false
    }

	if gb.Cartridge.ColorMode {
		gb.Color = true
	}

	if gb.Color {
		gb.Cpu.a = 0x11
		log.Printf("Color mode: GBC")
	} else {
		log.Printf("Color mode: DMG")
	}

	ramData, err := cartridge.ReadRam(gb.Cartridge.SavPath)

	if err != nil {
		ramData = make([]uint8, 0x8000)
	}

	gb.Cartridge.RamData = ramData

	switch gb.Cartridge.Type {
	case 0x00, 0x08, 0x09:
		println("ROM")

		gb.Cartridge.Mbc = &cartridge.Mbc0{
			RomBank: 1,
			RamBank: 1,
		}

	case 0x01, 0x02, 0x03:
		println("MBC1")

		gb.Cartridge.Mbc = &cartridge.Mbc1{
			RomBank: 1,
			RamBank: 0,
		}
	case 0x0F, 0x10, 0x11, 0x12, 0x13:
		println("MBC3")

		gb.Cartridge.Mbc = &cartridge.Mbc3{
			RomBank: 1,
			RamBank: 0,
			Rtc: cartridge.Rtc{
				Rtc:  make([]uint8, 0x10),
				Temp: make([]uint8, 0x10),
			},
		}
	case 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E:
		println("MBC5")

		gb.Cartridge.Mbc = &cartridge.Mbc5{
			RomBank: 1,
			RamBank: 0,
		}

	default:
		panic(fmt.Sprintf("UNSUPPORTED TYPE %X", gb.Cartridge.Type))
	}

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

    cycles := 20
    if gb.DoubleSpeed {
        cycles = 10
    }

    if !gb.Cpu.IME && gb.Cpu.Halted {
        gb.Cpu.Halted = false
        return cycles
    }

	for i := range 5 {

		if (handling>>i)&1 == 0 {
			continue
		}

		gb.Cpu.IME = false
		gb.Cpu.Halted = false
		gb.Cpu.IF &^= (1 << i)
		gb.StackPush(gb.Cpu.PC)
		gb.Cpu.PC = IRQ_SRC[i]
		gb.Cpu.isBranching = true
		return cycles
	}

	return 0
}

var IRQ_SRC = [...]uint16{0x40, 0x48, 0x50, 0x58, 0x60}

func (gb *GameBoy) UpdateTimers() {

	cycles := gb.Cycles

	if gb.DoubleSpeed {
		cycles *= 2
	}

	Mem := &gb.MemoryBus.Memory
	t := &gb.Timer

	//t.DivReg += gb.Cycles
	t.DivReg += cycles

	if t.DivReg >= 0xFF {
		t.DivReg -= 0xFF
		Mem[DIV]++
	}

	if disabled := gb.MemoryBus.Memory[TAC]&0b100 == 0; disabled {
		return
	}

	t.Counter += cycles

	// is tma handled properly?

	freq := freqs[Mem[TAC]&3]

	for t.Counter >= freq {
		t.Counter -= freq
		tima := Mem[TIMA]
		if overflow := tima == 0xFF; overflow {
			Mem[TIMA] = 0
			t.InterruptPending = true
		} else {
			Mem[TIMA] = tima + 1
		}
	}

	if t.InterruptPending {
		Mem[TIMA] = Mem[TMA]
		gb.RequestInterrupt(IRQ_TMR)
		t.InterruptPending = false
	}
}

var freqs = [...]int{1024, 16, 64, 256, 1024}

func (gb *GameBoy) RequestInterrupt(mask uint8) {
	gb.Cpu.IF |= mask | 0xE0
}

func (gb *GameBoy) toggleDoubleSpeed() {

	if !gb.PrepareSpeedToggle {
		return
	}

	gb.PrepareSpeedToggle = false
	gb.DoubleSpeed = !gb.DoubleSpeed
	gb.Cpu.Halted = false

	var v uint8 = 0
	if gb.DoubleSpeed {
		v |= 1 << 7
	}

	gb.MemoryBus.Memory[0xFF4D] = v
}

func (gb *GameBoy) Close() {
	gb.Muted = true
	gb.Paused = true
	gb.Apu.Close()

    if L != nil {
        L.Close()
    }
}
