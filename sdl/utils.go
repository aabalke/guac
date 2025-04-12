package sdl

const MAX_Z = 100

// Centers nent with parent vertically and horizontally
func positionCenter(c Component, p *Component) (int32, int32, int32, int32) {

	pLayout := (*p).GetLayout()
	cLayout := c.GetLayout()

	x := pLayout.X - (cLayout.W / 2) + (pLayout.W / 2)
	y := pLayout.Y - (cLayout.H / 2) + (pLayout.H / 2)
	w := cLayout.W
	h := cLayout.H

	return x, y, w, h
}

func distributeEvenlyVertical(c Component) {

	layout := c.GetLayout()
	children := c.GetChildren()
	size := int(layout.H) / (len(children))

	for i, child := range children {

		l := (*child).GetLayout()

		height := layout.Y + int32(size*i) + int32(size/2) - (l.H / 2)
		width := layout.X + (layout.W / 2) - (l.W / 2)

		(*child).SetLayout(Layout{X: width, Y: height, W: l.W, H: l.H, Z: l.Z})
	}
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

            if block := childFunc(child); block {
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
