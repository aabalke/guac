package gl

type Quad struct {
	V1, V2, V3, V4 Vertex
}

func NewQuad(v1, v2, v3, v4 Vertex) *Quad {
	q := Quad{v1, v2, v3, v4}
	q.FixNormals()
	return &q
}

func NewQuadForPoints(p1, p2, p3, p4 Vector) *Quad {
	v1 := Vertex{Position: p1}
	v2 := Vertex{Position: p2}
	v3 := Vertex{Position: p3}
	v4 := Vertex{Position: p4}
	return NewQuad(v1, v2, v3, v4)
}

func (q *Quad) IsDegenerate() bool {
	p1 := q.V1.Position
	p2 := q.V2.Position
	p3 := q.V3.Position
	p4 := q.V4.Position
	//if p1 == p2 || p1 == p3 || p2 == p3 || p4 == p1 || p4 == p2 || {
	//	return true
	//}
	if p1.IsDegenerate() || p2.IsDegenerate() || p3.IsDegenerate() || p4.IsDegenerate() {
		return true
	}
	return false
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

func (q *Quad) Area() float64 {
    panic("Quad Area")
	//e1 := t.V2.Position.Sub(t.V1.Position)
	//e2 := t.V3.Position.Sub(t.V1.Position)
	//n := e1.Cross(e2)
	//return n.Length() / 2
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

func (q *Quad) BoundingBox() Box {
	//min := t.V1.Position.Min(t.V2.Position).Min(t.V3.Position)
	//max := t.V1.Position.Max(t.V2.Position).Max(t.V3.Position)
	//return Box{min, max}
    panic("BBox quad")
    return Box{}
}

func (q *Quad) Transform(matrix Matrix) {
	q.V1.Position = matrix.MulPosition(q.V1.Position)
	q.V2.Position = matrix.MulPosition(q.V2.Position)
	q.V3.Position = matrix.MulPosition(q.V3.Position)
	q.V4.Position = matrix.MulPosition(q.V4.Position)
	q.V1.Normal = matrix.MulDirection(q.V1.Normal)
	q.V2.Normal = matrix.MulDirection(q.V2.Normal)
	q.V3.Normal = matrix.MulDirection(q.V3.Normal)
	q.V4.Normal = matrix.MulDirection(q.V4.Normal)
}

func (q *Quad) ReverseWinding() {
	q.V1, q.V2, q.V3, q.V4 = q.V4, q.V3, q.V2, q.V1
	q.V1.Normal = q.V1.Normal.Negate()
	q.V2.Normal = q.V2.Normal.Negate()
	q.V3.Normal = q.V3.Normal.Negate()
	q.V4.Normal = q.V4.Normal.Negate()
}

func (q *Quad) SetColor(c Color) {
	q.V1.Color = c
	q.V2.Color = c
	q.V3.Color = c
	q.V4.Color = c
}

// func (q *Quad) RandomPoint() Vector {
// 	v1 := t.V1.Position
// 	v2 := t.V2.Position.Sub(v1)
// 	v3 := t.V3.Position.Sub(v1)
// 	for {
// 		a := rand.Float64()
// 		b := rand.Float64()
// 		if a+b <= 1 {
// 			return v1.Add(v2.MulScalar(a)).Add(v3.MulScalar(b))
// 		}
// 	}
// }

// func (q *Quad) Area() float64 {
// 	e1 := t.V2.Position.Sub(t.V1.Position)
// 	e2 := t.V3.Position.Sub(t.V1.Position)
// 	return e1.Cross(e2).Length() / 2
// }
