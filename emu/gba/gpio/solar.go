package gpio

const (
	SOLAR_CLK = 0
	SOLAR_RST = 1
	SOLAR_FLG = 3
)

type Solar struct {
	Level   uint8
	counter uint8
	prevClk bool
}

func NewSolar(level uint8) *Solar {
	s := &Solar{}
	s.SetLevel(level)
	return s
}

func (s *Solar) Read() uint8 {
	if s.counter > s.Level {
		return 1 << SOLAR_FLG
	}

	return 0
}

func (s *Solar) Write(v uint8) {
	reset := (v>>SOLAR_RST)&1 != 0
	clk := (v>>SOLAR_CLK)&1 != 0

	switch {
	case reset:
		s.counter = 0
	case s.prevClk && !clk:
		s.counter++
	}

	s.prevClk = clk
}

func (s *Solar) SetLevel(level uint8) {
	// input level should be 0 - 100
	s.Level = uint8(float32(100-level)*1.6) + 80
}
