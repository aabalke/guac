package ui

import (
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
)

type Focus struct {
	ui *ebitenui.UI

	other   []widget.Focuser
	sidebar []widget.Focuser
	submenu []widget.Focuser

	horizontalGroup [][]widget.Focuser
}

func (f *Focus) ClearFocus() {
	f.other = []widget.Focuser{}
	f.sidebar = []widget.Focuser{}
	f.submenu = []widget.Focuser{}
	f.horizontalGroup = [][]widget.Focuser{}
}

func (f *Focus) BuildFocus(ui *ebitenui.UI) {
	f.ui = ui
	f.buildFocusGroup(&f.other)

	if len(f.sidebar) == 0 {
		return
	}
	if len(f.submenu) == 0 {
		return
	}

	f.buildSidebarFocus()
	f.buildSubFocus()
}

func (f *Focus) buildFocusGroup(group *[]widget.Focuser) {

	if len(*group) <= 0 {
		return
	}

	(*group)[0].AddFocus(widget.FOCUS_SOUTH, (*group)[1])

	for i := 1; i < len(*group)-1; i++ {
		(*group)[i].AddFocus(widget.FOCUS_NORTH, (*group)[i-1])
		(*group)[i].AddFocus(widget.FOCUS_SOUTH, (*group)[i+1])
	}

	(*group)[len(*group)-1].AddFocus(widget.FOCUS_NORTH, (*group)[len(*group)-2])
}

func (f *Focus) buildSidebarFocus() {
	for i, focuser := range f.sidebar {

		if i != 0 {
			focuser.AddFocus(widget.FOCUS_NORTH, f.sidebar[i-1])
		}

		if i != len(f.sidebar)-1 {
			focuser.AddFocus(widget.FOCUS_SOUTH, f.sidebar[i+1])
		}

		focuser.AddFocus(widget.FOCUS_EAST, f.submenu[0])
	}
}

func (f *Focus) buildSubFocus() {
	for i, focuser := range f.submenu {

		if i != 0 {
			focuser.AddFocus(widget.FOCUS_NORTH, f.submenu[i-1])
		}

		if i != len(f.submenu)-1 {
			focuser.AddFocus(widget.FOCUS_SOUTH, f.submenu[i+1])
		}

		focuser.AddFocus(widget.FOCUS_WEST, f.sidebar[0])
	}

	f.buildHorizontalGroupOverride()
}

// this is only allowed in submenus right now
// since radio groups are [button1][button2][button3]
// using identical functionality to other elements will cause l, r == up, down
func (f *Focus) buildHorizontalGroupOverride() {

	if f.sidebar == nil || len(f.sidebar) == 0 {
		return
	}

	sidebar := f.sidebar[0]

	for _, group := range f.horizontalGroup {
		above := group[0].GetFocus(widget.FOCUS_NORTH)
		below := group[len(group)-1].GetFocus(widget.FOCUS_SOUTH)

		for j, button := range group {

			button.AddFocus(widget.FOCUS_NORTH, above)
			button.AddFocus(widget.FOCUS_SOUTH, below)

			if j == 0 {
				button.AddFocus(widget.FOCUS_WEST, sidebar)
			} else if j < len(group) {
				button.AddFocus(widget.FOCUS_WEST, group[j-1])
			}

			if j < len(group)-1 {
				button.AddFocus(widget.FOCUS_EAST, group[j+1])
			}
		}
	}
}

func (f *Focus) FocusLastSubMenu() {
	if len(f.submenu) > 0 {
		f.ui.SetFocusedWidget(f.submenu[len(f.submenu)-1])
	}
}

func (f *Focus) FocusSidebar(idx int) {
	if len(f.sidebar) > idx {
		f.ui.SetFocusedWidget(f.sidebar[idx])
	}
}

func (f *Focus) FocusSubmenu() {
	if len(f.submenu) > 0 {
		f.ui.SetFocusedWidget(f.submenu[0])
	}
}

func (f *Focus) DeFocus() {
	if f.ui != nil {
		f.ui.ClearFocus()
	}
}

func (f *Focus) KeepFocusedInView(slider *widget.Slider) {

	if len(f.submenu) == 0 ||
		slider == nil ||
		f.ui == nil ||
		f.ui.GetFocusedWidget() == nil {
		return
	}

	var (
		currMax = f.ui.GetFocusedWidget().GetWidget().Rect.Max.Y
		currMin = f.ui.GetFocusedWidget().GetWidget().Rect.Min.Y
		rootMax = f.ui.Container.GetWidget().Rect.Max.Y
		rootMin = f.ui.Container.GetWidget().Rect.Min.Y
	)

	switch f.ui.GetFocusedWidget() {
	case f.submenu[0]:
		slider.Current = slider.Min
		return
	case f.submenu[len(f.submenu)-1]:
		slider.Current = slider.Max
		return
	}

	switch {
	case currMax > rootMax:
		slider.Current += currMax - rootMax
	case currMin < rootMin:
		slider.Current -= rootMin - currMin
	}
}
