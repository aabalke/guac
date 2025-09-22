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

func ConvertToFloat(v uint32, bitFractional uint8) float64 {
    return float64(int32(v)) / float64(int(1) << bitFractional)
}

func Convert16ToFloat(v uint16, bitFractional uint8) float64 {
    return float64(int16(v)) / float64(int(1) << bitFractional)
}

func Convert10ToFloat(v uint16, bitFractional uint8) float64 {
	v &= 0x3FF
	s := int16(v<<6) >> 6
	return float64(s) / float64(int(1)<<bitFractional)
}
