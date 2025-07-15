package apu

import (
	"fmt"

	"github.com/aabalke33/guac/emu/gba/utils"
	"github.com/hajimehoshi/oto"
)

// CURRENTLY USING OTO AND SDL
// SDL HAS PROBLEMS WITH OVER AND UNDER RUNS, SO OTO IS MAIN NOW

const OTO_VERSION = true

var (
    ApuInstance Apu
	sndCycles = uint32(0)
)

type Apu struct {
	Enable bool

    FifoA, FifoB *Fifo
    SoundCntL, SoundCntH, SoundCntX uint16
    SoundBuffer [BUFF_SIZE]int16
    ReadPointer, WritePointer uint32

    IO[0x120]uint8

    ToneChannel1 *ToneChannel
    ToneChannel2 *ToneChannel
    WaveChannel *WaveChannel
    NoiseChannel *NoiseChannel

    // These Are used by oto version only
	Stream []byte
    Context *oto.Context
    Player *oto.Player
}

func (a *Apu) isSoundChanEnable(ch uint8) bool {
    cntx := uint32(a.SoundCntX)
	return utils.BitEnabled(cntx, ch)
}

func InitApuInstance() {

    ApuInstance.Stream = make([]byte, STREAM_LEN)
    ApuInstance.WritePointer = 0x200
    ApuInstance.FifoA = &Fifo{}
    ApuInstance.FifoB = &Fifo{}

    ApuInstance.ToneChannel1 = &ToneChannel{
        Idx: 0,
    }
    ApuInstance.ToneChannel2 = &ToneChannel{
        Idx: 1,
    }
    ApuInstance.WaveChannel = &WaveChannel{
        Idx: 2,
    }
    ApuInstance.NoiseChannel = &NoiseChannel{
        Idx: 3,
    }


    if OTO_VERSION {

        context, err := oto.NewContext(SND_FREQUENCY, 2, 2, STREAM_LEN * 3)
        if err != nil {
            panic(err)
        }

        player := context.NewPlayer()

        ApuInstance.Context = context
        ApuInstance.Player = player
    }
}

func (a *Apu) Play() {

    if !OTO_VERSION {
        return
    }

	a.Enable = true

	if a.Stream == nil {
		return
	}

	if len(a.Stream) == 0 {
		return
	}

	a.soundMix()


	if a.Player == nil {
        return
	}

    //return

    go a.Player.Write(a.Stream)
}

func (a *Apu) soundMix() {
    
	for i := 0; i < STREAM_LEN; i += 4 {
        for j := range 2 {
            snd := a.SoundBuffer[a.ReadPointer&(BUFF_SIZE - 1)] << 6
            idx := i + (2 * j)
            a.Stream[idx] = uint8(snd)
            a.Stream[idx + 1] = uint8(snd>>8)
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

func (a *Apu) IsSoundEnabled() bool {
    return utils.BitEnabled(uint32(a.SoundCntX), 7)
}

func (a *Apu) GetSample() (int16, int16) {

    if OTO_VERSION {
        return 0, 0
    }

    if a.WritePointer == a.ReadPointer {
        fmt.Printf("WRITE AND READ OVERLAP\n")
    }

    l := a.SoundBuffer[a.ReadPointer&(BUFF_SIZE - 1)] << 6
    a.ReadPointer++

    r := a.SoundBuffer[a.ReadPointer&(BUFF_SIZE - 1)] << 6
    a.ReadPointer++

    return l, r
}

func (a *Apu) Sync() {

    if OTO_VERSION {
        return
    }

	delta := (int32(a.WritePointer-a.ReadPointer) >> 8) - (int32(a.WritePointer-a.ReadPointer)>>8)%4
    if delta > 0 {
        a.ReadPointer += uint32(delta)
    } else {
        a.ReadPointer -= uint32(delta)
    }
}

func (a *Apu) SoundBufferWrap() {
    l := a.ReadPointer / BUFF_SIZE
    r := a.WritePointer / BUFF_SIZE
	if l == r {
        a.ReadPointer &= (BUFF_SIZE - 1)
        a.WritePointer &= (BUFF_SIZE - 1)
	}
}

var (
    volLut = [8]int32{0x000, 0x024, 0x049, 0x06d, 0x092, 0x0b6, 0x0db, 0x100}
	rshLut = [4]int32{0xa, 0x9, 0x8, 0x7}
)

func (a *Apu) fifoFx(ch uint8, sample int16) (int16, int16) {

    if sample == 0 {
        return 0, 0
    }

    if ch != 0 && ch != 1 {
        panic("INVALID FIFO CHANNEL")
    }

	if half := !utils.BitEnabled(uint32(a.SoundCntH), 2 + ch); half {
		sample /= 2
	}

    sample = clip(int32(sample))

	sampleLeft := int16(0)
	sampleRight := int16(0)

	if leftEnabled := utils.BitEnabled(uint32(a.SoundCntH), 9 + (ch * 4)); leftEnabled {
		sampleLeft = sample
	}

	if rightEnabled := utils.BitEnabled(uint32(a.SoundCntH), 8 + (ch * 4)); rightEnabled {
		sampleRight = sample
    }

    return sampleLeft, sampleRight

}

func (a *Apu) SoundClock(cycles uint32) {

	sndCycles += cycles

	sampleA := int16(a.FifoA.Sample) << 1
	sampleB := int16(a.FifoB.Sample) << 1

    sampleLeftA, sampleRightA := a.fifoFx(0, sampleA)
    sampleLeftB, sampleRightB := a.fifoFx(1, sampleB)

    sampleLeft := clip(int32(sampleLeftA) + int32(sampleLeftB))
    sampleRight := clip(int32(sampleRightA) + int32(sampleRightB))

	for sndCycles >= SAMP_CYCLES {

        psgL, psgR := int32(0), int32(0)

        ch1Sample := int32(a.ToneChannel1.GetSample())

        if leftEnabled := utils.BitEnabled(uint32(a.SoundCntL), 12); leftEnabled {
            psgL = int32(clip(psgL + ch1Sample))
        }

        if rightEnabled := utils.BitEnabled(uint32(a.SoundCntL), 8); rightEnabled {
            psgR = int32(clip(psgR + ch1Sample))
        }

        ch2Sample := int32(a.ToneChannel2.GetSample())

        if leftEnabled := utils.BitEnabled(uint32(a.SoundCntL), 13); leftEnabled {
            psgL = int32(clip(psgL + ch2Sample))
        }

        if rightEnabled := utils.BitEnabled(uint32(a.SoundCntL), 9); rightEnabled {
            psgR = int32(clip(psgR + ch2Sample))
        }

        ch3Sample := int32(a.WaveChannel.GetSample())

        if leftEnabled := utils.BitEnabled(uint32(a.SoundCntL), 14); leftEnabled {
            psgL = int32(clip(psgL + ch3Sample))
        }

        if rightEnabled := utils.BitEnabled(uint32(a.SoundCntL), 10); rightEnabled {
            psgR = int32(clip(psgR + ch3Sample))
        }

        ch4Sample := int32(a.NoiseChannel.GetSample())

        if leftEnabled := utils.BitEnabled(uint32(a.SoundCntL), 15); leftEnabled {
            psgL = int32(clip(psgL + ch4Sample))
        }

        if rightEnabled := utils.BitEnabled(uint32(a.SoundCntL), 11); rightEnabled {
            psgR = int32(clip(psgR + ch4Sample))
        }

        psgL *= volLut[(a.SoundCntL>>4)&7]
		psgR *= volLut[(a.SoundCntL>>0)&7]

		psgL >>= rshLut[(a.SoundCntH)&3]
		psgR >>= rshLut[(a.SoundCntH)&3]

		a.SoundBuffer[a.WritePointer&(BUFF_SIZE - 1)] = clip(int32(sampleLeft) + psgL)
		a.WritePointer++
		a.SoundBuffer[a.WritePointer&(BUFF_SIZE - 1)] = clip(int32(sampleRight) + psgR)
		a.WritePointer++

		sndCycles -= SAMP_CYCLES
	}
}

func (a *Apu) enableSoundChan(ch int, enable bool) {

    if enable {
        a.SoundCntX |= (1 << ch)
        return
    }

    a.SoundCntX &^= (1 << ch)
}

func isResetSoundChan(addr uint32) bool {
	_, ok := resetSoundChanMap[addr]
	return ok
}

func (a *Apu) resetSoundChan(addr uint32, b byte) {
	a._resetSoundChan(resetSoundChanMap[addr], utils.BitEnabled(uint32(b), 7))
}

var resetSoundChanMap = map[uint32]int{0x65: 0, 0x6d: 1, 0x75: 2, 0x7d: 3}

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

            if utils.BitEnabled(uint32(a.NoiseChannel.CntH), 3) {
				a.NoiseChannel.lfsr = 0x0040 // 7bit
			} else {
				a.NoiseChannel.lfsr = 0x4000 // 15bit
			}
		}

		a.enableSoundChan(ch, true)
	}
}
