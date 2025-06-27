package apu

import "encoding/binary"

func Bit(val any, idx int) bool {
	switch val := val.(type) {
	case uint64:
		if idx < 0 || idx > 63 {
			return false
		}
		return (val & (1 << idx)) != 0

	case uint32:
		if idx < 0 || idx > 31 {
			return false
		}
		return (val & (1 << idx)) != 0

	case uint:
		if idx < 0 || idx > 31 {
			return false
		}
		return (val & (1 << idx)) != 0

	case uint16:
		if idx < 0 || idx > 15 {
			return false
		}
		return (val & (1 << idx)) != 0

	case byte:
		if idx < 0 || idx > 7 {
			return false
		}
		return (val & (1 << idx)) != 0

	case int64:
		if idx < 0 || idx > 63 {
			return false
		}
		return (val & (1 << idx)) != 0

	case int32:
		if idx < 0 || idx > 31 {
			return false
		}
		return (val & (1 << idx)) != 0

	case int:
		if idx < 0 || idx > 31 {
			return false
		}
		return (val & (1 << idx)) != 0

	case int16:
		if idx < 0 || idx > 15 {
			return false
		}
		return (val & (1 << idx)) != 0

	case int8:
		if idx < 0 || idx > 7 {
			return false
		}
		return (val & (1 << idx)) != 0
	}
	return false
}

// LE32 reads 32bit little-endian value from byteslice
func LE32(bs []byte) uint32 {
	switch len(bs) {
	case 0:
		return 0
	case 1:
		return uint32(bs[0])
	case 2:
		return uint32(bs[1])<<8 | uint32(bs[0])
	case 3:
		return uint32(bs[2])<<16 | uint32(bs[1])<<8 | uint32(bs[0])
	default:
		return binary.LittleEndian.Uint32(bs)
	}
}

func AddInt32(u uint32, i int32) uint32 {
	if i > 0 {
		u += uint32(i)
	} else {
		u -= uint32(-i)
	}
	return u
}
