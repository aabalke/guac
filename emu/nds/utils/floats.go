package utils

func Convert20_8Float(v int32) float64 {

	// sign extend
	sBit := 27
	if v&(1<<sBit) != 0 {
		v |= ^((1 << ((sBit) + 1)) - 1)
	}

	return float64(v>>8) + (float64(v&0xFF) / 256.0)
}

func Convert8_8Float(v int16) float64 {
	return float64(v>>8) + (float64(v&0xFF) / 256.0)
}
