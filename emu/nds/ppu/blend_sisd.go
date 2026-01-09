//go:build !rc

package ppu

func NewBlendPalette(bld *Blend, backdrop uint32) *BlendPalettes {

	bp := &BlendPalettes{}

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

	return bp
}

func (bp *BlendPalettes) Blend(winBlend, objTransparent bool, bld *Blend) uint32 {

	if winBlend {
		return bp.noBlend(objTransparent, bld)
	}

	switch bld.Mode {
	case BLD_MODE_OFF:
		return bp.noBlend(objTransparent, bld)
	case BLD_MODE_STD:
		return bp.alphaBlend(bld)
	case BLD_MODE_WHITE:
		return bp.grayscaleBlend(true, bld)
	case BLD_MODE_BLACK:
		return bp.grayscaleBlend(false, bld)
	default:
		return bp.noBlend(objTransparent, bld)
	}
}

func (bp *BlendPalettes) noBlend(objTransparent bool, bld *Blend) uint32 {
	if objTransparent {
		return bp.alphaBlend(bld)
	}

	return bp.NoBlendPalette
}

func (bp *BlendPalettes) alphaBlend(bld *Blend) uint32 {

	if !bp.hasA || !bp.hasB || !bp.targetATop || bp.alpha >= 1 {
		return bp.NoBlendPalette
	}

	rA := (bp.APalette) & 0x1F
	gA := (bp.APalette >> 5) & 0x1F
	bA := (bp.APalette >> 10) & 0x1F
	rB := (bp.BPalette) & 0x1F
	gB := (bp.BPalette >> 5) & 0x1F
	bB := (bp.BPalette >> 10) & 0x1F

	blend := func(a, b uint32) uint32 {

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

func (bp *BlendPalettes) grayscaleBlend(white bool, bld *Blend) uint32 {

	if !bp.hasA || !bp.targetATop {
		return bp.NoBlendPalette
	}

	rA := (bp.APalette) & 0x1F
	gA := (bp.APalette >> 5) & 0x1F
	bA := (bp.APalette >> 10) & 0x1F

	blend := func(v uint32) uint32 {

		if white {
			v += ((31 - v) * bld.yEv) >> 4
		} else {
			v -= (v * bld.yEv) >> 4
		}

		return uint32(min(31, v))
	}

	r := blend(rA)
	g := blend(gA)
	b := blend(bA)

	return r | (g << 5) | (b << 10)
}
