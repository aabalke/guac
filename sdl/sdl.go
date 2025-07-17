package sdl

import (
	"time"

	gameboy "github.com/aabalke33/guac/emu/gb"
	"github.com/aabalke33/guac/emu/gba"
	"github.com/aabalke33/guac/oto"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"

	"os"
	"runtime/pprof"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
)

const (
	FPS = 60.0
)

var (
	C_Brown         = sdl.Color{R: 194, G: 138, B: 51, A: 255}
	C_Brown50       = sdl.Color{R: 194, G: 138, B: 51, A: 127}
	C_Green         = sdl.Color{R: 101, G: 163, B: 13, A: 255}
	C_White         = sdl.Color{R: 255, G: 255, B: 255, A: 255}
	C_Black         = sdl.Color{R: 0, G: 0, B: 0, A: 255}
	C_Grey          = sdl.Color{R: 25, G: 25, B: 25, A: 255}
	C_Transparent   = sdl.Color{R: 0, G: 0, B: 0, A: 0}
	C_Transparent50 = sdl.Color{R: 0, G: 0, B: 0, A: 127}

	gbConsole *gameboy.GameBoy
	gbaConsole *gba.GBA

    isGBA bool
    isGB bool
)

type SDLStruct struct {
	name     string
	Window   *sdl.Window
	Renderer *sdl.Renderer
	DebugWindow   *sdl.Window
	DebugRenderer *sdl.Renderer
}

func (s *SDLStruct) Init() {

    oto.InitOto()
    InitSound()

	err := sdl.Init(sdl.INIT_EVERYTHING)
	err = ttf.Init()

	if err != nil {
		panic(err)
	}

	s.initController()

	window, err := sdl.CreateWindow(
		s.name,
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		1280,
		720,
		sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)

	if err != nil {
		panic(err)
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		panic(err)
	}

	renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

	s.Window = window
	s.Renderer = renderer

}

func (s *SDLStruct) initController() {
	var controller *sdl.GameController

	if sdl.NumJoysticks() <= 0 {
		return
	}

	if !sdl.IsGameController(0) {
		return
	}

	controller = sdl.GameControllerOpen(0)

	if !controller.Attached() {
		return
	}

	//println("Controller Attached")
}

func (s *SDLStruct) Close() {
	sdl.Quit()
	ttf.Quit()
	s.Window.Destroy()
	s.Renderer.Destroy()
}

func (s *SDLStruct) Run(romPath string, profile bool) {

	w, h := s.Window.GetSize()
	scene := NewScene(s.Renderer, w, h, 10, C_Grey)

    if romPath == "" {
        s.Menu(scene)
    } else {
        ActivateConsole(scene, romPath)
    }

    if profile {
        f, err := os.Create("cpu.prof")
        if err != nil {
            panic(err)
        }

        pprof.StartCPUProfile(f)

        var count int
        for range 1000 {
            s.Update(scene, count)
            count++
        }

        pprof.StopCPUProfile()
        f.Close()
    }

    var count int

	frameTime := time.Second / (FPS)
	ticker := time.NewTicker(frameTime)

	for i := range ticker.C {

		if !scene.Status.Active {
			ticker.Stop()
			break
		}

		// free inactive components every few seconds
		if i.UnixMicro()%7 == 0 {
			scene.DeleteInactive()
		}

		if i.UnixMicro()%13 == 0 {
            s.initController()
        }

        s.Update(scene, count)

        count++
	}
}

func (s *SDLStruct) Menu(scene *Scene) {

	frameTime := time.Second / (FPS)
	ticker := time.NewTicker(frameTime)
    duration := 3 * time.Second
    InitLoadingScreen(s.Renderer, scene, duration)
    InitMainMenu(scene, 0)

    for range ticker.C {
        s.Renderer.SetDrawColor(0, 0, 0, 255)
        s.Renderer.Clear()
        scene.Update(nil)
        scene.View()
        s.Renderer.Present()

        if isGB || isGBA {
            return
        }
    }
}

func (s *SDLStruct) Update(scene *Scene, count int) {

    if isGBA {
        gbaConsole.Update(&scene.Status.Active, count)
    }

    if isGB {
        gbConsole.Update(&scene.Status.Active, count)
    }

    s.Renderer.SetDrawColor(0, 0, 0, 255)
    s.Renderer.Clear()
    scene.Update(nil)
    scene.View()
    s.Renderer.Present()
}

func InitSound() {
	sampleRate := beep.SampleRate(44100)
	bufferSize := sampleRate.N(time.Second / 30)
	err := speaker.Init(sampleRate, bufferSize)
	if err != nil {
        panic(err)
	}

}
