package ppu

const (
	BLD_NONE     = 0
	BLD_ALPHA    = 1
	BLD_WHITE    = 2
	BLD_BLACK    = 3
	BLD_ALPHA_3D = 4
)

// blends are [6]... because Bg0, Bg1, Bg2, Bg3, Obj, Bd
type Blend struct {
	Mode          uint32
	a, b          [6]bool
	aEv, bEv, yEv uint16

	Blended        [SCREEN_WIDTH]uint16
	NoBlendPals    [SCREEN_WIDTH]uint16
	APals          [SCREEN_WIDTH]uint16
	BPals          [SCREEN_WIDTH]uint16
	alphas         [SCREEN_WIDTH]uint16
	hasA           [SCREEN_WIDTH]bool
	hasB           [SCREEN_WIDTH]bool
	targetATop     [SCREEN_WIDTH]bool
	targetA3d      [SCREEN_WIDTH]bool
	objTransparent [SCREEN_WIDTH]bool
	outWindow      [SCREEN_WIDTH]bool
	modes          [SCREEN_WIDTH]uint32

	// used to only run mode blend only if a pixel in scanline has mode
	// BLD_NONE == 0, so bit 0 is a flag for bld none
	modeFlags uint8

	whiteLut [17][32]uint16
	blackLut [17][32]uint16
	alphaLut [17][32]uint16
}

func NewBlend() *Blend {

	b := &Blend{}

	// set luts
	for y := range uint16(17) {

		scale := 16 - y
		whiteBias := (31 * y) >> 4

		for c := range uint16(32) {
			b.blackLut[y][c] = min(31, (c*scale)>>4)
			b.whiteLut[y][c] = min(31, ((c*scale)>>4)+whiteBias)
			b.alphaLut[y][c] = (c * y) >> 4
		}
	}

	return b
}

// occurs per priority
func (e *Engine) SetBgPals() {

	bld := e.Blend

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
				bld.alphas[x] = uint16(e.BgAlphas[x])
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

	bld := e.Blend

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

	bld := e.Blend

	backdrop := *e.Backdrop &^ 0x8000

	bld.modeFlags = 0

	//bld.APals = [SCREEN_WIDTH]uint16{}
	//bld.BPals = [SCREEN_WIDTH]uint16{}
	bld.alphas = [SCREEN_WIDTH]uint16{}
	bld.targetA3d = [SCREEN_WIDTH]bool{}
	bld.objTransparent = [SCREEN_WIDTH]bool{}

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

		if bld.outWindow[x] && wins.Enabled {
			bld.modes[x] = BLD_ALPHA
		}

		activeA := bld.hasA[x] && bld.targetATop[x] && bld.alphas[x] < 32
		allowWin := !bld.outWindow[x] || (bld.hasB[x] && bld.objTransparent[x])
		requireB := bld.modes[x] == BLD_ALPHA || (bld.modes[x] == BLD_NONE && bld.objTransparent[x])
		noBlending := !activeA || (wins.Enabled && !allowWin) || (requireB && !bld.hasB[x])

		if noBlending {
			bld.modes[x] = BLD_NONE
		}

		if bld.modes[x] == BLD_ALPHA && bld.targetA3d[x] {
			bld.modes[x] = BLD_ALPHA_3D
		}

		bld.modeFlags |= 1 << bld.modes[x]
	}

	// checks if all pixels are the same blend mode
	switch bld.modeFlags {
	case 1 << BLD_NONE:

		copy(bld.Blended[:], bld.NoBlendPals[:])

		for x := range uint32(SCREEN_WIDTH) {
			bld.Blended[x] &^= 0x8000
		}

		return

	case 1 << BLD_ALPHA:

		aLut := bld.alphaLut[bld.aEv]
		bLut := bld.alphaLut[bld.bEv]

		for x := range uint32(SCREEN_WIDTH) {
			var (
				pA = bld.APals[x]
				pB = bld.BPals[x]

				r = min(31, aLut[(pA>>0)&0x1F]+bLut[(pB>>0)&0x1F])
				g = min(31, aLut[(pA>>5)&0x1F]+bLut[(pB>>5)&0x1F])
				b = min(31, aLut[(pA>>10)&0x1F]+bLut[(pB>>10)&0x1F])
			)


			bld.Blended[x] = r | (g << 5) | (b << 10)
		}
		return

	case 1 << BLD_WHITE:

		lut := &bld.whiteLut[bld.yEv]

		for x := range uint32(SCREEN_WIDTH) {
			p := bld.APals[x]
			r := lut[p&0x1F]
			g := lut[(p>>5)&0x1F]
			b := lut[(p>>10)&0x1F]
			bld.Blended[x] = r | (g << 5) | (b << 10)
		}

		return
	case 1 << BLD_BLACK:

		lut := &bld.blackLut[bld.yEv]

		for x := range uint32(SCREEN_WIDTH) {
			p := bld.APals[x]
			r := lut[p&0x1F]
			g := lut[(p>>5)&0x1F]
			b := lut[(p>>10)&0x1F]
			bld.Blended[x] = r | (g << 5) | (b << 10)
		}

		return
	case 1 << BLD_ALPHA_3D:


		for x := range uint32(SCREEN_WIDTH) {
			var (
				pA = bld.APals[x]
				pB = bld.BPals[x]
				rA = (pA >> 0) & 0x1F
				gA = (pA >> 5) & 0x1F
				bA = (pA >> 10) & 0x1F
				rB = (pB >> 0) & 0x1F
				gB = (pB >> 5) & 0x1F
				bB = (pB >> 10) & 0x1F

				a  = bld.alphas[x]
				ai = 31 - a

				r = min(31, ((rA*a)+(rB*ai))>>5)
				g = min(31, ((gA*a)+(gB*ai))>>5)
				b = min(31, ((bA*a)+(bB*ai))>>5)
			)

			bld.Blended[x] = r | (g << 5) | (b << 10)
		}
		return
	}

	for x := range uint32(SCREEN_WIDTH) {
		switch bld.modes[x] {
		case BLD_NONE:
			bld.Blended[x] = bld.NoBlendPals[x] &^ 0x8000
		case BLD_ALPHA:

			var (
				aLut = bld.alphaLut[bld.aEv]
				bLut = bld.alphaLut[bld.bEv]
				pA   = bld.APals[x]
				pB   = bld.BPals[x]

				r = min(31, aLut[(pA>>0)&0x1F]+bLut[(pB>>0)&0x1F])
				g = min(31, aLut[(pA>>5)&0x1F]+bLut[(pB>>5)&0x1F])
				b = min(31, aLut[(pA>>10)&0x1F]+bLut[(pB>>10)&0x1F])
			)

			bld.Blended[x] = r | (g << 5) | (b << 10)

		case BLD_WHITE:

			lut := &bld.whiteLut[bld.yEv]
			p := bld.APals[x]
			r := lut[p&0x1F]
			g := lut[(p>>5)&0x1F]
			b := lut[(p>>10)&0x1F]
			bld.Blended[x] = r | (g << 5) | (b << 10)

		case BLD_BLACK:

			lut := &bld.blackLut[bld.yEv]
			p := bld.APals[x]
			r := lut[p&0x1F]
			g := lut[(p>>5)&0x1F]
			b := lut[(p>>10)&0x1F]
			bld.Blended[x] = r | (g << 5) | (b << 10)

		case BLD_ALPHA_3D:


			var (
				pA = bld.APals[x]
				pB = bld.BPals[x]
				rA = (pA >> 0) & 0x1F
				gA = (pA >> 5) & 0x1F
				bA = (pA >> 10) & 0x1F
				rB = (pB >> 0) & 0x1F
				gB = (pB >> 5) & 0x1F
				bB = (pB >> 10) & 0x1F

				a  = bld.alphas[x]
				ai = 31 - a

				r = min(31, ((rA*a)+(rB*ai))>>5)
				g = min(31, ((gA*a)+(gB*ai))>>5)
				b = min(31, ((bA*a)+(bB*ai))>>5)
			)

			bld.Blended[x] = r | (g << 5) | (b << 10)
		}
	}
}
