package sdl

import (
	"flag"
	"time"

	gameboy "github.com/aabalke33/guac/emu/gb"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	FPS = 60.0
)

var (
	C_Brown       = sdl.Color{R: 194, G: 138, B: 51, A: 255}
	C_Green       = sdl.Color{R: 101, G: 163, B: 13, A: 255}
	C_White       = sdl.Color{R: 255, G: 255, B: 255, A: 255}
	C_Black       = sdl.Color{R: 0, G: 0, B: 0, A: 255}
	C_Grey        = sdl.Color{R: 25, G: 25, B: 25, A: 255}
	C_Transparent = sdl.Color{R: 0, G: 0, B: 0, A: 0}
)

type SDLStruct struct {
	name     string
	Window   *sdl.Window
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

	gb := gameboy.NewGameBoy()
	gb.LoadGame(*romPath)
	defer gb.Logger.Close()

	w, h := s.Window.GetSize()

	scene := NewScene(s.Renderer, w, h, 10, C_Grey)

	scene.Add(NewGbFrame(s.Renderer, scene, 160.0/144, &scene.H, 0, 0, 1, gb))

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

		if !scene.Status.Active {
			ticker.Stop()
			break
		}

		// free inactive components every few seconds
		if i.UnixMicro()%7 == 0 {
			scene.DeleteInactive()
		}

		count = gb.Update(&scene.Status.Active, count)

		s.Renderer.SetDrawColor(0, 0, 0, 255)
		s.Renderer.Clear()
		scene.Update(nil)
		scene.View()
		s.Renderer.Present()
	}
}

func InitPauseMenu(renderer *sdl.Renderer, scene *Scene, gb *gameboy.GameBoy) {
	pause := NewGbMenu(renderer, scene, &scene.H, &scene.W, 0, 0, 2, gb, C_Brown)
	containerLayout := Layout{W: 200, H: 400, X: 100, Y: 100, Z: 3}
	container := NewContainer(renderer, pause, containerLayout, gb, C_Transparent)

	text := "mute"
	if gb.Muted {
		text = "unmute"
	}

	container.Add(NewText(renderer, container, 5, "resume", 48))
	container.Add(NewText(renderer, container, 5, text, 48))
	container.Add(NewText(renderer, container, 5, "exit", 48))
	pause.Add(container)
	//pause.Add(NewText(s.Renderer, container, 5, "always save your game in the emulator before exiting", 16))
	scene.Add(pause)

	pause.InitOptions()
}
