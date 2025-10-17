package gl

type Triangle struct {
	V1, V2, V3 Vertex
}

func NewTriangle(v1, v2, v3 Vertex) *Triangle {
	t := Triangle{v1, v2, v3}
	t.FixNormals()
	return &t
}

func (t *Triangle) Normal() Vector {
	e1 := t.V2.Position.Sub(t.V1.Position)
	e2 := t.V3.Position.Sub(t.V1.Position)
	return e1.Cross(e2).Normalize()
}

func (t *Triangle) FixNormals() {
	n := t.Normal()
	zero := Vector{}
	if t.V1.Normal == zero {
		t.V1.Normal = n
	}
	if t.V2.Normal == zero {
		t.V2.Normal = n
	}
	if t.V3.Normal == zero {
		t.V3.Normal = n
	}
}

type Quad struct {
	V1, V2, V3, V4 Vertex
}

func NewQuad(v1, v2, v3, v4 Vertex) *Quad {
	q := Quad{v1, v2, v3, v4}
	q.FixNormals()
	return &q
}

func (q *Quad) Normal() Vector {
    //panic("Quad Normal")
    //e1 := t.V2.Position.Sub(t.V1.Position)
    //e2 := t.V3.Position.Sub(t.V1.Position)
    //return e1.Cross(e2).Normalize()
    // First triangle
    e1 := q.V2.Position.Sub(q.V1.Position)
    e2 := q.V3.Position.Sub(q.V1.Position)
    n1 := e1.Cross(e2).Normalize()

    // Second triangle
    e3 := q.V3.Position.Sub(q.V1.Position)
    e4 := q.V4.Position.Sub(q.V1.Position)
    n2 := e3.Cross(e4).Normalize()

    // Average normals for the quad
    normal := n1.Add(n2).Normalize()
    return normal
}

func (q *Quad) FixNormals() {
	n := q.Normal()
	zero := Vector{}
	if q.V1.Normal == zero {
		q.V1.Normal = n
	}
	if q.V2.Normal == zero {
		q.V2.Normal = n
	}
	if q.V3.Normal == zero {
		q.V3.Normal = n
	}
	if q.V4.Normal == zero {
		q.V4.Normal = n
	}
}
