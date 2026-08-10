package ui

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/aabalke/guac/common/bus"
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/gb"
	"github.com/aabalke/guac/emu/gba"
	"github.com/aabalke/guac/emu/nds"
	"github.com/aabalke/guac/utils"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type PageId int

const (
	PAGE_HOME PageId = iota
	PAGE_PAUSE
	PAGE_SETTINGS
	PAGE_KEYBOARD
)

type Game struct {
	ui          *Ui
	Bus         *bus.EventBus
	DrawOptions ebiten.DrawImageOptions
	nds         *nds.Nds
	gba         *gba.GBA
	gb          *gb.GameBoy
	emuClose    func()

	audioCtx     *audio.Context
	pauseEndTick int64
	TargetFps    int
	paused       bool
	muted        bool
	quit         bool
}

type Ui struct {
	gamepadIdBuf []ebiten.GamepadID
	gamepadIds   map[ebiten.GamepadID]struct{}
	res          *Resources
	focus        *Focus
	ActiveInputs int

	PageId     PageId
	PrevPageId PageId

	ui    *ebitenui.UI
	toast *Toast

	sidebar    *widget.Container
	scrollable *widget.ScrollContainer
	content    *widget.Container
	slider     *widget.Slider
	keyboard   *Keyboard
}

func StartEngine() {
	res, err := NewUIResources()
	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("guac emulator")
	ebiten.SetWindowIcon(res.icon)
	ebiten.SetWindowSize(256*4, 192*4)

	ebiten.SetVsyncEnabled(config.Conf.General.Vsync)

	if config.Conf.General.InitFullscreen {
		ebiten.SetFullscreen(true)
	}

	g := NewGame(res)

	if ok := g.InitConsole(config.Conf.General.RomPath); !ok {
		NewHome(g)
	}

	err = ebiten.RunGame(g)
	if err != nil {
		log.Print(err)
	}
}

func NewGame(res *Resources) *Game {
	g := &Game{
		audioCtx: audio.NewContext(config.Conf.General.SampleRate),
		muted:    config.Conf.General.Muted,
		ui: &Ui{
			gamepadIds:   make(map[ebiten.GamepadID]struct{}),
			gamepadIdBuf: make([]ebiten.GamepadID, 0),
			focus:        &Focus{},
			res:          res,
			toast:        NewToast(res),
			keyboard:     NewKeyboard(res, res.localization.Settings.Ui.Alphabet),
		},
		Bus: bus.NewEventBus(),
	}

	ebiten.SetVsyncEnabled(config.Conf.General.Vsync)

	return g
}

func (g *Game) Layout(outsideWidth int, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func (g *Game) Update() error {
	g.ui.toast.Update()

	//g.Profile()

	switch {
	case g.quit:
		return ebiten.Termination
	case g.ui.ui != nil:

		justButtons, buttons := g.GetGamepadButtons()
		g.ButtonInput(justButtons, buttons)
		if g.ui.ActiveInputs <= 0 {
			justKeys := inpututil.AppendJustPressedKeys([]ebiten.Key{})
			g.HandleGlobalInputs(justKeys, justButtons, buttons)
		}

		if ebiten.Tick() < 1 &&
			len(g.ui.gamepadIds) != 0 &&
			g.ui.ui != nil && g.ui.ui.Container != nil &&
			len(g.ui.ui.Container.GetFocusers()) != 0 {
			g.ui.ui.SetFocusedWidget(g.ui.ui.Container.GetFocusers()[0])
		}

		if g.paused && ebiten.Tick()-g.pauseEndTick < 10 {
			// pressing select on pause can sometimes input into emulator,
			// this gives time from the pause and emulator starting again
			return nil
		}

		if g.ui.scrollable != nil && len(g.ui.gamepadIds) != 0 {
			g.ui.focus.KeepFocusedInView(g.ui.slider, g.ui.ui)
		}

		if g.ui.ui != nil {
			g.ui.ui.Update()
		}

	default:
		justKeys := inpututil.AppendJustPressedKeys([]ebiten.Key{})
		keys := inpututil.AppendPressedKeys([]ebiten.Key{})
		justButtons, buttons := g.GetGamepadButtons()
		g.HandleGlobalInputs(justKeys, justButtons, buttons)

		g.Bus.Publish(bus.Event{
			Type: bus.INPUT,
			Data: bus.InputData{
				Keys:        keys,
				JustKeys:    justKeys,
				Buttons:     buttons,
				JustButtons: justButtons,
			},
		})
	}

	return nil
}

const (
	ROT_0 = iota
	ROT_90
	ROT_180
	ROT_270
)

const (
	RAD_0   = float64(math.Pi/180) * 0
	RAD_90  = float64(math.Pi/180) * 90
	RAD_180 = float64(math.Pi/180) * 180
	RAD_270 = float64(math.Pi/180) * 270
)

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(config.Conf.Ui.Backdrop)

	switch {
	case g.ui.ui != nil:
		g.ui.ui.Draw(screen)
	case g.gb != nil:

		const (
			width  = 160
			height = 144
		)

		var (
			sw      = float64(screen.Bounds().Dx())
			sh      = float64(screen.Bounds().Dy())
			scale   = utils.ScaleImage(sw, sh, width, height)
			offsetX = (sw - (width * scale)) / 2
			offsetY = (sh - (height * scale)) / 2
		)

		g.DrawOptions.GeoM.Reset()
		g.DrawOptions.GeoM.Scale(scale, scale)
		g.DrawOptions.GeoM.Translate(offsetX, offsetY)
		g.gb.Mu.Lock()
		screen.DrawImage(g.gb.Image, &g.DrawOptions)
		g.gb.Mu.Unlock()

	case g.gba != nil:

		const (
			canvasW = float64(gba.SCREEN_WIDTH)
			canvasH = float64(gba.SCREEN_HEIGHT)
		)

		var (
			rotation = config.Conf.Gba.Rotation
			radians  = float64(rotation) * math.Pi / 2
			screenW  = float64(screen.Bounds().Dx())
			screenH  = float64(screen.Bounds().Dy())
			fitW     = canvasW
			fitH     = canvasH
		)

		if rot := rotation == ROT_90 || rotation == ROT_270; rot {
			fitW, fitH = canvasH, canvasW
		}

		scale := utils.ScaleImage(screenW, screenH, fitW, fitH)

		g.DrawOptions.GeoM.Reset()
		g.DrawOptions.GeoM.Translate(-canvasW/2, -canvasH/2)
		g.DrawOptions.GeoM.Rotate(radians)
		g.DrawOptions.GeoM.Scale(scale, scale)
		g.DrawOptions.GeoM.Translate(screenW/2, screenH/2)

		g.gba.Mu.Lock()
		screen.DrawImage(g.gba.Image, &g.DrawOptions)
		g.gba.Mu.Unlock()

	case g.nds != nil:
		g.nds.Screen.FillScreen(screen)
	}

	if g.ui.toast.enabled {
		g.ui.toast.ui.Draw(screen)
	}

	if config.Conf.General.ShowFps {

		target := config.Conf.General.TargetFps

		switch {
		case g.ui.ui != nil:
			ebitenutil.DebugPrint(screen, fmt.Sprintf("Target FPS: %d.00 Engine FPS: %.2f, TPS: %.2f", target, ebiten.ActualFPS(), ebiten.ActualTPS()))
		case g.gb != nil:
			ebitenutil.DebugPrint(screen, fmt.Sprintf("Target FPS: %d.00 Engine FPS: %.2f, TPS: %.2f\nEmulated FPS: %.2f Frame: %08d", target, ebiten.ActualFPS(), ebiten.ActualTPS(), g.gb.FPS(), g.gb.Frame()))
		case g.gba != nil:
			ebitenutil.DebugPrint(screen, fmt.Sprintf("Target FPS: %d.00 Engine FPS: %.2f, TPS: %.2f\nEmulated FPS: %.2f Frame: %08d", target, ebiten.ActualFPS(), ebiten.ActualTPS(), g.gba.FPS(), g.gba.Frame()))
		case g.nds != nil:
			ebitenutil.DebugPrint(screen, fmt.Sprintf("Target FPS: %d.00 Engine FPS: %.2f, TPS: %.2f\nEmulated FPS: %.2f Frame: %08d", target, ebiten.ActualFPS(), ebiten.ActualTPS(), g.nds.FPS(), g.nds.Frame()))
		default:
			ebitenutil.DebugPrint(screen, fmt.Sprintf("Target FPS: %d.00 Engine FPS: %.2f, TPS: %.2f", target, ebiten.ActualFPS(), ebiten.ActualTPS()))
		}
	}
}

func (g *Game) TogglePause() {
	if !g.paused && (g.nds == nil && g.gba == nil && g.gb == nil) {
		return
	}

	g.paused = !g.paused

	g.Bus.Publish(bus.Event{
		Type: bus.PAUSE,
		Data: g.paused,
	})

	if g.paused && g.ui.ui == nil {
		NewPause(g)
	}

	if !g.paused && (g.nds != nil || g.gba != nil || g.gb != nil) {
		g.pauseEndTick = ebiten.Tick()
		g.ui.ui = nil
	}
}

func (g *Game) ToggleMute() {
	g.muted = !g.muted

	g.Bus.Publish(bus.Event{
		Type: bus.MUTE,
		Data: g.muted,
	})

	if g.muted {
		g.ui.toast.AddMessage(g.ui.res.localization.Toast.Muted)
	} else {
		g.ui.toast.AddMessage(g.ui.res.localization.Toast.Unmuted)
	}
}

//var (
//	t time.Time
//	f *os.File
//)

//const UNLIMITED_FPS = 0x1800

//func (g *Game) Profile() {
//	p := &config.Conf.Profile
//
//	if !p.Enabled {
//		return
//	}
//
//	if ebiten.Tick() == p.StartTick {
//
//		if g.gb != nil {
//			g.gb.Apu.ToggleMute(true)
//		}
//		if g.gba != nil {
//			g.gba.Apu.ToggleMute(true)
//		}
//		if g.nds != nil {
//			g.nds.ToggleMute(true)
//		}
//
//		ebiten.SetTPS(UNLIMITED_FPS)
//
//		var err error
//		f, err = os.Create(p.FilePath)
//		if err != nil {
//			panic(err)
//		}
//
//		println("starting profiler")
//
//		pprof.StartCPUProfile(f)
//		t = time.Now()
//	}
//
//	if ebiten.Tick() >= p.EndTick {
//		dur := time.Since(t).Seconds()
//
//		reqDur := (float64(p.EndTick-p.StartTick) / 60.0)
//
//		fmt.Printf("DURATION %.2f seconds. %.2fx faster.\n", time.Since(t).Seconds(), reqDur/dur)
//
//		pprof.StopCPUProfile()
//		f.Close()
//
//		println("ending profiling")
//		g.quit = true
//	}
//}

func (g *Game) InitConsole(file string) bool {
	switch romType := utils.GetRomType(file); romType {
	case utils.GB:

		g.ui.ui = nil
		ctx, cancel := context.WithCancel(context.Background())
		g.emuClose = cancel

		g.gb = gb.NewGameBoy(g.audioCtx, file, g.muted)
		go g.gb.Run(ctx, g.Bus)
		return true

	case utils.GBA:

		g.ui.ui = nil
		ctx, cancel := context.WithCancel(context.Background())
		g.emuClose = cancel

		g.gba = gba.NewGBA(g.audioCtx, file, g.muted)
		go g.gba.Run(ctx, g.Bus)
		return true

	case utils.NDS:

		g.ui.ui = nil
		ctx, cancel := context.WithCancel(context.Background())
		g.emuClose = cancel

		g.nds = nds.NewNds(g.audioCtx, file, g.muted)
		go g.nds.Run(ctx, g.Bus)

		return true
	default:
		return false
	}
}
