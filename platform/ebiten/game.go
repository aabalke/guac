package ui

import (
	"context"
	"fmt"
	"log"

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
	ui       *Ui
	Bus      *bus.EventBus
	nds      *nds.Nds
	gba      *gba.GBA
	gb       *gb.GameBoy
	emuClose func()

	audioCtx     *audio.Context
	pauseEndTick int64
	paused       bool
	muted        bool
	quit         bool
}

type Ui struct {
	gamepadIdBuf []ebiten.GamepadID
	gamepadIds   map[ebiten.GamepadID]struct{}
	res          *Resources
	focus        *Focus
	ui           *ebitenui.UI
	toast        *Toast
	sidebar      *widget.Container
	scrollable   *widget.ScrollContainer
	content      *widget.Container
	slider       *widget.Slider
	keyboard     *Keyboard
	ActiveInputs int
	PageId       PageId
	PrevPageId   PageId
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
	var audioCtx *audio.Context
	if !config.Conf.Profile.Enabled {
		audioCtx = audio.NewContext(config.Conf.General.SampleRate)
	}

	g := &Game{
		audioCtx: audioCtx,
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

	switch {
	case g.quit:
		return ebiten.Termination
	case g.ui.ui != nil:

		if g.ui.ActiveInputs <= 0 {
			justButtons, buttons := g.GetGamepadButtons()
			g.ButtonInput(justButtons, buttons)
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

func (g *Game) Draw(dst *ebiten.Image) {
	dst.Fill(config.Conf.Ui.Backdrop)

	switch {
	case g.ui.ui != nil:
		g.ui.ui.Draw(dst)
	case g.gb != nil:
		g.gb.Draw(dst)
	case g.gba != nil:
		g.gba.Draw(dst)
	case g.nds != nil:
		g.nds.Screen.Draw(dst)
	}

	if g.ui.toast.enabled {
		g.ui.toast.ui.Draw(dst)
	}

	if config.Conf.General.ShowFps {

		target := config.Conf.General.TargetFps

		switch {
		case g.ui.ui != nil:
			ebitenutil.DebugPrint(dst, fmt.Sprintf("Target FPS: %d.00 Engine FPS: %.2f, TPS: %.2f", target, ebiten.ActualFPS(), ebiten.ActualTPS()))
		case g.gb != nil:
			ebitenutil.DebugPrint(dst, fmt.Sprintf("Target FPS: %d.00 Engine FPS: %.2f, TPS: %.2f\nEmulated FPS: %.2f Frame: %08d", target, ebiten.ActualFPS(), ebiten.ActualTPS(), g.gb.FPS(), g.gb.Frame()))
		case g.gba != nil:
			ebitenutil.DebugPrint(dst, fmt.Sprintf("Target FPS: %d.00 Engine FPS: %.2f, TPS: %.2f\nEmulated FPS: %.2f Frame: %08d", target, ebiten.ActualFPS(), ebiten.ActualTPS(), g.gba.FPS(), g.gba.Frame()))
		case g.nds != nil:
			ebitenutil.DebugPrint(dst, fmt.Sprintf("Target FPS: %d.00 Engine FPS: %.2f, TPS: %.2f\nEmulated FPS: %.2f Frame: %08d", target, ebiten.ActualFPS(), ebiten.ActualTPS(), g.nds.FPS(), g.nds.Frame()))
		default:
			ebitenutil.DebugPrint(dst, fmt.Sprintf("Target FPS: %d.00 Engine FPS: %.2f, TPS: %.2f", target, ebiten.ActualFPS(), ebiten.ActualTPS()))
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

		if g.gb != nil {
			g.gb.UpdateFromConfig() // this will get called every pause, better method?
		}
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

func (g *Game) InitConsole(file string) bool {
	romType := utils.GetRomType(file)
	if romType == utils.GB || romType == utils.GBA || romType == utils.NDS {

		g.ui.ui = nil
		ctx, cancel := context.WithCancel(context.Background())
		g.emuClose = cancel

		switch romType := utils.GetRomType(file); romType {
		case utils.GB:
			g.gb = gb.NewGameBoy(g.audioCtx, file, g.muted)
			go g.gb.Run(ctx, g.Bus)

		case utils.GBA:
			g.gba = gba.NewGBA(g.audioCtx, file, g.muted)
			go g.gba.Run(ctx, g.Bus)

		case utils.NDS:
			g.nds = nds.NewNds(g.audioCtx, file, g.muted)
			go g.nds.Run(ctx, g.Bus)
		}

		return true
	}

	if romType == utils.NONE_ZIP {
		g.ui.toast.AddMessage("Invalid Zip File")
	}

	return false
}
