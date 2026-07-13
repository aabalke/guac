package gpio

const (
	RTC_SCK = 0
	RTC_SIO = 1
	RTC_CS  = 2
)

const (
	RTC_REG_CTL = iota
	RTC_REG_DTT
	RTC_REG_TME
	RTC_REG_RST
	RTC_REG_IRQ
)

type Rtc struct {
	irq Irq

	control  uint8
	dateTime uint8
	time     uint8
}

type Irq interface {
	SetIRQ(irq uint32)
}

func NewRtc(irq Irq) *Rtc {
	return &Rtc{
		irq: irq,
	}
}
