package gl

//import "fmt"

type Shader struct {
	Texture *Texture
}

func NewShader() *Shader {
	return &Shader{}
}

const FACTOR = 32.0

func (s *Shader) Fragment(v *Vertex) {

	if s.Texture == nil {
		return
	}

	texClr := s.Texture.Sample(v.Texture.X, v.Texture.Y)

	switch s.Texture.Mode {
	case 0:

		con := func(v, t float64) float64 {
			v *= FACTOR
			t *= FACTOR
			return max(0, min(1, ((t) * (v) - 1)/(FACTOR * FACTOR)))
		}

		v.Color.R = con(v.Color.R, texClr.R)
		v.Color.G = con(v.Color.G, texClr.G)
		v.Color.B = con(v.Color.B, texClr.B)
		v.Color.A = con(v.Color.A, texClr.A)

	case 1:

		con := func(v, t float64, at float64) float64 {

			v *= FACTOR
			t *= FACTOR
			at *= FACTOR

			switch {
			case at <= 0:
				return v
			case at >= FACTOR:
				return t
			default:
				return (t*at + v*(FACTOR-at)) / (FACTOR * FACTOR)
			}
		}

		v.Color.R = con(v.Color.R, texClr.R, texClr.A)
		v.Color.G = con(v.Color.G, texClr.G, texClr.A)
		v.Color.B = con(v.Color.B, texClr.B, texClr.A)

	// toon and highlight shading not implimented

	default:
		v.Color = texClr
	}
}

func (s *Shader) SetTexture(texture *Texture) {
	s.Texture = texture
}
