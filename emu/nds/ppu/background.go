package ppu

import (
    "unsafe"

    "github.com/aabalke/guac/emu/nds/utils"
)

func (e *Engine) getBgPriority(y uint32) {

    priorities := &e.BgPriorities
    priorities[0].Cnt = 0
    priorities[1].Cnt = 0
    priorities[2].Cnt = 0
    priorities[3].Cnt = 0

	for i := range uint32(4) {
        bg := &e.Backgrounds[i]

        switch {
        case !bg.Enabled:
            bg.MasterEnabled = false
			continue

        case bg.Affine:
            // need to setup scanline check here
        case !bg.Affine:
            top := (int(y) - int(bg.YOffset)) & int((bg.H)-1)
            if (top < 0 || top-int(bg.H) >= 0) {
                bg.MasterEnabled = false
                continue
            }
        }

        bg.MasterEnabled = true
        p := &priorities[bg.Priority]
        p.Idx[p.Cnt] = i
        p.Cnt++
	}
}

func (ppu *PPU) threeScanline(e *Engine, bgIdx, y uint32) {

    bg := &e.Backgrounds[bgIdx]

    yIdx := (y + bg.YOffset) & ((bg.H) - 1)

    if yIdx >= SCREEN_HEIGHT {
        return
    }

    for x := range uint32(SCREEN_WIDTH) {

        xIdx := (x + bg.XOffset) & ((bg.W) - 1)
        i := (xIdx + (yIdx * SCREEN_WIDTH))

        if xIdx >= SCREEN_WIDTH {
			e.BgPalettes[bgIdx][x] = 0
			e.BgOks[bgIdx][x] = false
            continue
        }

        r := ppu.Rasterizer.Render

        pal, alpha := uint32(0), float32(0)

        if r.Rasterizer.Buffers.BisRendering {
            pal, alpha = r.Pixels.PalettesA[i], r.Pixels.AlphaA[i]
        } else {
            pal, alpha = r.Pixels.PalettesB[i], r.Pixels.AlphaB[i]
        }

        if noblend := ppu.EngineA.Blend.Mode == 0; noblend {

            // this is only important if the 3d screen is the only one, nothing is behind it, and alpha != 1.
            // really only noticed in devkit tests ( nehe/lesson10)

            r := float32(pal&0x1F) * alpha
            g := float32((pal>>5)&0x1F) * alpha
            b := float32((pal>>10)&0x1F) * alpha

            pal &= 0x8000
            pal |= uint32(r) & 0x1F
            pal |= ((uint32(g) & 0x1F) << 5)
            pal |= ((uint32(b) & 0x1F) << 10)
        }

        e.BgPalettes[bgIdx][x] = pal
        e.BgOks[bgIdx][x] = alpha != 0
        e.BgAlphas[bgIdx][x] = alpha
        continue
    }
}

func (ppu *PPU) affineScanline(e *Engine, bgIdx uint32) {

    bg := &e.Backgrounds[bgIdx]
    pa := float64(utils.Convert16ToFloat(uint16(bg.Pa), 8))
    pc := float64(utils.Convert16ToFloat(uint16(bg.Pc), 8))

    for x := range uint32(SCREEN_WIDTH) {

        xIdx := int(pa*float64(x) + bg.OutX)
        yIdx := int(pc*float64(x) + bg.OutY)

        if bg.Mosaic && e.Mosaic.BgH != 0 {
            xIdx -= xIdx % int(e.Mosaic.BgH+1)
        }

        if bg.Mosaic && e.Mosaic.BgV != 0 {
            yIdx -= yIdx % int(e.Mosaic.BgV+1)
        }

        out := xIdx < 0 || xIdx >= int(bg.W) || yIdx < 0 || yIdx >= int(bg.H)

        switch {
        case bg.AffineWrap:
            xIdx &= int(bg.W) - 1
            yIdx &= int(bg.H) - 1
        case !bg.AffineWrap && out:
			e.BgPalettes[bgIdx][x] = 0
			e.BgOks[bgIdx][x] = false
            continue
        }

        map_x := (uint32(xIdx)) & (bg.W - 1) >> 3
        map_y := ((uint32(yIdx)) & (bg.H - 1)) >> 3
        map_y *= bg.W >> 3
        mapIdx := map_y + map_x

        mapAddr := bg.ScreenBaseBlock + mapIdx

        if !e.IsB {
            mapAddr += e.Dispcnt.ScreenBase
        }

        vramOffset := uint32(0)

        if e.IsB {
            vramOffset = uint32(0x20_0000)
        }

        data := uint32(ppu.Vram.Read(vramOffset+mapAddr, true))

        tileAddr := bg.CharBaseBlock + (data << 6)

        inTileY := yIdx & 0b111 //% 8
        inTileX := xIdx & 0b111 //% 8

        if hFlip := data>>10&1 != 0; hFlip {
            inTileX = 7 - inTileX
        }

        if vFlip := data>>11&1 != 0; vFlip {
            inTileY = 7 - inTileY
        }

        inTileIdx := uint32(inTileX) + uint32(inTileY<<3)
        addr := vramOffset + tileAddr + inTileIdx
        palIdx := uint32(ppu.Vram.Read(addr, true))

        palIdx &= 0xFF

        if palIdx == 0 {
			e.BgPalettes[bgIdx][x] = 0
			e.BgOks[bgIdx][x] = false
            continue
        }

        if e.Dispcnt.BgExtPal {
            if e.Backgrounds[bgIdx].AltExtPalSlot {
                bgIdx += 2
            }

            e.BgPalettes[bgIdx][x] = uint32(b16(e.ExtBgSlots[bgIdx][palIdx<<1:]))
			e.BgOks[bgIdx][x] = true
            continue
        }

        e.BgPalettes[bgIdx][x] = uint32(e.Pram.Bg[palIdx])
        e.BgOks[bgIdx][x] = true
    }
}

func (ppu *PPU) affine16Scanline(e *Engine, bgIdx uint32) {

    bg := &e.Backgrounds[bgIdx]

    base := bg.ScreenBaseBlock
    var banks uint32
    if e.IsB {
        base += 0x20_0000
        banks = BANKS_B_2D_BG
    } else {
        banks = BANKS_A_2D_BG
    }

    pa := float64(utils.Convert16ToFloat(uint16(bg.Pa), 8))
    pc := float64(utils.Convert16ToFloat(uint16(bg.Pc), 8))

    for x := range uint32(SCREEN_WIDTH) {

        xIdx := int(pa*float64(x) + bg.OutX)
        yIdx := int(pc*float64(x) + bg.OutY)

        if bg.Mosaic && e.Mosaic.BgH != 0 {
            xIdx -= xIdx % int(e.Mosaic.BgH+1)
        }

        if bg.Mosaic && e.Mosaic.BgV != 0 {
            yIdx -= yIdx % int(e.Mosaic.BgV+1)
        }

        out := xIdx < 0 || xIdx >= int(bg.W) || yIdx < 0 || yIdx >= int(bg.H)

        switch {
        case bg.AffineWrap:
            xIdx &= int(bg.W) - 1
            yIdx &= int(bg.H) - 1
        case !bg.AffineWrap && out:
			e.BgPalettes[bgIdx][x] = 0
			e.BgOks[bgIdx][x] = false
            continue
        }

        const BYTE_SHIFT = 1

        map_x := (uint32(xIdx)) & (bg.W - 1) >> 3
        map_y := ((uint32(yIdx)) & (bg.H - 1)) >> 3
        map_y *= bg.W >> 3
        mapIdx := map_y + map_x
        mapIdx <<= BYTE_SHIFT

        mapAddr := base + mapIdx

        data := uint32(ppu.Vram.ReadGraphical(mapAddr, banks))

        tileIdx := (data & 0b11_1111_1111) << 5

        tileAddr := bg.CharBaseBlock + tileIdx
        tileAddr += tileIdx

        if !e.IsB {
            tileAddr += e.Dispcnt.CharBase
        } else {
            tileAddr += 0x20_0000
        }

        inTileY := yIdx & 0b111 //% 8
        inTileX := xIdx & 0b111 //% 8

        if hFlip := data>>10&1 != 0; hFlip {
            inTileX = 7 - inTileX
        }

        if vFlip := data>>11&1 != 0; vFlip {
            inTileY = 7 - inTileY
        }

        inTileIdx := uint32(inTileX) + uint32(inTileY<<3)
        palIdx := uint32(ppu.Vram.Read(tileAddr+inTileIdx, true))
        palNum := data >> 12

        palIdx &= 0xFF

        if palIdx == 0 {
			e.BgPalettes[bgIdx][x] = 0
			e.BgOks[bgIdx][x] = false
            continue
        }

        if e.Dispcnt.BgExtPal {
            //16 colors x 16 palettes --> standard palette memory (=256 colors)
            //256 colors x 16 palettes --> extended palette memory (=4096 colors)
            // this can probably be replaced in the ppu.
            // ex. vram.ExtABgSlot (unsafe) -> Background[bgIdx].BgSlot (*0x2000uint8)
            // removes un necessary palette slot switches

            if e.Backgrounds[bgIdx].AltExtPalSlot {
                bgIdx += 2
            }

            addr := (palNum << 9) + palIdx<<1

            e.BgPalettes[bgIdx][x] = uint32(b16(e.ExtBgSlots[bgIdx][addr:]))
			e.BgOks[bgIdx][x] = true
            continue
        }

        e.BgPalettes[bgIdx][x] = uint32(e.Pram.Bg[palIdx])
        e.BgOks[bgIdx][x] = true
    }
}

func (ppu *PPU) directBmpScanline(e *Engine, bgIdx, y uint32) {

    bg := &e.Backgrounds[bgIdx]

    base := bg.ScreenBaseBlock * 8
    var banks uint32
    if e.IsB {
        base += 0x20_0000
        banks = BANKS_B_2D_BG
    } else {
        banks = BANKS_A_2D_BG
    }

    ptr := ppu.Vram.ReadGraphicalPtr(base)

    pa := float64(utils.Convert16ToFloat(uint16(bg.Pa), 8))
    pc := float64(utils.Convert16ToFloat(uint16(bg.Pc), 8))

    for x := range SCREEN_WIDTH {

        xIdx := int(pa*float64(x) + bg.OutX)
        yIdx := int(pc*float64(x) + bg.OutY)

        if bg.Mosaic && e.Mosaic.BgH != 0 {
            xIdx -= xIdx % int(e.Mosaic.BgH+1)
        }

        if bg.Mosaic && e.Mosaic.BgV != 0 {
            yIdx -= yIdx % int(e.Mosaic.BgV+1)
        }

        out := xIdx < 0 || xIdx >= int(bg.W) || yIdx < 0 || yIdx >= int(bg.H)

        switch {
        case bg.AffineWrap:
            xIdx &= int(bg.W) - 1
            yIdx &= int(bg.H) - 1
        case !bg.AffineWrap && out:
			e.BgPalettes[bgIdx][x] = 0
			e.BgOks[bgIdx][x] = false
            continue
        }

        addr := uint32(xIdx+(yIdx*int(bg.W))) * 2

        var data uint32
        if ptr == nil {
            data = uint32(ppu.Vram.ReadGraphical(base + addr, banks))
        } else {
            data = uint32(*(*uint16)(unsafe.Add(ptr, addr)))
        }

        // required sonic dark brotherhood
        if transparent := (data & 0x8000) == 0; transparent {
			e.BgPalettes[bgIdx][x] = 0
			e.BgOks[bgIdx][x] = false
            continue
        }

        e.BgPalettes[bgIdx][x] = data
        e.BgOks[bgIdx][x] = true
    }
}

func (ppu *PPU) bmpScanline(e *Engine, bgIdx, y uint32) {

    bg := &e.Backgrounds[bgIdx]

    base := bg.ScreenBaseBlock * 8
    if e.IsB {
        base += 0x20_0000
    }

    ptr := ppu.Vram.ReadGraphicalPtr(base)

    pa := float64(utils.Convert16ToFloat(uint16(bg.Pa), 8))
    pc := float64(utils.Convert16ToFloat(uint16(bg.Pc), 8))

    for x := range SCREEN_WIDTH {

        xIdx := int(pa*float64(x) + bg.OutX)
        yIdx := int(pc*float64(x) + bg.OutY)

        if bg.Mosaic && e.Mosaic.BgH != 0 {
            xIdx -= xIdx % int(e.Mosaic.BgH+1)
        }

        if bg.Mosaic && e.Mosaic.BgV != 0 {
            yIdx -= yIdx % int(e.Mosaic.BgV+1)
        }

        out := xIdx < 0 || xIdx >= int(bg.W) || yIdx < 0 || yIdx >= int(bg.H)

        switch {
        case bg.AffineWrap:
            xIdx &= int(bg.W) - 1
            yIdx &= int(bg.H) - 1
        case !bg.AffineWrap && out:
			e.BgPalettes[bgIdx][x] = 0
			e.BgOks[bgIdx][x] = false
            continue
        }

        addr := uint32(xIdx + (yIdx * int(bg.W)))

        var palIdx uint32
        if ptr == nil {
            palIdx = uint32(ppu.Vram.Read(base + addr, true))
        } else {
            palIdx = uint32(*(*uint8)(unsafe.Add(ptr, addr)))
        }

        if palIdx == 0 {
			e.BgPalettes[bgIdx][x] = 0
			e.BgOks[bgIdx][x] = false
            continue
        }

        if e.IsB {
			e.BgPalettes[bgIdx][x] = uint32(ppu.EngineB.Pram.Bg[palIdx])
			e.BgOks[bgIdx][x] = true
            continue
        }

        e.BgPalettes[bgIdx][x] = uint32(ppu.EngineA.Pram.Bg[palIdx])
        e.BgOks[bgIdx][x] = true
    }
}

func (ppu *PPU) tiledScanline(e *Engine, bgIdx, y uint32) {

    bg := &e.Backgrounds[bgIdx]

	yIdx := (y + bg.YOffset) & ((bg.H) - 1)

	if bg.Mosaic && e.Mosaic.BgV != 0 {
		yIdx -= yIdx % (e.Mosaic.BgV + 1)
	}

	map_y := yIdx >> 3
	quad_x := uint32(10) //32 * 32
	quad_y := uint32(10) //32 * 32
	if bg.Size == 3 {
		quad_y = 11
	}

	var banks uint32
	if !e.IsB {
		banks = BANKS_A_2D_BG
	} else {
		banks = BANKS_B_2D_BG
	}

	for tileX := uint32(0); tileX < SCREEN_WIDTH; tileX += 8 {

		tileXIdx := (tileX + bg.XOffset) & ((bg.W) - 1)

		if bg.Mosaic && e.Mosaic.BgH != 0 {
			tileXIdx -= tileXIdx % (e.Mosaic.BgH + 1)
		}

		map_x := tileXIdx >> 3
		mapIdx := (map_y >> 5) << quad_y
		mapIdx += (map_x >> 5) << quad_x
		mapIdx += (map_y & 31) << 5
		mapIdx += (map_x & 31)
		mapIdx <<= 1

		mapAddr := bg.ScreenBaseBlock + mapIdx
		if !e.IsB {
			mapAddr += e.Dispcnt.ScreenBase
		} else {
			mapAddr += 0x20_0000
		}

		screenData := uint32(ppu.Vram.ReadGraphical(mapAddr, banks))
        palNum := screenData >> 12

		tileIdx := (screenData & 0b11_1111_1111) << 5

		tileAddr := bg.CharBaseBlock + tileIdx
		if bg.Palette256 {
			tileAddr += tileIdx
		}

		if !e.IsB {
			tileAddr += e.Dispcnt.CharBase
		} else {
			tileAddr += 0x20_0000
		}

		hFlip := (screenData>>10)&1 != 0

		inTileY := yIdx & 7
		if vFlip := (screenData>>11)&1 != 0; vFlip {
			inTileY = 7 - inTileY
		}

		ptr := ppu.Vram.ReadGraphicalPtr(tileAddr)

		for inTileX := range uint32(8) {

			xIdx := (tileXIdx + inTileX) & ((bg.W) - 1)

			if bg.Mosaic && e.Mosaic.BgH != 0 {
				xIdx -= xIdx % (e.Mosaic.BgH + 1)
			}

			inTileX := xIdx & 7
			if hFlip {
				inTileX = 7 - inTileX
			}

			var inTileIdx uint32
			if bg.Palette256 {
				inTileIdx = inTileX + (inTileY << 3)
			} else {
				inTileIdx = (inTileX >> 1) + (inTileY << 2)
			}

			var palIdx uint32
			if ptr == nil {
				palIdx = uint32(ppu.Vram.ReadGraphical(tileAddr+inTileIdx, banks))
			} else {
				palIdx = uint32(*(*uint16)(unsafe.Add(ptr, inTileIdx)))
			}

            if bg.Palette256 {
                palIdx &= 0xFF
            } else {
                palIdx = (palIdx >> ((inTileX & 1) << 2)) & 0xF
            }

            if palIdx == 0 {
                e.BgPalettes[bgIdx][tileX+inTileX] = 0
                e.BgOks[bgIdx][tileX+inTileX] = false
                continue
            }

            if e.Dispcnt.BgExtPal && bg.Palette256 {
                if e.Backgrounds[bgIdx].AltExtPalSlot {
                    bgIdx += 2
                }

                addr := (palNum << 9) + palIdx<<1

                e.BgPalettes[bgIdx][tileX+inTileX] = uint32(b16(e.ExtBgSlots[bgIdx][addr:]))
                e.BgOks[bgIdx][tileX+inTileX] = true
                continue
            }

            if bg.Palette256 {
                palNum = 0
            }

            addr := (palNum << 4) + palIdx
			e.BgPalettes[bgIdx][tileX+inTileX] = uint32(e.Pram.Bg[addr])
            e.BgOks[bgIdx][tileX+inTileX] = true
		}
	}
}
