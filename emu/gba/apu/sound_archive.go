package apu
//
//import (
//	"math"
//	"math/rand"
//
//	"github.com/aabalke33/guac/emu/gba/utils"
//	"github.com/gopxl/beep"
//	"github.com/gopxl/beep/effects"
//	"github.com/gopxl/beep/speaker"
//)
//
//const (
//	MASTER     = 0x84
//	PANNING    = 0x81
//	VOLUME_VIN = 0x80
//
//	// Channel 1
//	CH1_SWEEP       = 0x60
//	CH1_LENGTH_DUTY = 0x62
//	CH1_VOLUME      = 0x63
//	CH1_PERIOD_LO   = 0x64
//	CH1_CONTROL     = 0x65
//
//	// Channel 2
//	CH2_LENGTH_DUTY = 0x68
//	CH2_VOLUME      = 0x69
//	CH2_PERIOD_LO   = 0x6C
//	CH2_CONTROL     = 0x6D
//
//	// Channel 3
//	CH3_DAC       = 0x70
//	CH3_LENGTH    = 0x71
//	CH3_OUTPUT    = 0x72
//	CH3_PERIOD_LO = 0x73
//	CH3_CONTROL   = 0x74
//
//	// Channel 4
//	CH4_LENGTH  = 0x78
//	CH4_VOLUME  = 0x79
//	CH4_FREQ    = 0x7C
//	CH4_CONTROL = 0x7D
//
//    // Channel A, B
//    FIFO_A = 0xA0
//    FIFO_B = 0xA4
//
//    // Master
//    SOUNDCNT_DMA = 0x82
//    SOUNDCNT_BIAS = 0x88
//
//    FIFO_SIZE = 0x2000
//)
//
//type APU struct {
//    gba *GBA
//
//	SampleRate float64
//
//	Enabled bool
//
//	Channel1 Channel
//	Channel2 Channel
//	Channel3 Channel
//	Channel4 Channel
//    ChannelA DigitalChannel
//    ChannelB DigitalChannel
//
//    WavRam   [32]float64
//
//    Mixer      *beep.Mixer 
//}
//
//func (a *APU) Init() {
//
//	mixer := &beep.Mixer{}
//
//	a.Channel1 = Channel{
//        Idx: 0,
//		Enabled: false,
//		Apu:   a,
//		WavShaper: func(i int, samples *[][2]float64, c *Channel) {
//
//			tickInCycle := c.SampleTick * 2 * math.Pi
//			if math.Sin(tickInCycle) <= c.duty {
//				(*samples)[i][0] = 1 * c.Volume
//				(*samples)[i][1] = 1 * c.Volume
//			} else {
//				(*samples)[i][0] = 0
//				(*samples)[i][1] = 0
//			}
//
//		},
//	}
//	a.Channel2 = Channel{
//        Idx: 1,
//		Enabled: false,
//		Apu:     a,
//		WavShaper: func(i int, samples *[][2]float64, c *Channel) {
//
//			tickInCycle := c.SampleTick * 2 * math.Pi
//			if math.Sin(tickInCycle) <= c.duty {
//				(*samples)[i][0] = 1 * c.Volume
//				(*samples)[i][1] = 1 * c.Volume
//			} else {
//				(*samples)[i][0] = 0
//				(*samples)[i][1] = 0
//			}
//
//		},
//	}
//	a.Channel3 = Channel{
//        Idx: 2,
//		Enabled: false,
//		Apu:     a,
//		WavShaper: func(i int, samples *[][2]float64, c *Channel) {
//            id := math.Floor(math.Mod(c.SampleTick, 1.0) * 32)
//            v := c.Apu.WavRam[int(id)] * c.Volume
//
//			(*samples)[i][0] = v
//			(*samples)[i][1] = v
//		},
//	}
//	a.Channel4 = Channel{
//        Idx: 3,
//		Enabled: false,
//		Apu:     a,
//		WavShaper: func(i int, samples *[][2]float64, c *Channel) {
//
//            tickDiff := c.SampleTick - c.SamplePrev
//
//            if tickDiff > 1 || tickDiff < -1 {
//                sample := rand.Float64()*2 - 1
//				(*samples)[i][0] = sample * c.Volume
//				(*samples)[i][1] = sample * c.Volume
//
//                c.Sample = sample
//                c.SamplePrev = c.SampleTick
//                return
//            }
//                
//			(*samples)[i][0] = c.Sample * c.Volume
//			(*samples)[i][1] = c.Sample * c.Volume
//
//		},
//	}
//
//    digital := DigitalChannel{
//		Enabled: false,
//		Apu:     a,
//		WavShaper: func(i int, samples *[][2]float64, c *DigitalChannel) {
//
//            c.History[0] = c.History[1]
//            c.History[1] = c.History[2]
//            c.History[2] = c.History[3]
//            c.History[3] = 0
//            if utils.BitEnabled(c.Control, 7) {
//                c.History[3] = float64((c.Fifo[c.FifoReadSize]))
//            }
//
//            a := cubicInterpolate(c.History, c.Fraction)
//
//            c.Fraction += c.Ratio
//
//            if c.Fraction >= 1 {
//                c.Fraction -= float64((c.Fraction))
//
//                if (c.FifoReadSize + 1) % FIFO_SIZE != c.FifoWriteSize {
//                    c.FifoReadSize = (c.FifoReadSize + 1) % FIFO_SIZE
//                }
//            }
//
//            left := 0
//            right := 0
//
//            soundcnt := c.Apu.gba.Mem.ReadIODirect(0x82, 2)
//
//            if utils.BitEnabled(soundcnt, 8) {
//                right = clamp(int(a), -512, 511)
//            }
//
//            if utils.BitEnabled(soundcnt, 9) {
//                left = clamp(int(a), -512, 511)
//            }
//
//            // will need 12 and 13 for Channel B
//
//            nLeft := (float64(left) + 512) / 1023
//            nRight := (float64(right) + 512) / 1023
//
//			(*samples)[i][0] = nLeft
//			(*samples)[i][1] = nRight
//		},
//    }
//
//    digital.Idx = 0
//	a.ChannelA = digital
//    //digital.Idx = 1
//	//a.ChannelB = digital update clamp bits
//
//    //mixer.Add(&a.Channel1, &a.Channel2, &a.Channel3, &a.Channel4, &a.ChannelA)
//    //mixer.Add(&a.ChannelA)
//    //mixer.Add(&a.Channel1, &a.Channel2, &a.Channel3, &a.Channel4)
//    //a.ChannelA.InitCallback()
//
//	amplifier := &effects.Volume{
//		Streamer: mixer,
//		Base:     2,
//		//Volume:   -10,
//		Volume:   -5,
//	}
//
//	speaker.Play(amplifier)
//}
//
//func (a *APU) Close() {
//}
//
//func (a *APU) Update(addr uint16, data uint8) {
//
//    gba := a.gba
//    io := a.gba.Mem.IO
//
//    if gba.Paused || gba.Muted {
//        return
//    }
//
//    if addr >= 0x90 && addr < 0xA0 {
//        idx := 0
//        for i := range 16 {
//            a.WavRam[idx] = float64(io[0x90 + i] >> 4)
//            idx++
//            a.WavRam[idx] = float64(io[0x90 + i] & 0xF)
//        }
//    }
//
//	switch addr {
//	case MASTER:
//        a.Enabled = utils.BitEnabled(uint32(data), 7)
//
//	case CH1_CONTROL:
//        if !utils.BitEnabled(uint32(data), 7) {
//            return
//        }
//		a.Channel1.Reset()
//
//		// volume & envelope
//		a.Channel1.EnvelopeIncrease = utils.BitEnabled(uint32(io[CH1_VOLUME]), 3)
//		a.Channel1.InitialVolume = float64(io[CH1_VOLUME] & 0b11110000 >> 4)
//		a.Channel1.EnvelopePace = io[CH1_VOLUME] & 0b111
//		a.Channel1.Volume = a.Channel1.InitialVolume
//
//        // sweep
//        a.Channel1.SweepIncrease = utils.BitEnabled(uint32(io[CH1_SWEEP]), 3)
//        a.Channel1.SweepTime = io[CH1_SWEEP] & 0b1110000 >> 4
//        a.Channel1.SweepPace = io[CH1_SWEEP] & 0b111
//
//		// freq
//		periodHi := uint16(io[CH1_CONTROL]&0b111) << 8
//		periodLo := uint16(io[CH1_PERIOD_LO])
//		a.Channel1.Freq = 131072.0 / (2048.0 - float64(periodHi+periodLo))
//
//		// duration
//		MAX_TIMER := 64.0
//		DIV_APU_RATE := 1.0 / 256.0
//		INIT_LENGTH_TIMER := float64(io[CH1_LENGTH_DUTY&0b111111])
//		a.Channel1.duration = (MAX_TIMER - INIT_LENGTH_TIMER) * DIV_APU_RATE
//
//		a.Channel1.duty = getDuty(io[CH1_LENGTH_DUTY] >> 6)
//
//		a.Channel1.LengthTimer = utils.BitEnabled(uint32(data), 6)
//		a.Channel1.Enabled = true
//
//	case CH1_LENGTH_DUTY:
//		a.Channel1.duty = getDuty(io[CH1_LENGTH_DUTY] >> 6)
//
//	case CH2_CONTROL:
//        if !utils.BitEnabled(uint32(data), 7) {
//            return
//        }
//
//		a.Channel2.Reset()
//
//		// volume & envelope
//		a.Channel2.EnvelopeIncrease = utils.BitEnabled(uint32(io[CH2_VOLUME]), 3)
//		a.Channel2.InitialVolume = float64(io[CH2_VOLUME] & 0b11110000 >> 4)
//		a.Channel2.EnvelopePace = io[CH2_VOLUME] & 0b111
//		a.Channel2.Volume = a.Channel2.InitialVolume
//
//		// freq
//		periodHi := uint16(io[CH2_CONTROL]&0b111) << 8
//		periodLo := uint16(io[CH2_PERIOD_LO])
//		a.Channel2.Freq = 131072.0 / (2048.0 - float64(periodHi+periodLo))
//
//		// duration
//		MAX_TIMER := 64.0
//		DIV_APU_RATE := 1.0 / 256.0
//		INIT_LENGTH_TIMER := float64(io[CH2_LENGTH_DUTY&0b111111])
//		a.Channel2.duration = (MAX_TIMER - INIT_LENGTH_TIMER) * DIV_APU_RATE
//
//		a.Channel2.duty = getDuty(io[CH2_LENGTH_DUTY] >> 6)
//
//		a.Channel2.LengthTimer = utils.BitEnabled(uint32(data), 6)
//		a.Channel2.Enabled = true
//
//	case CH2_LENGTH_DUTY:
//		a.Channel2.duty = getDuty(io[CH2_LENGTH_DUTY] >> 6)
//
//
//    case CH3_DAC:
//        a.Channel3.Enabled = utils.BitEnabled(uint32(io[CH3_DAC]), 7)
//
//    case CH3_OUTPUT:
//
//        outputLevel := io[CH3_OUTPUT] & 0b1100000 >> 5
//        switch outputLevel {
//        case 0: a.Channel3.Volume = 0
//        case 1: a.Channel3.Volume = 1
//        case 2: a.Channel3.Volume = 0.5
//        case 3: a.Channel3.Volume = 0.25
//        }
//
//    case CH3_CONTROL:
//        if !utils.BitEnabled(uint32(data), 7) {
//            return
//        }
//
//        a.Channel3.Reset()
//
//		// duration
//		MAX_TIMER := 256.0
//		DIV_APU_RATE := 1.0 / 256.0
//		INIT_LENGTH_TIMER := float64(io[CH3_LENGTH])
//		a.Channel3.duration = (MAX_TIMER - INIT_LENGTH_TIMER) * DIV_APU_RATE
//
//		// freq
//		periodHi := uint16(io[CH3_CONTROL]&0b111) << 8
//		periodLo := uint16(io[CH3_PERIOD_LO])
//		a.Channel3.Freq = 65536 / (2048.0 - float64(periodHi+periodLo))
//
//        a.Channel3.duty = 1
//		a.Channel3.LengthTimer = utils.BitEnabled(uint32(data), 6)
//        a.Channel3.Enabled = true
//
//    case CH4_CONTROL:
//        if !utils.BitEnabled(uint32(data), 7) {
//            return
//        }
//
//        a.Channel4.Enabled = false
//        a.Channel4.SampleTick = 0
//        a.Channel4.EnvelopeTick = 0
//        a.Channel4.EnvelopePrev = 0
//
//		// volume & envelope
//		a.Channel4.EnvelopeIncrease = utils.BitEnabled(uint32(io[CH4_VOLUME]), 3)
//		a.Channel4.InitialVolume = float64((io[CH4_VOLUME] & 0b11110000) >> 4)
//		a.Channel4.EnvelopePace = io[CH4_VOLUME] & 0b111
//		a.Channel4.Volume = a.Channel4.InitialVolume
//
//		MAX_TIMER := 64.0
//		DIV_APU_RATE := 1.0 / 256.0
//		INIT_LENGTH_TIMER := float64(io[CH4_LENGTH&0b111111])
//		a.Channel4.duration = (MAX_TIMER - INIT_LENGTH_TIMER) * DIV_APU_RATE
//
//        a.Channel4.duty = 1
//
//		a.Channel4.LengthTimer = utils.BitEnabled(uint32(data), 6)
//        a.Channel4.Enabled = true
//
//    case CH4_FREQ:
//
//        shiftFreq := float64(data & 0b11110000 >> 4)
//        ratio := float64(data & 0b111)
//        if ratio == 0 {
//            ratio = 0.5
//        }
//        freq := 524288 / ratio / math.Pow(2, shiftFreq+1)
//
//        a.Channel4.Freq = freq
//    }
//}
//
//type Channel struct {
//    Idx int
//	Apu         *APU
//	Enabled     bool
//	LengthTimer bool
//
//    Sample float64
//	SampleTick float64
//    SamplePrev float64
//
//	duration   float64
//
//    // Wave Shape
//	WavShaper func(i int, samples *[][2]float64, c *Channel)
//	duty       float64
//
//	// Volume and Envelope
//	InitialVolume    float64
//	Volume           float64
//	EnvelopeIncrease bool
//	EnvelopePace     uint8
//	EnvelopeTick     float64
//    EnvelopePrev     float64
//
//    // Sweep and Freq
//	Freq       float64
//    LastFreq   float64
//    SweepIncrease bool
//    SweepPace uint8
//    SweepTick float64
//    SweepPrev float64
//    SweepTime uint8
//}
//
//func (c *Channel) Reset() {
//	c.Enabled = false
//	c.Freq = 0
//	c.SampleTick = 0
//    c.EnvelopeTick = 0
//    c.EnvelopePrev = 0
//}
//
//func (c *Channel) Stream(samples [][2]float64) (n int, ok bool) {
//
//	for i := range samples {
//
//		c.SampleTick += float64(c.Freq) / c.Apu.SampleRate
//
//		if c.BlockAudio() {
//            //c.setChannelBit(false)
//			samples[i][0] = 0
//			samples[i][1] = 0
//
//			c.Envelope()
//			c.Sweep()
//
//
//			continue
//		}
//
//        //c.setChannelBit(true)
//
//		c.WavShaper(i, &samples, c)
//
//		c.duration -= 1.0 / c.Apu.SampleRate
//
//		c.Envelope()
//		c.Sweep()
//
//	}
//
//	return len(samples), true
//}
//
////func (c *Channel) setChannelBit(on bool) {
////
////    if on {
////        c.Apu.gba.Mem.IO[0x84] |= (1 << c.Idx)
////    }
////
////    c.Apu.gba.Mem.IO[0x84] &^= (1 << c.Idx)
////}
//
//func (c *Channel) Err() error {
//	return nil
//}
//
//func (c *Channel) BlockAudio() bool {
//
//    if  c.Apu.gba.Muted ||
//        c.Apu.gba.Paused ||
//        !c.Apu.Enabled ||
//        !c.Enabled ||
//        c.Freq <= 0 ||
//        c.duty == 0 ||
//        (c.LengthTimer && c.duration <= 0) {
//		return true
//	}
//
//	return false
//}
//
//func getDuty(bit uint8) float64 {
//	switch bit {
//	case 0:
//		return -0.25
//	case 1:
//		return 0.25
//	case 2:
//		return 0.5
//	case 3:
//		return 0.75
//	}
//
//	return 0
//}
//
//func getSweepTime(bit uint8) float64 {
//
//    switch bit {
//    case 0: return 0.0
//	case 1: return 1.0 / 128
//	case 2: return 2.0 / 128
//	case 3: return 3.0 / 128
//	case 4: return 4.0 / 128
//	case 5: return 5.0 / 128
//	case 6: return 6.0 / 128
//	case 7: return 7.0 / 128
//    }
//    return 0
//}
//
//func (c *Channel) Envelope() {
//
//	tick := 1.0 / c.Apu.SampleRate
//	Rate_64Hz := 1.0 / 64.0
//
//	c.EnvelopeTick += tick
//
//	if c.EnvelopePace <= 0 {
//		return
//	}
//
//	step := float64(c.EnvelopePace) * Rate_64Hz
//
//	if c.EnvelopeTick - c.EnvelopePrev < step {
//		return
//	}
//
//	if c.InitialVolume <= 0 {
//		return
//	}
//
//	c.InitialVolume--
//
//	if c.EnvelopeIncrease {
//		c.Volume = 1 - c.InitialVolume
//		c.EnvelopePrev = c.EnvelopeTick
//		return
//	}
//
//	c.Volume = c.InitialVolume
//    c.EnvelopePrev = c.EnvelopeTick
//}
//
//func (c *Channel) Sweep() {
//
//	tick := 1.0 / c.Apu.SampleRate
//    c.SweepTick += tick
//
//    if c.SweepPace <= 0 {
//        return
//    }
//
//    if c.SweepTick - c.SweepPrev < getSweepTime(c.SweepTime) {
//        return
//    }
//
//    if c.Freq <= 0 {
//        return
//    }
//
//    if c.SweepIncrease {
//        c.LastFreq = float64(int(c.LastFreq + (c.LastFreq/2)) ^ int(c.SweepPace))
//        c.Freq = 131072.0 / (2048.0 - c.LastFreq)
//        c.SweepPrev = c.SweepTick
//        return
//    }
//
//    c.LastFreq = float64(int(c.LastFreq - (c.LastFreq/2)) ^ int(c.SweepPace))
//    c.Freq = 131072.0 / (2048.0 - c.LastFreq)
//    c.SweepPrev = c.SweepTick
//}
//
//type DigitalChannel struct {
//    Apu *APU
//    Enabled bool
//	WavShaper func(i int, samples *[][2]float64, c *DigitalChannel)
//    Ticks uint32
//    Refill bool
//    Fifo [FIFO_SIZE]uint32
//    FifoWriteSize, FifoReadSize uint32
//
//    SampleTick float64
//    Freq uint32
//    Duration float64
//    Volume float64
//
//    History [4]float64
//    Fraction float64
//    Timer bool
//    Idx uint32
//    Control, Reload uint32
//    Ratio float64
//}
//
//func (c *DigitalChannel) Write(sample uint32) {
//
//    //c.Fifo[c.FifoWriteSize] = sample
//    //c.FifoWriteSize = (c.FifoWriteSize + 4) % FIFO_SIZE
//    c.Fifo[c.FifoWriteSize] = sample
//    c.FifoWriteSize = (c.FifoWriteSize + 1) % FIFO_SIZE
//}
//
//func (c *DigitalChannel) Reset() {
//	c.Enabled = false
//	c.SampleTick = 0
//}
//
//func (c *DigitalChannel) Stream(samples [][2]float64) (n int, ok bool) {
//
//    c.InitCallback(samples)
//
//	for i := range samples {
//
//		//c.SampleTick += float64(c.Freq) / c.Apu.SampleRate
//		//c.SampleTick += float64(440) / c.Apu.SampleRate
//
//		//if c.BlockAudio() {
//		//	samples[i][0] = 0
//		//	samples[i][1] = 0
//		//	continue
//		//}
//
//		c.WavShaper(i, &samples, c)
//
//		//c.Duration -= 1.0 / c.Apu.SampleRate
//	}
//
//	return len(samples), true
//}
//func (c *DigitalChannel) Err() error {
//	return nil
//}
//
//func (c *DigitalChannel) BlockAudio() bool {
//
//    if  c.Apu.gba.Muted ||
//        c.Apu.gba.Paused ||
//        !c.Apu.Enabled ||
//        !c.Enabled {
//		return true
//	}
//
//	return false
//}
//
//func (c *DigitalChannel) InitCallback(samples[][2]float64) {
//
//    gba := c.Apu.gba
//
//    switch c.Idx {
//    case 0:
//        c.Timer = utils.BitEnabled(gba.Mem.ReadIODirect(0x82, 2), 10)
//    case 1:
//        c.Timer = utils.BitEnabled(gba.Mem.ReadIODirect(0x82, 2), 14)
//    default:
//        panic("SET ACCURATE IDXS ON CHANNEL A AND B")
//    }
//
//    if c.Timer {
//        c.Control = gba.Timers[1].CNT
//        c.Reload = gba.Timers[1].SavedInitialValue
//    } else {
//        c.Control = gba.Timers[0].CNT
//        c.Reload = gba.Timers[0].SavedInitialValue
//    }
//
//    srcRate := float64(16777216) / (65536 - float64(c.Reload))
//
//    //if c.Reload >= 65536 || c.Reload == 0 {
//    //    c.Ratio = 0
//    //}
//
//    dstRate := c.Apu.SampleRate
//    c.Ratio = srcRate / dstRate
//    c.Fraction = 0
//    c.History = [4]float64{}
//}
//
//func cubicInterpolate(history [4]float64, mu float64) float64 {
//    a := history[3] - history[2] - history[0] + history[1]
//    b := history[0] - history[1] - a
//    c := history[2] - history[0]
//    d := history[1]
//    return a * mu * mu * mu + b * mu * mu + c * mu + d;
//}
//
//func clamp(x, minimum, maximum int) int {
//    if x < minimum {
//        x = minimum
//    }
//
//    if x > maximum {
//        x= maximum
//    }
//    return x
//}
