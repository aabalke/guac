package cart

import (
	"os"
)

type Cartridge struct {
	RomPath   string
	SavPath   string
	RomLength uint32
	Header    Header

	Rom []uint8
}

func NewCartridge(rom, sav string) Cartridge {

	c := Cartridge{
		RomPath: rom,
		SavPath: sav,
		Rom:     make([]uint8, 0),
	}

	c.load()

	c.Header = NewHeader(&c)

	return c
}

func (c *Cartridge) load() {

	buf, err := os.ReadFile(c.RomPath)
	if err != nil {
		panic(err)
	}

	c.RomLength = uint32(len(buf))

	for i := range len(buf) {
		c.Rom = append(c.Rom, buf[i])
	}
}
