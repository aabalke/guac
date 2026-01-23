package ppu

func (e *Engine) getObjPriority(y uint32) {

    priorities := &e.ObjPriorities
    priorities[0].Cnt = 0
    priorities[1].Cnt = 0
    priorities[2].Cnt = 0
    priorities[3].Cnt = 0

	for i := range uint32(128) {

		obj := &e.Objects[i]

        switch {
        case !obj.RotScale && obj.Disable:
            obj.MasterEnabled = false
			continue
        case obj.RotScale && obj.RotParams >= 64:
            obj.MasterEnabled = false
			continue
        }

		if objNotScanline(obj, y) {
            obj.MasterEnabled = false
			continue
		}

        obj.MasterEnabled = true
        p := &priorities[obj.Priority]
        p.Idx[p.Cnt] = i
        p.Cnt++
	}
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

func getObjPaletteData(e *Engine, pal256 bool, pal, palIdx, inTileX uint32) (uint32, bool) {

	if pal256 {
		palIdx &= 0xFF
	} else {
		palIdx = (palIdx >> ((inTileX & 1) << 2)) & 0xF
	}

	if palIdx == 0 {
		return 0, false
	}

	if e.Dispcnt.ObjExtPal && pal256 {

        //16 colors x 16 palettes --> standard palette memory (=256 colors)
        //256 colors x 16 palettes --> extended palette memory (=4096 colors)
        // this can probably be replaced in the ppu.
        // ex. vram.ExtABgSlot (unsafe) -> Background[bgIdx].BgSlot (*0x2000uint8)
        // removes un necessary palette slot switches

        addr := (pal << 9) + palIdx<<1
        return uint32(b16(e.ExtObj[addr:])), true
	}

	// this is from gba, does not work with ext palettes, but assume it still
	// is needed for std
	if pal256 {
		pal = 0
	}

	addr := (pal << 4) + palIdx
    return uint32(e.Pram.Obj[addr]), true
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
	yWrappedInBound := !yWrapped && (y >= t && y < b)
	yUnwrappedInBound := yWrapped && (y >= t || y < b)
	xWrappedInBound := !xWrapped && (x >= l && x < r)
	xUnwrappedInBound := xWrapped && (x >= l || x < r)
	return !((yWrappedInBound || yUnwrappedInBound) &&
		(xWrappedInBound || xUnwrappedInBound))
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
	if e.IsB {
		vramOffset = 0x60_0000
	}

	tileData := uint32(ppu.Vram.ReadGraphical(vramOffset+addr))

	return getObjPaletteData(e, obj.Palette256, obj.Palette, tileData, inTileX)
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
	if e.IsB {
		vramOffset = uint32(0x60_0000)
	}

	data := uint32(ppu.Vram.ReadGraphical(vramOffset+addr))

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
