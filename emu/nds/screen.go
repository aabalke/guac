package nds

import (
	"math"

	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2"
)

// For emulated nds, see ppu. Screen is just how it is displayed with the emulator

// layout: "vertical", "horizontal", "hybrid"
// rotation: 0, 90, 180, 270
// sizing: "even", "only top", "only bottom"

// unsetup: gap, emphasized screen, hybrid with bottom

const (
	SCREEN_LAYOUT = iota
	SCREEN_SIZING
	SCREEN_ROTATION
)

const (
	LAYOUT_VERTICAL = iota
	LAYOUT_HORZONTAL
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

const (
	RAD_0   = float64(math.Pi/180) * 0
	RAD_90  = float64(math.Pi/180) * 90
	RAD_180 = float64(math.Pi/180) * 180
	RAD_270 = float64(math.Pi/180) * 270
)

type Screen struct {
	Layout   int
	Sizing   int
	Rotation int

	Top, Bottom *ebiten.Image
	BtmAbs      BtmAbs

	Options ebiten.DrawImageOptions
}

type BtmAbs struct {
	T, B, L, R, W, H int
}

func NewScreen() *Screen {

	s := &Screen{
		Top:      ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT),
		Bottom:   ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT),
		Layout:   config.Conf.Nds.Screen.OLayout,
		Sizing:   config.Conf.Nds.Screen.OSizing,
		Rotation: config.Conf.Nds.Screen.ORotation,
	}

	return s
}

func (s *Screen) FillScreen(screen *ebiten.Image) {
	switch {
	case s.Layout == LAYOUT_HYBRID:
		s.FillHybrid(screen)
		return
	case s.Sizing == SIZING_ONLY_TOP:
		s.FillOnly(screen, false)
		return
	case s.Sizing == SIZING_ONLY_BOTTOM:
		s.FillOnly(screen, true)
		return
	case s.Layout == LAYOUT_VERTICAL:
		s.FillEvenVertical(screen)
		return
	case s.Layout == LAYOUT_HORZONTAL:
		s.FillEvenHorizontal(screen)
		return
	}
}

func (s *Screen) ApplyTouchPositions(x, y, w, h int) {
	switch s.Rotation {
	case ROT_0:
		s.BtmAbs = BtmAbs{T: y, B: y + h, L: x, R: x + w, W: w, H: h}
	case ROT_90:
		s.BtmAbs = BtmAbs{L: y, R: y + w, B: x, T: x + h, H: w, W: h}
	case ROT_180:
		s.BtmAbs = BtmAbs{B: y, T: y + h, R: x, L: x + w, W: w, H: h}
	case ROT_270:
		s.BtmAbs = BtmAbs{R: y, L: y + w, T: x, B: x + h, H: w, W: h}
	default:
		panic("unallowed nds rotation value")
	}
}

func (s *Screen) FillOnly(screen *ebiten.Image, bottom bool) {

	var image *ebiten.Image
	if bottom {
		image = s.Bottom
	} else {
		image = s.Top
	}

	var (
		rot        bool
		rotRadians float64
		rotX, rotY float64
	)

	switch s.Rotation {
	case ROT_0: // skip
		rotRadians = RAD_0
	case ROT_90:
		rotX = SCREEN_HEIGHT
		rot = true
		rotRadians = RAD_90
	case ROT_180:
		rotX = SCREEN_WIDTH
		rotY = SCREEN_HEIGHT
		rotRadians = RAD_180
	case ROT_270:
		rotY = SCREEN_WIDTH
		rot = true
		rotRadians = RAD_270
	default:
		panic("unallowed nds rotation value")
	}

	var (
		screenW        = float64(screen.Bounds().Dx())
		screenH        = float64(screen.Bounds().Dy())
		canvasW        = float64(image.Bounds().Dx())
		canvasH        = float64(image.Bounds().Dy())
		scaleX, scaleY float64
	)

	if rot {
		screenH, screenW = screenW, screenH
	}

	scaleX = screenW / canvasW
	scaleY = screenH / canvasH
	scale := min(scaleX, scaleY)
	offsetX := (screenW - (canvasW * scale)) / 2
	offsetY := (screenH - (canvasH * scale)) / 2

	if rot {
		offsetX, offsetY = offsetY, offsetX
	}

	s.Options = ebiten.DrawImageOptions{}
	s.Options.GeoM.Rotate(rotRadians)
	s.Options.GeoM.Translate(rotX, rotY)
	s.Options.GeoM.Scale(scale, scale)
	s.Options.GeoM.Translate(offsetX, offsetY)
	screen.DrawImage(image, &s.Options)

	if !bottom {
		s.BtmAbs = BtmAbs{}
		return
	}

	var (
		realX = int(offsetX)
		realY = int(offsetY)
		realW = int(scale * canvasW)
		realH = int(scale * canvasH)
	)

	s.ApplyTouchPositions(realX, realY, realW, realH)
}

func (s *Screen) FillHybrid(screen *ebiten.Image) {

	var (
		bottomW = float64(s.Top.Bounds().Dx())
		bottomH = float64(s.Top.Bounds().Dy())
		canvasW = float64(s.Top.Bounds().Dx()) * 1.5
		canvasH = float64(s.Top.Bounds().Dy())
		screenW = float64(screen.Bounds().Dx())
		screenH = float64(screen.Bounds().Dy())
		scale   = min(screenW/canvasW, screenH/canvasH)
		scaledH = scale * SCREEN_HEIGHT
		scaledW = scale * SCREEN_WIDTH
		offsetX = (screenW - (canvasW * scale)) / 2
		offsetY = (screenH - (canvasH * scale)) / 2
	)

	s.Options = ebiten.DrawImageOptions{}
	s.Options.GeoM.Scale(0.5, 0.5)
	s.Options.GeoM.Translate(SCREEN_WIDTH, 0)
	s.Options.GeoM.Scale(scale, scale)
	s.Options.GeoM.Translate(offsetX, offsetY)
	screen.DrawImage(s.Top, &s.Options)

	s.Options = ebiten.DrawImageOptions{}
	s.Options.GeoM.Scale(0.5, 0.5)
	s.Options.GeoM.Translate(SCREEN_WIDTH, SCREEN_HEIGHT/2)
	s.Options.GeoM.Scale(scale, scale)
	s.Options.GeoM.Translate(offsetX, offsetY)
	screen.DrawImage(s.Bottom, &s.Options)

	if s.Sizing == SIZING_ONLY_BOTTOM {

		s.Options = ebiten.DrawImageOptions{}
		s.Options.GeoM.Scale(scale, scale)
		s.Options.GeoM.Translate(offsetX, offsetY)
		screen.DrawImage(s.Bottom, &s.Options)

		realX := int(offsetX)
		realY := int(offsetY)
		realW := int(scale * bottomW)
		realH := int(scale * bottomH)

		s.BtmAbs = BtmAbs{
			T: realY,
			B: realY + realH,
			L: realX,
			R: realX + realW,
			W: realW,
			H: realH,
		}

		return
	}

	s.Options = ebiten.DrawImageOptions{}
	s.Options.GeoM.Scale(scale, scale)
	s.Options.GeoM.Translate(offsetX, offsetY)
	screen.DrawImage(s.Top, &s.Options)

	realW := int(scale * bottomW * 0.5)
	realH := int(scale * bottomH * 0.5)

	s.BtmAbs = BtmAbs{
		T: int((scaledH / 2) + offsetY),
		B: int(scaledH + offsetY),
		L: int(scaledW + offsetX),
		R: int((scaledW * 1.5) + offsetX),
		W: realW,
		H: realH,
	}
}

func (s *Screen) FillEvenVertical(screen *ebiten.Image) {
	var (
		screenW = float64(screen.Bounds().Dx())
		screenH = float64(screen.Bounds().Dy())
		bottomW = float64(s.Bottom.Bounds().Dx())
		bottomH = float64(s.Bottom.Bounds().Dy())

		rotRadians float64
		rotX       float64
		rotY       float64
		rot        bool
		topOff     float64
		botOff     float64
		canvasW    float64
		canvasH    float64
	)

	switch s.Rotation {
	case ROT_0: // skip
		botOff = SCREEN_HEIGHT
		rotRadians = RAD_0
	case ROT_90:
		topOff = SCREEN_WIDTH
		rotX = SCREEN_HEIGHT
		rot = true
		rotRadians = RAD_90
	case ROT_180:
		topOff = SCREEN_HEIGHT
		rotX = SCREEN_WIDTH
		rotY = SCREEN_HEIGHT
		rotRadians = RAD_180
	case ROT_270:
		botOff = SCREEN_WIDTH
		rotY = SCREEN_WIDTH
		rot = true
		rotRadians = RAD_270
	default:
		panic("unallowed nds rotation value")
	}

	if rot {
		screenH, screenW = screenW, screenH
		canvasW = bottomW * 2
		canvasH = bottomH
	} else {
		canvasH = bottomH * 2
		canvasW = bottomW
	}

	scale := min(screenW/canvasW, screenH/canvasH)
	offsetX := (screenW - (canvasW * scale)) / 2
	offsetY := (screenH - (canvasH * scale)) / 2

	if rot {
		offsetX, offsetY = offsetY, offsetX
	}

	s.Options = ebiten.DrawImageOptions{}
	s.Options.GeoM.Rotate(rotRadians)
	s.Options.GeoM.Translate(rotX, rotY)
	s.Options.GeoM.Translate(0, topOff)
	s.Options.GeoM.Scale(scale, scale)
	s.Options.GeoM.Translate(offsetX, offsetY)
	screen.DrawImage(s.Top, &s.Options)

	s.Options = ebiten.DrawImageOptions{}
	s.Options.GeoM.Rotate(rotRadians)
	s.Options.GeoM.Translate(rotX, rotY)
	s.Options.GeoM.Translate(0, botOff)
	s.Options.GeoM.Scale(scale, scale)
	s.Options.GeoM.Translate(offsetX, offsetY)
	screen.DrawImage(s.Bottom, &s.Options)

	var (
		realX = int(offsetX)
		realY = int(offsetY + (scale * botOff))
		realW = int(scale * bottomW)
		realH = int(scale * bottomH)
	)

	s.ApplyTouchPositions(realX, realY, realW, realH)
}

func (s *Screen) FillEvenHorizontal(screen *ebiten.Image) {
	var (
		screenW = float64(screen.Bounds().Dx())
		screenH = float64(screen.Bounds().Dy())
		bottomW = float64(s.Bottom.Bounds().Dx())
		bottomH = float64(s.Bottom.Bounds().Dy())

		rotRadians float64
		rotX       float64
		rotY       float64
		rot        bool
		topOff     float64
		botOff     float64
		canvasW    float64
		canvasH    float64
	)

	switch s.Rotation {
	case ROT_0:
		botOff = SCREEN_WIDTH
		rotRadians = RAD_0
	case ROT_90:
		topOff = SCREEN_HEIGHT

		rotX = SCREEN_HEIGHT
		rot = true
		rotRadians = RAD_90
	case ROT_180:
		topOff = SCREEN_WIDTH

		rotX = SCREEN_WIDTH
		rotY = SCREEN_HEIGHT
		rotRadians = RAD_180
	case ROT_270:
		botOff = SCREEN_HEIGHT

		rotY = SCREEN_WIDTH
		rot = true
		rotRadians = RAD_270
	default:
		panic("unallowed nds rotation value")
	}

	if rot {
		screenH, screenW = screenW, screenH
		canvasW = bottomW
		canvasH = bottomH * 2
	} else {
		canvasH = bottomH
		canvasW = bottomW * 2
	}

	scale := min(screenW/canvasW, screenH/canvasH)
	offsetX := (screenW - (canvasW * scale)) / 2
	offsetY := (screenH - (canvasH * scale)) / 2

	if rot {
		offsetX, offsetY = offsetY, offsetX
	}

	s.Options = ebiten.DrawImageOptions{}
	s.Options.GeoM.Rotate(rotRadians)
	s.Options.GeoM.Translate(rotX, rotY)
	s.Options.GeoM.Translate(topOff, 0)
	s.Options.GeoM.Scale(scale, scale)
	s.Options.GeoM.Translate(offsetX, offsetY)
	screen.DrawImage(s.Top, &s.Options)

	s.Options = ebiten.DrawImageOptions{}
	s.Options.GeoM.Rotate(rotRadians)
	s.Options.GeoM.Translate(rotX, rotY)
	s.Options.GeoM.Translate(botOff, 0)
	s.Options.GeoM.Scale(scale, scale)
	s.Options.GeoM.Translate(offsetX, offsetY)
	screen.DrawImage(s.Bottom, &s.Options)

	var (
		realX = int(offsetX + (scale * botOff))
		realY = int(offsetY)
		realW = int(scale * bottomW)
		realH = int(scale * bottomH)
	)

	s.ApplyTouchPositions(realX, realY, realW, realH)
}

func (s *Screen) inputHandler(field int) {

	switch field {
	case SCREEN_LAYOUT:
		s.Layout = (s.Layout + 1) % 3
	case SCREEN_SIZING:
		s.Sizing = (s.Sizing + 1) % 3
	case SCREEN_ROTATION:
		s.Rotation = (s.Rotation + 1) % 4
	}
}
