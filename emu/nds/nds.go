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
)

const (
	SCREEN_WIDTH  = 256
	SCREEN_HEIGHT = 192

	NUM_SCANLINES   = SCREEN_HEIGHT + 70 // or 71 ???

	CYCLES_HDRAW    = 1606
	CYCLES_HBLANK   = 524 // need t overify
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


    arm9Irq := cpu.Irq{IsArm9: true}
	arm7Irq := cpu.Irq{}
    arm9Dma := [4]mem.DMA{}

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

    // this is init for hello world

    nds.arm7.Reg.R[2] =  0x4003EC8
    nds.arm7.Reg.R[4] =  0x27FFFF0
    nds.arm7.Reg.R[6] =  0x8000000
    nds.arm7.Reg.R[7] =  0x37F8000
    nds.arm7.Reg.R[12] = 0x80000C0
    nds.arm7.Reg.R[13] = 0x380FD80
    nds.arm7.Reg.R[14] = 0x80000FB

    // needs 0xFF returned from arm7 0x1c2 (spi bus data)

	return &nds
}

func (nds *Nds) checkBadPc() {
    r := &nds.arm9.Reg.R
    r7 := &nds.arm7.Reg.R

    if r[15] % 2 == 1 || (nds.mem.Read32(r[15], true) == 0x0 && nds.mem.Read32(r[15] + 4, true) == 0x0) {
        panic(fmt.Sprintf("BAD PC ARM9 @ PC %08X OP %08X CURR %d\n", r[15], nds.mem.Read32(r[15], true), CURR_INST))
    }

    if r7[15] % 2 == 1 || (nds.mem.Read32(r7[15], true) == 0x0 && nds.mem.Read32(r7[15] + 4, true) == 0x0) {
        panic(fmt.Sprintf("BAD PC ARM7 @ PC %08X OP %08X CURR %d\n", r7[15], nds.mem.Read32(r7[15], false), CURR_INST))
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
    _ = fmt.Sprintf("")

	nds.Drawn = false

    cycleArm7 := false
	for !nds.Drawn {

		cycles := 4

        //logger.Update(0,0,CURR_INST)

		if !nds.arm9.Halted {
			nds.arm9.Execute()
		}

        if cycleArm7 && !nds.arm7.Halted {
			nds.arm7.Execute()
        }

        nds.Tick(uint32(cycles))

		//// irq has to be at end (count up tests)
		nds.arm9.CheckIrq()
		nds.arm7.CheckIrq()

        cycleArm7 = !cycleArm7

        CURR_INST++
	}
}

func (nds *Nds) Tick(cycles uint32) {
    nds.VideoUpdate(cycles)
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
			//updateBackgrounds(gba, &gba.PPU.Dispcnt)
			//gba.PPU.bgPriorities = gba.getBgPriority(uint32(vcount), gba.PPU.Dispcnt.Mode, &gba.PPU.Backgrounds)
			//gba.PPU.objPriorities = gba.getObjPriority(uint32(vcount), &gba.PPU.Objects)
			nds.graphics(uint32(vcount))
			//gba.PPU.Backgrounds[2].BgAffineUpdate()
			//gba.PPU.Backgrounds[3].BgAffineUpdate()
			//nds.CheckDmas(DMA_MODE_HBL)
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
			//gba.PPU.Backgrounds[2].BgAffineReset()
			//gba.PPU.Backgrounds[3].BgAffineReset()
		case SCREEN_HEIGHT:
			dispstat.SetVBlank(true)
			//gba.checkDmas(DMA_MODE_VBL)
			// bios/bios.gba needs irq set on screen_height, iridion 3d needs screen_height + 1
			// I believe this is cycle related
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

func (nds *Nds) checkDmas(mode uint32) {

    for i := range 4 {

        //if ok := nds.arm7.Dma[i].CheckMode(mode); ok {
        //    nds.arm7.Dma[i].Transfer()
        //}

        if ok := nds.arm9.Dma[i].CheckMode(mode); ok {
            nds.arm9.Dma[i].Transfer()
        }
    }
}
