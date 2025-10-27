package utils

import "math"

func Convert28ToFloat(v uint32, bitFractional uint8) float64 {
	v &= 0xFFF_FFFF
	s := int32(v<<4) >> 4
	return float64(s) / float64(int(1)<<bitFractional)
}

//func Convert8_8Float(v int16) float64 {
//	return float64(v>>8) + (float64(v&0xFF) / 256.0)
//}

func ConvertFromFloat(f float64, bitFractional uint8) uint32 {
	scaled := f * float64(uint64(1)<<bitFractional)
	val := int32(math.Round(scaled))
	return uint32(val)
}

func ConvertToFloat(v uint32, bitFractional uint8) float64 {
	return float64(int32(v)) / float64(int(1)<<bitFractional)
}

func Convert16ToFloat(v uint16, bitFractional uint8) float64 {
	return float64(int16(v)) / float64(int(1)<<bitFractional)
}

func Convert10ToFloat(v uint16, bitFractional uint8) float64 {
	v &= 0x3FF
	s := int16(v<<6) >> 6
	return float64(s) / float64(int(1)<<bitFractional)
}

func ConvertFromFloat4_0_12(f float64) uint16 {
    const bitFractional = 12
    const totalBits = 16
    const signBits = 4

    scaled := f * float64(1<<bitFractional)
    val := int16(math.Round(scaled))

    maxAllowed := int16((1 << (signBits - 1)) - 1) << bitFractional // 7 << 12
    minAllowed := -int16(1 << (signBits - 1)) << bitFractional      // -8 << 12

    if val > maxAllowed {
        val = maxAllowed
    } else if val < minAllowed {
        val = minAllowed
    }

    return uint16(val)
}
