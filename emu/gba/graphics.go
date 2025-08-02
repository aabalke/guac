package gba

import (
	"fmt"
	"sync"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/gba/utils"
)

const (
	MAX_HEIGHT = 256
	MAX_WIDTH  = 512
)

func (bg *Background) BgAffineReset() {
	bg.OutX = convert20_8Float(int32(bg.aXOffset))
	bg.OutY = convert20_8Float(int32(bg.aYOffset))
}

func (bg *Background) BgAffineUpdate() {
	bg.OutX += convert8_8Float(int16(bg.Pb))
	bg.OutY += convert8_8Float(int16(bg.Pd))
}

func updateBackgrounds(gba *GBA, dispcnt *Dispcnt) *[4]Background {

	bgs := &gba.PPU.Backgrounds

	for i := range 4 {
		isAffine := ((dispcnt.Mode == 1 && i == 2) ||
			(dispcnt.Mode == 2 && (i == 2 || i == 3)))
		isStandard := ((dispcnt.Mode == 0) ||
			(dispcnt.Mode == 1 && (i == 0 || i == 1 || i == 2)))

		bgs[i].Invalid = !isAffine && !isStandard
		bgs[i].Affine = isAffine

		bgs[i].setSize()

		if (dispcnt.Mode == 1 && i == 2) || dispcnt.Mode == 2 {
			bgs[i].Palette256 = true
		}
	}

	return bgs
}

func (gba *GBA) scanlineGraphics(y uint32) {

	if y >= 160 {
		return
	}

	if gba.PPU.Dispcnt.ForcedBlank {
		x := uint32(0)
		for x = range SCREEN_WIDTH {
			index := (x + (y * SCREEN_WIDTH)) * 4
			(gba.Pixels)[index] = 0xFF
			(gba.Pixels)[index+1] = 0xFF
			(gba.Pixels)[index+2] = 0xFF
			(gba.Pixels)[index+3] = 0xFF
		}
	}

	switch gba.PPU.Dispcnt.Mode {
	case 0, 1, 2:
		gba.scanlineTileMode(y)
	case 3, 4, 5:
		gba.scanlineBitmapMode(y)
	default:
		panic("UNKNOWN MODE")
	}
}

func (gba *GBA) scanlineTileMode(y uint32) {

	wg := sync.WaitGroup{}

    bgPriorities := &gba.PPU.BackgroundPriorities
	objPriorities := gba.getObjPriority(y, &gba.PPU.Objects)

	dispcnt := &gba.PPU.Dispcnt

	bgs := updateBackgrounds(gba, dispcnt)
	wins := &gba.PPU.Windows

	if dispcnt.Mode >= 3 {
		return
	}

	renderPixel := func(x uint32) {


        bldPal := &gba.PPU.Blend.BlendPalettes[x]
		bldPal.reset(gba)

		var objMode uint32
		var inObjWindow bool

		for i := range 4 {

			// 0 is highest priority
			decIdx := 3 - i

			bgCount := len(bgPriorities[decIdx])
			for j := range bgCount {

				// need bgCount - 1 - j because of blends
				bgIdx := bgPriorities[decIdx][bgCount-1-j]

				bg := &bgs[bgIdx]
				if bg.Invalid || !bg.Enabled {
					continue
				}

				palData, ok := gba.background(x, y, bgIdx, bg, wins)

				if ok {
					bldPal.setBlendPalettes(palData, uint32(bgIdx), false, false)
                    break
				}
			}


			if objects := dispcnt.DisplayObj; objects {
				objPal := uint32(0)
				objExists := false
				objCount := len(objPriorities[decIdx])

				for j := range objCount {
					objIdx := objPriorities[decIdx][j]
                    obj := &gba.PPU.Objects[objIdx]

                    if disabledStd := obj.Disable && !obj.RotScale; disabledStd {
                        continue
                    }

					palData, ok := gba.object(x, y, obj, wins, dispcnt)

					if ok {
						if obj.Mode == 2 {
							inObjWindow = true
                            break
							// break here too? idk
						} else {
							objMode = obj.Mode
							objExists = true
							objPal = palData
							break
						}
					}
				}

				if objExists {
					bldPal.setBlendPalettes(objPal, 0, true, objMode == 1)
				}
			}
		}

		finalPalData := bldPal.blend(objMode, x, y, wins, inObjWindow)
		index := uint32(x + (y * SCREEN_WIDTH)) << 2
		gba.applyColor(finalPalData, index)
	}

    if config.Conf.Gba.Threads == 0 {
        for j := range SCREEN_WIDTH {
            renderPixel(uint32(j))
        }
        return
    }

    dx := SCREEN_WIDTH / config.Conf.Gba.Threads

	for i := range config.Conf.Gba.Threads {

		wg.Add(1)

		go func(i int) {

			defer wg.Done()

			for j := range dx {
				renderPixel(uint32((i * dx) + j))
			}
		}(i)
	}

	wg.Wait()
}

func (gba *GBA) scanlineBitmapMode(y uint32) {

	wg := sync.WaitGroup{}
	mem := &gba.Mem
	dispcnt := &gba.PPU.Dispcnt

	objPriorities := gba.getObjPriority(y, &gba.PPU.Objects)

	wins := &gba.PPU.Windows

	if dispcnt.Mode < 3 {
		return
	}

	renderPixel := func(x uint32) {

        bldPal := &gba.PPU.Blend.BlendPalettes[x]
		index := (x + (y * SCREEN_WIDTH)) * 4
		bldPal.reset(gba)

		var objMode uint32
		var inObjWindow bool

		BG_IDX := uint32(2)
		DEC_IDX := uint32(0) // this will have to be updated

		switch dispcnt.Mode {
		case 3:

			const (
				BASE           = 0x0600_0000
				BYTE_PER_PIXEL = 2
				WIDTH          = SCREEN_WIDTH
			)

			idx := BASE + ((x + (y * WIDTH)) * BYTE_PER_PIXEL)
			data := mem.Read16(idx)
			bldPal.setBlendPalettes(data, BG_IDX, false, false)

		case 4:

			const (
				BASE           = 0x0600_0000
				BYTE_PER_PIXEL = 1
				WIDTH          = SCREEN_WIDTH
			)

			idx := BASE + ((x + (y * WIDTH)) * BYTE_PER_PIXEL)

			if dispcnt.DisplayFrame1 {
				idx += 0xA000
			}

			palIdx := mem.Read8(idx)
			if palIdx != 0 {
				data := gba.getPalette(uint32(palIdx), 0, false)
				bldPal.setBlendPalettes(data, BG_IDX, false, false)
			}

		case 5:

			const (
				BASE           = 0x0600_0000
				BYTE_PER_PIXEL = 2
				WIDTH          = 160
				HEIGHT         = 128
			)

			if x >= WIDTH || y >= HEIGHT {
				palData := gba.getPalette(0, 0, false)
				gba.applyColor(palData, uint32(index))
				return
			}

			idx := BASE + ((x + (y * WIDTH)) * BYTE_PER_PIXEL)
			if dispcnt.DisplayFrame1 {
				idx += 0xA000
			}

			data := mem.Read16(idx)
			bldPal.setBlendPalettes(data, BG_IDX, false, false)
		}
		if objects := dispcnt.DisplayObj; objects {
			objPal := uint32(0)
			objExists := false
			objCount := len(objPriorities[DEC_IDX])

			for j := range objCount {
				objIdx := objPriorities[DEC_IDX][j]
                obj := &gba.PPU.Objects[objIdx]
				palData, ok := gba.object(x, y, obj, wins, dispcnt)

				if ok {
					if obj.Mode == 2 {
						inObjWindow = true
						// break here too? idk
					} else {
						objMode = obj.Mode
						objExists = true
						objPal = palData
						break
					}
				}
			}

			if objExists {
				bldPal.setBlendPalettes(objPal, 0, true, objMode == 1)
			}
		}

		finalPalData := bldPal.blend(objMode, x, y, wins, inObjWindow)
		gba.applyColor(finalPalData, uint32(index))
	}

    if config.Conf.Gba.Threads == 0 {
        for j := range SCREEN_WIDTH {
            renderPixel(uint32(j))
        }
        return
    }

    dx := SCREEN_WIDTH / config.Conf.Gba.Threads

	for i := range config.Conf.Gba.Threads {

		wg.Add(1)

		go func(i int) {

			defer wg.Done()

			for j := range dx {
				renderPixel(uint32((i * dx) + j))
			}
		}(i)
	}

	wg.Wait()
}

func (gba *GBA) object(x, y uint32, obj *Object, wins *Windows, dispcnt *Dispcnt) (uint32, bool) {


	if !windowObjPixelAllowed(x, y, wins) {
		return 0, false
	}

	obj.OneDimensional = dispcnt.OneDimensional

	if obj.RotScale {
		return gba.setObjectAffinePixel(obj, x, y)
	}

	return gba.setObjectPixel(obj, x, y)
}

func outOfBound(h, w, xIdx, yIdx int) bool {
	t := yIdx < 0
	b := yIdx-h >= 0
	l := xIdx < 0
	r := xIdx-w >= 0
	return t || b || l || r
}

func (gba *GBA) setObjectAffinePixel(obj *Object, x, y uint32) (uint32, bool) {

	if gba.outBoundsAffine(obj, x, y) {
		return 0, false
	}

	objX := obj.X
	objY := obj.Y
	if obj.DoubleSize {
		objX += obj.W / 2
		objY += obj.H / 2
	}

	mem := &gba.Mem

	xIdx := int(float32(x) - float32(objX))
	yIdx := int(float32(y)-float32(objY)) % 256

	if objY > SCREEN_HEIGHT {
		yIdx += 256 // i believe 256 is max
	}
	if objX > SCREEN_WIDTH {
		xIdx += 512 // i believe 512 is max
	}

	xOrigin := float32(xIdx - (int(obj.W) / 2))
	yOrigin := float32(yIdx - (int(obj.H) / 2))

	xIdx = int(obj.Pa*xOrigin+obj.Pb*yOrigin) + (int(obj.W) / 2)
	yIdx = int(obj.Pc*xOrigin+obj.Pd*yOrigin) + (int(obj.H) / 2)

	if outOfBound(int(obj.H), int(obj.W), xIdx, yIdx) {
		return 0, false
	}

	enTileX, enTileY, inTileX, inTileY := getPositions(obj, uint32(xIdx), uint32(yIdx))

    readVRAM := mem.readRegions[0x6]
	addr := getTileAddr(obj, enTileX, enTileY, inTileX, inTileY)
    tileData := uint32(readVRAM(mem, addr))

	return getPaletteData(gba, obj.Palette256, true, obj.Palette, tileData, uint32(inTileX))

}

func (gba *GBA) outBoundsAffine(obj *Object, x, y uint32) bool {

	const (
		MAX_X_MASK = 512 - 1
		MAX_Y_MASK = 256 - 1
	)

    dH := obj.H
    dW := obj.W

    if obj.DoubleSize {
        dH <<= 1
        dW <<= 1
    }

    t := obj.Y
    b := (obj.Y + dH) & MAX_Y_MASK
    l := obj.X
    r := (obj.X + dW) & MAX_X_MASK

    inY := (t <= b && y >= t && y < b) || (t > b && (y >= t || y < b))
    inX := (l <= r && x >= l && x < r) || (l > r && (x >= l || x < r))

    return !(inX && inY)
}

func (gba *GBA) setObjectPixel(obj *Object, x, y uint32) (uint32, bool) {

	mem := &gba.Mem

	yIdx := int(y) - int(obj.Y)
	xIdx := int(x) - int(obj.X)

	if obj.Y > SCREEN_HEIGHT {
		yIdx += 256 // i believe 256 is max
	}

	if obj.X > SCREEN_WIDTH {
		xIdx += 512 // i believe 512 is max
	}

	if outOfBound(int(obj.H), int(obj.W), xIdx, yIdx) {
		// blank color
		//gba.applyColor(0, (x + (y * SCREEN_WIDTH))*4)
		return 0, false
	}

	enTileX, enTileY, inTileX, inTileY := getPositions(obj, uint32(xIdx), uint32(yIdx))

	addr := getTileAddr(obj, enTileX, enTileY, inTileX, inTileY)

	tileData := mem.Read16(addr)

	return getPaletteData(gba, obj.Palette256, true, obj.Palette, tileData, uint32(inTileX))

}

func getPositions(obj *Object, xIdx, yIdx uint32) (uint32, uint32, uint32, uint32) {

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

func getTileAddr(obj *Object, enTileX, enTileY, inTileX, inTileY uint32) uint32 {

	tileHeight := int(obj.W) * 4
	tileWidth := 0x20

	if obj.Palette256 {
		enTileX *= 2
		tileHeight *= 2
	}

	const MAX_NUM_TILE = 1024
	var tileIdx int
	if obj.OneDimensional {
		tileIdx = (int(enTileX) * tileWidth) + (int(enTileY) * tileHeight)
		tileIdx = (tileIdx + int(obj.CharName)*tileWidth) % (MAX_NUM_TILE * tileWidth)
	} else {
		tileIdx = int(enTileX) + (int(enTileY) * 32)
		tileIdx = (tileIdx + int(obj.CharName)%MAX_NUM_TILE) * tileWidth
	}

	VRAM_BASE := int(0x0601_0000)
	tileAddr := uint32(VRAM_BASE + tileIdx)

	var inTileIdx uint32
	if obj.Palette256 {
		inTileIdx = uint32(inTileX) + uint32(inTileY*8)
	} else {
		inTileIdx = uint32(inTileX/2) + uint32(inTileY*4)
	}

	return uint32(tileAddr) + inTileIdx
}

func getPaletteData(gba *GBA, pal256, obj bool, pal, tileData, inTileX uint32) (uint32, bool) {

    if !pal256 {
        // 4 is bit depth

        if inTileX & 1 == 1 {
            palIdx := (tileData >> 4) & 0b1111
            if palIdx == 0 {
                return 0, false
            }
            palData := gba.getPalette(palIdx, pal, obj)
            return palData, true
        }

        palIdx := tileData & 0b1111
            if palIdx == 0 {
                return 0, false
            }
        palData := gba.getPalette(palIdx, pal, obj)
        return palData, true
    }

    if tileData & 0xFF == 0 {
        return 0, false
    }

    palData := gba.getPalette(tileData & 0xFF, 0, obj)
    return palData, true
}

func (gba *GBA) getPalette(palIdx uint32, paletteNum uint32, obj bool) uint32 {

	addr := (paletteNum << 5) + palIdx<<1

	if obj {
		addr += 0x200
	}

    return uint32(gba.Mem.PRAM[addr >> 1])
}


func (gba *GBA) getObjPriority(y uint32, objects *[128]Object) [4][]uint32 {

	priorities := [4][]uint32{}

	for i := range 128 {

		obj := &objects[i]

		priority := obj.Priority

		if disabled := obj.Disable && !obj.RotScale; disabled {
			continue
		}

		if objNotScanline(obj, y) {
			continue
		}

		priorities[priority] = append(priorities[priority], uint32(i))
	}

	return priorities
}

func objNotScanline(obj *Object, y uint32) bool {

	if obj.DoubleSize && obj.RotScale {

		offset := obj.H / 2

		localY := int(y) - int(obj.Y+offset)

		if obj.Y+uint32(offset) > SCREEN_HEIGHT {
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

	t := localY < 0
	b := localY-int(obj.H) >= 0

	return t || b
}

func (gba *GBA) background(x, y, idx uint32, bg *Background, wins *Windows) (uint32, bool) {

	if !windowPixelAllowed(idx, x, y, wins) {
		return 0, false
	}

	if bg.Affine {
		return gba.setAffineBackgroundPixel(bg, x)
	}

	return gba.setBackgroundPixel(bg, x, y)
}

func inRange(coord, start, end uint32) bool {
	if end < start {
		return coord >= start || coord < end
	}
	return coord >= start && coord < end
}

func inWindow(x, y, l, r, t, b uint32) bool {
	return inRange(x, l, r) && inRange(y, t, b)
}

func windowPixelAllowed(idx, x, y uint32, wins *Windows) bool {

	if !wins.Enabled {
		return true
	}

	if wins.Win0.Enabled && inWindow(x, y, wins.Win0.L, wins.Win0.R, wins.Win0.T, wins.Win0.B) {
		return wins.Win0.InBg[idx]
	}

	if wins.Win1.Enabled && inWindow(x, y, wins.Win1.L, wins.Win1.R, wins.Win1.T, wins.Win1.B) {
		return wins.Win1.InBg[idx]
	}

	return wins.OutBg[idx]
}

func windowObjPixelAllowed(x, y uint32, wins *Windows) bool {

    if !wins.Enabled {
        return true
    }

	if !wins.Win0.Enabled && !wins.Win1.Enabled {
		return true
	}

	if wins.Win0.Enabled && inWindow(x, y, wins.Win0.L, wins.Win0.R, wins.Win0.T, wins.Win0.B) {
		return wins.Win0.InObj
	}

	if wins.Win1.Enabled && inWindow(x, y, wins.Win1.L, wins.Win1.R, wins.Win1.T, wins.Win1.B) {
		return wins.Win1.InObj
	}

	return wins.OutObj
}

func windowBldPixelAllowed(x, y uint32, wins *Windows, inObjWindow bool) bool {

    if !wins.Enabled {
        return true
    }

	if !wins.Win0.Enabled && !wins.Win1.Enabled && !wins.WinObj.Enabled {
		return true
	}

	if wins.Win0.Enabled && inWindow(x, y, wins.Win0.L, wins.Win0.R, wins.Win0.T, wins.Win0.B) {
		return wins.Win0.InBld
	}

	if wins.Win1.Enabled && inWindow(x, y, wins.Win1.L, wins.Win1.R, wins.Win1.T, wins.Win1.B) {
		return wins.Win1.InBld
	}

	if wins.WinObj.Enabled && inObjWindow {
		return wins.WinObj.InBld
	}

	return wins.OutBld
}

func convert20_8Float(v int32) float64 {

	// sign extend
	sBit := 27
	if v&(1<<sBit) != 0 {
		v |= ^((1 << ((sBit) + 1)) - 1)
	}

	return float64(v>>8) + (float64(v&0xFF) / 256.0)
}

func convert8_8Float(v int16) float64 {
	return float64(v>>8) + (float64(v&0xFF) / 256.0)
}

func (gba *GBA) setAffineBackgroundPixel(bg *Background, x uint32) (uint32, bool) {

	if !bg.Palette256 {
		panic(fmt.Sprintf("AFFINE WITHOUT PAL 256"))
	}

	pa := convert8_8Float(int16(bg.Pa))
	pc := convert8_8Float(int16(bg.Pc))
	xIdx := int(pa*float64(x) + bg.OutX)
	yIdx := int(pc*float64(x) + bg.OutY)

	out := xIdx < 0 || xIdx >= int(bg.W)*8 || yIdx < 0 || yIdx >= int(bg.H)*8

	switch {
	case bg.AffineWrap:
		xIdx &= int(bg.W*8) - 1
		yIdx &= int(bg.H*8) - 1
	case !bg.AffineWrap && out:
		return 0, false
	}

	mem := &gba.Mem

	map_x := (uint32(xIdx) / 8) % (bg.W)
	map_y := (uint32(yIdx) / 8) % (bg.H)
	mapIdx := map_y*bg.W + map_x

	mapAddr := bg.ScreenBaseBlock + mapIdx

	tileIdx := mem.Read8(mapAddr)

	tileAddr := bg.CharBaseBlock + (tileIdx << 6)

	if inObjTiles := tileAddr >= 0x1_0000; inObjTiles {
		return 0, false
	}

	inTileX, inTileY := getPositionsBg(tileIdx, uint32(xIdx), uint32(yIdx))

	inTileIdx := uint32(inTileX) + uint32(inTileY*8)

	addr := tileAddr + inTileIdx
	palIdx := mem.Read8(addr)

    if palIdx == 0 {
        return 0, false
    }

	palData := gba.getPalette(uint32(palIdx), 0, false)
	return palData, true
}

func (gba *GBA) setBackgroundPixel(bg *Background, x, y uint32) (uint32, bool) {

	xIdx := (x + bg.XOffset) & ((bg.W << 3) - 1)
	yIdx := (y + bg.YOffset) & ((bg.H << 3) - 1)

	map_x := (xIdx >> 3) & (bg.W - 1)
	map_y := (yIdx >> 3) & (bg.H - 1)
	mapIdx := (map_y >> 5) << bg.QuadYShift
    mapIdx += (map_x >> 5) << bg.QuadXShift
    mapIdx += (map_y & 31) << 5
    mapIdx += (map_x & 31)

	mapAddr := bg.ScreenBaseBlock + mapIdx << 1

	mem := &gba.Mem
    //readVRAM := mem.readRegions[0x6]

    screenData := uint32(mem.VRAM[(mapAddr + 1) & 0x1FFFF]) << 8
    screenData |= uint32(mem.VRAM[mapAddr & 0x1FFFF])
    //screenData := uint32(readVRAM(mem, mapAddr + 1)) << 8
    //screenData |= uint32(readVRAM(mem, mapAddr))

	tileIdx := utils.GetVarData(screenData, 0, 9) << 5
    if bg.Palette256 {
        tileIdx <<= 1
    }

	tileAddr := bg.CharBaseBlock + tileIdx

	if inObjTiles := tileAddr >= 0x1_0000; inObjTiles {
		return 0, false
	}

	inTileY := yIdx & 0b111 //% 8
	inTileX := xIdx & 0b111 //% 8

    if hFlip := screenData >> 10 & 1 == 1; hFlip {
		inTileX = 7 - inTileX
    }

    if vFlip := screenData >> 11 & 1 == 1; vFlip {
		inTileY = 7 - inTileY
    }

    if bg.Palette256 {
        return gba.getBg256(tileAddr, inTileX, inTileY)
    }

    return gba.getBg16(screenData, tileAddr, inTileX, inTileY)
}

func (gba *GBA) getBg256(tileAddr, inTileX, inTileY uint32) (uint32, bool) {

    addr := tileAddr + inTileX + inTileY << 3
    //tileData := uint32(gba.Mem.readRegions[6](&gba.Mem, addr))
    tileData := uint32(gba.Mem.VRAM[addr & 0x1FFFF])
    palIdx := tileData & 0xFF

    if palIdx == 0 {
        return 0, false
    }

    palData := gba.getPalette(palIdx, 0, false)
    return palData, true
}

func (gba *GBA) getBg16(screenData, tileAddr, inTileX, inTileY uint32) (uint32, bool) {

    addr := tileAddr + (inTileX >> 1) + inTileY << 2
    //tileData := uint32(gba.Mem.readRegions[6](&gba.Mem, addr))
    tileData := uint32(gba.Mem.VRAM[addr & 0x1FFFF])


    // 4 is bit depth
    palIdx := (tileData >> ((inTileX & 1) << 2)) & 0b1111
    if palIdx == 0 {
        return 0, false
    }

    palette := screenData >> 12

	addr = ((palette << 5) + palIdx<<1) >> 1

    return uint32(gba.Mem.PRAM[addr]), true
}


func getPositionsBg(screenData, xIdx, yIdx uint32) (uint32, uint32) {

	inTileY := yIdx & 0b111 //% 8
	inTileX := xIdx & 0b111 //% 8

    if hFlip := screenData >> 10 & 1 == 1; hFlip {
		inTileX = 7 - inTileX
    }

    if vFlip := screenData >> 11 & 1 == 1; vFlip {
		inTileY = 7 - inTileY
    }

	return inTileX, inTileY
}

func (gba *GBA) applyColor(data, i uint32) {
	r := uint8((data) & 0b11111)
	g := uint8((data >> 5) & 0b11111)
	b := uint8((data >> 10) & 0b11111)

	r = (r << 3) | (r >> 2)
	g = (g << 3) | (g >> 2)
	b = (b << 3) | (b >> 2)

	gba.Pixels[i] = r
	gba.Pixels[i+1] = g
	gba.Pixels[i+2] = b
	gba.Pixels[i+3] = 0xFF
}

const (
	BLD_MODE_OFF   = 0
	BLD_MODE_STD   = 1
	BLD_MODE_WHITE = 2
	BLD_MODE_BLACK = 3
)

type BlendPalettes struct {
	Bld            *Blend
	NoBlendPalette uint32
	APalette       uint32
	BPalette       uint32
	hasA, hasB     bool
	targetATop bool
}

func (bp *BlendPalettes) reset(gba *GBA) {
    bp.Bld = &gba.PPU.Blend
	bp.NoBlendPalette = 0
	bp.APalette = 0
	bp.BPalette = 0
	bp.hasA = false
	bp.hasB = false
	bp.targetATop = false

	backdrop := gba.getPalette(0, 0, false)

	bp.NoBlendPalette = backdrop

	if bp.Bld.a[5] {
		bp.APalette = backdrop
		bp.hasA = true
		bp.targetATop = true
	} else {
		bp.targetATop = false
	}

	if bp.Bld.b[5] {
		bp.BPalette = backdrop
		bp.hasB = true
	}
}

func (bp *BlendPalettes) setBlendPalettes(palData uint32, bgIdx uint32, obj bool, semiTransparent bool) {

	bp.NoBlendPalette = palData

	if obj {

		if bp.Bld.a[4] || semiTransparent {
			bp.APalette = palData
			bp.hasA = true
			bp.targetATop = true
		} else {
			bp.targetATop = false
		}

		if bp.Bld.b[4] {
			bp.BPalette = palData
			bp.hasB = true
		}
		return
	}

	if bp.Bld.a[bgIdx] {
		bp.APalette = palData
		bp.hasA = true
		bp.targetATop = true
	} else {
		bp.targetATop = false
	}

	if bp.Bld.b[bgIdx] {
		bp.BPalette = palData
		bp.hasB = true
	}
}

func (bp *BlendPalettes) blend(objMode uint32, x ,y uint32, wins *Windows, inObjWindow bool) uint32 {

    objTransparent := objMode == 1

    if !windowBldPixelAllowed(x, y, wins, inObjWindow) {
        return bp.noBlend(objTransparent)
    }

    switch bp.Bld.Mode {
    case BLD_MODE_OFF: return bp.noBlend(objTransparent)
    case BLD_MODE_STD: return bp.alphaBlend()
    case BLD_MODE_WHITE: return bp.grayscaleBlend(true)
    case BLD_MODE_BLACK: return bp.grayscaleBlend(false)
    }

    return bp.noBlend(objTransparent)
}

func (bp *BlendPalettes) noBlend(objTransparent bool) uint32 {
    if objTransparent {
        return bp.alphaBlend()
    }
    return bp.NoBlendPalette
}

func (bp *BlendPalettes) alphaBlend() uint32 {

    if !bp.hasA || !bp.hasB || !bp.targetATop {
        return bp.NoBlendPalette
    }

    rA := float32((bp.APalette) & 0x1F)
    gA := float32((bp.APalette >> 5) & 0x1F)
    bA := float32((bp.APalette >> 10) & 0x1F)
    rB := float32((bp.BPalette) & 0x1F)
    gB := float32((bp.BPalette >> 5) & 0x1F)
    bB := float32((bp.BPalette >> 10) & 0x1F)

    blend := func(a, b float32) uint32 {
        val := a*bp.Bld.aEv + b*bp.Bld.bEv
        return uint32(min(31, val))
    }
    r := blend(rA, rB)
    g := blend(gA, gB)
    b := blend(bA, bB)

    return r | (g << 5) | (b << 10)
}

func (bp *BlendPalettes) grayscaleBlend(white bool) uint32 {

    if !bp.hasA || !bp.targetATop {
        return bp.NoBlendPalette
    }

    rA := float32((bp.APalette) & 0x1F)
    gA := float32((bp.APalette >> 5) & 0x1F)
    bA := float32((bp.APalette >> 10) & 0x1F)

    blend := func(v float32) uint32 {

        if white {
            v += (31 - v)*bp.Bld.yEv
        } else {
            v -= v*bp.Bld.yEv
        }
        
        return uint32(min(31, v))
    }

    r := blend(rA)
    g := blend(gA)
    b := blend(bA)

    return r | (g << 5) | (b << 10)
}
