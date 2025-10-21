package gl

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

	texClr := Color{1, 1, 1, 1}
	if s.Texture.CachedTexture != nil {
        texClr = s.Texture.Sample(v.Texture.X, v.Texture.Y)
    }

	switch s.Texture.Mode {
	case 0:

		con := func(v, t float64) float64 {
			v *= FACTOR
			t *= FACTOR
			return max(0, min(1, ((t)*(v)-1)/(FACTOR*FACTOR)))
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

	case 2:

		if s.Texture.IsHighlight {

			con := func(s, t float64) float64 {
				// assume s needs to be added as 0...1
				sb := s
				s *= FACTOR
				t *= FACTOR
				return max(0, min(1, ((t)*(s)-1)/(FACTOR*FACTOR)+sb))
			}

			toon := s.Texture.ToonTbl[uint32(v.Color.R*FACTOR)]

			v.Color.R = con(toon.R, texClr.R)
			v.Color.G = con(toon.G, texClr.G)
			v.Color.B = con(toon.B, texClr.B)
			v.Color.A = con(v.Color.A, texClr.A)

		} else {

			// toon

			con := func(s, t float64) float64 {
				s *= FACTOR
				t *= FACTOR
				return max(0, min(1, ((t)*(s)-1)/(FACTOR*FACTOR)))
			}

			toon := s.Texture.ToonTbl[uint32(v.Color.R*FACTOR)]

			v.Color.R = con(toon.R, texClr.R)
			v.Color.G = con(toon.G, texClr.G)
			v.Color.B = con(toon.B, texClr.B)
			v.Color.A = con(v.Color.A, texClr.A)
		}

    case 3:
        v.Color = Transparent
	}
}

func (s *Shader) SetTexture(texture *Texture) {
	s.Texture = texture
}
