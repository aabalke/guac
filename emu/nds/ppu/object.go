package ppu

import (
	"unsafe"
)

func (e *Engine) getObjPriority(y uint32) {

	priorities := &e.ObjPriorities
	priorities[0].Cnt = 0
	priorities[1].Cnt = 0
	priorities[2].Cnt = 0
	priorities[3].Cnt = 0

    if !e.Dispcnt.DisplayObj {
        return
    }

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

func (ppu *PPU) tiledObject(e *Engine, objIdx, priority, y uint32) {

    obj := &e.Objects[objIdx]

    yIdx := int(y) - int(obj.Y)

    if obj.Y > SCREEN_HEIGHT {
        yIdx += 256 // i believe 256 is max
    }

    if obj.Mosaic && e.Mosaic.ObjV != 0 {
        yIdx -= yIdx % int(e.Mosaic.ObjV+1)
    }

    enTileY := uint32(yIdx >> 3)
    inTileY := uint32(yIdx & 7)
    if obj.VFlip {
        enTileY = (obj.H / 8) - 1 - enTileY
        inTileY = 7 - inTileY
    }

    base := uint32(0x40_0000)
    if e.IsB {
        base = 0x60_0000
    }

    const BYTES_PER_PIXEL = 2
    w := obj.W << BYTES_PER_PIXEL
    if obj.Palette256 {
        w <<= 1
    }

    if e.Dispcnt.TileObj1D {
        base += (enTileY * w) + (obj.CharName<<obj.TileBoundaryShift)
    } else {
        base += ((enTileY << 5) + obj.CharName) << 5
    }

    ptr := ppu.Vram.ReadGraphicalPtr(base)

    for x := range uint32(SCREEN_WIDTH) {

        if e.ObjOk[x] {
            continue
        }

        if e.Windows.Enabled && !e.Windows.inWinObj(x, y) {
            continue
        }

        xIdx := int(x) - int(obj.X)
        if obj.X > SCREEN_WIDTH {
            xIdx += 512 // i believe 512 is max
        }

        if obj.Mosaic && e.Mosaic.ObjH != 0 {
            xIdx -= xIdx % int(e.Mosaic.ObjH+1)
        }

        if outObjectBound(obj, xIdx, yIdx) {
            continue
        }

        enTileX := uint32(xIdx >> 3)
        inTileX := uint32(xIdx & 7)
        if obj.HFlip {
            enTileX = (obj.W >> 3) - 1 - enTileX
            inTileX = 7 - inTileX
        }

        var inTileIdx uint32
        if obj.Palette256 {
            enTileX <<= 1
            inTileIdx = inTileX + (inTileY << 3)
        } else {
            inTileIdx = (inTileX >> 1) + (inTileY << 2)
        }

        addr := (enTileX << 5) + inTileIdx

        var palIdx uint32
        if ptr == nil {
            palIdx = uint32(ppu.Vram.Read16(base + addr))
        } else {
            palIdx = uint32(*(*uint16)(unsafe.Add(ptr, addr)))
        }

        ppu.getObjPalData(e, obj, palIdx, inTileX, priority, x)
    }
}

func (ppu *PPU) tiledObjectAffine(e *Engine, objIdx, priority, y uint32) {

    obj := &e.Objects[objIdx]

    base := uint32(0x40_0000)
    if e.IsB {
        base = 0x60_0000
    }

    const BYTES_PER_PIXEL = 2
    w := obj.W << BYTES_PER_PIXEL
    if obj.Palette256 {
        w <<= 1
    }

    for x := range uint32(SCREEN_WIDTH) {

        if covered := e.ObjOk[x]; covered {
            continue
        }

        if outBoundAffine(obj, x, y) {
            continue
        }
        if e.Windows.Enabled && !e.Windows.inWinObj(x, y) {
            continue
        }


        xIdx, yIdx := getAffineCoordinates(e, obj, x, y)

        if outObjectBound(obj, xIdx, yIdx) {
            continue
        }

        enTileY := uint32(yIdx >> 3)
        enTileX := uint32(xIdx >> 3)
        inTileY := uint32(yIdx & 7)
        inTileX := uint32(xIdx & 7)

        if obj.Palette256 {
            enTileX <<= 1
        }

        var tileOffset uint32
        if e.Dispcnt.TileObj1D {
            tileOffset = (enTileX << 5) + (enTileY * w)
            tileOffset = (tileOffset + obj.CharName<<obj.TileBoundaryShift)
        } else {
            tileOffset = enTileX + (enTileY << 5)
            tileOffset = (tileOffset + obj.CharName) << 5
        }

        var inTileIdx uint32
        if obj.Palette256 {
            inTileIdx = inTileX + (inTileY << 3)
        } else {
            inTileIdx = (inTileX >> 1) + (inTileY << 2)
        }

        offset := tileOffset + inTileIdx

        palIdx := uint32(ppu.Vram.Read16(base+offset))

        ppu.getObjPalData(e, obj, palIdx, inTileX, priority, x)
    }
}

func (ppu *PPU) getObjPalData(e *Engine, obj *Object, palIdx, inTileX, priority, x uint32) {

        pal := obj.Palette

        if obj.Palette256 {
            palIdx &= 0xFF
        } else {
            palIdx = (palIdx >> ((inTileX & 1) << 2)) & 0xF
        }

        if palIdx == 0 {
            return
        }

        e.ObjMode[x] = obj.Mode
        e.ObjOk[x] = true

        if e.Dispcnt.ObjExtPal && obj.Palette256 {
            addr := (pal << 9) + palIdx<<1
            e.ObjPals[x] = *(*uint16)(unsafe.Add(unsafe.Pointer(e.ExtObj), addr))
            return
        }

        if obj.Palette256 {
            e.ObjPals[x] = e.Pram.Obj[palIdx]
            return
        }

        e.ObjPals[x] = e.Pram.Obj[(pal << 4) + palIdx]
}

func (ppu *PPU) bitmapObject(e *Engine, objIdx, priority, y uint32) {

    obj := &e.Objects[objIdx]

    base := uint32(0x40_0000)
    if e.IsB {
        base = 0x60_0000
    }

    for x := range uint32(SCREEN_WIDTH) {

        if covered := e.ObjOk[x]; covered {
            continue
        }
        if e.Windows.Enabled && !e.Windows.inWinObj(x, y) {
            continue
        }


        xIdx, yIdx, ok := getObjNormalCoords(e, obj, x, y)

        if !ok {
            continue
        }

        var offset uint32
        const BYTES_PER_PIXEL = 2
        if e.Dispcnt.BitmapObj1D {
		    offset = uint32(xIdx+(yIdx<<obj.BmpBoundaryShift)) * BYTES_PER_PIXEL
        } else {
            offset = getBmp2d(obj, uint32(xIdx), uint32(yIdx))
        }

        data := ppu.Vram.Read16(base + offset)

        if alpha := (data & 0x8000) == 0; alpha {
            continue
        }

        e.ObjMode[x] = obj.Mode
        e.ObjOk[x] = true
        e.ObjPals[x] = data &^ 0x8000
    }
}

func (ppu *PPU) bitmapObjectAffine(e *Engine, objIdx, priority, y uint32) {

    obj := &e.Objects[objIdx]

    base := uint32(0x40_0000)
    if e.IsB {
        base = 0x60_0000
    }

    for x := range uint32(SCREEN_WIDTH) {

        if covered := e.ObjOk[x]; covered {
            continue
        }
        if e.Windows.Enabled && !e.Windows.inWinObj(x, y) {
            continue
        }


        xIdx, yIdx, ok := getObjAffineCoords(e, obj, x, y)

        if !ok {
            continue
        }

        var offset uint32
        const BYTES_PER_PIXEL = 2
        if e.Dispcnt.BitmapObj1D {
		    offset = uint32(xIdx+(yIdx<<obj.BmpBoundaryShift)) * BYTES_PER_PIXEL
        } else {
            offset = getBmp2d(obj, uint32(xIdx), uint32(yIdx))
        }

        data := ppu.Vram.Read16(base + offset)

        if alpha := (data & 0x8000) == 0; alpha {
            continue
        }

        e.ObjMode[x] = obj.Mode
        e.ObjOk[x] = true
        e.ObjPals[x] = data &^ 0x8000
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

func getBmp2d(obj *Object, xIdx, yIdx uint32) uint32 {

	const BYTES_PER_PIXEL = 2

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
