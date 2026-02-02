package ppu

import (
	"encoding/binary"
	"unsafe"
)

func (ppu *PPU) Graphics(y, frame uint32) {

	a := &ppu.EngineA
	b := &ppu.EngineB

	for i := range 4 {
		a.Backgrounds[i].SetSize()
		b.Backgrounds[i].SetSize()
	}

	a.getBgPriority(y)
	b.getBgPriority(y)
	a.getObjPriority(y)
	b.getObjPriority(y)

	//if frame & ppu.FrameSkipMask == 0 {
	ppu.buildFrame(y)
	//}

	a.Backgrounds[2].BgAffineUpdate()
	a.Backgrounds[3].BgAffineUpdate()
	b.Backgrounds[2].BgAffineUpdate()
	b.Backgrounds[3].BgAffineUpdate()
}

func (ppu *PPU) buildFrame(y uint32) {

	switch a := &ppu.EngineA; a.Dispcnt.DisplayMode {
	case 0:
		ppu.screenoff(y, a)
	case 1:
		ppu.standard(y, a)
		if ppu.Capture.ActiveCapture {
			ppu.Capture.CaptureLine(y, ppu.Rasterizer.Buffers.BisRendering)
		}
	case 2:
		if ppu.Capture.ActiveCapture {
			ppu.standard(y, a)
			ppu.Capture.CaptureLine(y, ppu.Rasterizer.Buffers.BisRendering)
		}

		ppu.vramDisplay(y, a)
	case 3:
		panic("main memory fifo display unsupported")
	}

	switch b := &ppu.EngineB; b.Dispcnt.DisplayMode {
	case 0:
		ppu.screenoff(y, b)
	case 1:
		ppu.standard(y, b)
	}
}

func (ppu *PPU) screenoff(y uint32, e *Engine) {
	copy(e.Pixels[y*SCREEN_WIDTH*4:(y+1)*SCREEN_WIDTH*4], ppu.WHITE_SCANLINE)
}

func (ppu *PPU) vramDisplay(y uint32, e *Engine) {

	bank := (*[0x20000]uint8)(ppu.Vram.Cnt[e.Dispcnt.VramBlock].bank)

	addr := (y * SCREEN_WIDTH) * 2
	for x := range uint32(SCREEN_WIDTH) {
		e.Blend.Blended[x] = binary.LittleEndian.Uint16(bank[addr:])
		addr += 2
	}

	p32 := (*[SCREEN_WIDTH]uint32)(unsafe.Pointer(&e.Pixels[(y*SCREEN_WIDTH)*4]))
	for x := range uint32(SCREEN_WIDTH) {
		p32[x] = e.MasterBright.LUT[e.Blend.Blended[x]&^0x8000]
	}
}

func (ppu *PPU) standard(y uint32, e *Engine) {

	ResetBlendPalettes(e)
	for x := range uint32(SCREEN_WIDTH) {
		e.Windows.inObjWindow[x] = false
	}

	for priority := 3; priority >= 0; priority-- {

		for x := range uint32(SCREEN_WIDTH) {
			e.BgOks[x] = false
			e.ObjOk[x] = false
		}

		bgPriority := &e.BgPriorities[priority]

		for j := range bgPriority.Cnt {
			bgIdx := bgPriority.Idx[j]

			if !e.Backgrounds[bgIdx].MasterEnabled {
				continue
			}

			switch e.Backgrounds[bgIdx].Type {
			case BG_TYPE_DIR:
				ppu.directBmpScanline(e, bgIdx, y)
			case BG_TYPE_TEX:
				ppu.tiledScanline(e, bgIdx, y)
			case BG_TYPE_256:
				ppu.bmpScanline(e, bgIdx, y)
			case BG_TYPE_BGM, BG_TYPE_LAR:
				ppu.affine16Scanline(e, bgIdx, y)
			case BG_TYPE_3D:
				ppu.threeScanline(e, bgIdx, y)
			case BG_TYPE_AFF:
				ppu.affineScanline(e, bgIdx, y)
			}
		}

		if bgPriority.Cnt != 0 {
			e.SetBgPals()
		}

		if !e.Dispcnt.DisplayObj {
			continue
		}

		objPriority := &e.ObjPriorities[priority]

		for j := range uint32(objPriority.Cnt) {

			i := objPriority.Idx[j]
			obj := &e.Objects[i]

			if !obj.MasterEnabled {
				continue
			}

			if bmp := obj.Mode == 3; bmp {
				if obj.RotScale {
					ppu.bitmapObjectAffine(e, i, uint32(priority), y)
				} else {
					ppu.bitmapObject(e, i, uint32(priority), y)
				}
			} else {

				if obj.RotScale {
					ppu.tiledObjectAffine(e, i, uint32(priority), y)
				} else {
					ppu.tiledObject(e, i, uint32(priority), y)
				}
			}
		}

		if wins := &e.Windows; wins.Enabled {
			for x := range uint32(SCREEN_WIDTH) {
				if e.ObjOk[x] {
					wins.inObjWindow[x] = e.ObjMode[x] == 2
				}
			}
		}

        if objPriority.Cnt != 0 {
            e.SetObjPals(uint32(priority))
        }
	}

	BlendAll(e.Blend, &e.Windows, y)

	p32 := (*[SCREEN_WIDTH]uint32)(unsafe.Pointer(&e.Pixels[(y*SCREEN_WIDTH)*4]))
	for x := range uint32(SCREEN_WIDTH) {
		p32[x] = e.MasterBright.LUT[e.Blend.Blended[x]]
	}
}
