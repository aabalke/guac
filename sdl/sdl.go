package sdl

import (
	"flag"
	"time"

	comp "github.com/aabalke33/go-sdl2-components/Components"
	gameboy "github.com/aabalke33/guac/emu/gb"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const FPS = 60.0

type SDLStruct struct {
    name string

    Window *sdl.Window
    Renderer *sdl.Renderer
}

func (s *SDLStruct) Init() {
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
		800,
		600,
		sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)
		//sdl.WINDOW_SHOWN)

	if err != nil {
		panic(err)
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		panic(err)
	}

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
    println("Controller Attached")
}

func (s *SDLStruct) Close() {
	sdl.Quit()
	ttf.Quit()
	s.Window.Destroy()
	s.Renderer.Destroy()
}

func (s *SDLStruct) Update() {

    romPath := flag.String("r", "", "rom path")
    debug := flag.Bool("debug", false, "debug")
    maxInstr := flag.Int("i", 0, "max instruction count")
    flag.Parse()

	emu := gameboy.NewGameBoy()
	emu.LoadGame(*romPath)
    defer emu.Logger.Close()

	w, h := s.Window.GetSize()

	scene := comp.NewScene(s.Renderer, w, h, 10, sdl.Color{
		R: 25, G: 25, B: 25, A: 255},
	)

	scene.Add(NewEmulatorFrame(s.Renderer, scene, 160.0/144, &scene.H, 0, 0, 1, emu))
    //var debug_h int32 = 128
	//scene.Add(NewDebugFrame(s.Renderer, scene, 1, &scene.H, 0, 0, 2, emu))

    if *debug && *maxInstr == 0 {
        panic("In debug mode max instruction count is required")
    }

	frameTime := time.Second / FPS
	ticker := time.NewTicker(frameTime)
    count := 0
    //switched := false
    for i := range ticker.C {

        //if emu.DoubleSpeed && !switched {
        //    switched = true
        //    emu.MemoryBus.Memory[0xFF26] = 0x80
        //}

		if !scene.Active {
			ticker.Stop()
			break
		}

        // free inactive components every few seconds
        if i.UnixMicro() % 7 == 0 {
            scene.DeleteInactive()
        }

		count = emu.Update(&scene.Active, count)

		s.Renderer.SetDrawColor(0, 0, 0, 255)
		s.Renderer.Clear()
		scene.Update(1/FPS, nil)
		scene.View(s.Renderer)
		s.Renderer.Present()
	}

}
