package ppu


func inRange(coord, start, end uint32) bool {
	if end < start {
		return coord >= start || coord < end
	}
	return coord >= start && coord < end
}

func inWindow(x, y, l, r, t, b uint32) bool {
	return inRange(x, l, r) && inRange(y, t, b)
}

func WindowPixelAllowed(idx, x, y uint32, wins *Windows) bool {

	if !wins.Enabled {
		return true
	}

	win := &wins.Win0
	if win.Enabled && inWindow(x, y, win.L, win.R, win.T, win.B) {
		return win.InBg[idx]
	}

	win = &wins.Win1
	if win.Enabled && inWindow(x, y, win.L, win.R, win.T, win.B) {
		return win.InBg[idx]
	}

	return wins.OutBg[idx]
}

func WindowObjPixelAllowed(x, y uint32, wins *Windows) bool {

	if !wins.Enabled {
		return true
	}

	if !wins.Win0.Enabled && !wins.Win1.Enabled {
		return true
	}

	win := &wins.Win0
	if win.Enabled && inWindow(x, y, win.L, win.R, win.T, win.B) {
		return win.InObj
	}

	win = &wins.Win1
	if win.Enabled && inWindow(x, y, win.L, win.R, win.T, win.B) {
		return win.InObj
	}

	return wins.OutObj
}

func windowBldPixelAllowed(x, y uint32, wins *Windows, inObjWindow bool) bool {
	if !wins.Enabled {
		return true
	}

	if !wins.Win0.Enabled && !wins.Win1.Enabled && !wins.WinObj.Enabled {
		return true
	}

	win := &wins.Win0
	if win.Enabled && inWindow(x, y, win.L, win.R, win.T, win.B) {
		return win.InBld
	}

	win = &wins.Win1
	if win.Enabled && inWindow(x, y, win.L, win.R, win.T, win.B) {
		return win.InBld
	}

	if wins.WinObj.Enabled && inObjWindow {
		return wins.WinObj.InBld
	}

	return wins.OutBld
}
