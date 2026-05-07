package ui

import (
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/oto"

	"fmt"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/gb"
	"github.com/aabalke/guac/emu/gba"
	"github.com/aabalke/guac/emu/nds"
	"github.com/aabalke/guac/input"
	"github.com/aabalke/guac/utils"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type PageId int

const (
	PAGE_HOME PageId = iota
	PAGE_PAUSE
	PAGE_SETTINGS
	PAGE_KEYBOARD
)

type Game struct {

	// flags

	ui       *Ui
	nds      *nds.Nds
	gba      *gba.GBA
	gb       *gb.GameBoy
	mouse    *input.Mouse
	audioCtx *oto.Context

	gamepadIdBuf []ebiten.GamepadID
	gamepadIds   map[ebiten.GamepadID]struct{}

	pauseEndTick int64
	TargetFps    int
	vsync        bool

	paused bool
	muted  bool
	quit   bool
}

type Ui struct {
	res   *Resources
	focus *Focus

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
		audioCtx:     NewAudioContext(),
		mouse:        input.NewMouse(),
		TargetFps:    config.Conf.General.TargetFps,
		vsync:        config.Conf.General.Vsync,
		gamepadIds:   make(map[ebiten.GamepadID]struct{}),
		gamepadIdBuf: make([]ebiten.GamepadID, 0),
		ui: &Ui{
			res:      res,
			focus:    &Focus{},
			toast:    NewToast(res),
			keyboard: NewKeyboard(res),
		},
	}

	return g
}

func (g *Game) Layout(outsideWidth int, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func (g *Game) Update() error {

	g.ui.toast.Update()

	if config.Conf.General.TargetFps != g.TargetFps {
		g.TargetFps = config.Conf.General.TargetFps
		ebiten.SetTPS(g.TargetFps)
	}

	if config.Conf.General.Vsync != g.vsync {
		g.vsync = config.Conf.General.Vsync
		ebiten.SetVsyncEnabled(g.vsync)
	}

	g.Profile()

	justKeys, keys, _, buttons := g.GetInput()

	switch {
	case g.quit:
		return ebiten.Termination
	case g.ui.ui != nil:

		if ebiten.Tick() < 1 &&
			len(g.gamepadIds) != 0 &&
			g.ui.ui != nil && g.ui.ui.Container != nil &&
			len(g.ui.ui.Container.GetFocusers()) != 0 {
			g.ui.ui.SetFocusedWidget(g.ui.ui.Container.GetFocusers()[0])
		}

		if g.paused && ebiten.Tick()-g.pauseEndTick < 10 {
			// pressing select on pause can sometimes input into emulator,
			// this gives time from the pause and emulator starting again
			return nil
		}

		if g.ui.scrollable != nil && len(g.gamepadIds) != 0 {
			g.ui.focus.KeepFocusedInView(g.ui.slider)
		}

		g.ui.ui.Update()

	case g.nds != nil:
		g.nds.InputHandler(justKeys, keys, buttons, g.mouse, uint64(ebiten.Tick()))
		g.nds.Update(g.TargetFps == 60)

	case g.gba != nil:
		g.gba.InputHandler(keys, buttons)
		g.gba.Update(g.TargetFps == 60)

	case g.gb != nil:
		g.gb.InputHandler(keys, buttons)
		g.gb.Update(g.TargetFps == 60)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	screen.Fill(config.Conf.Ui.Backdrop)

	switch {
	case g.ui.ui != nil:
		g.ui.ui.Draw(screen)
	case g.nds != nil:
		g.nds.Screen.FillScreen(screen)
	case g.gba != nil:
		g.gba.Draw(screen)
	case g.gb != nil:
		g.gb.Draw(screen)
	}

	if g.ui.toast.enabled {
		g.ui.toast.ui.Draw(screen)
	}

	if config.Conf.General.ShowFps {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %.2f, TPS: %.2f", ebiten.ActualFPS(), ebiten.ActualTPS()))
	}
}

func (g *Game) TogglePause() {

	if !g.paused && (g.nds == nil && g.gba == nil && g.gb == nil) {
		return
	}

	g.paused = !g.paused

	switch {
	case g.nds != nil:
		g.nds.TogglePause()
	case g.gba != nil:
		g.gba.TogglePause()
	case g.gb != nil:
		g.gb.TogglePause()
	}

	if g.paused && g.ui.ui == nil {
		NewPause(g)
	}

	if !g.paused && (g.nds != nil || g.gba != nil || g.gb != nil) {

		if g.gb != nil {
			g.gb.UpdateFromConfig() // this will get called every pause, better method?
		}

		g.pauseEndTick = ebiten.Tick()
		g.ui.ui = nil
	}
}

func (g *Game) ToggleMute() {

	g.muted = !g.muted

	switch {
	case g.nds != nil:
		g.nds.ToggleMute()
	case g.gba != nil:
		g.gba.ToggleMute()
	case g.gb != nil:
		g.gb.ToggleMute()
	}

	if g.muted {
		g.ui.toast.AddMessage(g.ui.res.localization.Toast.Muted)
	} else {
		g.ui.toast.AddMessage(g.ui.res.localization.Toast.Unmuted)
	}
}

var (
	t time.Time
	f *os.File
)

const UNLIMITED_FPS = 0x1000

func (g *Game) Profile() {

	p := &config.Conf.Profile

	if !p.Enabled {
		return
	}

	if ebiten.Tick() == p.StartTick {

		if g.gb != nil {
			g.gb.Muted = true
		}
		if g.gba != nil {
			g.gba.Muted = true
		}
		if g.nds != nil {
			g.nds.Muted = true
		}

		ebiten.SetTPS(UNLIMITED_FPS)

		var err error
		f, err = os.Create(p.FilePath)
		if err != nil {
			panic(err)
		}

		println("starting profiler")

		pprof.StartCPUProfile(f)
		t = time.Now()
	}

	if ebiten.Tick() >= p.EndTick {
		dur := time.Since(t).Seconds()

		reqDur := (float64(p.EndTick-p.StartTick) / 60.0)

		fmt.Printf("DURATION %.2f seconds. %.2fx faster.\n", time.Since(t).Seconds(), reqDur/dur)

		pprof.StopCPUProfile()
		f.Close()

		println("ending profiling")
		g.quit = true
	}
}

func (g *Game) InitConsole(file string) bool {
	switch romType := utils.GetRomType(file); romType {
	case utils.GB:
		g.gb = gb.NewGameBoy(file, g.audioCtx)
		g.ui.ui = nil
		if g.muted {
			g.gb.ToggleMute()
		}
		return true

	case utils.GBA:
		g.gba = gba.NewGBA(file, g.audioCtx)
		g.ui.ui = nil
		if g.muted {
			g.gba.ToggleMute()
		}
		return true

	case utils.NDS:
		g.nds = nds.NewNds(file, g.audioCtx)
		g.ui.ui = nil
		if g.muted {
			g.nds.ToggleMute()
		}
		return true
	default:
		return false
	}
}
