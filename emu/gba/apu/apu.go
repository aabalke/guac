package apu

import (
	"math"
	"time"

	"github.com/aabalke/guac/utils"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

// akatsuki105/magia MIT License

type Apu struct {
	stream utils.Stream
	Ctx    *audio.Context
	player *audio.Player

	FifoA, FifoB                    Fifo
	SoundCntL, SoundCntH, SoundCntX uint16
	SoundBias                       uint16

	ToneChannel1 ToneChannel
	ToneChannel2 ToneChannel
	WaveChannel  WaveChannel
	NoiseChannel NoiseChannel

	cpuFreqHz    int
	sndFrequency int
	sampCycles   int
	sampleTime   float64
}

func (a *Apu) Disable() {
	a.ToneChannel1.CntL = 0
	a.ToneChannel1.CntH = 0
	a.ToneChannel1.CntX = 0

	a.ToneChannel2.CntL = 0
	a.ToneChannel2.CntH = 0
	a.ToneChannel2.CntX = 0

	a.WaveChannel.CntL = 0
	a.WaveChannel.CntH = 0
	a.WaveChannel.CntX = 0

	a.NoiseChannel.CntL = 0
	a.NoiseChannel.CntH = 0

	a.SoundCntL = 0
	//a.SoundCntH = 0
	//a.SoundCntX = 0
}

func (a *Apu) isSoundChanEnable(ch uint8) bool {
	cntx := uint32(a.SoundCntX)
	return BitEnabled(cntx, ch)
}

func NewApu(ctx *audio.Context, bufferSize time.Duration, cpuFreq int) *Apu {
	a := &Apu{
		FifoA:     Fifo{},
		FifoB:     Fifo{},
		cpuFreqHz: cpuFreq,
	}

	if ctx != nil {

		a.Ctx = ctx

		a.sndFrequency = ctx.SampleRate()
		a.sampCycles = cpuFreq / ctx.SampleRate()
		a.sampleTime = 1.0 / float64(ctx.SampleRate())

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

func (a *Apu) IsSoundEnabled() bool {
	return BitEnabled(uint32(a.SoundCntX), 7)
}

var (
	volLut = [8]int32{0x000, 0x024, 0x049, 0x06d, 0x092, 0x0b6, 0x0db, 0x100}
	rshLut = [4]int32{0xa, 0x9, 0x8, 0x7}
)

func (a *Apu) SoundClock(doubleSpeed bool) {
	shift0 := int32(a.SoundCntH>>2) & 1
	shift1 := int32(a.SoundCntH>>3) & 1
	lpan0 := int32(a.SoundCntH>>9) & 1
	rpan0 := int32(a.SoundCntH>>8) & 1
	lpan1 := int32(a.SoundCntH>>13) & 1
	rpan1 := int32(a.SoundCntH>>12) & 1

	sampleA := int32(a.FifoA.Latched) << (1 - shift0)
	sampleB := int32(a.FifoB.Latched) << (1 - shift1)

	sampleLeft := sampleA*lpan0 + sampleB*lpan1
	sampleRight := sampleA*rpan0 + sampleB*rpan1

	cntL := uint32(a.SoundCntL)
	volL := volLut[(cntL>>4)&0b111]
	volR := volLut[(cntL>>0)&0b111]
	shift := rshLut[a.SoundCntH&0b11]

	ch1 := int32(a.ToneChannel1.GetSample(doubleSpeed))
	ch2 := int32(a.ToneChannel2.GetSample(doubleSpeed))
	ch3 := int32(a.WaveChannel.GetSample(doubleSpeed))
	ch4 := int32(a.NoiseChannel.GetSample(doubleSpeed))

	psgL := ch1*int32((cntL>>12)&1) +
		ch2*int32((cntL>>13)&1) +
		ch3*int32((cntL>>14)&1) +
		ch4*int32((cntL>>15)&1)

	psgR := ch1*int32((cntL>>8)&1) +
		ch2*int32((cntL>>9)&1) +
		ch3*int32((cntL>>10)&1) +
		ch4*int32((cntL>>11)&1)

	psgL = (psgL * volL) >> shift
	psgR = (psgR * volR) >> shift

	l := int16(clip((psgL + sampleLeft) * 50))
	r := int16(clip((psgR + sampleRight) * 50))
	a.stream.Write(l, r)
}

//go:inline
func clip(v int32) int32 {
	return min(math.MaxInt16, max(math.MinInt16, v))
}

func (a *Apu) enableSoundChan(ch int, enable bool) {
	if enable {
		a.SoundCntX |= (1 << ch)
		return
	}

	a.SoundCntX &^= (1 << ch)
}

func IsResetSoundChan(addr uint32, isGB bool) bool {
	if isGB {
		_, ok := resetSoundChanMapGB[addr]
		return ok
	}
	_, ok := resetSoundChanMapGBA[addr]
	return ok
}

func (a *Apu) ResetSoundChan(addr uint32, b byte, isGB bool) {
	if isGB {
		a._resetSoundChan(resetSoundChanMapGB[addr], BitEnabled(uint32(b), 7))
		return
	}
	a._resetSoundChan(resetSoundChanMapGBA[addr], BitEnabled(uint32(b), 7))
}

var (
	resetSoundChanMapGBA = map[uint32]int{0x65: 0, 0x6d: 1, 0x75: 2, 0x7d: 3}
	resetSoundChanMapGB  = map[uint32]int{0x14: 0, 0x19: 1, 0x1E: 2, 0x23: 3}
)

func (a *Apu) _resetSoundChan(ch int, enable bool) {
	if enable {
		switch ch {
		case 0:

			a.ToneChannel1.phase = false
			a.ToneChannel1.samples = 0
			a.ToneChannel1.lengthTime = 0
			a.ToneChannel1.sweepTime = 0
			a.ToneChannel1.envTime = 0

		case 1:

			a.ToneChannel2.phase = false
			a.ToneChannel2.samples = 0
			a.ToneChannel2.lengthTime = 0
			a.ToneChannel2.sweepTime = 0
			a.ToneChannel2.envTime = 0

		case 2:
			a.WaveChannel.samples = 0
			a.WaveChannel.lengthTime = 0
			a.WaveChannel.Reset()
		case 3:

			a.NoiseChannel.samples = 0
			a.NoiseChannel.lengthTime = 0
			a.NoiseChannel.envTime = 0

			if BitEnabled(uint32(a.NoiseChannel.CntH), 3) {
				a.NoiseChannel.lfsr = 0x0040 // 7bit
			} else {
				a.NoiseChannel.lfsr = 0x4000 // 15bit
			}
		}

		a.enableSoundChan(ch, true)
	}
}
