package oto

import "github.com/hajimehoshi/oto"

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
