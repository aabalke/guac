package nds

func (nds *Nds) graphics(y uint32) {

    //nds.DebugPalette(y)
    //return

    switch nds.ppu.Dispcnt.DisplayMode {
    case 0:
    case 2: nds.bitmap(y)
    }
}

func (nds *Nds) bitmap(y uint32) {

    // temp offset for vram banking
    offset := uint32(0x80_0000)

    x := uint32(0)
    for x = range SCREEN_WIDTH {

        const (
            BYTE_PER_PIXEL = 2
        )

        idx := (x+(y*SCREEN_WIDTH))*BYTE_PER_PIXEL + offset // no region

        data := uint32(nds.mem.Vram.Read(idx, true))
        data |= uint32(nds.mem.Vram.Read(idx+1, true)) << 8

        r := uint8((data) & 0b11111)
        g := uint8((data >> 5) & 0b11111)
        b := uint8((data >> 10) & 0b11111)

        r = (r << 3) | (r >> 2)
        g = (g << 3) | (g >> 2)
        b = (b << 3) | (b >> 2)

        index := (x + (y * SCREEN_WIDTH)) * 4
        nds.PixelsTop[index] = r
        nds.PixelsTop[index+1] = g
        nds.PixelsTop[index+2] = b
        nds.PixelsTop[index+3] = 0xFF
    }
}

func(nds *Nds) DebugPalette(y uint32) {

    x := uint32(0)
    for x = range SCREEN_WIDTH {

        index := (x + (y * SCREEN_WIDTH)) * 4

        data := uint32(nds.mem.Pram[0x21F])


        nds.applyColor(data, index)
    }
}

func (nds *Nds) applyColor(data, i uint32) {
	r := uint8((data) & 0b11111)
	g := uint8((data >> 5) & 0b11111)
	b := uint8((data >> 10) & 0b11111)

	r = (r << 3) | (r >> 2)
	g = (g << 3) | (g >> 2)
	b = (b << 3) | (b >> 2)

	nds.PixelsTop[i] = r
	nds.PixelsTop[i+1] = g
	nds.PixelsTop[i+2] = b
	nds.PixelsTop[i+3] = 0xFF
}


