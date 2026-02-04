package cart

import (
	"fmt"
)

const (
	INST_NONE = 0x00

	INST_READ = 0x03
	INST_RDSR = 0x05

	INST_WREN = 0x06
	INST_WRDI = 0x04

	// need eeprom and fram specific inst
	INST_WRSR = 0x01

	INST_RDLO = 0x03
	INST_RDHI = 0x0B
	INST_WRLO = 0x02
	INST_WRHI = 0x0A

	STAT_CONT = 1
	STAT_DONE = 2
)

type Backup struct {
    Cartridge *Cartridge

	Addr         uint32
	WriteEnabled bool

	AutoDetect bool
	WrittenCnt bool
	AddrSize   uint32

	WriteProtection uint8

	Size uint32
	Type uint32
}

func NewBackup(c *Cartridge) *Backup {
    return &Backup{
        Cartridge: c,
        AutoDetect: true,
    }
}

func (b *Backup) Detect(data []uint8) bool {
	switch {
	case len(data) == 1:
		b.WrittenCnt = false
		return false
	case b.WrittenCnt:
		b.AddrSize = uint32(len(data) - 2)
		b.AutoDetect = false
		return true
	default:
		return false
	}
}

func (b *Backup) calcAddr(data []uint8) {
	b.Addr = 0
	for _, v := range data[1:] {
		b.Addr <<= 8
		b.Addr |= uint32(v)
	}

	if b.AddrSize != 1 {
		return
	}

	if hiCmd := data[0] == 0xB || data[0] == 0xA; hiCmd {
		b.Addr += 0x100
	}
}

func (b *Backup) checkSize() {
	//if b.Addr >= DATA_SIZE {
	//	panic("BACKUP DATA TRANSFER IS BIGGER THAN DATA SIZE")
	//}
}

func (b *Backup) Transfer(data []uint8) (reply []uint8, stat uint8) {

	//fmt.Printf("BACKUP SPI % 2X\n", data)

	//if data[0] == 0x06 { panic("WHAT")}

	switch inst := data[0]; inst {
	case INST_NONE:

		return nil, STAT_DONE

	case INST_RDSR:

		if len(data) == 1 {
			return nil, STAT_CONT
		}

		v := uint8(0xF0)
		//v := uint8(0x0)

		if b.WriteEnabled {
			v |= 2
		}

		v |= b.WriteProtection << 2

		//for i := range data {
		//    fmt.Printf("%02X ", data[i])
		//}

		//fmt.Printf("\n")

		return []uint8{v}, STAT_CONT

	case INST_RDHI:

		panic("read hi")

	case INST_READ:

		if b.AutoDetect && !b.Detect(data) {
			return nil, STAT_CONT
		}

		//b.AddrSize = 256

		if uint32(len(data)) < b.AddrSize+1 {
			return nil, STAT_CONT
		}

		if uint32(len(data)) == b.AddrSize+1 {
			b.calcAddr(data)
		}

		b.checkSize()

		buf := make([]uint8, 256)
		sz := min(256, uint32(len(b.Cartridge.Sav))-b.Addr)

		copy(buf[:sz], b.Cartridge.Sav[b.Addr:b.Addr+sz])

		//for i := range data {
		//    fmt.Printf("%02X ", data[i])
		//}

		//fmt.Printf("\n")
		return buf, STAT_CONT

	case INST_WRLO, INST_WRHI:

		if !b.WriteEnabled || b.AutoDetect {
			panic(fmt.Sprintf("Writing BACKUP while disabled or autodetect Enabled: %t AutoDetect %t\n", b.WriteEnabled, b.AutoDetect))
		}

		if uint32(len(data)) < b.AddrSize+1 {
			return nil, STAT_CONT
		}

		if uint32(len(data)) == b.AddrSize+1 {
			b.calcAddr(data)
		}

		b.checkSize()

		b.Cartridge.SaveFlag = true

		copy(b.Cartridge.Sav[b.Addr:], data[1+b.AddrSize:])

		return nil, STAT_CONT

	case INST_WREN:

		b.WriteEnabled = true
		return nil, STAT_DONE

	case INST_WRDI:

		b.WriteEnabled = false
		return nil, STAT_DONE

	case INST_WRSR:

		if len(data) < 2 {
			return nil, STAT_CONT
		}

		b.WriteProtection = (data[1] >> 2) & 0b11

		return nil, STAT_DONE

	default:
		panic(fmt.Sprintf("UNKNOWN OR UN SETUP BACKUP INST CODE %02X. DATA %02X", inst, data))
		return nil, STAT_DONE
	}
}
