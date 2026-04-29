package ui

import (
	"log"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/oto"

	"fmt"

	"github.com/aabalke/guac/config"
	gameboy "github.com/aabalke/guac/emu/gb"
	"github.com/aabalke/guac/emu/gba"
	"github.com/aabalke/guac/emu/nds"
	"github.com/aabalke/guac/input"
	"github.com/aabalke/guac/utils"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type PageId int

const (
	PAGE_HOME PageId = iota
	PAGE_PAUSE
	PAGE_SETTINGS
)

type Game struct {

	// flags

	ui    *Ui
	nds   *nds.Nds
	gba   *gba.GBA
	gb    *gameboy.GameBoy
	mouse *input.Mouse

	pauseEndFrame uint64

	gamepadIdBuf []ebiten.GamepadID
	gamepadIds   map[ebiten.GamepadID]struct{}

	menuCtx *audio.Context
	emuCtx  *oto.Context

	TargetFps int

	unlimitedFPS bool
	StdFPS       bool

	frame uint64

	paused bool
	muted  bool
	quit   bool
}

type Ui struct {
	res   *Resources
	focus *Focus

	PageId     PageId
	PrevPageId PageId

	ui *ebitenui.UI

	sidebar    *widget.Container
	scrollable *widget.ScrollContainer
	content    *widget.Container
	slider     *widget.Slider
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
	//ebiten.SetVsyncEnabled(!config.Conf.VsyncDisabled)
	ebiten.SetVsyncEnabled(true)

	ebiten.SetCursorMode(ebiten.CursorModeHidden)
	if config.Conf.General.InitFullscreen {
		ebiten.SetFullscreen(true)
	}

	g := NewGame()
	g.ui = NewUi(res)

	// switch based on flags
	NewHome(g)

	err = ebiten.RunGame(g)
	if err != nil {
		log.Print(err)
	}
}

func NewUi(res *Resources) *Ui {
	return &Ui{
		res:   res,
		focus: &Focus{},
	}
}

func NewGame() *Game {

	g := &Game{
		//flags
		emuCtx:       NewAudioContext(),
		mouse:        input.NewMouse(),
		StdFPS:       config.Conf.General.TargetFps == 60,
		gamepadIds:   make(map[ebiten.GamepadID]struct{}),
		gamepadIdBuf: make([]ebiten.GamepadID, 0),
	}

	if !config.Conf.CancelAudioInit {
		g.menuCtx = audio.NewContext(SND_FREQUENCY)
	}

	return g
}

func (g *Game) Layout(outsideWidth int, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func (g *Game) Update() error {

	if config.Conf.General.TargetFps != g.TargetFps {
		g.TargetFps = config.Conf.General.TargetFps
		ebiten.SetTPS(g.TargetFps)
		g.StdFPS = g.TargetFps == 60
	}

	g.Profile()

	justKeys, keys, _, buttons := g.GetInput()

	if g.quit {
		return ebiten.Termination
	}

	switch {
	case g.ui != nil:

		if init := g.frame < 1; init && len(g.gamepadIds) != 0 {
			g.ui.ui.SetFocusedWidget(g.ui.ui.Container.GetFocusers()[0])
		}

		if g.paused && g.frame-g.pauseEndFrame < 10 {
			// pressing select on pause can sometimes input into emulator,
			// this gives time from the pause and emulator starting again
			return nil
		}

		g.ui.ui.Update()

	case g.nds != nil:
		g.nds.InputHandler(justKeys, keys, buttons, g.mouse, g.frame)
		g.nds.Update(g.StdFPS)

	case g.gba != nil:
		g.gba.InputHandler(keys, buttons)
		g.gba.Update(g.StdFPS)

	case g.gb != nil:
		g.gb.InputHandler(keys, buttons)
		g.gb.Update(g.StdFPS)

	}

	g.frame++

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	screen.Fill(config.Conf.Ui.Backdrop)

	defer g.mouse.Draw(screen)

	switch {
	case g.ui != nil:
		g.ui.ui.Draw(screen)
	case g.nds != nil:
		g.nds.Screen.FillScreen(screen)
	case g.gba != nil:
		g.gba.Draw(screen)
	case g.gb != nil:
		g.gb.Draw(screen)
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

	if g.paused && g.ui == nil {
		NewPause(g)
	}

	if !g.paused && (g.nds != nil || g.gba != nil || g.gb != nil) {

		if g.gb != nil {
			g.gb.UpdateFromConfig() // this will get called every pause, better method?
		}

		g.pauseEndFrame = g.frame
		g.ui = nil
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
}

//const (
//	//PRF_START = 1200
//	//PRF_END   = PRF_START + 2000
//	PRF_START = 0
//	PRF_END   = 10000
//)
//
//var t time.Time

func (g *Game) Profile() {

	//if g.flags.Profile && g.frame == PRF_START {
	//	println("starting profiling")
	//	//isProfiling = true
	//	t = time.Now()
	//	pprof.StartCPUProfile(f)

	//}

	//if g.flags.Profile && g.frame >= PRF_END {
	//	dur := time.Since(t).Seconds()

	//	reqDur := (float64(PRF_END-PRF_START) / 60.0)

	//	fmt.Printf("DURATION %.2f seconds. %.2fx faster.\n", time.Since(t).Seconds(), reqDur/dur)
	//	println("ending profiling")
	//	return exit
	//}
}

func (g *Game) InitConsole(file string) {
	switch romType := utils.GetRomType(file); romType {

	case utils.GB:
		g.gb = gameboy.NewGameBoy(file, g.emuCtx)
		g.ui = nil
		if g.muted {
			g.gb.ToggleMute()
		}

	case utils.GBA:
		g.gba = gba.NewGBA(file, g.emuCtx)
		g.ui = nil
		if g.muted {
			g.gba.ToggleMute()
		}

	case utils.NDS:
		g.nds = nds.NewNds(file, g.emuCtx)
		g.ui = nil
		if g.muted {
			g.nds.ToggleMute()
		}
	}
}
