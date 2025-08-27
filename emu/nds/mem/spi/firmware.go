package spi

import (
	_ "embed"
)

//go:embed res/firmware.bin
var firmware []byte

const (
    INST_NONE = 0x00
    INST_READ = 0x03

)

type Firmware struct {
    Transmitting bool
    Inst uint8
    Params []uint32
    Idx uint32
}

func (f *Firmware) Write(v uint8) {

    if f.Transmitting {
        return
    }


    switch f.Inst {
    case INST_NONE:
        f.Inst = v
        return

    case INST_READ:


        f.Params = append(f.Params, uint32(v))

        if len(f.Params) == 3 {
            f.Transmitting = true
            return
        }
    default:
        panic("UNKNOWN OR UN SETUP FIRMWARE INST CODE")
    }
}

func (f *Firmware) Read() uint8 {

    if !f.Transmitting {
        return 0
    }

    switch f.Inst {
    case INST_READ:

        addr := (f.Params[0] << 16) | (f.Params[1] << 8) | (f.Params[2])
        addr += f.Idx

        f.Idx++

        //fmt.Printf("INST %02X ADDR %02X V %02X\n", f.Inst, addr, firmware[addr])

        return firmware[addr]
    default:
        panic("READ FIRMWARE")
    }

    return 0
}

func (f *Firmware) Reset() {

    //fmt.Printf("RESET\n")

    // end transmition
    // write final byte to data?

    // reset inst, params, idx etc etc

    f.Transmitting = false
    f.Inst = INST_NONE
    f.Params = []uint32{}
    f.Idx = 0
}
