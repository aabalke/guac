package apu

import (
	"math"
	"time"

	"github.com/aabalke/guac/utils"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

type Apu struct {
	Stream utils.Stream
	Ctx    *audio.Context
	Player *audio.Player

	ToneChannel1 ToneChannel
	ToneChannel2 ToneChannel
	WaveChannel  WaveChannel
	NoiseChannel NoiseChannel

	FsCounter       uint32
	FsStep          uint8
	PendingPowerOff bool
	PendingPowerOn  bool
	Enabled         bool
	PanReg          uint8
	Master          uint8
}

func NewApu(ctx *audio.Context, bufferSize time.Duration) *Apu {
	a := &Apu{}

	if ctx != nil {

		a.Ctx = ctx

		var err error
		a.Player, err = a.Ctx.NewPlayer(&a.Stream)
		if err != nil {
			panic(err)
		}
		a.Player.SetBufferSize(bufferSize)
		a.Player.Play()
	}

	a.ToneChannel1 = ToneChannel{FsStep: &a.FsStep, Idx: 0}
	a.ToneChannel2 = ToneChannel{FsStep: &a.FsStep, Idx: 1}
	a.NoiseChannel = NoiseChannel{FsStep: &a.FsStep, Idx: 3}
	a.WaveChannel = WaveChannel{
		FsStep: &a.FsStep,
		Idx:    2,
		//WaveRam: [32]uint8{
		//	0x84, 0x40, 0x43, 0xAA, 0x2D, 0x78, 0x92, 0x3C,
		//	0x60, 0x59, 0x59, 0xB0, 0x34, 0xB8, 0x2E, 0xDA,
		//},
	}

	return a
}

func (a *Apu) ToggleMute(muted bool) {
	if a.Player != nil {
		if muted {
			a.Player.SetVolume(0)
		} else {
			a.Player.SetVolume(1)
		}
	}
}

func (a *Apu) TogglePause(paused bool) {
	if a.Player != nil {
		if paused {
			a.Player.Pause()
		} else {
			a.Player.Play()
		}
	}
}

func (a *Apu) Close() {
	if a.Player != nil {
		a.Player.Close()
	}
}

func (a *Apu) SoundClock() {
	var (
		pan  = a.PanReg
		volL = int32((a.Master>>4)&7) + 1
		volR = int32((a.Master>>0)&7) + 1

		ch1L = (pan & 0x10) != 0
		ch1R = (pan & 0x01) != 0
		ch2L = (pan & 0x20) != 0
		ch2R = (pan & 0x02) != 0
		ch3L = (pan & 0x40) != 0
		ch3R = (pan & 0x04) != 0
		ch4L = (pan & 0x80) != 0
		ch4R = (pan & 0x08) != 0

		ch1 = a.ToneChannel1.ChannelEnabled
		ch2 = a.ToneChannel2.ChannelEnabled
		ch3 = a.WaveChannel.ChannelEnabled
		ch4 = a.NoiseChannel.ChannelEnabled
	)

	psgL, psgR := int32(0), int32(0)

	if ch1 {
		ch := int32(a.ToneChannel1.GetSample(float64(a.Ctx.SampleRate())))
		if ch1L {
			psgL += ch
		}
		if ch1R {
			psgR += ch
		}
	}

	if ch2 {
		ch := int32(a.ToneChannel2.GetSample(float64(a.Ctx.SampleRate())))
		if ch2L {
			psgL += ch
		}
		if ch2R {
			psgR += ch
		}
	}

	if ch3 {
		ch := int32(a.WaveChannel.GetSample())
		if ch3L {
			psgL += ch
		}
		if ch3R {
			psgR += ch
		}
	}

	if ch4 {
		ch := int32(a.NoiseChannel.GetSample())
		if ch4L {
			psgL += ch
		}
		if ch4R {
			psgR += ch
		}
	}

	//l := clip(((psgL * volL) >> 3) >> 2)
	//r := clip(((psgR * volR) >> 3) >> 2)

	l := int16(Clip(psgL * volL))
	r := int16(Clip(psgR * volR))
	a.Stream.Write(l, r)
}

//go:inline
func Clip(v int32) int32 {
	return min(math.MaxInt16, max(math.MinInt16, v))
}

func (a *Apu) PowerOff() {
	a.ToneChannel1 = ToneChannel{FsStep: &a.FsStep, Idx: 0}
	a.ToneChannel2 = ToneChannel{FsStep: &a.FsStep, Idx: 1}
	a.WaveChannel = WaveChannel{FsStep: &a.FsStep, Idx: 2, Ram: a.WaveChannel.Ram}
	a.NoiseChannel = NoiseChannel{FsStep: &a.FsStep, Idx: 3}
	a.Master = 0
	a.PanReg = 0
	a.PendingPowerOff = true
}

func (a *Apu) PowerOn() {
	a.PendingPowerOn = true
	a.FsStep = 0
	a.FsCounter = 0
}

func (a *Apu) ClockFrameSequencer() {
	if a.PendingPowerOff {
		a.FsStep = 0
		a.PendingPowerOff = false
	}

	if a.PendingPowerOn {
		a.FsStep = 0
		a.PendingPowerOn = false
	}

	a.FsCounter++

	// frame sequencer runs at 512hz
	// length ctr at 256hz
	// sweep at 128hz
	// vol at 64hz

	if a.FsStep&1 == 0 {
		a.ToneChannel1.ClockLength()
		a.ToneChannel2.ClockLength()
		a.WaveChannel.ClockLength()
		a.NoiseChannel.ClockLength()
	}

	if a.FsStep == 2 || a.FsStep == 6 {
		a.ToneChannel1.ClockSweep()
	}

	if a.FsStep == 7 {
		a.ToneChannel1.ClockEnvelope()
		a.ToneChannel2.ClockEnvelope()
		a.NoiseChannel.ClockEnvelope()
	}

	a.FsStep = (a.FsStep + 1) & 7
}
