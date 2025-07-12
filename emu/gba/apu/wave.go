package apu

import "github.com/aabalke33/guac/emu/gba/utils"

type WaveChannel struct {
	Idx uint32

	CntL, CntH, CntX uint16
	WaveRam          [0x20]uint8

	samples, lengthTime float64

    WaveSamples, WavePosition uint8
}

func (ch *WaveChannel) GetSample() int8 {
	//wave := uint16(a.Load32(SOUND3CNT_L))
	//if !(a.isSoundChanEnable(2) && util.Bit(wave, 7)) {
	//	return 0
	//}

	// Actual frequency in Hertz
	rate := ch.CntX & 2047
	frequency := 2097152 / (2048 - float64(rate))

	// Full length of the generated wave (if enabled) in seconds
	soundLen := ch.CntH & 0xff
	length := (256 - float64(soundLen)) / 256.

	// Numbers of samples that a single "cycle" (all entries on Wave RAM) takes at output sample rate
	cycleSamples := SND_FREQUENCY / frequency

	// Length reached check (if so, just disable the channel and return silence)
	if utils.BitEnabled(uint32(ch.CntX), 14) {
		ch.lengthTime += SAMPLE_TIME
		if ch.lengthTime >= length {
            ApuInstance.enableSoundChan(int(ch.Idx), false)
			return 0
		}
	}

	ch.samples++
	if ch.samples >= cycleSamples {
		ch.samples -= cycleSamples

		ch.WaveSamples--
		if ch.WaveSamples != 0 {
			ch.WavePosition = (ch.WavePosition + 1) & 0b0011_1111
		} else {
            ch.Reset()
		}
	}

	wavedata := ch.WaveRam[(uint32(ch.WavePosition)>>1)&0x1f]
	sample := (float64((wavedata>>((ch.WavePosition&1)<<2))&0xf) - 0x8) / 8

	switch volume := (ch.CntH >> 13) & 0x7; volume {
	case 0:
		sample = 0 // 0%
	case 1: // 100%
	case 2:
		sample /= 2 // 50%
	case 3:
		sample /= 4 // 25%
	default:
		sample *= 3 / 4 // 75%
	}

	if sample >= 0 {
		return int8(sample / 7 * PSG_MAX)
	}
	return int8(sample / (-8) * PSG_MIN)

}

func (ch *WaveChannel) Reset() {

	if utils.BitEnabled(uint32(ch.CntL), 5) { // R/W Wave RAM Dimension
		// 64 samples (at 4 bits each, uses both banks so initial position is always 0)
		ch.WavePosition, ch.WaveSamples = 0, 64
		return
	}
	// 32 samples (at 4 bits each, bank selectable through Wave Control register)
	ch.WavePosition, ch.WaveSamples = byte((ch.CntL>>1)&0x20), 32
}
