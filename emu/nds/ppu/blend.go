package ppu

const (
	BLD_MODE_OFF   = 0
	BLD_MODE_STD   = 1
	BLD_MODE_WHITE = 2
	BLD_MODE_BLACK = 3
)

type BlendPalettes struct {
	Bld                                *Blend
	NoBlendPalette, APalette, BPalette uint32
	hasA, hasB, targetATop             bool

    targetA3d bool
    alpha float32
}

func NewBlendPalette(i uint32, bld *Blend, backdrop uint32) *BlendPalettes {

	bp := &BlendPalettes{
		Bld: bld,
	}

	//backdrop := gba.getPalette(0, 0, false)

	bp.NoBlendPalette = backdrop

	if bp.Bld.a[5] {
		bp.APalette = backdrop
		bp.hasA = true
		bp.targetATop = true
	}

	if bp.Bld.b[5] {
		bp.BPalette = backdrop
		bp.hasB = true
	}

	return bp
}

func (bp *BlendPalettes) SetBlendPalettes(palData uint32, bgIdx uint32, obj bool, semiTransparent, targetA3d bool, alpha float64) {

	bp.NoBlendPalette = palData

	if obj {

		if bp.Bld.a[4] || semiTransparent {
			bp.APalette = palData
			bp.hasA = true
			bp.targetATop = true
		} else {
			bp.targetATop = false
		}

		if bp.Bld.b[4] {
			bp.BPalette = palData
			bp.hasB = true
		}
		return
	}

	if bp.Bld.a[bgIdx] {
		bp.APalette = palData
		bp.hasA = true
		bp.targetATop = true

        if targetA3d && bgIdx == 0 {
            bp.targetA3d = true
            bp.alpha = float32(alpha)
        } else {
            bp.targetA3d = false
        }

	} else {
		bp.targetATop = false

        // not sure if this is required or correct
        bp.targetA3d = false
	}

	if bp.Bld.b[bgIdx] {
		bp.BPalette = palData
		bp.hasB = true
	}
}

func (bp *BlendPalettes) Blend(objTransparent bool, x, y uint32, wins *Windows, inObjWindow bool) uint32 {

    if !windowBldPixelAllowed(x, y, wins, inObjWindow) {
		return bp.noBlend(objTransparent)
    }

	switch bp.Bld.Mode {
	case BLD_MODE_OFF:
		return bp.noBlend(objTransparent)
	case BLD_MODE_STD:
		return bp.alphaBlend()
	case BLD_MODE_WHITE:
		return bp.grayscaleBlend(true)
	case BLD_MODE_BLACK:
		return bp.grayscaleBlend(false)
	default:
		return bp.noBlend(objTransparent)
	}
}

func (bp *BlendPalettes) noBlend(objTransparent bool) uint32 {
	if objTransparent {
		return bp.alphaBlend()
	}

	return bp.NoBlendPalette
}

func (bp *BlendPalettes) alphaBlend() uint32 {

	if !bp.hasA || !bp.hasB || !bp.targetATop {
		return bp.NoBlendPalette
	}

    if bp.targetA3d { // not sure if accurate

        rA := float32((bp.APalette) & 0x1F)
        gA := float32((bp.APalette >> 5) & 0x1F)
        bA := float32((bp.APalette >> 10) & 0x1F)
        rB := float32((bp.BPalette) & 0x1F)
        gB := float32((bp.BPalette >> 5) & 0x1F)
        bB := float32((bp.BPalette >> 10) & 0x1F)


        blend := func(a, b float32) uint32 {
            //val := a*bp.Bld.aEv + b*bp.Bld.bEv
            val := a*bp.alpha + b*(1 - bp.alpha)
            return uint32(min(31, val))
        }
        r := blend(rA, rB)
        g := blend(gA, gB)
        b := blend(bA, bB)

        return r | (g << 5) | (b << 10)

    }

	rA := float32((bp.APalette) & 0x1F)
	gA := float32((bp.APalette >> 5) & 0x1F)
	bA := float32((bp.APalette >> 10) & 0x1F)
	rB := float32((bp.BPalette) & 0x1F)
	gB := float32((bp.BPalette >> 5) & 0x1F)
	bB := float32((bp.BPalette >> 10) & 0x1F)

	blend := func(a, b float32) uint32 {
		val := a*bp.Bld.aEv + b*bp.Bld.bEv
		return uint32(min(31, val))
	}
	r := blend(rA, rB)
	g := blend(gA, gB)
	b := blend(bA, bB)

	return r | (g << 5) | (b << 10)
}

func (bp *BlendPalettes) grayscaleBlend(white bool) uint32 {

	if !bp.hasA || !bp.targetATop {
		return bp.NoBlendPalette
	}

	rA := float32((bp.APalette) & 0x1F)
	gA := float32((bp.APalette >> 5) & 0x1F)
	bA := float32((bp.APalette >> 10) & 0x1F)

	blend := func(v float32) uint32 {

		if white {
			v += (31 - v) * bp.Bld.yEv
		} else {
			v -= v * bp.Bld.yEv
		}

		return uint32(min(31, v))
	}

	r := blend(rA)
	g := blend(gA)
	b := blend(bA)

	return r | (g << 5) | (b << 10)
}
