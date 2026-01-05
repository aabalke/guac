package ppu

import "github.com/aabalke/guac/emu/nds/utils"

// this is temp share file for sisd and simd graphics
// once simd is completely implimented this will be removed

func (ppu *PPU) screenoff(y uint32, e *Engine) {
	start := y * SCREEN_WIDTH << 2
	end := start + SCREEN_WIDTH<<2
	copy(e.Pixels[start:end], ppu.WHITE_SCANLINE)
}

func (ppu *PPU) MemFifoDisplay(e *Engine) {
	copy(e.Pixels, ppu.DisplayFifo.Pixels)
}

var bgFuncs = [...]func(ppu *PPU, e *Engine, bg *Background, bgIdx, x, y uint32) (palData uint32, alpha float32, ok bool){
	func(ppu *PPU, e *Engine, bg *Background, bgIdx, x, y uint32) (uint32, float32, bool) {
		return ppu.setBackgroundPixel(e, bg, bgIdx, x, y)
	},
	func(ppu *PPU, e *Engine, bg *Background, bgIdx, x, y uint32) (uint32, float32, bool) {
		return ppu.setAffineBackgroundPixel(e, bg, bgIdx, x)
	},
	func(ppu *PPU, e *Engine, bg *Background, bgIdx, x, y uint32) (uint32, float32, bool) {
		return ppu.setAffine16BackgroundPixel(e, bg, bgIdx, x)
	},
	func(ppu *PPU, e *Engine, bg *Background, bgIdx, x, y uint32) (uint32, float32, bool) {
		return ppu.set3d(e, bg, x, y)
	},
	func(ppu *PPU, e *Engine, bg *Background, bgIdx, x, y uint32) (uint32, float32, bool) {
		return ppu.setAffine16BackgroundPixel(e, bg, bgIdx, x)
	},
	func(ppu *PPU, e *Engine, bg *Background, bgIdx, x, y uint32) (uint32, float32, bool) {
		return ppu.setBmpBackgroundPixel(e, bg, x)
	},
	func(ppu *PPU, e *Engine, bg *Background, bgIdx, x, y uint32) (uint32, float32, bool) {
		return ppu.setDirectBitmap(e, bg, x)
	},
}

func (ppu *PPU) render(x, y uint32, e *Engine, bldPal *BlendPalettes) (bool, bool) {

	var (
		dispcnt       = &e.Dispcnt
		wins          = &e.Windows
		bld           = &e.Blend
		objPriorities = &e.ObjPriorities
		bgPriorities  = &e.BgPriorities

		isSemiTransparent bool
		inObjWindow       bool
	)

	// work backwards for proper priorities
	for i := 3; i >= 0; i-- {

		for j := len(bgPriorities[i]) - 1; j >= 0; j-- {

			bgIdx := bgPriorities[i][j]

			if !WindowPixelAllowed(bgIdx, x, y, wins) {
				continue
			}

			bg := &e.Backgrounds[bgIdx]

			if palData, alpha, ok := bgFuncs[bg.Type](ppu, e, bg, bgIdx, x, y); ok {
				bldPal.SetBgPalettes(palData, bgIdx, bg.Type == BG_TYPE_3D, alpha, bld)
			}
		}

		if objDisabled := !dispcnt.DisplayObj; objDisabled {
			continue
		}

		if !WindowObjPixelAllowedX(x, y, wins) {
			continue
		}

	ObjectLoop:
		for j := 0; j < len((*objPriorities)[i]); j++ {

			objIdx := (*objPriorities)[i][j]
			obj := &e.Objects[objIdx]

			palData, ok := uint32(0), false
			if bmp := obj.Mode == 3; bmp {
				obj.OneDimensional = dispcnt.BitmapObj1D
				palData, ok = ppu.setObjBmpPixel(e, obj, x, y)
			} else {
				obj.OneDimensional = dispcnt.TileObj1D
				palData, ok = ppu.setObjTilePixel(e, obj, x, y)
			}

			if !ok {
				continue
			}

			if obj.Mode == 2 {
				inObjWindow = true
				break ObjectLoop
			}

			isSemiTransparent = obj.Mode == 1
			bldPal.SetObjPalettes(palData, isSemiTransparent, bld)
			break ObjectLoop
		}
	}

	return isSemiTransparent, inObjWindow
	//return bldPal.Blend(isSemiTransparent, x, y, wins, inObjWindow)
}

func (e *Engine) updateBackgrounds() *[4]Background {

	bgs := &e.Backgrounds

	getExtended := func(bg *Background) uint8 {

		if !bg.Palette256 {
			return BG_TYPE_BGM
		}

		// CharBaseBlock is precalced, need to / 0x4000
		if (bg.CharBaseBlock>>14)&1 == 1 {
			return BG_TYPE_DIR
		}

		return BG_TYPE_256
	}

	for i := range 4 {

		switch i {
		case 0:
			if !e.IsB && e.Dispcnt.Is3D {
				bgs[i].Type = BG_TYPE_3D
			} else {
				bgs[i].Type = BG_TYPE_TEX

				// does bg0 mode 6 panic?
			}
		case 1:
			bgs[i].Type = BG_TYPE_TEX
		case 2:

			switch e.Dispcnt.Mode {
			case 0, 1, 3:
				bgs[i].Type = BG_TYPE_TEX
			case 2, 4:
				bgs[i].Type = BG_TYPE_AFF
			case 5:
				bgs[i].Type = getExtended(&bgs[i])
			case 6:
				bgs[i].Type = BG_TYPE_LAR
			}

		case 3:

			switch e.Dispcnt.Mode {
			case 0:
				bgs[i].Type = BG_TYPE_TEX
			case 1, 2:
				bgs[i].Type = BG_TYPE_AFF
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

func (ppu *PPU) set3d(e *Engine, bg *Background, x, y uint32) (uint32, float32, bool) {

	//xIdx := int(x) + int(bg.XOffset)
	//yIdx := int(y) + int(bg.YOffset)
	xIdx := (x + bg.XOffset) & ((bg.W) - 1)
	yIdx := (y + bg.YOffset) & ((bg.H) - 1)

	i := (xIdx + (yIdx * SCREEN_WIDTH))

	if xIdx >= SCREEN_WIDTH || yIdx >= SCREEN_HEIGHT {
		return 0, 0, false
	}

	r := ppu.Rasterizer.Render

	pal, alpha := uint32(0), float32(0)

	if r.Rasterizer.Buffers.BisRendering {
		pal, alpha = r.Pixels.PalettesA[i], r.Pixels.AlphaA[i]
	} else {
		pal, alpha = r.Pixels.PalettesB[i], r.Pixels.AlphaB[i]
	}

	if alpha == 0 {
		return pal, alpha, false
	}

	if noblend := ppu.EngineA.Blend.Mode == 0; noblend {

		// this is only important if the 3d screen is the only one, nothing is behind it, and alpha != 1.
		// really only noticed in devkit tests ( nehe/lesson10)

		r := float32(pal&0x1F) * alpha
		g := float32((pal>>5)&0x1F) * alpha
		b := float32((pal>>10)&0x1F) * alpha

		pal &= 0x8000
		pal |= uint32(r) & 0x1F
		pal |= ((uint32(g) & 0x1F) << 5)
		pal |= ((uint32(b) & 0x1F) << 10)
	}

	return pal, alpha, alpha > 0
}

func (ppu *PPU) setBackgroundPixel(e *Engine, bg *Background, bgIdx, x, y uint32) (uint32, float32, bool) {

	xIdx := (x + bg.XOffset) & ((bg.W) - 1)
	yIdx := (y + bg.YOffset) & ((bg.H) - 1)

	if bg.Mosaic {

		if e.Mosaic.BgH != 0 {
			xIdx -= xIdx % (e.Mosaic.BgH + 1)
		}

		if e.Mosaic.BgV != 0 {
			yIdx -= yIdx % (e.Mosaic.BgV + 1)
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
	if !e.IsB {
		mapAddr += e.Dispcnt.ScreenBase
		banks = BANKS_A_2D_BG
	} else {
		mapAddr += 0x20_0000
		banks = BANKS_B_2D_BG
	}

	screenData := uint32(ppu.Vram.ReadGraphical(mapAddr, banks))

	tileIdx := (screenData & 0b11_1111_1111) << 5

	tileAddr := bg.CharBaseBlock + tileIdx
	if bg.Palette256 {
		tileAddr += tileIdx
	}

	if !e.IsB {
		tileAddr += e.Dispcnt.CharBase
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

	palIdx := uint32(ppu.Vram.ReadGraphical(tileAddr+inTileIdx, banks))
	palNum := screenData >> 12

	return getBgPaletteData(ppu, e, bgIdx, bg.Palette256, palNum, palIdx, inTileX)
}

func (ppu *PPU) setAffine16BackgroundPixel(e *Engine, bg *Background, bgIdx, x uint32) (uint32, float32, bool) {

	//if !bg.Palette256 {
	//	panic(fmt.Sprintf("AFFINE WITHOUT PAL 256"))
	//}

	pa := float64(utils.Convert16ToFloat(uint16(bg.Pa), 8))
	pc := float64(utils.Convert16ToFloat(uint16(bg.Pc), 8))
	xIdx := int(pa*float64(x) + bg.OutX)
	yIdx := int(pc*float64(x) + bg.OutY)

	if bg.Mosaic && e.Mosaic.BgH != 0 {
		xIdx -= xIdx % int(e.Mosaic.BgH+1)
	}

	if bg.Mosaic && e.Mosaic.BgV != 0 {
		yIdx -= yIdx % int(e.Mosaic.BgV+1)
	}

	out := xIdx < 0 || xIdx >= int(bg.W) || yIdx < 0 || yIdx >= int(bg.H)

	switch {
	case bg.AffineWrap:
		xIdx &= int(bg.W) - 1
		yIdx &= int(bg.H) - 1
	case !bg.AffineWrap && out:
		return 0, 1, false
	}

	const BYTE_SHIFT = 1

	map_x := (uint32(xIdx)) & (bg.W - 1) >> 3
	map_y := ((uint32(yIdx)) & (bg.H - 1)) >> 3
	map_y *= bg.W >> 3
	mapIdx := map_y + map_x
	mapIdx <<= BYTE_SHIFT

	mapAddr := bg.ScreenBaseBlock + mapIdx

	var banks uint32
	if !e.IsB {
		mapAddr += e.Dispcnt.ScreenBase
		banks = BANKS_A_2D_BG
	} else {
		mapAddr += 0x20_0000
		banks = BANKS_B_2D_BG
	}

	screenData := uint32(ppu.Vram.ReadGraphical(mapAddr, banks))

	tileIdx := (screenData & 0b11_1111_1111) << 5

	tileAddr := bg.CharBaseBlock + tileIdx
	tileAddr += tileIdx

	if !e.IsB {
		tileAddr += e.Dispcnt.CharBase
	} else {
		tileAddr += 0x20_0000
	}

	inTileX, inTileY := getPositionsBg(screenData, uint32(xIdx), uint32(yIdx))
	inTileIdx := uint32(inTileX) + uint32(inTileY<<3)
	palIdx := uint32(ppu.Vram.Read(tileAddr+inTileIdx, true))
	palNum := screenData >> 12

	return getBgPaletteData(ppu, e, bgIdx, true, palNum, palIdx, inTileX)
}

func (ppu *PPU) setAffineBackgroundPixel(e *Engine, bg *Background, bgIdx, x uint32) (uint32, float32, bool) {

	vramOffset := uint32(0)

	if e.IsB {
		vramOffset = uint32(0x20_0000)
	}

	pa := float64(utils.Convert16ToFloat(uint16(bg.Pa), 8))
	pc := float64(utils.Convert16ToFloat(uint16(bg.Pc), 8))
	xIdx := int(pa*float64(x) + bg.OutX)
	yIdx := int(pc*float64(x) + bg.OutY)

	if bg.Mosaic && e.Mosaic.BgH != 0 {
		xIdx -= xIdx % int(e.Mosaic.BgH+1)
	}

	if bg.Mosaic && e.Mosaic.BgV != 0 {
		yIdx -= yIdx % int(e.Mosaic.BgV+1)
	}

	out := xIdx < 0 || xIdx >= int(bg.W) || yIdx < 0 || yIdx >= int(bg.H)

	switch {
	case bg.AffineWrap:
		xIdx &= int(bg.W) - 1
		yIdx &= int(bg.H) - 1
	case !bg.AffineWrap && out:
		return 0, 1, false
	}

	map_x := (uint32(xIdx)) & (bg.W - 1) >> 3
	map_y := ((uint32(yIdx)) & (bg.H - 1)) >> 3
	map_y *= bg.W >> 3
	mapIdx := map_y + map_x

	mapAddr := bg.ScreenBaseBlock + mapIdx

	if !e.IsB {
		mapAddr += e.Dispcnt.ScreenBase
	}

	tileIdx := uint32(ppu.Vram.Read(vramOffset+mapAddr, true))

	tileAddr := bg.CharBaseBlock + (tileIdx << 6)

	inTileX, inTileY := getPositionsBg(tileIdx, uint32(xIdx), uint32(yIdx))
	inTileIdx := uint32(inTileX) + uint32(inTileY<<3)
	addr := vramOffset + tileAddr + inTileIdx
	palIdx := uint32(ppu.Vram.Read(addr, true))

	pal := uint32(0)

	return getBgPaletteData(ppu, e, bgIdx, true, pal, palIdx, inTileX)
}

func (ppu *PPU) setBmpBackgroundPixel(e *Engine, bg *Background, x uint32) (uint32, float32, bool) {

	//if !bg.Palette256 {
	//	panic(fmt.Sprintf("AFFINE WITHOUT PAL 256"))
	//}

	pa := float64(utils.Convert16ToFloat(uint16(bg.Pa), 8))
	pc := float64(utils.Convert16ToFloat(uint16(bg.Pc), 8))
	xIdx := int(pa*float64(x) + bg.OutX)
	yIdx := int(pc*float64(x) + bg.OutY)

	if bg.Mosaic && e.Mosaic.BgH != 0 {
		xIdx -= xIdx % int(e.Mosaic.BgH+1)
	}

	if bg.Mosaic && e.Mosaic.BgV != 0 {
		yIdx -= yIdx % int(e.Mosaic.BgV+1)
	}

	out := xIdx < 0 || xIdx >= int(bg.W) || yIdx < 0 || yIdx >= int(bg.H)

	switch {
	case bg.AffineWrap:
		xIdx &= int(bg.W) - 1
		yIdx &= int(bg.H) - 1
	case !bg.AffineWrap && out:
		return 0, 1, false
	}

	addr := uint32(xIdx + (yIdx * int(bg.W)))
	addr += bg.ScreenBaseBlock * 8

	if e.IsB {
		addr += 0x20_0000
	}

	palIdx := uint32(ppu.Vram.Read(addr, true))

	if palIdx == 0 {
		return 0, 1, false
	}

	data := ppu.getPalette(palIdx, 0, false, e.IsB)

	//if transparent := data & 0x80 != 0; transparent {
	//    return 0, false
	//}

	return data, 1, true
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

func (ppu *PPU) setDirectBitmap(e *Engine, bg *Background, x uint32) (uint32, float32, bool) {

	pa := float64(utils.Convert16ToFloat(uint16(bg.Pa), 8))
	pc := float64(utils.Convert16ToFloat(uint16(bg.Pc), 8))
	xIdx := int(pa*float64(x) + bg.OutX)
	yIdx := int(pc*float64(x) + bg.OutY)

	if bg.Mosaic && e.Mosaic.BgH != 0 {
		xIdx -= xIdx % int(e.Mosaic.BgH+1)
	}

	if bg.Mosaic && e.Mosaic.BgV != 0 {
		yIdx -= yIdx % int(e.Mosaic.BgV+1)
	}

	out := xIdx < 0 || xIdx >= int(bg.W) || yIdx < 0 || yIdx >= int(bg.H)

	switch {
	case bg.AffineWrap:
		xIdx &= int(bg.W) - 1
		yIdx &= int(bg.H) - 1
	case !bg.AffineWrap && out:
		return 0, 1, false
	}

	addr := uint32(xIdx+(yIdx*int(bg.W))) * 2

	addr += bg.ScreenBaseBlock * 8

	var banks uint32
	if e.IsB {
		addr += 0x20_0000
		banks = BANKS_B_2D_BG
	} else {
		banks = BANKS_A_2D_BG
	}

	data := uint32(ppu.Vram.ReadGraphical(addr, banks))

	// required sonic dark brotherhood
	if transparent := (data & 0x8000) == 0; transparent {
		return 0, 1, false
	}

	return data, 1, true
}

func (e *Engine) getBgPriority(y uint32) {

    var (
        //mode = e.Dispcnt.Mode
        //wins = &e.Windows
        bgs  = &e.Backgrounds
        priorities = &e.BgPriorities
    )

	p := [4][]uint32{}

	for i := range uint32(4) {

		if bgs[i].Invalid || !bgs[i].Enabled {
			continue
		}

		if bgNotScanline(&bgs[i], y) {
			continue
		}

		//if !ppu.WindowPixelAllowedScanline(i, y, wins) {
		//    continue
		//}

		priority := bgs[i].Priority

		p[priority] = append(p[priority], uint32(i))
	}

    *priorities = p
}

func (e *Engine) getObjPriority(y uint32) {

    var (
        //mode = e.Dispcnt.Mode
        wins = &e.Windows
        objects  = &e.Objects
        priorities = &e.ObjPriorities
    )

	p := [4][]uint32{}

	for i := range 128 {

		obj := &objects[i]

		if disabled := (obj.Disable && !obj.RotScale) || (obj.RotScale && obj.RotParams >= 64); disabled {
			continue
		}

		if objNotScanline(obj, y) {
			continue
		}

		if !WindowObjPixelAllowedScanline(y, wins) {
			continue
		}

		priority := obj.Priority

		p[priority] = append(p[priority], uint32(i))
	}

    *priorities = p
}

func objNotScanline(obj *Object, y uint32) bool {

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

func bgNotScanline(bg *Background, y uint32) bool {

	if bg.Affine {
		return false
	}

	localY := (int(y) - int(bg.YOffset)) & int((bg.H)-1)

	t := localY < 0
	b := localY-int(bg.H) >= 0

	return t || b
}

func (ppu *PPU) getExtendedPalette(e *Engine, bgIdx uint32, obj bool, palIdx, paletteNum uint32) uint32 {

	//16 colors x 16 palettes --> standard palette memory (=256 colors)
	//256 colors x 16 palettes --> extended palette memory (=4096 colors)

	addr := ((paletteNum << 5) + palIdx<<1)
	vram := &ppu.Vram

	// this can probably be replaced in the ppu.
	// ex. vram.ExtABgSlot (unsafe) -> Background[bgIdx].BgSlot (*0x2000uint8)
	// removes un necessary palette slot switches

	switch {
	case !obj && !e.IsB:

		slotIdx := bgIdx

		if altSlot := e.Backgrounds[bgIdx].AltExtPalSlot; altSlot {
			switch bgIdx {
			case 0:
				slotIdx = 2
			case 1:
				slotIdx = 3
			}
		}

		slot := vram.ExtABgSlots[slotIdx]

		return uint32(b16(slot[addr:]))
	case !obj && e.IsB:

		slotIdx := bgIdx

		if altSlot := e.Backgrounds[bgIdx].AltExtPalSlot; altSlot {
			switch bgIdx {
			case 0:
				slotIdx = 2
			case 1:
				slotIdx = 3
			}
		}

		slot := vram.ExtBBgSlots[slotIdx]

		return uint32(b16(slot[addr:]))
	case obj && !e.IsB:
		return uint32(b16(vram.ExtAPalObj[addr:]))
	case obj && e.IsB:
		return uint32(b16(vram.ExtBPalObj[addr:]))
	}

	return 0
}

func (ppu *PPU) getPalette(palIdx uint32, paletteNum uint32, obj, eB bool) uint32 {

	addr := (paletteNum << 5) + palIdx<<1

	if obj {
		addr += 0x200
	}

	if eB {
		addr += 0x400
	}

	addr >>= 1

	//if addr >= 0x400 {
	//    fmt.Printf("BAD PAL ADDR %04X (> 0x400), isB %t\n", addr, eB)
	//    return 0x0 //FFFF
	//}

	return uint32(ppu.Pram[addr])
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

func getObjTileAddr(obj *Object, enTileX, enTileY, inTileX, inTileY uint32) uint32 {

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
		tileIdx = (tileIdx + obj.CharName<<obj.TileBoundaryShift) //& MAX_TILE_MASK
	} else {
		tileIdx = enTileX + (enTileY << 5)
		tileIdx = (tileIdx + obj.CharName /*& MAX_TILE_MASK*/) << 5
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

func getBmpTileAddr(obj *Object, xIdx, yIdx uint32) uint32 {

	const BYTES_PER_PIXEL = 2

	if obj.OneDimensional {
		return uint32(xIdx+(yIdx<<obj.BmpBoundaryShift)) * BYTES_PER_PIXEL
	}

	maskX := obj.BmpBoundaryMask
	base := ((obj.CharName & maskX) << 4) + ((obj.CharName & ^maskX) << 7)

	var pixelOffset uint32
	if obj.ObjBmpMapping == OBJ_BMP_128_2D {
		pixelOffset = (yIdx*(obj.W<<1) + xIdx) * BYTES_PER_PIXEL
	} else {
		pixelOffset = (yIdx*(obj.W<<2) + xIdx) * BYTES_PER_PIXEL
	}

	return base + pixelOffset
}

func getBgPaletteData(ppu *PPU, e *Engine, bgIdx uint32, pal256 bool, palNum, tileData, inTileX uint32) (uint32, float32, bool) {
	//if !e.IsB && x == y && x == 0 {
	//    fmt.Printf("PAL IDX %0X PAL NUM %08X\n", palIdx, palNum)
	//}

	var palIdx uint32
	if pal256 {
		palIdx = tileData & 0xFF
	} else {
		palIdx = (tileData >> ((inTileX & 1) << 2)) & 0xF
	}

	if palIdx == 0 {
		return 0, 1, false
	}

	if e.Dispcnt.BgExtPal && pal256 {

		palNum <<= 4
		palData := ppu.getExtendedPalette(e, bgIdx, false, palIdx, palNum)
		return palData, 1, true
	}

	// this is from gba, does not work with ext palettes, but assume it still
	// is needed for std
	if pal256 {
		palNum = 0
	}

	palData := ppu.getPalette(uint32(palIdx), palNum, false, e.IsB)

	return palData, 1, true
}

func getPaletteData(ppu *PPU, e *Engine, pal256 bool, pal, tileData, inTileX uint32) (uint32, bool) {

	var palIdx uint32
	if pal256 {
		palIdx = tileData & 0xFF
	} else {
		palIdx = (tileData >> ((inTileX & 1) << 2)) & 0xF
	}

	if palIdx == 0 {
		return 0, false
	}

	if e.Dispcnt.ObjExtPal && pal256 {
		pal <<= 4
		palData := ppu.getExtendedPalette(e, 0, true, palIdx, pal)
		return palData, true
	}

	// this is from gba, does not work with ext palettes, but assume it still
	// is needed for std
	if pal256 {
		pal = 0
	}

	palData := ppu.getPalette(uint32(palIdx), pal, true, e.IsB)

	return palData, true
}

func outObjectBound(obj *Object, xIdx, yIdx int) bool {
	t := yIdx < 0
	b := yIdx-int(obj.H) >= 0
	l := xIdx < 0
	r := xIdx-int(obj.W) >= 0
	return t || b || l || r
}

func outBoundAffine(obj *Object, x, y uint32) bool {

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
	yWrappedInBouppu := !yWrapped && (y >= t && y < b)
	yUnwrappedInBouppu := yWrapped && (y >= t || y < b)
	xWrappedInBouppu := !xWrapped && (x >= l && x < r)
	xUnwrappedInBouppu := xWrapped && (x >= l || x < r)
	return !((yWrappedInBouppu || yUnwrappedInBouppu) &&
		(xWrappedInBouppu || xUnwrappedInBouppu))
}

func (ppu *PPU) setObjTilePixel(e *Engine, obj *Object, x, y uint32) (uint32, bool) {

	xIdx, yIdx, ok := int(0), int(0), false
	if obj.RotScale {
		xIdx, yIdx, ok = getObjAffineCoords(e, obj, x, y)
	} else {
		xIdx, yIdx, ok = getObjNormalCoords(e, obj, x, y)
	}

	if !ok {
		return 0, false
	}

	enTileX, enTileY, inTileX, inTileY := getPositions(obj, uint32(xIdx), uint32(yIdx))

	addr := getObjTileAddr(obj, enTileX, enTileY, inTileX, inTileY)

	vramOffset := uint32(0x40_0000)
	var banks uint32
	if e.IsB {
		vramOffset = uint32(0x60_0000)
		banks = BANKS_B_2D_OBJ
	} else {
		banks = BANKS_A_2D_OBJ
	}

	tileData := uint32(ppu.Vram.ReadGraphical(vramOffset+addr, banks))

	return getPaletteData(ppu, e, obj.Palette256, obj.Palette, tileData, uint32(inTileX))
}

func (ppu *PPU) setObjBmpPixel(e *Engine, obj *Object, x, y uint32) (uint32, bool) {

	xIdx, yIdx, ok := int(0), int(0), false
	if obj.RotScale {
		xIdx, yIdx, ok = getObjAffineCoords(e, obj, x, y)
	} else {
		xIdx, yIdx, ok = getObjNormalCoords(e, obj, x, y)
	}

	if !ok {
		return 0, false
	}

	addr := getBmpTileAddr(obj, uint32(xIdx), uint32(yIdx))

	vramOffset := uint32(0x40_0000)
	var banks uint32
	if e.IsB {
		vramOffset = uint32(0x60_0000)
		banks = BANKS_B_2D_OBJ
	} else {
		banks = BANKS_A_2D_OBJ
	}

	data := uint32(ppu.Vram.ReadGraphical(vramOffset+addr, banks))

	if alpha := (data & 0x8000) == 0; alpha {
		return 0, false
	}

	return data, true
}

func getObjAffineCoords(e *Engine, obj *Object, x, y uint32) (int, int, bool) {

	if outBoundAffine(obj, x, y) {
		return 0, 0, false
	}

	xIdx, yIdx := getAffineCoordinates(e, obj, x, y)

	if outObjectBound(obj, xIdx, yIdx) {
		return 0, 0, false
	}

	return xIdx, yIdx, true
}

func getObjNormalCoords(e *Engine, obj *Object, x, y uint32) (int, int, bool) {

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

	if obj.Mosaic && e.Mosaic.ObjH != 0 {
		xIdx -= xIdx % int(e.Mosaic.ObjH+1)
	}

	if obj.Mosaic && e.Mosaic.ObjV != 0 {
		yIdx -= yIdx % int(e.Mosaic.ObjV+1)
	}

	return xIdx, yIdx, true
}

func getAffineCoordinates(e *Engine, obj *Object, x, y uint32) (int, int) {

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

	if obj.Mosaic && e.Mosaic.ObjH != 0 {
		xIdx -= xIdx % int(e.Mosaic.ObjH+1)
	}

	if obj.Mosaic && e.Mosaic.ObjV != 0 {
		yIdx -= yIdx % int(e.Mosaic.ObjV+1)
	}

	xOrigin := float32(xIdx - (int(obj.W) / 2))
	yOrigin := float32(yIdx - (int(obj.H) / 2))

	xIdx = int(obj.Pa*xOrigin+obj.Pb*yOrigin) + (int(obj.W) / 2)
	yIdx = int(obj.Pc*xOrigin+obj.Pd*yOrigin) + (int(obj.H) / 2)

	return xIdx, yIdx
}
