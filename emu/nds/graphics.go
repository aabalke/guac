package nds

import (

	"github.com/aabalke/guac/emu/nds/ppu"
)

func (nds *Nds) graphics(y uint32) {

    a := &nds.ppu.EngineA
    b := &nds.ppu.EngineB

	switch a.Dispcnt.DisplayMode {
	//case 0:
	//	nds.screenoff(y, a)
	case 1:
		nds.standard(y, a)
	case 2:
		nds.bitmap(y, a)
	case 3:
		panic("UNSETUP MAIN MEMORY DISPLAY")
	}

    switch b.Dispcnt.DisplayMode {
    //case 0:
    //    nds.screenoff(y, b)
    case 1:
        nds.tileBg0(y, b)
    }
}

func (nds *Nds) screenoff(y uint32, engine *ppu.Engine) {

	x := uint32(0)
	for x = range SCREEN_WIDTH {

		index := (x + (y * SCREEN_WIDTH)) * 4
		(*engine.Pixels)[index]   = 0xFF
		(*engine.Pixels)[index+1] = 0xFF
		(*engine.Pixels)[index+2] = 0xFF
		(*engine.Pixels)[index+3] = 0xFF
	}
}

func (nds *Nds) standard(y uint32, engine *ppu.Engine) {

	//switch nds.ppu.DispcntA.Mode {
	//case 0:
	//case 1:
	//case 2:
	//case 3:
	//case 4:
	//case 5:
	//    switch nds.ppu.DispcntA
	//case 6:
	//}

	//if nds.ppu.DispcntA.Mode != 6 {
	//    return
	//}

	offset := uint32(0)

	x := uint32(0)
	for x = range SCREEN_WIDTH {

		const (
			BYTE_PER_PIXEL = 1
		)

		idx := (x+(y*SCREEN_WIDTH))*BYTE_PER_PIXEL + offset // no region

		palIdx := uint32(nds.mem.Vram.Read(idx, true))

		data := uint32(nds.mem.Pram[palIdx])

		r := uint8((data) & 0b11111)
		g := uint8((data >> 5) & 0b11111)
		b := uint8((data >> 10) & 0b11111)

		r = (r << 3) | (r >> 2)
		g = (g << 3) | (g >> 2)
		b = (b << 3) | (b >> 2)

		index := (x + (y * SCREEN_WIDTH)) * 4
		(*engine.Pixels)[index] = r
		(*engine.Pixels)[index+1] = g
		(*engine.Pixels)[index+2] = b
		(*engine.Pixels)[index+3] = 0xFF
	}
}

func (nds *Nds) tileBg0(y uint32, engine *ppu.Engine) {

    if !engine.Backgrounds[0].Enabled {
        return
    }

    updateBackgrounds(engine)

    x := uint32(0)
    for x = range SCREEN_WIDTH {
        palData, ok := nds.setBackgroundPixel(&engine.Backgrounds[0], x, y)

        if ok {
            index := (x + (y * SCREEN_WIDTH)) << 2
            nds.applyColor(palData, index, engine.Pixels)
        }
    }
}

func updateBackgrounds(engine *ppu.Engine) *[4]ppu.Background {

    bgs := &engine.Backgrounds
    //dispcnt := &engine.Dispcnt

	for i := range 4 {
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

func (nds *Nds) setBackgroundPixel(bg *ppu.Background, x, y uint32) (uint32, bool) {

	xIdx := (x) & ((bg.W) - 1)
	yIdx := (y) & ((bg.H) - 1)
	//xIdx := (x + bg.XOffset) & ((bg.W) - 1)
	//yIdx := (y + bg.YOffset) & ((bg.H) - 1)

//	if bg.Mosaic && nds.PPU.Mosaic.BgH != 0 {
//		xIdx -= xIdx % (nds.PPU.Mosaic.BgH + 1)
//	}
//
//	if bg.Mosaic && nds.PPU.Mosaic.BgV != 0 {
//		yIdx -= yIdx % (nds.PPU.Mosaic.BgV + 1)
//	}

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

	//mapAddr &= 0x1FFFF

	//if mapAddr >= 0x18000 {
	//    mapAddr -= 0x8000
	//}

    screenData := nds.mem.Read16(mapAddr + 0x620_0000, true)

	tileIdx := (screenData & 0b11_1111_1111) << 5


	tileAddr := bg.CharBaseBlock + tileIdx
	if bg.Palette256 {
		tileAddr += tileIdx
	}

	//if inObjTiles := tileAddr >= 0x1_0000; inObjTiles {
	//	return 0, false
	//}

	inTileX, inTileY := getPositionsBg(screenData, xIdx, yIdx)

	var inTileIdx uint32
	if bg.Palette256 {
		inTileIdx = inTileX + (inTileY << 3)
	} else {
		inTileIdx = (inTileX >> 1) + (inTileY << 2)
	}

    tileData := nds.mem.Read32(tileAddr + inTileIdx + 0x620_0000, true)

	if bg.Palette256 {
		palIdx := tileData
		if palIdx == 0 {
			return 0, false
		}

        return uint32(nds.mem.Pram[palIdx]), true
	}

	palIdx := (tileData >> ((inTileX & 1) << 2)) & 0xF

	//if palIdx == 0 {
	//	return 0, false
	//}

    pramOffset := uint32(0x400)

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


func (nds *Nds) bitmap(y uint32, engine *ppu.Engine) {

    bankIdx := engine.Dispcnt.VramBlock

    var bank [0x20000]uint8

    switch bankIdx {
    case 0: bank = nds.mem.Vram.A
    case 1: bank = nds.mem.Vram.B
    case 2: bank = nds.mem.Vram.C
    case 3: bank = nds.mem.Vram.D
    }

	x := uint32(0)
	for x = range SCREEN_WIDTH {

		const (
			BYTE_PER_PIXEL = 2
		)

		idx := (x+(y*SCREEN_WIDTH)) * BYTE_PER_PIXEL

        data := uint32(bank[idx])
        data |= uint32(bank[idx + 1]) << 8

		r := uint8((data) & 0b11111)
		g := uint8((data >> 5) & 0b11111)
		b := uint8((data >> 10) & 0b11111)

		r = (r << 3) | (r >> 2)
		g = (g << 3) | (g >> 2)
		b = (b << 3) | (b >> 2)

		index := (x + (y * SCREEN_WIDTH)) * 4
		(*engine.Pixels)[index] = r
		(*engine.Pixels)[index+1] = g
		(*engine.Pixels)[index+2] = b
		(*engine.Pixels)[index+3] = 0xFF
	}
}

func (nds *Nds) DebugPalette(y uint32) {

	x := uint32(0)
	for x = range SCREEN_WIDTH {

		index := (x + (y * SCREEN_WIDTH)) * 4

		data := uint32(nds.mem.Pram[0x5FE >> 1])

		nds.applyColor(data, index, &nds.PixelsTop)
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
