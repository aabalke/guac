package ppu

import (
	"encoding/binary"
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
			if top < 0 || top-int(bg.H) >= 0 {
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

func (ppu *PPU) threeScanline(e *Engine, bgIdx, priority, y uint32) {

    wins := &e.Windows
	bg := &e.Backgrounds[bgIdx]

	yIdx := (y + bg.YOffset) & ((bg.H) - 1)

	if yIdx >= SCREEN_HEIGHT {
		return
	}

	for x := range uint32(SCREEN_WIDTH) {

        if covered := e.BgOks[x]; covered {
            continue
        }

        if wins.Enabled && !wins.inWinBg(bgIdx, x, y) {
            continue
        }

		xIdx := (x + bg.XOffset) & ((bg.W) - 1)
		i := (xIdx + (yIdx * SCREEN_WIDTH))

		if xIdx >= SCREEN_WIDTH {
			continue
		}

		r := ppu.Rasterizer.Render

		pal, alpha := uint16(0), float32(0)

		if r.Rasterizer.Buffers.BisRendering {
			pal, alpha = uint16(r.Pixels.PalettesA[i]), r.Pixels.AlphaA[i]
		} else {
			pal, alpha = uint16(r.Pixels.PalettesB[i]), r.Pixels.AlphaB[i]
		}

		if noblend := ppu.EngineA.Blend.Mode == 0; noblend {

			// this is only important if the 3d screen is the only one, nothing is behind it, and alpha != 1.
			// really only noticed in devkit tests ( nehe/lesson10)

			r := float32(pal&0x1F) * alpha
			g := float32((pal>>5)&0x1F) * alpha
			b := float32((pal>>10)&0x1F) * alpha

			pal |= uint16(r) & 0x1F
			pal |= ((uint16(g) & 0x1F) << 5)
			pal |= ((uint16(b) & 0x1F) << 10)
		}

		e.BgPals[x] = pal
		e.BgOks[x] = alpha != 0
		e.BgAlphas[x] = alpha
        e.BgIdx[x] = bgIdx
		continue
	}
}

func (ppu *PPU) affineScanline(e *Engine, bgIdx, priority, y uint32) {

    wins := &e.Windows
	bg := &e.Backgrounds[bgIdx]
	pa := float64(utils.Convert16ToFloat(uint16(bg.Pa), 8))
	pc := float64(utils.Convert16ToFloat(uint16(bg.Pc), 8))

    base := bg.ScreenBaseBlock
    tileBase := bg.CharBaseBlock
    if e.IsB {
        base += 0x20_0000
        tileBase += 0x20_0000
    } else {
        base += e.Dispcnt.ScreenBase
    }

	for x := range uint32(SCREEN_WIDTH) {
        if covered := e.BgOks[x]; covered {
            continue
        }


        if wins.Enabled && !wins.inWinBg(bgIdx, x, y) {
            continue
        }

		xIdx := int(pa*float64(x) + bg.OutX)
		if bg.Mosaic && e.Mosaic.BgH != 0 {
			xIdx -= xIdx % int(e.Mosaic.BgH+1)
		}

		yIdx := int(pc*float64(x) + bg.OutY)
		if bg.Mosaic && e.Mosaic.BgV != 0 {
			yIdx -= yIdx % int(e.Mosaic.BgV+1)
		}

		out := xIdx < 0 || xIdx >= int(bg.W) || yIdx < 0 || yIdx >= int(bg.H)
		switch {
		case bg.AffineWrap:
			xIdx &= int(bg.W) - 1
			yIdx &= int(bg.H) - 1
		case out:
			continue
		}

        var (
            map_x  = (uint32(xIdx))  & (bg.W - 1) >> 3
            map_y  = (((uint32(yIdx)) & (bg.H - 1)) >> 3) * (bg.W >> 3)
            mapIdx = map_y + map_x
            data   = uint32(ppu.Vram.Read(base+mapIdx, true))
        )

		inTileX := xIdx & 7
		if hFlip := (data>>10)&1 != 0; hFlip {
			inTileX = 7 - inTileX
		}

		inTileY := yIdx & 7
		if vFlip := (data>>11)&1 != 0; vFlip {
			inTileY = 7 - inTileY
		}

        var (
            inTileIdx = uint32(inTileX) + uint32(inTileY<<3)
            addr      = tileBase + (data << 6) + inTileIdx
            palIdx    = ppu.Vram.Read(addr, true)
        )

		if palIdx == 0 {
			continue
		}

        e.BgOks[x] = true
        e.BgIdx[x] = bgIdx

		if e.Dispcnt.BgExtPal {
			if e.Backgrounds[bgIdx].AltExtPalSlot {
				bgIdx += 2
			}

			e.BgPals[x] = binary.LittleEndian.Uint16(e.ExtBgSlots[bgIdx][palIdx<<1:])
			continue
		}

		e.BgPals[x] = e.Pram.Bg[palIdx]
	}
}

func (ppu *PPU) affine16Scanline(e *Engine, bgIdx, priority, y uint32) {

    wins := &e.Windows
	bg := &e.Backgrounds[bgIdx]

	base := bg.ScreenBaseBlock
	if e.IsB {
		base += 0x20_0000
	}

	pa := float64(utils.Convert16ToFloat(uint16(bg.Pa), 8))
	pc := float64(utils.Convert16ToFloat(uint16(bg.Pc), 8))

	for x := range uint32(SCREEN_WIDTH) {
        if covered := e.BgOks[x]; covered {
            continue
        }

        if wins.Enabled && !wins.inWinBg(bgIdx, x, y) {
            continue
        }

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
			continue
		}

		const BYTE_SHIFT = 1

		map_x := (uint32(xIdx)) & (bg.W - 1) >> 3
		map_y := ((uint32(yIdx)) & (bg.H - 1)) >> 3
		map_y *= bg.W >> 3
		mapIdx := map_y + map_x
		mapIdx <<= BYTE_SHIFT

		mapAddr := base + mapIdx

		data := uint32(ppu.Vram.ReadGraphical(mapAddr))

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
			continue
		}

        e.BgOks[x] = true
        e.BgIdx[x] = bgIdx

		if e.Dispcnt.BgExtPal {
			if e.Backgrounds[bgIdx].AltExtPalSlot {
				bgIdx += 2
			}

			addr := (palNum << 9) + palIdx<<1

			e.BgPals[x] = binary.LittleEndian.Uint16(e.ExtBgSlots[bgIdx][addr:])
			continue
		}

		e.BgPals[x] = e.Pram.Bg[palIdx]
	}
}

func (ppu *PPU) directBmpScanline(e *Engine, bgIdx, priority, y uint32) {

    wins := &e.Windows
	bg := &e.Backgrounds[bgIdx]

	base := bg.ScreenBaseBlock * 8
	if e.IsB {
		base += 0x20_0000
	}

	ptr := ppu.Vram.ReadGraphicalPtr(base)

	pa := float64(utils.Convert16ToFloat(uint16(bg.Pa), 8))
	pc := float64(utils.Convert16ToFloat(uint16(bg.Pc), 8))

	for x := range uint32(SCREEN_WIDTH) {
        if covered := e.BgOks[x]; covered {
            continue
        }

        if wins.Enabled && !wins.inWinBg(bgIdx, x, y) {
            continue
        }

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
			continue
		}

		addr := uint32(xIdx+(yIdx*int(bg.W))) * 2

		var data uint16
		if ptr == nil {
			data = ppu.Vram.ReadGraphical(base + addr)
		} else {
			data = *(*uint16)(unsafe.Add(ptr, addr))
		}

		// required sonic dark brotherhood
		if transparent := (data & 0x8000) == 0; transparent {
			continue
		}

		e.BgPals[x] = data &^ 0x8000
		e.BgOks[x] = true
        e.BgIdx[x] = bgIdx
	}
}

func (ppu *PPU) bmpScanline(e *Engine, bgIdx, priority, y uint32) {

    wins := &e.Windows
	bg := &e.Backgrounds[bgIdx]

	base := bg.ScreenBaseBlock * 8
	if e.IsB {
		base += 0x20_0000
	}

	ptr := ppu.Vram.ReadGraphicalPtr(base)

	pa := float64(utils.Convert16ToFloat(uint16(bg.Pa), 8))
	pc := float64(utils.Convert16ToFloat(uint16(bg.Pc), 8))

	for x := range uint32(SCREEN_WIDTH) {
        if covered := e.BgOks[x]; covered {
            continue
        }

        if wins.Enabled && !wins.inWinBg(bgIdx, x, y) {
            continue
        }

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
			continue
		}

		addr := uint32(xIdx + (yIdx * int(bg.W)))

		var palIdx uint32
		if ptr == nil {
			palIdx = uint32(ppu.Vram.Read(base+addr, true))
		} else {
			palIdx = uint32(*(*uint8)(unsafe.Add(ptr, addr)))
		}

		if palIdx == 0 {
			continue
		}

        e.BgOks[x] = true
        e.BgIdx[x] = bgIdx

		if e.IsB {
			e.BgPals[x] = ppu.EngineB.Pram.Bg[palIdx]
			continue
		}

		e.BgPals[x] = ppu.EngineA.Pram.Bg[palIdx]
	}
}

func (ppu *PPU) tiledScanline(e *Engine, bgIdx, priority, y uint32) {

	const (
		TILE_SIZE     = 8
		TILE_MASK     = TILE_SIZE - 1
		MAP_COL_SH    = 10
		TILE_ROW_MASK = 0xF8
	)

    wins := &e.Windows
	bg := &e.Backgrounds[bgIdx]

	bgY := (y + bg.YOffset) & ((bg.H) - 1)
	if bg.Mosaic && e.Mosaic.BgV != 0 {
		bgY -= bgY % (e.Mosaic.BgV + 1)
	}

	mapRowShift := uint32(10) // 32x32
	if bg.Size == 3 {
		mapRowShift = 11 // 64x64
	}

	var (
		mapRowOffset = ((bgY>>8)<<mapRowShift) + ((bgY&TILE_ROW_MASK)<<2)
		tileBase     = bg.CharBaseBlock
		mapBase      = bg.ScreenBaseBlock
	)

	if !e.IsB {
		tileBase += e.Dispcnt.CharBase
		mapBase += e.Dispcnt.ScreenBase
	} else {
		tileBase += 0x20_0000
		mapBase += 0x20_0000
	}

	var (
		scrollX     = bg.XOffset & (bg.W - 1)
		startTileX  = scrollX >> 3
		startPixelX = scrollX & 7
		screenX     = uint32(0)
	)

	for tile := uint32(0); screenX < SCREEN_WIDTH; tile++ {

		var (
			tileX        = (startTileX + tile) & ((bg.W >> 3) - 1)
			mapColOffset = (tileX >> 5 << MAP_COL_SH) + (tileX & 31)
			mapAddr      = mapBase + ((mapRowOffset + mapColOffset) << 1)
			screenData   = ppu.Vram.ReadGraphical(mapAddr)
			palNum       = screenData >> 12
			tileNum      = uint32(screenData & 0x03FF)
			hFlip        = (screenData>>10)&1 != 0
			vFlip        = (screenData>>11)&1 != 0
			tileOffset   = tileNum << 5
		)

		if bg.Palette256 {
			tileOffset <<= 1
		}

		var (
			tileAddr = tileBase + tileOffset
			ptr      = ppu.Vram.ReadGraphicalPtr(tileAddr)
		)

		tilePxY := bgY & 7
		if vFlip {
			tilePxY = 7 - tilePxY
		}

		pxStart := uint32(0)
		if tile == 0 {
			pxStart = startPixelX
		}

		for px := pxStart; px < 8 && screenX < SCREEN_WIDTH; px++ {

            if e.BgOks[screenX] {
				screenX++
                continue
            }

            if wins.Enabled && !wins.inWinBg(bgIdx, screenX, y) {
				screenX++
                continue
            }

			tilePxX := px
			if hFlip {
				tilePxX = 7 - tilePxX
			}

			var inTileOffset uint32
			if bg.Palette256 {
				inTileOffset = tilePxX + (tilePxY << 3)
			} else {
				inTileOffset = (tilePxX >> 1) + (tilePxY << 2)
			}

			var palIdx uint16
			if ptr == nil {
				palIdx = ppu.Vram.ReadGraphical(tileAddr + inTileOffset)
			} else {
				palIdx = *(*uint16)(unsafe.Add(ptr, inTileOffset))
			}

			if bg.Palette256 {
				palIdx &= 0xFF
			} else {
				palIdx = (palIdx >> ((tilePxX & 1) << 2)) & 0xF
			}

			if palIdx == 0 {
				screenX++
				continue
			}

			e.BgOks[screenX] = true
			e.BgIdx[screenX] = bgIdx

			if e.Dispcnt.BgExtPal && bg.Palette256 {
				slot := bgIdx
				if bg.AltExtPalSlot {
					slot += 2
				}
				addr := (palNum << 9) + (palIdx << 1)
                e.BgPals[screenX] = *(*uint16)(unsafe.Pointer(&e.ExtBgSlots[slot][addr]))
                screenX++
                continue
            }

            if bg.Palette256 {
                e.BgPals[screenX] = e.Pram.Bg[palIdx]
                screenX++
                continue
            }

            addr := (palNum << 4) + palIdx
            e.BgPals[screenX] = e.Pram.Bg[addr]
            screenX++
        }
	}
}
