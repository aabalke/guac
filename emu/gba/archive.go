package gba


//func (gba *GBA) Update(exit *bool, instCount int) int {
//
//    gt := gba.Gt
//    r := &gba.Cpu.Reg.R
//
//    if gba.Paused {
//        return 0
//    }
//
//    gt.reset()
//
//    //for range MAX_COUNT + 1 {
//    for gba.Gt.RefreshCycles < (gba.Clock / gba.FPS) {
//
//        gba.Ct.instCycles = 0
//
//        cycles := 4
//
//        //gba.Logger.WriteLog()
//        //if CURR_INST == 10_000 {
//        //    gba.Logger.Close()
//        //}
//
//
//        opcode := gba.Mem.Read32(r[PC])
//
//        //if r[PC] > 0xA00_0000 {
//        //    panic("CRAZY")
//        //}
//
//        //if IN_EXCEPTION {
//        //    EXCEPTION_COUNT++
//        //    fmt.Printf("PC %08X: OP %08X MODE %02X CPSR %08X SPSR CURR %08X SPSR BANKS %08X LR %08X R11 %08X R12 %08X R13 %08X R14 %08X\n", r[PC], opcode, gba.Cpu.Reg.getMode(), gba.Cpu.Reg.CPSR, gba.Cpu.Reg.SPSR[BANK_ID[gba.Cpu.Reg.getMode()]], gba.Cpu.Reg.SPSR, gba.Cpu.Reg.R[LR], gba.Cpu.Reg.R[11], gba.Cpu.Reg.R[12], gba.Cpu.Reg.R[13], gba.Cpu.Reg.R[14])
//        //}
//
//        //if EXCEPTION_COUNT > 20 { panic("END EXCEPTION LOG")}
//
//        if gba.Halted {
//            AckIntrWait(gba)
//        }
//
//
//        if gba.Halted && gba.ExitHalt {
//            gba.Halted = false
//        } 
//
//        if !gba.Halted {
//            cycles = gba.Cpu.Execute(opcode)
//            //gba.Ct.prevAddr = r[PC]
//        }
//
//        //if gba.Mem.Read32(0x3007FFC) == 0x300_0000 {
//        //    panic("HEREHEREHERE")
//        //}
//
//        //const DEBUG_START = 204768
//
//        //if CURR_INST >= DEBUG_START - 2 {
//        //    gba.Debugger.debugIRQ()
//        //}
//
//        //if CURR_INST == DEBUG_START + 2 {
//        //    gba.Paused = true
//        //    gba.Debugger.print(CURR_INST)
//        //    return instCount
//        //}
//
//        // 35 // 458
//        //if r[0] == 0xADE03C && r[1] == 0xB65B5B22 {
//        //    gba.Paused = true
//        //    gba.Debugger.print(CURR_INST)
//        //    return instCount
//        //}
//
//        //if CURR_INST == 204500 { GOOD
//        //if CURR_INST == 204700 { GOOD
//        //if CURR_INST == 204722 { GOOD
//        //if CURR_INST == 204725 { FAILED
//        //if CURR_INST == 204750 { FAILED
//        //if CURR_INST == 204723 {
//        //    gba.Paused = true
//        //    gba.Debugger.print(CURR_INST)
//        //    return instCount
//        //}
//        //if r[PC] == 0x80029F8 && CURR_INST >= 295791 {
//        //    panic(fmt.Sprintf("CORRECT BEHAVIOR! %X\n", r[PC]))
//        //}
//
//
//        CURR_INST++
//
//        gt.update(cycles)
//        gba.Timers.Increment(uint32(cycles))
//	}
//
//    gba.checkDmas(DMA_MODE_REF)
//
//    //for y := range gt.Scanline - gt.PrevScanline {
//    //    gba.scanlineMode0(min(159, uint32(gt.PrevScanline + y)))
//    //}
//
//    //gt.PrevScanline = gt.Scanline
//
//    //gba.debugGraphics()
//    gba.graphics()
//
//    return instCount
//}

//type GraphicsTiming struct {
//    Gba             *GBA
//    RefreshCycles   int
//    PrevScanline    int
//    Scanline        int
//    HBlank          bool
//    VBlank          bool
//    hasVBlankDma    bool
//    hasHBlankDma    bool
//}
//
//func (gt *GraphicsTiming) reset() {
//    gt.RefreshCycles = 0
//    gt.PrevScanline = 0
//    gt.Scanline = 0
//    gt.HBlank = false
//    gt.VBlank = false
//    gt.hasVBlankDma = false
//    gt.hasHBlankDma = false
//}
//
//func (gt *GraphicsTiming) update(cycles int) {
//
//    const (
//        REFRESH = 280_896 // should this be replaced by clock / fps?
//        SCANLINE = 1232
//        HDRAW = 960
//        HBLANK = 272
//        VDRAW = 197120
//        VBLANK = 83776
//    )
//
//    //prevHBlank := gt.HBlank
//    //prevVBlank := gt.VBlank
//    
//    //gt.RefreshCycles += cycles
//    //gt.HBlank = gt.RefreshCycles % SCANLINE > HDRAW
//    //gt.Scanline = gt.RefreshCycles / SCANLINE
//    //gt.VBlank = gt.Scanline > 160
//
//    //dispstat := &gt.Gba.Mem.Dispstat
//
//    //dispstat.SetVBlank(gt.VBlank)
//    //dispstat.SetHBlank(gt.HBlank)
//    //dispstat.SetVCounter(gt.Scanline)
//    //vcounter := (uint32(*dispstat) >> 8) & 0b1111_1111
//
//    //if gt.VBlank {
//    //    gt.Gba.Mem.IO[0x202] |= 1
//    //    //fmt.Printf("IRQ EXCEPTION CHECK AT MEM IE\n")
//    //    //gt.Gba.checkIRQ()
//    //    //gt.Gba.triggerIRQ(0)
//    //}
//
//    //if gt.HBlank {
//    //    gt.Gba.Mem.IO[0x202] |= 10
//    //    //gt.Gba.checkIRQ()
//    //    //gt.Gba.triggerIRQ(1)
//    //}
//
//    //if gt.Scanline == int(vcounter) {
//    //    gt.Gba.Mem.IO[0x202] |= 100
//    //    *dispstat = Dispstat(uint32(*dispstat) | uint32(100))
//    //}
//
//    //if gt.VBlank && !prevVBlank {
//    //    gt.Gba.checkDmas(DMA_MODE_VBL)
//    //    //gt.Gba.triggerIRQ(0)
//    //}
//    //if gt.HBlank && !prevHBlank {
//    //    gt.Gba.checkDmas(DMA_MODE_HBL)
//    //    //gt.Gba.triggerIRQ(1)
//    //}
//}
//
//func (gba *GBA) updateMode0(dispcnt *Dispcnt) {
//
//    bgPriorities := gba.getBgPriority(0)
//    wins := NewWindows(dispcnt, gba)
//    bld := NewBlend(gba)
//    bldPal := NewBlendPalette(bld)
//
//    x, y := uint32(0), uint32(0)
//    for y = range SCREEN_HEIGHT {
//
//        objPriorities := gba.getObjPriority(y, dispcnt)
//
//        for x = range SCREEN_WIDTH {
//
//            index := (x + (y*SCREEN_WIDTH)) * 4
//
//            bldPal.reset(gba)
//
//            var objTransparent bool
//
//            for i := range 4 {
//
//                // 0 is highest priority
//                decIdx := 3 - i
//
//                // bg and objs are prioritized so obj0, is above obj1 if both same
//                // priority. this is why [bgCount - 1 - j]
//
//                bgCount := len(bgPriorities[decIdx])
//                for j := range bgCount {
//
//                    bgIdx := bgPriorities[decIdx][bgCount - 1 - j]
//
//                    if bgEnabled(dispcnt, int(bgIdx)) {
//                        palData, ok, palZero := gba.background(x, y, bgIdx, false, wins)
//                        if ok && !palZero {
//                            bldPal.setBlendPalettes(palData, bgIdx, false)
//                        }
//                    }
//                }
//
//                if obj := dispcnt.DisplayObj; obj {
//
//                    objPal := uint32(0)
//                    objExists := false
//                    objCount := len(objPriorities[decIdx])
//
//                    for j := range objCount {
//                        objIdx := objPriorities[decIdx][objCount - 1 - j]
//                        palData, ok, palZero, obj := gba.object(x, y, dispcnt, objIdx * 0x8, wins)
//
//                        if ok && !palZero {
//                            objTransparent = obj.Mode == 1
//                            objExists = true
//                            objPal = palData
//                        }
//                    }
//
//                    if objExists {
//                        bldPal.setBlendPalettes(objPal, 0, true)
//                    }
//                }
//            }
//
//            finalPalData := bldPal.blend(objTransparent)
//
//            gba.applyColor(finalPalData, uint32(index))
//        }
//    }
//}
//
//func (gba *GBA) updateMode1(dispcnt *Dispcnt) {
//
//    bgPriorities := gba.getBgPriority(0)
//    wins := NewWindows(dispcnt, gba)
//    bld := NewBlend(gba)
//    bldPal := NewBlendPalette(bld)
//
//    x, y := uint32(0), uint32(0)
//    for y = range SCREEN_HEIGHT {
//
//        objPriorities := gba.getObjPriority(y, dispcnt)
//
//        for x = range SCREEN_WIDTH {
//
//            index := (x + (y*SCREEN_WIDTH)) * 4
//
//            bldPal.reset(gba)
//
//            var objTransparent bool
//
//            for i := range 4 {
//
//                // 0 is highest priority
//                decIdx := 3 - i
//
//                // bg and objs are prioritized so obj0, is above obj1 if both same
//                // priority. this is why [bgCount - 1 - j]
//
//                bgCount := len(bgPriorities[decIdx])
//                for j := range bgCount {
//
//                    bgIdx := bgPriorities[decIdx][bgCount - 1 - j]
//
//                    if !bgEnabled(dispcnt, int(bgIdx)) {
//                        continue
//                    }
//
//                    palData, ok, palZero := uint32(0), false, false
//
//                    if bgIdx == 2 {
//                        palData, ok, palZero = gba.background(x, y, bgIdx, true, wins)
//                    } else {
//                        palData, ok, palZero = gba.background(x, y, bgIdx, false, wins)
//                    }
//
//                    if ok && !palZero {
//                        bldPal.setBlendPalettes(palData, bgIdx, false)
//                    }
//                }
//
//                if obj := dispcnt.DisplayObj; obj {
//
//                    objPal := uint32(0)
//                    objExists := false
//                    objCount := len(objPriorities[decIdx])
//
//                    for j := range objCount {
//                        objIdx := objPriorities[decIdx][objCount - 1 - j]
//                        palData, ok, palZero, obj := gba.object(x, y, dispcnt, objIdx * 0x8, wins)
//
//                        if ok && !palZero {
//                            objTransparent = obj.Mode == 1
//                            objExists = true
//                            objPal = palData
//                        }
//                    }
//
//                    if objExists {
//                        bldPal.setBlendPalettes(objPal, 0, true)
//                    }
//                }
//            }
//
//            finalPalData := bldPal.blend(objTransparent)
//
//            gba.applyColor(finalPalData, uint32(index))
//        }
//    }
//}
//
//func (gba *GBA) updateMode3() {
//
//
//	const (
//		SIZE           = 0x12C00
//		BASE           = 0x0600_0000
//		BYTE_PER_PIXEL = 2
//	)
//
//	Mem := gba.Mem
//
//	index := 0
//	for i := uint32(0); i < SIZE; i += BYTE_PER_PIXEL {
//		data := Mem.Read16(BASE + i)
//        gba.applyColor(data, uint32(index))
//		index += 4
//	}
//}
//
//func (gba *GBA) updateMode4(dispcnt *Dispcnt) {
//
//	const (
//		SIZE = 0x9600
//	)
//
//    BASE := uint32(0x0600_0000)
//
//    if dispcnt.DisplayFrame1 {
//        BASE += 0xA_000
//    }
//
//	Mem := gba.Mem
//
//	index := 0
//	for i := uint32(0); i < SIZE; i++ {
//
//		palIdx := Mem.Read8(BASE + i)
//
//		palData := gba.getPalette(uint32(palIdx), 0, false)
//
//        gba.applyColor(palData, uint32(index))
//		index += 4
//	}
//}
//
//func (gba *GBA) updateMode5(dispcnt *Dispcnt) {
//
//	const (
//        MAP_WIDTH = 160
//        MAP_HEIGHT = 128
//	)
//
//    BASE := uint32(0x0600_0000)
//
//    if dispcnt.DisplayFrame1 {
//        BASE += 0xA_000
//    }
//
//	Mem := gba.Mem
//
//	index := 0
//    i := uint32(0)
//    for range MAP_HEIGHT {
//        for range MAP_WIDTH {
//            data := Mem.Read16(BASE + i)
//            gba.applyColor(data, uint32(index))
//            index += 4
//            i += 2
//        }
//
//        index += 4 * (SCREEN_WIDTH - MAP_WIDTH) // map diff screen width and map width
//    }
//}

func (gba *GBA) graphics() {

    //addr := gba.Mem.Read16(0x0400_0000 + DISPCNT)
    //dispcnt := NewDispcnt(addr)

    //x, y := uint32(0), uint32(0)

    //for x = range SCREEN_WIDTH {
    //    for y = range SCREEN_HEIGHT {
    //        gba.background(x, y, 0, false, NewWindows(dispcnt, gba))
    //    }
    //}

    //gba.getTiles(0x601_0000, 0x1E, true, false)
    //gba.getTiles(0x600_3800, 0x1E, false, false)
    //gba.debugPalette()
    return

    //gba.clear()

	//switch dispcnt.Mode {
	//case 0:
	//	gba.updateMode0(dispcnt)
	//case 1:
	//	gba.updateMode1(dispcnt)
	//case 2:
    //    panic("mode 2")
	//	//gba.updateMode2()
	//case 3:
	//	gba.updateMode3()
	//case 4:
	//	gba.updateMode4(dispcnt)
	//case 5:
	//	gba.updateMode5(dispcnt)
    //default: panic("UNKNOWN MODE")
	//}
}
