package nds

import (
	"encoding/binary"
	"sync"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/nds/ppu"
	"github.com/aabalke/guac/emu/nds/utils"
)
var wg = sync.WaitGroup{}

func (nds *Nds) graphics(y uint32) {

    a := &nds.ppu.EngineA
    b := &nds.ppu.EngineB

	switch a.Dispcnt.DisplayMode {
	case 0:
		nds.screenoff(y, 0xFF, a)
	case 1:
		nds.standard(y, a)
	case 2:
		nds.vramDisplay(y, a)
	case 3:
		panic("UNSETUP MAIN MEMORY DISPLAY")
	}

    switch b.Dispcnt.DisplayMode {
    case 0:
        nds.screenoff(y, 0xFF, b)
    case 1:
        nds.standard(y, b)
    }
}

func (nds *Nds) screenoff(y uint32, colorByte uint8, engine *ppu.Engine) {

	x := uint32(0)
	for x = range SCREEN_WIDTH {

		index := (x + (y * SCREEN_WIDTH)) * 4
		(*engine.Pixels)[index]   = colorByte
		(*engine.Pixels)[index+1] = colorByte
		(*engine.Pixels)[index+2] = colorByte
		(*engine.Pixels)[index+3] = 0xFF
	}
}

func (nds *Nds) vramDisplay(y uint32, engine *ppu.Engine) {

	x := uint32(0)
	for x = range SCREEN_WIDTH {

        bankIdx := engine.Dispcnt.VramBlock

        var bank *[0x20000]uint8

        switch bankIdx {
        case 0: bank = &nds.mem.Vram.A
        case 1: bank = &nds.mem.Vram.B
        case 2: bank = &nds.mem.Vram.C
        case 3: bank = &nds.mem.Vram.D
        }

        offset := uint32(0)

        palData, ok := nds.directbitmap(x, y, offset, bank)

        if ok {
            index := (x + (y * SCREEN_WIDTH)) * 4
            nds.applyColor(palData, index, engine.Pixels)
        }
    }
}

func (nds *Nds) standard(y uint32, engine *ppu.Engine) {

	if config.Conf.Gba.Threads == 0 {

		x := uint32(0)
		for x = range SCREEN_WIDTH {
			nds.render(x, y, engine)
		}

		return
	}

	WAIT_GROUPS := config.Conf.Gba.Threads
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

			//if !windowPixelAllowed(bgIdx, x, y, wins) {
			//	continue
			//}

            palData, ok := uint32(0), false

            switch bg.Type {
            case ppu.BG_TYPE_TEX:
                palData, ok = nds.setBackgroundPixel(engine, bg, x, y)
            case ppu.BG_TYPE_AFF:
				palData, ok = nds.setAffineBackgroundPixel(engine, bg, x)
            case ppu.BG_TYPE_LAR:
                palData, ok = 0b11111 << 10, true // blue
            case ppu.BG_TYPE_3D :
                palData, ok = 0b11111, true // red
            case ppu.BG_TYPE_BGM:
                palData, ok = 0b1111111111, true // yellow
            case ppu.BG_TYPE_256:
                palData, ok = nds.largeBitmap(x, y)
            case ppu.BG_TYPE_DIR:

                offset := bg.ScreenBaseBlock * 8
                if engine.IsB {
                    offset += 0x20_0000
                }

                palData, ok = nds.directbitmap(x, y, offset, nil)
            }

			if ok {

				bldPal.SetBlendPalettes(palData, uint32(bgIdx), false, false)
			}
		}

		if objDisabled := !dispcnt.DisplayObj; objDisabled {
			continue
		}

        ObjectLoop:
		for j := len(objPriorities[i]) - 1; j >= 0; j-- {
			objIdx := objPriorities[i][j]
	        obj := &engine.Objects[objIdx]

            obj.OneDimensional = dispcnt.TileObj1D
			//if !windowObjPixelAllowed(x, y, wins) {
			//	continue
			//}

			var palData uint32
			var ok bool

            switch {
            case obj.Mode == 3:
                palData, ok = nds.setBmpObjectAffinePixel(engine, obj, x, y)
            case obj.RotScale:
                palData, ok = nds.setObjectAffinePixel(engine, obj, x, y)
            default:
                palData, ok = nds.setObjectPixel(engine, obj, x, y)
            }

			switch {
			case ok && obj.Mode == 2:
				inObjWindow = true
				break ObjectLoop
			case ok:
				objMode = obj.Mode
				bldPal.SetBlendPalettes(palData, 0, true, objMode == 1)
				break ObjectLoop
			}
		}
	}

	finalPalData := bldPal.Blend(objMode == 1, x, y, wins, inObjWindow)
	index := (x + (y * SCREEN_WIDTH)) << 2
	nds.applyColor(finalPalData, uint32(index), engine.Pixels)
}

func (nds *Nds) largeBitmap(x, y uint32) (uint32, bool) {

    // this will need affine support

    const (
        BYTE_PER_PIXEL = 1
    )

    idx := (x+(y*SCREEN_WIDTH))*BYTE_PER_PIXEL

    palIdx := uint32(nds.mem.Vram.Read(idx, true))

    data := uint32(nds.mem.Pram[palIdx])

    return data, true
}

func updateBackgrounds(engine *ppu.Engine) *[4]ppu.Background {

    bgs := &engine.Backgrounds

    getExtended := func(bg *ppu.Background) uint8 {

        if !bg.Palette256 {
            return ppu.BG_TYPE_BGM // what does lsb char block do? text vs affine?
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

func (nds *Nds) setBackgroundPixel(engine *ppu.Engine, bg *ppu.Background, x, y uint32) (uint32, bool) {

    vramOffset := uint32(0)
    pramOffset := uint32(0)

    if engine.IsB {
        vramOffset = uint32(0x20_0000)
        pramOffset = uint32(0x400)
    }

	xIdx := (x + bg.XOffset) & ((bg.W) - 1)
	yIdx := (y + bg.YOffset) & ((bg.H) - 1)

	if bg.Mosaic && engine.Mosaic.BgH != 0 {
		xIdx -= xIdx % (engine.Mosaic.BgH + 1)
	}

	if bg.Mosaic && engine.Mosaic.BgV != 0 {
		yIdx -= yIdx % (engine.Mosaic.BgV + 1)
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

    if !engine.IsB {
        mapAddr += engine.Dispcnt.ScreenBase
    }

    screenData := uint32(nds.mem.Vram.Read(vramOffset + mapAddr, true))
    screenData |= uint32(nds.mem.Vram.Read(vramOffset + mapAddr + 1, true)) << 8

	tileIdx := (screenData & 0b11_1111_1111) << 5

	tileAddr := bg.CharBaseBlock + tileIdx
	if bg.Palette256 {
		tileAddr += tileIdx
	}

    if !engine.IsB {
        tileAddr += engine.Dispcnt.CharBase
    }

	inTileX, inTileY := getPositionsBg(screenData, xIdx, yIdx)

	var inTileIdx uint32
	if bg.Palette256 {
		inTileIdx = inTileX + (inTileY << 3)
	} else {
		inTileIdx = (inTileX >> 1) + (inTileY << 2)
	}

    tileData := uint32(nds.mem.Vram.Read(vramOffset + tileAddr + inTileIdx, true))

	if bg.Palette256 {
		palIdx := tileData
		if palIdx == 0 {
			return 0, false
		}

        addr := (palIdx << 1) + pramOffset

        return uint32(nds.mem.Pram[addr >> 1]), true
	}

	palIdx := (tileData >> ((inTileX & 1) << 2)) & 0xF

	if palIdx == 0 {
		return 0, false
	}

	palette := screenData >> 12
	addr := ((palette << 5) + palIdx<<1) + pramOffset

	return uint32(nds.mem.Pram[addr >> 1]), true
}

func (nds *Nds) setAffineBackgroundPixel(engine *ppu.Engine, bg *ppu.Background, x uint32) (uint32, bool) {

	//if !bg.Palette256 {
	//	panic(fmt.Sprintf("AFFINE WITHOUT PAL 256"))
	//}

    vramOffset := uint32(0)
    //pramOffset := uint32(0)

    if engine.IsB {
        vramOffset = uint32(0x20_0000)
        //pramOffset = uint32(0x400)
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

    tileIdx := uint32(nds.mem.Vram.Read(vramOffset + mapAddr, true))

	tileAddr := bg.CharBaseBlock + (tileIdx << 6)

	inTileX, inTileY := getPositionsBg(tileIdx, uint32(xIdx), uint32(yIdx))

	inTileIdx := uint32(inTileX) + uint32(inTileY<<3)

	addr := vramOffset + tileAddr + inTileIdx
    palIdx := uint32(nds.mem.Vram.Read(addr, true))

	if palIdx == 0 {
		return 0, false
	}

	palData := nds.getPalette(palIdx, 0, false, engine.IsB)

	return palData, true
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

func (nds *Nds) directbitmap(x, y, offset uint32, bank *[0x20000]uint8) (uint32, bool) {

    // will need affine support

    const (
        BYTE_PER_PIXEL = 2
    )


    var data uint32

    if bank == nil {
        idx := ((x+(y*SCREEN_WIDTH)) * BYTE_PER_PIXEL) + offset
        data = uint32(nds.mem.Vram.Read(idx, true))
        data |= uint32(nds.mem.Vram.Read(idx + 1, true)) << 8
    } else {
        idx := ((x+(y*SCREEN_WIDTH)) * BYTE_PER_PIXEL)
        data = uint32(bank[idx])
        data |= uint32(bank[idx + 1]) << 8
    }

    if transparent := (data >> 16) & 1 == 1; transparent {
        return 0, false
    }

    return data, true
}

func (nds *Nds) DebugPalette(y, addr uint32, pixels *[]byte) {

	x := uint32(0)
	for x = range SCREEN_WIDTH {

		index := (x + (y * SCREEN_WIDTH)) * 4

		data := uint32(nds.mem.Pram[addr >> 1])

		nds.applyColor(data, index, pixels)
	}
}

func (nds *Nds) applyColor(data, i uint32, pixels *[]byte) {
	r := uint8((data) & 0b11111)
	g := uint8((data >> 5) & 0b11111)
	b := uint8((data >> 10) & 0b11111)

	r = (r << 3) | (r >> 2)
	g = (g << 3) | (g >> 2)
	b = (b << 3) | (b >> 2)

	(*pixels)[i] = r
	(*pixels)[i+1] = g
	(*pixels)[i+2] = b
	(*pixels)[i+3] = 0xFF
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

	//added := false
	//highestPriority := uint32(5)

	for i := range 128 {

		obj := &objects[i]

		if disabled := (obj.Disable && !obj.RotScale) || (obj.RotScale && obj.RotParams >= 32); disabled {
			continue
		}

		//if objNotScanline(obj, y) {
		//	continue
		//}

		priority := obj.Priority

		//if gba.PPU.Blend.Mode != BLD_MODE_STD && added && priority >= highestPriority {
		//	continue
		//}

		//added = true

		priorities[priority] = append(priorities[priority], uint32(i))
	}

	return priorities
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

func (nds *Nds) getExtendedPalette(engine *ppu.Engine, obj bool, palIdx, paletteNum uint32) uint32 {

    //16 colors x 16 palettes --> standard palette memory (=256 colors)
    //256 colors x 16 palettes --> extended palette memory (=4096 colors)

	addr := ((paletteNum << 5) + palIdx<<1)

    //addr &= 0x1FFF

    vram := &nds.mem.Vram

    switch {
    case !obj && !engine.IsB:
        return uint32(binary.LittleEndian.Uint16(vram.ExtAPalBg[addr:]))
    case !obj && engine.IsB:
        return uint32(binary.LittleEndian.Uint16(vram.ExtBPalBg[addr:]))
    case obj && !engine.IsB:
        return uint32(binary.LittleEndian.Uint16(vram.ExtAPalObj[addr:]))
    case obj && engine.IsB:
        return uint32(binary.LittleEndian.Uint16(vram.ExtBPalObj[addr:]))
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

	return uint32(nds.mem.Pram[addr>>1])
}

func (nds *Nds) setObjectPixel(engine *ppu.Engine, obj *ppu.Object, x, y uint32) (uint32, bool) {

    vramOffset := uint32(0x40_0000)
    if engine.IsB {
        vramOffset = uint32(0x60_0000)
    }
    //pramOffset := uint32(0)

	//mem := &nds.mem

	yIdx := int(y) - int(obj.Y)
	xIdx := int(x) - int(obj.X)

	if obj.Y > SCREEN_HEIGHT {
		yIdx += 256 // i believe 256 is max
	}

	if obj.X > SCREEN_WIDTH {
		xIdx += 512 // i believe 512 is max
	}

	if outObjectBound(obj, xIdx, yIdx) {
		return 0, false
	}

	if obj.Mosaic && engine.Mosaic.ObjH != 0 {
		xIdx -= xIdx % int(engine.Mosaic.ObjH+1)
	}

	if obj.Mosaic && engine.Mosaic.ObjV != 0 {
		yIdx -= yIdx % int(engine.Mosaic.ObjV+1)
	}

	enTileX, enTileY, inTileX, inTileY := getPositions(obj, uint32(xIdx), uint32(yIdx))

	addr := getObjTileAddr(obj, enTileX, enTileY, inTileX, inTileY)

    tileData := uint32(nds.mem.Vram.Read(vramOffset + addr, true))
    tileData |= uint32(nds.mem.Vram.Read(vramOffset + addr + 1, true)) << 8

	return getPaletteData(nds, engine, obj.Palette256, obj.Palette, tileData, uint32(inTileX))

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

    const BYTES_PER_PIXEL = 2
	w := obj.W << BYTES_PER_PIXEL

	if obj.Palette256 {
		enTileX <<= 1
		w <<= 1
	}

	//const MAX_NUM_TILE = 1024
	//const MAX_TILE_MASK = MAX_NUM_TILE - 1 
	var tileIdx uint32
	if obj.OneDimensional {
		//tileIdx = (enTileX << 5) + (enTileY * w)
		//tileIdx = (tileIdx + obj.CharName << 5) & MAX_TILE_MASK
		tileIdx = (enTileX << 5) + (enTileY * w)
		tileIdx = (tileIdx + obj.CharName << obj.TileBoundaryShift) //& MAX_TILE_MASK
	} else {
		tileIdx = enTileX + (enTileY << w)
		//tileIdx = (tileIdx + obj.CharName & MAX_TILE_MASK) << 5
		tileIdx = (tileIdx + obj.CharName) << obj.TileBoundaryShift
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

func getBmpTileAddr(obj *ppu.Object, xIdx, yIdx int) uint32 {

    const BYTES_PER_PIXEL = 2

	if obj.OneDimensional {
        return uint32((xIdx+(yIdx << int(obj.BmpBoundaryShift))) * BYTES_PER_PIXEL)
	}

    panic("OBJ BMP 2D, make sure addr calc accurate")

    return uint32((xIdx+(yIdx << int(obj.BmpBoundaryShift))) * BYTES_PER_PIXEL)
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

    if engine.Dispcnt.ObjExtPal {
        // palette is 256 each

        pal <<= 4
        palData := nds.getExtendedPalette(engine, true, palIdx, pal)
        return palData, true
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

func (nds *Nds) setObjectAffinePixel(engine *ppu.Engine, obj *ppu.Object, x, y uint32) (uint32, bool) {

	if nds.outBoundsAffine(obj, x, y) {
		return 0, false
	}

    xIdx, yIdx := nds.getAffineCoordinates(engine, obj, x, y)

	if outObjectBound(obj, xIdx, yIdx) {
		return 0, false
	}

	enTileX, enTileY, inTileX, inTileY := getPositions(obj, uint32(xIdx), uint32(yIdx))

	addr := getObjTileAddr(obj, enTileX, enTileY, inTileX, inTileY)

    vramOffset := uint32(0x40_0000)
    if engine.IsB {
        vramOffset = uint32(0x60_0000)
    }

    tileData := uint32(nds.mem.Vram.Read(vramOffset + addr, true))
    tileData |= uint32(nds.mem.Vram.Read(vramOffset + addr + 1, true)) << 8

	return getPaletteData(nds, engine, obj.Palette256, obj.Palette, tileData, uint32(inTileX))
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

func (nds *Nds) outBoundsAffine(obj *ppu.Object, x, y uint32) bool {

	const (
		MAX_X_MASK = 511
		MAX_Y_MASK = 255
	)

	if !obj.DoubleSize {

		t := obj.Y
		b := (obj.Y + obj.H) & MAX_Y_MASK
		l := obj.X
		r := (obj.X + obj.W) & MAX_X_MASK

		yWrapped := t > b
		xWrapped := l > r

		yWrappedInBounds := !yWrapped && (y >= t && y < b)
		yUnwrappedInBounds := yWrapped && (y >= t || y < b)
		xWrappedInBounds := !xWrapped && (x >= l && x < r)
		xUnwrappedInBounds := xWrapped && (x >= l || x < r)
		if (yWrappedInBounds || yUnwrappedInBounds) && (xWrappedInBounds || xUnwrappedInBounds) {
			return false
		}

		return true
	}

	// obj.Y is double Sized Y value already, have to adj because of

	dY := (obj.Y)
	dH := obj.H * 2
	dX := (obj.X)
	dW := obj.W * 2

	t := dY
	b := (dY + dH) & MAX_Y_MASK
	l := dX
	r := (dX + dW) & MAX_X_MASK

	yWrapped := t > b
	xWrapped := l > r

	yWrappedInBounds := !yWrapped && (y >= t && y < b)
	yUnwrappedInBounds := yWrapped && (y >= t || y < b)

	xWrappedInBounds := !xWrapped && (x >= l && x < r)
	xUnwrappedInBounds := xWrapped && (x >= l || x < r)
	if (yWrappedInBounds || yUnwrappedInBounds) && (xWrappedInBounds || xUnwrappedInBounds) {
		return false
	}

	return true
}

func (nds *Nds) setBmpObjectAffinePixel(engine *ppu.Engine, obj *ppu.Object, x, y uint32) (uint32, bool) {

    if obj.Palette256 {
        panic("BITMAP AND PAL 256")
    }

    xIdx, yIdx := nds.getAffineCoordinates(engine, obj, x, y)

	if outObjectBound(obj, xIdx, yIdx) {
		return 0, false
	}

	addr := getBmpTileAddr(obj, xIdx, yIdx)

    vramOffset := uint32(0x40_0000)
    if engine.IsB {
        vramOffset = uint32(0x60_0000)
    }

    data := uint32(nds.mem.Vram.Read(vramOffset + addr, true))
    data |= uint32(nds.mem.Vram.Read(vramOffset + addr + 1, true)) << 8

    if alpha := (data & 0x8000) == 0; alpha {
        return 0, false
    }

    data &^= 0x8000

    return data, true
}
