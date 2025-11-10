package gl

import (
	"math"
)

type Vector struct {
	X, Y, Z float64
}

func (a Vector) Dot(b Vector) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func (a Vector) Add(b Vector) Vector {
	return Vector{a.X + b.X, a.Y + b.Y, a.Z + b.Z}
}

func (a Vector) Sub(b Vector) Vector {
	return Vector{a.X - b.X, a.Y - b.Y, a.Z - b.Z}
}

func (a Vector) MulScalar(b float64) Vector {
	return Vector{a.X * b, a.Y * b, a.Z * b}
}

func (a Vector) Min(b Vector) Vector {
	return Vector{math.Min(a.X, b.X), math.Min(a.Y, b.Y), math.Min(a.Z, b.Z)}
}

func (a Vector) Max(b Vector) Vector {
	return Vector{math.Max(a.X, b.X), math.Max(a.Y, b.Y), math.Max(a.Z, b.Z)}
}

func (a Vector) Floor() Vector {
	return Vector{math.Floor(a.X), math.Floor(a.Y), math.Floor(a.Z)}
}

func (a Vector) Ceil() Vector {
	return Vector{math.Ceil(a.X), math.Ceil(a.Y), math.Ceil(a.Z)}
}

type VectorW struct {
	X, Y, Z, W float64
}

func (a VectorW) Vector() Vector {
	return Vector{a.X, a.Y, a.Z}
}

func (a VectorW) Outside() bool {
	x, y, z, w := a.X, a.Y, a.Z, a.W
	return x < -w || x > w || y < -w || y > w || z < -w || z > w
}

func (a VectorW) Dot(b VectorW) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z + a.W*b.W
}

func (a VectorW) Dot3(b VectorW) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func (a VectorW) Add(b VectorW) VectorW {
	return VectorW{a.X + b.X, a.Y + b.Y, a.Z + b.Z, a.W + b.W}
}

func (a VectorW) Sub(b VectorW) VectorW {
	return VectorW{a.X - b.X, a.Y - b.Y, a.Z - b.Z, a.W - b.W}
}

func (a VectorW) MulScalar(b float64) VectorW {
	return VectorW{a.X * b, a.Y * b, a.Z * b, a.W * b}
}

func (a VectorW) DivScalar(b float64) VectorW {
	return VectorW{a.X / b, a.Y / b, a.Z / b, a.W / b}
}
