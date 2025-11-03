package rast

type Buffers struct {
    A, B []Polygon
    BisRendering bool
    DepthBufferW bool
    ManualSort bool

    //SwapBuffers isn't executed until next VBlank
    SwapSet bool
}

func (b *Buffers) Append(p Polygon) {

    if b.BisRendering {
        b.A = append(b.A, p)
        return
    }

    b.B = append(b.B, p)
}

func (b *Buffers) GetPolygons() []Polygon {

    if b.BisRendering {
        return b.B
    }

    return b.A
}

func (b *Buffers) Swap() {

    if b.BisRendering {
        b.B = []Polygon{}
    } else {
        b.A = []Polygon{}
    }

    b.BisRendering = !b.BisRendering
    b.SwapSet = false
}
