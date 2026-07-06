package cart

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

type Cartridge struct {
	RomPath    string
	SavPath    string
	Title      string
	Code       string
	Id         int
	FlashType  int
	FlashMode  int
	FlashBank  uint32
	Device     [2]uint8
	FlashStage uint32

	EepromReadBits       uint64
	EepromWriteBits      uint64
	EepromReadBitsCount  uint32
	EepromWriteBitsCount uint32
	EepromWidth          uint32
	EepromAddr           uint32
	EepromState          uint32

	Rom *[]uint8
	Sav []uint8
}

const (
	NONE = iota
	EEPROM
	SRAM
	FLASH
	FLASH128
)

const (
	TYPE_SST = iota
	TYPE_MACRONIX64
	TYPE_PANASONIC
	TYPE_ATMEL
	TYPE_SANYO
	TYPE_MACRONIX128
)

var idType = [...]string{"none", "eeprom", "sram", "flash", "flash128"}

func (c *Cartridge) String() string {
	return fmt.Sprintf("Gba Header: Title %12s, Code %4s, Id %8s", c.Title, c.Code, idType[c.Id])
}

func NewCartridge(rom, sav string) *Cartridge {
	c := Cartridge{
		RomPath: rom,
		SavPath: sav,
	}

	buf, err := os.ReadFile(c.RomPath)
	if err != nil {
		panic(err)
	}

	c.Rom = &buf

	c.Id = c.getCartBackupId()
	switch c.Id {
	case SRAM:
		c.Sav = make([]uint8, 0x10000)
	case EEPROM:
		c.Sav = make([]uint8, 0x2000)
	case FLASH, FLASH128:
		c.Sav = make([]uint8, 0x20000)
	}

	if sBuf, err := os.ReadFile(c.SavPath); err != nil {
		for i := range len(c.Sav) {
			c.Sav[i] = 0xFF
		}
	} else {
		if len(c.Sav) != len(sBuf) {
			fmt.Printf("Sav Size != Save File Size %d != %d\n", len(c.Sav), len(sBuf))
			panic("BAD")
		}

		c.Sav = sBuf

	}

	c.Title = strings.ToUpper(strings.ReplaceAll(string((*c.Rom)[0xA0:0xA0+12]), "\x00", " "))
	c.Code = strings.ToUpper(string((*c.Rom)[0xAC : 0xAC+4]))

	return &c
}

func (c *Cartridge) getCartBackupId() int {
	// have to be word aligned // maybe not???

	for i := 0; i < len(*c.Rom)-4; i++ {
		switch string((*c.Rom)[i : i+4]) {
		case "EEPR":
			return EEPROM
		case "SRAM":
			return SRAM
		case "FLAS":

			if i < len(*c.Rom)-8 && string((*c.Rom)[i:i+8]) == "FLASH1M_" {
				c.Device = [2]uint8{0x62, 0x13}
				c.FlashType = TYPE_SANYO
				return FLASH128
			}

			c.Device = [2]uint8{0xBF, 0xD4}
			c.FlashType = TYPE_SST
			return FLASH
		}
	}

	return NONE
}

func (c *Cartridge) Save() {
	log.Printf("Saving Game Path: %s\n", c.SavPath)

	f, err := os.Create(c.SavPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	_, err = writer.Write(c.Sav[:])
	if err != nil {
		panic(err)
	}
}

func (c *Cartridge) Read(addr uint32) uint8 {
	switch c.Id {
	case SRAM:
		return c.Sav[addr]
	case EEPROM:
		log.Printf("Attempted GPIO Read with EEPROM. Not supported.\n")
		return 0
	case FLASH, FLASH128:
		return c.ReadFlash(addr)
	default:
		return 0xFF
	}
}

func (c *Cartridge) Write(addr uint32, v uint8) {
	switch c.Id {
	case SRAM:
		c.Sav[addr] = v
	case EEPROM:
		log.Printf("Attempted GPIO Write with EEPROM. Not supported.\n")
	case FLASH, FLASH128:
		c.WriteFlash(addr, v)
	}
}
