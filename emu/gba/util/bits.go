package utils

import "fmt"

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
