package apu

import (
	"github.com/hajimehoshi/oto"
)

const (
	SAMP_MAX = 0x1ff
	SAMP_MIN = -0x200
)

type Apu struct {
	Enabled bool

	PanReg uint8
	Master uint8

	SoundBuffer               []int16
	ReadPointer, WritePointer uint32

	ToneChannel1 ToneChannel
	ToneChannel2 ToneChannel
	WaveChannel  WaveChannel
	NoiseChannel NoiseChannel

	Stream []byte
	player *oto.Player

	sndFrequency int
	streamLen    int
	buffSize     uint32

	fsCounter uint32
	fsStep    uint8

	pendingPowerOff bool
	pendingPowerOn  bool
}

type Channel interface {
	GetSample() int8
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

func NewApu(audioContext *oto.Context, cpuFreq, sampleRate, sampleCnt int) *Apu {
	a := &Apu{
		WritePointer: 0x200,
		sndFrequency: sampleRate,
		streamLen:    (2 * 2 * sampleRate / 60) - (2*2*sampleRate/60)%4,
		buffSize:     uint32(sampleCnt * 16 * 2),
	}

	a.Stream = make([]byte, a.streamLen)
	a.SoundBuffer = make([]int16, a.buffSize)
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

	if audioContext != nil {
		a.player = audioContext.NewPlayer()
	}

	return a
}

func (a *Apu) Play(muted, stdFps bool) {
	a.SoundBufferWrap()

	if a.Stream == nil {
		return
	}

	if len(a.Stream) == 0 {
		return
	}

	a.soundMix()

	if muted {
		return
	}

	if a.player == nil {
		return
	}

	if !stdFps {
		return
	}

	a.player.Write(a.Stream)
}

func (a *Apu) Close() {
	a.player.Close()
}

func (a *Apu) soundMix() {
	for i := 0; i < a.streamLen; i += 4 {
		for j := range 2 {
			snd := a.SoundBuffer[a.ReadPointer&uint32(a.buffSize-1)] << 6
			idx := i + (2 * j)
			a.Stream[idx] = uint8(snd)
			a.Stream[idx+1] = uint8(snd >> 8)
			a.ReadPointer++
		}
	}

	// Avoid desync between the Play cursor and the Write cursor
	delta := (int32(a.WritePointer-a.ReadPointer) >> 8) - (int32(a.WritePointer-a.ReadPointer)>>8)%2
	if delta > 0 {
		a.ReadPointer += uint32(delta)
	} else {
		a.ReadPointer -= uint32(delta)
	}
}

func (a *Apu) SoundBufferWrap() {
	l := a.ReadPointer / uint32(a.buffSize)
	r := a.WritePointer / uint32(a.buffSize)
	if l == r {
		a.ReadPointer &= (uint32(a.buffSize) - 1)
		a.WritePointer &= (uint32(a.buffSize) - 1)
	}
}

func (a *Apu) SoundClock() {
	var (
		volL = int32((a.Master>>4)&7) + 1
		volR = int32((a.Master>>0)&7) + 1
		chs  = [4]struct {
			Channel       Channel
			Enabled, L, R bool
		}{
			{
				Channel: &a.ToneChannel1,
				Enabled: a.ToneChannel1.ChannelEnabled,
				L:       (a.PanReg & 0x10) != 0,
				R:       (a.PanReg & 0x01) != 0,
			},
			{
				Channel: &a.ToneChannel2,
				Enabled: a.ToneChannel2.ChannelEnabled,
				L:       (a.PanReg & 0x20) != 0,
				R:       (a.PanReg & 0x02) != 0,
			},
			{
				Channel: &a.WaveChannel,
				Enabled: a.WaveChannel.ChannelEnabled,
				L:       (a.PanReg & 0x40) != 0,
				R:       (a.PanReg & 0x04) != 0,
			},
			{
				Channel: &a.NoiseChannel,
				Enabled: a.NoiseChannel.ChannelEnabled,
				L:       (a.PanReg & 0x80) != 0,
				R:       (a.PanReg & 0x08) != 0,
			},
		}

		psgL int32
		psgR int32
	)

	for _, ch := range chs {
		if !ch.Enabled {
			continue
		}
		data := int32(ch.Channel.GetSample())
		if ch.L {
			psgL += data
		}
		if ch.R {
			psgR += data
		}
	}

	psgL = ((psgL * volL) >> 3) >> 2
	psgR = ((psgR * volR) >> 3) >> 2

	a.SoundBuffer[a.WritePointer&(a.buffSize-1)] = clip(psgL)
	a.WritePointer++
	a.SoundBuffer[a.WritePointer&(a.buffSize-1)] = clip(psgR)
	a.WritePointer++
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

//go:inline
func clip(v int32) int16 {
	return min(SAMP_MAX, max(SAMP_MIN, int16(v)))
}
