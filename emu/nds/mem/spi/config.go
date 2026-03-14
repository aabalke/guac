package spi

import (
	"encoding/binary"
	"unicode/utf16"
)

func WriteUTF16(d *[]byte, s string, start, cnt uint32) {

	encodedString := utf16.Encode([]rune(s))

	for i := range cnt / 2 {
		v := uint16(0)
		if i < uint32(len(encodedString)) {
			v = encodedString[i]
		}
		binary.LittleEndian.PutUint16((*d)[start+i*2:], v)
	}
}

//go:inline
func Crc16(bytes []uint8, crc uint32) uint16 {

	var vals = [8]uint32{
		0xC0C1,
		0xC181,
		0xC301,
		0xC601,
		0xCC01,
		0xD801,
		0xF001,
		0xA001,
	}

	// crc inits in 0xFFFF, or 0x0

	for i := range len(bytes) {
		crc ^= uint32(bytes[i])
		for j := range 8 {
			carry := crc&1 != 0
			crc >>= 1
			if carry {
				crc ^= vals[j] << (7 - j)
			}
		}
	}

	return uint16(crc)
}
