package cart

// RidgeX/ygba BSD3 License

import "fmt"

const (
	EE_MODE_IDLE = iota
	EE_MODE_WRITE_INIT
	EE_MODE_WRITE
	EE_MODE_READ
	EE_MODE_READ_INIT
)

func (c *Cartridge) EepromRead() uint16 {
	switch {
	case c.EepromReadBitsCount > 64:
		c.EepromReadBitsCount--
		return 1
	case c.EepromReadBitsCount > 0:
		c.EepromReadBitsCount--
		return uint16(c.EepromReadBits>>uint64(c.EepromReadBitsCount)) & 1
	default:
		return 1
	}
}

func (c *Cartridge) EepromWrite(v uint16) {
	if c.EepromWidth == 0 {
		panic("EEPROM WIDTH 0")
	}

	c.EepromWriteBits <<= 1
	c.EepromWriteBits |= uint64(v & 1)
	c.EepromWriteBitsCount++

	switch c.EepromState {
	case 0: // Start of stream
		if c.EepromWriteBitsCount < 2 {
			return
		}
		c.EepromState = uint32(c.EepromWriteBits)

		if c.EepromState != 2 && c.EepromState != 3 {
			panic("EEPROM INCORRECT START STREAM STATE")
		}

		c.EepromWriteBits = 0
		c.EepromWriteBitsCount = 0

	case 1: // End of stream

		c.EepromState = 0
		c.EepromWriteBits = 0
		c.EepromWriteBitsCount = 0

	case 2: // Write request
		if c.EepromWriteBitsCount < c.EepromWidth {
			return
		}
		c.EepromAddr = uint32(c.EepromWriteBits * 8)
		c.EepromReadBits = 0
		c.EepromReadBitsCount = 0
		c.EepromState = 4
		c.EepromWriteBits = 0
		c.EepromWriteBitsCount = 0

	case 3: // Read request
		if c.EepromWriteBitsCount < c.EepromWidth {
			return
		}
		c.EepromAddr = uint32(c.EepromWriteBits * 8)

		//if c.EepromAddr > uint32(len(c.c.Eeprom)) {
		//    fmt.Printf("3 ADDR %08X", c.EepromAddr)
		//    panic("TOO BIG")
		//}

		c.EepromReadBits = 0
		c.EepromReadBitsCount = 68
		for i := range 8 {
			b := c.Sav[int(c.EepromAddr)+i]
			for j := 7; j >= 0; j-- {
				c.EepromReadBits <<= 1
				c.EepromReadBits |= uint64(b>>j) & 1
			}
		}
		c.EepromState = 1
		c.EepromWriteBits = 0
		c.EepromWriteBitsCount = 0

	case 4: // Data
		if c.EepromWriteBitsCount < 64 {
			return
		}

		for i := range 8 {
			b := uint8(0)
			for j := 7; j >= 0; j-- {
				b <<= 1
				b |= uint8(c.EepromWriteBits>>((7-i)*8+j)) & 1
			}

			if c.EepromAddr+uint32(i) > 8192 {
				fmt.Printf("EEPROM ADDR WRITING V %02X, ADDR %08X, I %08X\n", b, c.EepromAddr, i)
				panic("TOO BIG")
			}

			c.Sav[c.EepromAddr+uint32(i)] = uint8(b)
		}

		c.EepromState = 1
		c.EepromWriteBits = 0
		c.EepromWriteBitsCount = 0

	default:
		panic("UNKNOWN EEPROM STATE")
	}
}
