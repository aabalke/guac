package sdl

import (
	"flag"
	"fmt"
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
	C_Transparent25 = sdl.Color{R: 0, G: 0, B: 0, A: 63}
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

    InitMainMenu(s.Renderer, scene, 0)

    //duration := 3 * time.Second
    //InitLoadingScreen(s.Renderer, scene, duration)

    //timer := time.NewTimer(duration - (time.Second / 5))

    //go func() {
    //    <- timer.C
    //    gb.Paused = false
    //}()

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
	c := sdl.Color{R: 228, G: 199, B: 153, A: 255}
    c2 := sdl.Color{R: 255, G: 255, B: 255, A: 255}

	pause := NewGbMenu(renderer, scene, &scene.H, &scene.W, 0, 0, 2, gb, C_Brown)
	containerLayout := Layout{W: 200, H: 400, X: 100, Y: 100, Z: 3}
	container := NewContainer(renderer, pause, containerLayout, C_Transparent, "evenlyVertical")

	text := "mute"
	if gb.Muted {
		text = "unmute"
	}

    container.Add(NewText(renderer, container, Layout{Z: 5}, "resume", 48, c, c2, ""))
	container.Add(NewText(renderer, container, Layout{Z: 5}, text, 48, c, c2, ""))
	container.Add(NewText(renderer, container, Layout{Z: 5}, "exit", 48, c, c2, ""))
	pause.Add(container)
	//pause.Add(NewText(s.Renderer, container, 5, "always save your game in the emulator before exiting", 16))
	scene.Add(pause)

	pause.InitOptions()
}

func InitLoadingScreen(renderer *sdl.Renderer, scene *Scene, duration time.Duration) {
    c := sdl.Color{R: 255, G: 255, B: 255, A: 255}

    loadingScreen := NewLoadingScreen(renderer, scene, &scene.H, &scene.W, 0, 0, 2, C_Green, duration)

	containerLayout := Layout{W: 200, H: 200, X: 100, Y: 100, Z: 3}
	container := NewContainer(renderer, loadingScreen, containerLayout, C_Transparent, "evenlyVertical")
    container.Add(NewText(renderer, container, Layout{Z: 5}, "guac emulator", 48, c, c, ""))

	containerLayout = Layout{W: 200, H: 50, X: 100, Y: 100, Z: 3}
	container2 := NewContainer(renderer, container, containerLayout, C_Transparent, "evenlyVertical")
    container2.Add(NewText(renderer, container2, Layout{Z: 5}, "alpha 0.0.1", 24, c, c, ""))
    container2.Add(NewText(renderer, container2, Layout{Z: 5}, "developed by aaron balke", 24, c, c, ""))

    container.Add(container2)
    loadingScreen.Add(container)
    scene.Add(loadingScreen)

    loadingScreen.InitOptions()
}

type GameData struct {
    Name string
    RomPath string
    SavPath string
    ArtPath string
    Year int
    Console string
}

func InitMainMenu(renderer *sdl.Renderer, scene *Scene, duration time.Duration) {


    gameDatas := []GameData{
        {
            Name: "Pokemon Gold",
            ArtPath: "./art/pokemon_gold.png",
            Year: 1999,
            Console: "Gameboy Color",
        },
        {
            Name: "Pokemon Red",
            ArtPath: "./art/pokemon_red.png",
            Year: 1996,
            Console: "Gameboy",
        },
        {
            Name: "Links Awakening DX",
            ArtPath: "./art/links_awakening_dx.png",
            Year: 1998,
            Console: "Gameboy Color",
        },
        {
            Name: "Oracle of Ages",
            ArtPath: "./art/oracle_of_ages.png",
            Year: 2001,
            Console: "Gameboy Color",
        },
        {
            Name: "Mario Tennis",
            ArtPath: "./art/mario_tennis.png",
            Year: 2000,
            Console: "Gameboy Color",
        },
    }

    menu := NewMainMenu(renderer, scene, &scene.H, &scene.W, 0, 0, 2, C_White, duration)

    //SelectedIdx := 0
    //initialHeight :=
    for i, v := range gameDatas {
        InitGameData(renderer, menu, duration, v, int32(i * 250))
    }

    scene.Add(menu)
}
func InitGameData(renderer *sdl.Renderer, menu *MainMenu, duration time.Duration, data GameData, y int32) {

    x := int32(*menu.W) / 2 - (800 / 2)

    container := NewContainer(renderer, menu, Layout{W: 800, H: 200, Z: 3, Y: y, X: x}, C_Green, "")
    container.Add(NewImage(renderer, container, Layout{W: 200, H: 200, Z: 5}, data.ArtPath, "relativeParent"))
	nestedC2 := NewContainer(renderer, container, Layout{W: 600, H: 200, X: 200, Z: 4}, C_Brown, "relativeParent")
    container.Add(nestedC2)
    nestedC2.Add(NewText(renderer, nestedC2, Layout{X: 25, Y: 50, Z: 5}, data.Name, 48, C_White, C_White, "relativeParent"))

    t := fmt.Sprintf("%s, %d", data.Console, data.Year)
    nestedC2.Add(NewText(renderer, nestedC2, Layout{X: 25, Y: 100, Z: 5}, t, 24, C_White, C_White, "relativeParent"))
    //container.Add(nestedC2)
    container.Add(NewContainer(renderer, container, Layout{W: 800, H: 200, Z: 6}, C_Transparent25, "relativeParent"))

    menu.Add(container)
}
