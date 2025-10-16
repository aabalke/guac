package gl

type Shader struct {
    Texture Texture
}

func NewShader() *Shader {
	return &Shader{}
}

func (s *Shader) Fragment(v *Vertex) {
    if s.Texture != nil {
        v.Color = s.Texture.Sample(v.Texture.X, v.Texture.Y)
    }
}

func (s *Shader) SetTexture(texture Texture) {
    s.Texture = texture
}
