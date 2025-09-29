package gl

import (
	"math"
)

type Shader interface {
	Vertex(Vertex) Vertex
	Fragment(Vertex) Color
    SetTexture(Texture)
}

// SolidColorShader renders with a single, solid color.
//type SolidColorShader struct {
//	Color  Color
//}
//
//func NewSolidColorShader(matrix Matrix, color Color) *SolidColorShader {
//	return &SolidColorShader{color}
//}
//
//func (shader *SolidColorShader) Vertex(v Vertex) Vertex {
//	return v
//}
//
//func (shader *SolidColorShader) Fragment(v Vertex) Color {
//	return shader.Color
//}
//
//func (shader *SolidColorShader) SetTexture(texture Texture) {
//}
//
type NdsShader struct {
    Texture Texture
    Lights [4]Light
    LightEnabled   [4]bool
}

func NewNdsShader(Lights [4]Light) *NdsShader {
	return &NdsShader{Lights: Lights}
}

func (shader *NdsShader) Vertex(v Vertex) Vertex {
	//v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

const (
    AMBIENT_COEF = 0.0
    DIFFUSE_COEF = 0.9
)

func (shader *NdsShader) Fragment(v Vertex) Color {

    vertexColor := v.Color
    //vertexColor.A = 0xFF

    if shader.Texture != nil {
        vertexColor = shader.Texture.Sample(v.Texture.X, v.Texture.Y)

        //bilinear sample needs texture coords
        //vertexColor = shader.Texture.BilinearSample(v.Texture.X, v.Texture.Y)
    }

    //vertexColor.A = 0xFF

    return vertexColor

    for i := range 4 {

        //if !shader.LightEnabled[i] {
        //    continue
        //}

        lightDirection := shader.Lights[i].Direction
        lightColor := shader.Lights[i].Color

        ambient := lightColor.MulScalar(AMBIENT_COEF)

        diffuse := math.Max(v.Normal.Dot(lightDirection), 0.0)
        diffuseColor := lightColor.MulScalar(diffuse * DIFFUSE_COEF)

        //vertexColor = vertexColor.Add(ambient).Add(diffuseColor)
        vertexColor = vertexColor.Add(ambient).Add(diffuseColor)
        //vertexColor = vertexColor.Add(White.Mul(ambient.Add(diffuseColor)))
    }

    vertexColor = vertexColor.Min(White)

    return vertexColor

    //color := White
    //color = Gray(0.5)

    ////switch {
    ////case shader.Texture == nil:
    ////    color = Gray(0.5)
    ////    //color = v.Color
    ////default:
    ////    color = shader.Texture.NearestNeightborSample(v.Texture.X, v.Texture.Y)
    ////    //color := shader.Texture.BilinearSample(v.Texture.X, v.Texture.Y)
    ////}

    ////return color

    //objColor := White

    //c := Gray(0.0)

    //for i := range 1 {
    //    ambient := shader.Lights[i].Color
    //    diffuse := math.Max(v.Normal.Dot(shader.Lights[i].Direction), 0)
    //    light = light.Add(White.MulScalar(diffuse))
    //    c = c.Add(color.Mul(light))
    //}
    //c = c.Min(White).Alpha(c.A)
	//return c

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

    //return color


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

type Light struct {
    Direction Vector
    Color Color
    Original Vector
}
