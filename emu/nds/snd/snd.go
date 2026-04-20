package snd

import (
	"fmt"

	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/oto"
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
	Mem Mem

	VolMaster float64
	LOut      uint8
	ROut      uint8

	NoOutCh1 bool
	NoOutCh3 bool
	Enabled  bool
	Bias     uint32

	Channels [16]Channel
	Capture  [2]Capture

	player    *oto.Player
	Stream    []uint8
	sndCycles uint32

	SoundBuffer               []int16
	ReadPointer, WritePointer uint32

	cpuFreqHz    int
	sndFrequency int
	sndSamples   int
	sampCycles   int
	buffSamples  int
	sampleTime   float64
	streamLen    int
	buffSize     uint32

	steamCh chan []uint8

	muted bool
}

func NewSnd(ctx *oto.Context, freq, rate, cnt int) *Snd {

	s := &Snd{
		WritePointer: 0x200,
		cpuFreqHz:    freq,
		sndFrequency: rate,
		sndSamples:   cnt,
		sampCycles:   freq / rate,
		buffSamples:  cnt * 16 * 2,
		sampleTime:   1.0 / float64(rate),
		streamLen:    (2 * 2 * rate / 60) - (2*2*rate/60)%4,
		buffSize:     uint32((cnt) * 16 * 2),
		//steamCh: make(chan []uint8, 10000),
	}

	s.Stream = make([]byte, s.streamLen)
	s.SoundBuffer = make([]int16, s.buffSize)

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

	if !config.Conf.CancelAudioInit {
		s.player = ctx.NewPlayer()
		//go s.runCh()
	}

	return s
}

func (s *Snd) runCh() {
	for stream := range s.steamCh {
		if s.muted {
			continue
		}

		if ebiten.ActualTPS() > 130 {
			continue
		}

		s.player.Write(stream)
	}
}

func (s *Snd) Play(muted, stdFps bool) {

	s.muted = muted

	s.SoundBufferWrap()

	if len(s.Stream) == 0 {
		return
	}

	s.Mix()

	if muted || s.player == nil {
		return
	}

    if !stdFps {
		return
	}

	s.player.Write(s.Stream)
}

func (s *Snd) Close() {
	s.player.Close()
}

func (s *Snd) Mix() {

	for i := 0; i < s.streamLen; i += 4 {
		for j := range 2 {
			snd := s.SoundBuffer[s.ReadPointer&uint32(s.buffSize-1)] << 6
			idx := i + (2 * j)
			s.Stream[idx] = uint8(snd)
			s.Stream[idx+1] = uint8(snd >> 8)
			s.ReadPointer++
		}
	}

	// Avoid desync between the Play cursor and the Write cursor
	delta := (int32(s.WritePointer-s.ReadPointer) >> 8) - (int32(s.WritePointer-s.ReadPointer)>>8)%2
	if delta > 0 {
		s.ReadPointer += uint32(delta)
	} else {
		s.ReadPointer -= uint32(delta)
	}
}

func (a *Snd) GetSample() (int16, int16) {

	if a.WritePointer == a.ReadPointer {
		fmt.Printf("WRITE AND READ OVERLAP\n")
	}

	l := a.SoundBuffer[a.ReadPointer&uint32(a.buffSize-1)] << 6
	a.ReadPointer++

	r := a.SoundBuffer[a.ReadPointer&uint32(a.buffSize-1)] << 6
	a.ReadPointer++

	return l, r
}

func (a *Snd) Sync() {

	delta := (int32(a.WritePointer-a.ReadPointer) >> 8) - (int32(a.WritePointer-a.ReadPointer)>>8)%4
	if delta > 0 {
		a.ReadPointer += uint32(delta)
	} else {
		a.ReadPointer -= uint32(delta)
	}
}

func (s *Snd) SoundBufferWrap() {
	l := s.ReadPointer / uint32(s.buffSize)
	r := s.WritePointer / uint32(s.buffSize)
	if l == r {
		s.ReadPointer &= (uint32(s.buffSize) - 1)
		s.WritePointer &= (uint32(s.buffSize) - 1)
	}
}

func (s *Snd) SoundClock(cycles uint32) {

	s.sndCycles += cycles

	for s.sndCycles >= uint32(s.sampCycles) {

		l := float64(0)
		r := float64(0)

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

		s.SoundBuffer[s.WritePointer&(s.buffSize-1)] = clip(int32(l))
		s.WritePointer++
		s.SoundBuffer[s.WritePointer&(s.buffSize-1)] = clip(int32(r))
		s.WritePointer++

		s.sndCycles -= uint32(s.sampCycles)
	}
}

const (
	SAMP_MAX = 0x1ff
	SAMP_MIN = -0x200
)

//go:inline
func clip(v int32) int16 {
	if v > SAMP_MAX {
		return SAMP_MAX
	}
	if v < SAMP_MIN {
		return SAMP_MIN
	}
	return int16(v)
}
