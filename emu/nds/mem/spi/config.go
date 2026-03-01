package spi

import (
	"encoding/binary"
	"unicode/utf16"

	"github.com/aabalke/guac/config"
)

func FirmwareConfig() {

	f := &config.Conf.Nds.Firmware

	const ofs = 0x3FE00

	FirmwareData[ofs+0x02] = f.Color
	FirmwareData[ofs+0x03] = f.BirthdayMonth
	FirmwareData[ofs+0x04] = f.BirthdayDay

	WriteUTF16(f.Nickname, ofs+0x6, 20)
	FirmwareData[ofs+0x1A] = uint8(len(f.Nickname))
	FirmwareData[ofs+0x1B] = 0

	WriteUTF16(f.Message, ofs+0x1C, 52)
	FirmwareData[ofs+0x50] = uint8(len(f.Message))
	FirmwareData[ofs+0x51] = 0

	crc := Crc16(FirmwareData[ofs:ofs+0x70], 0xFFFF)
	binary.LittleEndian.PutUint16(FirmwareData[ofs+0x72:], crc)
}

func WriteUTF16(s string, start, cnt uint32) {

	encodedString := utf16.Encode([]rune(s))

	for i := range cnt / 2 {
		v := uint16(0)
		if i < uint32(len(encodedString)) {
			v = encodedString[i]
		}
		binary.LittleEndian.PutUint16(FirmwareData[start+i*2:], v)
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
