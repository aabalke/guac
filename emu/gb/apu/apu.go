package apu

import (
	"math"
	"time"

	"github.com/aabalke/guac/utils"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

type Apu struct {
	stream utils.Stream
	Ctx    *audio.Context
	player *audio.Player

	ToneChannel1 ToneChannel
	ToneChannel2 ToneChannel
	WaveChannel  WaveChannel
	NoiseChannel NoiseChannel

	fsCounter       uint32
	fsStep          uint8
	Enabled         bool
	pendingPowerOff bool
	pendingPowerOn  bool
	PanReg          uint8
	Master          uint8
}

type Channel interface {
	GetSample() int8
}

func NewApu(ctx *audio.Context, bufferSize time.Duration) *Apu {
	a := &Apu{}

	if ctx != nil {

		a.Ctx = ctx

		var err error
		a.player, err = a.Ctx.NewPlayer(&a.stream)
		if err != nil {
			panic(err)
		}
		a.player.SetBufferSize(bufferSize)
		a.player.Play()
	}

	a.ToneChannel1 = ToneChannel{Apu: a, Idx: 0}
	a.ToneChannel2 = ToneChannel{Apu: a, Idx: 1}
	a.WaveChannel = WaveChannel{
		Apu: a,
		Idx: 2,
		//WaveRam: [32]uint8{
		//	0x84, 0x40, 0x43, 0xAA, 0x2D, 0x78, 0x92, 0x3C,
		//	0x60, 0x59, 0x59, 0xB0, 0x34, 0xB8, 0x2E, 0xDA,
		//},
	}
	a.NoiseChannel = NoiseChannel{Apu: a, Idx: 3}

	return a
}

func (a *Apu) ToggleMute(muted bool) {
	if a.player != nil {
		if muted {
			a.player.SetVolume(0)
		} else {
			a.player.SetVolume(1)
		}
	}
}

func (a *Apu) TogglePause(paused bool) {
	if a.player != nil {
		if paused {
			a.player.Pause()
		} else {
			a.player.Play()
		}
	}
}

func (a *Apu) Close() {
	if a.player != nil {
		a.player.Close()
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
		ch := int32(a.ToneChannel1.GetSample())
		if ch1L {
			psgL += ch
		}
		if ch1R {
			psgR += ch
		}
	}

	if ch2 {
		ch := int32(a.ToneChannel2.GetSample())
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

	l := int16(clip(psgL * volL))
	r := int16(clip(psgR * volR))
	a.stream.Write(l, r)
}

//go:inline
func clip(v int32) int32 {
	return min(math.MaxInt16, max(math.MinInt16, v))
}

func (a *Apu) PowerOff() {
	a.ToneChannel1 = ToneChannel{Idx: 0, Apu: a}
	a.ToneChannel2 = ToneChannel{Idx: 1, Apu: a}
	a.WaveChannel = WaveChannel{Idx: 2, Apu: a, Ram: a.WaveChannel.Ram}
	a.NoiseChannel = NoiseChannel{Idx: 3, Apu: a}
	a.Master = 0
	a.PanReg = 0
	a.pendingPowerOff = true
}

func (a *Apu) PowerOn() {
	a.pendingPowerOn = true
	a.fsStep = 0
	a.fsCounter = 0
}

func (a *Apu) ClockFrameSequencer() {
	if a.pendingPowerOff {
		a.fsStep = 0
		a.pendingPowerOff = false
	}

	if a.pendingPowerOn {
		a.fsStep = 0
		a.pendingPowerOn = false
	}

	a.fsCounter++

	// frame sequencer runs at 512hz
	// length ctr at 256hz
	// sweep at 128hz
	// vol at 64hz

	if a.fsStep&1 == 0 {
		a.ToneChannel1.clockLength()
		a.ToneChannel2.clockLength()
		a.WaveChannel.clockLength()
		a.NoiseChannel.clockLength()
	}

	if a.fsStep == 2 || a.fsStep == 6 {
		a.ToneChannel1.clockSweep()
	}

	if a.fsStep == 7 {
		a.ToneChannel1.clockEnvelope()
		a.ToneChannel2.clockEnvelope()
		a.NoiseChannel.clockEnvelope()
	}

	a.fsStep = (a.fsStep + 1) & 7
}
