package nds

import (
	"sync"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/nds/ppu"
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
	//objPriorities := &engine.ObjPriorities
	bgPriorities := &engine.BgPriorities

	bldPal := ppu.NewBlendPalette(x, &engine.Blend, nds.getPalette(0, 0, false))

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
			//	palData, ok = gba.setAffineBackgroundPixel(bg, x)
                palData, ok = 0b11111 << 5, true // green
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

		//for j := len(objPriorities[i]) - 1; j >= 0; j-- {
		//	objIdx := objPriorities[i][j]
		//	obj := &gba.PPU.Objects[objIdx]
		//	obj.OneDimensional = dispcnt.OneDimensional

		//	if !windowObjPixelAllowed(x, y, wins) {
		//		continue
		//	}

		//	var palData uint32
		//	var ok bool

		//	if obj.RotScale {
		//		palData, ok = gba.setObjectAffinePixel(obj, x, y)
		//	} else {
		//		palData, ok = gba.setObjectPixel(obj, x, y)
		//	}

		//	switch {
		//	case ok && obj.Mode == 2:
		//		inObjWindow = true
		//		break
		//	case ok:
		//		objMode = obj.Mode
		//		bldPal.setBlendPalettes(palData, 0, true, objMode == 1)
		//		break
		//	}
		//}
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
    //dispcnt := &engine.Dispcnt

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


		//isAffine := ((dispcnt.Mode == 1 && i == 2) ||
		//	(dispcnt.Mode == 2 && (i == 2 || i == 3)))
		//isStandard := ((dispcnt.Mode == 0) ||
		//	(dispcnt.Mode == 1 && (i == 0 || i == 1 || i == 2)))

		//bgs[i].Invalid = !isAffine && !isStandard
		//bgs[i].Affine = isAffine

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

        return uint32(nds.mem.Pram[palIdx]), true
	}

	palIdx := (tileData >> ((inTileX & 1) << 2)) & 0xF

	if palIdx == 0 {
		return 0, false
	}

	palette := screenData >> 12
	addr := ((palette << 5) + palIdx<<1) + pramOffset

	return uint32(nds.mem.Pram[addr >> 1]), true
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

func bgNotScanline(bg *ppu.Background, y uint32) bool {

	//if bg.Affine {
	//	return false
	//}

	localY := (int(y) - int(bg.YOffset)) & int((bg.H)-1)

	t := localY < 0
	b := localY-int(bg.H) >= 0

	return t || b
}

func (nds *Nds) getPalette(palIdx uint32, paletteNum uint32, obj bool) uint32 {

	addr := (paletteNum << 5) + palIdx<<1

	if obj {
		addr += 0x200
	}

	return uint32(nds.mem.Pram[addr>>1])
}
