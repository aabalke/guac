package nds

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/nds/ppu"
	"github.com/aabalke/guac/emu/nds/utils"
)

var b16 = binary.LittleEndian.Uint16

var wg = sync.WaitGroup{}

func (nds *Nds) graphics(y uint32) {

    a := &nds.ppu.EngineA
    b := &nds.ppu.EngineB

	switch a.Dispcnt.DisplayMode {
	case 0:
		nds.screenoff(y, a)
	case 1:
		nds.standard(y, a)
	case 2:
		nds.vramDisplay(y, a)
		//nds.standard(y, a)
	case 3:
        panic("MAIN MEM FIFO")
        nds.MemFifoDisplay(a)
	}

    nds.ppu.Capture.CaptureLine(y)

    switch b.Dispcnt.DisplayMode {
    case 0:
        nds.screenoff(y, b)
    case 1:
        nds.standard(y, b)
    }
}

func (nds *Nds) screenoff(y uint32, engine *ppu.Engine) {
    start := y * SCREEN_WIDTH << 2
    end := start + SCREEN_WIDTH << 2
    copy(engine.Pixels[start:end], nds.ppu.WHITE_SCANLINE)
}

func (nds *Nds) vramDisplay(y uint32, engine *ppu.Engine) {

	x := uint32(0)
	for x = range SCREEN_WIDTH {
        palData, _ := nds.setRawBitmap(engine, x, y)
        r, g, b := engine.MasterBright.Apply(palData)
        i := (x + (y * SCREEN_WIDTH)) << 2

        (engine.Pixels)[i] = r
        (engine.Pixels)[i+1] = g
        (engine.Pixels)[i+2] = b
        //(engine.Pixels)[i+3] = 0xFF
    }
}

func (nds *Nds) MemFifoDisplay(engine *ppu.Engine) {
    copy(engine.Pixels, nds.ppu.DisplayFifo.Pixels)
}

func (nds *Nds) standard(y uint32, engine *ppu.Engine) {

	if config.Conf.Nds.Threads == 0 {

		x := uint32(0)
		for x = range SCREEN_WIDTH {
			nds.render(x, y, engine)
		}

		return
	}

	WAIT_GROUPS := config.Conf.Nds.Threads
	dx := SCREEN_WIDTH / WAIT_GROUPS

	wg.Add(WAIT_GROUPS)

	for i := range WAIT_GROUPS {

		go func(i int) {

			defer wg.Done()

			for j := range dx {
				x := uint32((i * dx) + j)
				nds.render(x, y, engine)
			}
		}(i)
	}

	wg.Wait()
}

func (nds *Nds) render(x, y uint32, engine *ppu.Engine) {
	dispcnt := &engine.Dispcnt
	wins := &engine.Windows
	bgs := &engine.Backgrounds
	objPriorities := &engine.ObjPriorities
	bgPriorities := &engine.BgPriorities

	bldPal := ppu.NewBlendPalette(x, &engine.Blend, nds.getPalette(0, 0, false, engine.IsB))

	var objMode uint32
	var inObjWindow bool


	// work backwards for proper priorities
	for i := 3; i >= 0; i-- {

		for j := len(bgPriorities[i]) - 1; j >= 0; j-- {

			bgIdx := bgPriorities[i][j]
			bg := &bgs[bgIdx]

			if !ppu.WindowPixelAllowed(bgIdx, x, y, wins) {
				continue
			}

            palData, alpha, ok := uint32(0), float64(1), false

            switch bg.Type {
            case ppu.BG_TYPE_TEX:
                palData, ok = nds.setBackgroundPixel(engine, bg, bgIdx, x, y)
            case ppu.BG_TYPE_AFF:
				palData, ok = nds.setAffineBackgroundPixel(engine, bg, bgIdx, x)
            case ppu.BG_TYPE_LAR:
				palData, ok = nds.setAffine16BackgroundPixel(engine, bg, bgIdx, x)
            case ppu.BG_TYPE_3D :
                palData, alpha, ok = nds.set3d(engine, bg, x, y)
            case ppu.BG_TYPE_BGM:
				palData, ok = nds.setAffine16BackgroundPixel(engine, bg, bgIdx, x)
            case ppu.BG_TYPE_256:
				palData, ok = nds.setBmpBackgroundPixel(engine, bg, x)
            case ppu.BG_TYPE_DIR:
				palData, ok = nds.setDirectBitmap(engine, bg, x)
            }

			if ok {
				bldPal.SetBlendPalettes(palData, uint32(bgIdx), false, false, true, alpha)
			}
		}

		if objDisabled := !dispcnt.DisplayObj; objDisabled {
			continue
		}

        ObjectLoop:
		for j := 0; j < len((*objPriorities)[i]); j++ {

			objIdx := (*objPriorities)[i][j]
	        obj := &engine.Objects[objIdx]

            obj.OneDimensional = dispcnt.TileObj1D

			if !ppu.WindowObjPixelAllowed(x, y, wins) {
				continue
			}

			var palData uint32
			var ok bool

            if bmp:= obj.Mode == 3; bmp {
                palData, ok = nds.setObjBmpPixel(engine, obj, x, y)
            } else {
                palData, ok = nds.setObjTilePixel(engine, obj, x, y)
            }

			switch {
			case ok && obj.Mode == 2:
				inObjWindow = true
				break ObjectLoop
			case ok:
				objMode = obj.Mode
				bldPal.SetBlendPalettes(palData, 0, true, objMode == 1, false, 0)
				break ObjectLoop
			}
		}
	}

	palData := bldPal.Blend(objMode == 1, x, y, wins, inObjWindow)
    r, g, b := engine.MasterBright.Apply(palData)
	i := (x + (y * SCREEN_WIDTH)) << 2
	(engine.Pixels)[i] = r
	(engine.Pixels)[i+1] = g
	(engine.Pixels)[i+2] = b
	//(engine.Pixels)[i+3] = 0xFF
}

func updateBackgrounds(engine *ppu.Engine) *[4]ppu.Background {

    bgs := &engine.Backgrounds

    getExtended := func(bg *ppu.Background) uint8 {

        if !bg.Palette256 {
            return ppu.BG_TYPE_BGM
        }

        // CharBaseBlock is precalced, need to / 0x4000
        if (bg.CharBaseBlock >> 14) & 1 == 1 {
            return ppu.BG_TYPE_DIR
        }

        return ppu.BG_TYPE_256
    }

	for i := range 4 {

        switch i {
        case 0:
            if !engine.IsB && engine.Dispcnt.Is3D {
                bgs[i].Type = ppu.BG_TYPE_3D
            } else {
                bgs[i].Type = ppu.BG_TYPE_TEX

                // does bg0 mode 6 panic?
            }
        case 1:
            bgs[i].Type = ppu.BG_TYPE_TEX
        case 2:

            switch engine.Dispcnt.Mode {
            case 0, 1, 3:
                bgs[i].Type = ppu.BG_TYPE_TEX
            case 2, 4: 
                bgs[i].Type = ppu.BG_TYPE_AFF
            case 5:
                bgs[i].Type = getExtended(&bgs[i])
            case 6:
                bgs[i].Type = ppu.BG_TYPE_LAR
            }

        case 3:

            switch engine.Dispcnt.Mode {
            case 0:
                bgs[i].Type = ppu.BG_TYPE_TEX
            case 1, 2: 
                bgs[i].Type = ppu.BG_TYPE_AFF
            case 3, 4, 5:
                bgs[i].Type = getExtended(&bgs[i])
            }
        }

		bgs[i].SetSize()

		//if (dispcnt.Mode == 1 && i == 2) || dispcnt.Mode == 2 {
		//	bgs[i].Palette256 = true
		//}
	}

	return bgs
}

func (nds *Nds) set3d(engine *ppu.Engine, bg *ppu.Background, x, y uint32) (uint32, float64, bool) {

    //xIdx := int(x) + int(bg.XOffset)
    //yIdx := int(y) + int(bg.YOffset)
	xIdx := (x + bg.XOffset) & ((bg.W) - 1)
	yIdx := (y + bg.YOffset) & ((bg.H) - 1)

	index := (xIdx + (yIdx * SCREEN_WIDTH))

    if xIdx >= SCREEN_WIDTH || yIdx >= SCREEN_HEIGHT {
        return 0, 0, false
    }

    render := nds.ppu.Rasterizer.Render

    pal, alpha := uint32(render.PixelPalettes[index]), render.Alphas[index]

    //if nds.ppu.Rasterizer.Disp3dCnt.RearPlaneBitmapEnabled == true && alpha != 1 {
    //    pal, ok := nds.setRearBitmap(engine, bg, x, y)
    //    return pal, 1, ok
    //}

    return pal, alpha, alpha > 0
    //return pal, alpha, true
}

func (nds *Nds) setRearBitmap(engine *ppu.Engine, bg *ppu.Background, x, y uint32) (uint32, bool) {

    const slot3Offset = 0x60000

    xIdx := x + nds.ppu.Rasterizer.RearPlane.OffsetX
    yIdx := y + nds.ppu.Rasterizer.RearPlane.OffsetY

    addr := uint32(xIdx+(yIdx * SCREEN_WIDTH)) * 2
    addr += slot3Offset

    fmt.Printf("addr %08X Slots % v\n", addr, nds.ppu.Vram.TextureSlots)

    data := uint32(nds.ppu.Vram.ReadTexture(addr))
    data |= uint32(nds.ppu.Vram.ReadTexture(addr + 1)) << 8

    data &^= 0x80

    return data, true

}

func (nds *Nds) setBackgroundPixel(engine *ppu.Engine, bg *ppu.Background, bgIdx, x, y uint32) (uint32, bool) {

	xIdx := (x + bg.XOffset) & ((bg.W) - 1)
	yIdx := (y + bg.YOffset) & ((bg.H) - 1)

    if bg.Mosaic {

        if engine.Mosaic.BgH != 0 {
            xIdx -= xIdx % (engine.Mosaic.BgH + 1)
        }

        if engine.Mosaic.BgV != 0 {
            yIdx -= yIdx % (engine.Mosaic.BgV + 1)
        }
    }

	map_x := xIdx >> 3
	map_y := yIdx >> 3
	quad_x := uint32(10) //32 * 32
	quad_y := uint32(10) //32 * 32
	if bg.Size == 3 {
		quad_y = 11
	}
	mapIdx := (map_y >> 5) << quad_y
	mapIdx += (map_x >> 5) << quad_x
	mapIdx += (map_y & 31) << 5
	mapIdx += (map_x & 31)
	mapIdx <<= 1

	mapAddr := bg.ScreenBaseBlock + mapIdx

    var banks uint32
    if !engine.IsB {
        mapAddr += engine.Dispcnt.ScreenBase
        banks = ppu.BANKS_A_2D_BG
    } else {
        mapAddr += 0x20_0000
        banks = ppu.BANKS_B_2D_BG
    }

    screenData := uint32(nds.ppu.Vram.ReadGraphical(mapAddr, banks))

	tileIdx := (screenData & 0b11_1111_1111) << 5

	tileAddr := bg.CharBaseBlock + tileIdx
	if bg.Palette256 {
		tileAddr += tileIdx
	}

    if !engine.IsB {
        tileAddr += engine.Dispcnt.CharBase
    } else {
        tileAddr += 0x20_0000
    }

	inTileX, inTileY := getPositionsBg(screenData, xIdx, yIdx)

	var inTileIdx uint32
	if bg.Palette256 {
		inTileIdx = inTileX + (inTileY << 3)
	} else {
		inTileIdx = (inTileX >> 1) + (inTileY << 2)
	}

    palIdx := uint32(nds.ppu.Vram.ReadGraphical(tileAddr + inTileIdx, banks))
    palNum := screenData >> 12

    return getBgPaletteData(nds, engine, bgIdx, bg.Palette256, palNum, palIdx, inTileX)
}

func (nds *Nds) setAffine16BackgroundPixel(engine *ppu.Engine, bg *ppu.Background, bgIdx, x uint32) (uint32, bool) {

	//if !bg.Palette256 {
	//	panic(fmt.Sprintf("AFFINE WITHOUT PAL 256"))
	//}

    vramOffset := uint32(0)

    if engine.IsB {
        vramOffset = uint32(0x20_0000)
    }

	pa := utils.Convert8_8Float(int16(bg.Pa))
	pc := utils.Convert8_8Float(int16(bg.Pc))
	xIdx := int(pa*float64(x) + bg.OutX)
	yIdx := int(pc*float64(x) + bg.OutY)

	if bg.Mosaic && engine.Mosaic.BgH != 0 {
		xIdx -= xIdx % int(engine.Mosaic.BgH+1)
	}

	if bg.Mosaic && engine.Mosaic.BgV != 0 {
		yIdx -= yIdx % int(engine.Mosaic.BgV+1)
	}

	out := xIdx < 0 || xIdx >= int(bg.W) || yIdx < 0 || yIdx >= int(bg.H)

	switch {
	case bg.AffineWrap:
		xIdx &= int(bg.W) - 1
		yIdx &= int(bg.H) - 1
	case !bg.AffineWrap && out:
		return 0, false
	}

    const BYTE_SHIFT = 1

	map_x := (uint32(xIdx)) & (bg.W - 1) >> 3
	map_y := ((uint32(yIdx)) & (bg.H - 1)) >> 3
	map_y *= bg.W >> 3
	mapIdx := map_y + map_x
    mapIdx <<= BYTE_SHIFT

	mapAddr := bg.ScreenBaseBlock + mapIdx

    var banks uint32
    if !engine.IsB {
        mapAddr += engine.Dispcnt.ScreenBase
        banks = ppu.BANKS_A_2D_BG
    } else {
        banks = ppu.BANKS_B_2D_BG
    }

    screenData := uint32(nds.ppu.Vram.ReadGraphical(vramOffset + mapAddr, banks))

	tileIdx := (screenData & 0b11_1111_1111) << 5

	tileAddr := bg.CharBaseBlock + tileIdx
    tileAddr += tileIdx

    if !engine.IsB {
        tileAddr += engine.Dispcnt.CharBase
    }

	inTileX, inTileY := getPositionsBg(screenData, uint32(xIdx), uint32(yIdx))

	inTileIdx := uint32(inTileX) + uint32(inTileY<<3)

	addr := vramOffset + tileAddr + inTileIdx
    palIdx := uint32(nds.ppu.Vram.Read(addr, true))
    palNum := screenData >> 12

    return getBgPaletteData(nds, engine, bgIdx, true, palNum, palIdx, inTileX)
}

func (nds *Nds) setAffineBackgroundPixel(engine *ppu.Engine, bg *ppu.Background, bgIdx, x uint32) (uint32, bool) {

    vramOffset := uint32(0)

    if engine.IsB {
        vramOffset = uint32(0x20_0000)
    }

	pa := utils.Convert8_8Float(int16(bg.Pa))
	pc := utils.Convert8_8Float(int16(bg.Pc))
	xIdx := int(pa*float64(x) + bg.OutX)
	yIdx := int(pc*float64(x) + bg.OutY)

	if bg.Mosaic && engine.Mosaic.BgH != 0 {
		xIdx -= xIdx % int(engine.Mosaic.BgH+1)
	}

	if bg.Mosaic && engine.Mosaic.BgV != 0 {
		yIdx -= yIdx % int(engine.Mosaic.BgV+1)
	}

	out := xIdx < 0 || xIdx >= int(bg.W) || yIdx < 0 || yIdx >= int(bg.H)

	switch {
	case bg.AffineWrap:
		xIdx &= int(bg.W) - 1
		yIdx &= int(bg.H) - 1
	case !bg.AffineWrap && out:
		return 0, false
	}

	map_x := (uint32(xIdx)) & (bg.W - 1) >> 3
	map_y := ((uint32(yIdx)) & (bg.H - 1)) >> 3
	map_y *= bg.W >> 3
	mapIdx := map_y + map_x

	mapAddr := bg.ScreenBaseBlock + mapIdx

    if !engine.IsB {
        mapAddr += engine.Dispcnt.ScreenBase
    }

    tileIdx := uint32(nds.ppu.Vram.Read(vramOffset + mapAddr, true))

	tileAddr := bg.CharBaseBlock + (tileIdx << 6)

	inTileX, inTileY := getPositionsBg(tileIdx, uint32(xIdx), uint32(yIdx))

	inTileIdx := uint32(inTileX) + uint32(inTileY<<3)

	addr := vramOffset + tileAddr + inTileIdx
    palIdx := uint32(nds.ppu.Vram.Read(addr, true))

    pal := uint32(0)

    return getBgPaletteData(nds, engine, bgIdx, true, pal, palIdx, inTileX)
}

func (nds *Nds) setBmpBackgroundPixel(engine *ppu.Engine, bg *ppu.Background, x uint32) (uint32, bool) {

	//if !bg.Palette256 {
	//	panic(fmt.Sprintf("AFFINE WITHOUT PAL 256"))
	//}

	pa := utils.Convert8_8Float(int16(bg.Pa))
	pc := utils.Convert8_8Float(int16(bg.Pc))
	xIdx := int(pa*float64(x) + bg.OutX)
	yIdx := int(pc*float64(x) + bg.OutY)

	if bg.Mosaic && engine.Mosaic.BgH != 0 {
		xIdx -= xIdx % int(engine.Mosaic.BgH+1)
	}

	if bg.Mosaic && engine.Mosaic.BgV != 0 {
		yIdx -= yIdx % int(engine.Mosaic.BgV+1)
	}

	out := xIdx < 0 || xIdx >= int(bg.W) || yIdx < 0 || yIdx >= int(bg.H)

	switch {
	case bg.AffineWrap:
		xIdx &= int(bg.W) - 1
		yIdx &= int(bg.H) - 1
	case !bg.AffineWrap && out:
		return 0, false
	}

    addr := uint32(xIdx+(yIdx * int(bg.W)))
    addr += bg.ScreenBaseBlock * 8

    if engine.IsB {
        addr += 0x20_0000
    }


    palIdx := uint32(nds.ppu.Vram.Read(addr, true))

    if palIdx == 0 { return 0, false }

	data := nds.getPalette(palIdx, 0, false, engine.IsB)

    //if transparent := data & 0x80 != 0; transparent {
    //    return 0, false
    //}

    return data, true
}

func getPositionsBg(screenData, xIdx, yIdx uint32) (uint32, uint32) {

	inTileY := yIdx & 0b111 //% 8
	inTileX := xIdx & 0b111 //% 8

	if hFlip := screenData>>10&1 == 1; hFlip {
		inTileX = 7 - inTileX
	}

	if vFlip := screenData>>11&1 == 1; vFlip {
		inTileY = 7 - inTileY
	}

	return inTileX, inTileY
}

func (nds *Nds) setDirectBitmap(engine *ppu.Engine, bg *ppu.Background, x uint32) (uint32, bool) {

	pa := utils.Convert8_8Float(int16(bg.Pa))
	pc := utils.Convert8_8Float(int16(bg.Pc))
	xIdx := int(pa*float64(x) + bg.OutX)
	yIdx := int(pc*float64(x) + bg.OutY)

	if bg.Mosaic && engine.Mosaic.BgH != 0 {
		xIdx -= xIdx % int(engine.Mosaic.BgH+1)
	}

	if bg.Mosaic && engine.Mosaic.BgV != 0 {
		yIdx -= yIdx % int(engine.Mosaic.BgV+1)
	}

	out := xIdx < 0 || xIdx >= int(bg.W) || yIdx < 0 || yIdx >= int(bg.H)

	switch {
	case bg.AffineWrap:
		xIdx &= int(bg.W) - 1
		yIdx &= int(bg.H) - 1
	case !bg.AffineWrap && out:
		return 0, false
	}

    addr := uint32(xIdx+(yIdx * int(bg.W))) * 2

    addr += bg.ScreenBaseBlock * 8

    var banks uint32
    if engine.IsB {
        addr += 0x20_0000
        banks = ppu.BANKS_B_2D_BG
    } else {
        banks = ppu.BANKS_A_2D_BG
    }

    //palIdx := uint32(nds.ppu.Vram.Read(addr, true))
    data := uint32(nds.ppu.Vram.ReadGraphical(addr, banks))

    // if on transparent??? maybe not
    //if transparent := (data >> 15) & 1 == 0; transparent {
    //    return 0, false
    //}

    return data, true

}

func (nds *Nds) setRawBitmap(engine *ppu.Engine, x, y uint32) (uint32, bool) {

    addr := uint32(x+(y * SCREEN_WIDTH)) * 2

    bankIdx := engine.Dispcnt.VramBlock

    var bank *[0x20000]uint8

    switch bankIdx {
    case 0: bank = &nds.ppu.Vram.A
    case 1: bank = &nds.ppu.Vram.B
    case 2: bank = &nds.ppu.Vram.C
    case 3: bank = &nds.ppu.Vram.D
    }

    //bank = &nds.ppu.Vram.A

    //data := uint32(binary.LittleEndian.Uint16(bank[addr:]) &^ 0x80)

    data := uint32(b16(bank[addr:]))

    return data, true

}

func (nds *Nds) getBgPriority(y uint32, mode uint32, bgs *[4]ppu.Background) [4][]uint32 {

	priorities := [4][]uint32{}

	for i := range 4 {

		if bgs[i].Invalid || !bgs[i].Enabled {
			continue
		}

		if bgNotScanline(&bgs[i], y) {
			continue
		}

        priority := bgs[i].Priority

		priorities[priority] = append(priorities[priority], uint32(i))
	}

	return priorities
}

func (nds *Nds) getObjPriority(y uint32, objects *[128]ppu.Object) [4][]uint32 {

	priorities := [4][]uint32{}

	for i := range 128 {

		obj := &objects[i]

		if disabled := (obj.Disable && !obj.RotScale) || (obj.RotScale && obj.RotParams >= 64); disabled {
			continue
		}

		if objNotScanline(obj, y) {
			continue
		}

		priority := obj.Priority

		priorities[priority] = append(priorities[priority], uint32(i))
	}


	return priorities
}

func objNotScanline(obj *ppu.Object, y uint32) bool {

    const MAX_HEIGHT = 256

	if obj.DoubleSize && obj.RotScale {

		offset := obj.H / 2

		localY := int(y) - int(obj.Y+offset)

		if obj.Y+offset > SCREEN_HEIGHT {
			localY += MAX_HEIGHT
		}

		t := localY+int(offset) < 0
		b := localY-int(obj.H+obj.H+offset) >= 0

		return t || b
	}

	localY := int(y) - int(obj.Y)

	if obj.Y > SCREEN_HEIGHT {
		localY += MAX_HEIGHT
	}

    if aboveTop := localY < 0; aboveTop {
        return true
    }

    if belowBottom := localY >= int(obj.H); belowBottom {
        return true
    }

    return false
}

func bgNotScanline(bg *ppu.Background, y uint32) bool {

	if bg.Affine {
		return false
	}

	localY := (int(y) - int(bg.YOffset)) & int((bg.H)-1)

	t := localY < 0
	b := localY-int(bg.H) >= 0

	return t || b
}

func (nds *Nds) getExtendedPalette(engine *ppu.Engine, bgIdx uint32, obj bool, palIdx, paletteNum uint32) uint32 {

    //16 colors x 16 palettes --> standard palette memory (=256 colors)
    //256 colors x 16 palettes --> extended palette memory (=4096 colors)

	addr := ((paletteNum << 5) + palIdx<<1)
    vram := &nds.ppu.Vram

    // this can probably be replaced in the ppu.
    // ex. vram.ExtABgSlot (unsafe) -> Background[bgIdx].BgSlot (*0x2000uint8)
    // removes un necessary palette slot switches

    switch {
    case !obj && !engine.IsB:

        slotIdx := bgIdx

        if altSlot := engine.Backgrounds[bgIdx].AltExtPalSlot; altSlot {
            switch bgIdx {
            case 0: slotIdx = 2
            case 1: slotIdx = 3
            }
        }

        slot := vram.ExtABgSlots[slotIdx]

        return uint32(b16(slot[addr:]))
    case !obj && engine.IsB:

        slotIdx := bgIdx

        if altSlot := engine.Backgrounds[bgIdx].AltExtPalSlot; altSlot {
            switch bgIdx {
            case 0: slotIdx = 2
            case 1: slotIdx = 3
            }
        }

        slot := vram.ExtBBgSlots[slotIdx]

        return uint32(b16(slot[addr:]))
    case obj && !engine.IsB:
        return uint32(b16(vram.ExtAPalObj[addr:]))
    case obj && engine.IsB:
        return uint32(b16(vram.ExtBPalObj[addr:]))
    }

    return 0
}

func (nds *Nds) getPalette(palIdx uint32, paletteNum uint32, obj, engineB bool) uint32 {

	addr := (paletteNum << 5) + palIdx<<1

	if obj {
		addr += 0x200
	}

    if engineB {
		addr += 0x400
    }

    addr >>= 1

    if addr >= 0x400 {
        fmt.Printf("BAD PAL ADDR %04X (> 0x400), isB %t\n", addr, engineB)
        return 0x0 //FFFF
    }

	return uint32(nds.ppu.Pram[addr])
}

func getPositions(obj *ppu.Object, xIdx, yIdx uint32) (uint32, uint32, uint32, uint32) {

	enTileY := yIdx >> 3    // / 8
	enTileX := xIdx >> 3    // / 8
	inTileY := yIdx & 0b111 // % 8
	inTileX := xIdx & 0b111 // % 8

	if obj.RotScale {
		return enTileX, enTileY, inTileX, inTileY
	}

	if obj.HFlip {
		enTileX = (obj.W / 8) - 1 - enTileX
		inTileX = 7 - inTileX
	}
	if obj.VFlip {
		enTileY = (obj.H / 8) - 1 - enTileY
		inTileY = 7 - inTileY
	}

	return enTileX, enTileY, inTileX, inTileY
}

func getObjTileAddr(obj *ppu.Object, enTileX, enTileY, inTileX, inTileY uint32) uint32 {

    const (
        BYTES_PER_PIXEL = 2
	    //MAX_NUM_TILE = 1024
	    //MAX_TILE_MASK = MAX_NUM_TILE - 1 
    )

	w := obj.W << BYTES_PER_PIXEL

	if obj.Palette256 {
		enTileX <<= 1
		w <<= 1
	}

	var tileIdx uint32
	if obj.OneDimensional {
		tileIdx = (enTileX << 5) + (enTileY * w)
		tileIdx = (tileIdx + obj.CharName << obj.TileBoundaryShift) //& MAX_TILE_MASK
	} else {
        tileIdx = enTileX + (enTileY << 5)
		tileIdx = (tileIdx + obj.CharName/*& MAX_TILE_MASK*/) << 5
	}

	tileAddr := uint32(tileIdx)

	var inTileIdx uint32
	if obj.Palette256 {
		inTileIdx = inTileX + (inTileY << 3)
	} else {
		inTileIdx = (inTileX >> 1) + (inTileY << 2)
	}

	return tileAddr + inTileIdx
}

func getBmpTileAddr(obj *ppu.Object, xIdx, yIdx uint32) uint32 {

    const BYTES_PER_PIXEL = 2

	if obj.OneDimensional {
        return uint32(xIdx+(yIdx << obj.BmpBoundaryShift)) * BYTES_PER_PIXEL
	}

    maskX := obj.BmpBoundaryMask
    base := ((obj.CharName & maskX) << 4) + ((obj.CharName & ^maskX) << 7)

    var pixelOffset uint32
    if obj.ObjBmpMapping == ppu.OBJ_BMP_128_2D {
        pixelOffset = (yIdx*(obj.W << 1) + xIdx) * BYTES_PER_PIXEL
    } else {
        pixelOffset = (yIdx*(obj.W << 2) + xIdx) * BYTES_PER_PIXEL
    }

    return base + pixelOffset
}

func getBgPaletteData(nds *Nds, engine *ppu.Engine, bgIdx uint32, pal256 bool, palNum, tileData, inTileX uint32) (uint32, bool) {
    //if !engine.IsB && x == y && x == 0 {
    //    fmt.Printf("PAL IDX %0X PAL NUM %08X\n", palIdx, palNum)
    //}

	var palIdx uint32
	if pal256 {
		palIdx = tileData & 0xFF
	} else {
		palIdx = (tileData >> ((inTileX & 1) << 2)) & 0xF
	}

	if palIdx == 0 {
		return 0, false
	}

    if engine.Dispcnt.BgExtPal && pal256 {

        palNum <<= 4
        palData := nds.getExtendedPalette(engine, bgIdx, false, palIdx, palNum)
        return palData, true
    }

    // this is from gba, does not work with ext palettes, but assume it still
    // is needed for std
    if pal256 {
        palNum = 0
    }

	palData := nds.getPalette(uint32(palIdx), palNum, false, engine.IsB)

	return palData, true
}

func getPaletteData(nds *Nds, engine *ppu.Engine, pal256 bool, pal, tileData, inTileX uint32) (uint32, bool) {

	var palIdx uint32
	if pal256 {
		palIdx = tileData & 0xFF
	} else {
		palIdx = (tileData >> ((inTileX & 1) << 2)) & 0xF
	}

    if palIdx == 0 {
        return 0, false
    }

    if engine.Dispcnt.ObjExtPal && pal256 {
        pal <<= 4
        palData := nds.getExtendedPalette(engine, 0, true, palIdx, pal)
        return palData, true
    }
	if palIdx == 0 {
		return 0, false
	}

    // this is from gba, does not work with ext palettes, but assume it still
    // is needed for std
    if pal256 {
        pal = 0
    }

	palData := nds.getPalette(uint32(palIdx), pal, true, engine.IsB)

	return palData, true
}

func outObjectBound(obj *ppu.Object, xIdx, yIdx int) bool {
	t := yIdx < 0
	b := yIdx-int(obj.H) >= 0
	l := xIdx < 0
	r := xIdx-int(obj.W) >= 0
	return t || b || l || r
}

func (nds *Nds) outBoundsAffine(obj *ppu.Object, x, y uint32) bool {

	const (
		MAX_X_MASK = 511
		MAX_Y_MASK = 255
	)

    var t, b, l, r uint32

	if !obj.DoubleSize {

		t = obj.Y
		b = (obj.Y + obj.H) & MAX_Y_MASK
		l = obj.X
		r = (obj.X + obj.W) & MAX_X_MASK

    } else {

        // obj.Y is double Sized Y value already, have to adj because of

        dY := (obj.Y)
        dH := obj.H * 2
        dX := (obj.X)
        dW := obj.W * 2

        t = dY
        b = (dY + dH) & MAX_Y_MASK
        l = dX
        r = (dX + dW) & MAX_X_MASK
    }

	yWrapped := t > b
	xWrapped := l > r
	yWrappedInBounds := !yWrapped && (y >= t && y < b)
	yUnwrappedInBounds := yWrapped && (y >= t || y < b)
	xWrappedInBounds := !xWrapped && (x >= l && x < r)
	xUnwrappedInBounds := xWrapped && (x >= l || x < r)
	return !(
        (yWrappedInBounds || yUnwrappedInBounds) &&
        (xWrappedInBounds || xUnwrappedInBounds))
}

func (nds *Nds) setObjTilePixel(engine *ppu.Engine, obj *ppu.Object, x, y uint32) (uint32, bool) {

    xIdx, yIdx, ok := int(0), int(0), false
    if obj.RotScale {
        xIdx, yIdx, ok = nds.getObjAffineCoords(engine, obj, x, y)
    } else {
        xIdx, yIdx, ok = nds.getObjNormalCoords(engine, obj, x, y)
    }

    if !ok {
        return 0, false
    }

	enTileX, enTileY, inTileX, inTileY := getPositions(obj, uint32(xIdx), uint32(yIdx))

	addr := getObjTileAddr(obj, enTileX, enTileY, inTileX, inTileY)

    vramOffset := uint32(0x40_0000)
    var banks uint32
    if engine.IsB {
        vramOffset = uint32(0x60_0000)
        banks = ppu.BANKS_B_2D_OBJ
    } else {
        banks = ppu.BANKS_A_2D_OBJ
    }

    tileData := uint32(nds.ppu.Vram.ReadGraphical(vramOffset + addr, banks))

	return getPaletteData(nds, engine, obj.Palette256, obj.Palette, tileData, uint32(inTileX))
}

func (nds *Nds) setObjBmpPixel(engine *ppu.Engine, obj *ppu.Object, x, y uint32) (uint32, bool) {

    xIdx, yIdx, ok := int(0), int(0), false
    if obj.RotScale {
        xIdx, yIdx, ok = nds.getObjAffineCoords(engine, obj, x, y)
    } else {
        xIdx, yIdx, ok = nds.getObjNormalCoords(engine, obj, x, y)
    }

    if !ok {
        return 0, false
    }

	addr := getBmpTileAddr(obj, uint32(xIdx), uint32(yIdx))

    vramOffset := uint32(0x40_0000)
    var banks uint32
    if engine.IsB {
        vramOffset = uint32(0x60_0000)
        banks = ppu.BANKS_B_2D_OBJ
    } else {
        banks = ppu.BANKS_A_2D_OBJ
    }

    data := uint32(nds.ppu.Vram.ReadGraphical(vramOffset + addr, banks))

    if alpha := (data & 0x8000) == 0; alpha {
        return 0, false
    }

    return data, true
}

func (nds *Nds) getObjAffineCoords(engine *ppu.Engine, obj *ppu.Object, x, y uint32) (int, int, bool) {

        if nds.outBoundsAffine(obj, x, y) {
            return 0, 0, false
        }

        xIdx, yIdx := nds.getAffineCoordinates(engine, obj, x, y)

        if outObjectBound(obj, xIdx, yIdx) {
            return 0, 0, false
        }

        return xIdx, yIdx, true
}

func (nds *Nds) getObjNormalCoords(engine *ppu.Engine, obj *ppu.Object, x, y uint32) (int, int, bool) {

    yIdx := int(y) - int(obj.Y)
    xIdx := int(x) - int(obj.X)

    if obj.Y > SCREEN_HEIGHT {
        yIdx += 256 // i believe 256 is max
    }

    if obj.X > SCREEN_WIDTH {
        xIdx += 512 // i believe 512 is max
    }

    if outObjectBound(obj, xIdx, yIdx) {
        return 0, 0, false
    }

    if obj.Mosaic && engine.Mosaic.ObjH != 0 {
        xIdx -= xIdx % int(engine.Mosaic.ObjH+1)
    }

    if obj.Mosaic && engine.Mosaic.ObjV != 0 {
        yIdx -= yIdx % int(engine.Mosaic.ObjV+1)
    }

    return xIdx, yIdx, true
}

func (nds *Nds) getAffineCoordinates(engine *ppu.Engine, obj *ppu.Object, x, y uint32) (int, int) {

	objX := obj.X
	objY := obj.Y
	if obj.DoubleSize {
		objX += obj.W / 2
		objY += obj.H / 2
	}

	xIdx := int(float32(x) - float32(objX))
	yIdx := int(float32(y)-float32(objY)) % 256

	if objY > SCREEN_HEIGHT {
		yIdx += 256 // i believe 256 is max
	}
	if objX > SCREEN_WIDTH {
		xIdx += 512 // i believe 512 is max
	}

	if obj.Mosaic && engine.Mosaic.ObjH != 0 {
		xIdx -= xIdx % int(engine.Mosaic.ObjH+1)
	}

	if obj.Mosaic && engine.Mosaic.ObjV != 0 {
		yIdx -= yIdx % int(engine.Mosaic.ObjV+1)
	}

	xOrigin := float32(xIdx - (int(obj.W) / 2))
	yOrigin := float32(yIdx - (int(obj.H) / 2))

	xIdx = int(obj.Pa*xOrigin+obj.Pb*yOrigin) + (int(obj.W) / 2)
	yIdx = int(obj.Pc*xOrigin+obj.Pd*yOrigin) + (int(obj.H) / 2)

    return xIdx, yIdx
}
