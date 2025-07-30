package cart

import (
	"fmt"
	"strings"
)

type Header struct {
	Cartridge *Cartridge
	Title     string
	GameCode  string
}

func NewHeader(c *Cartridge) *Header {

	h := &Header{
		Cartridge: c,
		Title:     string(c.Rom[0xA0 : 0xA0+12]),
		GameCode:  string(c.Rom[0xAC : 0xAC+4]),
	}

	if strings.HasPrefix(h.GameCode, "F") {
		panic("NES CLASSIC GAME. NOT SUPPORTED")
	}

	h.valid()
	h.print()
	return h
}

func (h *Header) valid() bool {

	tests := []bool{
		h.Cartridge.Rom[0xB2] == 0x96,
		h.Cartridge.Rom[0xB5] == 0x00,
		h.Cartridge.Rom[0xBE] == 0x00,
	}

	for _, pass := range tests {
		if !pass {
			return false
		}
	}

	return true
}

func (h *Header) print() {
	println(fmt.Sprintf("GBA Cartridge Information"))
	println(fmt.Sprintf("Title: %s", h.Title))
}
