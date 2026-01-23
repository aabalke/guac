package ppu

func (ppu *PPU) screenoff(y uint32, e *Engine) {
	copy(e.Pixels[y*SCREEN_WIDTH*4:(y+1)*SCREEN_WIDTH*4],
    ppu.WHITE_SCANLINE)
}

func (ppu *PPU) vramDisplay(y uint32, e *Engine) {

    var bank *[0x20000]uint8
    switch bankIdx := e.Dispcnt.VramBlock; bankIdx {
    case 0:
        bank = &ppu.Vram.A
    case 1:
        bank = &ppu.Vram.B
    case 2:
        bank = &ppu.Vram.C
    case 3:
        bank = &ppu.Vram.D
    }

    addr := (y * SCREEN_WIDTH) * 2
    i    := (y * SCREEN_WIDTH) * 4
    for range uint32(SCREEN_WIDTH) {

        data := uint32(b16(bank[addr:]))
		r, g, b := e.MasterBright.Apply(data)

		(e.Pixels)[i+0] = r
		(e.Pixels)[i+1] = g
		(e.Pixels)[i+2] = b
		(e.Pixels)[i+3] = 0xFF

        i += 4
        addr += 2
	}
}

func (ppu *PPU) standard(y uint32, engine *Engine) {

	wins := &engine.Windows
	backdrop := uint32(*engine.Backdrop)
	bld := &engine.Blend

    for i := range uint32(4) {

        if !engine.Backgrounds[i].MasterEnabled {
            continue
        }

        switch engine.Backgrounds[i].Type {
        case BG_TYPE_DIR:
            ppu.directBmpScanline(engine, i, y)
        case BG_TYPE_TEX:
            ppu.tiledScanline(engine, i, y)
        case BG_TYPE_256:
            ppu.bmpScanline(engine, i, y)
        case BG_TYPE_BGM:
            ppu.affine16Scanline(engine, i)
        case BG_TYPE_3D:
            ppu.threeScanline(engine, i, y)
        case BG_TYPE_AFF:
            ppu.affineScanline(engine, i)
        }
    }

	for x := range uint32(SCREEN_WIDTH) {
		bldPal := NewBlendPalette(&engine.Blend, backdrop)
		isSemiTransparent, inObjWindow := ppu.render(x, y, engine, bldPal)

        winBlend := wins.Enabled && !wins.inWinBld(x, y, inObjWindow)
		palData := bldPal.Blend(winBlend, isSemiTransparent, bld)

		r, g, b := engine.MasterBright.Apply(palData)
		i := (x + (y * SCREEN_WIDTH)) << 2
		(engine.Pixels)[i] = r
		(engine.Pixels)[i+1] = g
		(engine.Pixels)[i+2] = b
		(engine.Pixels)[i+3] = 0xFF
	}
}
