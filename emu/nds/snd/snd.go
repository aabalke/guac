package snd

import "github.com/hajimehoshi/oto"

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
    Read(addr uint32) uint8
}

type Snd struct {
	VolMaster uint32
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

type Channel struct {
    Snd *Snd
    Mem *Mem
	VolMul     uint32
	VolDiv     uint32
	Hold       bool
	Panning    uint32
	Duty       uint32 // wave duty
	RepeatMode uint32
	Format     uint32
	Busy       bool

	SrcAddr       uint32
	TimerValue    uint16
	StartPosition uint16
	SndLength     uint32
}

func NewSnd(ctx *oto.Context, freq, rate, cnt int) *Snd {

    s := &Snd{
    }

    return s
}

func (s *Snd) Play(muted bool) {

    // snd buffer wrap

    s.Enabled = true

    if len(s.Stream) == 0 {
        return
    }

    // snd mix

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
}
