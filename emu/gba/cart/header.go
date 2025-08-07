package cart

import (
	"log"
	"strings"
)

type Header struct {
	Title     string
	GameCode  string
    Version uint8
}

func NewHeader(c *Cartridge) *Header {

	h := &Header{
		Title:    strings.ToUpper(strings.ReplaceAll(string(c.Rom[0xA0 : 0xA0+12]), "\x00", " ")),
		GameCode: strings.ToUpper(string(c.Rom[0xAC : 0xAC+4])),
        Version:  uint8(c.Rom[0xBC]),
	}

	if strings.HasPrefix(h.GameCode, "F") {
		panic("NES CLASSIC GAME. NOT SUPPORTED")
	}

	h.valid(c)
	h.print()
	return h
}

func (h *Header) valid(c *Cartridge) bool {

	tests := []bool{
		c.Rom[0xB2] == 0x96,
		c.Rom[0xB5] == 0x00,
		c.Rom[0xBE] == 0x00,
	}

	for _, pass := range tests {
		if !pass {
			return false
		}
	}

	return true
}

func (h *Header) print() {
    log.Printf("GBA ROM %12s C %4s V %d\n", h.Title, h.GameCode, h.Version )
}
