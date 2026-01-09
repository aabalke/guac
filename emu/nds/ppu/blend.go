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
	NoBlendPalette, APalette, BPalette uint32
	hasA, hasB, targetATop             bool

	targetA3d bool
	alpha     uint32
}

func (bp *BlendPalettes) SetBgPalettes(palData, bgIdx uint32, targetA3d bool, alpha float32, bld *Blend) {

	bp.NoBlendPalette = palData

	if bld.a[bgIdx] {
		bp.APalette = palData
		bp.hasA = true
		bp.targetATop = true
		bp.targetA3d = false

		if targetA3d && bgIdx == 0 {
			bp.targetA3d = true
			bp.alpha = min(16, uint32(alpha*16))
		}

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

func (bp *BlendPalettes) SetObjPalettes(palData uint32, semiTransparent bool, bld *Blend) {

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
