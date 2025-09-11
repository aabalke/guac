package cart

import "fmt"

const (
	INST_NONE = 0x00

	INST_RDID = 0x9F
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

var backupData [0x80_0000]uint8 // this size may not be necessary

type Backup struct {
	Data []uint8
	Idx  uint32

	Addr         uint32
	WriteEnabled bool
	WriteBuffer  []uint8

    AutoDetect bool
    WrittenCnt bool
    AddrSize uint32
}

func (b *Backup) Init() {

	for i := range len(backupData) {
		backupData[i] = 0xFF
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
        // For 0.5k EEPROMS, cmd 0xB means "read high"
        b.Addr += 0x100
    }
}

func (b *Backup) checkSize() {

    return

    size := uint32(len(b.Data))

    if b.Addr < size {
        return
    }

    if size == 0 {
        size = 512
    }

    for size <= b.Addr {
        size *= 2
    }






}


func (b *Backup) Transfer(data []uint8) (reply []uint8, stat uint8) {

    fmt.Printf("BACKUP SPI % 2X\n", data)

	switch inst := data[0]; inst {
	case INST_NONE:

		return nil, STAT_DONE

	case INST_RDID:

		// 9Fh  RDID Read JEDEC Identification (Read 1..3 ID Bytes)
		// (Manufacturer, Device Type, Capacity)

		if len(data) < 1 {
			return nil, STAT_CONT
		}

		//ID 20h,40h,11h - ST 45PE10V6 - 128 Kbytes (Nintendo DSi) (nocash)

		return []uint8{0x20, 0x40, 0x11}, STAT_DONE

	case INST_RDSR:

		//05h  RDSR Read Status Register (Read Status Register, endless repeated)
		//Bit7-2  Not used (zero)
		//Bit1    WEL Write Enable Latch             (0=No, 1=Enable)
		//Bit0    WIP Write/Program/Erase in Progess (0=No, 1=Busy)

		if b.WriteEnabled {
			return []uint8{2}, STAT_DONE
		}

		return []uint8{0}, STAT_DONE

	case INST_READ, INST_RDHI:

        if b.AutoDetect && !b.Detect(data) {
            return nil, STAT_CONT
        }

        if uint32(len(data)) < b.AddrSize + 1 {
            return nil, STAT_CONT
        }

        if uint32(len(data)) == b.AddrSize + 1 {
            b.calcAddr(data)
        }

        b.checkSize()

        buf := make([]uint8, 256)
        sz := min(256, uint32(len(b.Data)) - b.Addr)
		copy(buf[:sz], b.Data[b.Addr:b.Addr+sz])
		return buf, STAT_CONT

	case INST_WRLO, INST_WRHI:

        if !b.WriteEnabled || b.AutoDetect {
            panic(fmt.Sprintf("Writing BACKUP while disabled or autodetect Enabled: %t AutoDetect %t\n", b.WriteEnabled, b.AutoDetect))
        }

        if uint32(len(data)) < b.AddrSize + 1 {
            return nil, STAT_CONT
        }

        if uint32(len(data)) == b.AddrSize + 1 {
            b.calcAddr(data)
        }

        b.checkSize()

        copy(b.Data[b.Addr:], data[1+b.AddrSize:])

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

        return nil, STAT_DONE

	//case INST_PW:

	//	switch len(data) {
	//	case 0, 1, 2, 3:
	//		return nil, STAT_CONT
	//	case 4:
	//		b.Addr = uint32(data[1]) << 16
	//		b.Addr |= uint32(data[2]) << 8
	//		b.Addr |= uint32(data[3])
	//	}

	//	b.WriteBuffer = data[4:]

	//	return nil, STAT_CONT

	default:
		panic(fmt.Sprintf("UNKNOWN OR UN SETUP BACKUP INST CODE %02X", inst))
		return nil, STAT_DONE
	}
}
