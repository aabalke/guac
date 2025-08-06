package cart

// RidgeX/ygba BSD3 License

import "fmt"

const (
	EE_MODE_IDLE       = 0
	EE_MODE_WRITE_INIT = 1
	EE_MODE_WRITE      = 2
	EE_MODE_READ       = 3
	EE_MODE_READ_INIT  = 4
)

var (
	eepromReadBitsCount uint32
	eepromReadBits      uint64

	eepromWriteBitsCount uint32
	eepromWriteBits      uint64

	EepromWidth uint32
	EepromAddr  uint32
	EepromState uint32
)

func (c *Cartridge) EepromRead() uint16 {

	//println("READING EEPROM")

	switch {
	case eepromReadBitsCount > 64:
		eepromReadBitsCount--
		return 1
	case eepromReadBitsCount > 0:
		eepromReadBitsCount--
		return uint16(eepromReadBits>>uint64(eepromReadBitsCount)) & 1
	default:
		return 1
	}
}

func (c *Cartridge) EepromWrite(v uint16) {

	if EepromWidth == 0 {
		panic("EEPROM WIDTH 0")
	}

	eepromWriteBits <<= 1
	eepromWriteBits |= uint64(v & 1)
	eepromWriteBitsCount++

	switch EepromState {
	case 0: // Start of stream
		if eepromWriteBitsCount < 2 {
			return
		}
		EepromState = uint32(eepromWriteBits)

		if EepromState != 2 && EepromState != 3 {
			panic("EEPROM INCORRECT START STREAM STATE")
		}

		eepromWriteBits = 0
		eepromWriteBitsCount = 0

	case 1: // End of stream

		EepromState = 0
		eepromWriteBits = 0
		eepromWriteBitsCount = 0

	case 2: // Write request
		if eepromWriteBitsCount < EepromWidth {
			return
		}
		EepromAddr = uint32(eepromWriteBits * 8)
		eepromReadBits = 0
		eepromReadBitsCount = 0
		EepromState = 4
		eepromWriteBits = 0
		eepromWriteBitsCount = 0

	case 3: // Read request
		if eepromWriteBitsCount < EepromWidth {
			return
		}
		EepromAddr = uint32(eepromWriteBits * 8)

		//if EepromAddr > uint32(len(c.Eeprom)) {
		//    fmt.Printf("3 ADDR %08X", EepromAddr)
		//    panic("TOO BIG")
		//}

		eepromReadBits = 0
		eepromReadBitsCount = 68
		for i := range 8 {
			b := c.Eeprom[int(EepromAddr)+i]
			for j := 7; j >= 0; j-- {
				eepromReadBits <<= 1
				eepromReadBits |= uint64(b>>j) & 1
			}
		}
		EepromState = 1
		eepromWriteBits = 0
		eepromWriteBitsCount = 0

	case 4: // Data
		if eepromWriteBitsCount < 64 {
			return
		}

		for i := range 8 {
			b := uint8(0)
			for j := 7; j >= 0; j-- {
				b <<= 1
				b |= uint8(eepromWriteBits>>((7-i)*8+j)) & 1
			}

			if EepromAddr+uint32(i) > 8192 {
				fmt.Printf("EEPROM ADDR WRITING V %02X, ADDR %08X, I %08X\n", b, EepromAddr, i)
				panic("TOO BIG")
			}

			c.Eeprom[EepromAddr+uint32(i)] = uint8(b)
		}

		EepromState = 1
		eepromWriteBits = 0
		eepromWriteBitsCount = 0

	default:
		panic("UNKNOWN EEPROM STATE")
	}
}
