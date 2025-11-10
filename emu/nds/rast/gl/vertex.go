package gl

type Vertex struct {
	Position VectorW
	Texture  Vector
	Color    Color
	Output   VectorW
    S, T float64
    NdsTexture *Texture
}

func (a Vertex) Outside() bool {
	return a.Output.Outside()
}

func (vert *Vertex) CalcTextureVector(w, h int) {
    u := vert.S / float64(w)
    v := vert.T / float64(h)
    vert.Texture = Vector{X: u, Y: v, Z: 0}
}

func (v *Vertex) InterpolateVertexes(v1, v2, v3 *Vertex, b *VectorW) {
	v.Texture = InterpolateVectors(v1.Texture, v2.Texture, v3.Texture, b)
	v.Color   = InterpolateColors(v1.Color, v2.Color, v3.Color, b)
	v.Output  = InterpolateVectorWs(v1.Output, v2.Output, v3.Output, b)
}

func InterpolateColors(v1, v2, v3 Color, b *VectorW) Color {
	return Color{
		R: (v1.R*b.X + v2.R*b.Y + v3.R*b.Z) * b.W,
		G: (v1.G*b.X + v2.G*b.Y + v3.G*b.Z) * b.W,
		B: (v1.B*b.X + v2.B*b.Y + v3.B*b.Z) * b.W,
		A: (v1.A*b.X + v2.A*b.Y + v3.A*b.Z) * b.W,
	}
}

func InterpolateVectors(v1, v2, v3 Vector, b *VectorW) Vector {
	return Vector{
		X: (v1.X*b.X + v2.X*b.Y + v3.X*b.Z) * b.W,
		Y: (v1.Y*b.X + v2.Y*b.Y + v3.Y*b.Z) * b.W,
		Z: (v1.Z*b.X + v2.Z*b.Y + v3.Z*b.Z) * b.W,
	}
}

func InterpolateVectorWs(v1, v2, v3 VectorW, b *VectorW) VectorW {
	return VectorW{
		X: (v1.X*b.X + v2.X*b.Y + v3.X*b.Z) * b.W,
		Y: (v1.Y*b.X + v2.Y*b.Y + v3.Y*b.Z) * b.W,
		Z: (v1.Z*b.X + v2.Z*b.Y + v3.Z*b.Z) * b.W,
		W: (v1.W*b.X + v2.W*b.Y + v3.W*b.Z) * b.W,
	}
}

func Barycentric(p1, p2, p3, p Vector) VectorW {
	v0 := p2.Sub(p1)
	v1 := p3.Sub(p1)
	v2 := p.Sub(p1)
	d00 := v0.Dot(v0)
	d01 := v0.Dot(v1)
	d11 := v1.Dot(v1)
	d20 := v2.Dot(v0)
	d21 := v2.Dot(v1)
	d := d00*d11 - d01*d01
	v := (d11*d20 - d01*d21) / d
	w := (d00*d21 - d01*d20) / d
	u := 1 - v - w
	return VectorW{u, v, w, 1}
}
