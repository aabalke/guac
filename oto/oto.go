package oto

import (
    //"time"
	//"github.com/gopxl/beep"
	//"github.com/gopxl/beep/speaker"
	"github.com/hajimehoshi/oto"
)

var OtoContext *oto.Context
var OtoPlayer *oto.Player


func InitOto() {

	c, err := oto.NewContext(SND_FREQUENCY, 2, 2, STREAM_LEN*3)
	if err != nil {
		panic(err)
	}

	OtoContext = c
    OtoPlayer = OtoContext.NewPlayer()
}

func InitSound() {

    InitOto()

	//sampleRate := beep.SampleRate(44100)
	//bufferSize := sampleRate.N(time.Second / 30)
	//err := speaker.Init(sampleRate, bufferSize)
	//if err != nil {
    //    panic(err)
	//}
}
