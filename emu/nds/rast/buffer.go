package rast

type Buffers struct {
    A, B []Polygon
    BisRendering bool
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
