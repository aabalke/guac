package rast

type Buffers struct {
	A, B         Buffer
	BisRendering bool

	//SwapBuffers isn't executed until next VBlank
	SwapSet bool
}

type Buffer struct {
	DepthBufferW bool
	ManualSort   bool

	Polys []Polygon
}

func (b *Buffers) Append(p Polygon) {

	if b.BisRendering {
		b.A.Polys = append(b.A.Polys, p)
		return
	}

	b.B.Polys = append(b.B.Polys, p)
}

func (b *Buffers) GetBuffer() *Buffer {

	if b.BisRendering {
		return &b.B
	}

	return &b.A
}

func (b *Buffers) Swap() {

	b.BisRendering = !b.BisRendering

	buf := &b.B
	if b.BisRendering {
		buf = &b.A
	}

	buf.Polys = []Polygon{}

	b.SwapSet = false
}

func (b *Buffers) SwapCmd(data uint32) {

	buf := &b.B
	if b.BisRendering {
		buf = &b.A
	}

	buf.ManualSort = data&1 != 0
	buf.DepthBufferW = (data>>1)&1 != 0
	b.SwapSet = true
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

    polyCnt = min(2048, polyCnt)
    vertCnt = min(6144, vertCnt)

    return polyCnt, vertCnt
}
