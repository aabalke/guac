package nds

import (
	"fmt"
	"os"

	"github.com/aabalke/guac/emu/nds/cart"
	"github.com/aabalke/guac/emu/nds/cpu"
	"github.com/aabalke/guac/emu/nds/cpu/arm7"
	"github.com/aabalke/guac/emu/nds/cpu/arm9"
	"github.com/aabalke/guac/emu/nds/mem"
	"github.com/aabalke/guac/emu/nds/ppu"
	"github.com/aabalke/guac/emu/nds/utils"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/oto"
)

var (
	CURR_INST = uint64(0)
    _ = os.Args
    _ = fmt.Sprintf("")
)

const (
	SCREEN_WIDTH  = 256
	SCREEN_HEIGHT = 192

	NUM_SCANLINES   = SCREEN_HEIGHT + 70 // or 71 ???

	CYCLES_HDRAW    = 1606
	CYCLES_HBLANK   = 524 // need to verify
	CYCLES_SCANLINE = CYCLES_HDRAW + CYCLES_HBLANK
	CYCLES_VDRAW    = CYCLES_SCANLINE * SCREEN_HEIGHT
	CYCLES_VBLANK   = CYCLES_SCANLINE * 70 // or 71???
	CYCLES_FRAME    = CYCLES_VDRAW + CYCLES_VBLANK
)

type Nds struct {
	mem  mem.Mem
	arm7 arm7.Cpu
	arm9 arm9.Cpu
    ppu ppu.PPU


    Cartridge cart.Cartridge

	Debugger Debugger

	Muted, Paused, Drawn    bool
	PixelsTop, PixelsBottom []byte
	ImageTop, ImageBottom   *ebiten.Image

    BtmAbs struct{T, B, L, R, W, H int} 

    AccCycles uint32
}

var logger *Logger

func NewNds(path string, _ *oto.Context) *Nds {

	nds := Nds{
		PixelsTop:    make([]byte, SCREEN_WIDTH*SCREEN_HEIGHT*4),
		PixelsBottom: make([]byte, SCREEN_WIDTH*SCREEN_HEIGHT*4),
		ImageTop:     ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT),
		ImageBottom:  ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT),
	}

    nds.ppu.EngineA.Pixels = &nds.PixelsBottom
    nds.ppu.EngineB.Pixels = &nds.PixelsTop
    nds.ppu.EngineB.IsB = true

    arm9Irq := cpu.Irq{IsArm9: true}
	arm7Irq := cpu.Irq{}
    arm9Dma := [4]mem.DMA{}

    for i := range 8 {
        if i < 4 {
            nds.mem.Timers[i].IsArm9 = true
        }

        nds.mem.Timers[i].Idx = i % 4
    }

	nds.Debugger = Debugger{&nds}
	nds.arm7 = *arm7.NewCpu(&nds.mem, &arm7Irq)
	nds.arm9 = *arm9.NewCpu(&nds.mem, &arm9Irq)
	nds.mem = mem.NewMemory(&arm9Dma, &arm7Irq, &arm9Irq, &nds.Cartridge, &nds.ppu)

    arm9Dma[0].Init(0, &nds.mem, &arm9Irq)
    arm9Dma[1].Init(1, &nds.mem, &arm9Irq)
    arm9Dma[2].Init(2, &nds.mem, &arm9Irq)
    arm9Dma[3].Init(3, &nds.mem, &arm9Irq)

    nds.LoadGame(path)
	//nds.arm9.Reset()

    nds.DirtyInit()

    logger = NewLogger("./log.csv", &nds)

    //nds.arm7.Halted = true
    //nds.mem.Write32(0x23FFC80, 0xDEADBEEF, false)

	return &nds
}

func (nds *Nds) checkBadPc() {

    reg9 := &nds.arm9.Reg
    reg7 := &nds.arm7.Reg

    switch {
    case reg9.IsThumb && reg9.R[15] & 0b1 != 0:
        panic(fmt.Sprintf("BAD ARM9 THUMB PC %08X CPSR %08X CURR %d\n", reg9.R[15], reg9.CPSR, CURR_INST))
    case !reg9.IsThumb && reg9.R[15] & 0b11 != 0:
        panic(fmt.Sprintf("BAD ARM9 ARM   PC %08X CPSR %08X CURR %d\n", reg9.R[15], reg9.CPSR, CURR_INST))
    case reg7.IsThumb && reg7.R[15] & 0b1 != 0:
        panic(fmt.Sprintf("BAD ARM7 THUMB PC %08X CPSR %08X CURR %d\n", reg7.R[15], reg7.CPSR, CURR_INST))
    case !reg7.IsThumb && reg7.R[15] & 0b11 != 0:
        panic(fmt.Sprintf("BAD ARM7 ARM   PC %08X CPSR %08X CURR %d\n", reg7.R[15], reg7.CPSR, CURR_INST))
    }
}

func (nds *Nds) Update() {

    r := &nds.arm9.Reg.R
    r7 := &nds.arm7.Reg.R

	if nds.Paused {
		return
	}

    _ = r
    _ = r7

	nds.Drawn = false
    cycleArm7 := false

	for !nds.Drawn {

        // will need half time cycles for thumb

        nds.checkBadPc()

		cycles := 2
        //if nds.arm9.Reg.IsThumb {
        //    cycles = 1
        //}

        //logger.Update(0, 200_000, CURR_INST,true)

		if !nds.arm9.Halted {
			nds.arm9.Execute()
		}

        if cycleArm7 && !nds.arm7.Halted {
			nds.arm7.Execute()
        }

        nds.Tick(uint32(cycles))

		//// irq has to be at end (count up tests)
		nds.arm9.CheckIrq()

        if cycleArm7 {
            nds.arm7.CheckIrq()
        }

        cycleArm7 = !cycleArm7

        CURR_INST++
	}
}

func (nds *Nds) Tick(cycles uint32) {
    nds.VideoUpdate(cycles)
    nds.UpdateTimers(cycles)
}

func (nds *Nds) ToggleMute() bool {
	nds.Muted = !nds.Muted
	return nds.Muted
}

func (nds *Nds) TogglePause() bool {
	nds.Paused = !nds.Paused
	return nds.Paused
}

func (nds *Nds) Close() {
	nds.Muted = true
	nds.Paused = true

    logger.Close()
    logger = nil
}

func (nds *Nds) LoadGame(path string) {
	nds.Cartridge = cart.NewCartridge(path, path+".save")
}

//temp 
func (nds *Nds) DirtyInit() {

    nds.mem.DirtyTransfer()

    nds.arm9.Reg.R[12] = nds.Cartridge.Header.Arm9EntryAddr
    nds.arm9.Reg.R[13] = 0x3002F7C
    nds.arm9.Reg.R[14] = nds.Cartridge.Header.Arm9EntryAddr
    nds.arm9.Reg.R[15] = nds.Cartridge.Header.Arm9EntryAddr
    nds.arm9.Reg.CPSR = 0x000_001F

    nds.arm7.Reg.R[12] = nds.Cartridge.Header.Arm7EntryAddr
    //nds.arm7.Reg.R[13] = 0x3002F7C
    nds.arm7.Reg.R[14] = nds.Cartridge.Header.Arm7EntryAddr
    nds.arm7.Reg.R[15] = nds.Cartridge.Header.Arm7EntryAddr
    nds.arm7.Reg.CPSR = 0x000_001F

    nds.arm7.Halted = false
}

// RidgeX/ygba BSD3
func (nds *Nds) VideoUpdate(cycles uint32) {

	dispstat := &nds.mem.Dispstat
    vcount := nds.mem.Vcount

	prevFrameCycles := nds.AccCycles
	nds.AccCycles += cycles //% CYCLES_FRAME
	if nds.AccCycles >= CYCLES_FRAME {
		nds.AccCycles -= CYCLES_FRAME
	}
	currFrameCycles := nds.AccCycles

	prevScanlineCycles := prevFrameCycles % CYCLES_SCANLINE
	currScanlineCycles := currFrameCycles % CYCLES_SCANLINE

	inHblank := currScanlineCycles >= CYCLES_HDRAW
	prevInHdraw := prevScanlineCycles < CYCLES_HDRAW
	if enteredHblank := inHblank && prevInHdraw; enteredHblank {

		dispstat.SetHBlank(true)
		if utils.BitEnabled(uint32(*dispstat), 4) {
			nds.arm9.Irq.SetIRQ(1)
			nds.arm7.Irq.SetIRQ(1)
		}

		if vcount < SCREEN_HEIGHT {

            a := &nds.ppu.EngineA
            b := &nds.ppu.EngineB
			updateBackgrounds(a)
			updateBackgrounds(b)
			a.BgPriorities = nds.getBgPriority(uint32(vcount), a.Dispcnt.Mode, &a.Backgrounds)
			b.BgPriorities = nds.getBgPriority(uint32(vcount), b.Dispcnt.Mode, &b.Backgrounds)

			a.ObjPriorities = nds.getObjPriority(uint32(vcount), &a.Objects)
			b.ObjPriorities = nds.getObjPriority(uint32(vcount), &b.Objects)
			nds.graphics(uint32(vcount))
			nds.ppu.EngineA.Backgrounds[2].BgAffineUpdate()
			nds.ppu.EngineA.Backgrounds[3].BgAffineUpdate()
			nds.ppu.EngineB.Backgrounds[2].BgAffineUpdate()
			nds.ppu.EngineB.Backgrounds[3].BgAffineUpdate()
			nds.CheckDmas(mem.ARM9_DMA_MODE_HBL, true)
		}
	}

	if newScanline := currScanlineCycles < prevScanlineCycles; newScanline {

		// this 1232 cycle count is estimate, should replace with actual
		//gba.Apu.SoundClock(1232, false)

		dispstat.SetHBlank(false)

		vcount++
		if vcount == NUM_SCANLINES {
			vcount = 0
		}

        nds.mem.Vcount = vcount

		switch vcount {
		case 0:
			nds.ppu.EngineA.Backgrounds[2].BgAffineReset()
			nds.ppu.EngineA.Backgrounds[3].BgAffineReset()
			nds.ppu.EngineB.Backgrounds[2].BgAffineReset()
			nds.ppu.EngineB.Backgrounds[3].BgAffineReset()
		case SCREEN_HEIGHT:
			dispstat.SetVBlank(true)
			nds.CheckDmas(mem.ARM9_DMA_MODE_VBL, true)
		//case SCREEN_HEIGHT + 1:
			if utils.BitEnabled(uint32(*dispstat), 3) {
                nds.arm9.Irq.SetIRQ(0)
                nds.arm7.Irq.SetIRQ(0)
			}
		case NUM_SCANLINES - 1:
			dispstat.SetVBlank(false)
		}

		match := dispstat.GetLYC() == vcount
		dispstat.SetVCFlag(match)

		if vcounterIRQ := utils.BitEnabled(uint32(*dispstat), 5); vcounterIRQ && match {
			nds.arm9.Irq.SetIRQ(2)
			nds.arm7.Irq.SetIRQ(2)
		}
	}

	if currFrameCycles < prevFrameCycles {
		nds.Drawn = true
	}
}

func (nds *Nds) CheckDmas(mode uint32, arm9 bool) {

    if arm9 {
        for i := range 4 {
            if ok := nds.arm9.Dma[i].CheckMode(mode); ok {
                nds.arm9.Dma[i].Transfer()
            }
        }
        return
    }

    panic("ARM7 DMA")

    //for i := range 4 {

    //    //if ok := nds.arm7.Dma[i].CheckMode(mode); ok {
    //    //    nds.arm7.Dma[i].Transfer()
    //    //}
    //}
}

func (nds *Nds) UpdateTimers(cycles uint32) {

	overflow, setIrq := false, false

    for i := range 8 {
        if nds.mem.Timers[i].Enabled {
            overflow, setIrq = nds.mem.Timers[i].Update(overflow, cycles)

            if !setIrq {
                continue
            }

            if i < 4 {
                nds.arm9.Irq.SetIRQ(3 + uint32(i))
                continue
            }

            nds.arm7.Irq.SetIRQ(3 + uint32(i - 4))
        }
    }
}
