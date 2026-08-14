package cart

import (
	"fmt"
	"slices"
	"strings"

	"github.com/aabalke/guac/common/file"
)

type Cartridge struct {
	Title     string
	RomPath   string
	Data      []uint8
	RamData   []uint8
	Type      uint8
	RomSize   int
	RamSize   int
	Checksum  bool
	Valid     bool
	Mbc       Mbc
	ColorMode bool

	RomBank uint8
	RamBank uint8
	RomMask uint32
	RamMask uint32
}

const (
	TYPE = 0x147
	ROM  = 0x148
	RAM  = 0x149
	SUM  = 0x14D
)

func NewCartridge(path string) *Cartridge {
	c := &Cartridge{
		RomPath: path,
	}

	rom, sav, rtc := file.Read(path)
	if rom == nil {
		panic(fmt.Sprintf("gb: rom path is invalid, could not load %s", path))
	}

	c.ParseHeader(*rom)

	c.Data = make([]uint8, c.RomSize)
	copy(c.Data, *rom)

	if c.RamSize != 0 {
		c.RamData = make([]uint8, c.RamSize)
		if sav != nil {
			copy(c.RamData, *sav)
		}
	}

	fmt.Printf("Title: %s\n", c.Title)

	switch c.Type {
	case 0x00, 0x08, 0x09:
		c.Mbc = NewMbc0(c)
	case 0x01, 0x02, 0x03:
		c.Mbc = NewMbc1(c)
	case 0x05, 0x06:
		c.Mbc = NewMbc2(c)
	case 0x0F, 0x10, 0x11, 0x12, 0x13:
		c.Mbc = NewMbc3(c, rtc)
	case 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E:
		c.Mbc = NewMbc5(c)
	default:
		panic(fmt.Sprintf("UNSUPPORTED CART MAP TYPE %X", c.Type))
	}

	return c
}

var romSize = [...]int{
	1 << 15, // 32 * 1024
	1 << 16, // 64 * 1024
	1 << 17, // 128 * 1024
	1 << 18, // 256 * 1024
	1 << 19, // 512 * 1024
	1 << 20, // 1 * 1024 * 1024
	1 << 21, // 2 * 1024 * 1024
	1 << 22, // 4 * 1024 * 1024
	1 << 23, // 8 * 1024 * 1024
}

var romMask = [...]uint32{
	(1 << 15) - 1, // 32 * 1024
	(1 << 16) - 1, // 64 * 1024
	(1 << 17) - 1, // 128 * 1024
	(1 << 18) - 1, // 256 * 1024
	(1 << 19) - 1, // 512 * 1024
	(1 << 20) - 1, // 1 * 1024 * 1024
	(1 << 21) - 1, // 2 * 1024 * 1024
	(1 << 22) - 1, // 4 * 1024 * 1024
	(1 << 23) - 1, // 8 * 1024 * 1024
}

var ramSize = [...]int{
	0,
	0,
	1 << 13, // 8 * 1024
	1 << 15, // 32 * 1024
	1 << 17, // 128 * 1024
	1 << 16, // 64 * 1024
}

var ramMask = [...]uint32{
	0,
	0,
	(1 << 13) - 1,
	(1 << 15) - 1,
	(1 << 17) - 1,
	(1 << 16) - 1,
}

var validRamCodes = []uint8{
	0x02, 0x03, 0x08, 0x09, 0x0C,
	0x0D, 0x10, 0x12, 0x13, 0x1A,
	0x1B, 0x1D, 0x1E, 0x22, 0xFF,
}

func (c *Cartridge) ParseHeader(buf []uint8) {
	c.Type = buf[TYPE]
	c.Title = strings.Trim(string(buf[0x134:0x143]), string(byte(0)))
	c.RomSize = romSize[buf[ROM]]
	c.RomMask = romMask[buf[ROM]]

	if isRamCode := slices.Contains(validRamCodes, c.Type); isRamCode {
		c.RamSize = ramSize[buf[RAM]]
		c.RamMask = ramMask[buf[RAM]]
	}

	if c.RamSize == 0 && c.Type == 0x02 || c.Type == 0x03 {
		c.RamSize = ramSize[4]
		c.RamMask = ramMask[4]
	}

	if mbc2 := c.Type == 0x5 || c.Type == 0x6; mbc2 {
		c.RamSize = 1 << 18
		c.RamMask = (1 << 18) - 1
	}

	c.ColorMode = buf[0x143]&0x80 != 0

	check := uint8(0)
	for addr := 0x134; addr <= 0x14C; addr++ {
		check = check - buf[addr] - 1
	}

	c.Valid = check == buf[SUM]
}
