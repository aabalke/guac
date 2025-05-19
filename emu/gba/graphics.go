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
}

func (gt *GraphicsTiming) reset() {
    gt.RefreshCycles = 0
    gt.Scanline = 0
    gt.HBlank = false
    gt.VBlank = false
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

    gt.RefreshCycles += cycles
    gt.HBlank = gt.RefreshCycles % SCANLINE > HDRAW
    gt.Scanline = gt.RefreshCycles / SCANLINE
    gt.VBlank = gt.RefreshCycles > VDRAW

    dispstat := &gt.Gba.Mem.Dispstat
    dispstat.SetVBlank(gt.VBlank)
    dispstat.SetHBlank(gt.HBlank)
    dispstat.SetVCounter(gt.Scanline)

    if gt.VBlank { gt.Gba.checkDmas(DMA_MODE_VBL) }
    if gt.HBlank { gt.Gba.checkDmas(DMA_MODE_HBL) }
}

var ( 
    _ = fmt.Sprintf("")
)

func (gba *GBA) graphics() {

	dispcnt := gba.Mem.Read16(0x0400_0000 + DISPCNT)

	mode := dispcnt & 0b111

	flip := utils.BitEnabled(dispcnt, 4)


    //gba.getTiles(0x600_0000, 0x10, false)
    //gba.debugPalette()
    //return

    gba.clear()

	switch mode {
	case 0:
		gba.updateMode0(dispcnt)
	case 1:
		gba.updateMode1(dispcnt)
	case 2:
        panic("mode 2")
		//gba.updateMode2()
	case 3:
		gba.updateMode3()
	case 4:
		gba.updateMode4(flip)
	case 5:
		gba.updateMode5(flip)
    default: panic("UNKNOWN MODE")
	}

    if obj := utils.BitEnabled(dispcnt, 12); obj {
        for i := 127; i >= 0; i-- {
            gba.object(dispcnt, uint32(i) * 0x8)
        }
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

func (gba *GBA) object(dispcnt uint32, oamOffset uint32) {

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

    bitDepth := 4
    if obj.Palette256 {
        bitDepth = 8
    }

    VRAM_BASE := int(0x0601_0000)
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

    enTileY := int(yIdx / 8)
    enTileX := int(xIdx / 8)
    inTileY := int(yIdx % 8)
    inTileX := int(xIdx % 8)

    if obj.HFlip {
        enTileX = 7 - enTileX
        inTileX = 7 - inTileX
    }
    if obj.VFlip {
        enTileY = 7 - enTileY
        inTileY = 7 - inTileY
    }

    index := (x + (y*SCREEN_WIDTH)) * 4
    var addr uint32

    tileIdx := (enTileX * 0x20) + (enTileY * 0x100)
    if !obj.OneDimensional {
        tileIdx += enTileY * 0x400 - 0x100 // offsets previous 0x100 added
    }

    //MAX_NUM_TILES := 1024 // not sure if this is correct or two wrap at (obj h * obj W)

    tileOffset := 0x20
    if obj.Palette256 { tileOffset = 0x40}

    tileIdx = ((tileIdx + int(obj.CharName * uint32(tileOffset)) ) % (1024 * 0x20))

    tileAddr := uint32(VRAM_BASE + tileIdx)

    var inTileIdx uint32
    if obj.Palette256 {
        inTileIdx = uint32(inTileX) + uint32(inTileY * 8)
    } else {
        inTileIdx = uint32(inTileX / 2) + uint32(inTileY * 4)
    }

    addr = uint32(tileAddr) + inTileIdx

    tileData := mem.Read16(addr)

    var palIdx uint32
    if obj.Palette256 {
        palIdx = tileData & 0b1111_1111
        obj.Palette = 0
    } else {
        if inTileX % 2 == 0 {
            palIdx = tileData & 0b1111
        } else {
            palIdx = (tileData >> uint32(bitDepth)) & 0b1111
        }
    }

    palData := gba.getPalette(uint32(palIdx), obj.Palette, true)

    if !inScreenBounds(int(x), int(y)) {
        return
    }

    // this will need to be updated
    if palIdx == 0 {
        return
    }
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

    bitDepth := 4
    if obj.Palette256 {
        bitDepth = 8
    }

    // assummes 4bit

    VRAM_BASE := int(0x0601_0000)
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

    enTileY := int(yIdx / 8)
    enTileX := int(xIdx / 8)
    inTileY := int(yIdx % 8)
    inTileX := int(xIdx % 8)

    if obj.HFlip {
        enTileX = 7 - enTileX
        inTileX = 7 - inTileX
    }
    if obj.VFlip {
        enTileY = 7 - enTileY
        inTileY = 7 - inTileY
    }

    index := (x + (y*SCREEN_WIDTH)) * 4
    var addr uint32

    tileIdx := (enTileX * 0x20) + (enTileY * 0x100)
    if !obj.OneDimensional {
        tileIdx += enTileY * 0x400 - 0x100 // offsets previous 0x100 added
    }

    //MAX_NUM_TILES := 1024 // not sure if this is correct or two wrap at (obj h * obj W)

    tileOffset := 0x20
    if obj.Palette256 { tileOffset = 0x40}

    tileIdx = ((tileIdx + int(obj.CharName * uint32(tileOffset)) ) % (1024 * 0x20))

    tileAddr := uint32(VRAM_BASE + tileIdx)

    var inTileIdx uint32
    if obj.Palette256 {
        inTileIdx = uint32(inTileX) + uint32(inTileY * 8)
    } else {
        inTileIdx = uint32(inTileX / 2) + uint32(inTileY * 4)
    }

    addr = uint32(tileAddr) + inTileIdx

    tileData := mem.Read16(addr)

    var palIdx uint32
    if obj.Palette256 {
        palIdx = tileData & 0b1111_1111
        obj.Palette = 0
    } else {
        if inTileX % 2 == 0 {
            palIdx = tileData & 0b1111
        } else {
            palIdx = (tileData >> uint32(bitDepth)) & 0b1111
        }
    }

    palData := gba.getPalette(uint32(palIdx), obj.Palette, true)

    if !inScreenBounds(int(x), int(y)) {
        return
    }

    // this will need to be updated
    if palIdx == 0 {
        return
    }
    gba.applyColor(palData, uint32(index))
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

func NewObject(attr0, attr1, attr2, dispcnt uint32) *Object {

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

    obj.OneDimensional = utils.BitEnabled(dispcnt, 6)

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

func (gba *GBA) updateMode0(dispcnt uint32) {

    priorities := gba.getBgPriority(0)
    bgToggle := utils.GetByte(dispcnt, 8)

    for i := range priorities {

        if skipBgLayer := (bgToggle >> i) & 1 != 1; skipBgLayer {
            continue
        }

        controlAddr := uint32(0x0400_0008 + ((i) * 0x2))

        gba.background(controlAddr, false)
    }
}

func (gba *GBA) updateMode1(dispcnt uint32) {

    priorities := gba.getBgPriority(0)
    bgToggle := utils.GetByte(dispcnt, 8)

    for i := range priorities {

        if skipBgLayer := (bgToggle >> i) & 1 != 1; skipBgLayer {
            continue
        }

        controlAddr := uint32(0x0400_0008 + ((i) * 0x2))

        if i == 2 {
            gba.background(controlAddr, true)
            continue
        }

        gba.background(controlAddr, false)
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

func (gba *GBA) background(base uint32, affine bool) {

    mem := gba.Mem

    cnt := mem.Read16(base)
    hof := mem.Read16(base + 8)
    vof := mem.Read16(base + 10)
    bg := NewBackground(cnt, hof, vof, affine)

    x, y := uint32(0), uint32(0)
    for y = range SCREEN_HEIGHT {
        for x = range SCREEN_WIDTH {

            if affine {
                gba.setAffineBackgroundPixel(bg, x, y)
                continue
            }

            gba.setBackgroundPixel(bg, x, y)
        }
    }
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
    bg.CharBaseBlock = utils.GetVarData(cnt, 2, 5)
    bg.Mosaic = utils.BitEnabled(cnt, 6)
    bg.Palette256 = utils.BitEnabled(cnt, 7)
    bg.ScreenBaseBlock = utils.GetVarData(cnt, 8, 12)
    bg.AffineWrap = utils.BitEnabled(cnt, 13)
    bg.Size = utils.GetVarData(cnt, 14, 15)
    bg.XOffset = utils.GetVarData(hof, 0, 8)
    bg.YOffset = utils.GetVarData(vof, 0, 8)

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

func (gba *GBA) updateMode4(flip bool) {

	const (
		SIZE = 0x9600
	)

    BASE := uint32(0x0600_0000)

    if flip {
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

func (gba *GBA) updateMode5(flip bool) {

	const (
		SIZE           = 0xA000
		BYTE_PER_PIXEL = 2
        MAP_WIDTH = 160
        MAP_HEIGHT = 128
	)

    BASE := uint32(0x0600_0000)

    if flip {
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

func (gba *GBA) getTiles(baseAddr, count int, palette256 bool) {

	// base addr usually inc of 0x4000 over 0x0600_0000
	// count is # of tiles to view

	for offset := range count {
		tileOffset := offset * 0x20
		tileAddr := baseAddr + tileOffset
		gba.getTile(uint(tileAddr), 8, offset, 0, false, palette256)
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


			palData := gba.getPalette(uint32(palIdx), 0, false)
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
