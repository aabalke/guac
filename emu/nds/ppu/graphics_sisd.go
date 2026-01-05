//go:build !rc
package ppu

func (ppu *PPU) vramDisplay(y uint32, engine *Engine) {

	x := uint32(0)
	for x = range SCREEN_WIDTH {
		palData, _ := ppu.setRawBitmap(engine, x, y)
		r, g, b := engine.MasterBright.Apply(palData)
		i := (x + (y * SCREEN_WIDTH)) << 2

		(engine.Pixels)[i] = r
		(engine.Pixels)[i+1] = g
		(engine.Pixels)[i+2] = b
		(engine.Pixels)[i+3] = 0xFF
	}
}

func (ppu *PPU) standard(y uint32, engine *Engine) {

	wins := &engine.Windows
	backdrop := ppu.getPalette(0, 0, false, engine.IsB)
    bld := &engine.Blend

	for x := range uint32(SCREEN_WIDTH) {
		bldPal := NewBlendPalette(&engine.Blend, backdrop)
		isSemiTransparent, inObjWindow := ppu.render(x, y, engine, bldPal)
        winBlend := !WindowBldPixelAllowed(x, y, wins, inObjWindow)
		palData := bldPal.Blend(winBlend, isSemiTransparent, bld)

		r, g, b := engine.MasterBright.Apply(palData)
		i := (x + (y * SCREEN_WIDTH)) << 2
		(engine.Pixels)[i] = r
		(engine.Pixels)[i+1] = g
		(engine.Pixels)[i+2] = b
		(engine.Pixels)[i+3] = 0xFF
	}
}

func (ppu *PPU) setRawBitmap(engine *Engine, x, y uint32) (uint32, bool) {

	addr := uint32(x+(y*SCREEN_WIDTH)) * 2

	bankIdx := engine.Dispcnt.VramBlock

	var bank *[0x20000]uint8

	switch bankIdx {
	case 0:
		bank = &ppu.Vram.A
	case 1:
		bank = &ppu.Vram.B
	case 2:
		bank = &ppu.Vram.C
	case 3:
		bank = &ppu.Vram.D
	}

	//bank = &ppu.Vram.A
	//data := uint32(binary.LittleEndian.Uint16(bank[addr:]) &^ 0x80)

	data := uint32(b16(bank[addr:]))

	return data, true

}
