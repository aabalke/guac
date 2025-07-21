package apu

import (
	"fmt"

	//"github.com/aabalke33/guac/audio"
	"github.com/aabalke33/guac/emu/gba/utils"
	"github.com/hajimehoshi/oto"
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

	Stream []byte

    sndCycles uint32

    player *oto.Player
}

func (a *Apu) isSoundChanEnable(ch uint8) bool {
    cntx := uint32(a.SoundCntX)
	return utils.BitEnabled(cntx, ch)
}

func NewApu(audioContext *oto.Context) *Apu {
    a := &Apu {
        Stream: make([]byte, STREAM_LEN),
        WritePointer: 0x200,
        FifoA: &Fifo{},
        FifoB: &Fifo{},
    }

    a.ToneChannel1 = &ToneChannel{Apu: a, Idx: 0}
    a.ToneChannel2 = &ToneChannel{Apu: a, Idx: 1}
    a.WaveChannel  = &WaveChannel{Apu: a, Idx: 2}
    a.NoiseChannel = &NoiseChannel{Apu: a, Idx: 3}

    a.player = audioContext.NewPlayer()

    return a
}

func (a *Apu) Play(muted bool, frame uint64) {

	a.Enable = true

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


    a.player.Write(a.Stream)
    //audio.WriteGBA(a.Stream)
}

func (a *Apu) Close() {
    a.player.Close()
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

    //sample = clip(int32(sample))

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

	a.sndCycles += cycles

	sampleA := int16(a.FifoA.Sample) << 1
	sampleB := int16(a.FifoB.Sample) << 1

    sampleLeftA, sampleRightA := a.fifoFx(0, sampleA)
    sampleLeftB, sampleRightB := a.fifoFx(1, sampleB)

    //sampleLeft := clip(int32(sampleLeftA) + int32(sampleLeftB))
    //sampleRight := clip(int32(sampleRightA) + int32(sampleRightB))
    sampleLeft := int32(sampleLeftA) + int32(sampleLeftB)
    sampleRight := int32(sampleRightA) + int32(sampleRightB)

	for a.sndCycles >= SAMP_CYCLES {

        psgL, psgR := int32(0), int32(0)

        ch1Sample := int32(a.ToneChannel1.GetSample())

        if leftEnabled := utils.BitEnabled(uint32(a.SoundCntL), 12); leftEnabled {
            psgL = psgL + ch1Sample
        }

        if rightEnabled := utils.BitEnabled(uint32(a.SoundCntL), 8); rightEnabled {
            psgR = psgR + ch1Sample
        }

        ch2Sample := int32(a.ToneChannel2.GetSample())

        if leftEnabled := utils.BitEnabled(uint32(a.SoundCntL), 13); leftEnabled {
            psgL = psgL + ch2Sample
        }

        if rightEnabled := utils.BitEnabled(uint32(a.SoundCntL), 9); rightEnabled {
            psgR = psgR + ch2Sample
        }

        ch3Sample := int32(a.WaveChannel.GetSample())

        if leftEnabled := utils.BitEnabled(uint32(a.SoundCntL), 14); leftEnabled {
            psgL = psgL + ch3Sample
        }

        if rightEnabled := utils.BitEnabled(uint32(a.SoundCntL), 10); rightEnabled {
            psgR = psgR + ch3Sample
        }

        ch4Sample := int32(a.NoiseChannel.GetSample())

        if leftEnabled := utils.BitEnabled(uint32(a.SoundCntL), 15); leftEnabled {
            psgL = psgL + ch4Sample
        }

        if rightEnabled := utils.BitEnabled(uint32(a.SoundCntL), 11); rightEnabled {
            psgR = psgR + ch4Sample
        }

        psgL *= volLut[(a.SoundCntL>>4)&7]
		psgR *= volLut[(a.SoundCntL>>0)&7]

		psgL >>= rshLut[(a.SoundCntH)&3]
		psgR >>= rshLut[(a.SoundCntH)&3]

		a.SoundBuffer[a.WritePointer&(BUFF_SIZE - 1)] = clip(int32(sampleLeft) + psgL)
		a.WritePointer++
		a.SoundBuffer[a.WritePointer&(BUFF_SIZE - 1)] = clip(int32(sampleRight) + psgR)
		a.WritePointer++

		a.sndCycles -= SAMP_CYCLES
	}
}

func (a *Apu) enableSoundChan(ch int, enable bool) {

    if enable {
        a.SoundCntX |= (1 << ch)
        return
    }

    a.SoundCntX &^= (1 << ch)
}

func IsResetSoundChan(addr uint32) bool {
	_, ok := resetSoundChanMap[addr]
	return ok
}

func (a *Apu) ResetSoundChan(addr uint32, b byte) {
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
