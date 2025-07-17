package apu

const (
	SAMP_MAX = 0x1ff
	SAMP_MIN = -0x200
)

func clip(v int32) int16 {
    if v > SAMP_MAX { return SAMP_MAX }
    if v < SAMP_MIN { return SAMP_MIN }
	return int16(v)
}
