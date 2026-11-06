package ui

import (
	"image"

	"github.com/ebitenui/ebitenui/widget"
)

var dirs = [...]widget.FocusDirection{
	widget.FOCUS_NORTH,
	widget.FOCUS_EAST,
	widget.FOCUS_SOUTH,
	widget.FOCUS_WEST,
	//	widget.FOCUS_NORTHEAST,
	//	widget.FOCUS_SOUTHEAST,
	//	widget.FOCUS_SOUTHWEST,
	//	widget.FOCUS_NORTHWEST,
}

func BuildFocus(ui *Ui) {
	if ui.scrollableMenu != nil && ui.ui.GetFocusedWidget() != nil {
		ui.scrollableMenu.KeepFocusedInView(ui.ui.GetFocusedWidget().GetWidget())
	}

	var focusers []widget.Focuser
	if ui.scrollableMenu != nil {
		all := ui.ui.Container.GetFocusers()
		sliderWidget := ui.scrollableMenu.slider.GetWidget()
		focusers = make([]widget.Focuser, 0, len(all))
		for _, f := range all {
			if f.GetWidget() != sliderWidget {
				focusers = append(focusers, f)
			}
		}
	} else {
		focusers = ui.ui.Container.GetFocusers()
	}

	for _, f := range focusers {
		for _, d := range dirs {
			if neighbor := getFocusNeighbor(f, focusers, d); neighbor != nil {
				f.AddFocus(d, neighbor)
			}
		}
	}
}

func (menu *ScrollableMenu) KeepFocusedInView(focused *widget.Widget) {
	var (
		focusedTop    = focused.Rect.Min.Y
		focusedBottom = focused.Rect.Max.Y
		rootTop       = menu.root.GetWidget().Rect.Min.Y
		rootBottom    = menu.root.GetWidget().Rect.Max.Y
	)

	var firstTop int
	if fs := menu.content.GetFocusers(); len(fs) > 0 {
		firstTop = menu.content.GetFocusers()[0].GetWidget().Rect.Min.Y
	} else {
		firstTop = menu.content.GetWidget().Rect.Min.Y
	}

	switch {
	case focusedTop <= firstTop:
		menu.slider.Current = menu.slider.Min
	case focusedTop < rootTop:
		menu.slider.Current -= rootTop - focusedTop
	case focusedBottom > rootBottom:
		menu.slider.Current += focusedBottom - rootBottom
	}
}

func getFocusNeighbor(src widget.Focuser, pool []widget.Focuser, d widget.FocusDirection) widget.Focuser {
	var (
		s  = orient(src.GetWidget().Rect, d)
		sx = (s.Min.X + s.Max.X) / 2

		best               widget.Focuser
		bestScore, bestTie int64
		bestInBeam         bool
	)

	for _, cand := range pool {
		if cand == src {
			continue
		}
		w := cand.GetWidget()
		if w.Rect.Empty() || !w.IsVisible() {
			continue
		}

		c := orient(w.Rect, d)

		if c.Max.Y <= s.Max.Y {
			continue
		}

		major := int64(max(0, c.Min.Y-s.Max.Y))

		minor := int64(max(0, max(c.Min.X-s.Max.X, s.Min.X-c.Max.X)))

		if major > 500 || minor > 500 {
			continue
		}

		inBeam := minor == 0

		const slack = 8
		if !inBeam && minor > major+slack {
			continue
		}
		score := major*major + minor*minor*4

		cx := (c.Min.X + c.Max.X) / 2
		tie := int64(abs(cx - sx))

		switch {
		case best == nil,
			inBeam && !bestInBeam,
			inBeam == bestInBeam && (score < bestScore || (score == bestScore && tie < bestTie)):
			best, bestScore, bestTie, bestInBeam = cand, score, tie, inBeam
		}
	}
	return best
}

func orient(r image.Rectangle, d widget.FocusDirection) image.Rectangle {
	switch d {
	case widget.FOCUS_NORTH:
		return image.Rect(r.Min.X, -r.Max.Y, r.Max.X, -r.Min.Y)
	case widget.FOCUS_EAST:
		return image.Rect(r.Min.Y, r.Min.X, r.Max.Y, r.Max.X)
	case widget.FOCUS_WEST:
		return image.Rect(r.Min.Y, -r.Max.X, r.Max.Y, -r.Min.X)
	default:
		return r
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func DeFocus(ui *Ui) {
	if ui.ui != nil {
		ui.ui.ClearFocus()
	}
}

func FocusFirst(ui *Ui) {
	if fs := ui.ui.Container.GetFocusers(); len(fs) > 0 {
		ui.ui.SetFocusedWidget(fs[0])
	}
}

func FocusLast(ui *Ui) {
	if fs := ui.ui.Container.GetFocusers(); len(fs) > 0 {
		ui.ui.SetFocusedWidget(fs[len(fs)-1])
	}
}
