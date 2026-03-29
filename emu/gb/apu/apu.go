package apu

import (
	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/oto"
)

type Apu struct {
	Enabled bool

	FifoA, FifoB                    Fifo

    PanReg uint8
    Master uint8


	SoundBias                       uint16

	SoundBuffer               []int16
	ReadPointer, WritePointer uint32

	ToneChannel1 ToneChannel
	ToneChannel2 ToneChannel
	WaveChannel  WaveChannel
	NoiseChannel NoiseChannel

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
	buffSize     uint32

    fsCounter uint32
    fsStep    uint8

    pendingPowerOff bool
    pendingPowerOn  bool
    suppress bool
}

func (a *Apu) ClockFrameSequencer() {

    if a.pendingPowerOff {
        a.fsStep = 0
        a.pendingPowerOff = false
    }

    if a.pendingPowerOn {
        a.fsStep = 0
        a.pendingPowerOn = false
        a.suppress = true
    }

    a.fsCounter++

    // frame sequencer runs at 512hz
    // length ctr at 256hz
    // sweep at 128hz
    // vol at 64hz

    //a.ToneChannel1.InFirstHalf = a.fsStep & 1 != 0

    if a.fsStep & 1 == 0 {
        a.clockLengthCounters()
    }

    if a.fsStep & 7 == 2 || a.fsStep & 7 == 6 {
        //a.ToneChannel1.clockSweep2()
        a.ToneChannel1.clockSweep()
    }

    if a.fsStep & 7 == 7 {
        a.clockEnvelopes()
    }

    a.fsStep = (a.fsStep + 1) & 7
}

func (a *Apu) clockLengthCounters() {
    a.ToneChannel1.clockLength()
    a.ToneChannel2.clockLength()
    a.WaveChannel.clockLength()
    a.NoiseChannel.clockLength()
}

func (a *Apu) clockEnvelopes() {
    a.ToneChannel1.clockEnvelope()
    a.ToneChannel2.clockEnvelope()
    a.NoiseChannel.clockEnvelope()
}

func NewApu(audioContext *oto.Context, cpuFreq, sampleRate, sampleCnt int) *Apu {

	a := &Apu{
		WritePointer: 0x200,
		FifoA:        Fifo{},
		FifoB:        Fifo{},
		cpuFreqHz:    cpuFreq,
		sndFrequency: sampleRate,
		sndSamples:   sampleCnt,
		sampCycles:   cpuFreq / sampleRate,
		buffSamples:  sampleCnt * 16 * 2,
		sampleTime:   1.0 / float64(sampleRate),
		streamLen:    (2 * 2 * sampleRate / 60) - (2*2*sampleRate/60)%4,
		buffSize:     uint32((sampleCnt) * 16 * 2),
	}

	a.Stream = make([]byte, a.streamLen)
	a.SoundBuffer  = make([]int16, a.buffSize)
	a.ToneChannel1 = ToneChannel{Apu: a, Idx: 0}
	a.ToneChannel2 = ToneChannel{Apu: a, Idx: 1}
	a.WaveChannel  = WaveChannel{
        Apu: a,
        Idx: 2,
        WaveRam: [32]uint8{
            0x84, 0x40, 0x43, 0xAA, 0x2D, 0x78, 0x92, 0x3C,
            0x60, 0x59, 0x59, 0xB0, 0x34, 0xB8, 0x2E, 0xDA,
        },
    }
	a.NoiseChannel = NoiseChannel{Apu: a, Idx: 3}

	if !config.Conf.CancelAudioInit {
		a.player = audioContext.NewPlayer()
	}

	return a
}

func (a *Apu) Play(muted bool) {

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

	if ebiten.ActualTPS() > 65 {
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

func (a *Apu) SoundClock(cycles uint32, doubleSpeed bool) {

	a.sndCycles += cycles

    var (
        pan  = a.PanReg
        volL = (a.Master>>4)&7
        volR = (a.Master>>0)&7
    )

	clockCycles := uint32(a.sampCycles)
	if doubleSpeed {
		clockCycles <<= 1
	}

	for a.sndCycles >= clockCycles {

        psgL, psgR := int32(0), int32(0)

        if a.ToneChannel1.ChannelEnabled && (pan & 0x11 != 0) {
            ch := int32(a.ToneChannel1.GetSample(doubleSpeed))
            if pan & 0x10 != 0 {
                psgL += ch
            }
            if pan & 0x01 != 0 {
                psgR += ch
            }
        }
        if a.ToneChannel2.ChannelEnabled && (pan & 0x22 != 0) {
            ch := int32(a.ToneChannel2.GetSample(doubleSpeed))
            if pan & 0x20 != 0 {
                psgL += ch
            }
            if pan & 0x02 != 0 {
                psgR += ch
            }
        }
        if a.WaveChannel.ChannelEnabled && (pan & 0x44 != 0) {
            ch := int32(a.WaveChannel.GetSample(doubleSpeed))
            if pan & 0x40 != 0 {
                psgL += ch
            }
            if pan & 0x04 != 0 {
                psgR += ch
            }
        }
        if a.NoiseChannel.ChannelEnabled && (pan & 0x88 != 0) {
            ch := int32(a.NoiseChannel.GetSample(doubleSpeed))
            if pan & 0x80 != 0 {
                psgL += ch
            }
            if pan & 0x08 != 0 {
                psgR += ch
            }
        }

		psgL = ((psgL * int32(volL+1)) >> 3) >> 2
		psgR = ((psgR * int32(volR+1)) >> 3) >> 2

		a.SoundBuffer[a.WritePointer&(a.buffSize-1)] = clip(psgL)
		a.WritePointer++
		a.SoundBuffer[a.WritePointer&(a.buffSize-1)] = clip(psgR)
		a.WritePointer++

		a.sndCycles -= clockCycles
	}
}

func (a *Apu) PowerOff() {
    a.ToneChannel1 = ToneChannel{Idx: 0, Apu: a}
    a.ToneChannel2 = ToneChannel{Idx: 1, Apu: a}
    a.WaveChannel  = WaveChannel{Idx: 2, Apu: a, WaveRam: a.WaveChannel.WaveRam}
    a.NoiseChannel = NoiseChannel{Idx: 3, Apu: a, lfsr: a.NoiseChannel.lfsr}
    a.Master = 0
    a.PanReg = 0 
    a.pendingPowerOff = true
    //fmt.Printf("Power Off\n")
}

func (a *Apu) PowerOn() {
    a.pendingPowerOn = true
    //fmt.Printf("Power On\n")
}
