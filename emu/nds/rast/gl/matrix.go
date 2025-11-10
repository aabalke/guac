package gl

type Matrix struct {
	X00, X01, X02, X03 float64
	X10, X11, X12, X13 float64
	X20, X21, X22, X23 float64
	X30, X31, X32, X33 float64
}

func Identity() Matrix {
	return Matrix{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1}
}

func Translate(v Vector) Matrix {
    // row based
	return Matrix{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		v.X, v.Y, v.Z, 1}
		//1, 0, 0, v.X,
		//0, 1, 0, v.Y,
		//0, 0, 1, v.Z,
		//0,0,0, 1}
}

func Scale(v Vector) Matrix {
	return Matrix{
		v.X, 0, 0, 0,
		0, v.Y, 0, 0,
		0, 0, v.Z, 0,
		0, 0, 0, 1}
}

func Screen(w, h int) Matrix {
	w2 := float64(w) / 2
	h2 := float64(h) / 2
	return Matrix{
		w2, 0, 0, w2,
		0, -h2, 0, h2,
		0, 0, 0.5, 0.5,
		0, 0, 0, 1,
	}
}

func (m Matrix) Col3(i int) Vector {
	switch i {
	case 0:
		return Vector{m.X00, m.X10, m.X20}
	case 1:
		return Vector{m.X01, m.X11, m.X21}
	case 2:
		return Vector{m.X02, m.X12, m.X22}
	default:
		panic("invalid column index")
	}
}

func (m Matrix) Col(i int) VectorW {
	switch i {
	case 0:
		return VectorW{m.X00, m.X10, m.X20, m.X30}
	case 1:
		return VectorW{m.X01, m.X11, m.X21, m.X31}
	case 2:
		return VectorW{m.X02, m.X12, m.X22, m.X32}
	case 3:
		return VectorW{m.X03, m.X13, m.X23, m.X33}
	default:
		panic("invalid column index")
	}
}

func (m Matrix) Translate(v Vector) Matrix {
	return Translate(v).Mul(m)
}

func (m Matrix) Scale(v Vector) Matrix {
	return Scale(v).Mul(m)
}

func (a Matrix) Mul(b Matrix) Matrix {
	m := Matrix{}
	m.X00 = a.X00*b.X00 + a.X01*b.X10 + a.X02*b.X20 + a.X03*b.X30
	m.X10 = a.X10*b.X00 + a.X11*b.X10 + a.X12*b.X20 + a.X13*b.X30
	m.X20 = a.X20*b.X00 + a.X21*b.X10 + a.X22*b.X20 + a.X23*b.X30
	m.X30 = a.X30*b.X00 + a.X31*b.X10 + a.X32*b.X20 + a.X33*b.X30
	m.X01 = a.X00*b.X01 + a.X01*b.X11 + a.X02*b.X21 + a.X03*b.X31
	m.X11 = a.X10*b.X01 + a.X11*b.X11 + a.X12*b.X21 + a.X13*b.X31
	m.X21 = a.X20*b.X01 + a.X21*b.X11 + a.X22*b.X21 + a.X23*b.X31
	m.X31 = a.X30*b.X01 + a.X31*b.X11 + a.X32*b.X21 + a.X33*b.X31
	m.X02 = a.X00*b.X02 + a.X01*b.X12 + a.X02*b.X22 + a.X03*b.X32
	m.X12 = a.X10*b.X02 + a.X11*b.X12 + a.X12*b.X22 + a.X13*b.X32
	m.X22 = a.X20*b.X02 + a.X21*b.X12 + a.X22*b.X22 + a.X23*b.X32
	m.X32 = a.X30*b.X02 + a.X31*b.X12 + a.X32*b.X22 + a.X33*b.X32
	m.X03 = a.X00*b.X03 + a.X01*b.X13 + a.X02*b.X23 + a.X03*b.X33
	m.X13 = a.X10*b.X03 + a.X11*b.X13 + a.X12*b.X23 + a.X13*b.X33
	m.X23 = a.X20*b.X03 + a.X21*b.X13 + a.X22*b.X23 + a.X23*b.X33
	m.X33 = a.X30*b.X03 + a.X31*b.X13 + a.X32*b.X23 + a.X33*b.X33
	return m
}

func (a Matrix) VecMul3x3(b Vector) Vector {
    x := b.Dot(a.Col3(0))
    y := b.Dot(a.Col3(1))
    z := b.Dot(a.Col3(2))
    return Vector{x, y, z}
}

func (a Matrix) MulPosition(b Vector) Vector {
	x := a.X00*b.X + a.X01*b.Y + a.X02*b.Z + a.X03
	y := a.X10*b.X + a.X11*b.Y + a.X12*b.Z + a.X13
	z := a.X20*b.X + a.X21*b.Y + a.X22*b.Z + a.X23
	return Vector{x, y, z}
}

func (a Matrix) MulPositionW(b Vector) VectorW {
	x := a.X00*b.X + a.X01*b.Y + a.X02*b.Z + a.X03
	y := a.X10*b.X + a.X11*b.Y + a.X12*b.Z + a.X13
	z := a.X20*b.X + a.X21*b.Y + a.X22*b.Z + a.X23
	w := a.X30*b.X + a.X31*b.Y + a.X32*b.Z + a.X33
	return VectorW{x, y, z, w}
}

func (a Matrix) MulVectorW(b VectorW) VectorW {
    // row based
    x := b.X*a.X00 + b.Y*a.X10 + b.Z*a.X20 + b.W*a.X30
    y := b.X*a.X01 + b.Y*a.X11 + b.Z*a.X21 + b.W*a.X31
    z := b.X*a.X02 + b.Y*a.X12 + b.Z*a.X22 + b.W*a.X32
    w := b.X*a.X03 + b.Y*a.X13 + b.Z*a.X23 + b.W*a.X33
	return VectorW{x, y, z, w}
}
