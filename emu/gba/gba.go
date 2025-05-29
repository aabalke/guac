package gba

import (
	"fmt"

	cart "github.com/aabalke33/guac/emu/gba/cart"
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
    CURR_INST = 0
    //MAX_COUNT = 100_000

    IN_EXCEPTION = false
    EXCEPTION_COUNT = 0
)

type GBA struct {
    Debugger *Debugger
    Cartridge *cart.Cartridge
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

    Gt *GraphicsTiming
    Ct *CycleTiming

    //Logger *Logger

    IntrWait uint32
}


func (gba *GBA) Update(exit *bool, instCount int) int {

    gt := gba.Gt
    r := &gba.Cpu.Reg.R

    if gba.Paused {
        return 0
    }

    gt.reset()

    //for range MAX_COUNT + 1 {
    for gba.Gt.RefreshCycles < (gba.Clock / gba.FPS) {

        gba.Ct.instCycles = 0

        cycles := 4

        //gba.Logger.WriteLog()
        //if CURR_INST == 10_000 {
        //    gba.Logger.Close()
        //}


        opcode := gba.Mem.Read32(r[PC])

        //if r[PC] > 0xA00_0000 {
        //    panic("CRAZY")
        //}

        //if IN_EXCEPTION {
        //    EXCEPTION_COUNT++
        //    fmt.Printf("PC %08X: OP %08X MODE %02X CPSR %08X SPSR CURR %08X SPSR BANKS %08X LR %08X R11 %08X R12 %08X R13 %08X R14 %08X\n", r[PC], opcode, gba.Cpu.Reg.getMode(), gba.Cpu.Reg.CPSR, gba.Cpu.Reg.SPSR[BANK_ID[gba.Cpu.Reg.getMode()]], gba.Cpu.Reg.SPSR, gba.Cpu.Reg.R[LR], gba.Cpu.Reg.R[11], gba.Cpu.Reg.R[12], gba.Cpu.Reg.R[13], gba.Cpu.Reg.R[14])
        //}

        //if EXCEPTION_COUNT > 20 { panic("END EXCEPTION LOG")}

        if gba.Halted {
            AckIntrWait(gba)
        }


        if gba.Halted && gba.ExitHalt {
            gba.Halted = false
        } 

        if !gba.Halted {
            cycles = gba.Cpu.Execute(opcode)
            //gba.Ct.prevAddr = r[PC]
        }

        //if gba.Mem.Read32(0x3007FFC) == 0x300_0000 {
        //    panic("HEREHEREHERE")
        //}

        //const DEBUG_START = 204768

        //if CURR_INST >= DEBUG_START - 2 {
        //    gba.Debugger.debugIRQ()
        //}

        //if CURR_INST == DEBUG_START + 2 {
        //    gba.Paused = true
        //    gba.Debugger.print(CURR_INST)
        //    return instCount
        //}

        // 35 // 458
        //if r[0] == 0xADE03C && r[1] == 0xB65B5B22 {
        //    gba.Paused = true
        //    gba.Debugger.print(CURR_INST)
        //    return instCount
        //}

        //if CURR_INST == 204500 { GOOD
        //if CURR_INST == 204700 { GOOD
        //if CURR_INST == 204722 { GOOD
        //if CURR_INST == 204725 { FAILED
        //if CURR_INST == 204750 { FAILED
        //if CURR_INST == 204723 {
        //    gba.Paused = true
        //    gba.Debugger.print(CURR_INST)
        //    return instCount
        //}
        //if r[PC] == 0x80029F8 && CURR_INST >= 295791 {
        //    panic(fmt.Sprintf("CORRECT BEHAVIOR! %X\n", r[PC]))
        //}


        CURR_INST++

        gt.update(cycles)
        gba.Timers.Increment(uint32(cycles))
	}

    gba.checkDmas(DMA_MODE_REF)

    //for y := range gt.Scanline - gt.PrevScanline {
    //    gba.scanlineMode0(min(159, uint32(gt.PrevScanline + y)))
    //}

    //gt.PrevScanline = gt.Scanline

    //gba.debugGraphics()
    gba.graphics()

    return instCount
}

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
    gba.Gt = &GraphicsTiming{
        Gba: &gba,
    }

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
	gba.Cartridge = cart.NewCartridge(path, "")
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
