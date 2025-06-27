package apu

import (

	"github.com/aabalke33/guac/emu/gba/utils"
	"github.com/hajimehoshi/oto"
)

const (
	CPU_FREQ_HZ              = 16777216
	SND_FREQUENCY            = 32768 // sample rate
	SND_SAMPLES              = 512 * 2
	SAMP_CYCLES              = (CPU_FREQ_HZ / SND_FREQUENCY)
	BUFF_SIZE             = ((SND_SAMPLES) * 16 * 2)
	BUFF_MSK         = ((BUFF_SIZE) - 1)
	SAMPLE_TIME      float64 = 1.0 / SND_FREQUENCY
	STREAM_LEN               = (2 * 2 * SND_FREQUENCY / 60) - (2*2*SND_FREQUENCY/60)%4

	PSG_MAX = 0x7f
	PSG_MIN = -0x80

	SAMP_MAX = 0x1ff
	SAMP_MIN = -0x200
)

var (
	sndCycles = uint32(0)
)

type DigitalAPU struct {
	Enable bool
	Stream []byte
    FifoA, FifoB *Fifo
    SoundCntH, SoundCntX uint16
    SoundBuffer [BUFF_SIZE]int16
    ReadPointer, WritePointer uint32
    Context *oto.Context
    Player *oto.Player
}

func NewDigitalAPU() *DigitalAPU {

    context, err := oto.NewContext(SND_FREQUENCY, 2, 2, STREAM_LEN)
	if err != nil {
		panic(err)
	}

    player := context.NewPlayer()

    a := &DigitalAPU{
        Stream: make([]byte, STREAM_LEN),
        ReadPointer: 0,
        WritePointer: 0x200,
        FifoA: &Fifo{},
        FifoB: &Fifo{},
        Context: context,
        Player: player,
    }

    return a
}

func (a *DigitalAPU) IsSoundEnabled() bool {
    return utils.BitEnabled(uint32(a.SoundCntX), 7)
}

func (a *DigitalAPU) Play() {

	a.Enable = true

	if a.Stream == nil {
		return
	}

	if len(a.Stream) == 0 {
		return
	}

	a.soundMix()

	if a.Player != nil {
        go a.Player.Write(a.Stream)
	}
}

func (a *DigitalAPU) soundMix() {

	for i := 0; i < STREAM_LEN; i += 4 {
        for j := range 2 {
            snd := a.SoundBuffer[a.ReadPointer&BUFF_MSK] << 6
            idx := i + (2 * j)
            a.Stream[idx] = uint8(snd)
            a.Stream[idx + 1] = uint8(snd>>8)
            a.ReadPointer++
        }
	}

	// Avoid desync between the Play cursor and the Write cursor
	delta := (int32(a.WritePointer-a.ReadPointer) >> 8) - (int32(a.WritePointer-a.ReadPointer)>>8)%2
	a.ReadPointer = AddInt32(a.ReadPointer, delta)
}

type Fifo struct {
    Buffer [0x20]int8
    Length uint8
    Sample int8
}

func (f *Fifo) Copy(v uint32) {

    if fifoFull := f.Length > 28; fifoFull {
        f.Length = 0
        //f.Length -= 28
    }

    for i := range 4 {
        f.Buffer[f.Length] = int8(v >> (8 * i))
        f.Length++
    }
}

func (f *Fifo) Load() {

    if f.Length == 0 {
        return
    }

    f.Sample = f.Buffer[0]
    f.Length--

    for i := range f.Length {
        f.Buffer[i] = f.Buffer[i+1]
    }
}

func clip(v int32) int32 {
    v = min(v, SAMP_MAX)
    v = max(v, SAMP_MIN)
	return int32(int16(v))
}

func (a *DigitalAPU) SoundClock(cycles uint32) {

    // THIS IS ALL JUST A AND WILL NEED TO BE UPDATED TO SUPPORT B

	sndCycles += cycles

    sampleLeft := int32(0)
    sampleRight := int32(0)

	sample := int32(a.FifoA.Sample)<<1

    if halfA := !utils.BitEnabled(uint32(a.SoundCntH), 2); halfA {
        sample /= 2
    }

    if leftEnabled := utils.BitEnabled(uint32(a.SoundCntH), 9); leftEnabled {
        sampleLeft = clip(sample)
	}

    if rightEnabled := utils.BitEnabled(uint32(a.SoundCntH), 8); rightEnabled {
        sampleRight = clip(sample)
	}

	for sndCycles >= SAMP_CYCLES {

		a.SoundBuffer[a.WritePointer&BUFF_MSK] = int16(sampleLeft)
		a.WritePointer++
		a.SoundBuffer[a.WritePointer&BUFF_MSK] = int16(sampleRight)
		a.WritePointer++

		sndCycles -= SAMP_CYCLES
	}
}

