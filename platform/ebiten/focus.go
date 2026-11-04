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
	if ui.scrollableMenu != nil && len(ui.gamepadIds) != 0 {
		KeepFocusedInView(ui)
	}

	focusers := ui.ui.Container.GetFocusers()

	for _, f := range focusers {
		for _, d := range dirs {
			if neighbor := getFocusNeighbor(f, focusers, d); neighbor != nil {
				f.AddFocus(d, neighbor)
			}
		}
	}
}

func getFocusNeighbor(curr widget.Focuser, fs []widget.Focuser, d widget.FocusDirection) widget.Focuser {
	const MAJOR_PRIORITY = 13

	var (
		s     = orient(curr.GetWidget().Rect, d)
		currX = (s.Min.X + s.Max.X) / 2
		currY = (s.Min.Y + s.Max.Y) / 2

		best       widget.Focuser
		bestScore  int64
		bestInBeam bool
	)

	for _, cand := range fs {

		if cand == curr {
			continue
		}

		w := cand.GetWidget()
		if hidden := w.Rect.Empty() || !w.IsVisible(); hidden {
			continue
		}

		c := orient(w.Rect, d)
		candX, candY := (c.Min.X+c.Max.X)/2, (c.Min.Y+c.Max.Y)/2

		if wrongDir := candY <= currY; wrongDir {
			continue
		}

		var (
			major  = int64(max(0, c.Min.Y-s.Max.Y))
			minor  = int64(abs(candX - currX))
			inBeam = c.Min.X < s.Max.X && c.Max.X > s.Min.X
			score  = MAJOR_PRIORITY*major*major + minor*minor
		)

		if best == nil {
			best, bestScore, bestInBeam = cand, score, inBeam
			continue
		}

		if !inBeam {
			continue
		}

		if !bestInBeam || score < bestScore {
			best, bestScore, bestInBeam = cand, score, inBeam
			continue
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

func KeepFocusedInView(ui *Ui) {
	if ui.ui == nil || ui.ui.GetFocusedWidget() == nil {
		return
	}

	var (
		focusedTop    = ui.ui.GetFocusedWidget().GetWidget().Rect.Min.Y
		focusedBottom = ui.ui.GetFocusedWidget().GetWidget().Rect.Max.Y
		rootTop       = ui.scrollableMenu.root.GetWidget().Rect.Min.Y
		rootBottom    = ui.scrollableMenu.root.GetWidget().Rect.Max.Y
	)

	var firstTop int
	if fs := ui.scrollableMenu.content.GetFocusers(); len(fs) > 0 {
		firstTop = ui.scrollableMenu.content.GetFocusers()[0].GetWidget().Rect.Min.Y
	} else {
		firstTop = ui.scrollableMenu.content.GetWidget().Rect.Min.Y
	}

	switch {
	case focusedTop <= firstTop:
		ui.scrollableMenu.slider.Current = ui.scrollableMenu.slider.Min
	case focusedTop < rootTop:
		ui.scrollableMenu.slider.Current -= rootTop - focusedTop

	case focusedBottom > rootBottom:
		ui.scrollableMenu.slider.Current += focusedBottom - rootBottom
	}
}

func DeFocus(ui *Ui) {
	if ui.ui != nil {
		ui.ui.ClearFocus()
	}
}

func FocusFirst(ui *Ui) {
	fs := ui.ui.Container.GetFocusers()

	if len(fs) <= 0 {
		return
	}

	ui.ui.SetFocusedWidget(fs[0])
}

func FocusLast(ui *Ui) {
	fs := ui.ui.Container.GetFocusers()

	if len(fs) <= 0 {
		return
	}

	ui.ui.SetFocusedWidget(fs[len(fs)-1])
}
