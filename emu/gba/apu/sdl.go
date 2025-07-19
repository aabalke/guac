package apu

// typedef unsigned char Uint8;
// typedef signed char int16;
// void Callback(void *userdata, Uint8 *stream, int len);
//import "C"
import (
	//"unsafe"
	//"github.com/veandco/go-sdl2/sdl"
)

const (
//	CPU_FREQ_HZ              = 16777216
//	SND_FREQUENCY            = 32768 // sample rate
//	SND_SAMPLES              = 512
//	SAMP_CYCLES              = (CPU_FREQ_HZ / SND_FREQUENCY)
	BUFF_SIZE                = ((SND_SAMPLES) * 16 * 2)
//	STREAM_LEN               = (2 * 2 * SND_FREQUENCY / 60) - (2*2*SND_FREQUENCY/60)%4
//
)

func InitAudio() {

    if OTO_VERSION {
        return
    }

    //InitApuInstance()

    //spec := sdl.AudioSpec{
    //    Freq: SND_FREQUENCY,
    //    Format: sdl.AUDIO_S16,
    //    Channels: 2,
    //    Samples: STREAM_LEN,
    //    Callback: sdl.AudioCallback(C.Callback),
    //}

    //device, err := sdl.OpenAudioDevice("", false, &spec, nil, 0)
    //if err != nil || device == 0 {
    //    panic(err)
    //}

    //sdl.PauseAudioDevice(device, false)

    return
}

//export Callback
//func Callback(userdata unsafe.Pointer, stream *C.Uint8, length C.int) {
//	n := int(length)
//    buf := unsafe.Slice((*int16)(unsafe.Pointer(stream)), n/2)
//
//	for i := 0; i < n / 2; i += 2 {
//        l, r := ApuInstance.GetSample()
//        buf[i] = l
//        buf[i+1] = r
//	}
//
//    ApuInstance.Sync()
//}
