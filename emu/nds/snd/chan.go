package snd


const (
	MAX = 127
	MIN = -128

	AMPLIFICATION = 0.125
	BASE_FREQ = 33_513_982 / 2
)

var (

    adpcmIndexTable = [8]int16{-1, -1, -1, -1, 2, 4, 6, 8}
    adpcmTable = [89]uint16{
        0x0007, 0x0008, 0x0009, 0x000A, 0x000B, 0x000C, 0x000D, 0x000E, 0x0010, 0x0011, 0x0013, 0x0015,
        0x0017, 0x0019, 0x001C, 0x001F, 0x0022, 0x0025, 0x0029, 0x002D, 0x0032, 0x0037, 0x003C, 0x0042,
        0x0049, 0x0050, 0x0058, 0x0061, 0x006B, 0x0076, 0x0082, 0x008F, 0x009D, 0x00AD, 0x00BE, 0x00D1,
        0x00E6, 0x00FD, 0x0117, 0x0133, 0x0151, 0x0173, 0x0198, 0x01C1, 0x01EE, 0x0220, 0x0256, 0x0292,
        0x02D4, 0x031C, 0x036C, 0x03C3, 0x0424, 0x048E, 0x0502, 0x0583, 0x0610, 0x06AB, 0x0756, 0x0812,
        0x08E0, 0x09C3, 0x0ABD, 0x0BD0, 0x0CFF, 0x0E4C, 0x0FBA, 0x114C, 0x1307, 0x14EE, 0x1706, 0x1954,
        0x1BDC, 0x1EA5, 0x21B6, 0x2515, 0x28CA, 0x2CDF, 0x315B, 0x364B, 0x3BB9, 0x41B2, 0x4844, 0x4F7E,
        0x5771, 0x602F, 0x69CE, 0x7462, 0x7FFF,
    }

    duty = [8]float64{
        1 - 0.125,
        1 - 0.250,
        1 - 0.375,
        1 - 0.500,
        1 - 0.625,
        1 - 0.750,
        1 - 0.875,
        1 - 0.000,
    }
)

type Channel struct {
	Snd        *Snd
	Mem        *Mem

	Start   bool
	Playing bool

	VolMul     uint32
	VolDiv     uint32
	Panning    uint32
	Duty       uint32
	RepeatMode uint32
	Format     uint32
    Hold bool

	SrcAddr       uint32
	TimerValue    uint16
	StartPosition uint16
	SndLength     uint32

	SamplePos float64

	Samples []int16

	isNoise bool
    lfsr uint32

    isDuty bool
    phase float64
}

func NewChannel(s *Snd) Channel {

	c := Channel{
		Snd: s,
        lfsr: 0x7FFF,
	}

	return c
}

func (c *Channel) GetSample() (int8, int8) {

	if c.Start {
		c.Playing = true
		c.Start = false
		c.SamplePos = 0

		switch c.Format {
		case 2:
			c.DecompressADPCM()
        case 3:
            if c.isNoise {
                c.lfsr = 0x7FFF
            }
		}
	}

	if !c.Playing {
		return 0, 0
	}

	if c.TimerValue == 0 {
		return 0, 0
	}

    sndFreq := float64(c.Snd.sndFrequency)
    if c.Format == 3 && c.isDuty {
        //Each duty cycle consists of eight HIGH or LOW samples,
        //so the sound frequency is 1/8th of the selected sample rate (gbatek)
        sndFreq *= 8
    }

    // playbackRate / sndFreq == step
    playbackRate := BASE_FREQ / float64(-int16(c.TimerValue))
    c.SamplePos += playbackRate / sndFreq

	var sample float64
	switch c.Format {
	case 0: sample = c.GetPCM8()
	case 1: sample = c.GetPCM16()
	case 2: sample = c.GetADPCM()
    case 3: sample = c.GetPSG()
	}

	switch c.VolDiv {
	case 1: sample /= 2
	case 2: sample /= 4
	case 3: sample /= 16
	}

	sample *= float64(c.VolMul)

	l := sample * (float64(127-c.Panning) / 127)
	r := sample * (float64(c.Panning) / 127)

	return int8(l), int8(r)
}

func (c *Channel) GetPCM8() float64 {

	length := (c.SndLength + uint32(c.StartPosition)) * 4
	if uint32(c.SamplePos) >= length {

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

	length := ((c.SndLength) + uint32(c.StartPosition)) * 2
	if uint32(c.SamplePos) >= length {

		if loop := c.RepeatMode == 1; loop {
			c.SamplePos = float64(c.StartPosition * 2)

		} else {
			c.Playing = false
			return 0
		}
	}

    v := int16(uint16(c.Snd.Mem.Read16(
			c.SrcAddr+(uint32(c.SamplePos)*2),
			false)))

	return float64(v) / 32768
}

func (c *Channel) DecompressADPCM() {

	addr := c.SrcAddr
	length := (c.SndLength * 4) + uint32(4*(c.StartPosition))
	c.Samples = make([]int16, 0, length*2)

	head := uint32(c.Snd.Mem.Read32(addr, false))

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

		v := c.Snd.Mem.Read(addr+i, false)

		dec(v & 0xF)
		a := uint16(pcm & 0xFF)
		a |= uint16((pcm>>8)&0xFF) << 8
		c.Samples = append(c.Samples, int16(a))

		dec(v >> 4)
		b := uint16(pcm & 0xFF)
		b |= uint16((pcm>>8)&0xFF) << 8
		c.Samples = append(c.Samples, int16(b))
	}
}

func (c *Channel) GetADPCM() float64 {

	if int(c.SamplePos) >= len(c.Samples) {

        // this needs to be adjusted
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

func (c *Channel) GetPSG() float64 {
    switch {
    case c.isNoise: return c.GetNoise()
    case c.isDuty:  return c.GetWaveDuty()
    default:        return 0
    }
}

func (c *Channel) GetNoise() float64 {

    // untested

    carry := c.lfsr & 1 == 1
    c.lfsr >>= 1

    if carry {
        c.lfsr ^= 0x6000
        return -1
    }

    return 1
}

func (c *Channel) GetWaveDuty() float64 {

    if c.SamplePos >= 1.0 {
        c.SamplePos -= 1.0
    }

    if c.SamplePos < duty[c.Duty] {
        return -1
    }

    return 1
}
