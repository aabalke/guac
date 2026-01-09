//go:build rc

package ppu

import (
	simd "simd/archsimd"
	"unsafe"
)

func isSimd() bool {
	return simd.X86Features{}.AVX512()
}

func (ppu *PPU) vramDisplay(y uint32, engine *Engine) {

	ppu.setRawBitmap(engine, y, &engine.Simd.PalData)
	ptr := unsafe.Pointer(&engine.Pixels[(y*SCREEN_WIDTH)*4])
	engine.MasterBright.ApplySIMD(&engine.Simd.PalData, ptr)
}

func (ppu *PPU) setRawBitmap(engine *Engine, y uint32, pals *[SCREEN_WIDTH]uint32) {

	var bank *[0x20000]uint8

	switch bankIdx := engine.Dispcnt.VramBlock; bankIdx {
	case 0:
		bank = &ppu.Vram.A
	case 1:
		bank = &ppu.Vram.B
	case 2:
		bank = &ppu.Vram.C
	case 3:
		bank = &ppu.Vram.D
	}

	var (
		height = simd.BroadcastUint32x16(y * SCREEN_WIDTH)
		idxs   = [16]uint32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
		idx    = simd.LoadUint32x16(&idxs)
	)

	for x := uint32(0); x < SCREEN_WIDTH; x += SCREEN_WIDTH / 16 {
		calc := simd.BroadcastUint32x16(x)
		calc = calc.Add(idx)
		calc = calc.Add(height)
		calc = calc.ShiftAllLeft(1)
		addrs := [16]uint32{}
		calc.Store(&addrs)

		for i, addr := range addrs {
			pals[x+uint32(i)] = uint32(b16(bank[addr:]))
		}
	}
}

func (ppu *PPU) standard(y uint32, engine *Engine) {

	wins := &engine.Windows
	backdrop := ppu.getPalette(0, 0, false, engine.IsB)
	bld := &engine.Blend

	ResetBlendPalettesSIMD(engine.Simd, backdrop, bld)

	for x := range uint32(SCREEN_WIDTH) {

		bldPal := &engine.Simd.BlendPalettes[x]

		isSemiTransparent, inObjWindow := ppu.render(x, y, engine, bldPal)

		winBlend := !WindowBldPixelAllowed(x, y, wins, inObjWindow)

		bldPal.GetBlendModeSIMD(engine.Simd, winBlend, isSemiTransparent, bld, x)

		//engine.Simd.PalData[x] = bldPal.Blend(winBlend, isSemiTransparent, bld)
	}

	BlendSIMD(engine.Simd, bld)

	ptr := unsafe.Pointer(&engine.Pixels[(y*SCREEN_WIDTH)*4])
	engine.MasterBright.ApplySIMD(&engine.Simd.PalData, ptr)
}
