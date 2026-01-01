package utils

import "math"

func Convert28ToFloat(v uint32, bitFractional uint8) float32 {
	v &= 0xFFF_FFFF
	s := int32(v<<4) >> 4
	return float32(s) / float32(int(1)<<bitFractional)
}

//func Convert8_8Float(v int16) float32 {
//	return float32(v>>8) + (float32(v&0xFF) / 256.0)
//}

func ConvertFromFloat(f float32, bitFractional uint8) uint32 {
	scaled := f * float32(uint64(1)<<bitFractional)
	val := int32(math.Round(float64(scaled)))
	return uint32(val)
}

func ConvertToFloat(v uint32, bitFractional uint8) float32 {
	return float32(int32(v)) / float32(int(1)<<bitFractional)
}

func Convert16ToFloat(v uint16, bitFractional uint8) float32 {
	return float32(int16(v)) / float32(int(1)<<bitFractional)
}

func Convert10ToFloat(v uint16, bitFractional uint8) float32 {
	v &= 0x3FF
	s := int16(v<<6) >> 6
	return float32(s) / float32(int(1)<<bitFractional)
}

func ConvertFromFloat4_0_12(f float32) uint16 {
	const bitFractional = 12
	const totalBits = 16
	const signBits = 4

	scaled := f * float32(1<<bitFractional)
	val := int16(math.Round(float64(scaled)))

	maxAllowed := int16((1<<(signBits-1))-1) << bitFractional // 7 << 12
	minAllowed := -int16(1<<(signBits-1)) << bitFractional    // -8 << 12

	if val > maxAllowed {
		val = maxAllowed
	} else if val < minAllowed {
		val = minAllowed
	}

	return uint16(val)
}

func FloatRound(v, step float32) float32 {
	return float32(math.Round(float64(v/step)) * float64(step))
}

func FloatFloor(v, step float32) float32 {
	return float32(int(v/step)) * step
}
