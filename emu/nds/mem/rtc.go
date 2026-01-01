package mem

import (
	"time"
)

// interrupts not setup

type Rtc struct {
	wasCs    bool
	wasClk   bool
	wasWrite bool

	data uint8
	cnt  int

	RegStatus1 uint8
	RegStatus2 uint8

	IsWriting bool
	Buffer    []uint8
	Idx       int

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

	CMD_STS1 = 0
	CMD_ALM1 = 1
	CMD_DT   = 2
	CMD_CADJ = 3
	CMD_STS2 = 4
	CMD_ALM2 = 5
	CMD_TIME = 6
	CMD_FREE = 7
)

func (r *Rtc) InitRtc() {
	//r.RegStatus1 = 0x80
	r.RegStatus1 = 0x00
	r.RegStatus2 = 0x00
}

func (r *Rtc) Write(v uint8) {

	clk := v&BIT_CLK_OUT != 0
	cs := v&BIT_SEL_OUT != 0
	isWrite := v&BIT_DAT_DIR != 0

	if init := cs && !r.wasCs; init {
		r.data = 0
		if isWrite {
			r.cnt = 8
		} else {
			r.cnt = 0
		}
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
}

func (r *Rtc) Read() uint8 {
	v := uint8(0b0110_0000)
	if r.wasClk {
		v |= BIT_CLK_OUT
	}
	if r.wasCs {
		v |= BIT_SEL_OUT
	}
	if r.wasWrite {
		v |= BIT_DAT_DIR
	}
	v |= (r.data & 1)

	return v
}

func (r *Rtc) ReadData() uint8 {

	if r.IsWriting {
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
	r.IsWriting = false

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
	}
}

func (r *Rtc) WriteData(v uint8) {
	if r.IsWriting {
		r.writeReg(v)
		return
	}

	if invalidCmd := v&0xF != 0b0110; invalidCmd {
		return
	}

	reg := (v >> 4) & 7

	if write := v&0x80 == 0; write {
		r.IsWriting = true
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
		now := time.Now()

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
	}
}

func bcd(v uint) uint8 {

	if v > 99 {
		return 0xFF
	}

	return uint8((v/10)*16 + (v % 10))
}
