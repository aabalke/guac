package main

import "github.com/hajimehoshi/oto"


const (
	SND_FREQUENCY            = 32768 // sample rate
	//SND_FREQUENCY            = 44100 // sample rate
	//SND_FREQUENCY            = 48000 // sample rate
	STREAM_LEN               = (2 * 2 * SND_FREQUENCY / 60) - (2*2*SND_FREQUENCY/60)%4
)

func NewAudioContext() *oto.Context {

	c, err := oto.NewContext(SND_FREQUENCY, 2, 2, STREAM_LEN*3)
	if err != nil {
		panic(err)
	}

    return c
}
