package spi

import (
	"encoding/binary"
	"fmt"

	"github.com/aabalke/guac/emu/nds/utils"
)

const (
	CH_TEMP0   = 0
	CH_TOUCHY  = 1
	CH_BATTERY = 2
	CH_TOUCHZ1 = 3
	CH_TOUCHZ2 = 4
	CH_TOUCHX  = 5
	CH_AUX     = 6
	CH_TEMP1   = 7

	MODE_DIFF = 0
	MODE_SING = 1
)

type Tsc struct {
	Control uint8

	//Temp0 uint16
	//Temp1 uint16
	//Aux uint16

	TouchX uint16
	TouchY uint16

    IrqEnabled bool

    TouchActive bool
}

func (t *Tsc) Transfer(data []uint8) (reply []uint8, stat uint8) {

	inst := data[0]

    //log.Printf("SPI Touchscr % 02X\n", data)

    if invalidStart := !utils.BitEnabled(uint32(inst), 7); invalidStart {
        //panic("INVALID START TO TOUCH TRANSFER")
        return nil, STAT_DONE
    }

    
    var (
        out uint16
        conv8 = utils.BitEnabled(uint32(inst), 3)
    )

    switch ch := utils.GetVarData(uint32(inst), 4, 6); ch {
    case CH_TEMP0:
        out = 0x800
    case CH_TOUCHY:

        if t.TouchActive {
            adcY1 := binary.LittleEndian.Uint16(FirmwareData[0x3FE00 + 0x5A:])
            scrY1 := FirmwareData[0x3FE00 + 0x5D]
            adcY2 := binary.LittleEndian.Uint16(FirmwareData[0x3FE00 + 0x60:])
            scrY2 := FirmwareData[0x3FE00 + 0x63]
            out = uint16((int(t.TouchY)-int(scrY1)+1)*int(adcY2-adcY1)/int(scrY2-scrY1) + int(adcY1))
        } else {
            out = 0xFFF
        }


    case CH_TOUCHX:
        if t.TouchActive {
            adcX1 := binary.LittleEndian.Uint16(FirmwareData[0x3FE00 + 0x58:])
			scrX1 := FirmwareData[0x3FE00 + 0x5C]
			adcX2 := binary.LittleEndian.Uint16(FirmwareData[0x3FE00 + 0x5E:])
			scrX2 := FirmwareData[0x3FE00 + 0x62]
			out = uint16((int(t.TouchX)-int(scrX1)+1)*int(adcX2-adcX1)/int(scrX2-scrX1) + int(adcX1))
        } else {
            out = 0x0
        }

    case CH_TOUCHZ1, CH_TOUCHZ2:

        out = 0x0

    case CH_AUX:

        out = 0x0

    default:
        //out = 0
        fmt.Printf("UNSETUP TOUCH SPI CHANNEL %d\n", ch)
    }

    if !conv8 {
        return []uint8{
            uint8(out >> 5),
            uint8(out << 3),
        }, STAT_DONE
    }

    out >>= 4

    return []uint8{
        uint8(out >> 1),
        uint8(out << 7),
    }, STAT_DONE
}
