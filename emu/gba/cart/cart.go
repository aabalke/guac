package gba

import (
	"os"
    "fmt"
)

type Cartridge struct {
	RomPath string
	SavPath string
	Data    [0x0200_0000]uint8
    SRAM    [0x1_0000]uint8
	Header  *Header
    RomLength uint32
}

func NewCartridge(rom, sav string) *Cartridge {

	c := &Cartridge{
		RomPath: rom,
		SavPath: sav,
        Data: [0x0200_0000]uint8{},
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

    // no save means sram full of 0xFF
    for i := range len(c.SRAM) {
        c.SRAM[i] = 0xFF
    }

    c.RomLength = uint32(len(buf))

    fmt.Printf("LENGTH OF ROM IS %08X\n", len(buf))


    for i:= range len(buf) {
        c.Data[i] = uint8(buf[i])
        //c.Data = append(c.Data, uint8(buf[i]))
    }
}
