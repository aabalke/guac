package sdl

import (
	"time"
    "fmt"

	gameboy "github.com/aabalke33/guac/emu/gb"
	"github.com/aabalke33/guac/emu/gba"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"

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

	Gb *gameboy.GameBoy
	Gba *gba.GBA
)

type SDLStruct struct {
	name     string
	Window   *sdl.Window
	Renderer *sdl.Renderer
	DebugWindow   *sdl.Window
	DebugRenderer *sdl.Renderer
}

func (s *SDLStruct) Init(debugger bool) {
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
		//1080,
		//0,
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

    if debugger {

        debugWindow, err := sdl.CreateWindow(
            fmt.Sprintf("%s Debug\n", s.name),
            0,
            0,
            1080,
            1080,
            sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)

        if err != nil {
            panic(err)
        }

        debugRenderer, err := sdl.CreateRenderer(debugWindow, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
        if err != nil {
            panic(err)
        }

        debugRenderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

        s.DebugWindow = debugWindow
        s.DebugRenderer = debugRenderer
    }
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

func (s *SDLStruct) Close(debug bool) {
	sdl.Quit()
	ttf.Quit()
	s.Window.Destroy()
	s.Renderer.Destroy()

    if debug {
        s.DebugWindow.Destroy()
        s.DebugRenderer.Destroy()
    }
}

func (s *SDLStruct) Update(debug bool, romPath string, useSaveState bool) {

    InitSound()
	Gb = gameboy.NewGameBoy()
    Gba = gba.NewGBA()
	defer Gb.Logger.Close()

	w, h := s.Window.GetSize()
	scene := NewScene(s.Renderer, w, h, 10, C_Grey)

    var debugScene *Scene
    if debug {
        debugW, debugH := s.DebugWindow.GetSize()
        debugScene = NewScene(s.DebugRenderer, debugW, debugH, 10, C_Grey)
    }

    if romPath == "" {
        duration := 3 * time.Second
        InitLoadingScreen(s.Renderer, scene, duration)
        InitMainMenu(scene, 0)
    } else {
        ActivateConsole(scene, debugScene, romPath, debug, useSaveState)
    }

	frameTime := time.Second / FPS
	//frameTime := time.Second
	ticker := time.NewTicker(frameTime)

	count := 0
	for i := range ticker.C {

		if !scene.Status.Active || (debug && !debugScene.Status.Active) {
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

		//count = Gb.Update(&scene.Status.Active, count)
        count = Gba.Update(&scene.Status.Active, count)

		s.Renderer.SetDrawColor(0, 0, 0, 255)
		s.Renderer.Clear()
        if debug {
            s.DebugRenderer.SetDrawColor(0, 0, 0, 255)
            s.DebugRenderer.Clear()
        }

		scene.Update(nil)
		scene.View()
		s.Renderer.Present()

        if debug {
            debugScene.Update(nil)
            debugScene.View()
            s.DebugRenderer.Present()
        }
	}
}

func InitSound() {
	sampleRate := beep.SampleRate(44100)
	bufferSize := sampleRate.N(time.Second / 30)
	err := speaker.Init(sampleRate, bufferSize)
	if err != nil {
        panic(err)
	}
}
