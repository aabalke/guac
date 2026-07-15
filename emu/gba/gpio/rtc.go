package gpio

import (
	"fmt"
	"time"
)

const (
	RTC_SCK = 0
	RTC_SIO = 1
	RTC_CS  = 2
)

const (
	RTC_RESET = iota
	RTC_STATUS
	RTC_DATE
	RTC_TIME
	RTC_ALARM
	RTC_TEST_START // force irq according to gbatek
)

const (
	RTC_STAT_CMD = iota
	RTC_STAT_WR
	RTC_STAT_RD
)

var PARAM_CNT = [...]uint8{
	0, // reset
	1, // status
	7, // date
	3, // time
	0, // alarm
}

type Rtc struct {
	irq     Irq
	cnt     int
	prevClk bool
	prevCs  bool

	status   uint8
	selected uint8

	buf []uint8
	out uint8
	in  uint8

	ctrl struct {
		intfe  bool
		intme  bool
		intae  bool
		hour24 bool
		power  bool
	}
}

type Irq interface {
	SetIRQ(irq uint32)
}

func NewRtc(irq Irq) *Rtc {
	r := &Rtc{
		irq: irq,
	}

	r.ctrl.hour24 = true

	return r
}

func (r *Rtc) Read() uint8 {
	v := (r.out & 1) << 1

	if r.prevClk {
		v |= 1
	}

	if r.prevCs {
		v |= 1 << 2
	}

	return v
}

func (r *Rtc) Write(v uint8) {
	var (
		clk = (v>>RTC_SCK)&1 != 0
		sio = (v >> RTC_SIO) & 1
		cs  = (v>>RTC_CS)&1 != 0
	)

	if init := cs && !r.prevCs; init {
		r.in = 0
		r.out = 0
		r.cnt = 0
		r.status = RTC_STAT_CMD
	}

	if send := cs && !clk && r.prevClk; send {
		r.WriteData(sio)
	}

	r.prevClk = clk
	r.prevCs = cs
}

func (r *Rtc) WriteData(sio uint8) {
	switch r.status {
	case RTC_STAT_CMD:

		r.in >>= 1
		r.in |= sio << 7
		r.cnt++
		if r.cnt < 8 {
			return
		}

		r.cnt = 0

		v := r.in

		switch {
		case (v >> 4) == 0x6:

			// flip msb -> lsb if wrong direction

			v = (v << 4) | (v >> 4)
			v = ((v & 0x33) << 2) | ((v & 0xCC) >> 2)
			v = ((v & 0x55) << 1) | ((v & 0xAA) >> 1)

		case v&0xF != 0x6:
			return
		}

		v >>= 4

		read := v>>3 != 0
		r.selected = ((v >> 0) & 1) << 2
		r.selected |= ((v >> 1) & 1) << 1
		r.selected |= ((v >> 2) & 1) << 0

		if read {
			r.status = RTC_STAT_RD

			switch r.selected {
			case RTC_RESET:

				// registers should be reset, registers should also be reset on invalid values in registers
				// does reseting registers actually matter if emu provides exact value

				r.ctrl = struct {
					intfe  bool
					intme  bool
					intae  bool
					hour24 bool
					power  bool
				}{}

			case RTC_STATUS:

				v = 0

				if r.ctrl.intfe {
					v |= 1 << 1
				}

				if r.ctrl.intme {
					v |= 1 << 3
				}

				if r.ctrl.intae {
					v |= 1 << 5
				}

				if r.ctrl.hour24 {
					v |= 1 << 6
				}

				if r.ctrl.power {
					v |= 1 << 7
				}

				r.buf = []uint8{v}

			case RTC_DATE:

				now := time.Now()

				hour := convertDecimalToBcd(uint8(now.Hour()))

				if !r.ctrl.hour24 {
					hour %= 12
				}

				r.buf = []uint8{
					convertDecimalToBcd(uint8(now.Year() - 2000)),
					convertDecimalToBcd(uint8(now.Month())),
					convertDecimalToBcd(uint8(now.Day())),
					convertDecimalToBcd(uint8(now.Weekday())),
					hour,
					convertDecimalToBcd(uint8(now.Minute())),
					convertDecimalToBcd(uint8(now.Second())),
				}

			case RTC_TIME:

				now := time.Now()

				hour := convertDecimalToBcd(uint8(now.Hour()))
				if !r.ctrl.hour24 {
					hour %= 12
				}

				r.buf = []uint8{
					hour,
					convertDecimalToBcd(uint8(now.Minute())),
					convertDecimalToBcd(uint8(now.Second())),
				}

			case RTC_TEST_START:
				r.irq.SetIRQ(13)

			}
		} else {
			if PARAM_CNT[r.selected] > 0 {
				r.status = RTC_STAT_WR
			}
		}

	case RTC_STAT_WR:

		r.in >>= 1
		r.in |= sio << 7
		r.cnt++
		if r.cnt < 8 {
			return
		}

		r.cnt = 0
		v := r.in

		switch r.selected {
		case RTC_STATUS:

			r.ctrl.intfe = (v>>1)&1 != 0
			r.ctrl.intme = (v>>3)&1 != 0
			r.ctrl.intae = (v>>5)&1 != 0
			r.ctrl.hour24 = (v>>6)&1 != 0

			if r.ctrl.intme {
				fmt.Printf("Warning. Write to unsetup RTC per minute irq\n")
			}
		}

		r.status = RTC_STAT_CMD

	case RTC_STAT_RD:

		if len(r.buf) == 0 {
			r.status = RTC_STAT_CMD
			return
		}

		r.out = r.buf[0] & 1
		r.buf[0] >>= 1
		r.cnt++

		if r.cnt < 8 {
			return
		}

		r.cnt = 0

		if len(r.buf) > 1 {
			r.buf = r.buf[1:]
		} else {
			r.status = RTC_STAT_CMD
		}
	}
}

func convertDecimalToBcd(a uint8) uint8 {
	b, c := uint8(0), uint8(1)

	for a > 0 {
		b += (a % 10) * c
		c *= 16
		a /= 10
	}

	return b
}
