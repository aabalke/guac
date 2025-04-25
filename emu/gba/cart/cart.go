package gba

import (
	"os"
)

type Cartridge struct {
	RomPath string
	SavPath string
	Data    []uint8
	Header  *Header
}

func NewCartridge(rom, sav string) *Cartridge {

	c := &Cartridge{
		RomPath: rom,
		SavPath: sav,
        Data: []uint8{},
	}

    c.load()

	c.Header = NewHeader(c)

	return c
}

func (c *Cartridge) load() {

	buf, err := os.ReadFile(c.RomPath)
    if err != nil {
        panic(err)
    }

    for i:= range len(buf) {
        c.Data = append(c.Data, uint8(buf[i]))
    }
}
