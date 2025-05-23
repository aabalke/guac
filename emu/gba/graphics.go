package gba

import (
	"fmt"

	"github.com/aabalke33/guac/emu/gba/utils"
)

type GraphicsTiming struct {
    Gba             *GBA
    RefreshCycles   int
    Scanline        int
    HBlank          bool
    VBlank          bool
    hasVBlankDma    bool
    hasHBlankDma    bool
}

func (gt *GraphicsTiming) reset() {
    gt.RefreshCycles = 0
    gt.Scanline = 0
    gt.HBlank = false
    gt.VBlank = false
    gt.hasVBlankDma = false
    gt.hasHBlankDma = false
}

func (gt *GraphicsTiming) update(cycles int) {

    const (
        REFRESH = 280_896 // should this be replaced by clock / fps?
        SCANLINE = 1232
        HDRAW = 960
        HBLANK = 272
        VDRAW = 197120
        VBLANK = 83776
    )

    prevHBlank := gt.HBlank
    prevVBlank := gt.VBlank
    
    gt.RefreshCycles += cycles
    gt.HBlank = gt.RefreshCycles % SCANLINE > HDRAW
    gt.Scanline = gt.RefreshCycles / SCANLINE
    gt.VBlank = gt.Scanline > 160

    dispstat := &gt.Gba.Mem.Dispstat
    dispstat.SetVBlank(gt.VBlank)
    dispstat.SetHBlank(gt.HBlank)
    dispstat.SetVCounter(gt.Scanline)

    if gt.VBlank && !prevVBlank {
    //if gt.VBlank {
        gt.Gba.checkDmas(DMA_MODE_VBL)

        gt.Gba.triggerIRQ(1)

    }
    if gt.HBlank && !prevHBlank {
    //if gt.HBlank {
        gt.Gba.checkDmas(DMA_MODE_HBL)
        gt.Gba.triggerIRQ(0)
    }
}

var ( 
    _ = fmt.Sprintf("")
)

type Dispcnt struct {
    Mode uint32
    CGB bool
    DisplayFrame1 bool
    HBlankIntervalFree bool
    OneDimensional bool
    ForcedBlank bool
    DisplayBg0 bool
    DisplayBg1 bool
    DisplayBg2 bool
    DisplayBg3 bool
    DisplayObj bool
    DisplayWin0 bool
    DisplayWin1 bool
    DisplayObjWin bool
}

func NewDispcnt(dispcnt uint32) *Dispcnt {

    d := &Dispcnt{}
    d.Mode = utils.GetVarData(dispcnt, 0, 2)
    d.CGB = utils.BitEnabled(dispcnt, 3)
    d.DisplayFrame1 = utils.BitEnabled(dispcnt, 4)
    d.HBlankIntervalFree = utils.BitEnabled(dispcnt, 5)
    d.OneDimensional = utils.BitEnabled(dispcnt, 6)
    d.ForcedBlank = utils.BitEnabled(dispcnt, 7)
    d.DisplayBg0 = utils.BitEnabled(dispcnt, 8)
    d.DisplayBg1 = utils.BitEnabled(dispcnt, 9)
    d.DisplayBg2 = utils.BitEnabled(dispcnt, 10)
    d.DisplayBg3 = utils.BitEnabled(dispcnt, 11)
    d.DisplayObj = utils.BitEnabled(dispcnt, 12)
    d.DisplayWin0 = utils.BitEnabled(dispcnt, 13)
    d.DisplayWin1 = utils.BitEnabled(dispcnt, 14)
    d.DisplayObjWin = utils.BitEnabled(dispcnt, 15)

    return d
}

func (gba *GBA) graphics() {

    addr := gba.Mem.Read16(0x0400_0000 + DISPCNT)
    dispcnt := NewDispcnt(addr)

    //gba.background(0x400_0008, dispcnt, false)
    //gba.getTiles(0x601_0000, 0x1E, true, false)
    //gba.debugPalette()
    //return

    gba.clear()

	switch dispcnt.Mode {
	case 0:
		gba.updateMode0(dispcnt)
	case 1:
        panic("mode 1")
		gba.updateMode1(dispcnt)
	case 2:
        panic("mode 2")
		//gba.updateMode2()
	case 3:
		gba.updateMode3()
	case 4:
		gba.updateMode4(dispcnt)
	case 5:
		gba.updateMode5(dispcnt)
    default: panic("UNKNOWN MODE")
	}

}

func (gba *GBA) clear() {
    y, x := uint32(0), uint32(0)
    for y = range SCREEN_HEIGHT {
        for x = range SCREEN_WIDTH {
            gba.applyColor(0, (x + (y * SCREEN_WIDTH))*4)
        }
    }
}

func (gba *GBA) object(dispcnt *Dispcnt, oamOffset uint32, wins *Windows) {

    OAM_BASE := uint32(0x0700_0000)
    mem := gba.Mem

    attr0 := mem.Read16(OAM_BASE + oamOffset)
    attr1 := mem.Read16(OAM_BASE + oamOffset + 2)
    attr2 := mem.Read16(OAM_BASE + oamOffset + 4)

    obj := NewObject(attr0, attr1, attr2, dispcnt)

    if obj.Disable && !obj.RotScale {
        return
    }

    if obj.RotScale {
        paramsAddr := OAM_BASE + (obj.RotParams * 0x20)
        obj.Pa = mem.Read16(paramsAddr + 0x06)
        obj.Pb = mem.Read16(paramsAddr + 0x0E)
        obj.Pc = mem.Read16(paramsAddr + 0x16)
        obj.Pd = mem.Read16(paramsAddr + 0x1E)
    }

    x, y := uint32(0), uint32(0)

    for y = range SCREEN_HEIGHT {
        for x = range SCREEN_WIDTH {

            if !windowObjPixelAllowed(x, y, wins) {
                continue
            }

            if obj.RotScale {
                gba.setObjectAffinePixel(obj, x, y)
                continue
            }

            gba.setObjectPixel(obj, x, y)
        }
    }
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

func (gba *GBA) setObjectAffinePixel(obj *Object, x, y uint32) {

    // will need to fix. Large Scaled sprites "pop" into place when wrapping on bottom and right

    pa := float32(int16(obj.Pa)) / 256
    pb := float32(int16(obj.Pb)) / 256
    pc := float32(int16(obj.Pc)) / 256
    pd := float32(int16(obj.Pd)) / 256

    objX := obj.X
    objY := obj.Y
    if obj.DoubleSize {
        objX += obj.W / 2
        objY += obj.H / 2
    }

	mem := gba.Mem

    xIdx := int(float32(x) - float32(objX))
    yIdx := int(float32(y) - float32(objY)) % 256

    if objY > SCREEN_HEIGHT {
        yIdx += 256 // i believe 256 is max
    }
    if objX > SCREEN_WIDTH {
        xIdx += 512 // i believe 512 is max
    }

    screenX := xIdx
    screenY := yIdx

    xOrigin := screenX - (int(obj.W) / 2)
    yOrigin := screenY - (int(obj.H) / 2)

    xIdx = int(pa * float32(xOrigin) + pb * float32(yOrigin)) + (int(obj.W) / 2 )
    yIdx = int(pc * float32(xOrigin) + pd * float32(yOrigin)) + (int(obj.H) / 2 )

    if gba.outBoundsAffine(obj, x, y) {
        return
    }

    if outObjectBound(obj, xIdx, yIdx) {
        return
    }

    enTileX, enTileY, inTileX, inTileY := getPositions(obj, uint32(xIdx), uint32(yIdx))

    addr := getTileAddr(obj, enTileX, enTileY, inTileX, inTileY)

    tileData := mem.Read16(addr)

    palIdx, palData := getPaletteData(gba, obj, tileData, uint32(inTileX))

    if !inScreenBounds(int(x), int(y)) {
        return
    }

    if palIdx == 0 {
        // this will need to be updated
        return
    }

    index := (x + (y*SCREEN_WIDTH)) * 4
    gba.applyColor(palData, uint32(index))
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

func (gba *GBA) setObjectPixel(obj *Object, x, y uint32) {

	mem := gba.Mem

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
        return
    }

    enTileX, enTileY, inTileX, inTileY := getPositions(obj, uint32(xIdx), uint32(yIdx))

    addr := getTileAddr(obj, enTileX, enTileY, inTileX, inTileY)

    tileData := mem.Read16(addr)

    palIdx, palData := getPaletteData(gba, obj, tileData, uint32(inTileX))

    if !inScreenBounds(int(x), int(y)) {
        return
    }

    if palIdx == 0 {
        // this will need to be updated
        return
    }

    index := (x + (y*SCREEN_WIDTH)) * 4
    gba.applyColor(palData, uint32(index))
}

func getPositions (obj *Object, xIdx, yIdx uint32) (uint32, uint32, uint32, uint32) {

    enTileY := yIdx / 8
    enTileX := xIdx / 8
    inTileY := yIdx % 8
    inTileX := xIdx % 8

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
    tileWidth := 0x20
    tileHeight := int(obj.W) * 4

    if obj.Palette256 {
        tileWidth *= 2
        tileHeight *= 2
    }

    tileIdx := (int(enTileX) * tileWidth) + (int(enTileY) * tileHeight)
    if !obj.OneDimensional {
        tileIdx += (int(enTileY) * 0x400) - tileHeight
    }

    const MAX_NUM_TILE = 1024
    tileIdx = (tileIdx + int(obj.CharName) * tileWidth) % (MAX_NUM_TILE * tileWidth)

    tileAddr := uint32(VRAM_BASE + tileIdx)

    var inTileIdx uint32
    if obj.Palette256 {
        inTileIdx = uint32(inTileX) + uint32(inTileY * 8)
    } else {
        inTileIdx = uint32(inTileX / 2) + uint32(inTileY * 4)
    }

    return uint32(tileAddr) + inTileIdx
}

func getPaletteData(gba *GBA, obj *Object, tileData, inTileX uint32) (uint32, uint32) {

    var palIdx uint32
    if obj.Palette256 {
        palIdx = tileData & 0b1111_1111
        obj.Palette = 0
    } else {
        if inTileX % 2 == 0 {
            palIdx = tileData & 0b1111
        } else {
            const BIT_DEPTH = 4
            palIdx = (tileData >> 4) & 0b1111
        }
    }

    palData := gba.getPalette(uint32(palIdx), obj.Palette, true)

    return palIdx, palData
}

type Object struct {
    X, Y, W, H uint32
    Pa, Pb, Pc, Pd uint32
    RotScale bool
    DoubleSize bool
    Disable bool
    Mode uint32
    Mosaic bool
    Palette256 bool
    Shape uint32
    HFlip, VFlip bool
    Size uint32
    RotParams uint32
    CharName uint32
    Priority uint32
    Palette uint32
    OneDimensional bool
}

func NewObject(attr0, attr1, attr2 uint32, dispcnt *Dispcnt) *Object {

    obj := &Object{}

    obj.Y = attr0 & 0b1111_1111
    obj.RotScale = utils.BitEnabled(attr0, 8)

    obj.Mode = utils.GetVarData(attr0, 10, 11)
    obj.Mosaic = utils.BitEnabled(attr0, 12)
    obj.Palette256 = utils.BitEnabled(attr0, 13)
    obj.Shape = utils.GetVarData(attr0, 14, 15)
    
    obj.X = attr1 & 0b1_1111_1111

    if obj.RotScale {
        obj.DoubleSize = utils.BitEnabled(attr0, 9)
        obj.RotParams = utils.GetVarData(attr1, 9, 13)
    } else {
        obj.Disable = utils.BitEnabled(attr0, 9)
        obj.HFlip = utils.BitEnabled(attr1, 12)
        obj.VFlip = utils.BitEnabled(attr1, 13)
    }
    obj.Size = utils.GetVarData(attr1, 14, 15)

    obj.CharName = utils.GetVarData(attr2, 0, 9)
    obj.Priority = utils.GetVarData(attr2, 10, 11)
    obj.Palette = utils.GetVarData(attr2, 12, 15)

    obj.OneDimensional = dispcnt.OneDimensional

    obj.setSize(obj.Shape, obj.Size)

    return obj
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
    default: panic("PROHIBITTED OBJ SHAPE")
    }
}


func bgEnabled(dispcnt *Dispcnt, idx int) bool {
    switch idx {
    case 0: return dispcnt.DisplayBg0
    case 1: return dispcnt.DisplayBg1
    case 2: return dispcnt.DisplayBg2
    case 3: return dispcnt.DisplayBg3
    }

    return false
}

func (gba *GBA) updateMode0(dispcnt *Dispcnt) {

    priorities := gba.getBgPriority(0)
    objPriorities := gba.getObjPriority()

    wins := NewWindows(dispcnt, gba)

    for i := range priorities {

        // 0 is highest priority
        decIdx := 3 - i

        if bgEnabled(dispcnt, decIdx) {
            gba.background(uint32(decIdx), dispcnt, false, wins)
        }

        if obj := dispcnt.DisplayObj; obj {
            objCount := len(objPriorities[decIdx])
            for j := range objCount {
                objIdx := objPriorities[decIdx][objCount - 1 - j]
                gba.object(dispcnt, objIdx * 0x8, wins)
            }
        }
    }
}

func (gba *GBA) updateMode1(dispcnt *Dispcnt) {

    priorities := gba.getBgPriority(1)

    wins := NewWindows(dispcnt, gba)

    for i := range priorities {

        // 0 is highest priority
        decIdx := 3 - i

        if !bgEnabled(dispcnt, decIdx) {
            continue
        }

        if decIdx == 2 {
            gba.background(uint32(decIdx), dispcnt, true, wins)
            return
        }

        gba.background(uint32(decIdx), dispcnt, false, wins)
    }
}

func (gba *GBA) getBgPriority(mode uint32) [4]uint32 {

    mem := gba.Mem

    pr0 := mem.Read16(0x400_0008) & 0b11
    pr1 := mem.Read16(0x400_000A) & 0b11
    pr2 := mem.Read16(0x400_000C) & 0b11
    pr3 := mem.Read16(0x400_000E) & 0b11

    out := [4]uint32{}

    for i, v := range []uint32{pr0, pr1, pr2, pr3} {
        if mode == 1 && i > 2 { continue }
        if mode == 2 && i < 2 { continue }
        out[v] = uint32(i)
    }

    return out
}

func (gba *GBA) getObjPriority() [4][]uint32 {

    mem := gba.Mem
    attr2Base := uint32(0x700_0004)

    priorities := [4][]uint32{}

    for i := range 128 {
        attr2 := mem.Read16(attr2Base + (uint32(i) * 0x8))
        priority := utils.GetVarData(attr2, 10, 11)
        priorities[priority] = append(priorities[priority], uint32(i))
    }

    return priorities

}

func (gba *GBA) background(idx uint32, dispcnt *Dispcnt, affine bool, wins *Windows) {

    mem := gba.Mem

    cnt := mem.Read16(uint32(0x0400_0008 + ((idx) * 0x2)))
    hof := mem.Read16(uint32(0x0400_0010 + ((idx) * 0x4)))
    vof := mem.Read16(uint32(0x0400_0012 + ((idx) * 0x4)))
    bg := NewBackground(cnt, hof, vof, affine)


    x, y := uint32(0), uint32(0)
    for y = range SCREEN_HEIGHT {
        for x = range SCREEN_WIDTH {

            if !windowPixelAllowed(idx, x, y, wins) {
                continue
            }

            if affine {
                gba.setAffineBackgroundPixel(bg, x, y)
                continue
            }

            gba.setBackgroundPixel(bg, x, y)
        }
    }
}

func windowPixelAllowed(idx, x, y uint32, wins *Windows) bool {

    inWindow := func(win *Window) bool {

        l := win.L
        r := win.R
        t := win.T
        b := win.B

        // need to handle wrapping

        // should max be removed in NewWindow?

        return (x >= l && x < r) && (y >= t && y < b)
    }

    for _, win := range []*Window{wins.Win0, wins.Win1} {
        if win.Enabled && inWindow(win) {
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

    return false
}
func windowObjPixelAllowed(x, y uint32, wins *Windows) bool {

    inWindow := func(win *Window) bool {
        return (x >= win.L && x < win.R) && (y >= win.T && y < win.B)
    }

    for _, win := range []*Window{wins.Win0, wins.Win1} {
        if win.Enabled && inWindow(win) {
            return win.InObj
        }
    }

    return wins.OutObj
}

func (gba *GBA) setAffineBackgroundPixel(bg *Background, x, y uint32) {

    index := (x + (y*SCREEN_WIDTH)) * 4

    if !inScreenBounds(int(x), int(y)) {
        return
    }

    x = (x + bg.XOffset) % (bg.W * 8)
    y = (y + bg.YOffset) % (bg.H * 8)
    tileX := x / 8
    tileY := y / 8

    VRAM_BASE := int(0x0600_0000)
	mem := gba.Mem

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
        tileAddr += uint32(VRAM_BASE) + cbb + (tileIdx * 0x40)
    } else {
        tileAddr += uint32(VRAM_BASE) + cbb + (tileIdx * 0x20)
    }

    if inObjTiles := tileAddr >= 0x601_0000; inObjTiles {
        return
    }

    hFlip := utils.BitEnabled(screenData, 10)
    vFlip := utils.BitEnabled(screenData, 11)
    palette := utils.GetVarData(screenData, 12, 15)


    inTileY := int(y % 8)
    inTileX := int(x % 8)

    if hFlip {
        inTileX = 7 - inTileX
    }
    if vFlip {
        inTileY = 7 - inTileY
    }

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

    // this will need to be updated
    if palIdx == 0 {
        return
    }
    gba.applyColor(palData, uint32(index))
}

func (gba *GBA) setBackgroundPixel(bg *Background, x, y uint32) {


    index := (x + (y*SCREEN_WIDTH)) * 4

    if !inScreenBounds(int(x), int(y)) {
        return
    }

    x = (x + bg.XOffset) % (bg.W * 8)
    y = (y + bg.YOffset) % (bg.H * 8)
    tileX := x / 8
    tileY := y / 8

    VRAM_BASE := int(0x0600_0000)
	mem := gba.Mem

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
        tileAddr += uint32(VRAM_BASE) + cbb + (tileIdx * 0x40)
    } else {
        tileAddr += uint32(VRAM_BASE) + cbb + (tileIdx * 0x20)
    }

    if inObjTiles := tileAddr >= 0x601_0000; inObjTiles {
        return
    }

    hFlip := utils.BitEnabled(screenData, 10)
    vFlip := utils.BitEnabled(screenData, 11)
    palette := utils.GetVarData(screenData, 12, 15)

    inTileY := int(y % 8)
    inTileX := int(x % 8)

    if hFlip {
        inTileX = 7 - inTileX
    }
    if vFlip {
        inTileY = 7 - inTileY
    }

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

    // this will need to be updated
    if palIdx == 0 {
        return
    }

    gba.applyColor(palData, uint32(index))
}

type Background struct {
    W, H uint32
    Priority uint32
    CharBaseBlock uint32
    Mosaic bool
    Palette256 bool
    ScreenBaseBlock uint32
    AffineWrap bool
    Size uint32
    XOffset, YOffset uint32
    Affine bool

    // need hof and vof
}

func NewBackground(cnt, hof, vof uint32, affine bool) *Background {
    bg := &Background{}
    bg.Affine = affine
    bg.Priority = utils.GetVarData(cnt, 0, 1)
    //bg.CharBaseBlock = utils.GetVarData(cnt, 2, 5)
    bg.CharBaseBlock = (cnt >> 2) & 0b11
    bg.Mosaic = utils.BitEnabled(cnt, 6)
    bg.Palette256 = utils.BitEnabled(cnt, 7)
    bg.ScreenBaseBlock = utils.GetVarData(cnt, 8, 12)
    bg.AffineWrap = utils.BitEnabled(cnt, 13)
    bg.Size = utils.GetVarData(cnt, 14, 15)
    bg.XOffset = utils.GetVarData(hof, 0, 9)
    bg.YOffset = utils.GetVarData(vof, 0, 9)

    bg.setSize()
    return bg
}

func (bg *Background) setSize() {

    // need to early escape if affine
    if bg.Affine {
        switch bg.Size {
        case 0: bg.W, bg.H = 16, 16
        case 1: bg.W, bg.H = 32, 32
        case 2: bg.W, bg.H = 64, 64
        case 3: bg.W, bg.H = 128, 128
        default: panic("PROHIBITTED AFFINE BG SIZE")
        }

        return
    }

    switch bg.Size {
    case 0: bg.W, bg.H = 32, 32
    case 1: bg.W, bg.H = 64, 32
    case 2: bg.W, bg.H = 32, 64
    case 3: bg.W, bg.H = 64, 64
    default: panic("PROHIBITTED BG SIZE")
    }
}

func (gba *GBA) updateMode3() {


	const (
		SIZE           = 0x12C00
		BASE           = 0x0600_0000
		BYTE_PER_PIXEL = 2
	)

	Mem := gba.Mem

	index := 0
	for i := uint32(0); i < SIZE; i += BYTE_PER_PIXEL {
		data := Mem.Read16(BASE + i)
        gba.applyColor(data, uint32(index))
		index += 4
	}
}

func (gba *GBA) updateMode4(dispcnt *Dispcnt) {

	const (
		SIZE = 0x9600
	)

    BASE := uint32(0x0600_0000)

    if dispcnt.DisplayFrame1 {
        BASE += 0xA_000
    }

	Mem := gba.Mem

	index := 0
	for i := uint32(0); i < SIZE; i++ {

		palIdx := Mem.Read8(BASE + i)

		palData := gba.getPalette(uint32(palIdx), 0, false)

        gba.applyColor(palData, uint32(index))
		index += 4
	}
}

func (gba *GBA) updateMode5(dispcnt *Dispcnt) {

	const (
		SIZE           = 0xA000
		BYTE_PER_PIXEL = 2
        MAP_WIDTH = 160
        MAP_HEIGHT = 128
	)

    BASE := uint32(0x0600_0000)

    if dispcnt.DisplayFrame1 {
        BASE += 0xA_000
    }

	Mem := gba.Mem

	index := 0
    i := uint32(0)
    for range MAP_HEIGHT {
        for range MAP_WIDTH {
            data := Mem.Read16(BASE + i)
            gba.applyColor(data, uint32(index))
            index += 4
            i += 2
        }

        index += 4 * (SCREEN_WIDTH - MAP_WIDTH) // map diff screen width and map width
    }
}

func (gba *GBA) getPalette(palIdx uint32, paletteNum uint32, obj bool) uint32 {
	pram := gba.Mem.PRAM

    addr := (paletteNum * 0x20) + palIdx * 2

    if obj {
        addr += 0x200
    }

	return uint32(pram[addr]) | uint32(pram[addr+1])<<8
}

func (gba *GBA) debugPalette() {

	// prints single palette in corner
	// palIdx is idx of palette not memory address (which is palIdx * 2)

	palIdx := 0x20
	index := 0
	for y := range 8 {
		iY := SCREEN_WIDTH * y

		for x := range 8 {
			palData := gba.getPalette(uint32(palIdx), 0, false)
			index = (iY + x) * 4
            gba.applyColor(palData, uint32(index))
		}
	}
}

func (gba *GBA) applyColor(data, index uint32) {
	r := uint8((data) & 0b11111)
	g := uint8((data >> 5) & 0b11111)
	b := uint8((data >> 10) & 0b11111)
	c := convertTo24bit(r, g, b)

	(*gba.Pixels)[index] = c.R
	(*gba.Pixels)[index+1] = c.G
	(*gba.Pixels)[index+2] = c.B
	(*gba.Pixels)[index+3] = c.A
}

func (gba *GBA) applyDebugColor(data, index uint32) {
	r := uint8((data) & 0b11111)
	g := uint8((data >> 5) & 0b11111)
	b := uint8((data >> 10) & 0b11111)
	c := convertTo24bit(r, g, b)

	(*gba.DebugPixels)[index] = c.R
	(*gba.DebugPixels)[index+1] = c.G
	(*gba.DebugPixels)[index+2] = c.B
	(*gba.DebugPixels)[index+3] = c.A
}

func (gba *GBA) getTiles(baseAddr, count int, obj, palette256 bool) {

	// base addr usually inc of 0x4000 over 0x0600_0000
	// count is # of tiles to view

	for offset := range count {
		tileOffset := offset * 0x20
		tileAddr := baseAddr + tileOffset
		gba.getTile(uint(tileAddr), 8, offset, 0, obj, palette256)
	}
}

func (gba *GBA) getTile(tileAddr uint, tileSize, xOffset, yOffset int, obj, palette256 bool) {

	xOffset *= tileSize
	yOffset *= tileSize

	indexOffset := xOffset + (yOffset * SCREEN_WIDTH)

	mem := gba.Mem
	index := 0
	byteOffset := 0

	for y := range 8 {

		iY := SCREEN_WIDTH * y

		for x := range 8 {

			tileData := mem.Read16(uint32(tileAddr) + uint32(byteOffset))

            //fmt.Printf("%08X %08X\n", tileAddr, mem.VRAM[0x20])

            var palIdx uint32
            if !palette256 {
                bitDepth := 4
                palIdx = (tileData >> uint32(bitDepth)) & 0b1111
                if x%2 == 0 {
                    palIdx = tileData & 0b1111
                }
            } else {
                palIdx = tileData & 0b1111_1111
            }


			palData := gba.getPalette(uint32(palIdx), 0, obj)
			index = (iY + x + indexOffset) * 4

			gba.applyColor(palData, uint32(index))

            if !palette256 {

                if x%2 == 1 {
                    byteOffset += 1
                }

            } else {
                byteOffset += 1
            }
		}
	}
}

type Windows struct {
    Win0, Win1, WinObj *Window
    OutBg0, OutBg1, OutBg2, OutBg3, OutObj, OutBld bool
}

type Window struct {
    Enabled bool
    L, R, T, B uint32
    InBg0, InBg1, InBg2, InBg3, InObj, InBld bool
}

func NewWindows(dispcnt *Dispcnt, gba *GBA) *Windows {

    wins := &Windows{
        Win0: &Window{},
        Win1: &Window{},
        WinObj: &Window{},
    }

    mem := gba.Mem

    if dispcnt.DisplayWin0 {
        wins.Win0.Enabled = true

        winH := mem.Read16(0x400_0040)
        wins.Win0.R = utils.GetVarData(winH, 0, 7) + 1
        wins.Win0.L = utils.GetVarData(winH, 8, 15)

        if wins.Win0.R  > 240 || wins.Win0.L > wins.Win0.R {
            wins.Win0.R = 240
        }

        winV := mem.Read16(0x400_0044)
        wins.Win0.B = utils.GetVarData(winV, 0, 7) + 1
        wins.Win0.T = utils.GetVarData(winV, 8, 15)


        if wins.Win0.B  > 160 || wins.Win0.T > wins.Win0.B {
            wins.Win0.B = 160
        }

        winIn := mem.Read16(0x400_0048)
        wins.Win0.InBg0 = utils.BitEnabled(winIn, 0)
        wins.Win0.InBg1 = utils.BitEnabled(winIn, 1)
        wins.Win0.InBg2 = utils.BitEnabled(winIn, 2)
        wins.Win0.InBg3 = utils.BitEnabled(winIn, 3)
        wins.Win0.InObj = utils.BitEnabled(winIn, 4)
        wins.Win0.InBld = utils.BitEnabled(winIn, 5)
    }

    if dispcnt.DisplayWin1 {
        wins.Win1.Enabled = true

        winH := mem.Read16(0x400_0042)
        wins.Win1.R = utils.GetVarData(winH, 0, 7) + 1
        wins.Win1.L = utils.GetVarData(winH, 8, 15)

        if wins.Win1.R  > 240 || wins.Win1.L > wins.Win1.R {
            wins.Win1.R = 240
        }

        winV := mem.Read16(0x400_0046)
        wins.Win1.B = utils.GetVarData(winV, 0, 7) + 1
        wins.Win1.T = utils.GetVarData(winV, 8, 15)

        if wins.Win1.B  > 160 || wins.Win1.T > wins.Win1.B {
            wins.Win1.B = 160
        }

        winIn := mem.Read16(0x400_0048)
        wins.Win1.InBg0 = utils.BitEnabled(winIn, 8)
        wins.Win1.InBg1 = utils.BitEnabled(winIn, 9)
        wins.Win1.InBg2 = utils.BitEnabled(winIn, 10)
        wins.Win1.InBg3 = utils.BitEnabled(winIn, 11)
        wins.Win1.InObj = utils.BitEnabled(winIn, 12)
        wins.Win1.InBld = utils.BitEnabled(winIn, 13)
    }

    if dispcnt.DisplayObjWin {
        panic("OBJ WINDOW UNSET") // i dont know how dims are set with obj mode objs

        wins.WinObj.Enabled = true
        winOut := mem.Read16(0x400_004A)
        wins.WinObj.InBg0 = utils.BitEnabled(winOut, 8)
        wins.WinObj.InBg1 = utils.BitEnabled(winOut, 9)
        wins.WinObj.InBg2 = utils.BitEnabled(winOut, 10)
        wins.WinObj.InBg3 = utils.BitEnabled(winOut, 11)
        wins.WinObj.InObj = utils.BitEnabled(winOut, 12)
        wins.WinObj.InBld = utils.BitEnabled(winOut, 13)
    }

    winOut := mem.Read16(0x400_004A)
    wins.OutBg0 = utils.BitEnabled(winOut, 0)
    wins.OutBg1 = utils.BitEnabled(winOut, 1)
    wins.OutBg2 = utils.BitEnabled(winOut, 2)
    wins.OutBg3 = utils.BitEnabled(winOut, 3)
    wins.OutObj = utils.BitEnabled(winOut, 4)
    wins.OutBld = utils.BitEnabled(winOut, 5)

    return wins
}
