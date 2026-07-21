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

	SoundCntL uint16
	SoundCntH uint16
	SoundBias uint16

	FifoA, FifoB Fifo
	ToneChannel1 ToneChannel
	ToneChannel2 ToneChannel
	WaveChannel  WaveChannel
	NoiseChannel NoiseChannel

	fsCounter       uint32
	fsStep          uint8
	pendingPowerOff bool
	pendingPowerOn  bool
	Enabled         bool
}

func NewApu(ctx *audio.Context, bufferSize time.Duration, cpuFreq int) *Apu {
	a := &Apu{
		FifoA: Fifo{},
		FifoB: Fifo{},
	}

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
	a.WaveChannel = WaveChannel{Apu: a, Idx: 2}
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

var psgShiftTbl = [4]int32{2, 1, 0, 0}

func (a *Apu) SoundClock(doubleSpeed bool) {
	var (
		cntL      = a.SoundCntL
		cntH      = a.SoundCntH
		psgVolL   = int32((cntL>>4)&7) + 1
		psgVolR   = int32((cntL>>0)&7) + 1
		psgShift  = psgShiftTbl[cntH&3]
		pcmAShift = (cntH >> 2) & 1
		pcmBShift = (cntH >> 3) & 1
		rpanA     = int32(cntH>>8) & 1
		lpanA     = int32(cntH>>9) & 1
		rpanB     = int32(cntH>>12) & 1
		lpanB     = int32(cntH>>13) & 1
	)

	var ch1, ch2, ch3, ch4, pcmA, pcmB int32

	pcmA = int32(a.FifoA.Get()) >> (1 - pcmAShift)
	pcmB = int32(a.FifoB.Get()) >> (1 - pcmBShift)

	pcmL := pcmA*lpanA + pcmB*lpanB
	pcmR := pcmA*rpanA + pcmB*rpanB

	if a.ToneChannel1.ChannelEnabled {
		ch1 = int32(a.ToneChannel1.GetSample())
	}

	if a.ToneChannel2.ChannelEnabled {
		ch2 = int32(a.ToneChannel2.GetSample())
	}

	if a.WaveChannel.ChannelEnabled {
		ch3 = int32(a.WaveChannel.GetSample())
	}

	if a.NoiseChannel.ChannelEnabled {
		ch4 = int32(a.NoiseChannel.GetSample())
	}

	psgL := ch1*int32((cntL>>12)&1) +
		ch2*int32((cntL>>13)&1) +
		ch3*int32((cntL>>14)&1) +
		ch4*int32((cntL>>15)&1)

	psgR := ch1*int32((cntL>>8)&1) +
		ch2*int32((cntL>>9)&1) +
		ch3*int32((cntL>>10)&1) +
		ch4*int32((cntL>>11)&1)
	psgL = ((psgL * psgVolL) >> psgShift)
	psgR = ((psgR * psgVolR) >> psgShift)

	// not sure on proper volume adj
	psgL >>= 4
	psgR >>= 4

	l := int16(clip((psgL + pcmL) * 100))
	r := int16(clip((psgR + pcmR) * 100))
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

	a.SoundCntL = 0
	a.SoundCntH = 0

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
