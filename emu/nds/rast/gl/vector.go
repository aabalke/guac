package gl

import (
	"math"
)

type Vector struct {
	X, Y, Z float32
}

func (a Vector) Dot(b Vector) float32 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func (a Vector) Add(b Vector) Vector {
	return Vector{a.X + b.X, a.Y + b.Y, a.Z + b.Z}
}

func (a Vector) Sub(b Vector) Vector {
	return Vector{a.X - b.X, a.Y - b.Y, a.Z - b.Z}
}

func (a Vector) MulScalar(b float32) Vector {
	return Vector{a.X * b, a.Y * b, a.Z * b}
}

func (a Vector) Min(b Vector) Vector {
	return Vector{
		float32(math.Min(float64(a.X), float64(b.X))),
		float32(math.Min(float64(a.Y), float64(b.Y))),
		float32(math.Min(float64(a.Z), float64(b.Z)))}
}

func (a Vector) Max(b Vector) Vector {
	return Vector{
		float32(math.Max(float64(a.X), float64(b.X))),
		float32(math.Max(float64(a.Y), float64(b.Y))),
		float32(math.Max(float64(a.Z), float64(b.Z)))}
}

func (a Vector) Floor() Vector {
	return Vector{
		float32(math.Floor(float64(a.X))),
		float32(math.Floor(float64(a.Y))),
		float32(math.Floor(float64(a.Z)))}
}

func (a Vector) Ceil() Vector {
	return Vector{
		float32(math.Ceil(float64(a.X))),
		float32(math.Ceil(float64(a.Y))),
		float32(math.Ceil(float64(a.Z)))}
}

type VectorW struct {
	X, Y, Z, W float32
}

func (a VectorW) Vector() Vector {
	return Vector{a.X, a.Y, a.Z}
}

func (a VectorW) Outside() bool {
	x, y, z, w := a.X, a.Y, a.Z, a.W
	return x < -w || x > w || y < -w || y > w || z < -w || z > w
}

func (a VectorW) Dot(b VectorW) float32 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z + a.W*b.W
}

func (a VectorW) Dot3(b VectorW) float32 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func (a VectorW) Add(b VectorW) VectorW {
	return VectorW{a.X + b.X, a.Y + b.Y, a.Z + b.Z, a.W + b.W}
}

func (a VectorW) Sub(b VectorW) VectorW {
	return VectorW{a.X - b.X, a.Y - b.Y, a.Z - b.Z, a.W - b.W}
}

func (a VectorW) MulScalar(b float32) VectorW {
	return VectorW{a.X * b, a.Y * b, a.Z * b, a.W * b}
}

func (a VectorW) DivScalar(b float32) VectorW {
	return VectorW{a.X / b, a.Y / b, a.Z / b, a.W / b}
}
