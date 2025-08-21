package mem

import (
	"fmt"
	"log"

	"github.com/aabalke/guac/emu/nds/utils"
)

const (
    STATE_NONE   = 0
    STATE_INIT   = 1
    STATE_CMD    = 2
    STATE_PARAM  = 3

    CMD_STATUS1 = 0
    CMD_STATUS2 = 4
    CMD_DT = 2
    CMD_TIME = 6
    CMD_INT1 = 1
    CMD_INT2 = 5
    CMD_CLKADJ = 3
    CMD_FREE = 7
)

var cmdParams = map[uint8]uint8{
    CMD_STATUS1: 1,
    CMD_STATUS2: 1,
    CMD_DT: 7,
    CMD_TIME: 3,
    CMD_INT1: 1,
    CMD_INT2: 3,
    CMD_CLKADJ: 1,
    CMD_FREE: 1,

}

type Rtc struct {
	CNT uint8
    State uint8

    RegCommand uint8
    RegStatus1 uint8
    RegStatus2 uint8
    RegClkAdj uint8
    RegFree uint8
    RegYear uint8
    RegMonth uint8
    RegDay uint8
    RegDow uint8
    RegHour uint8
    RegMin uint8
    RegSec uint8
    RegAlarm1 uint32
    RegAlarm2 uint32

    Idx uint8

    Cmd uint8
    ParamRead bool
}

func (r *Rtc) Write(v uint8) {

    v &= 0b1_0111

	r.CNT = v

    switch r.State {
    case STATE_NONE:
        if v & 0b110 == 0b010 {
            r.State = STATE_INIT
            return
        }

    case STATE_INIT:

        if v & 0b100 == 0b100 {
            r.State = STATE_CMD
            return
        }

    case STATE_CMD:

        if read := v & 0b10 != 0b10; read {
            return
        }

        if r.Idx < 8 {
            r.RegCommand &^= 0b1 << r.Idx
            r.RegCommand |= (v & 1) << r.Idx
            r.Idx++
            return
        }

        r.Idx = 0

        if r.RegCommand & 0b0110 != 0b0110 {
            log.Printf("Malformed RTC Command\n")
            r.State = STATE_NONE
        }

        r.State = STATE_PARAM
        r.Cmd = uint8(utils.GetVarData(uint32(r.RegCommand), 4, 6))

    case STATE_PARAM:

        if read := v & 0b10 != 0b10; read {
            return
        }

        if r.Idx < cmdParams[r.Cmd] * 8 {

            // will need to mask writes

            switch r.Cmd {
            case CMD_STATUS1:
                r.RegStatus1 &^= 1 << r.Idx
                r.RegStatus1 |= (v & 1) << r.Idx
            case CMD_STATUS2:
                r.RegStatus2 &^= 1 << r.Idx
                r.RegStatus2 |= (v & 1) << r.Idx


            case CMD_TIME:
                if r.Idx < 8 {
                    r.RegHour &^= 1 << r.Idx
                    r.RegHour |= (v & 1) << r.Idx
                } else if r.Idx < 16 {
                    r.RegMin &^= 1 << (r.Idx - 8)
                    r.RegMin |= (v & 1) << (r.Idx - 8)
                } else {
                    r.RegSec &^= 1 << (r.Idx - 16)
                    r.RegSec |= (v & 1) << (r.Idx - 16)
                }

            case CMD_INT2:

                r.RegAlarm2 &^= 1 << r.Idx
                r.RegAlarm2 |= (uint32(v) & 1) << r.Idx


            default:
                panic(fmt.Sprintf("Unimplimented RTC Registers %d", r.Cmd))
            }

            r.Idx++
            return
        }

        r.Idx = 0
        r.State = STATE_NONE

    }

}

func (r *Rtc) Read() uint8 {

    var outBit uint8 = 0

    switch r.State {
    case STATE_PARAM:

        switch r.Cmd {
        case CMD_STATUS1:
            outBit = (r.RegStatus1 >> r.Idx) & 1
        case CMD_STATUS2:
            outBit = (r.RegStatus2 >> r.Idx) & 1
        case CMD_TIME:
            if r.Idx < 8 {
                outBit = (r.RegHour >> r.Idx) & 1
            } else if r.Idx < 16 {
                outBit = (r.RegMin >> (r.Idx - 8)) & 1
            } else {
                outBit = (r.RegSec >> (r.Idx - 16)) & 1
            }
        case CMD_INT2:
            outBit = uint8(r.RegAlarm2 >> r.Idx) & 1
        default:
            panic(fmt.Sprintf("Unimplemented RTC read for %d", r.Cmd))
        }

        r.Idx++
        if r.Idx >= cmdParams[r.Cmd]*8 {
            r.Idx = 0
            r.State = STATE_NONE
        }
    }

    // Build return: the "CNT" latch with SO bit in position 0
    return (r.CNT &^ 1) | outBit | 0b110_0000
}
