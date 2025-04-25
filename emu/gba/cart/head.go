package gba

import "fmt"


type Header struct {
	Cartridge *Cartridge
	Title     string
}

func NewHeader(c *Cartridge) *Header {

	h := &Header{
		Cartridge: c,
		Title:     string(c.Data[0xA0 : 0xA0+12]),
	}

    h.valid()
    h.print()
	return h
}

func (h *Header) valid() bool {

    tests := []bool{
        h.Cartridge.Data[0xB2] == 0x96,
        h.Cartridge.Data[0xB5] == 0x00,
        h.Cartridge.Data[0xBE] == 0x00,
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
