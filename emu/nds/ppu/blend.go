package ppu

const (
	BLD_NONE  = 0
	BLD_ALPHA = 1
	BLD_WHITE = 2
	BLD_BLACK = 3
    BLD_ALPHA_3D = 4
)

// blends are [6]... because Bg0, Bg1, Bg2, Bg3, Obj, Bd
type Blend struct {
	Mode          uint32
	a, b          [6]bool
	aEv, bEv, yEv uint16

	Blended [SCREEN_WIDTH]uint16

	NoBlendPals [SCREEN_WIDTH]uint16
	APals       [SCREEN_WIDTH]uint16
	BPals       [SCREEN_WIDTH]uint16
	alphas      [SCREEN_WIDTH]uint16

	hasA           [SCREEN_WIDTH]bool
	hasB           [SCREEN_WIDTH]bool
	targetATop     [SCREEN_WIDTH]bool
	targetA3d      [SCREEN_WIDTH]bool
	objTransparent [SCREEN_WIDTH]bool

	outWindow [SCREEN_WIDTH]bool
    modes [SCREEN_WIDTH]uint32
}

// occurs per priority
func (e *Engine) SetBgPals(priority uint32) {

	bld := &e.Blend

	for x := range uint32(SCREEN_WIDTH) {

		if !e.BgOks[x] {
			continue
		}

		pal := e.BgPals[x]
		bgIdx := e.BgIdx[x]

		bld.NoBlendPals[x] = pal
		bld.targetATop[x] = bld.a[bgIdx]
		bld.targetA3d[x] = e.Dispcnt.Is3D && bgIdx == 0

		if bld.a[bgIdx] {
			bld.APals[x] = pal
			bld.hasA[x] = true
			if bld.targetA3d[x] {
				bld.alphas[x] = min(16, uint16(e.BgAlphas[x]*16))
			}
			continue
		}

		if bld.b[bgIdx] {
			bld.BPals[x] = pal
			bld.hasB[x] = true
		}
	}
}

// occurs per priority
func (e *Engine) SetObjPals(priority uint32) {

	bld := &e.Blend

	for x := range uint32(SCREEN_WIDTH) {

		if !e.ObjOk[x] {
			continue
		}

		pal := e.ObjPals[x]
		mode := e.ObjMode[x]

		bld.NoBlendPals[x] = pal
		bld.objTransparent[x] = mode == 1
		bld.targetATop[x] = bld.a[4] || bld.objTransparent[x]

		if bld.a[4] || bld.objTransparent[x] {
			bld.APals[x] = pal
			bld.hasA[x] = true
			continue
		}

		if bld.b[4] {
			bld.BPals[x] = pal
			bld.hasB[x] = true
		}
	}
}

// occurs per scanline
func ResetBlendPalettes(e *Engine) {

	bld := &e.Blend

	backdrop := *e.Backdrop &^ 0x8000

	bld.APals = [SCREEN_WIDTH]uint16{}
	bld.BPals = [SCREEN_WIDTH]uint16{}
	bld.alphas = [SCREEN_WIDTH]uint16{}
	bld.targetA3d = [SCREEN_WIDTH]bool{}
	bld.objTransparent = [SCREEN_WIDTH]bool{}
	bld.outWindow = [SCREEN_WIDTH]bool{}

	for x := range uint32(SCREEN_WIDTH) {
		bld.NoBlendPals[x] = backdrop
		bld.hasA[x] = bld.a[5]
		bld.targetATop[x] = bld.a[5]
		bld.hasB[x] = bld.b[5]
	}

	if bld.a[5] {
		copy(bld.APals[:], bld.NoBlendPals[:])
	}

	if bld.b[5] {
		copy(bld.BPals[:], bld.NoBlendPals[:])
	}
}

// occurs per scanline
func BlendAll(bld *Blend, wins *Windows, y uint32) {

	if wins.Enabled {
		for x := range uint32(SCREEN_WIDTH) {
			bld.outWindow[x] = !wins.inWinBld(x, y)
		}
	}

    for x := range uint32(SCREEN_WIDTH) {

		bld.modes[x] = bld.Mode
		if bld.outWindow[x] {
			bld.modes[x] = BLD_ALPHA
		}

		activeA :=
			bld.hasA[x] &&
			bld.targetATop[x] &&
			bld.alphas[x] < 1

		allowWin :=
			!bld.outWindow[x] ||
			(bld.hasB[x] && bld.objTransparent[x])

		requireB :=
			bld.modes[x] == BLD_ALPHA ||
			(bld.modes[x] == BLD_NONE && bld.objTransparent[x])

		if noBlending :=
			!activeA ||
			!allowWin ||
			(requireB && !bld.hasB[x]); noBlending {

            bld.modes[x] = BLD_NONE
        }

        if bld.modes[x] == BLD_ALPHA && bld.targetA3d[x] {
            bld.modes[x] = BLD_ALPHA_3D
        }
    }

	for x := range uint32(SCREEN_WIDTH) {
		switch bld.modes[x] {
        case BLD_NONE:
			bld.Blended[x] = bld.NoBlendPals[x]
		case BLD_ALPHA:

            r := (bld.APals[x]) & 0x1F
            g := (bld.APals[x] >> 5) & 0x1F
            b := (bld.APals[x] >> 10) & 0x1F
            rB := (bld.BPals[x]) & 0x1F
            gB := (bld.BPals[x] >> 5) & 0x1F
            bB := (bld.BPals[x] >> 10) & 0x1F

            r = min(31, (r*bld.aEv+rB*bld.bEv)>>4)
            g = min(31, (g*bld.aEv+gB*bld.bEv)>>4)
            b = min(31, (b*bld.aEv+bB*bld.bEv)>>4)

            bld.Blended[x] = r | (g << 5) | (b << 10)

		case BLD_WHITE:

            r := (bld.APals[x]) & 0x1F
            g := (bld.APals[x] >> 5) & 0x1F
            b := (bld.APals[x] >> 10) & 0x1F

            r = min(31, r + ((31 - r) * bld.yEv) >> 4)
            g = min(31, g + ((31 - g) * bld.yEv) >> 4)
            b = min(31, b + ((31 - b) * bld.yEv) >> 4)

            bld.Blended[x] = r | (g << 5) | (b << 10)

		case BLD_BLACK:

            r := (bld.APals[x]) & 0x1F
            g := (bld.APals[x] >> 5) & 0x1F
            b := (bld.APals[x] >> 10) & 0x1F

            r = min(31, r - (r * bld.yEv) >> 4)
            g = min(31, g - (g * bld.yEv) >> 4)
            b = min(31, b - (b * bld.yEv) >> 4)

            bld.Blended[x] = r | (g << 5) | (b << 10)

        case BLD_ALPHA_3D:

            panic("untested 3d target blend sisd")
            //return max(0, min(31, (a*bld.alpha+b*(1-bp.alpha))>>4))

            r := (bld.APals[x]) & 0x1F
            g := (bld.APals[x] >> 5) & 0x1F
            b := (bld.APals[x] >> 10) & 0x1F
            rB := (bld.BPals[x]) & 0x1F
            gB := (bld.BPals[x] >> 5) & 0x1F
            bB := (bld.BPals[x] >> 10) & 0x1F

            r = min(31, (r*bld.aEv+rB*bld.bEv)>>4)
            g = min(31, (g*bld.aEv+gB*bld.bEv)>>4)
            b = min(31, (b*bld.aEv+bB*bld.bEv)>>4)

            bld.Blended[x] = r | (g << 5) | (b << 10)
		}
	}
}
