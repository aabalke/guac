package gl

type Triangle struct {
	V1, V2, V3 Vertex
}

func NewTriangle(v1, v2, v3 Vertex) *Triangle {
	t := Triangle{v1, v2, v3}
	return &t
}

type Quad struct {
	V1, V2, V3, V4 Vertex
}

func NewQuad(v1, v2, v3, v4 Vertex) *Quad {
	q := Quad{v1, v2, v3, v4}
	return &q
}
