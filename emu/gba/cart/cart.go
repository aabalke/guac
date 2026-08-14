package cart

import (
	"fmt"
	"log"
	"strings"

	"github.com/aabalke/guac/common/file"
	"github.com/aabalke/guac/config"
)

type Cartridge struct {
	path       string
	Title      string
	Code       string
	Id         int
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

	RomMask  uint32
	Mirrored bool

	Rom *[]uint8
	Sav *[]uint8
}

const (
	AUTO = iota
	NONE
	EEPROM
	SRAM
	FLASH
	FLASH128
)

func (c *Cartridge) String() string {
	return fmt.Sprintf("Gba Header: Title %12s, Code %4s, Id %d", c.Title, c.Code, c.Id)
}

func NewCartridge(path string) *Cartridge {
	c := Cartridge{
		path:    path,
		RomMask: 0x1FF_FFFF,
	}

	rom, sav, _ := file.Read(path)
	if rom == nil {
		panic(fmt.Sprintf("gba: rom path is invalid, could not load %s", path))
	}

	c.Rom = rom

	c.Id = config.Conf.Gba.Hardware.BackupType
	if c.Id == AUTO {
		c.Id = c.getCartBackupId()
	}

	if c.Id == FLASH || c.Id == FLASH128 {
		c.setFlashDevice(c.Id == FLASH128)
	}

	expectedLen := 0
	switch c.Id {
	case SRAM:
		expectedLen = 0x10000
	case EEPROM:
		expectedLen = 0x2000
	case FLASH, FLASH128:
		expectedLen = 0x20000
	}

	if sav == nil && expectedLen != 0 {
		s := make([]uint8, expectedLen)
		sav = &s
		for i := range len(s) {
			s[i] = 0xFF
		}
	}

	if expectedLen != len(*sav) {
		panic(fmt.Sprintf("gba: Sav Size != Save File Size %d != %d\n", expectedLen, len(*sav)))
	}

	c.Sav = sav

	c.Title = strings.ToUpper(strings.ReplaceAll(string((*c.Rom)[0xA0:0xA0+12]), "\x00", " "))
	c.Code = strings.ToUpper(string((*c.Rom)[0xAC : 0xAC+4]))

	// some dumps of classic nes games are only 1mb, since mirrored the rest of 4mb minimum
	// to fix we lower rom mask to fit smaller carts
	if classic := (*c.Rom)[0xAC] == 'F'; classic {
		c.RomMask = RoundPowerOfTwo(uint32(len(*c.Rom))) - 1
		c.Mirrored = true
	}

	return &c
}

func (c *Cartridge) getCartBackupId() int {
	// gbatek says has to be word aligned - need to confirm

	for i := range len(*c.Rom) - 4 {
		switch string((*c.Rom)[i : i+4]) {
		case "EEPR":
			return EEPROM
		case "SRAM":
			return SRAM
		case "FLAS":

			if i >= len(*c.Rom)-8 {
				continue
			}

			if string((*c.Rom)[i:i+8]) == "FLASH1M_" {
				return FLASH128
			}

			if string((*c.Rom)[i:i+6]) == "FLASH_" ||
				string((*c.Rom)[i:i+8]) == "FLASH512" {
				return FLASH
			}
		}
	}

	return NONE
}

func (c *Cartridge) setFlashDevice(flash128 bool) {
	if flash128 {
		// sanyo
		c.Device = [2]uint8{0x62, 0x13}
		return
	}
	// sst
	c.Device = [2]uint8{0xBF, 0xD4}
}

func (c *Cartridge) Save() {
	file.Write(c.path, c.Sav)
}

func (c *Cartridge) Read(addr uint32) uint8 {
	switch c.Id {
	case SRAM:
		return (*c.Sav)[addr]
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
		(*c.Sav)[addr] = v
	case EEPROM:
		log.Printf("Attempted GPIO Write with EEPROM. Not supported.\n")
	case FLASH, FLASH128:
		c.WriteFlash(addr, v)
	}
}

func RoundPowerOfTwo(v uint32) uint32 {
	s := uint32(1)

	for s < v {
		s *= 2
	}

	return s
}
