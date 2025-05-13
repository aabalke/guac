package gba

import (
	"fmt"

	"github.com/aabalke33/guac/emu/gba/utils"
)

func (gba *GBA) graphics() {

	dispcnt := gba.Mem.Read16(0x0400_0000 + DISPCNT)

	mode := dispcnt & 0b111

	flip := utils.BitEnabled(dispcnt, 4)

	switch mode {
	case 0:
		gba.updateMode0()
	case 3:
		gba.updateMode3()
	case 4:
		gba.updateMode4(flip)
	case 5:
		gba.updateMode5(flip)
	}

    if obj := utils.BitEnabled(dispcnt, 12); obj {
        gba.object(dispcnt, 0)
    }

}

func (gba *GBA) object(dispcnt uint32, oamOffset uint32) {

    OAM_BASE := uint32(0x0700_0000)
    mem := gba.Mem

    attr0 := mem.Read16(OAM_BASE + oamOffset)
    attr1 := mem.Read16(OAM_BASE + oamOffset + 2)
    attr2 := mem.Read16(OAM_BASE + oamOffset + 4)

    obj := NewObject(attr0, attr1, attr2, dispcnt)

    if obj.Disable {
        return
    }

    x, y := uint32(0), uint32(0)

    for y = range SCREEN_HEIGHT {
        for x = range SCREEN_WIDTH {
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

func (gba *GBA) setObjectPixel(obj *Object, x, y uint32) {

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
        gba.applyColor(0, (x + (y * SCREEN_WIDTH))*4)
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

    bitDepth := 4
    index := (x + (y*SCREEN_WIDTH)) * 4
    var addr uint32

    tileIdx := (enTileX * 0x20) + (enTileY * 0x100)
    if !obj.OneDimensional {
        tileIdx += enTileY * 0x400 - 0x100 // offsets previous 0x100 added
    }

    //MAX_NUM_TILES := 1024 // not sure if this is correct or two wrap at (obj h * obj W)
    tileIdx = ((tileIdx + int(obj.CharName * 0x20) ) % (1024 * 0x20))

    tileAddr := uint32(VRAM_BASE + tileIdx)

    inTileIdx := uint32(inTileX / 2) + uint32(inTileY * 4)
    addr = uint32(tileAddr) + inTileIdx

    tileData := mem.Read16(addr)
    var palIdx uint32
    if inTileX % 2 == 0 {
        palIdx = tileData & 0b1111
    } else {
        palIdx = (tileData >> uint32(bitDepth)) & 0b1111
    }

    palData := gba.getPalette(uint32(palIdx), obj.Palette, true)

    if !inScreenBounds(int(x), int(y)) {
        return
    }
    // this will need to be updated
    if palIdx == 0 {
        gba.applyColor(0, (x + (y * SCREEN_WIDTH))*4)
        return
    }
    gba.applyColor(palData, uint32(index))
}

type Object struct {
    X, Y, W, H uint32
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
    obj.DoubleSize = utils.BitEnabled(attr0, 9)
    obj.Disable = utils.BitEnabled(attr0, 9)
    obj.Mode = utils.GetVarData(attr0, 10, 11)
    obj.Mosaic = utils.BitEnabled(attr0, 12)
    obj.Palette256 = utils.BitEnabled(attr0, 13)
    obj.Shape = utils.GetVarData(attr0, 14, 15)
    
    obj.X = attr1 & 0b1_1111_1111
    obj.RotParams = utils.GetVarData(attr1, 9, 13)
    obj.HFlip = utils.BitEnabled(attr1, 12)
    obj.VFlip = utils.BitEnabled(attr1, 13)
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

func (gba *GBA) updateMode0() {

    gba.background()

	//base := uint32(0x0600_0000)
	//tileBaseAddr := base + uint32(Bg0Control.getCharacterBaseBlock())*0x4000 + 0x4000
	//mapBaseAddr := base + uint32(Bg0Control.getScreenBaseBlock())*0x800
	//tileBaseAddr := base
	//mapBaseAddr := base + 0xF000

	//for y := range 0x14 {
	//	for x := range 0x1E {

	//		addr := int(mapBaseAddr) + (x+(y*0x20))*2

	//		v := uint32(tileBaseAddr) + 0x20*gba.Mem.Read8(uint32(addr))

	//		gba.getTile(uint(v), 8, x, y, false)
	//	}
	//}
}

func (gba *GBA) background() {

    mem := gba.Mem

    regAddr := uint32(0x400_0008)

    cnt := mem.Read16(regAddr)
    hof := mem.Read16(regAddr+2)
    vof := mem.Read16(regAddr+4)

    bg := NewBackground(cnt, hof, vof)

    x, y := uint32(0), uint32(0)

    for y = range SCREEN_HEIGHT {
        for x = range SCREEN_WIDTH {
            gba.setBackgroundPixel(bg, x, y)
        }
    }
}

func (gba *GBA) setBackgroundPixel(bg *Background, x, y uint32) {
    VRAM_BASE := int(0x0600_0000)
	mem := gba.Mem

    tileAddr := VRAM_BASE + (int(bg.CharBaseBlock) * 0x4000)
    //mapAddr := VRAM_BASE + (int(bg.ScreenBaseBlock) * 0x800)

    enTileY := int(y / 8)
    enTileX := int(x / 8)
    inTileY := int(y % 8)
    inTileX := int(x % 8)

    bitDepth := 4
    index := (x + (y*SCREEN_WIDTH)) * 4
    var addr uint32

    tileIdx := (enTileX * 0x20) + (enTileY * 0x100)

    tileAddr = int(tileAddr + tileIdx)

    inTileIdx := uint32(inTileX / 2) + uint32(inTileY * 4)
    addr = uint32(tileAddr) + inTileIdx

    tileData := mem.Read16(addr)
    var palIdx uint32
    if inTileX % 2 == 0 {
        palIdx = tileData & 0b1111
    } else {
        palIdx = (tileData >> uint32(bitDepth)) & 0b1111
    }

    palette := uint32(0)
    palData := gba.getPalette(uint32(palIdx), palette, true)

    if !inScreenBounds(int(x), int(y)) {
        return
    }
    // this will need to be updated
    //if palIdx == 0 {
    //    gba.applyColor(0, (x + (y * SCREEN_WIDTH))*4)
    //    return
    //}
    gba.applyColor(palData, uint32(index))

    _ = fmt.Sprintf("")
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

    // need hof and vof
}

func NewBackground(cnt, hof, vof uint32) *Background {
    bg := &Background{}
    bg.Priority = utils.GetVarData(cnt, 0, 1)
    bg.CharBaseBlock = utils.GetVarData(cnt, 2, 3)
    bg.Mosaic = utils.BitEnabled(cnt, 6)
    bg.Palette256 = utils.BitEnabled(cnt, 7)
    bg.ScreenBaseBlock = utils.GetVarData(cnt, 8, 12)
    bg.AffineWrap = utils.BitEnabled(cnt, 13)
    bg.Size = utils.GetVarData(cnt, 14, 15)

    bg.setSize()
    return bg
}

func (bg *Background) setSize() {

    // need to early escape if affine

    switch bg.Size {
    case 0: bg.H, bg.W = 32, 32
    case 1: bg.H, bg.W = 64, 32
    case 2: bg.H, bg.W = 32, 64
    case 3: bg.H, bg.W = 64, 64
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

	palIdx := 0xF
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

func (gba *GBA) getTiles(baseAddr, count int) {

	// base addr usually inc of 0x4000 over 0x0600_0000
	// count is # of tiles to view

	for offset := range count {
		tileOffset := offset * 0x20
		tileAddr := baseAddr + tileOffset
		gba.getTile(uint(tileAddr), 8, offset, 0, false)
	}
}

func (gba *GBA) getTile(tileAddr uint, tileSize, xOffset, yOffset int, obj bool) {

	xOffset *= tileSize
	yOffset *= tileSize

	indexOffset := xOffset + (yOffset * SCREEN_WIDTH)

	mem := gba.Mem
	index := 0
	bitDepth := 4
	byteOffset := 0

	for y := range 8 {

		iY := SCREEN_WIDTH * y

		for x := range 8 {

			tileData := mem.Read16(uint32(tileAddr) + uint32(byteOffset))

			palIdx := (tileData >> uint32(bitDepth)) & 0b1111
			if x%2 == 0 {
				palIdx = tileData & 0b1111
			}

			palData := gba.getPalette(uint32(palIdx), 0, obj)
			index = (iY + x + indexOffset) * 4

			gba.applyColor(palData, uint32(index))

			if x%2 == 1 {
				byteOffset += 1
			}
		}
	}
}
