package utils

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

func HexToColor(h string) color.Color {
	h, _ = strings.CutPrefix(h, "#")
	h, _ = strings.CutPrefix(h, "0x")

	u, err := strconv.ParseUint(h, 16, 0)
	if err != nil {
		return color.NRGBA{
			R: uint8(0),
			G: uint8(0),
			B: uint8(0),
			A: uint8(255),
		}
	}

	return color.NRGBA{
		R: uint8(u >> 16),
		G: uint8(u >> 8),
		B: uint8(u),
		A: uint8(255),
	}
}

func ColorToHex(c color.Color) string {
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	return fmt.Sprintf("%02X%02X%02X", n.R, n.G, n.B)
}

func ColorToUint32(c color.Color) uint32 {
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	return (uint32(n.A) << 24) | (uint32(n.R) << 0) | (uint32(n.G) << 8) | (uint32(n.B) << 16)
}
