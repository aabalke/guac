package nds

import (
	"math"
	"sync"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/utils"
	"github.com/hajimehoshi/ebiten/v2"
)

// For emulated nds, see ppu. Screen is just how it is displayed with the emulator

// layout: "vertical", "horizontal", "hybrid"
// rotation: 0, 90, 180, 270
// sizing: "even", "only top", "only bottom"
// unsetup: gap, emphasized screen

const (
	SCREEN_LAYOUT = iota
	SCREEN_SIZING
	SCREEN_ROTATION
)

const (
	LAYOUT_VERTICAL = iota
	LAYOUT_HORIZONTAL
	LAYOUT_HYBRID
)

const (
	SIZING_EVEN = iota
	SIZING_ONLY_TOP
	SIZING_ONLY_BOTTOM
)

const (
	ROT_0 = iota
	ROT_90
	ROT_180
	ROT_270
)

type Screen struct {
	Mu       sync.Mutex
	Layout   *int
	Sizing   *int
	Rotation *int

	Top, Bottom *ebiten.Image
	opts        ebiten.DrawImageOptions
	TouchGeoM   ebiten.GeoM
}

func NewScreen() *Screen {
	return &Screen{
		Top:      ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT),
		Bottom:   ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT),
		Layout:   &config.Conf.Nds.Screen.Layout,
		Sizing:   &config.Conf.Nds.Screen.Sizing,
		Rotation: &config.Conf.Nds.Screen.Rotation,
	}
}

func (s *Screen) FillScreen(screen *ebiten.Image) {
	switch {
	case *s.Layout == LAYOUT_HYBRID:
		s.FillHybrid(screen)
	case *s.Sizing == SIZING_ONLY_TOP:
		s.FillOnly(screen, s.Top, false)
	case *s.Sizing == SIZING_ONLY_BOTTOM:
		s.FillOnly(screen, s.Bottom, true)
	default:
		s.FillEven(screen, *s.Layout == LAYOUT_HORIZONTAL)
	}
}

func (s *Screen) FillEven(screen *ebiten.Image, horizontal bool) {
	const (
		singleW = float64(SCREEN_WIDTH)
		singleH = float64(SCREEN_HEIGHT)
	)

	canvasW := singleW
	canvasH := singleH * 2

	if horizontal {
		canvasW = singleW * 2
		canvasH = singleH
	}

	var (
		rotation = *s.Rotation
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

	s.opts.GeoM.Reset()

	if horizontal {
		s.opts.GeoM.Translate(0, -singleH/2)
	} else {
		s.opts.GeoM.Translate(-singleW/2, -singleH)
	}

	s.opts.GeoM.Rotate(radians)
	s.opts.GeoM.Scale(scale, scale)
	s.opts.GeoM.Translate(screenW/2, screenH/2)

	s.Mu.Lock()
	screen.DrawImage(s.Top, &s.opts)
	s.Mu.Unlock()

	s.opts.GeoM.Reset()

	if horizontal {
		s.opts.GeoM.Translate(-singleW, -singleH/2)
	} else {
		s.opts.GeoM.Translate(-singleW/2, 0)
	}

	s.opts.GeoM.Rotate(radians)
	s.opts.GeoM.Scale(scale, scale)
	s.opts.GeoM.Translate(screenW/2, screenH/2)

	s.Mu.Lock()
	screen.DrawImage(s.Bottom, &s.opts)
	s.Mu.Unlock()

	s.TouchGeoM = s.opts.GeoM
}

func (s *Screen) FillOnly(screen, image *ebiten.Image, bottom bool) {
	const (
		canvasW = float64(SCREEN_WIDTH)
		canvasH = float64(SCREEN_HEIGHT)
	)

	var (
		rotation = *s.Rotation
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

	s.opts.GeoM.Reset()
	s.opts.GeoM.Translate(-canvasW/2, -canvasH/2)
	s.opts.GeoM.Rotate(radians)
	s.opts.GeoM.Scale(scale, scale)
	s.opts.GeoM.Translate(screenW/2, screenH/2)
	s.Mu.Lock()
	screen.DrawImage(image, &s.opts)
	s.Mu.Unlock()

	if bottom {
		s.TouchGeoM = s.opts.GeoM
	} else {
		s.TouchGeoM = ebiten.GeoM{}
	}
}

func (s *Screen) FillHybrid(screen *ebiten.Image) {
	const (
		singleW = float64(SCREEN_WIDTH)
		singleH = float64(SCREEN_HEIGHT)
	)

	var (
		canvasW  = singleW * 3
		canvasH  = singleH * 2
		rotation = *s.Rotation
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

	s.opts.GeoM.Reset()
	s.opts.GeoM.Translate(singleW*2, 0)
	s.opts.GeoM.Translate(-canvasW/2, -canvasH/2)
	s.opts.GeoM.Rotate(radians)
	s.opts.GeoM.Scale(scale, scale)
	s.opts.GeoM.Translate(screenW/2, screenH/2)
	s.Mu.Lock()
	screen.DrawImage(s.Top, &s.opts)
	s.Mu.Unlock()

	s.opts.GeoM.Reset()
	s.opts.GeoM.Translate(singleW*2, singleH)
	s.opts.GeoM.Translate(-canvasW/2, -canvasH/2)
	s.opts.GeoM.Rotate(radians)
	s.opts.GeoM.Scale(scale, scale)
	s.opts.GeoM.Translate(screenW/2, screenH/2)
	s.Mu.Lock()
	screen.DrawImage(s.Bottom, &s.opts)
	s.TouchGeoM = s.opts.GeoM
	s.Mu.Unlock()

	s.opts.GeoM.Reset()
	s.opts.GeoM.Scale(2, 2)
	s.opts.GeoM.Translate(-canvasW/2, -canvasH/2)
	s.opts.GeoM.Rotate(radians)
	s.opts.GeoM.Scale(scale, scale)
	s.opts.GeoM.Translate(screenW/2, screenH/2)

	s.Mu.Lock()
	if *s.Sizing == SIZING_ONLY_BOTTOM {
		screen.DrawImage(s.Bottom, &s.opts)
		s.TouchGeoM = s.opts.GeoM
	} else {
		screen.DrawImage(s.Top, &s.opts)
	}
	s.Mu.Unlock()
}
