package nds

import (
	"fmt"
	"os"

	"github.com/aabalke/guac/emu/nds/cart"
	"github.com/aabalke/guac/emu/nds/cpu"
	"github.com/aabalke/guac/emu/nds/cpu/arm7"
	"github.com/aabalke/guac/emu/nds/cpu/arm9"
	"github.com/aabalke/guac/emu/nds/cpu/cp15"
	"github.com/aabalke/guac/emu/nds/mem"
	"github.com/aabalke/guac/emu/nds/mem/dma"
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

	//NUM_SCANLINES   = SCREEN_HEIGHT + 70 // or 71 ???
	NUM_SCANLINES   = SCREEN_HEIGHT + 70 // or 71 ???

	CYCLES_HDRAW    = 1606
	//CYCLES_HBLANK   = 524 // need to verify
	CYCLES_HBLANK   = 526 // need to verify
	CYCLES_SCANLINE = CYCLES_HDRAW + CYCLES_HBLANK
	CYCLES_VDRAW    = CYCLES_SCANLINE * SCREEN_HEIGHT
	CYCLES_VBLANK   = CYCLES_SCANLINE * 70 // or 71???
	CYCLES_FRAME    = CYCLES_VDRAW + CYCLES_VBLANK)

type Nds struct {
	mem  mem.Mem
	arm7 arm7.Cpu
	arm9 arm9.Cpu
    ppu *ppu.PPU
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
        ppu: ppu.NewPPU(),
	}

    nds.ppu.EngineA.Pixels = &nds.PixelsBottom
    nds.ppu.EngineB.Pixels = &nds.PixelsTop
    nds.ppu.EngineB.IsB = true

    irq9 := cpu.Irq{IsArm9: true}
	irq7 := cpu.Irq{}

    for i := range 8 {
        if i < 4 {
            nds.mem.Timers[i].IsArm9 = true
        }

        nds.mem.Timers[i].Idx = i % 4
    }

    cp15 := &cp15.Cp15{}
    cp15.Init(&nds.mem)

	nds.Debugger = Debugger{&nds}
	nds.arm7 = *arm7.NewCpu(&nds.mem, &irq7)
	nds.arm9 = *arm9.NewCpu(&nds.mem, &irq9, cp15)
	nds.mem = mem.NewMemory(
        &nds.arm7.Reg.R[15],
        &nds.arm7.Halted, &nds.arm9.Halted,
        &nds.arm7.Dma, &nds.arm9.Dma,
        &irq7, &irq9, &nds.Cartridge, nds.ppu)

    nds.arm9.Dma[0].Init(0, &nds.mem, &irq9, true)
    nds.arm9.Dma[1].Init(1, &nds.mem, &irq9, true)
    nds.arm9.Dma[2].Init(2, &nds.mem, &irq9, true)
    nds.arm9.Dma[3].Init(3, &nds.mem, &irq9, true)

    nds.arm7.Dma[0].Init(0, &nds.mem, &irq7, false)
    nds.arm7.Dma[1].Init(1, &nds.mem, &irq7, false)
    nds.arm7.Dma[2].Init(2, &nds.mem, &irq7, false)
    nds.arm7.Dma[3].Init(3, &nds.mem, &irq7, false)

    nds.LoadGame(path)
	//nds.arm9.Reset()

    nds.DirtyInit()

    logger = NewLogger("./log.csv", &nds)

	return &nds
}

func (nds *Nds) checkBadPc() {

    reg9 := &nds.arm9.Reg
    reg7 := &nds.arm7.Reg

    if reg9.R[15] > 0x400_0000 && reg9.R[15] < 0xFFFF_0000 {
        panic(fmt.Sprintf("BAD ARM9 PC %08X CPSR %08X CURR %d\n", reg9.R[15], reg9.CPSR, CURR_INST))
    }
    if (reg7.R[15] > 0x400_0000 && reg7.R[15] < 0x600_0000) || reg7.R[15] >= 0x700_0000 {
        panic(fmt.Sprintf("BAD ARM7 PC %08X CPSR %08X CURR %d\n", reg7.R[15], reg7.CPSR, CURR_INST))
    }

    // should probably check proper vramwram for arm7

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


    //zeroWordcnt := 0x20

    //if nds.mem.Read32(reg9.R[15], true) == 0x0 {

    //    zeros := true

    //    for i := uint32(0); i < uint32(zeroWordcnt); i += 4 {
    //        if nds.mem.Read32(reg9.R[15] + i, true) != 0x0 {
    //            zeros = false
    //            break
    //        }
    //    }

    //    if zeros {
    //        panic(fmt.Sprintf("BAD ARM9 PC %08X (ZEROS) CPSR %08X CURR %d\n", reg9.R[15], reg9.CPSR, CURR_INST))
    //    }
    //}

    //if nds.mem.Read32(reg7.R[15], false) == 0x0 {

    //    zeros := true

    //    for i := uint32(0); i < uint32(zeroWordcnt); i += 4 {
    //        if nds.mem.Read32(reg7.R[15] + i, false) != 0x0 {
    //            zeros = false
    //            break
    //        }
    //    }

    //    if zeros {
    //        panic(fmt.Sprintf("BAD ARM7 PC %08X (ZEROS) CPSR %08X CURR %d\n", reg7.R[15], reg7.CPSR, CURR_INST))
    //    }
    //}


    if reg9.R[15] < 0x30 && !nds.arm9.LowVector {
            panic(fmt.Sprintf("BAD ARM9 PC %08X (LOW WHEN HIGH) CPSR %08X CURR %d\n", reg9.R[15], reg9.CPSR, CURR_INST))
    }
}

func (nds *Nds) checkMode() {

    validModes := []uint32 {
        0b10000,
        0b10001,
        0b10010,
        0b10011,
        0b10111,
        0b11011,
        0b11111,
    }

    m9 := uint32(nds.arm9.Reg.CPSR) & 0x1F
    m7 := uint32(nds.arm7.Reg.CPSR) & 0x1F

    validM9 := false
    validM7 := false

    for _, v := range validModes {
        if v == m9 { validM9 = true }
        if v == m7 { validM7 = true }
    }

    if !validM9 {
        panic(fmt.Sprintf("ARM9 MODE INVALID %02X CURR %d\n", m9, CURR_INST))
    }

    if !validM7 {
        panic(fmt.Sprintf("ARM7 MODE INVALID %02X CURR %d\n", m7, CURR_INST))
    }
}

var prev uint32

func (nds *Nds) Update() {

    r := &nds.arm9.Reg.R
    r7 := &nds.arm7.Reg.R

    _ = r
    _ = r7

	if nds.Paused {
		return
	}

	for nds.Drawn = false; !nds.Drawn; {

        nds.checkBadPc()
        nds.checkMode()

        // arm9 thumb ~1 cycles, arm ~2 cycles
        // arm7 thumb ~2 cycles, arm ~4 cycles

        //if r[15] == 0x020D9980 {
        //    panic(fmt.Sprintf("PC %08X CURR %d", r[15], CURR_INST))
        //}

        //if v := nds.mem.Read32(0x03807564, false); v != prev {
        //    fmt.Printf("UPDATING VALUE V %08X\n", v)
        //    prev = v
        //}

        //if r7[15] == 0x037F847C {
        //    fmt.Printf("HIT IE \n")
        //}

        //if r7[15] == 0x037FB4D8 && CURR_INST >= 8_223_888 && r7[0] == 0x12 {
        //    fmt.Printf("R0 %08X IF %08X IE %08X CURR %d\n", r7[0], nds.arm7.Irq.IF, nds.arm9.Irq.IE, CURR_INST)
        //}

        //if CURR_INST >= 50_000_000 {
        //    nds.Debugger.print(int(CURR_INST))
        //    os.Exit(0)
        //}


		if !nds.arm9.Halted {
            thumbExec :=  nds.arm9.Reg.IsThumb
            armExec := !nds.arm9.Reg.IsThumb && nds.AccCycles & 0b1 == 0

            if thumbExec || armExec  {
                //logger.Update(4_000_000, 8_000_000, CURR_INST, true)

                _, ok := nds.arm9.Execute()
                if !ok {
                    fmt.Printf("ARM9 Decode Error: PC %08X CURR %d\n", r[15], CURR_INST)
                    os.Exit(0)
                }
            }
		}

        if !nds.arm7.Halted {
            thumbExec :=  nds.arm7.Reg.IsThumb && nds.AccCycles & 0b1 == 0
            armExec := !nds.arm7.Reg.IsThumb && nds.AccCycles & 0b11 == 0

            if thumbExec || armExec  {
                // fmt.Printf("PC %08X CURR %d\n", r7[15] , CURR_INST)
                //logger.Update(8_912_144, 20_000_000, CURR_INST, false)
                // fmt.Printf("PC %08X CURR %d\n", r7[15] , CURR_INST)

                // logger.Update(8937080, 8984200, CURR_INST, false)
                // logger.Update(8223890,10223890, CURR_INST, false)
                // fmt.Printf("PC %08X CURR %d\n", r7[15], CURR_INST)
                _, ok := nds.arm7.Execute()
                if !ok {
                    fmt.Printf("ARM7 Decode Error: PC %08X CURR %d\n", r7[15], CURR_INST)
                    os.Exit(0)
                }
            }
        }

        nds.Tick(1)

		// irq has to be at end (count up tests)
		nds.arm9.CheckIrq()
        nds.arm7.CheckIrq()

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
    nds.arm9.Halted = false
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
			nds.CheckDmas(dma.ARM9_DMA_MODE_HBL, true)
		}
	}

	if newScanline := currScanlineCycles < prevScanlineCycles; newScanline {

		//gba.Apu.SoundClock(1232, false)

		dispstat.SetHBlank(false)

		vcount++
		if vcount >= NUM_SCANLINES {
			vcount = 0
		}

        nds.mem.Vcount = vcount

		switch vcount {
		case 0:
			nds.CheckDmas(dma.ARM9_DMA_MODE_STA, true)
			nds.ppu.EngineA.Backgrounds[2].BgAffineReset()
			nds.ppu.EngineA.Backgrounds[3].BgAffineReset()
			nds.ppu.EngineB.Backgrounds[2].BgAffineReset()
			nds.ppu.EngineB.Backgrounds[3].BgAffineReset()
		case SCREEN_HEIGHT:
			dispstat.SetVBlank(true)
			nds.CheckDmas(dma.ARM9_DMA_MODE_VBL, true)
			nds.CheckDmas(dma.ARM7_DMA_MODE_VBL, true)
		case SCREEN_HEIGHT + 1:
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
        nds.ppu.Rasterizer.Render.UpdateRender()
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

    for i := range 4 {

        if ok := nds.arm7.Dma[i].CheckMode(mode); ok {
            nds.arm7.Dma[i].Transfer()
        }
    }
}

func (nds *Nds) UpdateTimers(cycles uint32) {

	overflow, setIrq := false, false

    for i := range 4 {
        if !nds.mem.Timers[i].Enabled {
            continue
        }

        overflow, setIrq = nds.mem.Timers[i].Update(overflow, cycles)

        if !setIrq {
            continue
        }

        nds.arm9.Irq.SetIRQ(3 + uint32(i))
    }

	overflow, setIrq = false, false

    for i := 4; i < 8; i++ {

        if !nds.mem.Timers[i].Enabled {
            continue
        }

        overflow, setIrq = nds.mem.Timers[i].Update(overflow, cycles)

        if !setIrq {
            continue
        }

        nds.arm7.Irq.SetIRQ(3 + uint32(i - 4))
    }
}
