package ppu

// stores simd data to remove gc and init of large arrays
type Simd struct {
	PalData [SCREEN_WIDTH]uint32

    BlendPalettes [SCREEN_WIDTH]BlendPalettes

    // masks are uint16s that get filled if true
    BlendModeUse [5]bool
    BlendMasks[5][SCREEN_WIDTH / 16]uint16
}

func NewSimd(engine *Engine) *Simd {
    return &Simd{}
}
