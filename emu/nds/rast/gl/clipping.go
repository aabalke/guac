package gl

// old clipping assumed 3d planes, not 4d float method

//var clipPlanes = []clipPlane{
//	{VectorW{1, 0, 0, 1}, VectorW{-1, 0, 0, 1}},
//	{VectorW{-1, 0, 0, 1}, VectorW{1, 0, 0, 1}},
//	{VectorW{0, 1, 0, 1}, VectorW{0, -1, 0, 1}},
//	{VectorW{0, -1, 0, 1}, VectorW{0, 1, 0, 1}},
//	{VectorW{0, 0, 1, 1}, VectorW{0, 0, -1, 1}},
//	{VectorW{0, 0, -1, 1}, VectorW{0, 0, 1, 1}},
//}

//type clipPlane struct {
//	P, N VectorW
//}
//
//func (p clipPlane) pointInFront(v VectorW) bool {
//	return v.Sub(p.P).Dot(p.N) > 0
//}
//
//func (p clipPlane) intersectSegment(v0, v1 VectorW) VectorW {
//	u := v1.Sub(v0)
//	w := v0.Sub(p.P)
//	d := p.N.Dot(u)
//	n := -p.N.Dot(w)
//	return v0.Add(u.MulScalar(n / d))
//}

var clipPlanes = []clipPlane{
	{1, 0, 0, 1},  // x + w >= 0
	{-1, 0, 0, 1}, // -x + w >= 0
	{0, 1, 0, 1},  // y + w >= 0
	{0, -1, 0, 1}, // -y + w >= 0
	{0, 0, 1, 1},  // z + w >= 0   (OpenGL style)
	{0, 0, -1, 1}, // -z + w >= 0
}

type clipPlane struct {
	A, B, C, D float32
}

func (p clipPlane) pointInFront(v VectorW) bool {
	return p.A*v.X+p.B*v.Y+p.C*v.Z+p.D*v.W >= 0
}

func (p clipPlane) intersectSegment(v0, v1 VectorW) VectorW {
	d0 := p.A*v0.X + p.B*v0.Y + p.C*v0.Z + p.D*v0.W
	d1 := p.A*v1.X + p.B*v1.Y + p.C*v1.Z + p.D*v1.W
	t := d0 / (d0 - d1)
	return v0.Add(v1.Sub(v0).MulScalar(t))
}

func sutherlandHodgman(points []VectorW, planes []clipPlane) []VectorW {
	output := points
	for _, plane := range planes {
		input := output
		output = nil
		if len(input) == 0 {
			return nil
		}
		s := input[len(input)-1]
		for _, e := range input {
			if plane.pointInFront(e) {
				if !plane.pointInFront(s) {
					x := plane.intersectSegment(s, e)
					output = append(output, x)
				}
				output = append(output, e)
			} else if plane.pointInFront(s) {
				x := plane.intersectSegment(s, e)
				output = append(output, x)
			}
			s = e
		}
	}
	return output
}

// removes allocation overhead
//var v1, v2, v3 Vertex

func ClipTriangle(t *Triangle) []*Triangle {
	w1 := t.V1.Output
	w2 := t.V2.Output
	w3 := t.V3.Output
	p1 := w1.Vector()
	p2 := w2.Vector()
	p3 := w3.Vector()
	points := []VectorW{w1, w2, w3}
	newPoints := sutherlandHodgman(points, clipPlanes)
	var result []*Triangle
	for i := 2; i < len(newPoints); i++ {
		b1 := Barycentric(p1, p2, p3, newPoints[0].Vector())
		b2 := Barycentric(p1, p2, p3, newPoints[i-1].Vector())
		b3 := Barycentric(p1, p2, p3, newPoints[i].Vector())

		var v1, v2, v3 Vertex
		v1.InterpolateVertexes(&t.V1, &t.V2, &t.V3, &b1)
		v2.InterpolateVertexes(&t.V1, &t.V2, &t.V3, &b2)
		v3.InterpolateVertexes(&t.V1, &t.V2, &t.V3, &b3)
		result = append(result, NewTriangle(v1, v2, v3))
	}
	return result
}
