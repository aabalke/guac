package spi

import (
	"encoding/binary"
	"fmt"
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
	Firmware *Firmware

	Control uint8

	// Temp0 uint16
	// Temp1 uint16
	// Aux uint16

	TouchX uint16
	TouchY uint16

	IrqEnabled bool

	Input2 *uint8
}

func NewTsc(input2 *uint8) *Tsc {
	return &Tsc{
		Input2: input2,
	}
}

func (t *Tsc) Transfer(data []uint8) (reply []uint8, stat uint8) {
	inst := data[0]

	if invalidStart := inst&0x80 == 0; invalidStart {
		return nil, STAT_DONE
	}

	var (
		out   uint16
		conv8 = (inst>>3)&1 != 0
	)

	switch ch := (inst >> 4) & 7; ch {
	case CH_TEMP0:
		out = 0x800
	case CH_TOUCHY:

		if pressed := *t.Input2&0x40 == 0; pressed {
			adcY1 := int(binary.LittleEndian.Uint16(t.Firmware.Data[0x3FE00+0x5A:]))
			scrY1 := int(t.Firmware.Data[0x3FE00+0x5D])
			adcY2 := int(binary.LittleEndian.Uint16(t.Firmware.Data[0x3FE00+0x60:]))
			scrY2 := int(t.Firmware.Data[0x3FE00+0x63])
			out = uint16((int(t.TouchY)-scrY1+1)*(adcY2-adcY1)/(scrY2-scrY1) + adcY1)
		} else {
			out = 0xFFF
		}

	case CH_TOUCHX:
		if pressed := *t.Input2&0x40 == 0; pressed {
			adcX1 := int(binary.LittleEndian.Uint16(t.Firmware.Data[0x3FE00+0x58:]))
			scrX1 := int(t.Firmware.Data[0x3FE00+0x5C])
			adcX2 := int(binary.LittleEndian.Uint16(t.Firmware.Data[0x3FE00+0x5E:]))
			scrX2 := int(t.Firmware.Data[0x3FE00+0x62])
			out = uint16((int(t.TouchX)-scrX1+1)*(adcX2-adcX1)/(scrX2-scrX1) + adcX1)
		} else {
			out = 0x0
		}

	case CH_TOUCHZ1, CH_TOUCHZ2, CH_AUX:
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
