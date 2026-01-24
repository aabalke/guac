package ppu

const (
	BLD_MODE_OFF   = 0
	BLD_MODE_STD   = 1
	BLD_MODE_WHITE = 2
	BLD_MODE_BLACK = 3
)

type BlendType uint16

const (
	BLEND_NONE BlendType = iota
	BLEND_ALPHA
	BLEND_ALPHA_3D
	BLEND_WHITE
	BLEND_BLACK
)

type BlendPalettes struct {
	NoBlendPalette, APalette, BPalette uint16
	hasA, hasB, targetATop             bool

	targetA3d bool
	alpha     uint16

    winBlend, objTransparent bool
}

func (bp *BlendPalettes) SetBgPalettes3d(palData uint16, alpha float32, bld *Blend) {

	bp.NoBlendPalette = palData

	if bld.a[0] {
		bp.APalette = palData
		bp.hasA = true
		bp.targetATop = true
        bp.targetA3d = true
        bp.alpha = min(16, uint16(alpha*16))
		return
	}

	bp.targetATop = false

	// not sure if this is required or correct
	bp.targetA3d = false

	if bld.b[0] {
		bp.BPalette = palData
		bp.hasB = true
	}
}

func (bp *BlendPalettes) SetBgPalettes(palData uint16, bgIdx uint32, bld *Blend) {

	bp.NoBlendPalette = palData

	if bld.a[bgIdx] {
		bp.APalette = palData
		bp.hasA = true
		bp.targetATop = true
		bp.targetA3d = false
		return
	}

	bp.targetATop = false

	// not sure if this is required or correct
	bp.targetA3d = false

	if bld.b[bgIdx] {
		bp.BPalette = palData
		bp.hasB = true
	}
}

func (bp *BlendPalettes) SetObjPalettes(palData uint16, semiTransparent bool, bld *Blend) {

	bp.NoBlendPalette = palData

	if bld.a[4] || semiTransparent {
		bp.APalette = palData
		bp.hasA = true
		bp.targetATop = true
		return
	}

	bp.targetATop = false
	if bld.b[4] {
		bp.BPalette = palData
		bp.hasB = true
	}

}

func ResetBlendPalettes(e *Engine) {

    bld := &e.Blend

    backdrop := *e.Backdrop

	bp := BlendPalettes{}

	bp.NoBlendPalette = backdrop

	if bld.a[5] {
		bp.APalette = backdrop
		bp.hasA = true
		bp.targetATop = true
	}

	if bld.b[5] {
		bp.BPalette = backdrop
		bp.hasB = true
	}

    for x := range uint32(SCREEN_WIDTH) {
        e.BlendPalettes[x] = bp
    }
}

func BlendAll(bld *Blend, bps *[SCREEN_WIDTH]BlendPalettes, blended *[SCREEN_WIDTH]uint16) {

    for x := range uint32(SCREEN_WIDTH) {
		blended[x] = bps[x].Blend(bld)
    }
}

func (bp *BlendPalettes) Blend(bld *Blend) uint16 {

	if bp.winBlend {
		return bp.noBlend(bld)
	}

	switch bld.Mode {
	case BLD_MODE_OFF:
		return bp.noBlend(bld)
	case BLD_MODE_STD:
		return bp.alphaBlend(bld)
	case BLD_MODE_WHITE:
		return bp.grayscaleBlend(true, bld)
	case BLD_MODE_BLACK:
		return bp.grayscaleBlend(false, bld)
	default:
		return bp.noBlend(bld)
	}
}

func (bp *BlendPalettes) noBlend(bld *Blend) uint16 {
	if bp.objTransparent {
		return bp.alphaBlend(bld)
	}

	return bp.NoBlendPalette
}

func (bp *BlendPalettes) alphaBlend(bld *Blend) uint16 {

	if !bp.hasA || !bp.hasB || !bp.targetATop || bp.alpha >= 1 {
		return bp.NoBlendPalette
	}

	rA := (bp.APalette) & 0x1F
	gA := (bp.APalette >> 5) & 0x1F
	bA := (bp.APalette >> 10) & 0x1F
	rB := (bp.BPalette) & 0x1F
	gB := (bp.BPalette >> 5) & 0x1F
	bB := (bp.BPalette >> 10) & 0x1F

	blend := func(a, b uint16) uint16 {

		if bp.targetA3d {
			panic("untested 3d target blend sisd")
			return max(0, min(31, (a*bp.alpha+b*(1-bp.alpha))>>4))
		}

		return min(31, (a*bld.aEv+b*bld.bEv)>>4)
	}
	r := blend(rA, rB)
	g := blend(gA, gB)
	b := blend(bA, bB)

	return r | (g << 5) | (b << 10)
}

func (bp *BlendPalettes) grayscaleBlend(white bool, bld *Blend) uint16 {

	if !bp.hasA || !bp.targetATop {
		return bp.NoBlendPalette
	}

	rA := (bp.APalette) & 0x1F
	gA := (bp.APalette >> 5) & 0x1F
	bA := (bp.APalette >> 10) & 0x1F

	blend := func(v uint16) uint16 {

		if white {
			v += ((31 - v) * bld.yEv) >> 4
		} else {
			v -= (v * bld.yEv) >> 4
		}

		return uint16(min(31, v))
	}

	r := blend(rA)
	g := blend(gA)
	b := blend(bA)

	return r | (g << 5) | (b << 10)
}
