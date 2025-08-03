package gba

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/gba/utils"
)

var (
	_ = fmt.Sprintf("")

    wg = sync.WaitGroup{}
)

const (
	WAIT_GROUPS = 8
	dx          = SCREEN_WIDTH / WAIT_GROUPS

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
    switch {
    case gba.PPU.Dispcnt.ForcedBlank:
		x := uint32(0)
		for x = range SCREEN_WIDTH {
			index := (x + (y * SCREEN_WIDTH)) * 4
			(gba.Pixels)[index] = 0xFF
			(gba.Pixels)[index+1] = 0xFF
			(gba.Pixels)[index+2] = 0xFF
			(gba.Pixels)[index+3] = 0xFF
		}
    case gba.PPU.Dispcnt.Mode < 3:
		gba.scanlineTileMode(y)
    default:
		gba.scanlineBitmapMode(y)
    }
}

func (gba *GBA) scanlineTileMode(y uint32) {

	dispcnt := &gba.PPU.Dispcnt
	wins := &gba.PPU.Windows

	bgs := updateBackgrounds(gba, dispcnt)

	bgPriorities := gba.getBgPriority(y, dispcnt.Mode, &gba.PPU.Backgrounds)
	objPriorities := gba.getObjPriority(y, &gba.PPU.Objects)

    if config.Conf.Gba.Threads == 0 {

        x := uint32(0)
        for x = range SCREEN_WIDTH {
            gba.renderTilePixel(x, y, dispcnt, bgs, wins, &bgPriorities, &objPriorities)
        }

        return
    }

    wg.Add(WAIT_GROUPS)

	for i := range WAIT_GROUPS {

		go func(i int) {

			defer wg.Done()

			for j := range dx {
                x := uint32((i * dx) + j)
                gba.renderTilePixel(x, y, dispcnt, bgs, wins, &bgPriorities, &objPriorities)
			}
		}(i)
	}

	wg.Wait()
}

func (gba *GBA) renderTilePixel(x, y uint32, dispcnt *Dispcnt, bgs *[4]Background, wins *Windows, bgPriorities, objPriorities *[4][]uint32) {

    bldPal := NewBlendPalette(x, &gba.PPU.Blend, gba)

    var objMode uint32
    var inObjWindow bool

    // work backwards for proper priorities
    for i := 3; i >= 0; i-- {
        for j := len(bgPriorities[i]) - 1; j >= 0; j-- {

            bgIdx := bgPriorities[i][j]
            bg := &bgs[bgIdx]
            palData, ok := gba.background(x, y, bgIdx, bg, wins)

            if ok {
                bldPal.setBlendPalettes(palData, uint32(bgIdx), false, false)
            }
        }

        if objDisabled := !dispcnt.DisplayObj; objDisabled {
            continue
        }

        for j := len(objPriorities[i]) - 1; j >= 0; j-- {
            objIdx := objPriorities[i][j]
            obj := &gba.PPU.Objects[objIdx]
            obj.OneDimensional = dispcnt.OneDimensional
            palData, ok := gba.object(x, y, obj, wins)

            switch {
            case ok && obj.Mode == 2:
                inObjWindow = true
                break
            case ok:
                objMode = obj.Mode
                bldPal.setBlendPalettes(palData, 0, true, objMode == 1)
                break
            }
        }
    }

    finalPalData := bldPal.blend(objMode == 1, x, y, wins, inObjWindow)
    index := (x + (y * SCREEN_WIDTH)) << 2
    gba.applyColor(finalPalData, uint32(index))
}

func (gba *GBA) scanlineBitmapMode(y uint32) {

	mem := &gba.Mem
	dispcnt := &gba.PPU.Dispcnt

	objPriorities := gba.getObjPriority(y, &gba.PPU.Objects)

	wins := &gba.PPU.Windows

	if dispcnt.Mode < 3 {
		return
	}

	renderPixel := func(x uint32) {

		bldPal := NewBlendPalette(x, &gba.PPU.Blend, gba)
		index := (x + (y * SCREEN_WIDTH)) * 4

		var objMode uint32
		var inObjWindow bool

		BG_IDX := uint32(2)
		DEC_IDX := uint32(0) // this will have to be updated

		switch dispcnt.Mode {
		case 3:

			const (
				BYTE_PER_PIXEL = 2
				WIDTH          = SCREEN_WIDTH
			)

			idx := (x + (y * WIDTH)) * BYTE_PER_PIXEL
            data := uint32(mem.readRegions[6](mem, idx + 1)) << 8
            data |= uint32(mem.readRegions[6](mem, idx))

			bldPal.setBlendPalettes(data, BG_IDX, false, false)

		case 4:

			const (
				BYTE_PER_PIXEL = 1
				WIDTH          = SCREEN_WIDTH
			)

			idx := (x + (y * WIDTH)) * BYTE_PER_PIXEL

			if dispcnt.DisplayFrame1 {
				idx += 0xA000
			}

            palIdx := uint32(mem.readRegions[6](mem, idx))

			if palIdx != 0 {
				data := gba.getPalette(uint32(palIdx), 0, false)
				bldPal.setBlendPalettes(data, BG_IDX, false, false)
			}

		case 5:

			const (
				BYTE_PER_PIXEL = 2
				WIDTH          = 160
				HEIGHT         = 128
			)

			if x >= WIDTH || y >= HEIGHT {
				palData := gba.getPalette(0, 0, false)
				gba.applyColor(palData, uint32(index))
				return
			}

			idx := (x + (y * WIDTH)) * BYTE_PER_PIXEL
			if dispcnt.DisplayFrame1 {
				idx += 0xA000
			}

            data := uint32(mem.readRegions[6](mem, idx + 1)) << 8
            data |= uint32(mem.readRegions[6](mem, idx))
			bldPal.setBlendPalettes(data, BG_IDX, false, false)
		}

        if objs := dispcnt.DisplayObj; objs {

            for j := len(objPriorities[DEC_IDX]) - 1; j >= 0; j-- {
                objIdx := objPriorities[DEC_IDX][j]
                obj := &gba.PPU.Objects[objIdx]

                if obj.Disable && !obj.RotScale {
                    continue
                }

                obj.OneDimensional = dispcnt.OneDimensional

                palData, ok := gba.object(x, y, obj, wins)

                switch {
                case ok && obj.Mode == 2:
                    inObjWindow = true
                    break
                case ok:
                    objMode = obj.Mode
                    bldPal.setBlendPalettes(palData, 0, true, objMode == 1)
                    break
                }
            }
        }

		finalPalData := bldPal.blend(objMode == 1, x, y, wins, inObjWindow)
		gba.applyColor(finalPalData, uint32(index))
	}

    if config.Conf.Gba.Threads == 0 {

        x := uint32(0)
        for x = range SCREEN_WIDTH {
            renderPixel(x)
        }

        return
    }

    wg.Add(WAIT_GROUPS)

	for i := range WAIT_GROUPS {


		go func(i int) {

			defer wg.Done()

			for j := range dx {
				renderPixel(uint32((i * dx) + j))
			}
		}(i)
	}

	wg.Wait()
}

func (gba *GBA) object(x, y uint32, obj *Object, wins *Windows) (uint32, bool) {


	if !windowObjPixelAllowed(x, y, wins) {
		return 0, false
	}

	if obj.RotScale {
		return gba.setObjectAffinePixel(obj, x, y)
	}

	return gba.setObjectPixel(obj, x, y)
}

func outObjectBound(obj *Object, xIdx, yIdx int) bool {
	t := yIdx < 0
	b := yIdx-int(obj.H) >= 0
	l := xIdx < 0
	r := xIdx-int(obj.W) >= 0
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

	if outObjectBound(obj, xIdx, yIdx) {
		return 0, false
	}

	enTileX, enTileY, inTileX, inTileY := getPositions(obj, uint32(xIdx), uint32(yIdx))

	addr := getTileAddr(obj, enTileX, enTileY, inTileX, inTileY)

    //addr &= 0x1FFFF

    //if addr >= 0x18000 {
    //    addr -= 0x8000
    //}

    tileData := uint32(binary.LittleEndian.Uint16(mem.VRAM[addr:]))

	return getPaletteData(gba, obj.Palette256, obj.Palette, tileData, uint32(inTileX))
}

func (gba *GBA) outBoundsAffine(obj *Object, x, y uint32) bool {

	const (
		MAX_X = 512
		MAX_Y = 256
	)

	if !obj.DoubleSize {

		t := obj.Y
		b := (obj.Y + obj.H) % MAX_Y
		l := obj.X
		r := (obj.X + obj.W) % MAX_X

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
	b := (dY + dH) % MAX_Y
	l := dX
	r := (dX + dW) % MAX_X

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

	if outObjectBound(obj, xIdx, yIdx) {
		return 0, false
	}

	enTileX, enTileY, inTileX, inTileY := getPositions(obj, uint32(xIdx), uint32(yIdx))

	addr := getTileAddr(obj, enTileX, enTileY, inTileX, inTileY)

    // checks are only needed if getTileAddr return addr >= 0x18000
    //addr &= 0x1FFFF

    //if addr >= 0x18000 {
    //    addr -= 0x8000
    //}

    tileData := uint32(binary.LittleEndian.Uint16(mem.VRAM[addr:]))

	return getPaletteData(gba, obj.Palette256, obj.Palette, tileData, uint32(inTileX))

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
    const MAX_TILE_OFFSET = 10
	var tileIdx int
	if obj.OneDimensional {
		tileIdx = (int(enTileX) * tileWidth) + (int(enTileY) * tileHeight)
		tileIdx = (tileIdx + int(obj.CharName)*tileWidth) % (MAX_NUM_TILE * tileWidth)
	} else {
		tileIdx = int(enTileX) + (int(enTileY) * 32)
		tileIdx = (tileIdx + int(obj.CharName) >> 10) * tileWidth
	}

	tileAddr := uint32(0x1_0000 + tileIdx)

	var inTileIdx uint32
	if obj.Palette256 {
		inTileIdx = uint32(inTileX) + uint32(inTileY*8)
	} else {
		inTileIdx = uint32(inTileX/2) + uint32(inTileY*4)
	}

    if tileAddr + inTileIdx >= 0x18000 { panic("TILE ADDR IN 2nD MEM BANK need check in objpixel and affine")}

	return uint32(tileAddr) + inTileIdx
}

func getPaletteData(gba *GBA, pal256 bool, pal, tileData, inTileX uint32) (uint32, bool) {

	var palIdx uint32
	if pal256 {
		palIdx = tileData & 0xFF
		pal = 0
	} else {
		// 4 is bit depth
		palIdx = (tileData >> (4 * (inTileX & 1))) & 0b1111
	}

    if palIdx == 0 {
        return 0, false
    }

	palData := gba.getPalette(uint32(palIdx), pal, true)

	return palData, true
}

func (gba *GBA) getBgPriority(y uint32, mode uint32, bgs *[4]Background) [4][]uint32 {

	mem := &gba.Mem
	priorities := [4][]uint32{}

	for i := range 4 {

		if mode == 1 && i > 2 {
			continue
		}
		if mode == 2 && i < 2 {
			continue
		}

		if bgs[i].Invalid || !bgs[i].Enabled {
			continue
		}

		if bgNotScanline(&bgs[i], y) {
			continue
		}

		priority := mem.IO[0x8+(i*2)] & 0b11

		//priority := mem.ReadIODirect(0x8 + uint32(i * 0x2), 2) & 0b11
		priorities[priority] = append(priorities[priority], uint32(i))
	}

	return priorities
}

func (gba *GBA) getObjPriority(y uint32, objects *[128]Object) [4][]uint32 {

	priorities := [4][]uint32{}

	for i := range 128 {

		obj := &objects[i]

		if disabled := (obj.Disable && !obj.RotScale) || (obj.RotScale && obj.RotParams >= 32); disabled {
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

func bgNotScanline(bg *Background, y uint32) bool {

    if bg.Affine {
        return false
    }

	localY := (int(y) - int(bg.YOffset)) & int((bg.H * 8) - 1)

	t := localY < 0
	b := localY-int(bg.H * 8) >= 0

	return t || b
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

	for _, win := range []Window{wins.Win0, wins.Win1} {
		if win.Enabled && inWindow(x, y, win.L, win.R, win.T, win.B) {
			return win.InBg[idx]
		}
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

	for _, win := range []Window{wins.Win0, wins.Win1} {
		if win.Enabled && inWindow(x, y, win.L, win.R, win.T, win.B) {
			return win.InObj
		}
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

	for _, win := range []Window{wins.Win0, wins.Win1} {
		if win.Enabled && inWindow(x, y, win.L, win.R, win.T, win.B) {
			return win.InBld
		}
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

	out := xIdx < 0 || xIdx >= int(bg.W) || yIdx < 0 || yIdx >= int(bg.H)

	switch {
	case bg.AffineWrap:
		xIdx &= int(bg.W) - 1
		yIdx &= int(bg.H) - 1
	case !bg.AffineWrap && out:
		return 0, false
	}

	mem := &gba.Mem

	map_x := (uint32(xIdx)) & (bg.W - 1) >> 3
	map_y := ((uint32(yIdx)) & (bg.H - 1)) >> 3
    map_y *= bg.W >> 3
	mapIdx := map_y + map_x

	mapAddr := bg.ScreenBaseBlock + mapIdx

    tileIdx := uint32(mem.readRegions[6](mem, mapAddr))

	tileAddr := bg.CharBaseBlock + (tileIdx * 64)

	if inObjTiles := tileAddr >= 0x1_0000; inObjTiles {
		return 0, false
	}

	inTileX, inTileY := getPositionsBg(tileIdx, uint32(xIdx), uint32(yIdx))

	inTileIdx := uint32(inTileX) + uint32(inTileY*8)

	addr := tileAddr + inTileIdx
    palIdx := uint32(mem.readRegions[6](mem, addr))

	if palIdx == 0 {
		return 0, false
	}

	palData := gba.getPalette(palIdx, 0, false)

	return palData, true
}

func (gba *GBA) setBackgroundPixel(bg *Background, x, y uint32) (uint32, bool) {

	xIdx := (x + bg.XOffset) & ((bg.W) - 1)
	yIdx := (y + bg.YOffset) & ((bg.H) - 1)
	map_x := xIdx >> 3
	map_y := yIdx >> 3
	quad_x := uint32(10)//32 * 32
	quad_y := uint32(10)//32 * 32
	if bg.Size == 3 {
		quad_y = 11
	}
	mapIdx := (map_y >> 5) << quad_y
    mapIdx += (map_x >> 5) << quad_x
    mapIdx += (map_y&31) << 5
    mapIdx += (map_x&31)
    mapIdx <<= 1

	mapAddr := bg.ScreenBaseBlock + mapIdx
    //mapAddr &= 0x1FFFF

    //if mapAddr >= 0x18000 {
    //    mapAddr -= 0x8000
    //}

	mem := &gba.Mem

    screenData := uint32(binary.LittleEndian.Uint16(mem.VRAM[mapAddr:]))

	tileIdx := utils.GetVarData(screenData, 0, 9) << 5

	tileAddr := bg.CharBaseBlock + tileIdx
	if bg.Palette256 {
		tileAddr += tileIdx
	}

	if inObjTiles := tileAddr >= 0x1_0000; inObjTiles {
		return 0, false
	}

	inTileX, inTileY := getPositionsBg(screenData, xIdx, yIdx)

	var inTileIdx uint32
	if bg.Palette256 {
		inTileIdx = inTileX + (inTileY << 3)
	} else {
		inTileIdx = (inTileX >> 1) + (inTileY << 2)
	}

	addr := tileAddr + inTileIdx

    //addr &= 0x1FFFF

    //if addr >= 0x18000 {
    //    addr -= 0x8000
    //}

    tileData := uint32(mem.VRAM[addr])

	var palIdx, palette uint32
	if bg.Palette256 {
		palIdx = tileData & 0xFF
		palette = 0
	} else {
        palIdx = (tileData >> ((inTileX & 1) << 2)) & 0xF
        palette = screenData >> 12
	}

    if palIdx == 0 {
        return 0, false
    }

	palData := gba.getPalette(uint32(palIdx), palette, false)
    return palData, true
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

func (gba *GBA) getPalette(palIdx uint32, paletteNum uint32, obj bool) uint32 {

	addr := (paletteNum << 5) + palIdx<<1

	if obj {
		addr += 0x200
	}

    return uint32(gba.Mem.PRAM[addr >> 1])
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
	NoBlendPalette, APalette, BPalette uint32
	hasA, hasB, targetATop bool
}

func NewBlendPalette(i uint32, bld *Blend, gba *GBA) *BlendPalettes {

    bp := &BlendPalettes{
        Bld: bld,
    }

	backdrop := gba.getPalette(0, 0, false)

	bp.NoBlendPalette = backdrop

	if bp.Bld.a[5] {
		bp.APalette = backdrop
		bp.hasA = true
		bp.targetATop = true
	}

	if bp.Bld.b[5] {
		bp.BPalette = backdrop
		bp.hasB = true
	}

    return bp
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

func (bp *BlendPalettes) blend(objTransparent bool, x, y uint32, wins *Windows, inObjWindow bool) uint32 {

	if !windowBldPixelAllowed(x, y, wins, inObjWindow) {
		return bp.noBlend(objTransparent)
	}

	switch bp.Bld.Mode {
	case BLD_MODE_OFF:
		return bp.noBlend(objTransparent)
	case BLD_MODE_STD:
		return bp.alphaBlend()
	case BLD_MODE_WHITE:
		return bp.grayscaleBlend(true)
	case BLD_MODE_BLACK:
		return bp.grayscaleBlend(false)
    default:
        return bp.noBlend(objTransparent)
	}
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
			v += (31 - v) * bp.Bld.yEv
		} else {
			v -= v * bp.Bld.yEv
		}

		return uint32(min(31, v))
	}

	r := blend(rA)
	g := blend(gA)
	b := blend(bA)

	return r | (g << 5) | (b << 10)
}
