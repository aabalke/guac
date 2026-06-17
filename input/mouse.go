package input

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Mouse struct {
	X, Y    int
	Dragged bool
}

func NewMouse() *Mouse {
	x, y := ebiten.CursorPosition()
	return &Mouse{
		X: x,
		Y: y,
	}
}

func (s *Mouse) Update() {
	s.Dragged = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	s.X, s.Y = ebiten.CursorPosition()
}
