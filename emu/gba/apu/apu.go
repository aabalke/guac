package apu

import (
	"time"

	"github.com/aabalke/guac/emu/gb/apu"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

type Apu struct {
	apu.Apu

	FifoA, FifoB Fifo

	SoundCntL uint16
	SoundCntH uint16
	SoundBias uint16
}

func NewApu(ctx *audio.Context, bufferSize time.Duration) *Apu {
	a := &Apu{
		FifoA: Fifo{},
		FifoB: Fifo{},
	}

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

	a.ToneChannel1 = apu.ToneChannel{FsStep: &a.FsStep, Idx: 0}
	a.ToneChannel2 = apu.ToneChannel{FsStep: &a.FsStep, Idx: 1}
	a.WaveChannel = apu.WaveChannel{FsStep: &a.FsStep, Idx: 2}
	a.NoiseChannel = apu.NoiseChannel{FsStep: &a.FsStep, Idx: 3}

	return a
}

var psgShiftTbl = [4]int32{2, 1, 0, 0}

func (a *Apu) SoundClock() {
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

	ch3, ch4 = 0, 0
	pcmL, pcmR = 0, 0

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

	l := int16(apu.Clip((psgL + pcmL) * 100))
	r := int16(apu.Clip((psgR + pcmR) * 100))
	a.Stream.Write(l, r)
}

func (a *Apu) PowerOff() {
	a.ToneChannel1 = apu.ToneChannel{Idx: 0, FsStep: &a.FsStep}
	a.ToneChannel2 = apu.ToneChannel{Idx: 1, FsStep: &a.FsStep}
	a.WaveChannel = apu.WaveChannel{Idx: 2, FsStep: &a.FsStep, Ram: a.WaveChannel.Ram}
	a.NoiseChannel = apu.NoiseChannel{Idx: 3, FsStep: &a.FsStep}

	a.SoundCntL = 0
	a.SoundCntH = 0

	a.PendingPowerOff = true
}
