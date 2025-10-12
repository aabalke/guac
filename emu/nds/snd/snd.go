package snd

import (
	"fmt"

	"github.com/aabalke/guac/config"
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

    player *oto.Player
    Stream []uint8
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
    }

	s.Stream = make([]byte, s.streamLen)
	s.SoundBuffer = make([]int16, s.buffSize)

    for i := range 16 {
        s.Channels[i] = NewChannel(s)
    }

    s.Channels[8].isDuty = true
    s.Channels[9].isDuty = true
    s.Channels[10].isDuty = true
    s.Channels[11].isDuty = true
    s.Channels[12].isDuty = true
    s.Channels[13].isDuty = true
    s.Channels[14].isNoise = true
    s.Channels[15].isNoise = true

	if !config.Conf.CancelAudioInit {
		s.player = ctx.NewPlayer()
	}

    return s
}

func (s *Snd) Play(muted bool) {

    s.SoundBufferWrap()

    if len(s.Stream) == 0 {
        return
    }

    s.Mix()

    if muted || s.player == nil {
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
            snd := s.SoundBuffer[s.ReadPointer& uint32(s.buffSize-1)] << 6
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

        l := int32(0)
        r := int32(0)

        if s.Enabled {
            for i := range 16 {
                c := &s.Channels[i]
                cl, cr := c.GetSample()
                l += int32(cl)
                r += int32(cr)
            }

            l = int32(float64(l) * float64(s.VolMaster))
            r = int32(float64(r) * float64(s.VolMaster))
        }

		s.SoundBuffer[s.WritePointer&(s.buffSize-1)] = clip(l)
		s.WritePointer++
		s.SoundBuffer[s.WritePointer&(s.buffSize-1)] = clip(r)
		s.WritePointer++

		s.sndCycles -= uint32(s.sampCycles)
	}
}
