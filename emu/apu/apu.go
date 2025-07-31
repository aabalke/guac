package apu

import (
	"fmt"

	"github.com/hajimehoshi/oto"
)

// akatsuki105/magia MIT License

type Apu struct {
	Enable bool

	FifoA, FifoB                    *Fifo
	SoundCntL, SoundCntH, SoundCntX uint16
	SoundBias                       uint16

	SoundBuffer               []int16
	ReadPointer, WritePointer uint32

	ToneChannel1 *ToneChannel
	ToneChannel2 *ToneChannel
	WaveChannel  *WaveChannel
	NoiseChannel *NoiseChannel

	Stream []byte

	sndCycles uint32

	player *oto.Player

	cpuFreqHz    int
	sndFrequency int
	sndSamples   int
	sampCycles   int
	buffSamples  int
	sampleTime   float64
	streamLen    int
	buffSize     int
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

func NewApu(audioContext *oto.Context, cpuFreq, sampleRate, sampleCnt int) *Apu {

	a := &Apu{
		WritePointer: 0x200,
		FifoA:        &Fifo{},
		FifoB:        &Fifo{},
		cpuFreqHz:    cpuFreq,
		sndFrequency: sampleRate,
		sndSamples:   sampleCnt,
		sampCycles:   cpuFreq / sampleRate,
		buffSamples:  sampleCnt * 16 * 2,
		sampleTime:   1.0 / float64(sampleRate),
		streamLen:    (2 * 2 * sampleRate / 60) - (2*2*sampleRate/60)%4,
		buffSize:     ((sampleCnt) * 16 * 2),
	}

	a.Stream = make([]byte, a.streamLen)
	a.SoundBuffer = make([]int16, a.buffSize)
	a.ToneChannel1 = &ToneChannel{Apu: a, Idx: 0}
	a.ToneChannel2 = &ToneChannel{Apu: a, Idx: 1}
	a.WaveChannel = &WaveChannel{Apu: a, Idx: 2}
	a.NoiseChannel = &NoiseChannel{Apu: a, Idx: 3}

    return a

	a.player = audioContext.NewPlayer()

	return a
}

func (a *Apu) Play(muted bool) {

	a.SoundBufferWrap()

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

	//a.player.Write(a.Stream)
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

func (a *Apu) IsSoundEnabled() bool {
	return BitEnabled(uint32(a.SoundCntX), 7)
}

func (a *Apu) GetSample() (int16, int16) {

	if a.WritePointer == a.ReadPointer {
		fmt.Printf("WRITE AND READ OVERLAP\n")
	}

	l := a.SoundBuffer[a.ReadPointer&uint32(a.buffSize-1)] << 6
	a.ReadPointer++

	r := a.SoundBuffer[a.ReadPointer&uint32(a.buffSize-1)] << 6
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
	l := a.ReadPointer / uint32(a.buffSize)
	r := a.WritePointer / uint32(a.buffSize)
	if l == r {
		a.ReadPointer &= (uint32(a.buffSize) - 1)
		a.WritePointer &= (uint32(a.buffSize) - 1)
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

	if ch > 1 {
		panic("INVALID FIFO CHANNEL")
	}

    cntH := uint32(a.SoundCntH)

	if half := !BitEnabled(cntH, 2+ch); half {
		sample >>= 1
	}

	var sampleLeft, sampleRight int16

	if leftEnabled := BitEnabled(cntH, 9+(ch << 2)); leftEnabled {
		sampleLeft = sample
	}

	if rightEnabled := BitEnabled(cntH, 8+(ch << 2)); rightEnabled {
		sampleRight = sample
	}

	return sampleLeft, sampleRight
}

func (a *Apu) SoundClock(cycles uint32, doubleSpeed bool) {

	a.sndCycles += cycles

	sampleA := int16(a.FifoA.Sample) << 1
	sampleB := int16(a.FifoB.Sample) << 1

	sampleLeftA, sampleRightA := a.fifoFx(0, sampleA)
	sampleLeftB, sampleRightB := a.fifoFx(1, sampleB)

	sampleLeft := int32(sampleLeftA) + int32(sampleLeftB)
	sampleRight := int32(sampleRightA) + int32(sampleRightB)

	multiplier := 1
	if doubleSpeed {
		multiplier = 2
	}

    cntL := uint32(a.SoundCntL)

	for a.sndCycles >= uint32(a.sampCycles*multiplier) {

		psgL, psgR := int32(0), int32(0)

		ch1Sample := int32(a.ToneChannel1.GetSample(doubleSpeed))

		if leftEnabled := BitEnabled(uint32(cntL), 12); leftEnabled {
			psgL = psgL + ch1Sample
		}

		if rightEnabled := BitEnabled(uint32(cntL), 8); rightEnabled {
			psgR = psgR + ch1Sample
		}

		ch2Sample := int32(a.ToneChannel2.GetSample(doubleSpeed))

		if leftEnabled := BitEnabled(uint32(cntL), 13); leftEnabled {
			psgL = psgL + ch2Sample
		}

		if rightEnabled := BitEnabled(uint32(cntL), 9); rightEnabled {
			psgR = psgR + ch2Sample
		}

		ch3Sample := int32(a.WaveChannel.GetSample(doubleSpeed))

		if leftEnabled := BitEnabled(uint32(cntL), 14); leftEnabled {
			psgL = psgL + ch3Sample
		}

		if rightEnabled := BitEnabled(uint32(cntL), 10); rightEnabled {
			psgR = psgR + ch3Sample
		}

		ch4Sample := int32(a.NoiseChannel.GetSample(doubleSpeed))

		if leftEnabled := BitEnabled(uint32(cntL), 15); leftEnabled {
			psgL = psgL + ch4Sample
		}

		if rightEnabled := BitEnabled(uint32(cntL), 11); rightEnabled {
			psgR = psgR + ch4Sample
		}

		psgL *= volLut[(cntL>>4)&7]
		psgR *= volLut[(cntL>>0)&7]

		psgL >>= rshLut[(a.SoundCntH)&3]
		psgR >>= rshLut[(a.SoundCntH)&3]

		a.SoundBuffer[a.WritePointer&uint32(a.buffSize-1)] = clip(int32(sampleLeft) + psgL)
		a.WritePointer++
		a.SoundBuffer[a.WritePointer&uint32(a.buffSize-1)] = clip(int32(sampleRight) + psgR)
		a.WritePointer++

		a.sndCycles -= uint32(a.sampCycles * multiplier)
	}
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

var resetSoundChanMapGBA = map[uint32]int{0x65: 0, 0x6d: 1, 0x75: 2, 0x7d: 3}
var resetSoundChanMapGB = map[uint32]int{0x14: 0, 0x19: 1, 0x1E: 2, 0x23: 3}

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
