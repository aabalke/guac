package apu

import (
    "fmt"

	"github.com/hajimehoshi/oto"
)

const base = -0x60

// Sound IO
const (
	SOUNDCNT_L, SOUNDCNT_H, SOUNDCNT_X    = base + 0x80, base + 0x82, base + 0x84
	SOUNDBIAS                             = base + 0x88
	FIFO_A                                = base + 0xa0
)

var (
    _ = fmt.Sprint()
    context *oto.Context
    player *oto.Player
    Stream []byte

    CNTH uint32
)

func Init() {
	Stream = make([]byte, STREAM_LEN)

	var err error
	context, err = oto.NewContext(SND_FREQUENCY, 2, 2, STREAM_LEN)
	if err != nil {
		panic(err)
	}

	player = context.NewPlayer()
}

func Play() {
	if player == nil {
		return
	}

	go player.Write(Stream)
}

const (
	CPU_FREQ_HZ              = 16777216
	SND_FREQUENCY            = 32768 // sample rate
	SND_SAMPLES              = 512 * 2
	SAMP_CYCLES              = (CPU_FREQ_HZ / SND_FREQUENCY)
	BUFF_SAMPLES             = ((SND_SAMPLES) * 16 * 2)
	BUFF_SAMPLES_MSK         = ((BUFF_SAMPLES) - 1)
	SAMPLE_TIME      float64 = 1.0 / SND_FREQUENCY
	STREAM_LEN               = (2 * 2 * SND_FREQUENCY / 60) - (2*2*SND_FREQUENCY/60)%4

	PSG_MAX = 0x7f
	PSG_MIN = -0x80

	SAMP_MAX = 0x1ff
	SAMP_MIN = -0x200
)

type DigitalAPU struct {
	Enable bool
	Buffer [72]byte
	Stream []byte
}

func (a *DigitalAPU) Load32(ofs uint32) uint32 {
	return LE32(a.Buffer[ofs:])
}

func (a *DigitalAPU) Store8(ofs uint32, val byte) {

	a.Buffer[ofs] = val
}
func (a *DigitalAPU) IsSoundMasterEnable() bool {
	cntx := byte(a.Load32(SOUNDCNT_X))
	return Bit(cntx, 7)
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
}

var (
	sndCurPlay  uint32 = 0
	sndCurWrite uint32 = 0x200
    sndBuffer [BUFF_SAMPLES]int16
)

func (a *DigitalAPU) soundMix() {
	for i := 0; i < STREAM_LEN; i += 4 {
		snd := sndBuffer[sndCurPlay&BUFF_SAMPLES_MSK] << 6
		a.Stream[i+0], a.Stream[i+1] = byte(snd), byte(snd>>8)
		sndCurPlay++
		snd = sndBuffer[sndCurPlay&BUFF_SAMPLES_MSK] << 6
		a.Stream[i+2], a.Stream[i+3] = byte(snd), byte(snd>>8)
		sndCurPlay++

	}

	// Avoid desync between the Play cursor and the Write cursor
	delta := (int32(sndCurWrite-sndCurPlay) >> 8) - (int32(sndCurWrite-sndCurPlay)>>8)%2
	sndCurPlay = AddInt32(sndCurPlay, delta)
}

var (
	FifoALen    byte
	fifoA       [0x20]int8
	fifoASamp int8
	sndCycles = uint32(0)
	psgVolLut = [8]int32{0x000, 0x024, 0x049, 0x06d, 0x092, 0x0b6, 0x0db, 0x100}
	psgRshLut = [4]int32{0xa, 0x9, 0x8, 0x7}
)


func FifoACopy(val uint32) {
	if FifoALen > 28 { // FIFO A full
		FifoALen -= 28
	}

	for i := uint32(0); i < 4; i++ {
		fifoA[FifoALen] = int8(val >> (8 * i))
		FifoALen++
	}
}

func FifoALoad() {
	if FifoALen == 0 {
		return
	}

	fifoASamp = fifoA[0]
	FifoALen--

	for i := byte(0); i < FifoALen; i++ {
		fifoA[i] = fifoA[i+1]
	}
}

func clip(val int32) int16 {

	if val > SAMP_MAX {
		val = SAMP_MAX
	}
	if val < SAMP_MIN {
		val = SAMP_MIN
	}

	return int16(val)
}

func (a *DigitalAPU) SoundClock(cycles uint32) {

	sndCycles += cycles

	sampPcmL, sampPcmR := int16(0), int16(0)

	cnth := uint16(a.Load32(SOUNDCNT_H)) // snd_pcm_vol
	volADiv := int16((cnth>>2)&0b1)^1
	sampCh4 := (int16(fifoASamp)<<1)>>volADiv

	// Left
	if Bit(cnth, 9) {
		sampPcmL = clip(int32(sampPcmL) + int32(sampCh4))
	}

	// Right
	if Bit(cnth, 8) {
		sampPcmR = clip(int32(sampPcmR) + int32(sampCh4))
	}

	for sndCycles >= SAMP_CYCLES {
		sampPsgL, sampPsgR := int32(0), int32(0)

		cntl := uint16(a.Load32(SOUNDCNT_L)) // snd_psg_vol

		sampPsgL *= psgVolLut[(cntl>>4)&7]
		sampPsgR *= psgVolLut[(cntl>>0)&7]

		sampPsgL >>= psgRshLut[(cnth>>0)&3]
		sampPsgR >>= psgRshLut[(cnth>>0)&3]

		sndBuffer[sndCurWrite&BUFF_SAMPLES_MSK] = clip(sampPsgL + int32(sampPcmL))
		sndCurWrite++
		sndBuffer[sndCurWrite&BUFF_SAMPLES_MSK] = clip(sampPsgR + int32(sampPcmR))
		sndCurWrite++

		sndCycles -= SAMP_CYCLES
	}
}
