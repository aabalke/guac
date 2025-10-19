package gl

type Shader struct {
    Texture *Texture
}

func NewShader() *Shader {
	return &Shader{}
}

const a = float64(1)/ 32

func (s *Shader) Fragment(v *Vertex) {

    if s.Texture == nil {
        return
    }

    texClr := s.Texture.Sample(v.Texture.X, v.Texture.Y)

    switch s.Texture.Mode {
    case 0:

        v.Color.R = ((texClr.R + a)*(v.Color.R + a) - a)
        v.Color.G = ((texClr.G + a)*(v.Color.G + a) - a)
        v.Color.B = ((texClr.B + a)*(v.Color.B + a) - a)
        v.Color.A = ((texClr.A + a)*(v.Color.A + a) - a)

    case 1:

        // alpha causes errors in mario kart

        if texClr.A != 0 {
            v.Color.R = ((texClr.R * texClr.A)+(v.Color.R*(1-texClr.A)) - a)
            v.Color.G = ((texClr.G * texClr.A)+(v.Color.G*(1-texClr.A)) - a)
            v.Color.B = ((texClr.B * texClr.A)+(v.Color.B*(1-texClr.A)) - a)
        }

        if texClr.A == 1 {
            v.Color.R = texClr.R
            v.Color.G = texClr.G
            v.Color.B = texClr.B
        }

    // toon and highlight shading not implimented

    default:
        v.Color = texClr
    }
}

func (s *Shader) SetTexture(texture *Texture) {
    s.Texture = texture
}
