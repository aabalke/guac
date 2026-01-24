package ppu

import (
	"encoding/binary"
	"unsafe"
)

func (ppu *PPU) Graphics(y uint32, singleThread bool) {

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

	ppu.buildFrame(y)

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
	copy(e.Pixels[y*SCREEN_WIDTH*4:(y+1)*SCREEN_WIDTH*4],ppu.WHITE_SCANLINE)
}

func (ppu *PPU) vramDisplay(y uint32, e *Engine) {

    bank := ppu.Vram.VramBlocks[e.Dispcnt.VramBlock]

    addr := (y * SCREEN_WIDTH) * 2
    for x := range uint32(SCREEN_WIDTH) {
        e.BlendedPalettes[x] = binary.LittleEndian.Uint16(bank[addr:])
        addr += 2
	}

    p32 := (*[SCREEN_WIDTH]uint32)(unsafe.Pointer(&e.Pixels[(y*SCREEN_WIDTH)*4]))
    for x := range uint32(SCREEN_WIDTH) {
        p32[x] = e.MasterBright.LUT[e.BlendedPalettes[x]]
    }
}

func (ppu *PPU) standard(y uint32, e *Engine) {

	bld := &e.Blend

    for i := range uint32(4) {

        if !e.Backgrounds[i].MasterEnabled {
            continue
        }

        switch e.Backgrounds[i].Type {
        case BG_TYPE_DIR:
            ppu.directBmpScanline(e, i, y)
        case BG_TYPE_TEX:
            ppu.tiledScanline(e, i, y)
        case BG_TYPE_256:
            ppu.bmpScanline(e, i, y)
        case BG_TYPE_BGM:
            ppu.affine16Scanline(e, i, y)
        case BG_TYPE_3D:
            ppu.threeScanline(e, i, y)
        case BG_TYPE_AFF:
            ppu.affineScanline(e, i, y)
        }
    }

    ResetBlendPalettes(e)
    Render(ppu, e, y)
    BlendAll(bld, &e.BlendPalettes, &e.BlendedPalettes)

    p32 := (*[SCREEN_WIDTH]uint32)(unsafe.Pointer(&e.Pixels[(y*SCREEN_WIDTH)*4]))
    for x := range uint32(SCREEN_WIDTH) {
        p32[x] = e.MasterBright.LUT[e.BlendedPalettes[x]]
    }
}

func Render(ppu *PPU, e *Engine, y uint32) {

    var (
        dispcnt = &e.Dispcnt
        wins    = &e.Windows
        bld     = &e.Blend
    )

    for x := range uint32(SCREEN_WIDTH) {
        bldPal := &e.BlendPalettes[x]
        inObjWindow := false

        for priority := 3; priority >= 0; priority-- {

            bgPriority := &e.BgPriorities[priority]
            for j := bgPriority.Cnt - 1; j >= 0; j-- {

                bgIdx := bgPriority.Idx[j]

                if !e.BgOks[bgIdx][x] {
                    continue
                }

                if e.Backgrounds[bgIdx].Type == BG_TYPE_3D {
                    bldPal.SetBgPalettes3d(e.BgPalettes[bgIdx][x], e.BgAlphas[bgIdx][x], bld)
                    continue
                }

                bldPal.SetBgPalettes(e.BgPalettes[bgIdx][x], bgIdx, bld)
            }

            if !dispcnt.DisplayObj {
                continue
            }

            if wins.Enabled && !wins.inWinObj(x, y) {
                continue
            }

            objPriority := &e.ObjPriorities[priority]

            ObjectLoop:
            for j := range objPriority.Cnt {

                objIdx := objPriority.Idx[j]
                obj := &e.Objects[objIdx]

                if bmp := obj.Mode == 3; bmp {
                    obj.OneDimensional = dispcnt.BitmapObj1D
                    palData, ok := ppu.setObjBmpPixel(e, obj, x, y)
                    if !ok {
                        continue
                    }
                    bldPal.SetObjPalettes(palData, false, bld)
                    break ObjectLoop
                }

                obj.OneDimensional = dispcnt.TileObj1D
                palData, ok := ppu.setObjTilePixel(e, obj, x, y)
                if !ok {
                    continue
                }

                if obj.Mode == 2 {
                    inObjWindow = true
                    break ObjectLoop
                }

                isSemiTransparent := obj.Mode == 1
                e.BlendPalettes[x].objTransparent = isSemiTransparent
                bldPal.SetObjPalettes(palData, isSemiTransparent, bld)
                break ObjectLoop
            }
        }

        winBlend := wins.Enabled && !wins.inWinBld(x, y, inObjWindow)
        e.BlendPalettes[x].winBlend = winBlend
    }
}
