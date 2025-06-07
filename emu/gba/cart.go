package gba

import (
	"bufio"
	"fmt"
	"os"
	"slices"
)

type Cartridge struct {
    Gba     *GBA
	RomPath string
	SavPath string
	Header  *Header
    RomLength uint32
}

func NewCartridge(gba *GBA, rom, sav string) *Cartridge {

	c := &Cartridge{
        Gba: gba,
		RomPath: rom,
		SavPath: sav,
	}

    c.load()

	c.Header = NewHeader(c)

	return c
}

func (c *Cartridge) load() {

    mem := c.Gba.Mem

	buf, err := os.ReadFile(c.RomPath)
    if err != nil {
        panic(err)
    }

    c.RomLength = uint32(len(buf))
    fmt.Printf("LENGTH OF ROM IS %08X\n", len(buf))

    for i:= range len(buf) {
        mem.GamePak0[i] = uint8(buf[i])
    }

    // no save means sram full of 0xFF
    for i := range len(mem.SRAM) {
        mem.SRAM[i] = 0xFF
    }

    for i := range len(mem.Flash) {
        mem.Flash[i] = 0xFF
    }

	sBuf, err := os.ReadFile(c.SavPath)
    if err == nil {
        for i := range len(sBuf) {
            if i < len(mem.Flash) {
                mem.Flash[i] = uint8(sBuf[i])
                return
            }

            mem.SRAM[i - len(mem.Flash)] = uint8(sBuf[i])
        }
    }
}


func (c *Cartridge) save() {

    fmt.Printf("SAVING\n")

    mem := c.Gba.Mem

    f, err := os.Create(c.SavPath)
    if err != nil {
        panic(err)
    }
    defer f.Close()

    writer := bufio.NewWriter(f)

    bytes := slices.Concat(mem.Flash[:], mem.SRAM[:])
    _, err = writer.Write(bytes)
    if err != nil {
        panic(err)
    }
}
