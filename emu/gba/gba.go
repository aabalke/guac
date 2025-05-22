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
    MAX_COUNT = 100_000
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

        cycles := 4

        opcode := gba.Mem.Read32(r[PC])

        if gba.Halted && gba.ExitHalt {
            gba.Halted = false
        } 

        if !gba.Halted {
            cycles = gba.Cpu.Execute(opcode)
        }

        //if CURR_INST > 100_000 && r[PC] == 0x80B063E { 237603
        // 243_270
        //if CURR_INST == 243_285 pass
        //if CURR_INST == 243242 || CURR_INST == 243243 { //grabs dma3 which is 0 and should be 1
        ////if r[PC] == 0x80B159A {
        //    //gba.Paused = true
        //    gba.Debugger.print(CURR_INST)
        //    //return instCount
        //}

        //if CURR_INST == 243244 {
        //    gba.Paused = true
        //    return instCount
        //}

        CURR_INST++

        gt.update(cycles)
        gba.Timers.Increment(uint32(cycles))
        gba.updateIRQ()
	}

    //fmt.Printf("DMA3 WC %08X\n", gba.Mem.Read8(0x400_00DC))

    gba.checkDmas(DMA_MODE_REF)

    gba.debugGraphics()
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

    gba.LoadBios("./emu/gba/res/bios.gba")

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

func (gba *GBA) updateIRQ() {

    //for i := range 160 {
    //    // scanline
    //}

    const DISPSTAT = 0x0400_0004
    if dispstat := gba.Mem.Read8(DISPSTAT); utils.BitEnabled(dispstat, 3) {
        const IRQ_VBLANK = 0x00
        gba.triggerIRQ(IRQ_VBLANK)
    }
}

func (gba *GBA) triggerIRQ(irq uint32) {

    const IF = 0x202

    gba.Mem.BIOS_MODE = BIOS_IRQ

    iack := uint16(gba.Mem.Read8(IF))
	iack = iack | (1 << irq)
	gba.Mem.IO[IF] = uint8(iack)
    gba.Mem.IO[IF+1] = uint8(iack>>8)
	gba.Halted = false

	gba.checkIRQ()
}

func (gba *GBA) checkIRQ() {

    const IE = 0x200
    const IF = 0x202
    const IME = 0x208

    cond1 := !gba.Cpu.Reg.CPSR.GetFlag(FLAG_I)
    cond2 := gba.Mem.IO[IME] & 1 > 0
    cond3 := (uint16(gba.Mem.Read8(IE)) & uint16(gba.Mem.Read8(IF))) > 0
    if cond1 && cond2 && cond3 {
        panic("EXECPTION IN CHECK IRQ")
        //g.exception(irqVec, IRQ)
    }
}

func (gba *GBA) exception(addr uint32, mode uint32) {

    reg := &gba.Cpu.Reg

	cpsr := reg.CPSR
    curr := reg.getMode()
	reg.setMode(curr, mode)
	reg.SPSR[BANK_ID[curr]] = cpsr

	reg.R[14] = gba.exceptionReturn(addr)
	reg.CPSR.SetFlag(FLAG_T, false)
	reg.CPSR.SetFlag(FLAG_I, true)

    const (
        RESET_VEC = 0x0
        FIQ_VEC = 0x1C
    )

	switch addr & 0xff {
	case RESET_VEC, FIQ_VEC:
        reg.CPSR.SetFlag(FLAG_F, true)
	}
	reg.R[15] = addr
	//gba.pipelining()
}

func (gba *GBA) exceptionReturn(vec uint32) uint32 {
    reg := &gba.Cpu.Reg

	pc := reg.R[15]

	t := reg.CPSR.GetFlag(FLAG_T)
	switch vec {
	case UND_VEC, SWI_VEC:
		if t {
			pc -= 2
		} else {
			pc -= 4
		}
	case FIQ_VEC, IRQ_VEC, PREFETCH_VEC:
		if !t {
			pc -= 4
		}
	}
	return pc
}
