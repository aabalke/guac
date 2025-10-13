package cpu


const (
    IRQ_VBL  = 0
    IRQ_HBL  = 1
    IRQ_VCT  = 2
    IRQ_TMR0 = 3
    IRQ_TMR1 = 4
    IRQ_TMR2 = 5
    IRQ_TMR3 = 6
    IRQ_RTC  = 7 // arm7 only
    IRQ_DMA0 = 8
    IRQ_DMA1 = 9
    IRQ_DMA2 = 10
    IRQ_DMA3 = 11
    IRQ_KEY  = 12
    IRQ_GBA  = 13

    IRQ_IPC_SYNC = 16
    IRQ_IPC_SEND_FIFO = 17
    IRQ_IPC_RECV_FIFO = 18
    IRQ_CARD_TRANS_COMPLETE = 19
    IRQ_CARD_IREQ_MC = 20
    IRQ_GEO_CMD_FIFO = 21 // arm9 only
    IRQ_SCREEN_UNFOLDING = 22 // arm7 only
    IRQ_SPI_BUS = 23 // arm7 only
    IRQ_WIFI = 24 // arm7 only
)

type Irq struct {
	IF, IE  uint32
	IME     bool
	IdleIrq uint32

    IsArm9 bool
}

func (s *Irq) WriteIME(v uint8) {
	s.IME = v&1 == 1
}

func (s *Irq) ReadIME() uint8 {

	if s.IME {
		return 1
	}

	return 0
}

func (s *Irq) ReadIE(byte uint8) uint8 {
	return uint8(s.IE >> (8 * byte))
}

func (s *Irq) ReadIF(byte uint8) uint8 {
	return uint8(s.IF >> (8 * byte))
}

func (s *Irq) WriteIE(v uint8, byte uint8) {
	s.IE &^= 0xFF << (8 * byte)
	s.IE |= (uint32(v) << (8 * byte))
}

func (s *Irq) WriteIF(v uint8, byte uint8) {
	s.IF &^= uint32(v) << (8 * byte)
}

func (s *Irq) SetIRQ(irq uint32) {
	s.IF |= (1 << irq)
}
