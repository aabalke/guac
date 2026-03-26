package apu

const (
	SAMP_MAX = 0x1ff
	SAMP_MIN = -0x200
)

//go:inline
func clip(v int32) int16 {
	if v > SAMP_MAX {
		return SAMP_MAX
	}
	if v < SAMP_MIN {
		return SAMP_MIN
	}
	return int16(v)
}

//func GetVarData(i uint32, s, e uint8) uint32 {
//	return (i >> s) & ((1 << (e - s + 1)) - 1)
//}
