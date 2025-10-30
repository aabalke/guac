package gl

type Shader struct {
	Texture *Texture
}

func NewShader() *Shader {
	return &Shader{}
}

const FACTOR = 32.0

var blendFunc = [...]func(texture *Texture, vColor, tColor Color) Color {

    func(texture *Texture, vColor, tColor Color) Color {

		con := func(v, t float64) float64 {
			v *= FACTOR
			t *= FACTOR
			return max(0, min(1, ((t)*(v)-1)/(FACTOR*FACTOR)))
		}

		vColor.R = con(vColor.R, tColor.R)
		vColor.G = con(vColor.G, tColor.G)
		vColor.B = con(vColor.B, tColor.B)
		vColor.A = con(vColor.A, tColor.A)
        return vColor
    },
    func(texture *Texture, vColor, tColor Color) Color {

		con := func(v, t, at float64) float64 {
			v *= FACTOR
			t *= FACTOR
			at *= FACTOR
            return max(0, min(1, (t*at + v*(FACTOR-at)) / (FACTOR * FACTOR)))
		}

		vColor.R = con(vColor.R, tColor.R, tColor.A)
		vColor.G = con(vColor.G, tColor.G, tColor.A)
		vColor.B = con(vColor.B, tColor.B, tColor.A)
        return vColor
    },
    func(texture *Texture, vColor, tColor Color) Color {

		if texture.IsHighlight {

			con := func(s, t float64) float64 {
				// assume s needs to be added as 0...1
				sb := s
				s *= FACTOR
				t *= FACTOR
				return max(0, min(1, ((t)*(s)-1)/(FACTOR*FACTOR)+sb))
			}

			toon := texture.ToonTbl[uint32(vColor.R*FACTOR)& 0x1F]

			vColor.R = con(toon.R, tColor.R)
			vColor.G = con(toon.G, tColor.G)
			vColor.B = con(toon.B, tColor.B)
			vColor.A = con(vColor.A, tColor.A)
            return vColor
        }

        // toon

        con := func(s, t float64) float64 {
            s *= FACTOR
            t *= FACTOR
            return max(0, min(1, ((t)*(s)-1)/(FACTOR*FACTOR)))
        }

        toon := texture.ToonTbl[uint32(vColor.R*FACTOR)]

        vColor.R = con(toon.R, tColor.R)
        vColor.G = con(toon.G, tColor.G)
        vColor.B = con(toon.B, tColor.B)
        vColor.A = con(vColor.A, tColor.A)
        return vColor
    },
    func(texture *Texture, vColor, tColor Color) Color {
        return Transparent
    },
}

func (s *Shader) Fragment(v *Vertex) {

    if s.Texture == nil {
        return
    }

	tColor := White
	if s.Texture.CachedTexture != nil {
        tColor = s.Texture.Sample(v.Texture.X, v.Texture.Y)
    }

    v.Color = blendFunc[s.Texture.Mode](s.Texture, v.Color, tColor)
}

func (s *Shader) SetTexture(texture *Texture) {
	s.Texture = texture
}
