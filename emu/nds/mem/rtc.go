package mem

import (
	"fmt"
	"time"

	"github.com/aabalke/guac/config"
)

// interrupts not setup

type Rtc struct {
	wasCs    bool
	wasClk   bool
	wasWrite bool

	isWriteData bool
	isWriteClk  bool
	isWriteSel  bool

	data uint8
	cnt  int

	RegStatus1 uint8
	RegStatus2 uint8

	Buffer []uint8
	Idx    int

	Alarms [2]Alarm
}

type Alarm struct {
	Dow     uint8
	Hr      uint8
	MinFreq uint8
}

const (
	BIT_CLK_OUT = (1 << 1)
	BIT_SEL_OUT = (1 << 2)
	BIT_DAT_DIR = (1 << 4)
	BIT_CLK_DIR = (1 << 5)
	BIT_SEL_DIR = (1 << 6)
)

const (
	CMD_STS1 = iota
	CMD_ALM1
	CMD_DT
	CMD_CADJ
	CMD_STS2
	CMD_ALM2
	CMD_TIME
	CMD_FREE
)

func (r *Rtc) InitRtc() {
	r.RegStatus1 = 0x02
	r.RegStatus2 = 0x00
}

func (r *Rtc) Write(v uint8) {

	//r.Print(v, false)

	clk := v&BIT_CLK_OUT != 0
	cs := v&BIT_SEL_OUT != 0
	isWrite := v&BIT_DAT_DIR != 0

	if init := cs && !r.wasCs; init {
		r.data = 0
		r.cnt = 0
	}

	if send := cs && !clk && r.wasClk; send {
		r.data >>= 1
		r.cnt++
		if isWrite {
			r.data |= (v & 1) << 7
			if fullByte := r.cnt == 8; fullByte {
				r.WriteData(r.data)
				r.cnt = 0
			}
		} else {
			if fullByte := r.cnt == 8 || r.wasWrite; fullByte {
				r.data = r.ReadData()
				r.cnt = 0
			}
		}
	}

	r.wasCs = cs
	r.wasClk = clk
	r.wasWrite = isWrite
	r.isWriteClk = (v>>5)&1 != 0
	r.isWriteSel = (v>>6)&1 != 0
}

func (r *Rtc) Read() uint8 {

	v := uint8(r.data & 1)
	if r.wasClk {
		v |= BIT_CLK_OUT
	}
	if r.wasCs {
		v |= BIT_SEL_OUT
	}
	if r.wasWrite {
		v |= BIT_DAT_DIR
	}
	if r.isWriteClk {
		v |= BIT_CLK_DIR
	}
	if r.isWriteSel {
		v |= BIT_SEL_DIR
	}

	//r.Print(v, true)
	return v
}

func (r *Rtc) ReadData() uint8 {

	if r.isWriteData {
		return 0
	}

	if r.Idx >= len(r.Buffer) {
		return 0
	}

	d := r.Buffer[r.Idx]
	r.Idx++

	return d
}

func (r *Rtc) isFreqDuty() bool {
	return r.RegStatus2&(1<<2) == 0
}

var paramCnts = [8]int{1, 3, 7, 1, 1, 3, 3, 1}

func (r *Rtc) writeReg(v uint8) {
	if r.isFreqDuty() {
		paramCnts[1] = 1
	} else {
		paramCnts[1] = 3
	}

	r.Buffer = append(r.Buffer, v)
	if needParams := len(r.Buffer) != paramCnts[r.Idx]; needParams {
		return
	}
	r.isWriteData = false

	//fmt.Printf("WRITING PARAM %d V %02X\n", r.Idx, r.Buffer)

	switch r.Idx {
	case CMD_STS1:
		r.RegStatus1 = (r.RegStatus1 & 0xF0) | (v & 0xE)
	case CMD_STS2:
		r.RegStatus2 = v
	case CMD_ALM1:
		if len(r.Buffer) == 1 {
			r.Alarms[0].MinFreq = r.Buffer[0]
		} else {
			r.Alarms[0].Dow = r.Buffer[0]
			r.Alarms[0].Hr = r.Buffer[1]
			r.Alarms[0].MinFreq = r.Buffer[2]
		}
	case CMD_ALM2:
		r.Alarms[1].Dow = r.Buffer[0]
		r.Alarms[1].Hr = r.Buffer[1]
		r.Alarms[1].MinFreq = r.Buffer[2]
	default:
		panic(fmt.Sprintf("bad rtc write reg 0x%X\n", r.Idx))
	}
}

func (r *Rtc) WriteData(v uint8) {
	if r.isWriteData {
		r.writeReg(v)
		return
	}

	if invalidCmd := v&0xF != 0b0110; invalidCmd {
		return
	}

	reg := (v >> 4) & 7

	if write := v&0x80 == 0; write {
		r.isWriteData = true
		r.Buffer = nil
		r.Idx = int(reg)
		return
	}

	r.Buffer = nil
	r.Idx = 0
	switch reg {
	case CMD_STS1:
		r.Buffer = append(r.Buffer, r.RegStatus1)
		r.RegStatus1 &= 0x0F
	case CMD_STS2:
		r.Buffer = append(r.Buffer, r.RegStatus2)

	case CMD_DT, CMD_TIME:

		now := time.Now().Add(time.Hour * time.Duration(config.Conf.Nds.NdsRtc.AdditionalHours))

		var hour uint8
		if hr24 := r.RegStatus1&2 != 0; hr24 {
			hour = bcd(uint(now.Hour()))
		} else {
			hour = bcd(uint(now.Hour() % 12))
			if now.Hour() >= 12 {
				hour |= 0x40
			}
		}

		if reg == 2 {
			r.Buffer = append(r.Buffer,
				bcd(uint(now.Year()-2000)),
				bcd(uint(now.Month())),
				bcd(uint(now.Day())),
				bcd(uint(now.Weekday())),
			)
		}
		r.Buffer = append(r.Buffer,
			hour,
			bcd(uint(now.Minute())),
			bcd(uint(now.Second())),
		)

	case CMD_ALM1:
		if r.isFreqDuty() {
			r.Buffer = append(r.Buffer,
				r.Alarms[0].MinFreq,
			)

			return
		}

		r.Buffer = append(r.Buffer,
			r.Alarms[0].Dow,
			r.Alarms[0].Hr,
			r.Alarms[0].MinFreq,
		)

	case CMD_ALM2:
		r.Buffer = append(r.Buffer,
			r.Alarms[1].Dow,
			r.Alarms[1].Hr,
			r.Alarms[1].MinFreq,
		)
	default:
		panic(fmt.Sprintf("bad rtc read reg 0x%X\n", r.Idx))
	}
}

func bcd(v uint) uint8 {

	if v > 99 {
		return 0xFF
	}

	return uint8((v/10)*16 + (v % 10))
}

func (r *Rtc) Print(v uint8, read bool) {

	s := "RTC "

	if read {
		s += "R "
	} else {
		s += "W "
	}

	s += fmt.Sprintf("V %02X ", v)

	//s += fmt.Sprintf("STAT1 %02X ", r.RegStatus1)
	//s += fmt.Sprintf("STAT2 %02X ", r.RegStatus2)

	fmt.Printf("%s\n", s)
}
