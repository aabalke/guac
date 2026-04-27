package input

import (
	"image/color"
	_ "image/png"

	_ "embed"

	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Mouse struct {
	X, Y        int
	img         *ebiten.Image
	DraggedLeft bool

	alpha          float32
	size           int
	clr, clrStroke color.Color
	stroke, fill   bool
	strokeWidth    int
}

func NewMouse() *Mouse {

	x, y := ebiten.CursorPosition()

	mouse := config.Conf.Mouse

	if mouse.CursorSize == 0 {
		panic("uhhhhh")
	}

	m := &Mouse{
		X:           x,
		Y:           y,
		fill:        mouse.Fill,
		stroke:      mouse.Stroke,
		alpha:       mouse.UnSelectedAlpha,
		size:        mouse.CursorSize,
		strokeWidth: mouse.StrokeSize,
		clr: color.RGBA{
			R: mouse.FillColor[0],
			G: mouse.FillColor[1],
			B: mouse.FillColor[2],
			A: 0xFF,
		},
		clrStroke: color.RGBA{
			R: mouse.StrokeColor[0],
			G: mouse.StrokeColor[1],
			B: mouse.StrokeColor[2],
			A: 0xFF,
		},
	}

	r := m.size >> 1
	m.img = ebiten.NewImage(m.size, m.size)

	m.img.Clear()

	if m.fill {
		vector.DrawFilledCircle(m.img, float32(r), float32(r), float32(r), m.clr, true)
	}

	if m.stroke {
		vector.StrokeCircle(m.img, float32(r), float32(r), float32(r)-(float32(m.strokeWidth)/2), float32(m.strokeWidth), m.clrStroke, true)
	}

	return m
}

func (m *Mouse) Draw(screen *ebiten.Image) {

	r := m.size >> 1

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(m.X-r), float64(m.Y-r))

	if !m.DraggedLeft {
		op.ColorScale.ScaleAlpha(m.alpha)
	}

	screen.DrawImage(m.img, op)
}

func (s *Mouse) Update() {
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		s.DraggedLeft = false
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		s.DraggedLeft = true
	}

	s.X, s.Y = ebiten.CursorPosition()
}
