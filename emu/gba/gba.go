package gba

const (
    SCREEN_WIDTH = 240
    SCREEN_HEIGHT = 160
)

type GBA struct {

    Screen [SCREEN_WIDTH][SCREEN_HEIGHT]uint32
    Pixels *[]byte

    Paused bool
    Muted bool
}

func NewGBA() *GBA {

    gba := GBA{}
    pixels := make([]byte, SCREEN_WIDTH*SCREEN_HEIGHT*4)
    gba.Pixels = &pixels

    return &gba
}

func (gba *GBA) GetSize() (int32, int32) {
    return SCREEN_HEIGHT, SCREEN_WIDTH
}


func (gba *GBA) GetPixels() []byte {
    return *gba.Pixels
}

func (gba *GBA) ToggleMute() bool {
    gba.Muted = !gba.Muted
    return gba.Muted
}

func (gba *GBA) TogglePause() bool {
    gba.Paused = !gba.Paused
    return gba.Paused
}

func (gba *GBA) Close() {
    gba.Muted = true
    gba.Paused = true
}

func (gba *GBA) LoadGame(path string) {
    println("Loading GBA Game")
}
