package spi

import (
	_ "embed"
	"fmt"
)

//go:embed res/firmware.bin
var FirmwareData []byte

const (
    INST_NONE = 0x00
    INST_READ = 0x03
    INST_RDSR = 0x05


    INST_RDID = 0x9F

)

type Firmware struct {
    Transmitting bool
    Data []uint8
    Idx uint32

    Addr uint32
    WriteEnabled bool
}

func (f *Firmware) Transfer(data []uint8) (reply []uint8, stat uint8){

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

        if f.WriteEnabled {
            return []uint8{2}, STAT_DONE
        }

        return []uint8{0}, STAT_DONE

    case INST_READ:

        switch len(data) {
        case 0, 1, 2, 3:
            return nil, STAT_CONT
        case 4:
            f.Addr =  uint32(data[1]) << 16
            f.Addr |= uint32(data[2]) << 8
            f.Addr |= uint32(data[3])
        }

        const BUF_SIZE = 1024

        var (
            buffer []uint8
            i uint32
        )

        for i = range BUF_SIZE {
            buffer = append(buffer, FirmwareData[f.Addr + i])
        }

        f.Addr += BUF_SIZE

        return buffer, STAT_CONT


    default:
        panic(fmt.Sprintf("UNKNOWN OR UN SETUP FIRMWARE INST CODE %02X", inst))
        // this will be the same as INST_NONE but not yet
    }
}
