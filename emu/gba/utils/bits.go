package utils

func BitEnabled(v uint32, bit uint8) bool {
	return v&(1<<bit) != 0
}

func GetByte(i uint32, offsetBit uint8) uint32 {
	return GetVarData(i, offsetBit, offsetBit+3)
}

func GetVarData(i uint32, s, e uint8) uint32 {
	return (i >> s) & ((1 << (e - s + 1)) - 1)
}

func CountBits(v uint32) uint32 {
	count := uint32(0)
	for v != 0 {
		count += (v & 1)
		v >>= 1
	}
	return count
}

func ReplaceByte(value uint32, newByte uint32, byteOffset uint32) uint32 {
	bitOffset := 8 * byteOffset
	mask := uint32(0b1111_1111)
	return (value &^ (mask << bitOffset)) | (newByte << bitOffset)
}
