package snd

type Capture struct {
    Snd *Snd

    Add bool
    ChanSrc bool
    OneShot bool
    PCM8 bool

    Start bool
    Playing bool

    Dest uint32
    Len  uint16

    TimerValue *uint16

    SamplePos float64
}

func NewCaptures(snd *Snd) [2]Capture {

    c := [2]Capture{}

    c[0].Snd = snd
    c[1].Snd = snd

    c[0].TimerValue = &snd.Channels[1].TimerValue
    c[1].TimerValue = &snd.Channels[3].TimerValue

    return c
}

func (c *Capture) Capture(sample float64) {

	if c.Start {
		c.Playing = true
		c.Start = false
		c.SamplePos = 0
	}

	if !c.Playing {
		return
	}

    sndFreq := float64(c.Snd.sndFrequency)
    playbackRate := BASE_FREQ / float64(-int16(*c.TimerValue))
    c.SamplePos += playbackRate / sndFreq

    if c.PCM8 {

        length := uint32(c.Len) * 4
        if uint32(c.SamplePos) >= length {
            if c.OneShot {
                c.Playing = false
                return
            }

            c.SamplePos = float64(0)
        }

        c.Snd.Mem.Write(c.Dest + uint32(c.SamplePos), uint8(int8(sample)), false)

        return
    }

	length := uint32(c.Len) * 2
	if uint32(c.SamplePos) >= length {
        if c.OneShot {
			c.Playing = false
			return
        }

        c.SamplePos = float64(0)
	}

    c.Snd.Mem.Write16(c.Dest + uint32(c.SamplePos)*2, uint16(int16(sample)), false)
}
