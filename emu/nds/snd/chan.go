package snd

const (
	MAX = 127
	MIN = -128

	AMPLIFICATION = 0.125
)

type Channel struct {
	Snd        *Snd
	Mem        *Mem
	VolMul     uint32
	VolDiv     uint32
	Hold       bool
	Panning    uint32
	Duty       uint32 // wave duty
	RepeatMode uint32
	Format     uint32

	Start   bool
	Playing bool

	SrcAddr       uint32
	TimerValue    uint16
	StartPosition uint16
	SndLength     uint32

	pulse uint32
	phase float64
	freq  float64

	SampleIdx uint32

	SamplePos float64

    Samples []int16
}

func NewChannel(s *Snd, freq float64) Channel {

	c := Channel{
		Snd:  s,
		freq: freq,
	}

	return c
}

func (c *Channel) GetSample() (int8, int8) {

    if c.Format != 2 && c.Format != 0 {
		return 0, 0
	}

	if c.Start {
		c.Playing = true
		c.Start = false
		c.SamplePos = 0

        if c.Format == 2 {
            c.DecompressADPCM()
        }
	}

	if !c.Playing {
		return 0, 0
	}

	const BASE_FREQ = 33_513_982 / 2
	if c.TimerValue == 0 {
		return 0, 0
	}

	playbackRate := BASE_FREQ / float64(-int16(c.TimerValue))
	step := playbackRate / float64(c.Snd.sndFrequency)

	c.SamplePos += step

	var sample float64
	switch c.Format {
	case 0:
	    sample = c.GetPCM8()
	case 1:
		sample = c.GetPCM16()
    case 2:
		sample = c.GetADPCM()
	default:
		return 0, 0
	}

	switch c.VolDiv {
	case 1:
		sample /= 2
	case 2:
		sample /= 4
	case 3:
		sample /= 16
	}
	sample *= float64(c.VolMul)

	l := sample
	r := sample

	l *= (float64(127-c.Panning) / 127)
	r *= (float64(c.Panning) / 127)

	//l *= AMPLIFICATION
	//r *= AMPLIFICATION

	return int8(l), int8(r)
}

func (c *Channel) GetPCM8() float64 {

	sampleLen := (c.SndLength + uint32(c.StartPosition)) * 4

	if uint32(c.SamplePos) >= sampleLen {

		if loop := c.RepeatMode == 1; loop {
			c.SamplePos = float64(c.StartPosition) * 4

		} else {
			c.Playing = false
			return 0
		}
	}

	addr := c.SrcAddr + uint32(c.SamplePos)
	return float64(int8(c.Snd.Mem.Read(addr, false))) / 128
}

func (c *Channel) GetPCM16() float64 {

	sampleLen := (c.SndLength + uint32(c.StartPosition)) * 2

	if uint32(c.SamplePos) >= sampleLen {

		if loop := c.RepeatMode == 1; loop {
			c.SamplePos = float64(c.StartPosition * 2)

		} else {
			c.Playing = false
			return 0
		}
	}

	addr := c.SrcAddr + uint32(c.SamplePos*2)

	v := uint16(c.Snd.Mem.Read(addr+0, false))
	v |= uint16(c.Snd.Mem.Read(addr+1, false)) << 8

	return float64(int16(v)) / 32768
}

var adpcmIndexTable = [8]int16{-1, -1, -1, -1, 2, 4, 6, 8}
var adpcmTable      = [89]uint16{
		0x0007, 0x0008, 0x0009, 0x000A, 0x000B, 0x000C, 0x000D, 0x000E, 0x0010, 0x0011, 0x0013, 0x0015,
		0x0017, 0x0019, 0x001C, 0x001F, 0x0022, 0x0025, 0x0029, 0x002D, 0x0032, 0x0037, 0x003C, 0x0042,
		0x0049, 0x0050, 0x0058, 0x0061, 0x006B, 0x0076, 0x0082, 0x008F, 0x009D, 0x00AD, 0x00BE, 0x00D1,
		0x00E6, 0x00FD, 0x0117, 0x0133, 0x0151, 0x0173, 0x0198, 0x01C1, 0x01EE, 0x0220, 0x0256, 0x0292,
		0x02D4, 0x031C, 0x036C, 0x03C3, 0x0424, 0x048E, 0x0502, 0x0583, 0x0610, 0x06AB, 0x0756, 0x0812,
		0x08E0, 0x09C3, 0x0ABD, 0x0BD0, 0x0CFF, 0x0E4C, 0x0FBA, 0x114C, 0x1307, 0x14EE, 0x1706, 0x1954,
		0x1BDC, 0x1EA5, 0x21B6, 0x2515, 0x28CA, 0x2CDF, 0x315B, 0x364B, 0x3BB9, 0x41B2, 0x4844, 0x4F7E,
		0x5771, 0x602F, 0x69CE, 0x7462, 0x7FFF,
	}

func (c *Channel) DecompressADPCM() {

    addr := c.SrcAddr
    length := (c.SndLength * 4) + uint32(4 * (c.StartPosition))
	c.Samples = make([]int16, 0, length)

    head := uint32(c.Snd.Mem.Read(addr+0, false))
    head |= uint32(c.Snd.Mem.Read(addr+1, false)) << 8
    head |= uint32(c.Snd.Mem.Read(addr+2, false)) << 16
    head |= uint32(c.Snd.Mem.Read(addr+3, false)) << 24

    addr += 4

	pcm := int32(int16(head & 0xFFFF))
	index := int16(head>>16) & 0x7F

	dec := func(sample uint8) {
		diff := adpcmTable[index] / 8
		diff += (adpcmTable[index] / 4) * uint16((sample>>0)&1)
		diff += (adpcmTable[index] / 2) * uint16((sample>>1)&1)
		diff += (adpcmTable[index] / 1) * uint16((sample>>2)&1)
		if sample&8 == 0 {
			pcm += int32(diff)
			if pcm > 0x7FFF {
				pcm = 0x7FFF
			}
		} else {
			pcm -= int32(diff)
			if pcm < -0x7FFF {
				pcm = -0x7FFF
			}
		}

		index += adpcmIndexTable[sample&7]
		if index < 0 {
			index = 0
		} else if index > 88 {
			index = 88
		}
	}

	for i := range length {

        v := c.Snd.Mem.Read(addr + i, false)

		dec(v & 0xF)
        a := uint16(pcm&0xFF)
		a |= uint16((pcm>>8)&0xFF) << 8
        c.Samples = append(c.Samples, int16(a))

		dec(v >> 4)
        b := uint16(pcm&0xFF)
		b |= uint16((pcm>>8)&0xFF) << 8
        c.Samples = append(c.Samples, int16(b))
	}
}

func (c *Channel) GetADPCM() float64 {

	if int(c.SamplePos) >= len(c.Samples) {

		if loop := c.RepeatMode == 1; loop {
			c.SamplePos = (float64((c.StartPosition) * 4) - 4) / 2

		} else {
			c.Playing = false
			return 0
		}
	}

    v := c.Samples[uint32(c.SamplePos)]
	return float64(v) / 32768
}
