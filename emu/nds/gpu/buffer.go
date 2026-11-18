package gpu

type Buffers struct {
	bufGeo *Buffer
	bufRen *Buffer

	// SwapBuffers isn't executed until next VBlank
	SwapSet bool
}

type Buffer struct {
	Polys        []Polygon
	DepthBufferW bool
	ManualSort   bool
}

func NewBuffers() *Buffers {
	return &Buffers{
		bufGeo: &Buffer{},
		bufRen: &Buffer{},
	}
}

func (b *Buffers) Swap() {
	tmp := b.bufGeo
	b.bufGeo = b.bufRen
	b.bufRen = tmp
	b.bufGeo.Polys = []Polygon{}
	b.SwapSet = false
}

func (b *Buffer) GetCnts() (int, int) {
	// will need to handle culling as well in cnt
	// will need box test to remove

	polyCnt, vertCnt := 0, 0

	for i := range len(b.Polys) {

		poly := &b.Polys[i]

		switch poly.PrimitiveType {
		case PRIM_SEP_TRI:
			polyCnt += len(poly.Vertices) / 3
			vertCnt += len(poly.Vertices)

		case PRIM_SEP_QUAD:
			polyCnt += len(poly.Vertices) / 4
			vertCnt += len(poly.Vertices)

		case PRIM_TRI_STRIP:
			polyCnt += len(poly.Vertices) - 2
			vertCnt += len(poly.Vertices)

		case PRIM_QUAD_STRIP:
			polyCnt += (len(poly.Vertices) - 2) / 2
			vertCnt += len(poly.Vertices)
		}
	}

	// zelda spirit tracks checks for vertices to be below amount that is not the case, need to figure out why
	return 0, 0

	polyCnt = min(2048, polyCnt)
	vertCnt = min(6144, vertCnt)

	return polyCnt, vertCnt
}
