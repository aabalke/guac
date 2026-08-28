package ui

import (
	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2"

	_ "embed"
)

//go:embed shader.kage
var shader []byte

var corrections = [config.ColorCorrectionTypeCnt][]float32{
	{},
	{26. / 32, 0. / 32., 6. / 32., 4. / 32, 24 / 32., 4. / 32., 2. / 32, 8. / 32, 22. / 32},
	{1, 0.05, 0.0, 0.05, 1, 0.05, 0, 0.05, 1.0},
	{1, 0.039, 0.196, 0.196, 0.901, 0.039, 0, 0.117, 0.862},
}

var gammas = [config.ColorCorrectionTypeCnt]float32{
	0.0,
	2.2,
	3.7,
	4.0,
}

type ColorCorrectionShader struct {
	Shader     *ebiten.Shader
	Image, Alt *ebiten.Image
	vertices   [4]ebiten.Vertex
	Type       *int
	Percent    *float32
}

func NewColorCorrectionShader(width, height int, ccType *int, percent *float32) *ColorCorrectionShader {
	shader, err := ebiten.NewShader(shader)
	if err != nil {
		panic(err)
	}
	c := &ColorCorrectionShader{
		Shader:  shader,
		Image:   ebiten.NewImage(width, height),
		Alt:     ebiten.NewImage(width, height),
		Type:    ccType,
		Percent: percent,
	}

	c.vertices[0].DstX = float32(0)
	c.vertices[0].DstY = float32(0)
	c.vertices[1].DstX = float32(width)
	c.vertices[1].DstY = float32(0)
	c.vertices[2].DstX = float32(0)
	c.vertices[2].DstY = float32(height)
	c.vertices[3].DstX = float32(width)
	c.vertices[3].DstY = float32(height)

	c.vertices[0].SrcX = float32(0)
	c.vertices[0].SrcY = float32(0)
	c.vertices[1].SrcX = float32(width)
	c.vertices[1].SrcY = float32(0)
	c.vertices[2].SrcX = float32(0)
	c.vertices[2].SrcY = float32(height)
	c.vertices[3].SrcX = float32(width)
	c.vertices[3].SrcY = float32(height)

	return c
}

func (c *ColorCorrectionShader) Draw(src *ebiten.Image) *ebiten.Image {
	// triangle shader options
	var shaderOpts ebiten.DrawTrianglesShaderOptions
	shaderOpts.Images[0] = src
	shaderOpts.Uniforms = make(map[string]any)
	shaderOpts.Uniforms["Percent"] = max(min(1, *c.Percent), 0)
	shaderOpts.Uniforms["ColorMatrix"] = corrections[*c.Type]
	shaderOpts.Uniforms["Gamma"] = gammas[*c.Type]

	// draw shader
	indices := []uint16{0, 1, 2, 2, 1, 3} // map vertices to triangles
	c.Image.DrawTrianglesShader(c.vertices[:], indices, c.Shader, &shaderOpts)

	return c.Image
}

func (c *ColorCorrectionShader) DrawAlt(src *ebiten.Image) *ebiten.Image {
	// triangle shader options
	var shaderOpts ebiten.DrawTrianglesShaderOptions
	shaderOpts.Images[0] = src
	shaderOpts.Uniforms = make(map[string]any)
	shaderOpts.Uniforms["Percent"] = max(min(1, *c.Percent), 0)
	shaderOpts.Uniforms["ColorMatrix"] = corrections[*c.Type]
	shaderOpts.Uniforms["Gamma"] = gammas[*c.Type]

	// draw shader
	indices := []uint16{0, 1, 2, 2, 1, 3} // map vertices to triangles
	c.Alt.DrawTrianglesShader(c.vertices[:], indices, c.Shader, &shaderOpts)

	return c.Alt
}
