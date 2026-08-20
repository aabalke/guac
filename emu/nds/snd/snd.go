package snd

import (
	"math"
	"time"

	"github.com/aabalke/guac/utils"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

const (
	REPEAT_MAN = 0
	REPEAT_INF = 1
	REPEAT_ONE = 2

	FMT_PCM8  = 0
	FMT_PCM16 = 1
	FMT_ADPCM = 2
	FMT_PSG   = 3
)

type Mem interface {
	Read(addr uint32, arm9 bool) uint8
	Read16(addr uint32, arm9 bool) uint32
	Read32(addr uint32, arm9 bool) uint32

	Write(addr uint32, v uint8, arm9 bool)
	Write16(addr uint32, v uint16, arm9 bool)
}

type Snd struct {
	Mem      Mem
	Stream   *utils.Stream
	Ctx      *audio.Context
	player   *audio.Player
	Channels [16]Channel
	Capture  [2]Capture

	VolMaster float64
	LOut      uint8
	ROut      uint8

	NoOutCh1 bool
	NoOutCh3 bool
	Enabled  bool
	Bias     uint32
}

func NewSnd(ctx *audio.Context, mem Mem, bufferSize time.Duration) *Snd {
	s := &Snd{
		Mem: mem,
	}

	if ctx != nil {

		s.Ctx = ctx

		s.Stream = utils.NewStream(bufferSize, ctx.SampleRate())

		var err error
		s.player, err = s.Ctx.NewPlayer(s.Stream)
		if err != nil {
			panic(err)
		}
		s.player.SetBufferSize(bufferSize)
		s.player.Play()
	}

	for i := range 16 {

		s.Channels[i] = NewChannel(i, s)

		switch {
		case i < 8:
			continue
		case i < 14:
			s.Channels[i].isDuty = true
		default:
			s.Channels[i].isNoise = true
		}
	}

	s.Capture = NewCaptures(s)

	return s
}

func (s *Snd) ToggleMute(muted bool) {
	if s.player != nil {
		if muted {
			s.player.SetVolume(0)
		} else {
			s.player.SetVolume(1)
		}
	}
}

func (s *Snd) TogglePause(paused bool) {
	if s.player != nil {
		if paused {
			s.player.Pause()
		} else {
			s.player.Play()
		}
	}
}

func (s *Snd) Close() {
	if s.player != nil {
		s.player.Close()
	}
}

func (s *Snd) SoundClock() {
	l, r := float64(0), float64(0)

	if s.Enabled {
		for i := range 16 {
			c := &s.Channels[i]
			cl, cr := c.GetSample()
			l += float64(cl)
			r += float64(cr)
		}

		if mixCapture := !s.Capture[0].ChanSrc; mixCapture {
			s.Capture[0].Capture(l)
		}
		if mixCapture := !s.Capture[1].ChanSrc; mixCapture {
			s.Capture[1].Capture(r)
		}

		l = (float64(l) * float64(s.VolMaster))
		r = (float64(r) * float64(s.VolMaster))
	}

	l *= 50
	r *= 50

	s.Stream.Write(int16(clip(int32(l))), int16(clip(int32(r))))
}

//go:inline
func clip(v int32) int32 {
	return min(math.MaxInt16, max(math.MinInt16, v))
}
