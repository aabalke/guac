package sdl

const MAX_Z = 100

// Centers nent with parent vertically and horizontally
func positionCenter(c Component, p *Component) (int32, int32, int32, int32) {

	pLayout := (*p).GetLayout()
	cLayout := c.GetLayout()

	x := GetI32(pLayout.X) - (GetI32(cLayout.W) / 2) + (GetI32(pLayout.W) / 2)
	y := GetI32(pLayout.Y) - (GetI32(cLayout.H) / 2) + (GetI32(pLayout.H) / 2)
	w := GetI32(cLayout.W)
	h := GetI32(cLayout.H)

	return x, y, w, h
}

func positionHorizontal(c Component, p *Component) (int32, int32, int32, int32) {

	pLayout := (*p).GetLayout()
	cLayout := c.GetLayout()

	x := GetI32(pLayout.X) - (GetI32(cLayout.W) / 2) + (GetI32(pLayout.W) / 2)
	y := GetI32(cLayout.Y)
	w := GetI32(cLayout.W)
	h := GetI32(cLayout.H)

	return x, y, w, h
}

func distributeEvenlyVertical(c Component) {

	children := c.GetChildren()
	if len(children) == 0 {
		return
	}

	cLayout := c.GetLayout()
	x := GetI32(cLayout.X)
	y := GetI32(cLayout.Y)
	w := GetI32(cLayout.W)
	h := GetI32(cLayout.H)

	size := int(h / int32(len(children)))

	for i, child := range children {

		l := (*child).GetLayout()
		cW := GetI32(l.W)
		cH := GetI32(l.H)

		height := I32(y + int32(size*i) + int32(size/2) - (cH / 2))
		width := I32(x + (w / 2) - (cW / 2))

		(*child).SetLayout(Layout{X: width, Y: height, W: l.W, H: l.H, Z: l.Z})
	}
}

func positionRelative(cLayout Layout, pLayout Layout) (int32, int32, int32, int32, int32) {

	x := GetI32(pLayout.X) + GetI32(cLayout.X)
	y := GetI32(pLayout.Y) + GetI32(cLayout.Y)
	w := GetI32(cLayout.W)
	h := GetI32(cLayout.H)
	z := GetI32(cLayout.Z)

	return x, y, w, h, z
	//return Layout{X: x, Y: y, W: w, H: h, Z: z}
}

/** Apply childFunc to all Component children with z index as priority **/
func ChildFuncUpdate(c Component, childFunc func(*Component) bool) {

	children := c.GetChildren()

	countRendered := 0
	var z, i int32 = 0, MAX_Z
	for i = range MAX_Z {
		z = MAX_Z - i

		if len(children) == countRendered {
			return
		}

		for _, child := range children {

			if cLayout := (*child).GetLayout(); cLayout.Z != z {
				continue
			}

			block := childFunc(child)
			if block {
				return
			}
		}
	}
}

/** Apply childFunc to all Component children with z index as priority **/
func ChildFunc(c Component, childFunc func(*Component)) {

	children := c.GetChildren()

	countRendered := 0
	var z int32 = 0
	for z = range MAX_Z {

		if len(children) == countRendered {
			return
		}

		for _, child := range children {

			if cLayout := (*child).GetLayout(); cLayout.Z != z {
				continue
			}

			childFunc(child)
		}
	}
}

/** Remove child from list of children on parent **/
func RemoveChild(child *Component) {

	parent := (*(*child).GetParent())
	temp := parent.GetChildren()

	for i, c := range parent.GetChildren() {

		if c != child {
			continue
		}

		temp = append(temp[:i], temp[i+1:]...)
		break
	}

	parent.SetChildren(temp)
}
