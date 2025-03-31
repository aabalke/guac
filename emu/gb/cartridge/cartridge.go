package cartridge

import (
	"fmt"
	"strings"
    re "regexp"
)

type Cartridge struct {
    Title    string
    Path     string
	Data     []uint8
    RamData  []uint8
	Type     uint8
	RomSize  int
	RamSize  int
	Checksum bool
	Valid    bool
	Mbc      Mbc
    ColorMode   bool

    RomBank  uint8
    RamBank  uint8
}

const (
	// bytes
	KiB = 32768
	MiB = KiB * 1024

	// Header Addrs
	TYPE = 0x147
	ROM  = 0x148
	RAM  = 0x149
	SUM  = 0x14D
)

func (c *Cartridge) ParseHeader() {

	c.Type = c.Data[TYPE]

    c.setTitle()
    c.setSavePath()
	c.setRomSize()
	c.setRamSize()
    c.setCGB()
	// may need to validate size here
	c.validateChecksum()
}

func (c *Cartridge) setTitle() {
    c.Title = strings.Trim(string(c.Data[0x134:0x143]), string(byte(0b0)))
}

func (c *Cartridge) setSavePath() {
    s := strings.ReplaceAll(strings.TrimSpace(strings.ToLower(c.Title)), " ", "_")

    r, err := re.Compile("[^a-z_]")
    if err != nil {
        panic(err)
    }

    println(s)
    s = r.ReplaceAllLiteralString(s, "")

    c.Path = fmt.Sprintf("./sav/%s.sav", s)
}

func (c *Cartridge) setRomSize() {

	switch c.Data[ROM] {
	case 0x00:
		c.RomSize = 32 * KiB
	case 0x01:
		c.RomSize = 64 * KiB
	case 0x02:
		c.RomSize = 128 * KiB
	case 0x03:
		c.RomSize = 256 * KiB
	case 0x04:
		c.RomSize = 512 * KiB
	case 0x05:
		c.RomSize = 1 * MiB
	case 0x06:
		c.RomSize = 2 * MiB
	case 0x07:
		c.RomSize = 4 * MiB
	case 0x08:
		c.RomSize = 8 * MiB
	}
}

func (c *Cartridge) setRamSize() {

	validCodes := []uint8{
		0x02, 0x03, 0x08, 0x09, 0x0C,
		0x0D, 0x10, 0x12, 0x13, 0x1A,
		0x1B, 0x1D, 0x1E, 0x22, 0xFF,
	}

	for _, code := range validCodes {

		if c.Type != code {
			continue
		}

		switch c.Data[RAM] {
		case 0x00:
			c.RamSize = 0
		case 0x02:
			c.RamSize = 8 * KiB
		case 0x03:
			c.RamSize = 32 * KiB
		case 0x04:
			c.RamSize = 128 * KiB
		case 0x05:
			c.RamSize = 64 * KiB
		}
	}

	c.RamSize = 0
}

func (c *Cartridge) validateChecksum() {
	var check uint8 = 0
	for addr := 0x134; addr <= 0x14C; addr++ {
		check = check - c.Data[addr] - 1
	}

	c.Valid = check == c.Data[SUM]
}

func (c *Cartridge) setCGB() {

    flag := c.Data[0x143]

    if flag == 0x80 || flag == 0xC0 {
        c.ColorMode = true
    }
}
