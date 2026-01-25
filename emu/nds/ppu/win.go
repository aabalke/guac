package ppu

//go:inline
func inRange(coord, start, end uint32) bool {
	if end < start {
		return coord >= start || coord < end
	}
	return coord >= start && coord < end
}

//go:inline
func inWindow(x, y, l, r, t, b uint32) bool {
	return inRange(x, l, r) && inRange(y, t, b)
}

//go:inline
func (wins *Windows) inWinBg(bgIdx, x, y uint32) bool {
    win0 := &wins.Win0
    win1 := &wins.Win1
    if win0.Enabled && inWindow(x, y, win0.L, win0.R, win0.T, win0.B) {
        return win0.InBg[bgIdx]
    } else if win1.Enabled && inWindow(x, y, win1.L, win1.R, win1.T, win1.B) {
        return win1.InBg[bgIdx]
    } 
    return wins.OutBg[bgIdx]
}

//go:inline
func (wins *Windows) inWinObj(x, y uint32) bool {

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

//go:inline
func (wins *Windows) inWinBld(x, y uint32) bool {

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

	if wins.WinObj.Enabled && wins.inObjWindow[x] {
		return wins.WinObj.InBld
	}

	return wins.OutBld
}
