package cart

import (
	"bufio"
	"fmt"
	"log"
	"os"

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

    // will replace this later
    DATA_SIZE = 0x80_0000
)

type Backup struct {
    SavPath string
    Data [DATA_SIZE]uint8
	Idx  uint32
    SaveFlag *bool

	Addr         uint32
	WriteEnabled bool

    AutoDetect bool
    WrittenCnt bool
    AddrSize uint32

    WriteProtection uint8
}

func (b *Backup) Init(savpath string, saveFlag *bool) {

    b.SavPath = savpath
    b.SaveFlag = saveFlag

	for i := range len(b.Data) {
		b.Data[i] = 0xFF
	}

    b.AutoDetect = true

	sBuf, err := os.ReadFile(b.SavPath)
	if err != nil {
		return
	}

	for i := range len(sBuf) {
        b.Data[i] = uint8(sBuf[i])
	}

}

func (b *Backup) Save() {

    log.Printf("Saving Game Path: %s\n", b.SavPath)

	f, err := os.Create(b.SavPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)

    _, err = writer.Write(b.Data[:])
	if err != nil {
		panic(err)
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
    if b.Addr >= DATA_SIZE {
        panic("BACKUP DATA TRANSFER IS BIGGER THAN DATA SIZE")
    }
}


func (b *Backup) Transfer(data []uint8) (reply []uint8, stat uint8) {

    //fmt.Printf("BACKUP SPI % 2X\n", data)

    //if data[0] == 0x06 { panic("WHAT")}

	switch inst := data[0]; inst {
	case INST_NONE:

		return nil, STAT_DONE

	case INST_RDSR:

        v := uint8(0xF0)
        //v := uint8(0x0)

		if b.WriteEnabled {
            v |= 2
		}

        v |= b.WriteProtection << 2

		return []uint8{v}, STAT_CONT

	case INST_READ, INST_RDHI:

        if b.AutoDetect && !b.Detect(data) {
            return nil, STAT_CONT
        }

        //b.AddrSize = 256

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

        *b.SaveFlag = true

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

        b.WriteProtection = (data[1] >> 2) & 0b11

        return nil, STAT_DONE

	default:
		panic(fmt.Sprintf("UNKNOWN OR UN SETUP BACKUP INST CODE %02X", inst))
		return nil, STAT_DONE
	}
}
