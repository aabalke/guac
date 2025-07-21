package gameboy

import (
	"fmt"
	"os"

	"github.com/aabalke33/guac/emu/gb/cartridge"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	DIV  = 0xFF04
	TIMA = 0xFF05
	TMA  = 0xFF06
	TAC  = 0xFF07

	tileWidth = 8
	width     = 160
	height    = 144
)

type GameBoy struct {

    Palette   [][]uint8
    Pixels *[]byte

    Color     bool
    bgPalette *ColorPalette
    spPalette *ColorPalette

	Cartridge cartridge.Cartridge
	Cpu       Cpu
	Apu       APU
	MemoryBus MemoryBus
	FPS       int

	Clock int
    DoubleSpeed bool
    PrepareSpeedToggle bool
	Timer Timer

	Joypad uint8

    Image *ebiten.Image
	Screen [width][height]uint32
	bgPriority [width][height]bool

    Cycles int

    Paused bool
    Muted bool
}

type Timer struct {
	DivReg          int
	Counter         int
	ScanlineCounter int
    InterruptPending bool
}

func NewGameBoy(path string) *GameBoy {

    img := ebiten.NewImage(width, height)

    gb := GameBoy{
        Image: img,
		Cpu: *NewCpu(),
		Apu: APU{
			SampleRate: 44100,
			Enabled:    true,
		},
		FPS:    60,
		Clock:  4194304,
		Joypad: 0xFF,
		Cartridge: cartridge.Cartridge{
			Data: make([]uint8, 0),
		},
        Palette: palettes["greyscale"],
        bgPalette: NewColorPalette(),
        spPalette: NewColorPalette(),
	}

    pixels := make([]byte, width*height*4)
    gb.Pixels = &pixels

    gb.Apu.GameBoy = &gb

	gb.Apu.Init()

    gb.LoadGame(path)

	return &gb
}

func (gb *GameBoy) GetSize() (int32, int32) {
    return height, width
}

func (gb *GameBoy) Update() {

    if gb.Paused {
        return
    }

    multiplier := 1
    if gb.DoubleSpeed {
        multiplier = 2
    }

    for updateCycles := 0; updateCycles < (gb.Clock / gb.FPS * multiplier); {

		cycles := 4

		opcode, err := gb.ReadByte(gb.Cpu.PC)

		if err != nil {
			panic(err)
		}

		if !gb.Cpu.Halted {
			cycles = gb.Execute(opcode)
		}

        if gb.DoubleSpeed {
            cycles /= 2
        }

		updateCycles += cycles
        gb.Cycles = cycles

		gb.UpdateGraphics()

		interruptCycles := gb.UpdateInterrupt()
        if gb.DoubleSpeed {
            interruptCycles /= 2
        }
        updateCycles += interruptCycles
        gb.Cycles += interruptCycles

		gb.UpdateTimers()
	}

    gb.UpdateDisplay()
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
	gb.Apu.MemoryBus = &gb.MemoryBus
}

func (gb *GameBoy) loadCartridge() {

	fmt.Printf("Title: %s\n", gb.Cartridge.Title)

    // Debug DMG mode
    //gb.Cartridge.ColorMode = false

    if gb.Cartridge.ColorMode {
        gb.Color = true
    }

    if gb.Color {
        gb.Cpu.Registers.a = 0x11
        println("Color Mode: CMG")
    } else {
        println("Color Mode: DMG")
    }

	ramData, err := cartridge.ReadRam(gb.Cartridge.Path)

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
                Rtc: make([]uint8, 0x10),
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

func (gb *GameBoy) UpdateInterrupt() (cycles int) {

	if gb.Cpu.PendingInterrupt {
		gb.Cpu.InterruptMaster = true
		gb.Cpu.PendingInterrupt = false
		return 0
	}

	if !gb.Cpu.InterruptMaster && !gb.Cpu.Halted {
		return 0
	}

	interruptFlag := gb.MemoryBus.Memory[0xFF0F]
	interruptEnabled := gb.MemoryBus.Memory[0xFFFF]

	if interruptFlag == 0 {
		return 0
	}

	const (
		VBLANK = iota
		STAT   //LCD
		TIMER
		SERIAL
		JOYPAD
	)

	// interrupts are servered by priority, see above const
	for i := range 5 {
		handlerAvailable := gb.flagEnabled(interruptFlag, uint8(i))
		handlerRequested := gb.flagEnabled(interruptEnabled, uint8(i))

		if !(handlerAvailable && handlerRequested) {
			continue
		}

		if !gb.Cpu.InterruptMaster && gb.Cpu.Halted {
			gb.Cpu.Halted = false
			return 20
		}

		gb.Cpu.InterruptMaster = false
		gb.Cpu.Halted = false

        req := gb.MemoryBus.Memory[0xFF0F]
		newFlag := req & ^(1 << i)
		err := gb.WriteByte(0xFF0F, newFlag)
		if err != nil {
			panic(err)
		}

		gb.StackPush(gb.Cpu.PC)

		switch i {
		case VBLANK:
			gb.Cpu.PC = 0x40
		case STAT:
			gb.Cpu.PC = 0x48
		case TIMER:
			gb.Cpu.PC = 0x50
		case SERIAL:
			gb.Cpu.PC = 0x58
		case JOYPAD:
			gb.Cpu.PC = 0x60
		}

		return 20
	}

	return 0
}

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

	if !gb.EnableClock() {
		return
	}

	//t.Counter += gb.Cycles
	t.Counter += cycles

	freq := gb.SelectCycleFreq()

	for t.Counter >= freq {
		t.Counter -= freq
		tima := Mem[TIMA]
		if tima == 0xFF {
            Mem[TIMA] = 0
            t.InterruptPending = true
		} else {
            Mem[TIMA] = tima + 1
        }
	}

    if t.InterruptPending {
        Mem[TIMA] = Mem[TMA]
        gb.RequestInterrupt(0b100)
        t.InterruptPending = false
    }
}

func (gb *GameBoy) EnableClock() bool {
	tac := gb.MemoryBus.Memory[TAC]
    return gb.flagEnabled(tac, 2)
}

func (gb *GameBoy) SelectCycleFreq() int {

    tac := gb.MemoryBus.Memory[TAC]

	switch clock := tac & 0b11; clock {
	case 0b00:
		return 1024
	case 0b01:
		return 16
	case 0b10:
		return 64
	case 0b11:
		return 256
	default:
		return 1024
	}
}

func (gb *GameBoy) RequestInterrupt(mask uint8) {

	interruptFlag := gb.MemoryBus.Memory[0xFF0F] | 0xE0

	newFlag := interruptFlag | mask // may need ^
	err := gb.WriteByte(0xFF0F, newFlag)
	if err != nil {
		panic(err)
	}
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
        v = 0b10000000
    }

    gb.MemoryBus.Memory[0xFF4D] = v
}

func (gb *GameBoy) Close() {
    gb.Muted = true
    gb.Paused = true
}
