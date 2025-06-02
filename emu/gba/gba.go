package gba

import (
	"fmt"
    "os"

	"github.com/aabalke33/guac/emu/gba/utils"
)

const (
	SCREEN_WIDTH  = 240
	SCREEN_HEIGHT = 160

	RESET_VEC           uint32 = 0x00
	UND_VEC             uint32 = 0x04
	SWI_VEC             uint32 = 0x08
	PREFETCH_VEC        uint32 = 0x0C
	DATA_ABORT_VEC      uint32 = 0x10
	ADDR_26_VEC         uint32 = 0x14
	IRQ_VEC             uint32 = 0x18
	FIQ_VEC             uint32 = 0x1C
)

var (
    _ = fmt.Sprintln("")
    _ = os.Args
    CURR_INST = 0
    //MAX_COUNT = 100_000

    IN_EXCEPTION = false
    EXCEPTION_COUNT = 0
)

type GBA struct {
    Debugger *Debugger
    Cartridge *Cartridge
    Cpu *Cpu
    Mem *Memory
	Screen [SCREEN_WIDTH][SCREEN_HEIGHT]uint32
	Pixels *[]byte
	DebugPixels *[]byte

	Paused bool
	Muted  bool 
    Joypad uint16
    Halted bool
    ExitHalt bool

    Clock int
    FPS int
    Cycles int
    Scanline int
 
    Timers Timers
    Dma    [4]DMA

    Ct *CycleTiming

    VCOUNT uint32

    //Logger *Logger

    IntrWait uint32

    Save bool
}


var SAVED_CYCLES = uint32(0)

func (gba *GBA) Update(exit *bool, instCount int) int {

    if gba.Paused {
        return 0
    }

    gba.VCOUNT = 0

    frameCycles := uint32(0)
    for range 160 {
        frameCycles = gba.UpdateScanline(frameCycles)
    }

    dispstat := uint32(gba.Mem.Dispstat)
    if utils.BitEnabled(dispstat, 3) {
        gba.triggerIRQ(0)
    }

    gba.Mem.Dispstat.SetVBlank(true)
    gba.Mem.IO[0x202] |= 1

    gba.checkDmas(DMA_MODE_VBL)

    for range 67 {
        frameCycles = gba.UpdateScanline(frameCycles)
    }

    gba.Mem.Dispstat.SetVBlank(false)
    gba.Mem.IO[0x202] &^= 1

    frameCycles = gba.UpdateScanline(frameCycles) // 227

    SAVED_CYCLES = frameCycles - (1232 * 228)

    //gba.checkDmas(DMA_MODE_REF)

    gba.graphics()

    return instCount
}

func (gba *GBA) UpdateScanline(frameCycles uint32) uint32 {

    dispstat := gba.Mem.Read16(0x400_0004)
    lyc := gba.Mem.Read8(0x400_0005)
    if vcounter := utils.BitEnabled(dispstat, 5) && gba.VCOUNT == lyc; vcounter {
        gba.triggerIRQ(2)
    }

    //Although the drawing time is only 960 cycles (240*4), the H-Blank flag is "0" for a total of 1006 cycles. gbatek
    frameCycles = gba.Exec(1006, frameCycles)

    dispstat = uint32(gba.Mem.Dispstat)
    vblank := utils.BitEnabled(dispstat, 0) 
    if utils.BitEnabled(dispstat, 4) && !vblank {
        gba.triggerIRQ(1)
    }

    gba.Mem.Dispstat.SetHBlank(true)
    gba.Mem.IO[0x202] |= 10
    gba.checkDmas(DMA_MODE_HBL)

    frameCycles = gba.Exec(1232 - 1006, frameCycles)

    gba.Mem.Dispstat.SetHBlank(false)
    gba.Mem.IO[0x202] &^= 10

    // draw scanline here once converted to scanline version

    gba.VCOUNT++

    return frameCycles
}

func (gba *GBA) Exec(requiredCycles, frameCycles uint32) uint32 {

    r := &gba.Cpu.Reg.R

    accCycles := frameCycles

    for accCycles < requiredCycles {

        cycles := 4

        gba.Ct.instCycles = 0

        opcode := gba.Mem.Read32(r[PC])


        if gba.Halted {
            AckIntrWait(gba)
        }

        if gba.Halted && gba.ExitHalt {
            gba.Halted = false
        } 

        //if CURR_INST == 542764 { panic(fmt.Sprintf("??? HALTED %t CURR %08d PC %08X, OPCODE %04X\n", gba.Halted, CURR_INST, r[PC], opcode)) }
        if !gba.Halted {
            cycles = gba.Cpu.Execute(opcode)
        }

        //if CURR_INST == 238970 { // GOOD PALETTE
        //if CURR_INST == 238849 { // GOOD PALETTE
        //if CURR_INST == 542_800 { // GOOD PALETTE
        //if CURR_INST == 542_853 { // GOOD PALETTE
        //if CURR_INST == 543669 { // BADDDDD
        //if CURR_INST == 543669 { // BADDDDD
        //if r[PC] == 0x08073B04 { // GOOD PALETTE
        //if CURR_INST == 543669 || CURR_INST == 542853 { // GOOD PALETTE
        ////if r[PC] == 0x804128A && r[4] == 0x6F7B { // GOOD PALETTE
        //if CURR_INST > 544000  && CURR_INST < 545000 {

        if CURR_INST == 542980 {
            //r[SP] = 0x3007E20
            //panic("HERE")
        }

        //if r[PC] == 0x8073AFE {
        //    //fmt.Printf("r[PC] %08X CURR %d\n", r[PC], CURR_INST)
        //    gba.Debugger.print(CURR_INST)
        //    gba.Paused = true
        //    //return accCycles - requiredCycles
        //    os.Exit(0)
        //}

        if !gba.Halted {
            CURR_INST++
        }
        gba.Timers.Increment(uint32(cycles))

        accCycles += uint32(cycles)
    }

    return accCycles - requiredCycles
}

var PREV_VALUE uint32

func (gba *GBA) checkDmas(mode uint32) {
    for i := range gba.Dma {
        if gba.Dma[i].checkMode(mode) {
            gba.Dma[i].transfer()
        }
    }
}

func NewGBA() *GBA {

	pixels := make([]byte, SCREEN_WIDTH*SCREEN_HEIGHT*4)
	debugPixels := make([]byte, 1080*1080*4)

	gba := GBA{
        Clock: 16_780_000,
        FPS: 60, // 59.7374117111
        Pixels: &pixels,
        DebugPixels: &debugPixels,
    }

    gba.Mem = NewMemory(&gba)
    gba.Cpu = NewCpu(&gba)
    gba.Debugger = &Debugger{&gba}

    gba.Ct = &CycleTiming{
        prevAddr: 0x800_0000,
    }

    //gba.Cpu.Reg.setMode(MODE_SYS, MODE_SWI)

    gba.Cpu.Reg.CPSR.SetFlag(FLAG_I, false)

    gba.Timers[0].Gba = &gba
    gba.Timers[1].Gba = &gba
    gba.Timers[2].Gba = &gba
    gba.Timers[3].Gba = &gba

    gba.Timers[0].Idx = 0
    gba.Timers[1].Idx = 1
    gba.Timers[2].Idx = 2
    gba.Timers[3].Idx = 3

    gba.Dma[0].Gba = &gba
    gba.Dma[1].Gba = &gba
    gba.Dma[2].Gba = &gba
    gba.Dma[3].Gba = &gba

    gba.Dma[0].Idx = 0
    gba.Dma[1].Idx = 1
    gba.Dma[2].Idx = 2
    gba.Dma[3].Idx = 3

    //gba.Logger = NewLogger(".log.txt", &gba)

    gba.LoadBios("./emu/gba/res/bios_magia.gba")

    //gba.SoftReset()

	return &gba
}

func (gba *GBA) GetSize() (int32, int32) {
	return SCREEN_HEIGHT, SCREEN_WIDTH
}

func (gba *GBA) GetPixels() []byte {
	return *gba.Pixels
}

func (gba *GBA) GetDebugPixels() []byte {
	return *gba.DebugPixels
}

func (gba *GBA) ToggleMute() bool {
	gba.Muted = !gba.Muted
	return gba.Muted
}

func (gba *GBA) TogglePause() bool {
	gba.Paused = !gba.Paused
	return gba.Paused
}

func (gba *GBA) Close() {
	gba.Muted = true
	gba.Paused = true
}

func (gba *GBA) LoadGame(path string) {
    gba.Cartridge = NewCartridge(gba, path, path + ".save")
}

func (gba *GBA) toggleThumb() {

    reg := &gba.Cpu.Reg

    newFlag := reg.R[PC] & 1 > 0

    reg.CPSR.SetFlag(FLAG_T, newFlag)

    if newFlag {
        reg.R[PC] &^= 1
        // pipe
        return
    }

    reg.R[PC] &^= 3
    // pipe
}

func (gba *GBA) triggerIRQ(irq uint32) {


    mem := gba.Mem

    mem.BIOS_MODE = BIOS_IRQ

    iack := mem.Read16(0x400_0202)

    iack |= (1 << irq)
    mem.IO[0x202] = uint8(iack)
    mem.IO[0x203] = uint8(iack>>8)
	gba.Halted = false

    //fmt.Printf("IRQ EXCEPTION CHECK AT TRIGGER\n")
	gba.checkIRQ()
}

func (gba *GBA) checkIRQ() {

    interruptEnabled := !gba.Cpu.Reg.CPSR.GetFlag(FLAG_I)
    ime := utils.BitEnabled(gba.Mem.Read16(0x400_0208), 0)
    interrupts := (gba.Mem.Read16(0x400_0200) & gba.Mem.Read16(0x400_0202)) != 0

    if interruptEnabled && ime && interrupts {
        gba.handleInterrupt()
    }
}
