package apu

const (
	SAMP_MAX = 0x1ff
	SAMP_MIN = -0x200
)

func clip(v int32) int16 {
    v = min(v, SAMP_MAX)
    v = max(v, SAMP_MIN)
	return int16(v)
}
