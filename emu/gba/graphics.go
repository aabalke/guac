package gba

import (
)

func (gba *GBA) graphics() {

	//bg0Priority := Bg0Control.getPriority()
	//bg1Priority := Bg1Control.getPriority()
	//bg2Priority := Bg2Control.getPriority()
	//bg3Priority := Bg3Control.getPriority()

	//fmt.Printf("PRIORITY %X, %X, %X, %X\n", bg0Priority, bg1Priority, bg2Priority, bg3Priority)

	//for i := 3; i >= 0; i-- {

	//	switch i {
	//	case int(bg0Priority):
	//	case int(bg1Priority):
	//	case int(bg2Priority):
	//	case int(bg3Priority):
	//		gba.updateDisplay()
	//	}
	//}
	//gba.updateDisplay()

	gba.updateBg2()
}

func (gba *GBA) updateBg2() {

	Mem := gba.Mem

	index := 0
	for y := range SCREEN_HEIGHT {
		for x := range SCREEN_WIDTH {

			palIdx := Mem.VRAM[(y*SCREEN_WIDTH)+x]

			palData := uint32(Mem.PRAM[(palIdx*2)+1])<<8 | uint32(Mem.PRAM[palIdx*2])

			r := uint8((palData) & 0b11111)
			g := uint8((palData >> 5) & 0b11111)
			b := uint8((palData >> 10) & 0b11111)

			c := convertTo24bit(r, g, b)

			(*gba.Pixels)[index] = c.R
			(*gba.Pixels)[index+1] = c.G
			(*gba.Pixels)[index+2] = c.B
			(*gba.Pixels)[index+3] = c.A
			index += 4
		}
	}
}

func (gba *GBA) getPalette(palIdx uint32) uint32 {
	pram := gba.Mem.PRAM
	return uint32(pram[palIdx*2]) | uint32(pram[palIdx*2+1])<<8
}

func (gba *GBA) debugPalette() {

	// prints single palette in corner
	// palIdx is idx of palette not memory address (which is palIdx * 2)

	palIdx := 0xF
	index := 0
	for y := range 8 {
		iY := SCREEN_WIDTH * y

		for x := range 8 {
			palData := gba.getPalette(uint32(palIdx))
			r := uint8((palData) & 0b11111)
			g := uint8((palData >> 5) & 0b11111)
			b := uint8((palData >> 10) & 0b11111)

			c := convertTo24bit(r, g, b)

			index = (iY + x) * 4

			(*gba.Pixels)[index] = c.R
			(*gba.Pixels)[index+1] = c.G
			(*gba.Pixels)[index+2] = c.B
			(*gba.Pixels)[index+3] = c.A
		}
	}
}

func (gba *GBA) applyColor(palData, index uint32) {
	r := uint8((palData) & 0b11111)
	g := uint8((palData >> 5) & 0b11111)
	b := uint8((palData >> 10) & 0b11111)
	c := convertTo24bit(r, g, b)

	(*gba.Pixels)[index] = c.R
	(*gba.Pixels)[index+1] = c.G
	(*gba.Pixels)[index+2] = c.B
	(*gba.Pixels)[index+3] = c.A
}

func (gba *GBA) updateBg0() {

	base := uint32(0x0600_0000)
	tileBaseAddr := base + uint32(Bg0Control.getCharacterBaseBlock())*0x4000+0x4000
	mapBaseAddr := base + uint32(Bg0Control.getScreenBaseBlock())*0x800

	for y := range 0x14 {
		for x := range 0x1E {

			addr := int(mapBaseAddr) + (x+(y*0x20))*2

			v := uint32(tileBaseAddr) + 0x20*gba.Mem.Read8(uint32(addr))

			gba.getTile(uint(v), 8, x, y)
		}
	}
}

func (gba *GBA) getTiles(baseAddr, count int) {

    // base addr usually inc of 0x4000 over 0x0600_0000
    // count is # of tiles to view

	for offset := range count {
	    tileOffset := offset * 0x20
	    tileAddr := baseAddr + tileOffset
	    gba.getTile(uint(tileAddr), 8, offset, 0)
	}
}

func (gba *GBA) getTile(tileAddr uint, tileSize, xOffset, yOffset int) {

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

			palData := gba.getPalette(uint32(palIdx))
			index = (iY + x + indexOffset) * 4

			gba.applyColor(palData, uint32(index))

            if x % 2 == 1 {
                byteOffset += 1
            }
		}
	}
}
