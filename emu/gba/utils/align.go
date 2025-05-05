package utils


func WordAlign(addr uint32) uint32 {
    return addr &^ 0b11
}

func HalfAlign(addr uint32) uint32 {
    return addr &^ 0b1
}
