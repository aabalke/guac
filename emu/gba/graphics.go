package gba

import (
	"fmt"
	"sync"

	"github.com/aabalke33/guac/emu/gba/utils"
)

var ( 
    _ = fmt.Sprintf("")
)
const (
    WAIT_GROUPS = 8
    dx = SCREEN_WIDTH / WAIT_GROUPS
)

func NewBackgrounds(gba *GBA, dispcnt *Dispcnt) *[4]Background {

    bgs := &[4]Background{}

    for i := range 4 {
        isAffine := (
            (dispcnt.Mode == 1 && i == 2) ||
            (dispcnt.Mode == 2 && (i == 2 || i == 3)))
        isStandard := (
            (dispcnt.Mode == 0) ||
            (dispcnt.Mode == 1 && (i == 0 || i == 1 || i == 2)))

        if !isAffine && !isStandard {
            bgs[i].Invalid = true
            continue
        }

        bgs[i] = *NewBackground(gba, dispcnt, uint32(i), isAffine)
    }

    return bgs
}

func NewBackground(gba *GBA, dispcnt *Dispcnt, idx uint32, affine bool) *Background {

    mem := &gba.Mem

    cnt := mem.ReadIODirect(uint32(0x08 + (idx * 0x2)), 2)
    hof := mem.ReadIODirect(uint32(0x10 + (idx * 0x4)), 2)
    vof := mem.ReadIODirect(uint32(0x12 + (idx * 0x4)), 2)

    bg := &gba.PPU.Backgrounds[idx]
    bg.Affine = affine
    bg.Priority = utils.GetVarData(cnt, 0, 1)
    bg.CharBaseBlock = (cnt >> 2) & 0b11
    bg.Mosaic = utils.BitEnabled(cnt, 6)
    bg.Palette256 = utils.BitEnabled(cnt, 7)
    bg.ScreenBaseBlock = utils.GetVarData(cnt, 8, 12)
    bg.AffineWrap = utils.BitEnabled(cnt, 13)
    bg.Size = utils.GetVarData(cnt, 14, 15)
    bg.XOffset = utils.GetVarData(hof, 0, 9)
    bg.YOffset = utils.GetVarData(vof, 0, 9)
    switch idx {
        case 0: bg.Enabled = dispcnt.DisplayBg0
        case 1: bg.Enabled = dispcnt.DisplayBg1
        case 2: bg.Enabled = dispcnt.DisplayBg2
        case 3: bg.Enabled = dispcnt.DisplayBg3
    }
    bg.setSize()

    if bg.Affine {
        paramsAddr := 0x20 * (idx - 1)
        bg.Pa = mem.ReadIODirect(paramsAddr + 0x0, 2)
        bg.Pb = mem.ReadIODirect(paramsAddr + 0x2, 2)
        bg.Pc = mem.ReadIODirect(paramsAddr + 0x4, 2)
        bg.Pd = mem.ReadIODirect(paramsAddr + 0x6, 2)
        bg.aXOffset = mem.ReadIODirect(paramsAddr + 0x8, 4)
        bg.aYOffset = mem.ReadIODirect(paramsAddr + 0xC, 4)
    }

    if (dispcnt.Mode == 1 && idx == 2) || dispcnt.Mode == 2 {
        bg.Palette256 = true
    }

    //if idx == 0 {
    //    bg.Palette256 = true
    //}

    return bg
}

func (obj *Object) setSize(shape, size uint32) {

    const (
        SQUARE = 0
        HORIZONTAL = 1
        VERTICAL = 2
    )

    switch shape {
    case SQUARE:
        switch size {
        case 0: obj.H, obj.W = 8, 8
        case 1: obj.H, obj.W = 16, 16
        case 2: obj.H, obj.W = 32, 32
        case 3: obj.H, obj.W = 64, 64
        }
    case HORIZONTAL:
        switch size {
        case 0: obj.H, obj.W = 8, 16
        case 1: obj.H, obj.W = 8, 32
        case 2: obj.H, obj.W = 16, 32
        case 3: obj.H, obj.W = 32, 64
        }
    case VERTICAL:
        switch size {
        case 0: obj.H, obj.W = 16, 8
        case 1: obj.H, obj.W = 32, 8
        case 2: obj.H, obj.W = 32, 16
        case 3: obj.H, obj.W = 64, 32
        }
    //default: panic("PROHIBITTED OBJ SHAPE")
    }
}

func (gba *GBA) scanlineGraphics(y uint32) {

    if y >= 160 {
        return
    }

	switch gba.PPU.Dispcnt.Mode {
	case 0, 1, 2: gba.scanlineTileMode(y)
	case 3, 4, 5: gba.scanlineBitmapMode(y)
    default: panic("UNKNOWN MODE")
	}
}

func (gba *GBA) scanlineTileMode(y uint32) {

    wg := sync.WaitGroup{}

    bgPriorities := gba.getBgPriority(0)
    objPriorities := gba.getObjPriority(y, &gba.PPU.Objects)

    dispcnt := &gba.PPU.Dispcnt
    bgs := *NewBackgrounds(gba, dispcnt)
    wins := &gba.PPU.Windows

    if dispcnt.Mode >= 3 {
        return
    }

    renderPixel := func(x uint32) {

        bldPal := NewBlendPalette(x, &gba.PPU.Blend)

        index := (x + (y*SCREEN_WIDTH)) * 4

        bldPal.reset(gba)

        var objMode uint32
        var inObjWindow bool

        for i := range 4 {

            // 0 is highest priority
            decIdx := 3 - i

            bgCount := len(bgPriorities[decIdx])
            for j := range bgCount {

                // need bgCount - 1 - j because of blends
                bgIdx := bgPriorities[decIdx][bgCount - 1 - j]

                bg := bgs[bgIdx]
                if bg.Invalid || !bg.Enabled {
                    continue
                }

                palData, ok, palZero := gba.background(x, y, bgIdx, &bg, wins)

                if ok && !palZero {
                    bldPal.setBlendPalettes(palData, uint32(bgIdx), false, false)
                }
            }

            if objects := dispcnt.DisplayObj; objects {
                objPal := uint32(0)
                objExists := false
                objCount := len(objPriorities[decIdx])

                for j := range objCount {
                    objIdx := objPriorities[decIdx][j]
                    palData, ok, palZero, obj := gba.object(x, y, &gba.PPU.Objects[objIdx], wins, dispcnt)

                    if ok && !palZero {
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
        }

        finalPalData := bldPal.blend(objMode, x, y, wins, inObjWindow)

        gba.applyColor(finalPalData, uint32(index))
    }

    for i := range WAIT_GROUPS {

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

    if dispcnt.Mode < 3 {
        return
    }

    renderPixel := func(x uint32) {

        index := (x + (y*SCREEN_WIDTH)) * 4

        switch dispcnt.Mode {
        case 3:

            const (
                BASE           = 0x0600_0000
                BYTE_PER_PIXEL = 2
                WIDTH = SCREEN_WIDTH
            )

            idx := BASE + ((x + (y * WIDTH)) * BYTE_PER_PIXEL)
            data := mem.Read16(idx)
            gba.applyColor(data, uint32(index))
            return

        case 4:

            const (
                BASE           = 0x0600_0000
                BYTE_PER_PIXEL = 1
                WIDTH = SCREEN_WIDTH
            )

            idx := BASE + ((x + (y * WIDTH)) * BYTE_PER_PIXEL)

            if dispcnt.DisplayFrame1 {
                idx += 0xA000
            }

            palIdx := mem.Read8(idx)
            //if palIdx != 0 {
            ////    palData := gba.getPalette(uint32(palIdx), 0, false)
            ////    gba.applyColor(palData, uint32(index))
            ////}
            //    palData := gba.getPalette(uint32(palIdx), 0, false)
            //    gba.applyColor(palData, uint32(index))
            //}
            palData := gba.getPalette(uint32(palIdx), 0, false)
            gba.applyColor(palData, uint32(index))

            // need objs

            return

        case 5:

            const (
                BASE           = 0x0600_0000
                BYTE_PER_PIXEL = 2
                WIDTH = 160
                HEIGHT = 128
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
            gba.applyColor(data, uint32(index))
            return

        default:
            //log.Printf("Invalid Bitmap Mode Graphics: %d \n", dispcnt.Mode)
            return
        }
    }

    for i := range WAIT_GROUPS {

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

func (gba *GBA) object(x, y uint32, obj *Object, wins *Windows, dispcnt *Dispcnt) (uint32, bool, bool, *Object) {

    obj.OneDimensional = dispcnt.OneDimensional

    if disabledStd := obj.Disable && !obj.RotScale; disabledStd {
        return 0, false, false, obj
    }

    if !windowObjPixelAllowed(x, y, wins) {
        return 0, false, false, obj
    }

    if obj.RotScale {
        palData, ok, palZero := gba.setObjectAffinePixel(obj, x, y)
        return palData, ok, palZero, obj
    }

    palData, ok, palZero := gba.setObjectPixel(obj, x, y)
    return palData, ok, palZero, obj
}

func outObjectBoundScanline(obj *Object, y int) bool {
    yIdx := int(y) - int(obj.Y)
    if obj.Y > SCREEN_HEIGHT {
        yIdx += 256 // i believe 256 is max
    }

    t := yIdx < 0
    b := yIdx - int(obj.H) >= 0
    return t || b

}

func outObjectBound(obj *Object, xIdx, yIdx int) bool {
    t := yIdx < 0
    b := yIdx - int(obj.H) >= 0
    l := xIdx < 0
    r := xIdx - int(obj.W) >= 0
    return t || b || l || r
}

func inScreenBounds(x, y int) bool {
    if x < 0 || y < 0 || x > SCREEN_WIDTH || y > SCREEN_HEIGHT {return false}
    return true
}

func (gba *GBA) setObjectAffinePixel(obj *Object, x, y uint32) (uint32, bool, bool) {

    // will need to fix. Large Scaled sprites "pop" into place when wrapping on bottom and right

    if !inScreenBounds(int(x), int(y)) {
        return 0, false, false
    }

    if gba.outBoundsAffine(obj, x, y) {
        return 0, false, false
    }

    objX := obj.X
    objY := obj.Y
    if obj.DoubleSize {
        objX += obj.W / 2
        objY += obj.H / 2
    }

	mem := &gba.Mem

    xIdx := int(float32(x) - float32(objX))
    yIdx := int(float32(y) - float32(objY)) % 256

    if objY > SCREEN_HEIGHT {
        yIdx += 256 // i believe 256 is max
    }
    if objX > SCREEN_WIDTH {
        xIdx += 512 // i believe 512 is max
    }


    xOrigin := float32(xIdx - (int(obj.W) / 2))
    yOrigin := float32(yIdx - (int(obj.H) / 2))

    xIdx = int(obj.Pa * xOrigin + obj.Pb * yOrigin) + (int(obj.W) / 2 )
    yIdx = int(obj.Pc * xOrigin + obj.Pd * yOrigin) + (int(obj.H) / 2 )

    if outObjectBound(obj, xIdx, yIdx) {
        return 0, false, false
    }

    enTileX, enTileY, inTileX, inTileY := getPositions(obj, uint32(xIdx), uint32(yIdx))

    addr := getTileAddr(obj, enTileX, enTileY, inTileX, inTileY)

    tileData := mem.Read16(addr)

    palIdx, palData := getPaletteData(gba, obj.Palette256, obj.Palette, tileData, uint32(inTileX))

    return palData, !(palIdx == 0), palIdx == 0
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

func (gba *GBA) setObjectPixel(obj *Object, x, y uint32) (uint32, bool, bool) {

    if !inScreenBounds(int(x), int(y)) {
        return 0, false, false
    }

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
        // blank color
        //gba.applyColor(0, (x + (y * SCREEN_WIDTH))*4)
        return 0, false, false
    }

    enTileX, enTileY, inTileX, inTileY := getPositions(obj, uint32(xIdx), uint32(yIdx))

    addr := getTileAddr(obj, enTileX, enTileY, inTileX, inTileY)

    tileData := mem.Read16(addr)

    palIdx, palData := getPaletteData(gba, obj.Palette256, obj.Palette, tileData, uint32(inTileX))

    return palData, true, palIdx == 0

}

func getPositions(obj *Object, xIdx, yIdx uint32) (uint32, uint32, uint32, uint32) {

    enTileY := yIdx / 8
    enTileX := xIdx / 8
    inTileY := yIdx % 8
    inTileX := xIdx % 8

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

    VRAM_BASE := int(0x0601_0000)
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
        tileIdx = (tileIdx + int(obj.CharName) * tileWidth) % (MAX_NUM_TILE * tileWidth)
    } else {
        tileIdx = int(enTileX) + (int(enTileY) * 32)
        tileIdx = (tileIdx + int(obj.CharName) % MAX_NUM_TILE) * tileWidth

    }

    tileAddr := uint32(VRAM_BASE + tileIdx)

    var inTileIdx uint32
    if obj.Palette256 {
        inTileIdx = uint32(inTileX) + uint32(inTileY * 8)
    } else {
        inTileIdx = uint32(inTileX / 2) + uint32(inTileY * 4)
    }

    return uint32(tileAddr) + inTileIdx
}

func getPaletteData(gba *GBA, pal256 bool, pal uint32, tileData, inTileX uint32) (uint32, uint32) {

    var palIdx uint32
    if pal256 {
        palIdx = tileData & 0b1111_1111
        pal = 0
    } else {
        if inTileX % 2 == 0 {
            palIdx = tileData & 0b1111
        } else {
            const BIT_DEPTH = 4
            palIdx = (tileData >> 4) & 0b1111
        }
    }

    palData := gba.getPalette(uint32(palIdx), pal, true)

    return palIdx, palData
}

func (gba *GBA) getBgPriority(mode uint32) [4][]uint32 {

    mem := &gba.Mem
    priorities := [4][]uint32{}

    for i := range 4 {

        if mode == 1 && i > 2 { continue }
        if mode == 2 && i < 2 { continue }

        priority := mem.IO[0x8 + (i * 2)] & 0b11

        //priority := mem.ReadIODirect(0x8 + uint32(i * 0x2), 2) & 0b11
        priorities[priority] = append(priorities[priority], uint32(i))
    }

    return priorities
}

func (gba *GBA) getObjPriority(y uint32, objects *[128]Object) [4][]uint32 {

    priorities := [4][]uint32{}

    for i := range 128 {

        obj := &objects[i]

        priority := obj.Priority

        if disabled := obj.Disable && !obj.RotScale; disabled { // may need to make more effective
            continue
        }

        outOfBoundsStandard := outObjectBoundScanline(obj, int(y)) && !obj.RotScale
        //outOfBoundsAffine := !gba.inObjectBoundScanlineAffine(obj, y) && obj.RotScale
        outOfBoundsAffine := (y < obj.Y || y > obj.Y + obj.H) && obj.RotScale && !obj.DoubleSize

        //if outOfBoundsStandard || outOfBoundsAffine {
        if outOfBoundsStandard || outOfBoundsAffine {
            continue
        }

        priorities[priority] = append(priorities[priority], uint32(i))
    }

    return priorities

}

func (gba *GBA) inObjectBoundScanlineAffine(obj *Object, y uint32) bool {

    if obj.DoubleSize {
        // need to setup check on double
        return false
    }

    if y >= obj.Y && y < obj.Y + obj.H {
        return true
    }

    //if obj.Y + obj.H > SCREEN_HEIGHT {
    //    if y >= obj.Y || y < (obj.Y + obj.H) % SCREEN_HEIGHT {
    //        return true
    //    }
    //}


    return false
}

func (gba *GBA) background(x, y, idx uint32, bg *Background, wins *Windows) (uint32, bool, bool) {

    if !windowPixelAllowed(idx, x, y, wins) {
        return 0, false, false
    }

    if bg.Affine {
        palData, ok, palZero := gba.setAffineBackgroundPixel(bg, x, y)
        return palData, ok, palZero
    }

    if !inScreenBounds(int(x), int(y)) {
        return 0, false, false
    }

    palData, ok, palZero := gba.setBackgroundPixel(bg, x, y)

    return palData, ok, palZero
}

func windowPixelAllowed(idx, x, y uint32, wins *Windows) bool {

    if !wins.Enabled {
        return true
    }

    inWindow := func(win Window) bool {

        l := win.L
        r := win.R
        t := win.T
        b := win.B

        switch {
        case r < l && b < t:
            return (x >= l || x < r) && (y >= t || y < b)
        case r < l:                                    
            return (x >= l || x < r) && (y >= t && y < b)
        case b < t:                                    
            return (x >= l && x < r) && (y >= t || y < b)
        default:                                       
            return (x >= l && x < r) && (y >= t && y < b)
        }
    }

    for _, win := range []Window{wins.Win0, wins.Win1} {
        if win.Enabled && inWindow(win) {
            //return false
            switch idx {
            case 0: return win.InBg0
            case 1: return win.InBg1
            case 2: return win.InBg2
            case 3: return win.InBg3
            }
        }
    }
    switch idx {
    case 0: return wins.OutBg0
    case 1: return wins.OutBg1
    case 2: return wins.OutBg2
    case 3: return wins.OutBg3
    }

    return true
}

func windowObjPixelAllowed(x, y uint32, wins *Windows) bool {

    if !wins.Win0.Enabled && !wins.Win1.Enabled {
        return true
    }

    inWindow := func(win Window) bool {
        return (x >= win.L && x < win.R) && (y >= win.T && y < win.B)
    }

    for _, win := range []Window{wins.Win0, wins.Win1} {
        if win.Enabled && inWindow(win) {
            return win.InObj
        }
    }

    return wins.OutObj
}

func windowBldPixelAllowed(x, y uint32, wins *Windows, inObjWindow bool) bool {

    if !wins.Win0.Enabled && !wins.Win1.Enabled && !wins.WinObj.Enabled {
        return true
    }

    inWindow := func(win Window) bool {
        return (x >= win.L && x < win.R) && (y >= win.T && y < win.B)
    }

    for _, win := range []Window{wins.Win0, wins.Win1} {
        if win.Enabled && inWindow(win) {
            return win.InBld
        }
    }

    if wins.WinObj.Enabled && inObjWindow {
        return wins.WinObj.InBld
    }

    return wins.OutBld
}

func convert28Float(v uint32) float32 {
    if negative := utils.BitEnabled(v, 27); negative {
        // this I think is wrong but this works for now
        return float32((v & 0b111_1111_1111_1111_1111_1111_1111)) / 256
    }

    return float32(v & 0b111_1111_1111_1111_1111_1111_1111) / 256
}

func (gba *GBA) setAffineBackgroundPixel(bg *Background, x, y uint32) (uint32, bool, bool) {

    if !inScreenBounds(int(x), int(y)) {
        return 0, false, false
    }

    if !bg.Palette256 {
        panic(fmt.Sprintf("AFFINE WITHOUT PAL 256"))
    }

    pa := float32(int16(bg.Pa)) / 256
    pb := float32(int16(bg.Pb)) / 256
    pc := float32(int16(bg.Pc)) / 256
    pd := float32(int16(bg.Pd)) / 256
    xOffset := convert28Float(bg.aXOffset)
    yOffset := convert28Float(bg.aYOffset)
    xIdx := int(pa * float32(x) + pb * float32(y) + xOffset)
    yIdx := int(pc * float32(x) + pd * float32(y) + yOffset)

    if bg.AffineWrap {
        xIdx %= int(bg.W * 8)
        yIdx %= int(bg.H * 8)
    } else {

        // this does NOT WORK
        //xBound := xIdx
        //if xIdx <= 0x7FFFF {
        //    xBound = -xIdx
        //}

        //yBound := yIdx
        //if yIdx <= 0x7FFFF {
        //    yBound = -yIdx
        //}
        ////if xBound < 0 || xBound - 0x7FFFF >= int(bg.W * 8) || yBound < 0 || yBound - 0x7FFFF >= int(bg.H * 8) {
        ////if xBound < 0 || yBound < 0 {
        //if xBound < 0 || xBound - 0x7FFFF >= int(bg.W * 0x10) || yBound < 0 || yBound - 0x7FFFF >= int(bg.H * 0x10) {

        //    return 0, false, false
        //}
    }

    //tileX := uint32(xIdx / 8) / 2
    tileX := uint32(xIdx / 8)
    tileY := uint32(yIdx / 8)

    VRAM_BASE := int(0x0600_0000)
	mem := &gba.Mem

    //pitch := bg.W / 2
    //sbb := (tileY/32) * (pitch/32) + (tileX/32)
    //mapIdx := (sbb * 1024 + (tileY %32) * 32 + (tileX %32)) * 2
    //mapIdx := ((tileY %32) * 32 + (tileX %32))
    mapIdx := ((tileY % bg.H) * bg.H + (tileX % bg.W))

    screenAddr := bg.ScreenBaseBlock * 0x800

    mapAddr := uint32(VRAM_BASE) + screenAddr + mapIdx

    screenData := mem.Read8(mapAddr)
    tileIdx := screenData

    cbb := (bg.CharBaseBlock * 0x4000)

    tileAddr := uint32(VRAM_BASE) + cbb + (tileIdx * 0x40)

    if inObjTiles := tileAddr >= 0x601_0000; inObjTiles {
        return 0, false, false
    }

    palette := utils.GetVarData(screenData, 12, 15)
    inTileX, inTileY := getPositionsBg(screenData, uint32(xIdx), uint32(yIdx))

    inTileIdx := uint32(inTileX) + uint32(inTileY * 8)

    addr := tileAddr + inTileIdx
    tileData := mem.Read8(addr)

    var palIdx uint32
    if bg.Palette256 {
        palIdx = tileData & 0b1111_1111
        palette = 0
    } else {
        if inTileX % 2 == 0 {
            palIdx = tileData & 0b1111
        } else {
            bitDepth := uint32(4)
            palIdx = (tileData >> uint32(bitDepth)) & 0b1111
        }
    }

    palData := gba.getPalette(uint32(palIdx), palette, false)

    if palIdx == 0 {
        return palData, true, true
    }

    return palData, true, false
}

func (gba *GBA) setBackgroundPixel(bg *Background, x, y uint32) (uint32, bool, bool) {

    xIdx := (x + bg.XOffset) % (bg.W * 8)
    yIdx := (y + bg.YOffset) % (bg.H * 8)

    tileX := xIdx / 8
    tileY := yIdx / 8

    VRAM_BASE := int(0x0600_0000)
	mem := &gba.Mem

    pitch := bg.W
    sbb := (tileY/32) * (pitch/32) + (tileX/32)
    mapIdx := (sbb * 1024 + (tileY %32) * 32 + (tileX %32)) * 2

    screenAddr := bg.ScreenBaseBlock * 0x800

    mapAddr := uint32(VRAM_BASE) + screenAddr + mapIdx

    screenData := mem.Read16(mapAddr)

    tileIdx := utils.GetVarData(screenData, 0, 9)

    cbb := (bg.CharBaseBlock * 0x4000)

    var tileAddr uint32
    if bg.Palette256 {
        //tileAddr += uint32(VRAM_BASE) + cbb + (tileIdx * 0x40)
        tileAddr += uint32(VRAM_BASE) + cbb + (tileIdx * 0x40)
    } else {
        tileAddr += uint32(VRAM_BASE) + cbb + (tileIdx * 0x20)
    }

    if inObjTiles := tileAddr >= 0x601_0000; inObjTiles {
        return 0, false, false
    }

    palette := utils.GetVarData(screenData, 12, 15)
    inTileX, inTileY := getPositionsBg(screenData, xIdx, yIdx)

    var inTileIdx uint32
    if bg.Palette256 {
        inTileIdx = uint32(inTileX) + uint32(inTileY * 8)
    } else {
        inTileIdx = uint32(inTileX / 2) + uint32(inTileY * 4)
    }

    addr := tileAddr + inTileIdx

    tileData := mem.Read8(addr)

    var palIdx uint32
    if bg.Palette256 {
        palIdx = tileData & 0xFF
        palette = 0
    } else {
        if inTileX % 2 == 0 {
            palIdx = tileData & 0b1111
        } else {
            bitDepth := uint32(4)
            palIdx = (tileData >> uint32(bitDepth)) & 0b1111
        }
    }

    palData := gba.getPalette(uint32(palIdx), palette, false)

    return palData, true, palIdx == 0
}

func getPositionsBg(screenData, xIdx, yIdx uint32) (uint32, uint32) {

    inTileY := yIdx % 8
    inTileX := xIdx % 8

    if hFlip := utils.BitEnabled(screenData, 10); hFlip {
        inTileX = 7 - inTileX
    }
    if vFlip := utils.BitEnabled(screenData, 11); vFlip {
        inTileY = 7 - inTileY
    }

    return inTileX, inTileY
}

func (gba *GBA) getPalette(palIdx uint32, paletteNum uint32, obj bool) uint32 {
	pram := &gba.Mem.PRAM

    addr := (paletteNum * 0x20) + palIdx * 2

    if obj {
        addr += 0x200
    }

	return uint32(pram[addr]) | uint32(pram[addr+1])<<8
}

func (gba *GBA) applyColor(data, index uint32) {
	r := uint8((data) & 0b11111)
	g := uint8((data >> 5) & 0b11111)
	b := uint8((data >> 10) & 0b11111)
	c := convertTo24bit(r, g, b)

	(gba.Pixels)[index] = c.R
	(gba.Pixels)[index+1] = c.G
	(gba.Pixels)[index+2] = c.B
	(gba.Pixels)[index+3] = c.A
}

const (
    BLD_MODE_OFF = 0
    BLD_MODE_STD = 1
    BLD_MODE_WHITE = 2
    BLD_MODE_BLACK = 3
)

type BlendPalettes struct {
    Bld *Blend
    NoBlendPalette uint32
    APalette uint32
    BPalette uint32
    hasA, hasB bool

    targetATop bool

    inObjWin bool
    objOutPalette uint32
    objInPalette uint32
}

func NewBlendPalette(i uint32, bld *Blend) *BlendPalettes {
    return &BlendPalettes{ Bld: bld }
}

func (bp *BlendPalettes) reset(gba *GBA) {
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

    aEv := float32(min(int(bp.Bld.aEv), 16)) / 16
    bEv := float32(min(int(bp.Bld.bEv), 16)) / 16

    rA := float32((bp.APalette) & 0x1F)
    gA := float32((bp.APalette >> 5) & 0x1F)
    bA := float32((bp.APalette >> 10) & 0x1F)
    rB := float32((bp.BPalette) & 0x1F)
    gB := float32((bp.BPalette >> 5) & 0x1F)
    bB := float32((bp.BPalette >> 10) & 0x1F)

    blend := func(a, b float32) uint32 {
        val := a*aEv + b*bEv
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

    yEv := float32(min(int(bp.Bld.yEv), 16)) / 16

    rA := float32((bp.APalette) & 0x1F)
    gA := float32((bp.APalette >> 5) & 0x1F)
    bA := float32((bp.APalette >> 10) & 0x1F)

    blend := func(v float32) uint32 {

        if white {
            v += (31 - v)*yEv
        } else {
            v -= v*yEv
        }
        
        return uint32(min(31, v))
    }

    r := blend(rA)
    g := blend(gA)
    b := blend(bA)

    return r | (g << 5) | (b << 10)
}
