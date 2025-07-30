package cart

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

type Cartridge struct {
	RomPath      string
	SavPath      string
	Header       *Header
	RomLength    uint32
	Id           int
	FlashType    int
	FlashMode    int
	FlashBank    uint32
	Manufacturer uint32
	Device       uint32
	FlashStage   uint32

	Rom    [0x200_0000]uint8
	SRAM   [0x1_0000]uint8
	Flash  [0x2_0000]uint8 // multiple banks
	Eeprom [0x2000]uint8
}

const (
	NONE     = 0
	EEPROM   = 1
	SRAM     = 2
	FLASH    = 3
	FLASH128 = 4

	TYPE_SST         = 0
	TYPE_MACRONIX64  = 1
	TYPE_PANASONIC   = 2
	TYPE_ATMEL       = 3
	TYPE_SANYO       = 4
	TYPE_MACRONIX128 = 5
)

func NewCartridge(rom, sav string) Cartridge {

	c := Cartridge{
		RomPath: rom,
		SavPath: sav,
	}

	c.load()

	c.Header = NewHeader(&c)

	switch c.Id {
	case NONE:
		log.Printf("Cartridge Type NONE\n")
	case EEPROM:
		log.Printf("Cartridge Type EEPROM\n")
	case SRAM:
		log.Printf("Cartridge Type SRAM\n")
	case FLASH:
		log.Printf("Cartridge Type FLASH64\n")
	case FLASH128:
		log.Printf("Cartridge Type FLASH128\n")
	}

	return c
}

func (c *Cartridge) getCartBackupId() int {

	// have to be word aligned // maybe not???

	for i := range len(c.Rom) {

		if i >= len(c.Rom)-4 {
			continue
		}

		switch string(c.Rom[i : i+4]) {
		case "EEPR":
			return EEPROM
		case "SRAM":
			return SRAM
		case "FLAS":

			if i < len(c.Rom)-8 && string(c.Rom[i:i+8]) == "FLASH1M_" {
				c.setDeviceManufacturer(true)
				c.FlashType = c.getDeviceManufacturer()
				return FLASH128
			}

			c.setDeviceManufacturer(false)
			c.FlashType = c.getDeviceManufacturer()
			return FLASH
		}
	}

	return NONE
}

func (c *Cartridge) setDeviceManufacturer(size128 bool) {

	// set small to SST, and large to Sanyo - this avoids Macronix and Atmel specific cmds
	// do some games require the special cmds???

	if size128 {
		c.Flash[1] = 0x62
		c.Flash[0] = 0x13
		return
	}

	c.Flash[1] = 0xBF
	c.Flash[0] = 0xD4
}

func (c *Cartridge) getDeviceManufacturer() int {

	code := (uint16(c.Flash[0]) << 8) | uint16(c.Flash[1])

	c.Device = uint32(c.Flash[0])
	c.Manufacturer = uint32(c.Flash[1])

	switch code {
	case 0xD4BF:
		return TYPE_SST
	case 0x1CC2:
		return TYPE_MACRONIX64
	case 0x1B32:
		return TYPE_PANASONIC
	case 0x3D1F:
		return TYPE_ATMEL
	case 0x1362:
		return TYPE_SANYO
	case 0x09C2:
		return TYPE_MACRONIX128
	}

	panic("UNKNOWN FLASH ROM DEVICE AND MANUFACTUER")
}

func (c *Cartridge) load() {

	buf, err := os.ReadFile(c.RomPath)
	if err != nil {
		panic(err)
	}

	c.RomLength = uint32(len(buf))
	fmt.Printf("LENGTH OF ROM IS %08X\n", len(buf))

	for i := range len(buf) {
		c.Rom[i] = uint8(buf[i])
	}

	c.Id = c.getCartBackupId()

	// sav

	sBuf, err := os.ReadFile(c.SavPath)

	if err != nil {
		switch c.Id {
		case SRAM:
			for i := range len(c.SRAM) {
				c.SRAM[i] = 0xFF
			}
		case EEPROM:
			for i := range len(c.Eeprom) {
				c.Eeprom[i] = 0xFF
			}
		default:
			for i := range len(c.Flash) {
				c.Flash[i] = 0xFF
			}
		}

		return
	}

	for i := range len(sBuf) {
		switch c.Id {
		case SRAM:
			c.SRAM[i] = uint8(sBuf[i])
		case EEPROM:
			c.Eeprom[i] = uint8(sBuf[i])
		default:
			c.Flash[i] = uint8(sBuf[i])
		}
	}
}

func (c *Cartridge) Save() {

	log.Printf("Saving Game Path: %s\n", c.SavPath)

	f, err := os.Create(c.SavPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)

	var bytes []uint8
	switch c.Id {
	case SRAM:
		bytes = c.SRAM[:]
	case EEPROM:
		bytes = c.Eeprom[:]
	default:
		bytes = c.Flash[:]
	}

	_, err = writer.Write(bytes)
	if err != nil {
		panic(err)
	}
}

func (c *Cartridge) Read(addr uint32) uint8 {

	switch c.Id {
	case SRAM:
		return c.SRAM[addr]
	case EEPROM:
		panic("EEPROM READ USING STD IMPLIMENTATION")
	default:
		return c.ReadFlash(addr)
	}
}

func (c *Cartridge) Write(addr uint32, v uint8) {
	switch c.Id {
	case SRAM:
		c.SRAM[addr] = v
	case EEPROM:
		panic("EEPROM WRITE USING STD IMPLIMENTATION")
	default:
		c.WriteFlash(addr, v)
	}
}
