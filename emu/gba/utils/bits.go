package utils

import (
	"fmt"
)

func BitEnabled(value uint32, bit uint8) bool {

    if bit >= 0x20 {
        msg := fmt.Sprintf("bitEnabled max bit is 31. got=%d", bit)
        panic(msg)
    }

    var mask uint32 = 0b1 << uint32(bit)
    if value & mask == mask {
        return true
    }

    return false
}

func GetByte(i uint32, offsetBit uint8) uint32 {
    return GetVarData(i, offsetBit, offsetBit + 3)
}

func GetVarData(i uint32, s, e uint8) uint32 {

    switch {
    case s >= 0x21:
        msg := fmt.Sprintf("GetVarData max s is 31. got=%d", s)
        panic(msg)
    case e >= 0x21:
        msg := fmt.Sprintf("GetVarData max e is 31. got=%d", e)
        panic(msg)
    case e <= s:
        msg := fmt.Sprintf("GetVarData e >= s. got=%d >= %d", e, s)
        panic(msg)
    }

    width := e - s + 1
    mask := uint32((1 << width) - 1)
    return (i >> s) & mask
}

