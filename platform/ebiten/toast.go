package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type Toast struct {
	active  int
	face    *text.Face
	bgClr   *color.Color
	bus     *TimedEventBus
	drawers []func(dst *ebiten.Image)
}

func NewToast(bus *TimedEventBus, bgColor *color.Color, face *text.Face) *Toast {
	return &Toast{
		bus:   bus,
		bgClr: bgColor,
		face:  face,
	}
}

func (t *Toast) Draw(dst *ebiten.Image) {
	for i := range t.drawers {
		t.drawers[i](dst)
	}
}

func (t *Toast) Message(message string) {
	const (
		DURATION       = 120
		TRANS_DURATION = 40
	)

	var (
		padding = struct{ L, R, T, B int }{32, 32, 4, 4}
		margin  = float64(16)
		tw, th  = text.Measure(message, *t.face, 0)
		iw, ih  = int(tw) + padding.L + padding.R, int(th) + padding.T + padding.B
		dt      = float64(iw) + margin + margin
	)

	img := ebiten.NewImage(iw, ih)
	img.Fill(*t.bgClr)

	textOpts := &text.DrawOptions{}
	textOpts.GeoM.Translate(float64(padding.L), float64(padding.T))
	text.Draw(img, message, *t.face, textOpts)

	opts := &ebiten.DrawImageOptions{}
	if TRANS_DURATION != 0 {
		opts.GeoM.Translate(margin-dt, 0)
	} else {
		opts.GeoM.Translate(margin, 0)
	}

	e := &TimedEvent{
		eventType: EVENT_TOAST,
		TicksLeft: DURATION,
	}

	t.drawers = append(t.drawers, func(dst *ebiten.Image) {
		dy := dst.Bounds().Dy() + (-img.Bounds().Dy() - int(margin))
		opts.GeoM.Translate(0, float64(dy))
		dst.DrawImage(img, opts)
	})

	t.active++

	e.OnTick = func() {
		opts.GeoM.Reset()

		if TRANS_DURATION == 0 {
			opts.GeoM.Translate(margin, 0)
			return
		}
		switch {
		case e.TicksLeft > DURATION-TRANS_DURATION:
			transLast := DURATION - TRANS_DURATION
			mul := (float64(e.TicksLeft-transLast) / TRANS_DURATION)
			mul = CalcBezierBlend(mul)
			opts.GeoM.Translate(margin-(dt*mul), 0)
		case e.TicksLeft < TRANS_DURATION:
			transLast := TRANS_DURATION
			mul := (float64(e.TicksLeft-transLast) / TRANS_DURATION)
			mul = CalcBezierBlend(mul)
			opts.GeoM.Translate(margin-(dt*mul), 0)
		default:
			opts.GeoM.Translate(margin, 0)
		}
	}

	e.OnDone = func() {
		t.drawers = t.drawers[1:]
		t.active--
	}

	t.bus.AddEvent(e)
}

// https://stackoverflow.com/questions/13462001/ease-in-and-ease-out-animation-formula
func CalcBezierBlend(t float64) float64 {
	return t * t * (3.0 - 2.0*t)
}
