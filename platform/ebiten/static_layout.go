package ui

import (
	"image"

	"github.com/ebitenui/ebitenui/widget"
)

// StaticLayout places each child at an explicit position, relative to the
// top-left of the container (after the layout's own padding is applied).
//
// Widget.LayoutData of children must be StaticLayoutData or *StaticLayoutData.
type StaticLayout struct {
	padding *widget.Insets
}

type StaticLayoutOpt func(s *StaticLayout)

type StaticLayoutOptions struct{}

var StaticLayoutOpts StaticLayoutOptions

// StaticLayoutData specifies where a child goes.
type StaticLayoutData struct {
	// Position of the child's top-left corner, relative to the container.
	Position image.Point

	// Width and Height override the child's preferred size when > 0.
	Width  int
	Height int
}

func NewStaticLayout(opts ...StaticLayoutOpt) *StaticLayout {
	s := &StaticLayout{}
	for _, o := range opts {
		o(s)
	}
	if s.padding == nil {
		s.padding = &widget.Insets{}
	}
	return s
}

// Padding configures the layout to use padding i. This affects all children.
func (o StaticLayoutOptions) Padding(i *widget.Insets) StaticLayoutOpt {
	return func(s *StaticLayout) {
		s.padding = i
	}
}

// getData accepts both value and pointer layout data, so you can keep a
// pointer and mutate it every frame.
func getData(w widget.PreferredSizeLocateableWidget) (StaticLayoutData, bool) {
	switch d := w.GetWidget().LayoutData.(type) {
	case StaticLayoutData:
		return d, true
	case *StaticLayoutData:
		if d != nil {
			return *d, true
		}
	}
	return StaticLayoutData{}, false
}

func childSize(w widget.PreferredSizeLocateableWidget, d StaticLayoutData, ok bool) (int, int) {
	ww, wh := w.PreferredSize()
	if ok {
		if d.Width > 0 {
			ww = d.Width
		}
		if d.Height > 0 {
			wh = d.Height
		}
	}
	return ww, wh
}

// PreferredSize implements widget.Layouter. It returns the bounding extent of
// all children (from the container's origin) plus padding.
func (s *StaticLayout) PreferredSize(widgets []widget.PreferredSizeLocateableWidget) (int, int) {
	maxX, maxY := 0, 0
	for _, w := range widgets {
		if w.GetWidget().GetVisibility() == widget.Visibility_Hide {
			continue
		}
		d, ok := getData(w)
		ww, wh := childSize(w, d, ok)
		if r := d.Position.X + ww; r > maxX {
			maxX = r
		}
		if b := d.Position.Y + wh; b > maxY {
			maxY = b
		}
	}
	return maxX + s.padding.Dx(), maxY + s.padding.Dy()
}

// Layout implements widget.Layouter.
func (s *StaticLayout) Layout(widgets []widget.PreferredSizeLocateableWidget, rect image.Rectangle) {
	origin := s.padding.Apply(rect).Min

	for _, w := range widgets {
		if w.GetWidget().GetVisibility() == widget.Visibility_Hide {
			continue
		}
		d, ok := getData(w)
		ww, wh := childSize(w, d, ok)

		r := image.Rect(0, 0, ww, wh).Add(origin).Add(d.Position)
		w.SetLocation(r)
	}
}
