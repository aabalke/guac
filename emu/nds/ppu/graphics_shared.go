package ppu

func (ppu *PPU) render(x, y uint32, e *Engine, bldPal *BlendPalettes) (bool, bool) {

	var (
		dispcnt       = &e.Dispcnt
		wins          = &e.Windows
		bld           = &e.Blend
		objPriorities = &e.ObjPriorities
		bgPriorities  = &e.BgPriorities

		isSemiTransparent bool
		inObjWindow       bool
	)

	for i := 3; i >= 0; i-- {

		for j := bgPriorities[i].Cnt - 1; j >= 0; j-- {

			bgIdx := bgPriorities[i].Idx[j]

            if !e.BgOks[bgIdx][x] {
                continue
            }

            if wins.Enabled && !wins.inWinBg(bgIdx, x, y) {
                continue
            }

            if e.Backgrounds[bgIdx].Type == BG_TYPE_3D {
                bldPal.SetBgPalettes3d(e.BgPalettes[bgIdx][x], e.BgAlphas[bgIdx][x], bld)
            } else {
                bldPal.SetBgPalettes(e.BgPalettes[bgIdx][x], bgIdx, bld)
            }
        }

		if objDisabled := !dispcnt.DisplayObj; objDisabled {
			continue
		}

		if wins.Enabled && !wins.inWinObj(x, y) {
			continue
		}

        ObjectLoop:
        for j := range objPriorities[i].Cnt {

            objIdx := (*objPriorities)[i].Idx[j]
			obj := &e.Objects[objIdx]

			palData, ok := uint32(0), false

            switch {
            case obj.Mode == 3:
				obj.OneDimensional = dispcnt.BitmapObj1D
				palData, ok = ppu.setObjBmpPixel(e, obj, x, y)
            default:
				obj.OneDimensional = dispcnt.TileObj1D
				palData, ok = ppu.setObjTilePixel(e, obj, x, y)
			}

			if !ok {
				continue
			}

			if obj.Mode == 2 {
				inObjWindow = true
				break ObjectLoop
			}

			isSemiTransparent = obj.Mode == 1
			bldPal.SetObjPalettes(palData, isSemiTransparent, bld)
			break ObjectLoop
		}
	}

	return isSemiTransparent, inObjWindow
}
