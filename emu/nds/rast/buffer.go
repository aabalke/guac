package rast

type Buffers struct {
	A, B         Buffer
	BisRendering bool

	//SwapBuffers isn't executed until next VBlank
	SwapSet bool
}

type Buffer struct {
	Polys        []Polygon
	DepthBufferW bool
	ManualSort   bool
}

func (b *Buffers) Append(p Polygon) {

	if b.BisRendering {
		b.A.Polys = append(b.A.Polys, p)
		return
	}

	b.B.Polys = append(b.B.Polys, p)
}

func (b *Buffers) GetPolygons() (p []Polygon, manualSort, depthW bool) {

	if b.BisRendering {
		return b.B.Polys, b.B.ManualSort, b.B.DepthBufferW
	}

	return b.A.Polys, b.A.ManualSort, b.A.DepthBufferW
}

func (b *Buffers) Swap() {

	if b.BisRendering {
		b.B = Buffer{}
	} else {
		b.A = Buffer{}
	}

	b.BisRendering = !b.BisRendering
	b.SwapSet = false
}

func (b *Buffers) SwapCmd(data uint32) {

	buf := &b.B
	if b.BisRendering {
		buf = &b.A
	}

	buf.ManualSort = data&0b1 != 0
	buf.DepthBufferW = data&0b10 != 0
	b.SwapSet = true
}
