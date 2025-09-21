package gl

type Shader interface {
	Vertex(Vertex) Vertex
	Fragment(Vertex) Color
    SetTexture(Texture)
}

// SolidColorShader renders with a single, solid color.
type SolidColorShader struct {
	Matrix Matrix
	Color  Color
}

func NewSolidColorShader(matrix Matrix, color Color) *SolidColorShader {
	return &SolidColorShader{matrix, color}
}

func (shader *SolidColorShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

func (shader *SolidColorShader) Fragment(v Vertex) Color {
	return shader.Color
}

func (shader *SolidColorShader) SetTexture(texture Texture) {
}

type NdsShader struct {
	Matrix Matrix
    Texture Texture
}

func NewNdsShader(matrix Matrix) *NdsShader {
	return &NdsShader{Matrix: matrix}
}

func (shader *NdsShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

func (shader *NdsShader) Fragment(v Vertex) Color {

    if shader.Texture == nil {
        return v.Color
    }

    //return shader.Texture.NearestNeightborSample(v.Texture.X, v.Texture.Y)
    return shader.Texture.BilinearSample(v.Texture.X, v.Texture.Y)

    //diffuse := math.Max(v.Normal.Dot(shader.LightDirection), 0)
	//light = light.Add(shader.DiffuseColor.MulScalar(diffuse))
	//if diffuse > 0 && shader.SpecularPower > 0 {
	//	camera := shader.CameraPosition.Sub(v.Position).Normalize()
	//	reflected := shader.LightDirection.Negate().Reflect(v.Normal)
	//	specular := math.Max(camera.Dot(reflected), 0)
	//	if specular > 0 {
	//		specular = math.Pow(specular, shader.SpecularPower)
	//		light = light.Add(shader.SpecularColor.MulScalar(specular))
	//	}
	//}
	//return color.Mul(light).Min(White).Alpha(color.A)

	//return shader.Color
}

func (shader *NdsShader) SetTexture(texture Texture) {
    shader.Texture = texture
}
