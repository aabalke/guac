package cart

import "fmt"

const (
	FL_READ = iota
	FL_ID
	FL_ERASE_ALL
	FL_ERASE
	FL_WRITE
	FL_BANKSWITCH
	FL_ERASE_SECTOR
)

func (c *Cartridge) ReadFlash(addr uint32) uint8 {
	var v uint8

	if c.FlashMode == FL_ID {
		switch addr {
		case 0:
			v = uint8(c.Manufacturer)
		case 1:
			v = uint8(c.Device)
		default:
			v = 0xFF
		}
	} else {
		bankAddr := (c.FlashBank * 0x1_0000) + addr

		v = c.Flash[bankAddr]

	}

	fmt.Printf("R ADDR %08X V %02X\n", addr, v)
	return v
}

func (c *Cartridge) WriteFlash(addr uint32, v uint8) {
	fmt.Printf("W ADDR %08X V %02X\n", addr, v)

	if c.SetMode(addr, v) {
		return
	}

	switch c.FlashMode {
	case FL_ID, FL_READ:
		return

	case FL_WRITE:
		bankAddr := (c.FlashBank * 0x1_0000) + addr

		// The target memory location must have been previously erased.
		c.Flash[bankAddr] &= v

	case FL_ERASE_ALL:

		for i := range len(c.Flash) {
			c.Flash[i] = 0xFF
		}

		c.FlashMode = FL_READ
		c.FlashStage = 0
		return

	case FL_ERASE_SECTOR:

		if addr&0xFFF != 0 {
			return
		}

		bankAddr := (c.FlashBank * 0x1_0000) + addr
		for i := range uint32(0x1000) {
			c.Flash[bankAddr+i] = 0xFF
		}

		c.FlashMode = FL_READ
		c.FlashStage = 0

	case FL_BANKSWITCH:
		if addr == 0 && c.FlashType == FLASH128 {
			c.FlashBank = uint32(v & 1)
		}
	}

	c.FlashMode = FL_READ
}

func (c *Cartridge) SetMode(addr uint32, v uint8) bool {
	switch c.FlashStage {
	case 0:
		if addr == 0x5555 && v == 0xAA {
			c.FlashStage = 1
			return true
		}
	case 1:
		if addr == 0x2AAA && v == 0x55 {
			c.FlashStage = 2
			return true
		}
	case 2:

		if c.FlashMode == FL_ERASE && v == 0x10 {
			c.FlashMode = FL_ERASE_ALL
			c.FlashStage = 0
			return false
		}

		if c.FlashMode == FL_ERASE && v == 0x30 {
			c.FlashMode = FL_ERASE_SECTOR
			c.FlashStage = 0
			return false
		}

		switch v {
		case 0x90:
			c.FlashMode = FL_ID
		case 0xF0:
			c.FlashMode = FL_READ
		case 0xA0:
			c.FlashMode = FL_WRITE
		case 0x80:
			c.FlashMode = FL_ERASE
		case 0xB0:
			c.FlashMode = FL_BANKSWITCH
		}
		c.FlashStage = 0
		return true

	default:
		c.FlashStage = 0
	}

	return false
}
