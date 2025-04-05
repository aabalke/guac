package gameboy

import (
	"math"
	"math/rand"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/speaker"
)

const (
	MASTER     = 0xFF26
	PANNING    = 0xFF25
	VOLUME_VIN = 0xFF24

	// Channel 1
	CH1_SWEEP       = 0xFF10
	CH1_LENGTH_DUTY = 0xFF11
	CH1_VOLUME      = 0xFF12
	CH1_PERIOD_LO   = 0xFF13
	CH1_CONTROL     = 0xFF14

	// Channel 2
	CH2_LENGTH_DUTY = 0xFF16
	CH2_VOLUME      = 0xFF17
	CH2_PERIOD_LO   = 0xFF18
	CH2_CONTROL     = 0xFF19

	// Channel 3
	CH3_DAC       = 0xFF1A
	CH3_LENGTH    = 0xFF1B
	CH3_OUTPUT    = 0xFF1C
	CH3_PERIOD_LO = 0xFF1D
	CH3_CONTROL = 0xFF1E

	// Channel 4
	CH4_LENGTH  = 0xFF20
	CH4_VOLUME  = 0xFF21
	CH4_FREQ    = 0xFF22
	CH4_CONTROL = 0xFF23
)

type APU struct {
	MemoryBus *MemoryBus

	SampleRate float64

	Enabled bool

	Channel1 Channel
	Channel2 Channel
	Channel3 Channel
	Channel4 Channel

    WavRam   [32]float64
}

func (a *APU) Init() {

	sampleRate := beep.SampleRate(a.SampleRate)
	bufferSize := sampleRate.N(time.Second / 30)
	err := speaker.Init(sampleRate, bufferSize)
	if err != nil {
		panic(err)
	}

	mixer := &beep.Mixer{}

	a.Channel1 = Channel{
		Enabled: false,
		Apu:   a,
		WavShaper: func(i int, samples *[][2]float64, c Channel) {

			tickInCycle := c.SampleTick * 2 * math.Pi
			if math.Sin(tickInCycle) <= c.duty {
				(*samples)[i][0] = 1 * c.Volume
				(*samples)[i][1] = 1 * c.Volume
			} else {
				(*samples)[i][0] = 0
				(*samples)[i][1] = 0
			}

		},
	}
	a.Channel2 = Channel{
		Enabled: false,
		Apu:     a,
		WavShaper: func(i int, samples *[][2]float64, c Channel) {

			tickInCycle := c.SampleTick * 2 * math.Pi
			if math.Sin(tickInCycle) <= c.duty {
				(*samples)[i][0] = 1 * c.Volume
				(*samples)[i][1] = 1 * c.Volume
			} else {
				(*samples)[i][0] = 0
				(*samples)[i][1] = 0
			}

		},
	}
	a.Channel3 = Channel{
		Enabled: false,
		Apu:     a,
		WavShaper: func(i int, samples *[][2]float64, c Channel) {
            id := math.Floor(math.Mod(c.SampleTick, 1.0) * 32)
            v := c.Apu.WavRam[int(id)] * c.Volume

			(*samples)[i][0] = v
			(*samples)[i][1] = v
		},
	}
	a.Channel4 = Channel{
		Enabled: false,
		Apu:     a,
		WavShaper: func(i int, samples *[][2]float64, c Channel) {

            tickDiff := c.SampleTick - c.SamplePrev

            if tickDiff > 1 || tickDiff < -1 {
                sample := rand.Float64()*2 - 1
				(*samples)[i][0] = sample * c.Volume
				(*samples)[i][1] = sample * c.Volume

                c.Sample = sample
                c.SamplePrev = c.SampleTick
                return
            }
                
			(*samples)[i][0] = c.Sample * c.Volume
			(*samples)[i][1] = c.Sample * c.Volume

		},
	}

    mixer.Add(&a.Channel1, &a.Channel2, &a.Channel3, &a.Channel4)
	//mixer.Add(&a.Channel1)

	amplifier := &effects.Volume{
		Streamer: mixer,
		Base:     2,
		Volume:   -10,
	}

	speaker.Play(amplifier)
}

func (a *APU) Update(addr uint16, data uint8, gb *GameBoy) {

    multipler := 1.0
    if gb.DoubleSpeed {
        multipler = 2 //maybe 1/2?
    }

	Mem := &gb.MemoryBus.Memory

    if addr >= 0xFF30 && addr <= 0xFF3F {
        idx := 0
        for i := range 16 {
            a.WavRam[idx] = float64(Mem[0xFF30 + i] >> 4)
            idx++
            a.WavRam[idx] = float64(Mem[0xFF30 + i] & 0xF)
        }
    }

	switch addr {
	case MASTER:
		a.Enabled = gb.flagEnabled(data, 7)

	case CH1_CONTROL:
		if !gb.flagEnabled(data, 7) {
            return
        }
		a.Channel1.Reset()

		// volume & envelope
		a.Channel1.EnvelopeIncrease = gb.flagEnabled(Mem[CH1_VOLUME], 3)
		a.Channel1.InitialVolume = float64(Mem[CH1_VOLUME] & 0b11110000 >> 4)
		a.Channel1.EnvelopePace = Mem[CH1_VOLUME] & 0b111
		a.Channel1.Volume = a.Channel1.InitialVolume

        // sweep
        a.Channel1.SweepIncrease = gb.flagEnabled(Mem[CH1_SWEEP], 3)
        a.Channel1.SweepTime = Mem[CH1_SWEEP] & 0b1110000 >> 4
        a.Channel1.SweepPace = Mem[CH1_SWEEP] & 0b111

		// freq
		periodHi := uint16(Mem[CH1_CONTROL]&0b111) << 8
		periodLo := uint16(Mem[CH1_PERIOD_LO])
		a.Channel1.Freq = 131072.0 / (2048.0 - float64(periodHi+periodLo))

		// duration
		MAX_TIMER := 64.0 * multipler
		DIV_APU_RATE := 1.0 / 256.0 * multipler
		INIT_LENGTH_TIMER := float64(Mem[CH1_LENGTH_DUTY&0b111111])
		a.Channel1.duration = (MAX_TIMER - INIT_LENGTH_TIMER) * DIV_APU_RATE

		a.Channel1.duty = getDuty(Mem[CH1_LENGTH_DUTY] >> 6)

		a.Channel1.LengthTimer = gb.flagEnabled(data, 6)
		a.Channel1.Enabled = true

	case CH1_LENGTH_DUTY:
		a.Channel1.duty = getDuty(Mem[CH1_LENGTH_DUTY] >> 6)

	case CH2_CONTROL:
		if !gb.flagEnabled(data, 7) {
            return
        }

		a.Channel2.Reset()

		// volume & envelope
		a.Channel2.EnvelopeIncrease = gb.flagEnabled(Mem[CH2_VOLUME], 3)
		a.Channel2.InitialVolume = float64(Mem[CH2_VOLUME] & 0b11110000 >> 4)
		a.Channel2.EnvelopePace = Mem[CH2_VOLUME] & 0b111
		a.Channel2.Volume = a.Channel2.InitialVolume

		// freq
		periodHi := uint16(Mem[CH2_CONTROL]&0b111) << 8
		periodLo := uint16(Mem[CH2_PERIOD_LO])
		a.Channel2.Freq = 131072.0 / (2048.0 - float64(periodHi+periodLo))

		// duration
		MAX_TIMER := 64.0 * multipler
		DIV_APU_RATE := 1.0 / 256.0 * multipler
		INIT_LENGTH_TIMER := float64(Mem[CH2_LENGTH_DUTY&0b111111])
		a.Channel2.duration = (MAX_TIMER - INIT_LENGTH_TIMER) * DIV_APU_RATE

		a.Channel2.duty = getDuty(Mem[CH2_LENGTH_DUTY] >> 6)

		a.Channel2.LengthTimer = gb.flagEnabled(data, 6)
		a.Channel2.Enabled = true

	case CH2_LENGTH_DUTY:
		a.Channel2.duty = getDuty(Mem[CH2_LENGTH_DUTY] >> 6)


    case CH3_DAC:
        a.Channel3.Enabled = gb.flagEnabled(Mem[CH3_DAC], 7)

    case CH3_OUTPUT:

        outputLevel := Mem[CH3_OUTPUT] & 0b1100000 >> 5
        switch outputLevel {
        case 0: a.Channel3.Volume = 0
        case 1: a.Channel3.Volume = 1
        case 2: a.Channel3.Volume = 0.5
        case 3: a.Channel3.Volume = 0.25
        }

    case CH3_CONTROL:
        if !gb.flagEnabled(data, 7) {
            return
        }

        a.Channel3.Reset()

		// duration
		MAX_TIMER := 256.0 * multipler
		DIV_APU_RATE := 1.0 / 256.0 * multipler
		INIT_LENGTH_TIMER := float64(Mem[CH3_LENGTH])
		a.Channel3.duration = (MAX_TIMER - INIT_LENGTH_TIMER) * DIV_APU_RATE

		// freq
		periodHi := uint16(Mem[CH3_CONTROL]&0b111) << 8
		periodLo := uint16(Mem[CH3_PERIOD_LO])
		a.Channel3.Freq = 65536 / (2048.0 - float64(periodHi+periodLo))

        a.Channel3.duty = 1
		a.Channel3.LengthTimer = gb.flagEnabled(data, 6)
        a.Channel3.Enabled = true

    case CH4_CONTROL:
        if !gb.flagEnabled(data, 7) {
            return
        }

        a.Channel4.Enabled = false
        a.Channel4.SampleTick = 0
        a.Channel4.EnvelopeTick = 0
        a.Channel4.EnvelopePrev = 0

		// volume & envelope
		a.Channel4.EnvelopeIncrease = gb.flagEnabled(Mem[CH4_VOLUME], 3)
		a.Channel4.InitialVolume = float64((Mem[CH4_VOLUME] & 0b11110000) >> 4)
		a.Channel4.EnvelopePace = Mem[CH4_VOLUME] & 0b111
		a.Channel4.Volume = a.Channel4.InitialVolume

		MAX_TIMER := 64.0 * multipler
		DIV_APU_RATE := 1.0 / 256.0 * multipler
		INIT_LENGTH_TIMER := float64(Mem[CH4_LENGTH&0b111111])
		a.Channel4.duration = (MAX_TIMER - INIT_LENGTH_TIMER) * DIV_APU_RATE

        a.Channel4.duty = 1

		a.Channel4.LengthTimer = gb.flagEnabled(data, 6)
        a.Channel4.Enabled = true

    case CH4_FREQ:

        shiftFreq := float64(data & 0b11110000 >> 4)
        ratio := float64(data & 0b111)
        if ratio == 0 {
            ratio = 0.5
        }
        freq := 524288 / ratio / math.Pow(2, shiftFreq+1)

        a.Channel4.Freq = freq
    }
}

type Channel struct {
	Apu         *APU
	Enabled     bool
	LengthTimer bool

    Sample float64
	SampleTick float64
    SamplePrev float64

	duration   float64

    // Wave Shape
	WavShaper func(i int, samples *[][2]float64, c Channel)
	duty       float64

	// Volume and Envelope
	InitialVolume    float64
	Volume           float64
	EnvelopeIncrease bool
	EnvelopePace     uint8
	EnvelopeTick     float64
    EnvelopePrev     float64

    // Sweep and Freq
	Freq       float64
    LastFreq   float64
    SweepIncrease bool
    SweepPace uint8
    SweepTick float64
    SweepPrev float64
    SweepTime uint8
}

func (c *Channel) Reset() {
	c.Enabled = false
	c.Freq = 0
	c.SampleTick = 0
    c.EnvelopeTick = 0
    c.EnvelopePrev = 0
}

func (c *Channel) Stream(samples [][2]float64) (n int, ok bool) {

	for i := range samples {

		c.SampleTick += float64(c.Freq) / c.Apu.SampleRate

		if c.BlockAudio() {
			samples[i][0] = 0
			samples[i][1] = 0

			c.Envelope()
			c.Sweep()

			continue
		}

		c.WavShaper(i, &samples, *c)

		c.duration -= 1.0 / c.Apu.SampleRate

		c.Envelope()
		c.Sweep()
	}

	return len(samples), true
}

func (c *Channel) Err() error {
	return nil
}

func (c *Channel) BlockAudio() bool {

	if !c.Apu.Enabled || !c.Enabled || c.Freq <= 0 || c.duty == 0 || (c.LengthTimer && c.duration <= 0) {
		return true
	}

	return false
}

func getDuty(bit uint8) float64 {
	switch bit {
	case 0:
		return -0.25
	case 1:
		return 0.25
	case 2:
		return 0.5
	case 3:
		return 0.75
	}

	return 0
}

func getSweepTime(bit uint8) float64 {

    switch bit {
    case 0: return 0.0
	case 1: return 1.0 / 128
	case 2: return 2.0 / 128
	case 3: return 3.0 / 128
	case 4: return 4.0 / 128
	case 5: return 5.0 / 128
	case 6: return 6.0 / 128
	case 7: return 7.0 / 128
    }
    return 0
}

func (c *Channel) Envelope() {

	tick := 1.0 / c.Apu.SampleRate
	Rate_64Hz := 1.0 / 64.0

	c.EnvelopeTick += tick

	if c.EnvelopePace <= 0 {
		return
	}

	step := float64(c.EnvelopePace) * Rate_64Hz

	if c.EnvelopeTick - c.EnvelopePrev < step {
		return
	}

	if c.InitialVolume <= 0 {
		return
	}

	c.InitialVolume--

	if c.EnvelopeIncrease {
		c.Volume = 1 - c.InitialVolume
		c.EnvelopePrev = c.EnvelopeTick
		return
	}

	c.Volume = c.InitialVolume
    c.EnvelopePrev = c.EnvelopeTick
}

func (c *Channel) Sweep() {

	tick := 1.0 / c.Apu.SampleRate
    c.SweepTick += tick

    if c.SweepPace <= 0 {
        return
    }

    if c.SweepTick - c.SweepPrev < getSweepTime(c.SweepTime) {
        return
    }

    if c.Freq <= 0 {
        return
    }

    if c.SweepIncrease {
        c.LastFreq = float64(int(c.LastFreq + (c.LastFreq/2)) ^ int(c.SweepPace))
        c.Freq = 131072.0 / (2048.0 - c.LastFreq)
        c.SweepPrev = c.SweepTick
        return
    }

    c.LastFreq = float64(int(c.LastFreq - (c.LastFreq/2)) ^ int(c.SweepPace))
    c.Freq = 131072.0 / (2048.0 - c.LastFreq)
    c.SweepPrev = c.SweepTick
}
